package Argus

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Callback struct {
	Starter  string
	Callback any
}
type GenerationProps struct {
	Errors    Errors
	Callbacks []Callback
}

func (l Lib) HandleCli(props GenerationProps) int {
	if props.Errors == (Errors{}) {
		props.Errors = DefaultErrors
	}

	args := l.deps.Args

	// Need at least program name + command
	if len(args) < 2 {
		l.deps.Print(fmt.Sprintf(props.Errors.UnknowAction, "<none>"))
		return 1
	}

	command := args[1]
	commandArgs := args[2:]

	// Find the matching callback
	var matched *Callback
	for i := range props.Callbacks {
		if props.Callbacks[i].Starter == command {
			matched = &props.Callbacks[i]
			break
		}
	}

	if matched == nil {
		l.deps.Print(fmt.Sprintf(props.Errors.UnknowAction, command))
		return 1
	}

	// Use reflection to inspect the callback function
	callbackValue := reflect.ValueOf(matched.Callback)
	callbackType := callbackValue.Type()

	if callbackType.Kind() != reflect.Func {
		l.deps.Print(fmt.Sprintf("callback for '%s' is not a function", command))
		return 1
	}

	if callbackType.NumIn() != 1 {
		l.deps.Print(fmt.Sprintf("callback for '%s' must accept exactly one argument", command))
		return 1
	}

	// Create the entries struct
	entriesType := callbackType.In(0)
	entriesPtr := reflect.New(entriesType)
	entries := entriesPtr.Elem()

	// First pass: extract flags from args, collect remaining as positional
	positional, err := l.populateEntries(entries, entriesType, commandArgs, props.Errors)
	if err != "" {
		l.deps.Print(err)
		return 1
	}

	// Second pass: populate positional arguments
	errMsg := l.populatePositional(entries, entriesType, positional, props.Errors)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1
	}

	// Validate required fields
	errMsg = l.validateRequired(entries, entriesType, props.Errors)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1
	}

	// Call the callback
	if callbackType.NumOut() != 1 {
		l.deps.Print(fmt.Sprintf("callback for '%s' must return exactly one int", command))
		return 1
	}

	results := callbackValue.Call([]reflect.Value{entries})
	if len(results) == 1 && results[0].Kind() == reflect.Int {
		return int(results[0].Int())
	}

	return 0
}

// populateEntries extracts flags from the argument list and returns the remaining positional args.
func (l Lib) populateEntries(entries reflect.Value, entriesType reflect.Type, args []string, errs Errors) ([]string, string) {
	positional := []string{}
	consumed := make(map[int]bool)

	// First pass: extract all flags
	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")

		if entryType == "Flag" {
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				continue
			}
			ids := strings.Split(identifiers, ",")

			if field.Type.Kind() == reflect.Bool {
				// Boolean presence flag
				for j := 0; j < len(args); j++ {
					if consumed[j] {
						continue
					}
					for _, id := range ids {
						if args[j] == id {
							entries.Field(i).SetBool(true)
							consumed[j] = true
							break
						}
					}
				}
			} else {
				// Value flag
				found := false
				for j := 0; j < len(args)-1; j++ {
					if consumed[j] {
						continue
					}
					for _, id := range ids {
						if args[j] == id {
							consumed[j] = true
							consumed[j+1] = true
							errMsg := setFieldValue(entries.Field(i), field.Type, args[j+1])
							if errMsg != "" {
								return nil, fmt.Sprintf(errs.UnknowArg, args[j+1])
							}
							found = true
							break
						}
					}
					if found {
						break
					}
				}

				if !found {
					// Apply default if present
					defaultVal := field.Tag.Get("default")
					if defaultVal != "" {
						setFieldValue(entries.Field(i), field.Type, defaultVal)
					}
				}
			}
		} else if entryType == "ArrayFlag" {
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				continue
			}
			ids := strings.Split(identifiers, ",")

			// Collect all occurrences
			sliceType := field.Type
			elemType := sliceType.Elem()
			slice := reflect.MakeSlice(sliceType, 0, 0)

			for j := 0; j < len(args)-1; j++ {
				if consumed[j] {
					continue
				}
				for _, id := range ids {
					if args[j] == id {
						consumed[j] = true
						consumed[j+1] = true
						elem := reflect.New(elemType).Elem()
						errMsg := setFieldValue(elem, elemType, args[j+1])
						if errMsg != "" {
							return nil, fmt.Sprintf(errs.UnknowArg, args[j+1])
						}
						slice = reflect.Append(slice, elem)
						break
					}
				}
			}

			if slice.Len() > 0 {
				entries.Field(i).Set(slice)
			}

			// Validate min_size / max_size
			minSizeStr := field.Tag.Get("min_size")
			maxSizeStr := field.Tag.Get("max_size")
			errMsg := validateArraySize(entries.Field(i), field.Name, minSizeStr, maxSizeStr)
			if errMsg != "" {
				return nil, errMsg
			}
		}
	}

	// Collect positional arguments (everything not consumed by flags)
	for j := 0; j < len(args); j++ {
		if !consumed[j] {
			positional = append(positional, args[j])
		}
	}

	return positional, ""
}

// populatePositional fills Arg, NextArg, and ArrayArg fields from positional arguments.
func (l Lib) populatePositional(entries reflect.Value, entriesType reflect.Type, positional []string, errs Errors) string {
	nextArgIdx := 0

	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")

		switch entryType {
		case "Arg":
			posStr := field.Tag.Get("position")
			if posStr == "" {
				return fmt.Sprintf(errs.MissingArg, field.Name+" (missing position tag)")
			}
			pos, err := strconv.Atoi(posStr)
			if err != nil {
				return fmt.Sprintf(errs.MissingArg, field.Name+" (invalid position)")
			}
			if pos < len(positional) {
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[pos])
				if errMsg != "" {
					return fmt.Sprintf(errs.UnknowArg, positional[pos])
				}
			}

		case "NextArg":
			if nextArgIdx < len(positional) {
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[nextArgIdx])
				if errMsg != "" {
					return fmt.Sprintf(errs.UnknowArg, positional[nextArgIdx])
				}
				nextArgIdx++
			}

		case "ArrayArg":
			startStr := field.Tag.Get("start")
			endStr := field.Tag.Get("end")

			start := 0
			end := len(positional)

			if startStr != "" {
				s, err := strconv.Atoi(startStr)
				if err == nil {
					start = s
				}
			}
			if endStr != "" && endStr != "-1" {
				e, err := strconv.Atoi(endStr)
				if err == nil {
					end = e
				}
			}

			if start > len(positional) {
				start = len(positional)
			}
			if end > len(positional) {
				end = len(positional)
			}
			if start > end {
				start = end
			}

			sliceType := field.Type
			elemType := sliceType.Elem()
			slice := reflect.MakeSlice(sliceType, 0, end-start)

			for j := start; j < end; j++ {
				elem := reflect.New(elemType).Elem()
				errMsg := setFieldValue(elem, elemType, positional[j])
				if errMsg != "" {
					return fmt.Sprintf(errs.UnknowArg, positional[j])
				}
				slice = reflect.Append(slice, elem)
			}

			entries.Field(i).Set(slice)

			// Validate min_size / max_size
			minSizeStr := field.Tag.Get("min_size")
			maxSizeStr := field.Tag.Get("max_size")
			errMsg := validateArraySize(entries.Field(i), field.Name, minSizeStr, maxSizeStr)
			if errMsg != "" {
				return errMsg
			}
		}
	}

	return ""
}

// validateRequired checks that all required fields have been populated.
func (l Lib) validateRequired(entries reflect.Value, entriesType reflect.Type, errs Errors) string {
	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")

		// Bool flags are always optional
		if entryType == "Flag" && field.Type.Kind() == reflect.Bool {
			continue
		}

		// Check required status
		required := field.Tag.Get("required")
		defaultVal := field.Tag.Get("default")

		// Default is required:true, a default value implies optional
		if required == "false" || defaultVal != "" {
			continue
		}

		// Check if field is at its zero value
		fieldVal := entries.Field(i)
		if fieldVal.IsZero() {
			switch entryType {
			case "Flag", "ArrayFlag":
				return fmt.Sprintf(errs.MissingFlag, field.Name)
			default:
				return fmt.Sprintf(errs.MissingArg, field.Name)
			}
		}
	}

	return ""
}

// setFieldValue parses the string value and sets it on the given reflect.Value.
func setFieldValue(field reflect.Value, fieldType reflect.Type, value string) string {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Sprintf("cannot parse '%s' as int", value)
		}
		field.SetInt(n)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Sprintf("cannot parse '%s' as float64", value)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Sprintf("cannot parse '%s' as bool", value)
		}
		field.SetBool(b)
	default:
		return fmt.Sprintf("unsupported type: %s", fieldType.Kind())
	}
	return ""
}

// validateArraySize checks min_size and max_size constraints on a slice field.
func validateArraySize(field reflect.Value, fieldName string, minSizeStr string, maxSizeStr string) string {
	if minSizeStr != "" {
		minSize, err := strconv.Atoi(minSizeStr)
		if err == nil && field.Len() < minSize {
			return fmt.Sprintf("'%s' requires at least %d element(s), got %d", fieldName, minSize, field.Len())
		}
	}
	if maxSizeStr != "" && maxSizeStr != "-1" {
		maxSize, err := strconv.Atoi(maxSizeStr)
		if err == nil && field.Len() > maxSize {
			return fmt.Sprintf("'%s' allows at most %d element(s), got %d", fieldName, maxSize, field.Len())
		}
	}
	return ""
}

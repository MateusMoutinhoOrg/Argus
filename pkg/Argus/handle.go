package Argus

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type Callback struct {
	Starter     string
	Description string
	Callback    any
}
type GenerationProps struct {
	Name        string
	Description string
	Errors      Errors
	Callbacks   []Callback
}

func (l Lib) HandleCli(props GenerationProps) (int, error) {
	if props.Errors == (Errors{}) {
		props.Errors = DefaultErrors
	}

	// Validate all callbacks upfront (config/developer errors)
	if len(props.Callbacks) == 0 {
		return 1, fmt.Errorf("no callbacks provided")
	}

	for _, cb := range props.Callbacks {
		if cb.Starter == "" {
			return 1, fmt.Errorf("callback has an empty Starter")
		}

		if cb.Callback == nil {
			return 1, fmt.Errorf("callback for '%s' is nil", cb.Starter)
		}

		cbValue := reflect.ValueOf(cb.Callback)
		cbType := cbValue.Type()

		if cbType.Kind() != reflect.Func {
			return 1, fmt.Errorf("callback for '%s' is not a function", cb.Starter)
		}

		if cbType.NumIn() != 1 {
			return 1, fmt.Errorf("callback for '%s' must accept exactly one argument, got %d", cb.Starter, cbType.NumIn())
		}

		paramType := cbType.In(0)
		if paramType.Kind() != reflect.Struct {
			return 1, fmt.Errorf("callback for '%s' parameter must be a struct, got %s", cb.Starter, paramType.Kind())
		}

		if cbType.NumOut() != 1 {
			return 1, fmt.Errorf("callback for '%s' must return exactly one value, got %d", cb.Starter, cbType.NumOut())
		}

		if cbType.Out(0).Kind() != reflect.Int {
			return 1, fmt.Errorf("callback for '%s' must return int, got %s", cb.Starter, cbType.Out(0).Kind())
		}
	}

	args := l.deps.Args

	// Need at least program name + command
	if len(args) < 2 {
		l.deps.Print(fmt.Sprintf(props.Errors.UnknowAction, "<none>"))
		return 1, nil
	}

	command := args[1]
	commandArgs := args[2:]

	if command == "help" || command == "--help" || command == "-h" {
		if len(commandArgs) > 0 {
			cmd := commandArgs[0]
			for i := range props.Callbacks {
				if props.Callbacks[i].Starter == cmd {
					l.printCommandHelp(props, props.Callbacks[i])
					return 0, nil
				}
			}
		}
		l.printGlobalHelp(props)
		return 0, nil
	}

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
		return 1, nil
	}

	for _, arg := range commandArgs {
		if arg == "help" || arg == "--help" || arg == "-h" {
			l.printCommandHelp(props, *matched)
			return 0, nil
		}
	}

	// Use reflection to inspect the callback function
	callbackValue := reflect.ValueOf(matched.Callback)
	callbackType := callbackValue.Type()

	// Create the entries struct
	entriesType := callbackType.In(0)
	entriesPtr := reflect.New(entriesType)
	entries := entriesPtr.Elem()

	// First pass: extract flags from args, collect remaining as positional
	positional, err := l.populateEntries(entries, entriesType, commandArgs, props.Errors)
	if err != "" {
		l.deps.Print(err)
		return 1, nil
	}

	// Second pass: populate positional arguments
	errMsg := l.populatePositional(entries, entriesType, positional, props.Errors)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1, nil
	}

	// Validate required fields
	errMsg = l.validateRequired(entries, entriesType, props.Errors)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1, nil
	}

	// Call the callback
	results := callbackValue.Call([]reflect.Value{entries})
	return int(results[0].Int()), nil
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
				found := false
				for j := 0; j < len(args); j++ {
					if consumed[j] {
						continue
					}
					for _, id := range ids {
						if args[j] == id {
							entries.Field(i).SetBool(true)
							consumed[j] = true
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				
				if !found {
					defaultVal := field.Tag.Get("default")
					if defaultVal == "true" {
						entries.Field(i).SetBool(true)
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
							errMsg := setFieldValue(entries.Field(i), field.Type, args[j+1], field.Name, errs)
							if errMsg != "" {
								return nil, errMsg
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
						setFieldValue(entries.Field(i), field.Type, defaultVal, field.Name, errs)
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
						errMsg := setFieldValue(elem, elemType, args[j+1], field.Name, errs)
						if errMsg != "" {
							return nil, errMsg
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
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[pos], field.Name, errs)
				if errMsg != "" {
					return errMsg
				}
			}

		case "NextArg":
			if nextArgIdx < len(positional) {
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[nextArgIdx], field.Name, errs)
				if errMsg != "" {
					return errMsg
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
				errMsg := setFieldValue(elem, elemType, positional[j], field.Name, errs)
				if errMsg != "" {
					return errMsg
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
func setFieldValue(field reflect.Value, fieldType reflect.Type, value string, fieldName string, errs Errors) string {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Sprintf(errs.NaN, strings.ToLower(fieldName))
		}
		field.SetInt(n)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Sprintf(errs.NaN, strings.ToLower(fieldName))
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

func (l Lib) printGlobalHelp(props GenerationProps) {
	appName := props.Name
	if appName == "" && len(l.deps.Args) > 0 {
		appName = filepath.Base(l.deps.Args[0])
	}
	if appName == "" {
		appName = "app"
	}

	if props.Description != "" {
		l.deps.Print(fmt.Sprintf("%s\n", props.Description))
	}

	l.deps.Print("USAGE:")
	l.deps.Print(fmt.Sprintf("  %s <command> [arguments...]\n", appName))

	l.deps.Print("COMMANDS:")
	maxLen := 0
	for _, cb := range props.Callbacks {
		if len(cb.Starter) > maxLen {
			maxLen = len(cb.Starter)
		}
	}

	for _, cb := range props.Callbacks {
		desc := cb.Description
		if desc == "" {
			desc = "No description provided."
		}
		padding := strings.Repeat(" ", maxLen-len(cb.Starter)+2)
		l.deps.Print(fmt.Sprintf("  %s%s%s", cb.Starter, padding, desc))
	}

	l.deps.Print(fmt.Sprintf("\nRun '%s help <command>' for more information on a command.", appName))
}

func (l Lib) printCommandHelp(props GenerationProps, cb Callback) {
	appName := props.Name
	if appName == "" && len(l.deps.Args) > 0 {
		appName = filepath.Base(l.deps.Args[0])
	}
	if appName == "" {
		appName = "app"
	}

	if cb.Description != "" {
		l.deps.Print(fmt.Sprintf("%s\n", cb.Description))
	}

	l.deps.Print("USAGE:")
	l.deps.Print(fmt.Sprintf("  %s %s [arguments...]\n", appName, cb.Starter))

	cbValue := reflect.ValueOf(cb.Callback)
	entriesType := cbValue.Type().In(0)

	type fieldInfo struct {
		name   string
		desc   string
		isFlag bool
	}

	var infos []fieldInfo
	maxFlagLen := 0
	maxArgLen := 0

	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")
		helpMsg := field.Tag.Get("help")
		if helpMsg == "" {
			helpMsg = "No description"
		}

		info := fieldInfo{desc: helpMsg}

		if entryType == "Flag" || entryType == "ArrayFlag" {
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				continue
			}
			info.name = identifiers
			info.isFlag = true
			if len(identifiers) > maxFlagLen {
				maxFlagLen = len(identifiers)
			}
		} else if entryType == "Arg" || entryType == "NextArg" || entryType == "ArrayArg" {
			info.name = field.Name
			info.isFlag = false
			if len(field.Name) > maxArgLen {
				maxArgLen = len(field.Name)
			}
		} else {
			continue
		}
		infos = append(infos, info)
	}

	hasFlags := false
	hasArgs := false
	for _, info := range infos {
		if info.isFlag {
			hasFlags = true
		} else {
			hasArgs = true
		}
	}

	if hasArgs {
		l.deps.Print("ARGUMENTS:")
		for _, info := range infos {
			if !info.isFlag {
				padding := strings.Repeat(" ", maxArgLen-len(info.name)+2)
				l.deps.Print(fmt.Sprintf("  %s%s%s", info.name, padding, info.desc))
			}
		}
		if hasFlags {
			l.deps.Print("")
		}
	}

	if hasFlags {
		l.deps.Print("FLAGS:")
		for _, info := range infos {
			if info.isFlag {
				padding := strings.Repeat(" ", maxFlagLen-len(info.name)+2)
				l.deps.Print(fmt.Sprintf("  %s%s%s", info.name, padding, info.desc))
			}
		}
	}
}

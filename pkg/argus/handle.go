package argus

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

type Callback struct {
	Starter     string
	Description string
	Samples     []string
	Callback    any
}
type GenerationProps struct {
	Name             string
	DisableQuiet     bool     // if true, the quiet system will not work (default: false)
	QuietIdentifiers []string // the quiet flags to set quiet mode (default: ["--quiet", "-q"])
	Description      string
	Messages         Messages
	Callbacks        []Callback
}

func (l Lib) HandleCli(props GenerationProps) (int, error) {
	if props.Messages.MissingFlag == nil {
		props.Messages = DefaultMessages
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

		numIn := cbType.NumIn()
		if numIn < 1 || numIn > 2 {
			return 1, fmt.Errorf("callback for '%s' must accept 1 or 2 arguments, got %d", cb.Starter, numIn)
		}

		paramType := cbType.In(0)
		if paramType.Kind() != reflect.Struct {
			return 1, fmt.Errorf("callback for '%s' first parameter must be a struct, got %s", cb.Starter, paramType.Kind())
		}

		if numIn == 2 {
			depsType := reflect.TypeOf((*deps.Deps)(nil)).Elem()
			if !cbType.In(1).Implements(depsType) && cbType.In(1) != depsType {
				return 1, fmt.Errorf("callback for '%s' second parameter must be deps.Deps, got %s", cb.Starter, cbType.In(1))
			}
		}

		if cbType.NumOut() != 1 {
			return 1, fmt.Errorf("callback for '%s' must return exactly one value, got %d", cb.Starter, cbType.NumOut())
		}

		if cbType.Out(0).Kind() != reflect.Int {
			return 1, fmt.Errorf("callback for '%s' must return int, got %s", cb.Starter, cbType.Out(0).Kind())
		}

		// Validate struct field tags
		if err := validateStructTags(paramType); err != nil {
			return 1, fmt.Errorf("callback for '%s': %w", cb.Starter, err)
		}
	}

	args := l.deps.GetArgs()

	// Quiet mode: strip quiet flags from the args and silence all output
	if !props.DisableQuiet {
		quietIdentifiers := props.QuietIdentifiers
		if len(quietIdentifiers) == 0 {
			quietIdentifiers = []string{"--quiet", "-q"}
		}

		filtered := args[:0:0]
		for i, arg := range args {
			isQuiet := false
			if i > 0 {
				for _, id := range quietIdentifiers {
					if arg == id {
						isQuiet = true
						break
					}
				}
			}
			if isQuiet {
				l.deps.SetQuiet()
			} else {
				filtered = append(filtered, arg)
			}
		}
		args = filtered
	}

	// Need at least program name + command
	if len(args) < 2 {
		l.deps.Print(props.Messages.UnknownAction(""))
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
		l.deps.Print(props.Messages.UnknownAction(command))
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
	positional, err := l.populateEntries(entries, entriesType, commandArgs, props.Messages)
	if err != "" {
		l.deps.Print(err)
		return 1, nil
	}

	// Second pass: populate positional arguments
	errMsg := l.populatePositional(entries, entriesType, positional, props.Messages)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1, nil
	}

	// Validate required fields
	errMsg = l.validateRequired(entries, entriesType, props.Messages)
	if errMsg != "" {
		l.deps.Print(errMsg)
		return 1, nil
	}

	// Call the callback
	callArgs := []reflect.Value{entries}
	if callbackType.NumIn() == 2 {
		callArgs = append(callArgs, reflect.ValueOf(l.deps))
	}
	results := callbackValue.Call(callArgs)
	return int(results[0].Int()), nil
}

// populateEntries extracts flags from the argument list and returns the remaining positional args.
func (l Lib) populateEntries(entries reflect.Value, entriesType reflect.Type, args []string, msgs Messages) ([]string, string) {
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
							errMsg := setFieldValue(entries.Field(i), field.Type, args[j+1], field.Name, msgs)
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
						setFieldValue(entries.Field(i), field.Type, defaultVal, field.Name, msgs)
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
						errMsg := setFieldValue(elem, elemType, args[j+1], field.Name, msgs)
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
func (l Lib) populatePositional(entries reflect.Value, entriesType reflect.Type, positional []string, msgs Messages) string {
	nextArgIdx := 0

	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")
		description := field.Tag.Get("description")

		switch entryType {
		case "Arg":
			posStr := field.Tag.Get("position")
			if posStr == "" {
				return msgs.MissingArg(field.Name+" (missing position tag)", description, "")
			}
			pos, err := strconv.Atoi(posStr)
			if err != nil {
				return msgs.MissingArg(field.Name+" (invalid position)", description, posStr)
			}
			if pos < len(positional) {
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[pos], field.Name, msgs)
				if errMsg != "" {
					return errMsg
				}
			}

		case "NextArg":
			if nextArgIdx < len(positional) {
				errMsg := setFieldValue(entries.Field(i), field.Type, positional[nextArgIdx], field.Name, msgs)
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
				errMsg := setFieldValue(elem, elemType, positional[j], field.Name, msgs)
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
func (l Lib) validateRequired(entries reflect.Value, entriesType reflect.Type, msgs Messages) string {
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
			description := field.Tag.Get("description")
			switch entryType {
			case "Flag", "ArrayFlag":
				identifiers := field.Tag.Get("identifiers")
				if identifiers != "" {
					return msgs.MissingFlag(strings.Split(identifiers, ",")[0], description)
				}
				return msgs.MissingFlag(field.Name, description)
			default:
				position := field.Tag.Get("position")
				return msgs.MissingArg(field.Name, description, position)
			}
		}
	}

	return ""
}

// setFieldValue parses the string value and sets it on the given reflect.Value.
func setFieldValue(field reflect.Value, fieldType reflect.Type, value string, fieldName string, msgs Messages) string {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return msgs.NaN(strings.ToLower(fieldName))
		}
		field.SetInt(n)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return msgs.NaN(strings.ToLower(fieldName))
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
	if appName == "" && len(l.deps.GetArgs()) > 0 {
		appName = filepath.Base(l.deps.GetArgs()[0])
	}
	if appName == "" {
		appName = "app"
	}

	if props.Description != "" {
		l.deps.Print(props.Description)
		l.deps.Print("")
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

// validateStructTags validates that all struct field tags are correctly formed.
func validateStructTags(entriesType reflect.Type) error {
	validTypes := map[string]bool{
		"Flag":      true,
		"ArrayFlag": true,
		"Arg":       true,
		"NextArg":   true,
		"ArrayArg":  true,
	}

	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")

		// type tag is required for exported fields
		if entryType == "" {
			return fmt.Errorf("field '%s' missing 'type' tag", field.Name)
		}

		// Check if type is valid
		if !validTypes[entryType] {
			return fmt.Errorf("field '%s' has invalid type '%s', must be one of: Flag, ArrayFlag, Arg, NextArg, ArrayArg", field.Name, entryType)
		}

		// Validate type-specific requirements
		switch entryType {
		case "Flag", "ArrayFlag":
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				return fmt.Errorf("field '%s' (type:%s) missing 'identifiers' tag", field.Name, entryType)
			}
		case "Arg":
			position := field.Tag.Get("position")
			if position == "" {
				return fmt.Errorf("field '%s' (type:Arg) missing 'position' tag", field.Name)
			}
		}
	}

	return nil
}

func (l Lib) printCommandHelp(props GenerationProps, cb Callback) {
	appName := props.Name
	if appName == "" && len(l.deps.GetArgs()) > 0 {
		appName = filepath.Base(l.deps.GetArgs()[0])
	}
	if appName == "" {
		appName = "app"
	}

	if cb.Description != "" {
		l.deps.Print(cb.Description)
		l.deps.Print("")
	}

	l.deps.Print("USAGE:")
	l.deps.Print(fmt.Sprintf("  %s %s [arguments...]\n", appName, cb.Starter))

	cbValue := reflect.ValueOf(cb.Callback)
	entriesType := cbValue.Type().In(0)

	type fieldInfo struct {
		name        string
		desc        string
		isFlag      bool
		required    bool
		identifiers string
	}

	var infos []fieldInfo
	maxFlagLen := 0
	maxArgLen := 0

	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		entryType := field.Tag.Get("type")
		helpMsg := field.Tag.Get("help")
		if helpMsg == "" {
			helpMsg = field.Tag.Get("description")
		}
		if helpMsg == "" {
			helpMsg = "No description"
		}

		required := field.Tag.Get("required")
		defaultVal := field.Tag.Get("default")
		isRequired := required != "false" && defaultVal == "" && !(entryType == "Flag" && field.Type.Kind() == reflect.Bool)

		info := fieldInfo{desc: helpMsg, required: isRequired}

		if entryType == "Flag" || entryType == "ArrayFlag" {
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				continue
			}
			info.name = identifiers
			info.identifiers = identifiers
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
				required := ""
				if info.required {
					required = " (required)"
				}
				padding := strings.Repeat(" ", maxArgLen-len(info.name)+2)
				l.deps.Print(fmt.Sprintf("  %s%s%s%s", info.name, padding, info.desc, required))
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
				required := ""
				if info.required {
					required = " (required)"
				}
				padding := strings.Repeat(" ", maxFlagLen-len(info.name)+2)
				l.deps.Print(fmt.Sprintf("  %s%s%s%s", info.name, padding, info.desc, required))
			}
		}
	}

	if len(cb.Samples) > 0 {
		l.deps.Print("")
		l.deps.Print("SAMPLES:")
		for _, sample := range cb.Samples {
			l.deps.Print(fmt.Sprintf("  %s %s %s", appName, cb.Starter, sample))
		}
	}
}

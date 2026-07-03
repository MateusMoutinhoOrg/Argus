package argus

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

var depsType = reflect.TypeOf(deps.Deps{})

// entryKind describes how a field is populated, inferred from its
// containing struct (Args/Flags), its type, and its tags.
type entryKind int

const (
	entryFlag entryKind = iota
	entryArrayFlag
	entryArg
	entryNextArg
	entryArrayArg
)

// classifyArgField infers the entry kind for a field declared in an Args struct:
//   - slice type            -> ArrayArg
//   - has a "position" tag  -> Arg
//   - otherwise             -> NextArg
func classifyArgField(field reflect.StructField) entryKind {
	if field.Type.Kind() == reflect.Slice {
		return entryArrayArg
	}
	if field.Tag.Get("position") != "" {
		return entryArg
	}
	return entryNextArg
}

// classifyFlagField infers the entry kind for a field declared in a Flags struct:
//   - slice type -> ArrayFlag
//   - otherwise  -> Flag
func classifyFlagField(field reflect.StructField) entryKind {
	if field.Type.Kind() == reflect.Slice {
		return entryArrayFlag
	}
	return entryFlag
}

type Callback struct {
	Starter     string
	Description string
	Callback    any
}
type GenerationProps struct {
	Name        string
	Description string
	Messages    Messages
	Callbacks   []Callback
	// InvalidateQuietMode disables the global --quiet/-q flag. By default
	// (false) the app supports --quiet/-q anywhere in the CLI args, which
	// suppresses all further output through deps.Deps.Print. Set to true to
	// turn this behavior off.
	InvalidateQuietMode bool
	// InvalidateHelpMode disables the built-in "help"/"--help"/"-h" handling.
	// By default (false) the app supports these to print global or
	// command-specific help. Set to true to turn this behavior off.
	InvalidateHelpMode bool
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

		// Validate struct field tags
		if err := validateCallbackStruct(paramType); err != nil {
			return 1, fmt.Errorf("callback for '%s': %w", cb.Starter, err)
		}
	}

	args := l.deps.Args

	// Global quiet mode: strip --quiet/-q from anywhere in the args and
	// suppress all further output via deps.Deps.Print.
	if !props.InvalidateQuietMode {
		filtered := make([]string, 0, len(args))
		for _, arg := range args {
			if arg == "--quiet" || arg == "-q" {
				if l.deps.Quiet != nil {
					*l.deps.Quiet = true
				}
				continue
			}
			filtered = append(filtered, arg)
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

	if !props.InvalidateHelpMode && (command == "help" || command == "--help" || command == "-h") {
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

	if !props.InvalidateHelpMode {
		for _, arg := range commandArgs {
			if arg == "help" || arg == "--help" || arg == "-h" {
				l.printCommandHelp(props, *matched)
				return 0, nil
			}
		}
	}

	// Use reflection to inspect the callback function
	callbackValue := reflect.ValueOf(matched.Callback)
	callbackType := callbackValue.Type()

	// Create the entries struct
	entriesType := callbackType.In(0)
	entriesPtr := reflect.New(entriesType)
	entries := entriesPtr.Elem()

	// Inject the Deps used by this Lib into any deps.Deps field (exported or not)
	injectDeps(entries, entriesType, l.deps)

	// First pass: extract flags from args, collect remaining as positional
	positional, err := l.populateFlags(entries, entriesType, commandArgs, props.Messages)
	if err != "" {
		l.deps.Print(err)
		return 1, nil
	}

	// Second pass: populate positional arguments
	errMsg := l.populateArgs(entries, entriesType, positional, props.Messages)
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
	results := callbackValue.Call([]reflect.Value{entries})
	return int(results[0].Int()), nil
}

// injectDeps finds a field of type deps.Deps on the callback struct (exported
// or not) and populates it with the Lib's Deps, so callbacks can access
// Args/Print without going through fmt/os directly.
func injectDeps(entries reflect.Value, entriesType reflect.Type, d deps.Deps) {
	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)
		if field.Type != depsType {
			continue
		}
		target := entries.Field(i)
		if target.CanSet() {
			target.Set(reflect.ValueOf(d))
			return
		}
		// Unexported field: bypass Go's reflect write protection.
		ptr := unsafe.Pointer(target.UnsafeAddr())
		reflect.NewAt(target.Type(), ptr).Elem().Set(reflect.ValueOf(d))
		return
	}
}

// findSubStruct locates a nested struct field ("Args" or "Flags") on the
// top-level callback struct, returning its type/value and whether it exists.
func findSubStruct(entries reflect.Value, entriesType reflect.Type, name string) (reflect.Value, reflect.Type, bool) {
	field, ok := entriesType.FieldByName(name)
	if !ok {
		return reflect.Value{}, nil, false
	}
	return entries.FieldByName(name), field.Type, true
}

// populateFlags extracts flags from the argument list (via the Flags sub-struct,
// if present) and returns the remaining positional args.
func (l Lib) populateFlags(entries reflect.Value, entriesType reflect.Type, args []string, msgs Messages) ([]string, string) {
	positional := []string{}
	consumed := make(map[int]bool)

	flagsValue, flagsType, hasFlags := findSubStruct(entries, entriesType, "Flags")

	if hasFlags {
		for i := 0; i < flagsType.NumField(); i++ {
			field := flagsType.Field(i)
			fieldValue := flagsValue.Field(i)

			switch classifyFlagField(field) {
			case entryFlag:
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
								fieldValue.SetBool(true)
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
							fieldValue.SetBool(true)
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
								errMsg := setFieldValue(fieldValue, field.Type, args[j+1], field.Name, msgs)
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
							setFieldValue(fieldValue, field.Type, defaultVal, field.Name, msgs)
						}
					}
				}
			case entryArrayFlag:
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
					fieldValue.Set(slice)
				}

				// Validate min_size / max_size
				minSizeStr := field.Tag.Get("min_size")
				maxSizeStr := field.Tag.Get("max_size")
				errMsg := validateArraySize(fieldValue, field.Name, minSizeStr, maxSizeStr)
				if errMsg != "" {
					return nil, errMsg
				}
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

// populateArgs fills Arg, NextArg, and ArrayArg fields (via the Args sub-struct,
// if present) from positional arguments.
func (l Lib) populateArgs(entries reflect.Value, entriesType reflect.Type, positional []string, msgs Messages) string {
	argsValue, argsType, hasArgs := findSubStruct(entries, entriesType, "Args")
	if !hasArgs {
		return ""
	}

	nextArgIdx := 0

	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		fieldValue := argsValue.Field(i)
		description := field.Tag.Get("description")

		switch classifyArgField(field) {
		case entryArg:
			posStr := field.Tag.Get("position")
			pos, err := strconv.Atoi(posStr)
			if err != nil {
				return msgs.MissingArg(field.Name+" (invalid position)", description, posStr)
			}
			if pos < len(positional) {
				errMsg := setFieldValue(fieldValue, field.Type, positional[pos], field.Name, msgs)
				if errMsg != "" {
					return errMsg
				}
			}

		case entryNextArg:
			if nextArgIdx < len(positional) {
				errMsg := setFieldValue(fieldValue, field.Type, positional[nextArgIdx], field.Name, msgs)
				if errMsg != "" {
					return errMsg
				}
				nextArgIdx++
			}

		case entryArrayArg:
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

			fieldValue.Set(slice)

			// Validate min_size / max_size
			minSizeStr := field.Tag.Get("min_size")
			maxSizeStr := field.Tag.Get("max_size")
			errMsg := validateArraySize(fieldValue, field.Name, minSizeStr, maxSizeStr)
			if errMsg != "" {
				return errMsg
			}
		}
	}

	return ""
}

// validateRequired checks that all required fields (across Args and Flags) have been populated.
func (l Lib) validateRequired(entries reflect.Value, entriesType reflect.Type, msgs Messages) string {
	if flagsValue, flagsType, ok := findSubStruct(entries, entriesType, "Flags"); ok {
		for i := 0; i < flagsType.NumField(); i++ {
			field := flagsType.Field(i)
			kind := classifyFlagField(field)

			// Bool flags are always optional
			if kind == entryFlag && field.Type.Kind() == reflect.Bool {
				continue
			}

			required := field.Tag.Get("required")
			defaultVal := field.Tag.Get("default")
			if required == "false" || defaultVal != "" {
				continue
			}

			if flagsValue.Field(i).IsZero() {
				description := field.Tag.Get("description")
				identifiers := field.Tag.Get("identifiers")
				if identifiers != "" {
					return msgs.MissingFlag(strings.Split(identifiers, ",")[0], description)
				}
				return msgs.MissingFlag(field.Name, description)
			}
		}
	}

	if argsValue, argsType, ok := findSubStruct(entries, entriesType, "Args"); ok {
		for i := 0; i < argsType.NumField(); i++ {
			field := argsType.Field(i)

			required := field.Tag.Get("required")
			defaultVal := field.Tag.Get("default")
			if required == "false" || defaultVal != "" {
				continue
			}

			if argsValue.Field(i).IsZero() {
				description := field.Tag.Get("description")
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
	if appName == "" && len(l.deps.Args) > 0 {
		appName = filepath.Base(l.deps.Args[0])
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

// validateCallbackStruct validates the shape of a callback parameter struct.
// It may contain, at the top level, only:
//   - a field named "Args"  (struct)  — positional arguments
//   - a field named "Flags" (struct)  — named flags
//   - a field of type deps.Deps (exported or not) — auto-injected dependencies
func validateCallbackStruct(entriesType reflect.Type) error {
	for i := 0; i < entriesType.NumField(); i++ {
		field := entriesType.Field(i)

		if field.Type == depsType {
			continue
		}

		switch field.Name {
		case "Args":
			if field.Type.Kind() != reflect.Struct {
				return fmt.Errorf("field 'Args' must be a struct, got %s", field.Type.Kind())
			}
			if err := validateArgsStruct(field.Type); err != nil {
				return err
			}
		case "Flags":
			if field.Type.Kind() != reflect.Struct {
				return fmt.Errorf("field 'Flags' must be a struct, got %s", field.Type.Kind())
			}
			if err := validateFlagsStruct(field.Type); err != nil {
				return err
			}
		default:
			return fmt.Errorf("field '%s' is not recognized; the callback struct may only contain 'Args', 'Flags', and an optional deps.Deps field", field.Name)
		}
	}

	return nil
}

// validateArgsStruct validates the fields of an Args sub-struct.
func validateArgsStruct(argsType reflect.Type) error {
	for i := 0; i < argsType.NumField(); i++ {
		field := argsType.Field(i)
		if !field.IsExported() {
			return fmt.Errorf("field '%s' in Args must be exported", field.Name)
		}

		if classifyArgField(field) == entryArg {
			posStr := field.Tag.Get("position")
			if _, err := strconv.Atoi(posStr); err != nil {
				return fmt.Errorf("field '%s' (Arg) has an invalid 'position' tag %q", field.Name, posStr)
			}
		}
	}
	return nil
}

// validateFlagsStruct validates the fields of a Flags sub-struct.
func validateFlagsStruct(flagsType reflect.Type) error {
	for i := 0; i < flagsType.NumField(); i++ {
		field := flagsType.Field(i)
		if !field.IsExported() {
			return fmt.Errorf("field '%s' in Flags must be exported", field.Name)
		}

		identifiers := field.Tag.Get("identifiers")
		if identifiers == "" {
			kind := "Flag"
			if classifyFlagField(field) == entryArrayFlag {
				kind = "ArrayFlag"
			}
			return fmt.Errorf("field '%s' (%s) missing 'identifiers' tag", field.Name, kind)
		}
	}
	return nil
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

	describe := func(field reflect.StructField) string {
		helpMsg := field.Tag.Get("help")
		if helpMsg == "" {
			helpMsg = field.Tag.Get("description")
		}
		if helpMsg == "" {
			helpMsg = "No description"
		}
		return helpMsg
	}

	if flagsField, ok := entriesType.FieldByName("Flags"); ok {
		flagsType := flagsField.Type
		for i := 0; i < flagsType.NumField(); i++ {
			field := flagsType.Field(i)
			identifiers := field.Tag.Get("identifiers")
			if identifiers == "" {
				continue
			}

			required := field.Tag.Get("required")
			defaultVal := field.Tag.Get("default")
			isRequired := required != "false" && defaultVal == "" && field.Type.Kind() != reflect.Bool

			info := fieldInfo{
				name:        identifiers,
				desc:        describe(field),
				isFlag:      true,
				required:    isRequired,
				identifiers: identifiers,
			}
			if len(identifiers) > maxFlagLen {
				maxFlagLen = len(identifiers)
			}
			infos = append(infos, info)
		}
	}

	if argsField, ok := entriesType.FieldByName("Args"); ok {
		argsType := argsField.Type
		for i := 0; i < argsType.NumField(); i++ {
			field := argsType.Field(i)

			required := field.Tag.Get("required")
			defaultVal := field.Tag.Get("default")
			isRequired := required != "false" && defaultVal == ""

			info := fieldInfo{
				name:     field.Name,
				desc:     describe(field),
				isFlag:   false,
				required: isRequired,
			}
			if len(field.Name) > maxArgLen {
				maxArgLen = len(field.Name)
			}
			infos = append(infos, info)
		}
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
}

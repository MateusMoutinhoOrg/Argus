package argus

import "fmt"

type Messages struct {
	MissingFlag  func(flag, description string) string
	MissingArg   func(arg, description, position string) string
	UnknownAction func(string) string
	UnknownArg    func(string) string
	NaN          func(string) string
}

var DefaultMessages = Messages{
	MissingFlag: func(flag, description string) string {
		if description != "" {
			return fmt.Sprintf("error: missing required flag %s  (%s)\n", flag, description)
		}
		return fmt.Sprintf("error: missing required flag %s (description not provided)\n", flag)
	},

	MissingArg: func(arg, description, position string) string {
		msg := fmt.Sprintf("error: missing required argument %s", arg)
		if position != "" {
			msg += fmt.Sprintf(" (position %s)", position)
		}
		if description != "" {
			msg += fmt.Sprintf("\n  %s", description)
		}
		return msg
	},
	UnknownAction: func(action string) string {
		if action == "" {
			return "error: action (argv[1]) not provided"
		}
		return fmt.Sprintf("error: unknown action %s", action)
	},
	UnknownArg: func(arg string) string {
		return fmt.Sprintf("error: unknown arg %s", arg)
	},
	NaN: func(flag string) string {
		return fmt.Sprintf("error: flag %s is not a number", flag)
	},
}

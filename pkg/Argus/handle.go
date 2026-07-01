package Argus

type Callback struct {
	Starter  string
	Callback any
}
type GenerationProps struct {
	Errors          Errors
	PrintHelp       bool     //defaults to true
	HelpIdentifiers []string //defaults to ["-h", "--help"]
	Commands        map[string]Callback
}

func (l Lib) HandleCli(props GenerationProps) int {
	if props.Errors == (Errors{}) {
		props.Errors = DefaultErrors
	}
	//return 0 on fail
	return 0
}

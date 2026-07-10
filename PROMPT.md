quiet flag implementation:
implement the --quiet flag system , that allows the app to not print any msg  in the terminal.

Changes: 
### GenerationProps:
~~~go
type GenerationProps struct {
	Name        string
    DisableQuiet bool //if true,quiet system will not work (default: false)
    QuietIdentifiers []string // the quiet flags to set quiet mode (default: ["--quiet", "-q"])
	Description string
	Messages    Messages
	Callbacks   []Callback
}

Add on the documentation the explanation to quiet mode.
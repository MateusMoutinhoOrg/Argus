package deps

type Deps interface {
	GetArgs() []string
	Print(s string) // if quiet is true, print nothing
	SetQuiet()      // set the application in quiet mode
}

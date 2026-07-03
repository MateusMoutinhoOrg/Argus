package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

// NextArgArgs demonstrates sequential positional argument consumption.
// Each field without a `position` tag consumes the next unclaimed
// positional argument, in the order the fields are declared — Argus infers
// NextArg for any non-slice field in an Args sub-struct that has no
// `position` tag.
type NextArgArgs struct {
	Src  string `description:"source file path"`
	Dest string `description:"destination file path"`
}

type NextArgEntries struct {
	Args NextArgArgs
}

func copyFile(e NextArgEntries) int {
	fmt.Printf("Copying '%s' → '%s'\n", e.Args.Src, e.Args.Dest)
	return 0
}

// FixedArgArgs demonstrates fixed-position positional arguments.
// A `position` tag on a field in an Args sub-struct makes Argus infer Arg
// instead of NextArg.
type FixedArgArgs struct {
	Filename string `position:"0" description:"path to the file to open"`
	LineNum  int    `position:"1" description:"line number to navigate to"`
	ColNum   int    `position:"2" required:"false" description:"column number (optional)"`
}

type FixedArgEntries struct {
	Args FixedArgArgs
}

func gotoLine(e FixedArgEntries) int {
	if e.Args.ColNum > 0 {
		fmt.Printf("Opening '%s' at line %d, column %d\n", e.Args.Filename, e.Args.LineNum, e.Args.ColNum)
	} else {
		fmt.Printf("Opening '%s' at line %d\n", e.Args.Filename, e.Args.LineNum)
	}
	return 0
}

// Usage:
//
//	go run samples/positional/positional.go copy readme.md /tmp/backup.md
//	go run samples/positional/positional.go goto main.go 42
//	go run samples/positional/positional.go goto main.go 42 10
func main() {
	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{Starter: "copy", Callback: copyFile, Description: "Copy a file from source to destination"},
			{Starter: "goto", Callback: gotoLine, Description: "Go to a specific line in a file"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

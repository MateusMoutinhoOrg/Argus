package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// NextArgEntries demonstrates sequential positional argument consumption.
// Each NextArg field binds to the next unclaimed positional argument
// in the order the fields are declared.
type NextArgEntries struct {
	Src  string `type:"NextArg" description:"source file path"`
	Dest string `type:"NextArg" description:"destination file path"`
}

func copyFile(e NextArgEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Copying '%s' → '%s'\n", e.Src, e.Dest))
	return 0
}

// FixedArgEntries demonstrates fixed-position positional arguments.
// Each Arg field binds to a specific positional index via the position tag.
type FixedArgEntries struct {
	Filename string `type:"Arg" position:"0" description:"path to the file to open"`
	LineNum  int    `type:"Arg" position:"1" description:"line number to navigate to"`
	ColNum   int    `type:"Arg" position:"2" required:"false" description:"column number (optional)"`
}

func gotoLine(e FixedArgEntries, deps argus_dep.Deps) int {
	if e.ColNum > 0 {
		deps.Print(fmt.Sprintf("Opening '%s' at line %d, column %d\n", e.Filename, e.LineNum, e.ColNum))
	} else {
		deps.Print(fmt.Sprintf("Opening '%s' at line %d\n", e.Filename, e.LineNum))
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

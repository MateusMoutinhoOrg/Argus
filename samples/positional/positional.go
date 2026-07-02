package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

// NextArgEntries demonstrates sequential positional argument consumption.
// Each NextArg field binds to the next unclaimed positional argument
// in the order the fields are declared.
type NextArgEntries struct {
	Src  string `type:"NextArg"`
	Dest string `type:"NextArg"`
}

func copyFile(e NextArgEntries) int {
	fmt.Printf("Copying '%s' → '%s'\n", e.Src, e.Dest)
	return 0
}

// FixedArgEntries demonstrates fixed-position positional arguments.
// Each Arg field binds to a specific positional index via the position tag.
type FixedArgEntries struct {
	Filename  string `type:"Arg" position:"0"`
	LineNum   int    `type:"Arg" position:"1"`
	ColNum    int    `type:"Arg" position:"2" required:"false"`
}

func gotoLine(e FixedArgEntries) int {
	if e.ColNum > 0 {
		fmt.Printf("Opening '%s' at line %d, column %d\n", e.Filename, e.LineNum, e.ColNum)
	} else {
		fmt.Printf("Opening '%s' at line %d\n", e.Filename, e.LineNum)
	}
	return 0
}

// Usage:
//
//	go run samples/positional/positional.go copy readme.md /tmp/backup.md
//	go run samples/positional/positional.go goto main.go 42
//	go run samples/positional/positional.go goto main.go 42 10
func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "copy", Callback: copyFile, Description: "Copy a file from source to destination"},
			{Starter: "goto", Callback: gotoLine, Description: "Go to a specific line in a file"},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

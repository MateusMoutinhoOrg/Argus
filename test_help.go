package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

type ServeEntries struct {
	Help bool `type:"Flag" identifiers:"--help,help" default:"true"`
}

func serve(e ServeEntries) int {
	fmt.Printf("Help: %v\n", e.Help)
	return 0
}

func main() {
	argus := Argus.New(native.New())
	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "serve", Callback: serve},
		},
	}
	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

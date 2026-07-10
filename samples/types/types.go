package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// This sample demonstrates every supported scalar type:
//   - string
//   - int
//   - int64
//   - float64
//   - bool
//
// Both as flags and as positional arguments.

// TypesAsFlagsEntries shows all scalar types used as Flag entries.
type TypesAsFlagsEntries struct {
	Name   string  `type:"Flag" identifiers:"-n,--name" description:"person's name (string)"`
	Age    int     `type:"Flag" identifiers:"-a,--age" description:"person's age (int)"`
	ID     int64   `type:"Flag" identifiers:"--id" description:"unique identifier (int64)"`
	Score  float64 `type:"Flag" identifiers:"-s,--score" description:"numeric score (float64)"`
	Active bool    `type:"Flag" identifiers:"--active" description:"active status (bool)"`
}

func showFlags(e TypesAsFlagsEntries, deps argus_dep.Deps) int {
	deps.Print(strings.Repeat("─", 40) + "\n")
	deps.Print("  All scalar types via Flags:\n")
	deps.Print(strings.Repeat("─", 40) + "\n")
	deps.Print(fmt.Sprintf("  Name   (string):  %s\n", e.Name))
	deps.Print(fmt.Sprintf("  Age    (int):     %d\n", e.Age))
	deps.Print(fmt.Sprintf("  ID     (int64):   %d\n", e.ID))
	deps.Print(fmt.Sprintf("  Score  (float64): %.2f\n", e.Score))
	deps.Print(fmt.Sprintf("  Active (bool):    %v\n", e.Active))
	deps.Print(strings.Repeat("─", 40) + "\n")
	return 0
}

// TypesAsArgsEntries shows all non-bool scalar types used as positional NextArg.
type TypesAsArgsEntries struct {
	Label string  `type:"NextArg" description:"item label (string)"`
	Count int     `type:"NextArg" description:"quantity (int)"`
	Price float64 `type:"NextArg" description:"unit price (float64)"`
}

func showArgs(e TypesAsArgsEntries, deps argus_dep.Deps) int {
	deps.Print(strings.Repeat("─", 40) + "\n")
	deps.Print("  Scalar types via NextArg:\n")
	deps.Print(strings.Repeat("─", 40) + "\n")
	deps.Print(fmt.Sprintf("  Label (string):  %s\n", e.Label))
	deps.Print(fmt.Sprintf("  Count (int):     %d\n", e.Count))
	deps.Print(fmt.Sprintf("  Price (float64): %.2f\n", e.Price))
	deps.Print(strings.Repeat("─", 40) + "\n")
	return 0
}

// IntArrayEntries shows []int via ArrayArg.
type IntArrayEntries struct {
	Numbers []int `type:"ArrayArg" start:"0" end:"-1" min_size:"1" description:"list of integers to sum"`
}

func sumInts(e IntArrayEntries, deps argus_dep.Deps) int {
	total := 0
	for _, n := range e.Numbers {
		total += n
	}
	deps.Print(fmt.Sprintf("Numbers: %v\n", e.Numbers))
	deps.Print(fmt.Sprintf("Sum:     %d\n", total))
	return 0
}

// StringArrayFlagEntries shows []string via ArrayFlag.
type StringArrayFlagEntries struct {
	Hosts []string `type:"ArrayFlag" identifiers:"-H,--host" min_size:"1" description:"hosts to ping (can be repeated)"`
}

func ping(e StringArrayFlagEntries, deps argus_dep.Deps) int {
	for _, h := range e.Hosts {
		deps.Print(fmt.Sprintf("Pinging %s … ok\n", h))
	}
	return 0
}

// Usage:
//
//	go run samples/types/types.go flags -n Alice -a 30 --id 999999999 -s 97.5 --active
//	go run samples/types/types.go args widget 42 19.99
//	go run samples/types/types.go sum-ints 10 20 30 40
//	go run samples/types/types.go ping -H google.com -H github.com -H example.org
func main() {
	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{Starter: "flags", Callback: showFlags, Description: "Show all scalar types via flags", Samples: []string{"-n Alice -a 30 --id 999999999 -s 97.5 --active"}},
			{Starter: "args", Callback: showArgs, Description: "Show scalar types via positional arguments", Samples: []string{"widget 42 19.99"}},
			{Starter: "sum-ints", Callback: sumInts, Description: "Sum an array of integers", Samples: []string{"10 20 30 40"}},
			{Starter: "ping", Callback: ping, Description: "Ping a list of hosts", Samples: []string{"-H google.com -H github.com -H example.org"}},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

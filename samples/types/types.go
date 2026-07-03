package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

// This sample demonstrates every supported scalar type:
//   - string
//   - int
//   - int64
//   - float64
//   - bool
//
// Both as flags and as positional arguments.

// TypesAsFlagsFlags shows all scalar types used as Flag entries. Every
// non-slice field in a Flags sub-struct is inferred as a Flag.
type TypesAsFlagsFlags struct {
	Name   string  `identifiers:"-n,--name" description:"person's name (string)"`
	Age    int     `identifiers:"-a,--age" description:"person's age (int)"`
	ID     int64   `identifiers:"--id" description:"unique identifier (int64)"`
	Score  float64 `identifiers:"-s,--score" description:"numeric score (float64)"`
	Active bool    `identifiers:"--active" description:"active status (bool)"`
}

type TypesAsFlagsEntries struct {
	Flags TypesAsFlagsFlags
}

func showFlags(e TypesAsFlagsEntries) int {
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("  All scalar types via Flags:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Name   (string):  %s\n", e.Flags.Name)
	fmt.Printf("  Age    (int):     %d\n", e.Flags.Age)
	fmt.Printf("  ID     (int64):   %d\n", e.Flags.ID)
	fmt.Printf("  Score  (float64): %.2f\n", e.Flags.Score)
	fmt.Printf("  Active (bool):    %v\n", e.Flags.Active)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// TypesAsArgsArgs shows all non-bool scalar types used as positional
// NextArg. Every non-slice field in an Args sub-struct without a
// `position` tag is inferred as a NextArg.
type TypesAsArgsArgs struct {
	Label string  `description:"item label (string)"`
	Count int     `description:"quantity (int)"`
	Price float64 `description:"unit price (float64)"`
}

type TypesAsArgsEntries struct {
	Args TypesAsArgsArgs
}

func showArgs(e TypesAsArgsEntries) int {
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("  Scalar types via NextArg:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Label (string):  %s\n", e.Args.Label)
	fmt.Printf("  Count (int):     %d\n", e.Args.Count)
	fmt.Printf("  Price (float64): %.2f\n", e.Args.Price)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// IntArrayArgs shows []int via ArrayArg.
type IntArrayArgs struct {
	Numbers []int `start:"0" end:"-1" min_size:"1" description:"list of integers to sum"`
}

type IntArrayEntries struct {
	Args IntArrayArgs
}

func sumInts(e IntArrayEntries) int {
	total := 0
	for _, n := range e.Args.Numbers {
		total += n
	}
	fmt.Printf("Numbers: %v\n", e.Args.Numbers)
	fmt.Printf("Sum:     %d\n", total)
	return 0
}

// StringArrayFlagFlags shows []string via ArrayFlag.
type StringArrayFlagFlags struct {
	Hosts []string `identifiers:"-H,--host" min_size:"1" description:"hosts to ping (can be repeated)"`
}

type StringArrayFlagEntries struct {
	Flags StringArrayFlagFlags
}

func ping(e StringArrayFlagEntries) int {
	for _, h := range e.Flags.Hosts {
		fmt.Printf("Pinging %s … ok\n", h)
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
			{Starter: "flags", Callback: showFlags, Description: "Show all scalar types via flags"},
			{Starter: "args", Callback: showArgs, Description: "Show scalar types via positional arguments"},
			{Starter: "sum-ints", Callback: sumInts, Description: "Sum an array of integers"},
			{Starter: "ping", Callback: ping, Description: "Ping a list of hosts"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

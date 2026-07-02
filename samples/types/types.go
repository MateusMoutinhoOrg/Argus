package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
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
	Name    string  `type:"Flag" identifiers:"-n,--name"`
	Age     int     `type:"Flag" identifiers:"-a,--age"`
	ID      int64   `type:"Flag" identifiers:"--id"`
	Score   float64 `type:"Flag" identifiers:"-s,--score"`
	Active  bool    `type:"Flag" identifiers:"--active"`
}

func showFlags(e TypesAsFlagsEntries) int {
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("  All scalar types via Flags:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Name   (string):  %s\n", e.Name)
	fmt.Printf("  Age    (int):     %d\n", e.Age)
	fmt.Printf("  ID     (int64):   %d\n", e.ID)
	fmt.Printf("  Score  (float64): %.2f\n", e.Score)
	fmt.Printf("  Active (bool):    %v\n", e.Active)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// TypesAsArgsEntries shows all non-bool scalar types used as positional NextArg.
type TypesAsArgsEntries struct {
	Label  string  `type:"NextArg"`
	Count  int     `type:"NextArg"`
	Price  float64 `type:"NextArg"`
}

func showArgs(e TypesAsArgsEntries) int {
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("  Scalar types via NextArg:")
	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("  Label (string):  %s\n", e.Label)
	fmt.Printf("  Count (int):     %d\n", e.Count)
	fmt.Printf("  Price (float64): %.2f\n", e.Price)
	fmt.Println(strings.Repeat("─", 40))
	return 0
}

// IntArrayEntries shows []int via ArrayArg.
type IntArrayEntries struct {
	Numbers []int `type:"ArrayArg" start:"0" end:"-1" min_size:"1"`
}

func sumInts(e IntArrayEntries) int {
	total := 0
	for _, n := range e.Numbers {
		total += n
	}
	fmt.Printf("Numbers: %v\n", e.Numbers)
	fmt.Printf("Sum:     %d\n", total)
	return 0
}

// StringArrayFlagEntries shows []string via ArrayFlag.
type StringArrayFlagEntries struct {
	Hosts []string `type:"ArrayFlag" identifiers:"-H,--host" min_size:"1"`
}

func ping(e StringArrayFlagEntries) int {
	for _, h := range e.Hosts {
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

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "flags", Callback: showFlags, Description: "Show all scalar types via flags"},
			{Starter: "args", Callback: showArgs, Description: "Show scalar types via positional arguments"},
			{Starter: "sum-ints", Callback: sumInts, Description: "Sum an array of integers"},
			{Starter: "ping", Callback: ping, Description: "Ping a list of hosts"},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

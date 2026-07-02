package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

// ArrayArgEntries demonstrates ArrayArg — collecting a contiguous range
// of positional arguments into a slice.
//   - start:"0" end:"-1" means capture ALL positional args.
//   - min_size:"2" enforces at least 2 files.
type ArrayArgEntries struct {
	Files []string `type:"ArrayArg" start:"0" end:"-1" min_size:"2" description:"list of files to merge"`
}

func merge(e ArrayArgEntries) int {
	fmt.Printf("Merging %d files:\n", len(e.Files))
	for i, f := range e.Files {
		fmt.Printf("  [%d] %s\n", i+1, f)
	}
	return 0
}

// BoundedArrayEntries demonstrates a bounded ArrayArg window.
//   - start:"0" end:"2" captures only the first 2 positional args.
type BoundedArrayEntries struct {
	Pair []string `type:"ArrayArg" start:"0" end:"2" min_size:"2" max_size:"2" description:"two files to swap"`
}

func swap(e BoundedArrayEntries) int {
	fmt.Printf("Swapping: '%s' ↔ '%s'\n", e.Pair[0], e.Pair[1])
	return 0
}

// ArrayFlagEntries demonstrates ArrayFlag — a flag that can be repeated
// multiple times to build a slice.
//   - min_size:"1" requires at least one tag.
//   - max_size:"-1" means unbounded.
type ArrayFlagEntries struct {
	Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" min_size:"1" max_size:"-1" description:"labels to apply (can be repeated)"`
}

func tag(e ArrayFlagEntries) int {
	fmt.Printf("Tags applied (%d):\n", len(e.Tags))
	for _, t := range e.Tags {
		fmt.Printf("  • %s\n", t)
	}
	return 0
}

// ArrayFlagNumEntries demonstrates ArrayFlag with numeric types.
type ArrayFlagNumEntries struct {
	Scores []float64 `type:"ArrayFlag" identifiers:"-s,--score" min_size:"1" description:"numeric scores (can be repeated)"`
}

func average(e ArrayFlagNumEntries) int {
	sum := 0.0
	for _, s := range e.Scores {
		sum += s
	}
	avg := sum / float64(len(e.Scores))
	fmt.Printf("Scores: %v\n", e.Scores)
	fmt.Printf("Average: %.2f\n", avg)
	return 0
}

// Usage:
//
//	go run samples/arrays/arrays.go merge file1.txt file2.txt file3.txt
//	go run samples/arrays/arrays.go swap left.txt right.txt
//	go run samples/arrays/arrays.go tag -t bug -t urgent -t backend
//	go run samples/arrays/arrays.go average -s 9.5 -s 8.0 -s 7.2 -s 10.0
func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "merge", Callback: merge, Description: "Merge multiple files together"},
			{Starter: "swap", Callback: swap, Description: "Swap the contents of two files"},
			{Starter: "tag", Callback: tag, Description: "Apply tags"},
			{Starter: "average", Callback: average, Description: "Calculate the average of scores"},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

// ArrayArgArgs demonstrates ArrayArg — collecting a contiguous range
// of positional arguments into a slice. Any slice-typed field inside an
// Args sub-struct is inferred as ArrayArg.
//   - start:"0" end:"-1" means capture ALL positional args.
//   - min_size:"2" enforces at least 2 files.
type ArrayArgArgs struct {
	Files []string `start:"0" end:"-1" min_size:"2" description:"list of files to merge"`
}

type ArrayArgEntries struct {
	Args ArrayArgArgs
}

func merge(e ArrayArgEntries) int {
	fmt.Printf("Merging %d files:\n", len(e.Args.Files))
	for i, f := range e.Args.Files {
		fmt.Printf("  [%d] %s\n", i+1, f)
	}
	return 0
}

// BoundedArrayArgs demonstrates a bounded ArrayArg window.
//   - start:"0" end:"2" captures only the first 2 positional args.
type BoundedArrayArgs struct {
	Pair []string `start:"0" end:"2" min_size:"2" max_size:"2" description:"two files to swap"`
}

type BoundedArrayEntries struct {
	Args BoundedArrayArgs
}

func swap(e BoundedArrayEntries) int {
	fmt.Printf("Swapping: '%s' ↔ '%s'\n", e.Args.Pair[0], e.Args.Pair[1])
	return 0
}

// ArrayFlagFlags demonstrates ArrayFlag — a flag that can be repeated
// multiple times to build a slice. Any slice-typed field inside a Flags
// sub-struct is inferred as ArrayFlag.
//   - min_size:"1" requires at least one tag.
//   - max_size:"-1" means unbounded.
type ArrayFlagFlags struct {
	Tags []string `identifiers:"-t,--tag" min_size:"1" max_size:"-1" description:"labels to apply (can be repeated)"`
}

type ArrayFlagEntries struct {
	Flags ArrayFlagFlags
}

func tag(e ArrayFlagEntries) int {
	fmt.Printf("Tags applied (%d):\n", len(e.Flags.Tags))
	for _, t := range e.Flags.Tags {
		fmt.Printf("  • %s\n", t)
	}
	return 0
}

// ArrayFlagNumFlags demonstrates ArrayFlag with numeric types.
type ArrayFlagNumFlags struct {
	Scores []float64 `identifiers:"-s,--score" min_size:"1" description:"numeric scores (can be repeated)"`
}

type ArrayFlagNumEntries struct {
	Flags ArrayFlagNumFlags
}

func average(e ArrayFlagNumEntries) int {
	sum := 0.0
	for _, s := range e.Flags.Scores {
		sum += s
	}
	avg := sum / float64(len(e.Flags.Scores))
	fmt.Printf("Scores: %v\n", e.Flags.Scores)
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

	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{Starter: "merge", Callback: merge, Description: "Merge multiple files together"},
			{Starter: "swap", Callback: swap, Description: "Swap the contents of two files"},
			{Starter: "tag", Callback: tag, Description: "Apply tags"},
			{Starter: "average", Callback: average, Description: "Calculate the average of scores"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// ArrayArgEntries demonstrates ArrayArg — collecting a contiguous range
// of positional arguments into a slice.
//   - start:"0" end:"-1" means capture ALL positional args.
//   - min_size:"2" enforces at least 2 files.
type ArrayArgEntries struct {
	Files []string `type:"ArrayArg" start:"0" end:"-1" min_size:"2" description:"list of files to merge"`
}

func merge(e ArrayArgEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Merging %d files:\n", len(e.Files)))
	for i, f := range e.Files {
		deps.Print(fmt.Sprintf("  [%d] %s\n", i+1, f))
	}
	return 0
}

// BoundedArrayEntries demonstrates a bounded ArrayArg window.
//   - start:"0" end:"2" captures only the first 2 positional args.
type BoundedArrayEntries struct {
	Pair []string `type:"ArrayArg" start:"0" end:"2" min_size:"2" max_size:"2" description:"two files to swap"`
}

func swap(e BoundedArrayEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Swapping: '%s' ↔ '%s'\n", e.Pair[0], e.Pair[1]))
	return 0
}

// ArrayFlagEntries demonstrates ArrayFlag — a flag that can be repeated
// multiple times to build a slice.
//   - min_size:"1" requires at least one tag.
//   - max_size:"-1" means unbounded.
type ArrayFlagEntries struct {
	Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" min_size:"1" max_size:"-1" description:"labels to apply (can be repeated)"`
}

func tag(e ArrayFlagEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Tags applied (%d):\n", len(e.Tags)))
	for _, t := range e.Tags {
		deps.Print(fmt.Sprintf("  • %s\n", t))
	}
	return 0
}

// ArrayFlagNumEntries demonstrates ArrayFlag with numeric types.
type ArrayFlagNumEntries struct {
	Scores []float64 `type:"ArrayFlag" identifiers:"-s,--score" min_size:"1" description:"numeric scores (can be repeated)"`
}

func average(e ArrayFlagNumEntries, deps argus_dep.Deps) int {
	sum := 0.0
	for _, s := range e.Scores {
		sum += s
	}
	avg := sum / float64(len(e.Scores))
	deps.Print(fmt.Sprintf("Scores: %v\n", e.Scores))
	deps.Print(fmt.Sprintf("Average: %.2f\n", avg))
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
			{Starter: "merge", Callback: merge, Description: "Merge multiple files together", Samples: []string{"file1.txt file2.txt file3.txt"}},
			{Starter: "swap", Callback: swap, Description: "Swap the contents of two files", Samples: []string{"left.txt right.txt"}},
			{Starter: "tag", Callback: tag, Description: "Apply tags", Samples: []string{"-t bug -t urgent -t backend"}},
			{Starter: "average", Callback: average, Description: "Calculate the average of scores", Samples: []string{"-s 9.5 -s 8.0 -s 7.2 -s 10.0"}},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

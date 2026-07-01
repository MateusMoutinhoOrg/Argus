package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

// DeployEntries demonstrates combining positional args and flags
// in a single entries struct. Flags are extracted first, then the
// remaining tokens are treated as positional arguments.
type DeployEntries struct {
	// Positional args
	Service     string `type:"NextArg"`
	Environment string `type:"NextArg"`
	// Flags
	Replicas int    `type:"Flag" identifiers:"-r,--replicas" default:"1"`
	Image    string `type:"Flag" identifiers:"-i,--image"`
	DryRun   bool   `type:"Flag" identifiers:"--dry-run"`
	Force    bool   `type:"Flag" identifiers:"-f,--force"`
}

func deploy(e DeployEntries) int {
	fmt.Println(strings.Repeat("═", 45))
	fmt.Println("  DEPLOYMENT PLAN")
	fmt.Println(strings.Repeat("═", 45))
	fmt.Printf("  Service:     %s\n", e.Service)
	fmt.Printf("  Environment: %s\n", e.Environment)
	fmt.Printf("  Image:       %s\n", e.Image)
	fmt.Printf("  Replicas:    %d\n", e.Replicas)
	fmt.Printf("  Dry run:     %v\n", e.DryRun)
	fmt.Printf("  Force:       %v\n", e.Force)
	fmt.Println(strings.Repeat("═", 45))

	if e.DryRun {
		fmt.Println("  ⚠  Dry run — no changes applied.")
	} else {
		fmt.Println("  ✓  Deployment submitted!")
	}
	return 0
}

// Usage:
//
//	go run samples/mixed/mixed.go deploy api production --image api:v2.1 -r 3
//	go run samples/mixed/mixed.go deploy api staging --image api:latest --dry-run
//	go run samples/mixed/mixed.go deploy worker production --image worker:v1.0 --force -r 5
func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "deploy", Callback: deploy},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

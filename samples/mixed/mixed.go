package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

// DeployArgs demonstrates positional arguments in a Deploy command.
type DeployArgs struct {
	Service     string `description:"name of the service to deploy"`
	Environment string `description:"target environment (staging, production, etc)"`
}

// DeployFlags demonstrates flags alongside DeployArgs.
type DeployFlags struct {
	Replicas int    `identifiers:"-r,--replicas" default:"1" description:"number of replicas (default: 1)"`
	Image    string `identifiers:"-i,--image" description:"container image to deploy"`
	DryRun   bool   `identifiers:"--dry-run" description:"preview changes without applying them"`
	Force    bool   `identifiers:"-f,--force" description:"force deployment even if health checks fail"`
}

// DeployEntries demonstrates combining positional args and flags in a
// single callback struct, via the Args and Flags sub-structs. Flags are
// extracted first, then the remaining tokens are treated as positional
// arguments.
type DeployEntries struct {
	Args  DeployArgs
	Flags DeployFlags
}

func deploy(e DeployEntries) int {
	fmt.Println(strings.Repeat("═", 45))
	fmt.Println("  DEPLOYMENT PLAN")
	fmt.Println(strings.Repeat("═", 45))
	fmt.Printf("  Service:     %s\n", e.Args.Service)
	fmt.Printf("  Environment: %s\n", e.Args.Environment)
	fmt.Printf("  Image:       %s\n", e.Flags.Image)
	fmt.Printf("  Replicas:    %d\n", e.Flags.Replicas)
	fmt.Printf("  Dry run:     %v\n", e.Flags.DryRun)
	fmt.Printf("  Force:       %v\n", e.Flags.Force)
	fmt.Println(strings.Repeat("═", 45))

	if e.Flags.DryRun {
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

	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{Starter: "deploy", Callback: deploy, Description: "Deploy a service to an environment"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

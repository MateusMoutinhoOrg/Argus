package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	argus_dep "github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

// DeployEntries demonstrates combining positional args and flags
// in a single entries struct. Flags are extracted first, then the
// remaining tokens are treated as positional arguments.
type DeployEntries struct {
	// Positional args
	Service     string `type:"NextArg" description:"name of the service to deploy"`
	Environment string `type:"NextArg" description:"target environment (staging, production, etc)"`
	// Flags
	Replicas int    `type:"Flag" identifiers:"-r,--replicas" default:"1" description:"number of replicas (default: 1)"`
	Image    string `type:"Flag" identifiers:"-i,--image" description:"container image to deploy"`
	DryRun   bool   `type:"Flag" identifiers:"--dry-run" description:"preview changes without applying them"`
	Force    bool   `type:"Flag" identifiers:"-f,--force" description:"force deployment even if health checks fail"`
}

func deploy(e DeployEntries, deps argus_dep.Deps) int {
	deps.Print(strings.Repeat("═", 45) + "\n")
	deps.Print("  DEPLOYMENT PLAN\n")
	deps.Print(strings.Repeat("═", 45) + "\n")
	deps.Print(fmt.Sprintf("  Service:     %s\n", e.Service))
	deps.Print(fmt.Sprintf("  Environment: %s\n", e.Environment))
	deps.Print(fmt.Sprintf("  Image:       %s\n", e.Image))
	deps.Print(fmt.Sprintf("  Replicas:    %d\n", e.Replicas))
	deps.Print(fmt.Sprintf("  Dry run:     %v\n", e.DryRun))
	deps.Print(fmt.Sprintf("  Force:       %v\n", e.Force))
	deps.Print(strings.Repeat("═", 45) + "\n")

	if e.DryRun {
		deps.Print("  ⚠  Dry run — no changes applied.\n")
	} else {
		deps.Print("  ✓  Deployment submitted!\n")
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
			{Starter: "deploy", Callback: deploy, Description: "Deploy a service to an environment", Samples: []string{"api production --image api:v2.1 -r 3", "api staging --image api:latest --dry-run", "worker production --image worker:v1.0 --force -r 5"}},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

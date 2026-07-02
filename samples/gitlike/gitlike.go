package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/Argus"
)

// This sample simulates a real-world git-like CLI tool with multiple
// subcommands, each demonstrating a different combination of Argus
// features working together.

// --- init: no arguments at all (empty struct) ---

type InitEntries struct{}

func initRepo(e InitEntries) int {
	fmt.Println("Initialized empty repository in ./")
	return 0
}

// --- clone: NextArg + optional Flag ---

type CloneEntries struct {
	URL   string `type:"NextArg" description:"repository URL to clone"`
	Depth int    `type:"Flag" identifiers:"--depth" required:"false" description:"depth of the shallow clone"`
}

func clone(e CloneEntries) int {
	fmt.Printf("Cloning %s\n", e.URL)
	if e.Depth > 0 {
		fmt.Printf("  (shallow clone, depth=%d)\n", e.Depth)
	}
	return 0
}

// --- commit: all flags, some required, some optional ---

type CommitEntries struct {
	Message string `type:"Flag" identifiers:"-m,--message" description:"commit message"`
	Author  string `type:"Flag" identifiers:"--author" default:"current user" description:"commit author name"`
	Amend   bool   `type:"Flag" identifiers:"--amend" description:"amend the previous commit"`
}

func commit(e CommitEntries) int {
	action := "Created"
	if e.Amend {
		action = "Amended"
	}
	fmt.Printf("%s commit: \"%s\" (author: %s)\n", action, e.Message, e.Author)
	return 0
}

// --- add: ArrayArg + boolean flag ---

type AddEntries struct {
	Files   []string `type:"ArrayArg" start:"0" end:"-1" min_size:"1" description:"files to add to staging"`
	Verbose bool     `type:"Flag" identifiers:"-v,--verbose" description:"verbose output"`
}

func add(e AddEntries) int {
	for _, f := range e.Files {
		if e.Verbose {
			fmt.Printf("  staging: %s\n", f)
		}
	}
	fmt.Printf("Added %d file(s) to staging area.\n", len(e.Files))
	return 0
}

// --- remote: Arg (fixed position) + ArrayFlag ---

type RemoteEntries struct {
	Action string   `type:"Arg" position:"0" description:"remote action (list, add, etc)"`
	Names  []string `type:"ArrayFlag" identifiers:"-n,--name" required:"false" description:"remote repository names"`
}

func remote(e RemoteEntries) int {
	switch e.Action {
	case "list":
		fmt.Println("Remote repositories:")
		names := []string{"origin", "upstream"}
		for _, n := range names {
			fmt.Printf("  • %s\n", n)
		}
	case "add":
		if len(e.Names) == 0 {
			fmt.Println("Error: --name required for 'add'")
			return 1
		}
		for _, n := range e.Names {
			fmt.Printf("Added remote: %s\n", n)
		}
	default:
		fmt.Printf("Unknown remote action: %s\n", e.Action)
		return 1
	}
	return 0
}

// --- log: all optional flags with defaults ---

type LogEntries struct {
	Count  int    `type:"Flag" identifiers:"-n,--count" default:"10" description:"number of commits to show"`
	Format string `type:"Flag" identifiers:"--format" default:"short" description:"output format (short, full, oneline)"`
	All    bool   `type:"Flag" identifiers:"--all" description:"show commits from all branches"`
}

func logCmd(e LogEntries) int {
	scope := "current branch"
	if e.All {
		scope = "all branches"
	}
	fmt.Printf("Showing last %d commits (%s, format=%s)\n", e.Count, scope, e.Format)
	fmt.Println(strings.Repeat("─", 50))
	for i := 1; i <= e.Count && i <= 3; i++ {
		fmt.Printf("  %da1b2c3  feat: example commit #%d\n", i, i)
	}
	if e.Count > 3 {
		fmt.Printf("  ... and %d more\n", e.Count-3)
	}
	return 0
}

// Usage:
//
//	go run samples/gitlike/gitlike.go init
//	go run samples/gitlike/gitlike.go clone https://github.com/user/repo.git
//	go run samples/gitlike/gitlike.go clone https://github.com/user/repo.git --depth 1
//	go run samples/gitlike/gitlike.go commit -m "initial commit"
//	go run samples/gitlike/gitlike.go commit -m "fix typo" --amend --author "Alice"
//	go run samples/gitlike/gitlike.go add main.go utils.go README.md -v
//	go run samples/gitlike/gitlike.go remote list
//	go run samples/gitlike/gitlike.go remote add -n upstream -n mirror
//	go run samples/gitlike/gitlike.go log
//	go run samples/gitlike/gitlike.go log -n 5 --format oneline --all
func main() {

	argus := Argus.New(native.New())

	props := Argus.GenerationProps{
		Callbacks: []Argus.Callback{
			{Starter: "init", Callback: initRepo, Description: "Initialize an empty repository"},
			{Starter: "clone", Callback: clone, Description: "Clone a repository"},
			{Starter: "commit", Callback: commit, Description: "Record changes to the repository"},
			{Starter: "add", Callback: add, Description: "Add file contents to the index"},
			{Starter: "remote", Callback: remote, Description: "Manage set of tracked repositories"},
			{Starter: "log", Callback: logCmd, Description: "Show commit logs"},
		},
	}

	exitCode, err := argus.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

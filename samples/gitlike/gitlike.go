package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
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

type CloneArgs struct {
	URL string `description:"repository URL to clone"`
}

type CloneFlags struct {
	Depth int `identifiers:"--depth" required:"false" description:"depth of the shallow clone"`
}

type CloneEntries struct {
	Args  CloneArgs
	Flags CloneFlags
}

func clone(e CloneEntries) int {
	fmt.Printf("Cloning %s\n", e.Args.URL)
	if e.Flags.Depth > 0 {
		fmt.Printf("  (shallow clone, depth=%d)\n", e.Flags.Depth)
	}
	return 0
}

// --- commit: all flags, some required, some optional ---
//
// This command also demonstrates dependency injection: declaring an
// unexported field of type deps.Deps causes Argus to populate it
// automatically, giving the callback access to Print/Args without
// importing fmt/os directly.

type CommitFlags struct {
	Message string `identifiers:"-m,--message" description:"commit message"`
	Author  string `identifiers:"--author" default:"current user" description:"commit author name"`
	Amend   bool   `identifiers:"--amend" description:"amend the previous commit"`
}

type CommitEntries struct {
	Flags CommitFlags
	deps  deps.Deps
}

func commit(e CommitEntries) int {
	action := "Created"
	if e.Flags.Amend {
		action = "Amended"
	}
	e.deps.Print(fmt.Sprintf("%s commit: \"%s\" (author: %s)", action, e.Flags.Message, e.Flags.Author))
	return 0
}

// --- add: ArrayArg + boolean flag ---

type AddArgs struct {
	Files []string `start:"0" end:"-1" min_size:"1" description:"files to add to staging"`
}

type AddFlags struct {
	Verbose bool `identifiers:"-v,--verbose" description:"verbose output"`
}

type AddEntries struct {
	Args  AddArgs
	Flags AddFlags
}

func add(e AddEntries) int {
	for _, f := range e.Args.Files {
		if e.Flags.Verbose {
			fmt.Printf("  staging: %s\n", f)
		}
	}
	fmt.Printf("Added %d file(s) to staging area.\n", len(e.Args.Files))
	return 0
}

// --- remote: Arg (fixed position) + ArrayFlag ---

type RemoteArgs struct {
	Action string `position:"0" description:"remote action (list, add, etc)"`
}

type RemoteFlags struct {
	Names []string `identifiers:"-n,--name" required:"false" description:"remote repository names"`
}

type RemoteEntries struct {
	Args  RemoteArgs
	Flags RemoteFlags
}

func remote(e RemoteEntries) int {
	switch e.Args.Action {
	case "list":
		fmt.Println("Remote repositories:")
		names := []string{"origin", "upstream"}
		for _, n := range names {
			fmt.Printf("  • %s\n", n)
		}
	case "add":
		if len(e.Flags.Names) == 0 {
			fmt.Println("Error: --name required for 'add'")
			return 1
		}
		for _, n := range e.Flags.Names {
			fmt.Printf("Added remote: %s\n", n)
		}
	default:
		fmt.Printf("Unknown remote action: %s\n", e.Args.Action)
		return 1
	}
	return 0
}

// --- log: all optional flags with defaults ---

type LogFlags struct {
	Count  int    `identifiers:"-n,--count" default:"10" description:"number of commits to show"`
	Format string `identifiers:"--format" default:"short" description:"output format (short, full, oneline)"`
	All    bool   `identifiers:"--all" description:"show commits from all branches"`
}

type LogEntries struct {
	Flags LogFlags
}

func logCmd(e LogEntries) int {
	scope := "current branch"
	if e.Flags.All {
		scope = "all branches"
	}
	fmt.Printf("Showing last %d commits (%s, format=%s)\n", e.Flags.Count, scope, e.Flags.Format)
	fmt.Println(strings.Repeat("─", 50))
	for i := 1; i <= e.Flags.Count && i <= 3; i++ {
		fmt.Printf("  %da1b2c3  feat: example commit #%d\n", i, i)
	}
	if e.Flags.Count > 3 {
		fmt.Printf("  ... and %d more\n", e.Flags.Count-3)
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

	a := argus.New(native.New())

	props := argus.GenerationProps{
		Quiet: true,
		Callbacks: []argus.Callback{
			{Starter: "init", Callback: initRepo, Description: "Initialize an empty repository"},
			{Starter: "clone", Callback: clone, Description: "Clone a repository"},
			{Starter: "commit", Callback: commit, Description: "Record changes to the repository"},
			{Starter: "add", Callback: add, Description: "Add file contents to the index"},
			{Starter: "remote", Callback: remote, Description: "Manage set of tracked repositories"},
			{Starter: "log", Callback: logCmd, Description: "Show commit logs"},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("config error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}

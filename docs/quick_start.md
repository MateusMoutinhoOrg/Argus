# Quick Start

Get up and running with Argus in 5 minutes.

## Installation

Add Argus to your Go project:

```bash
go get github.com/MateusMoutinhoOrg/Argus@v0.0.3
```

## Your First CLI Application

Here's a minimal example that creates a `greet` command:

```go
package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

type GreetArgs struct {
	Name string `description:"The person's name"`
}

type GreetEntries struct {
	Args GreetArgs
}

func greet(e GreetEntries) int {
	fmt.Printf("Hello, %s!\n", e.Args.Name)
	return 0
}

func main() {
	a := argus.New(native.New())
	
	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{
				Starter:     "greet",
				Callback:    greet,
				Description: "Greet a person by name",
			},
		},
	}
	
	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
```

Run it:

```bash
go run main.go greet Alice
# Output: Hello, Alice!
```

## Key Concepts

### 1. **Entries Struct**

Each command receives an **entries struct**. It's built out of two optional
sub-structs — `Flags` for named flags, `Args` for positional arguments — plus
an optional injected `deps.Deps` field:

```go
type ServeFlags struct {
	Host string `identifiers:"-h,--host"`
	Port int    `identifiers:"-p,--port" default:"8080"`
	TLS  bool   `identifiers:"--tls"`
}

type ServeEntries struct {
	Flags ServeFlags
}
```

- `Flags ServeFlags` — a struct field named `Flags` holds all named flags.
- `identifiers:"-h,--host"` — short and long names for the flag.
- `default:"8080"` — optional; if not provided, use this value.

There is **no `type` tag** — Argus infers `Flag` vs `ArrayFlag` (in `Flags`) and
`Arg` vs `NextArg` vs `ArrayArg` (in `Args`) automatically. See *Entry Types* below.

### 2. **Struct Tags**

Tags tell Argus how to extract values from the command line:

| Tag | Purpose |
|-----|---------|
| `identifiers` | Flag names, e.g. `-p,--port` — presence of this tag (inside `Flags`) is what makes a field a flag |
| `position` | For fixed positional args inside `Args` (e.g., `position:"0"`); presence makes a field an `Arg` instead of `NextArg` |
| `required` | Is this field mandatory? Defaults to `"true"` |
| `default` | Fallback value if missing; implies optional |
| `help` | **Deprecated.** Use `description` instead |
| `description` | Description shown in help text and errors |

### 3. **Entry Types**

Argus infers the entry type from **where** a field lives and **what shape** it has:

| Kind        | Lives in | Inferred when                          |
|-------------|----------|------------------------------------------|
| `Flag`      | `Flags`  | non-slice field                          |
| `ArrayFlag` | `Flags`  | slice field                              |
| `NextArg`   | `Args`   | non-slice field, no `position` tag       |
| `Arg`       | `Args`   | non-slice field, has a `position` tag    |
| `ArrayArg`  | `Args`   | slice field                              |

#### Flags (Named Arguments)

```go
type BuildFlags struct {
	// Regular flag with a value
	Output string `identifiers:"-o,--output" default:"a.out"`
	
	// Boolean flag (presence-only)
	Verbose bool `identifiers:"-v,--verbose"`
}

type BuildEntries struct {
	Flags BuildFlags
}

func build(e BuildEntries) int {
	fmt.Printf("Building to %s (verbose=%v)\n", e.Flags.Output, e.Flags.Verbose)
	return 0
}
```

Usage:

```bash
go run main.go build -o bin/app --verbose
```

#### Positional Arguments

```go
type AddArgs struct {
	A float64 `description:"First number"`
	B float64 `description:"Second number"`
}

type AddEntries struct {
	Args AddArgs
}

func add(e AddEntries) int {
	fmt.Printf("%.1f + %.1f = %.1f\n", e.Args.A, e.Args.B, e.Args.A+e.Args.B)
	return 0
}
```

Usage:

```bash
go run main.go add 10 20
# Output: 10.0 + 20.0 = 30.0
```

#### Array Arguments

Repeated flags or multiple positional arguments, both inferred from a **slice type**:

```go
type CollectFlags struct {
	// Repeat the flag multiple times
	Tags []string `identifiers:"-t,--tag"`
}
type CollectArgs struct {
	// Collect multiple positional args
	Files []string `start:"0" end:"-1"`
}
type CollectEntries struct {
	Flags CollectFlags
	Args  CollectArgs
}
```

Usage:

```bash
go run main.go collect -t bug -t feature file1.txt file2.txt
```

## Complete Example

Here's a more realistic CLI with multiple commands:

```go
package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

type ServeFlags struct {
	Host     string `identifiers:"-h,--host" required:"false" default:"localhost"`
	Port     int    `identifiers:"-p,--port" default:"8080"`
	TLS      bool   `identifiers:"--tls"`
}

type ServeEntries struct {
	Flags ServeFlags
}

func serve(e ServeEntries) int {
	scheme := "http"
	if e.Flags.TLS {
		scheme = "https"
	}
	fmt.Printf("Server running on %s://%s:%d\n", scheme, e.Flags.Host, e.Flags.Port)
	return 0
}

type VersionEntries struct{}

func version(e VersionEntries) int {
	fmt.Println("v1.0.0")
	return 0
}

func main() {
	a := argus.New(native.New())
	
	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{
				Starter:     "serve",
				Callback:    serve,
				Description: "Start the server",
			},
			{
				Starter:     "version",
				Callback:    version,
				Description: "Show version",
			},
		},
	}
	
	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
```

## Next Steps

- **See samples** — Check `samples/` in the repository for real examples
- **Learn entries** — Read [docs/entries.md](entries.md) for the complete API
- **Custom messages** — Localize or customize error messages with [docs/msg_format.md](msg_format.md)
- **Testing with deps** — Mock CLI input for testing with [docs/deps.md](deps.md)

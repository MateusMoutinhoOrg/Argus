# Quick Start

Get up and running with Argus in 5 minutes.

## Installation

Add Argus to your Go project:

```bash
go get github.com/MateusMoutinhoOrg/Argus@v0.0.1
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

type GreetEntries struct {
	Name string `type:"NextArg" help:"The person's name"`
}

func greet(e GreetEntries) int {
	fmt.Printf("Hello, %s!\n", e.Name)
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

Each command receives an **entries struct** that defines which command-line arguments it accepts:

```go
type ServeEntries struct {
	Host string `type:"Flag" identifiers:"-h,--host"`
	Port int    `type:"Flag" identifiers:"-p,--port" default:"8080"`
	TLS  bool   `type:"Flag" identifiers:"--tls"`
}
```

- `type:"Flag"` — This is a named flag (like `--host`).
- `identifiers:"-h,--host"` — Short and long names for the flag.
- `default:"8080"` — Optional; if not provided, use this value.

### 2. **Struct Tags**

Tags tell Argus how to extract values from the command line:

| Tag | Purpose |
|-----|---------|
| `type` | How to read this field: `Flag`, `Arg`, `NextArg`, `ArrayFlag`, `ArrayArg` |
| `identifiers` | Flag names, e.g. `-p,--port` |
| `position` | For fixed positional args (e.g., `position:"0"`) |
| `required` | Is this field mandatory? Defaults to `"true"` |
| `default` | Fallback value if missing; implies optional |
| `help` | Description shown in help text |

### 3. **Entry Types**

#### Flags (Named Arguments)

```go
type BuildEntries struct {
	// Regular flag with a value
	Output string `type:"Flag" identifiers:"-o,--output" default:"a.out"`
	
	// Boolean flag (presence-only)
	Verbose bool `type:"Flag" identifiers:"-v,--verbose"`
}

func build(e BuildEntries) int {
	fmt.Printf("Building to %s (verbose=%v)\n", e.Output, e.Verbose)
	return 0
}
```

Usage:

```bash
go run main.go build -o bin/app --verbose
```

#### Positional Arguments

```go
type AddEntries struct {
	A float64 `type:"NextArg" help:"First number"`
	B float64 `type:"NextArg" help:"Second number"`
}

func add(e AddEntries) int {
	fmt.Printf("%.1f + %.1f = %.1f\n", e.A, e.B, e.A+e.B)
	return 0
}
```

Usage:

```bash
go run main.go add 10 20
# Output: 10.0 + 20.0 = 30.0
```

#### Array Arguments

Repeated flags or multiple positional arguments:

```go
type CollectEntries struct {
	// Repeat the flag multiple times
	Tags []string `type:"ArrayFlag" identifiers:"-t,--tag"`
	
	// Collect multiple positional args
	Files []string `type:"ArrayArg" start:"0" end:"-1"`
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

type ServeEntries struct {
	Host     string `type:"Flag" identifiers:"-h,--host" required:"false" default:"localhost"`
	Port     int    `type:"Flag" identifiers:"-p,--port" default:"8080"`
	TLS      bool   `type:"Flag" identifiers:"--tls"`
}

func serve(e ServeEntries) int {
	scheme := "http"
	if e.TLS {
		scheme = "https"
	}
	fmt.Printf("Server running on %s://%s:%d\n", scheme, e.Host, e.Port)
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

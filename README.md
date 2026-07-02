# Argus

A powerful, reflection-based CLI argument parser library for Go. Build structured, type-safe command-line applications with minimal boilerplate.

![Go Version](https://img.shields.io/badge/go-1.25+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## Features

- **Struct-based binding** — Use Go struct fields and tags to declaratively define CLI arguments, flags, and positional parameters
- **Type safety** — Automatic parsing and validation of int, float64, bool, string, and slice types
- **Reflection-powered** — No code generation; struct tags do all the work
- **Flexible argument handling** — Support for named flags, positional arguments, optional parameters, defaults, and array types
- **Auto-generated help** — Leverage field descriptions to build informative help text
- **Localization & customization** — Override all user-facing messages for localization or custom branding
- **Dependency injection** — Mock CLI input and output for easy testing without `os.Args`
- **Clean error handling** — Structured validation errors that guide users to correct usage
## Quick Installation

Add Argus to your Go project:

```bash
go get github.com/MateusMoutinhoOrg/Argus@v0.0.1
```

## Quick Example

```go
package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

type ServeEntries struct {
	Host string `type:"Flag" identifiers:"-h,--host" default:"localhost"`
	Port int    `type:"Flag" identifiers:"-p,--port" default:"8080"`
	TLS  bool   `type:"Flag" identifiers:"--tls"`
}

func serve(e ServeEntries) int {
	scheme := "http"
	if e.TLS {
		scheme = "https"
	}
	fmt.Printf("Listening on %s://%s:%d\n", scheme, e.Host, e.Port)
	return 0
}

func main() {
	a := argus.New(native.New())

	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{
				Starter:     "serve",
				Callback:    serve,
				Description: "Start the application server",
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

Usage:

```bash
go run main.go serve --host 0.0.0.0 -p 3000 --tls
# Output: Listening on https://0.0.0.0:3000
```

## Table of Contents

### Getting Started
- [Quick Start](docs/quick_start.md) — Install Argus and build your first CLI in 5 minutes
- [Flags and Arguments Guide](docs/flags_and_args.md) — Named flags, positional arguments, and array types with descriptions
- [API Reference — Entries](docs/entries.md) — Complete guide to struct tags and entry types

### Core Concepts
- [Dependency Injection & Testing](docs/deps.md) — Understand how to test your CLI without `os.Args`, mock input, and capture output
- [Custom Messages & Localization](docs/msg_format.md) — Customize all error messages and user-facing text for your language or brand
- [Glossary & Troubleshooting](docs/glossary.md) — Common issues, solutions, best practices, and quick reference

### Examples

Explore real working examples in the `samples/` directory. Each sample demonstrates flags and positional arguments with descriptions for user-friendly CLIs.

- **[flags/](samples/flags)** — Named flags with defaults and boolean presence flags
- **[positional/](samples/positional)** — Positional argument handling with `Arg` and `NextArg`
- **[arrays/](samples/arrays)** — Array arguments with `ArrayFlag` and `ArrayArg`
- **[mixed/](samples/mixed)** — Combining flags and positional arguments in one command
- **[gitlike/](samples/gitlike)** — Multi-command pattern (subcommands like `git commit`)
- **[types/](samples/types)** — Type conversions (int, float64, bool, string)
- **[custom_errors/](samples/custom_errors)** — Localized error messages (Portuguese example)

See [samples/README.md](samples/README.md) for detailed walkthroughs of each example.

Run any sample:

```bash
go run samples/flags/flags.go serve --host 0.0.0.0 -p 9090
go run samples/positional/positional.go copy src.txt dst.txt
go run samples/custom_errors/custom_errors.go greet "Your Name"
```

## Installation

```bash
go get github.com/MateusMoutinhoOrg/Argus@v0.0.1
```

## Architecture Overview

### Core Components

- **`pkg/Argus/`** — Core parsing engine using Go reflection
  - `handle.go` — Main CLI parsing logic; struct tag inspection, flag/positional extraction, help generation
  - `new.go` — Factory for creating Argus instances
  - `errors.go` — Error message templates (being refactored to `messages.go`)

- **`pkg/deps/`** — Dependency injection layer
  - `deps.go` — `Deps` interface with `Args` and `Print` for testable CLI interactions

- **`adapters/native/`** — OS integration
  - Reads `os.Args[1:]` and prints to stdout

### Parsing Flow

1. Inspect callback function parameter struct; read struct tags
2. Extract named flags from CLI arguments (first pass)
3. Populate positional arguments from remaining arguments (second pass)
4. Validate required fields, apply defaults, check constraints
5. Invoke callback with populated struct; capture exit code

## Common Use Cases

### 1. Simple Command with Flags

```go
type DeployEntries struct {
	Env    string `type:"Flag" identifiers:"-e,--env" default:"staging" 
	                 description:"deployment environment (default: staging)"`
	Force  bool   `type:"Flag" identifiers:"-f,--force"
	               description:"force deployment without confirmation"`
}

func deploy(e DeployEntries) int {
	fmt.Printf("Deploying to %s (force=%v)\n", e.Env, e.Force)
	return 0
}
```

### 2. Multi-Command Application

```go
props := argus.GenerationProps{
	Callbacks: []argus.Callback{
		{Starter: "serve", Callback: serve, Description: "Start server"},
		{Starter: "build", Callback: build, Description: "Build project"},
		{Starter: "test", Callback: test, Description: "Run tests"},
	},
}
```

### 3. Positional Arguments

```go
type CopyEntries struct {
	Src   string `type:"NextArg" description:"source file path"`
	Dst   string `type:"NextArg" description:"destination file path"`
	Force bool   `type:"Flag" identifiers:"-f,--force" 
	               description:"overwrite destination without prompting"`
}

func copy(e CopyEntries) int {
	fmt.Printf("Copying %s → %s\n", e.Src, e.Dst)
	return 0
}
```

### 4. Testing with Dependency Injection

```go
import (
	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

func TestServe(t *testing.T) {
	testDeps := deps.Deps{
		Args: []string{"serve", "--port", "9090"},
		Print: func(s string) { /* capture output */ },
	}
	
	a := argus.New(&testDeps)
	exitCode, _ := a.HandleCli(props)
	if exitCode != 0 {
		t.Fatal("serve command failed")
	}
}
```

## Describing Flags and Arguments

Each field can have a `description` tag that documents its purpose. This description appears in auto-generated help output and error messages.

```go
type ServeEntries struct {
    Host     string `type:"Flag" identifiers:"--host,-h" 
                      description:"hostname or IP to bind to"`
    Port     int    `type:"Flag" identifiers:"--port,-p" default:"8080"
                      description:"port number (default: 8080)"`
    TLS      bool   `type:"Flag" identifiers:"--tls"
                      description:"enable HTTPS"`
    Files    []string `type:"ArrayArg" start:"0" end:"-1" min_size:"1"
                        description:"input files (at least one required)"`
}
```

Descriptions guide users when they run `--help` or encounter errors. Keep descriptions concise, mention defaults when non-obvious, and specify constraints for array types.

See [Flags and Arguments Guide](docs/flags_and_args.md) for patterns and best practices.

## Tag System Reference

Each struct field declares how it's populated via tags:

| Tag | Values | Example |
|-----|--------|---------|
| `type` | `Flag`, `Arg`, `NextArg`, `ArrayFlag`, `ArrayArg` | `type:"Flag"` |
| `identifiers` | Comma-separated flag aliases | `identifiers:"-p,--port"` |
| `position` | Index for `Arg` type | `position:"0"` |
| `required` | `"true"` or `"false"` | `required:"false"` |
| `default` | Default value as string | `default:"8080"` |
| `start`, `end` | Array bounds (for `ArrayArg`) | `start:"0" end:"-1"` |
| `min_size`, `max_size` | Array size constraints | `min_size:"1" max_size:"10"` |
| `description` | Description for help and errors | `description:"port to listen on"` |

See [docs/entries.md](docs/entries.md) for complete details.

## Supported Types

- **Scalars** — `string`, `int`, `int64`, `float64`, `bool`
- **Slices** — `[]string`, `[]int`, `[]int64`, `[]float64` (for `ArrayFlag` and `ArrayArg`)

## Error Handling

Argus validates arguments and produces user-friendly errors:

- Missing required flag → usage error with flag description
- Unparseable value (e.g., non-numeric for `int`) → usage error with type info
- Missing required positional argument → usage error with position/description
- Array size constraints violated → usage error with bounds

All error messages are customizable via the `Messages` struct. See [docs/msg_format.md](docs/msg_format.md).

## Testing

Since Argus uses dependency injection, testing is straightforward:

```go
func runCLI(args []string) (int, string) {
	var output strings.Builder
	testDeps := deps.Deps{
		Args: args,
		Print: func(s string) { output.WriteString(s) },
	}
	a := argus.New(&testDeps)
	exitCode, _ := a.HandleCli(props)
	return exitCode, output.String()
}

// In tests:
exitCode, output := runCLI([]string{"serve", "--port", "9090"})
if exitCode != 0 {
	t.Errorf("serve failed: %s", output)
}
```

See [docs/deps.md](docs/deps.md) for comprehensive testing patterns.

## Design Philosophy

- **Reflection over codegen** — Struct tags + reflection = zero build-time overhead
- **Explicit over implicit** — Tags make intent clear; no magic defaults
- **Testing first** — Dependency injection built in from the start
- **Validation upfront** — Errors surface at configuration time, not parse time
- **Open-ended by default** — Array arguments are unbounded; constrain with `min_size`/`max_size`

## Known Limitations & Future Work

- **No automated tests** — Consider adding test suite once API stabilizes
- **Single-level commands** — Currently one level of nesting; samples show patterns for deeper trees
- **Flag syntax** — Currently only `--flag value` (space-separated); `--flag=value` not yet supported
- **Negative numbers** — `-5` in context of int field may be ambiguous with flag syntax

See [docs/entries.md](docs/entries.md) for open questions and design decisions.

## License

MIT

## Contributing

Contributions welcome! Please ensure:
- Struct tags are well-documented
- Examples in `samples/` demonstrate each feature
- Docs are updated for new capabilities
- Tests cover the happy path and error cases

## FAQ

**Q: Do I need to export struct fields?**  
A: Yes, Argus uses reflection and can only access exported fields. The examples use capitalized field names.

**Q: Can I combine multiple entry types in one struct?**  
A: Yes! Flags are extracted first, then positional arguments fill the remaining tokens.

**Q: How do I handle subcommands deeper than one level?**  
A: Argus handles single-level commands natively. For deeper nesting, write a wrapper command that parses the next level manually. See `samples/gitlike/` for patterns.

**Q: Can I customize the help text format?**  
A: Not directly via the API yet, but you can override the `Messages.Help*` functions to control output (see [docs/msg_format.md](docs/msg_format.md)).

---

Start with the [Quick Start guide](docs/quick_start.md) or explore the [samples](samples/).

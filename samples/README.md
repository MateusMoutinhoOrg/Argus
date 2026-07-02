# Argus Samples — Flags and Arguments in Practice

This directory contains working examples of Argus CLI applications, demonstrating how to use flags, positional arguments, and the `description` tag to build user-friendly command-line tools.

## Overview

Each sample showcases different argument patterns and how to add descriptions that appear in help output and error messages.

---

## Samples

### [flags/](flags) — Named Flags

Demonstrates flags with various configurations:

- **Required flags** — Must be provided by the user
- **Optional flags with defaults** — Fallback when omitted
- **Boolean presence flags** — No value consumed, just presence/absence
- **Descriptions** — Each flag documents its purpose

```go
type ServeEntries struct {
    Host     string `type:"Flag" identifiers:"-h,--host" 
                      description:"the host address to bind to"`
    Port     int    `type:"Flag" identifiers:"-p,--port" default:"8080"
                      description:"the port number to listen on (default: 8080)"`
    TLS      bool   `type:"Flag" identifiers:"--tls"
                      description:"enable TLS/HTTPS"`
}
```

**Run it:**
```bash
go run samples/flags/flags.go serve --host 0.0.0.0 -p 9090 --tls
go run samples/flags/flags.go status --verbose --format json
```

---

### [positional/](positional) — Positional Arguments

Demonstrates two types of positional argument handling:

#### NextArg — Sequential Consumption
Fields are filled in declaration order from remaining arguments.

```go
type NextArgEntries struct {
    Src  string `type:"NextArg" description:"source file path"`
    Dest string `type:"NextArg" description:"destination file path"`
}
```

#### Arg — Fixed Position
Fields bind to a specific positional index.

```go
type FixedArgEntries struct {
    Filename string `type:"Arg" position:"0" 
                      description:"path to the file to open"`
    LineNum  int    `type:"Arg" position:"1"
                      description:"line number to navigate to"`
}
```

**Run it:**
```bash
go run samples/positional/positional.go copy readme.md /tmp/backup.md
go run samples/positional/positional.go goto main.go 42 10
```

---

### [arrays/](arrays) — Array Arguments

Demonstrates handling multiple values as arrays:

#### ArrayFlag — Repeated Flag
A flag that appears multiple times, each adding an element.

```go
type ArrayFlagEntries struct {
    Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" min_size:"1"
                    description:"labels to apply (can be repeated)"`
}
```

#### ArrayArg — Range of Positionals
Collects positional arguments into a slice with optional bounds.

```go
type ArrayArgEntries struct {
    Files []string `type:"ArrayArg" start:"0" end:"-1" min_size:"2"
                     description:"list of files to merge"`
}
```

**Run it:**
```bash
go run samples/arrays/arrays.go merge file1.txt file2.txt file3.txt
go run samples/arrays/arrays.go tag -t bug -t urgent -t backend
go run samples/arrays/arrays.go average -s 9.5 -s 8.0 -s 7.2
```

---

### [mixed/](mixed) — Combining Flags and Positional Args

Shows how to mix flags and positional arguments in a single command. **Flags are extracted first**, then remaining tokens are treated as positional.

```go
type DeployEntries struct {
    Service     string `type:"NextArg" description:"name of the service to deploy"`
    Environment string `type:"NextArg" description:"target environment"`
    Image       string `type:"Flag" identifiers:"-i,--image"
                         description:"container image to deploy"`
    Replicas    int    `type:"Flag" identifiers:"-r,--replicas" default:"1"
                         description:"number of replicas (default: 1)"`
}
```

**Run it:**
```bash
go run samples/mixed/mixed.go deploy api production --image api:v2.1 -r 3
go run samples/mixed/mixed.go deploy worker staging --image worker:latest --dry-run
```

---

### [gitlike/](gitlike) — Multi-Command Application

Demonstrates a real-world pattern with multiple subcommands, each with different argument styles.

- `init` — No arguments
- `clone` — Positional URL + optional depth flag
- `commit` — Required flag + optional flags with defaults
- `add` — Array positional args + boolean flag
- `remote` — Fixed positional + array flag
- `log` — All optional flags

Each command showcases different combinations and includes descriptions.

**Run it:**
```bash
go run samples/gitlike/gitlike.go init
go run samples/gitlike/gitlike.go clone https://github.com/user/repo --depth 1
go run samples/gitlike/gitlike.go commit -m "fix bug" --amend
go run samples/gitlike/gitlike.go add main.go utils.go -v
go run samples/gitlike/gitlike.go log -n 5 --format oneline --all
```

---

### [types/](types) — Type Conversions

Demonstrates Argus's built-in type support:

- `string` — Accepted as-is
- `int`, `int64` — Parsed from decimal strings
- `float64` — Parsed from decimal strings  
- `bool` — Presence flags or parsed from strings
- Array types — Slices of any scalar

Each type is shown both as a flag and as a positional argument, with descriptions indicating the expected type.

```go
type TypesAsFlagsEntries struct {
    Name   string  `type:"Flag" identifiers:"-n,--name" 
                     description:"person's name (string)"`
    Age    int     `type:"Flag" identifiers:"-a,--age"
                     description:"person's age (int)"`
    Score  float64 `type:"Flag" identifiers:"-s,--score"
                     description:"numeric score (float64)"`
}
```

**Run it:**
```bash
go run samples/types/types.go flags -n Alice -a 30 -s 97.5 --active
go run samples/types/types.go args widget 42 19.99
go run samples/types/types.go sum-ints 10 20 30 40
go run samples/types/types.go ping -H google.com -H github.com
```

---

### [custom_errors/](custom_errors) — Localized Messages

Shows how to customize error messages and help output via the `Messages` struct. This example uses Portuguese messages to demonstrate localization.

Each entry includes a description that's passed to the custom error handlers, allowing you to include contextual information in localized error messages.

```go
type GreetEntries struct {
    Name string `type:"NextArg" description:"nome da pessoa a cumprimentar"`
}

// Custom error handler receives the description
MissingArg: func(arg, description, position string) string {
    msg := fmt.Sprintf("erro: argumento obrigatório '%s' não foi informado", arg)
    if description != "" {
        msg += fmt.Sprintf("\n  %s", description)
    }
    return msg
}
```

**Run it:**
```bash
go run samples/custom_errors/custom_errors.go greet Mateus
go run samples/custom_errors/custom_errors.go add -a 10 -b 20
go run samples/custom_errors/custom_errors.go greet  # Missing required arg
```

---

## Description Tags in Practice

All samples use the `description` tag to document CLI arguments. These descriptions:

- **Appear in auto-generated help** — Users see them when running `--help`
- **Appear in error messages** — When validation fails, the description provides context
- **Support localization** — Custom error handlers receive descriptions for contextual messages

### When to Use Descriptions

- **Always** for public-facing flags and arguments
- **Be specific** — "Port to listen on" is better than "port"
- **Mention constraints** — For arrays, specify min/max sizes
- **Include defaults** — Non-obvious defaults should be noted

### Description Examples

```go
// Good: specific, mentions default
Port int `type:"Flag" identifiers:"-p,--port" default:"8080"
           description:"port number (default: 8080)"`

// Good: specifies constraint
Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" min_size:"1"
               description:"labels to apply (1+ required, can repeat)"`

// Bad: too vague
Port int `type:"Flag" identifiers:"-p,--port" description:"port"`

// Bad: redundant with identifier
Host string `type:"Flag" identifiers:"-h,--host" description:"host flag"`
```

---

## Running All Samples

```bash
# Flags
go run samples/flags/flags.go serve --host localhost -p 3000 --tls

# Positional
go run samples/positional/positional.go copy src.txt dst.txt
go run samples/positional/positional.go goto file.go 42

# Arrays
go run samples/arrays/arrays.go merge f1.txt f2.txt f3.txt
go run samples/arrays/arrays.go tag -t bug -t urgent

# Mixed
go run samples/mixed/mixed.go deploy api production --image api:latest -r 5

# Git-like
go run samples/gitlike/gitlike.go clone https://github.com/user/repo
go run samples/gitlike/gitlike.go log -n 5 --all

# Types
go run samples/types/types.go flags -n Alice -a 30 --active
go run samples/types/types.go sum-ints 1 2 3 4 5

# Custom errors (Portuguese)
go run samples/custom_errors/custom_errors.go greet Alice
go run samples/custom_errors/custom_errors.go add -a 10 -b 20
```

---

## Key Takeaways

1. **Descriptions are essential** — They make CLIs user-friendly and support localization
2. **Flags are optional by design** — Use `required:"true"` to make one mandatory, or add `default` for fallback
3. **Order matters for positional args** — `NextArg` consumes in declaration order; `Arg` uses fixed positions
4. **Arrays can be bounded** — Use `start:`, `end:`, `min_size:`, `max_size:` to constrain
5. **Mix flags and positionals freely** — Extract flags first, then fill positionals from remaining args

---

## Learn More

- [Flags and Arguments Guide](../docs/flags_and_args.md) — Comprehensive patterns and best practices
- [API Reference — Entries](../docs/entries.md) — Complete tag system documentation
- [Main README](../README.md) — Quick start and feature overview

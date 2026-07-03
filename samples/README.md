# Argus Samples — Flags and Arguments in Practice

This directory contains working examples of Argus CLI applications, demonstrating how to use flags, positional arguments, and the `description` tag to build user-friendly command-line tools.

## Overview

Each sample showcases different argument patterns and how to add descriptions that appear in help output and error messages. There is no `type` tag — Argus infers the entry kind from whether a field lives in an `Args` or `Flags` sub-struct, whether it's a slice, and whether it has a `position` tag.

---

## Samples

### [flags/](flags) — Named Flags

Demonstrates flags with various configurations:

- **Required flags** — Must be provided by the user
- **Optional flags with defaults** — Fallback when omitted
- **Boolean presence flags** — No value consumed, just presence/absence
- **Descriptions** — Each flag documents its purpose

```go
type ServeFlags struct {
    Host     string `identifiers:"-h,--host" 
                      description:"the host address to bind to"`
    Port     int    `identifiers:"-p,--port" default:"8080"
                      description:"the port number to listen on (default: 8080)"`
    TLS      bool   `identifiers:"--tls"
                      description:"enable TLS/HTTPS"`
}
type ServeEntries struct {
    Flags ServeFlags
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
Fields without a `position` tag are filled in declaration order from remaining arguments.

```go
type NextArgArgs struct {
    Src  string `description:"source file path"`
    Dest string `description:"destination file path"`
}
type NextArgEntries struct {
    Args NextArgArgs
}
```

#### Arg — Fixed Position
Fields with a `position` tag bind to a specific positional index.

```go
type FixedArgArgs struct {
    Filename string `position:"0" 
                      description:"path to the file to open"`
    LineNum  int    `position:"1"
                      description:"line number to navigate to"`
}
type FixedArgEntries struct {
    Args FixedArgArgs
}
```

**Run it:**
```bash
go run samples/positional/positional.go copy readme.md /tmp/backup.md
go run samples/positional/positional.go goto main.go 42 10
```

---

### [arrays/](arrays) — Array Arguments

Demonstrates handling multiple values as arrays. Any **slice-typed** field in `Args` or `Flags` is automatically treated as an array entry:

#### ArrayFlag — Repeated Flag
A slice-typed field in `Flags`; the flag appears multiple times, each adding an element.

```go
type ArrayFlagFlags struct {
    Tags []string `identifiers:"-t,--tag" min_size:"1"
                    description:"labels to apply (can be repeated)"`
}
type ArrayFlagEntries struct {
    Flags ArrayFlagFlags
}
```

#### ArrayArg — Range of Positionals
A slice-typed field in `Args`; collects positional arguments into a slice with optional bounds.

```go
type ArrayArgArgs struct {
    Files []string `start:"0" end:"-1" min_size:"2"
                     description:"list of files to merge"`
}
type ArrayArgEntries struct {
    Args ArrayArgArgs
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

Shows how to mix flags and positional arguments in a single command by declaring both an `Args` and a `Flags` sub-struct. **Flags are extracted first**, then remaining tokens are treated as positional.

```go
type DeployArgs struct {
    Service     string `description:"name of the service to deploy"`
    Environment string `description:"target environment"`
}
type DeployFlags struct {
    Image       string `identifiers:"-i,--image"
                         description:"container image to deploy"`
    Replicas    int    `identifiers:"-r,--replicas" default:"1"
                         description:"number of replicas (default: 1)"`
}
type DeployEntries struct {
    Args  DeployArgs
    Flags DeployFlags
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
- `commit` — Required flag + optional flags with defaults, plus a `deps.Deps` field auto-injected so the callback can call `e.deps.Print(...)`
- `add` — Array positional args + boolean flag
- `remote` — Fixed positional + array flag
- `log` — All optional flags

Each command showcases different combinations and includes descriptions.

```go
// commit demonstrates dependency injection: an unexported deps.Deps field
// is populated by Argus before the callback runs.
type CommitFlags struct {
    Message string `identifiers:"-m,--message" description:"commit message"`
    Author  string `identifiers:"--author" default:"current user" description:"commit author name"`
    Amend   bool   `identifiers:"--amend" description:"amend the previous commit"`
}
type CommitEntries struct {
    Flags CommitFlags
    deps  deps.Deps
}
```

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
type TypesAsFlagsFlags struct {
    Name   string  `identifiers:"-n,--name" 
                     description:"person's name (string)"`
    Age    int     `identifiers:"-a,--age"
                     description:"person's age (int)"`
    Score  float64 `identifiers:"-s,--score"
                     description:"numeric score (float64)"`
}
type TypesAsFlagsEntries struct {
    Flags TypesAsFlagsFlags
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
type GreetArgs struct {
    Name string `description:"nome da pessoa a cumprimentar"`
}
type GreetEntries struct {
    Args GreetArgs
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
Port int `identifiers:"-p,--port" default:"8080"
           description:"port number (default: 8080)"`

// Good: specifies constraint
Tags []string `identifiers:"-t,--tag" min_size:"1"
               description:"labels to apply (1+ required, can repeat)"`

// Bad: too vague
Port int `identifiers:"-p,--port" description:"port"`

// Bad: redundant with identifier
Host string `identifiers:"-h,--host" description:"host flag"`
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

1. **No `type` tag** — Argus infers `Flag`/`ArrayFlag`/`Arg`/`NextArg`/`ArrayArg` from where a field lives (`Args` vs `Flags`), whether it's a slice, and whether it has a `position` tag.
2. **Descriptions are essential** — They make CLIs user-friendly and support localization
3. **Flags are optional by design** — Use `required:"true"` to make one mandatory, or add `default` for fallback
4. **Order matters for positional args** — Fields without `position` consume in declaration order; fields with `position` use fixed indices
5. **Arrays can be bounded** — Use `start:`, `end:`, `min_size:`, `max_size:` to constrain
6. **Mix flags and positionals freely** — Declare both `Args` and `Flags`; flags are extracted first, then positionals fill from remaining args
7. **Callbacks can access Print/Args directly** — Add a `deps deps.Deps` field (exported or not) to have Argus auto-inject it

---

## Learn More

- [Flags and Arguments Guide](../docs/flags_and_args.md) — Comprehensive patterns and best practices
- [API Reference — Entries](../docs/entries.md) — Complete tag system documentation
- [Main README](../README.md) — Quick start and feature overview

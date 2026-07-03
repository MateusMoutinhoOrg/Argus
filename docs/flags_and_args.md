# Flags and Arguments — Complete Guide

This guide covers everything about flags and positional arguments in Argus: their types, how to describe them, validation patterns, and best practices.

## Overview

CLI arguments fall into two categories, each living in its own sub-struct on the entries struct:

- **`Flags`** — Arguments preceded by identifiers (`--host`, `-p`). Flags are optional by design.
- **`Args`** — Values without a preceding flag. Position matters, either fixed or sequential.

There is no `type` tag to set. Argus infers the entry kind from which sub-struct
a field lives in, whether it's a slice, and whether it has a `position` tag:

| Kind        | Sub-struct | Rule                                      |
|-------------|-----------|---------------------------------------------|
| `Flag`      | `Flags`   | non-slice field                              |
| `ArrayFlag` | `Flags`   | slice field                                  |
| `NextArg`   | `Args`    | non-slice field, no `position` tag           |
| `Arg`       | `Args`    | non-slice field, has a `position` tag        |
| `ArrayArg`  | `Args`    | slice field                                  |

Each argument has a `description` tag to document its purpose in help output and error messages.

---

## Flags

### Flag — Named Value

A **flag** is a named argument that consumes the following token as its value. Any non-slice field declared in a `Flags` sub-struct is a `Flag`.

```go
type ServerConfigFlags struct {
    Host string `identifiers:"--host,-h" description:"hostname or IP to bind to"`
    Port int    `identifiers:"--port,-p" default:"8080" description:"port number (default: 8080)"`
}

type ServerConfig struct {
    Flags ServerConfigFlags
}
```

**Usage:**
```bash
myapp --host 0.0.0.0 --port 9090
myapp -h localhost -p 3000
```

#### Key Points

- **Identifiers** — Comma-separated aliases. Short (`-p`) and long (`--port`) are both common. Order doesn't matter. **Required** on every `Flags` field.
- **Value binding** — The next token after the identifier becomes the value.
- **Optional by design** — Flags are always optional (even without `required:"false"`).
- **Defaults** — Use `default:"<value>"` for fallback when the flag is absent.
- **Type coercion** — String flags accept any value; int/float/bool types are parsed with validation.

#### Description Tag

The `description` tag documents the flag's purpose. It appears in:

- Auto-generated help text (`--help`)
- Error messages when parsing fails
- Tool documentation

```go
type DeployConfigFlags struct {
    Env string `identifiers:"-e,--env" default:"staging" 
                 description:"deployment environment: dev, staging, prod"`
    Force bool `identifiers:"-f,--force" 
               description:"skip confirmation prompts"`
}
type DeployConfig struct {
    Flags DeployConfigFlags
}
```

#### Boolean (Presence) Flags

A `bool` field in `Flags` is a **presence flag** — no value is consumed. The flag is present or absent.

```go
type BuildOptionsFlags struct {
    Verbose bool `identifiers:"-v,--verbose" 
                   description:"print detailed build logs"`
    Release bool `identifiers:"--release" 
                  description:"build in release mode with optimizations"`
}
type BuildOptions struct {
    Flags BuildOptionsFlags
}
```

**Usage:**
```bash
myapp build --verbose
myapp build --release --verbose
myapp build
# Flags.Verbose=false, Flags.Release=false
```

---

### ArrayFlag — Repeated Flag

A **slice-typed** field in `Flags` can appear multiple times; each occurrence appends an element to the slice.

```go
type PublishConfigFlags struct {
    Tags []string `identifiers:"-t,--tag" 
                    description:"labels to apply (can be repeated)"`
    Servers []string `identifiers:"-s,--server" min_size:"1"
                       description:"target servers (at least one required)"`
}
type PublishConfig struct {
    Flags PublishConfigFlags
}
```

**Usage:**
```bash
myapp publish -t bug -t urgent -t backend -s prod-1 -s prod-2
# Flags.Tags = ["bug", "urgent", "backend"]
# Flags.Servers = ["prod-1", "prod-2"]
```

#### Constraints on ArrayFlag

- **`min_size`** — Minimum number of occurrences. Default is 0 (optional).
- **`max_size`** — Maximum number of occurrences. `-1` or absent = unbounded.

```go
type ImageOptionsFlags struct {
    Layers []string `identifiers:"-l,--layer" min_size:"1" max_size:"5"
                      description:"image layers to apply (1-5)"`
}
type ImageOptions struct {
    Flags ImageOptionsFlags
}
```

---

## Positional Arguments

Positional arguments are identified by their **position** in the command line, not by a flag name. They live in an `Args` sub-struct.

### NextArg — Sequential Consumption

Any non-slice field in `Args` **without** a `position` tag consumes the next positional argument in declaration order.

```go
type CopyEntriesArgs struct {
    Src  string `description:"source file path"`
    Dest string `description:"destination file path"`
}
type CopyEntries struct {
    Args CopyEntriesArgs
}
```

**Usage:**
```bash
myapp copy readme.md /tmp/readme-backup.md
# Args.Src = "readme.md", Args.Dest = "/tmp/readme-backup.md"
```

#### Key Points

- **Order matters** — Field declaration order determines consumption order.
- **Optional** — Use `required:"false"` to allow omission; `default:"<value>"` for a fallback.
- **Type coercion** — Parsed according to field type (string, int, float64, etc.).

```go
type GreetEntriesArgs struct {
    Name string `description:"person's name"`
    Age  int    `required:"false" default:"0" 
                 description:"person's age (optional)"`
}
type GreetEntries struct {
    Args GreetEntriesArgs
}
```

---

### Arg — Fixed Position

A non-slice field in `Args` with a **`position`** tag binds to a specific positional index.

```go
type NavigateEntriesArgs struct {
    Filename string `position:"0" description:"file path"`
    LineNum  int    `position:"1" description:"line number"`
    ColNum   int    `position:"2" required:"false" 
                     description:"column number (optional)"`
}
type NavigateEntries struct {
    Args NavigateEntriesArgs
}
```

**Usage:**
```bash
myapp goto main.go 42        # Args.Filename="main.go", Args.LineNum=42, Args.ColNum=0
myapp goto main.go 42 10     # Args.Filename="main.go", Args.LineNum=42, Args.ColNum=10
```

#### Key Points

- **Fixed indexing** — `position` is relative to the command's positional args (0-based).
- **Reorderable** — Unlike `NextArg`, you can skip positions or reorder fields.
- **Optional positions** — Use `required:"false"` for optional fixed arguments.

---

### ArrayArg — Range of Positionals

A **slice-typed** field in `Args` collects a contiguous range of positional arguments. Use `start` and `end` to define bounds; `-1` = to the end.

```go
type MergeEntriesArgs struct {
    Files []string `start:"0" end:"-1" min_size:"2"
                     description:"list of files to merge (at least 2)"`
}
type MergeEntries struct {
    Args MergeEntriesArgs
}
```

**Usage:**
```bash
myapp merge file1.txt file2.txt file3.txt
# Args.Files = ["file1.txt", "file2.txt", "file3.txt"]
```

#### Bounded Windows

Capture only specific positional indices:

```go
type SwapEntriesArgs struct {
    Pair []string `start:"0" end:"2" min_size:"2" max_size:"2"
                    description:"two files to swap"`
}
type SwapEntries struct {
    Args SwapEntriesArgs
}
```

**Usage:**
```bash
myapp swap left.txt right.txt extra.txt
# Args.Pair = ["left.txt", "right.txt"]  (end:"2" stops after index 1)
```

#### Constraints on ArrayArg

- **`min_size`** — Minimum element count. Default is 0.
- **`max_size`** — Maximum element count. `-1` or absent = unbounded.

```go
type ProcessOptionsArgs struct {
    Inputs []string `start:"0" end:"-1" min_size:"1" max_size:"10"
                      description:"input files (1-10)"`
}
type ProcessOptions struct {
    Args ProcessOptionsArgs
}
```

---

## Combining Flags and Positional Arguments

`Args` and `Flags` can coexist on the same entries struct. **Flags are extracted first**, then remaining tokens are treated as positional.

```go
type ExtractEntriesArgs struct {
    Archive string   `description:"archive file to extract from"`
    Files   []string `start:"1" end:"-1" 
                       description:"files to extract (optional)"`
}
type ExtractEntriesFlags struct {
    Output  string `identifiers:"-o,--output" default:"."
                     description:"output directory"`
    Verbose bool   `identifiers:"-v,--verbose"
                     description:"print extraction details"`
}
type ExtractEntries struct {
    Args  ExtractEntriesArgs
    Flags ExtractEntriesFlags
}
```

**Usage:**
```bash
myapp extract archive.tar.gz file1.txt file2.txt -o /tmp -v
# Args.Archive = "archive.tar.gz"
# Args.Files = ["file1.txt", "file2.txt"]
# Flags.Output = "/tmp"
# Flags.Verbose = true
```

---

## Description Tag — Best Practices

The `description` tag appears in help text and error messages. Write descriptions that:

1. **Be concise** — One sentence, <80 characters where possible.
2. **Mention defaults** — If a flag has a non-obvious default, include it in the description.
3. **Specify constraints** — For arrays, note min/max sizes or valid values.
4. **Use active voice** — "Enable TLS" not "TLS enabling".
5. **Avoid redundancy** — Don't repeat the flag name; users already see `--port`.

#### Good Descriptions

```go
Host     string `identifiers:"--host" 
                  description:"hostname or IP to bind to"`
Port     int    `identifiers:"--port" default:"8080"
                  description:"port number (default: 8080)"`
Tags     []string `identifiers:"-t,--tag"
                    description:"labels to apply (can be repeated)"`
Src      string `description:"source file path"`
```

#### Descriptions to Avoid

```go
// Too vague
Port int `identifiers:"--port" description:"the port"`

// Redundant with flag name
Host string `identifiers:"--host" description:"host flag"`

// Unclear constraints
Tags []string `identifiers:"-t,--tag" 
               description:"tags"`
```

---

## Validation and Error Handling

Argus validates arguments and produces user-friendly errors:

| Issue | Error | Example |
|-------|-------|---------|
| Missing required flag | Usage error with description | `--host is required: hostname or IP to bind to` |
| Unparseable value | Type error with expected type | `--port requires an int, got "abc"` |
| Array size violated | Size error with bounds | `--tag requires 1-5 values, got 10` |
| Missing required positional | Usage error with position and description | `Argument 0 (source file path) is required` |

All errors reference the `description` tag to guide users.

---

## Type Support

| Type | Flag | NextArg | Arg | ArrayFlag | ArrayArg |
|------|------|---------|-----|-----------|----------|
| `string` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `int`, `int64` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `float64` | ✓ | ✓ | ✓ | ✓ | ✓ |
| `bool` | ✓ (presence) | ✓ | ✓ | — | ✓ |
| `[]string` | — | — | — | ✓ | ✓ |
| `[]int`, `[]int64` | — | — | — | ✓ | ✓ |
| `[]float64` | — | — | — | ✓ | ✓ |

---

## Examples by Pattern

### Web Server with Host/Port Flags

```go
type ServeEntriesFlags struct {
    Host      string `identifiers:"-h,--host" default:"localhost"
                       description:"hostname to bind to"`
    Port      int    `identifiers:"-p,--port" default:"8080"
                       description:"port number (default: 8080)"`
    TLS       bool   `identifiers:"--tls"
                       description:"enable HTTPS"`
    CertFile  string `identifiers:"--cert-file" required:"false"
                       description:"path to SSL certificate (required if --tls)"`
}
type ServeEntries struct {
    Flags ServeEntriesFlags
}
```

### File Processing with Output Flag

```go
type ProcessEntriesArgs struct {
    Input string `description:"input file path"`
}
type ProcessEntriesFlags struct {
    Output   string `identifiers:"-o,--output" required:"false"
                      description:"output file path (default: stdout)"`
    Parallel bool   `identifiers:"-p,--parallel"
                      description:"process in parallel"`
}
type ProcessEntries struct {
    Args  ProcessEntriesArgs
    Flags ProcessEntriesFlags
}
```

### Multi-File Operations

```go
type MergeEntriesFlags struct {
    Output   string   `identifiers:"-o,--output" default:"merged.txt"
                        description:"output file path"`
    Format   string   `identifiers:"-f,--format" default:"text"
                        description:"output format: text, json, csv"`
}
type MergeEntriesArgs struct {
    Files []string `start:"0" end:"-1" min_size:"2"
                     description:"input files to merge (minimum 2)"`
}
type MergeEntries struct {
    Args  MergeEntriesArgs
    Flags MergeEntriesFlags
}

// Usage: merge file1 file2 file3 -o output.txt -f json
```

---

## See Also

- [Entries Reference](entries.md) — Full tag system documentation
- [Custom Messages](msg_format.md) — Customizing error messages and help text
- [Samples](../samples) — Working examples of flags and positional arguments

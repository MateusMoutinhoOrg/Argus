# Flags and Arguments — Complete Guide

This guide covers everything about flags and positional arguments in Argus: their types, how to describe them, validation patterns, and best practices.

## Overview

CLI arguments fall into two categories:

- **Named Flags** — Arguments preceded by identifiers (`--host`, `-p`). Flags are optional by design.
- **Positional Arguments** — Values without a preceding flag. Position matters, either fixed or sequential.

Each argument type has a `description` tag to document its purpose in help output and error messages.

---

## Flags

### Flag — Named Value

A **flag** is a named argument that consumes the following token as its value.

```go
type ServerConfig struct {
    Host string `type:"Flag" identifiers:"--host,-h" description:"hostname or IP to bind to"`
    Port int    `type:"Flag" identifiers:"--port,-p" default:"8080" description:"port number (default: 8080)"`
}
```

**Usage:**
```bash
myapp --host 0.0.0.0 --port 9090
myapp -h localhost -p 3000
```

#### Key Points

- **Identifiers** — Comma-separated aliases. Short (`-p`) and long (`--port`) are both common. Order doesn't matter.
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
type DeployConfig struct {
    Env string `type:"Flag" identifiers:"-e,--env" default:"staging" 
                 description:"deployment environment: dev, staging, prod"`
    Force bool `type:"Flag" identifiers:"-f,--force" 
               description:"skip confirmation prompts"`
}
```

#### Boolean (Presence) Flags

A `bool` field with `type:"Flag"` is a **presence flag** — no value is consumed. The flag is present or absent.

```go
type BuildOptions struct {
    Verbose bool `type:"Flag" identifiers:"-v,--verbose" 
                   description:"print detailed build logs"`
    Release bool `type:"Flag" identifiers:"--release" 
                  description:"build in release mode with optimizations"`
}
```

**Usage:**
```bash
myapp build --verbose
myapp build --release --verbose
myapp build
# Verbose=false, Release=false
```

---

### ArrayFlag — Repeated Flag

An **array flag** can appear multiple times; each occurrence appends an element to the slice.

```go
type PublishConfig struct {
    Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" 
                    description:"labels to apply (can be repeated)"`
    Servers []string `type:"ArrayFlag" identifiers:"-s,--server" min_size:"1"
                       description:"target servers (at least one required)"`
}
```

**Usage:**
```bash
myapp publish -t bug -t urgent -t backend -s prod-1 -s prod-2
# Tags = ["bug", "urgent", "backend"]
# Servers = ["prod-1", "prod-2"]
```

#### Constraints on ArrayFlag

- **`min_size`** — Minimum number of occurrences. Default is 0 (optional).
- **`max_size`** — Maximum number of occurrences. `-1` or absent = unbounded.

```go
type ImageOptions struct {
    Layers []string `type:"ArrayFlag" identifiers:"-l,--layer" min_size:"1" max_size:"5"
                      description:"image layers to apply (1-5)"`
}
```

---

## Positional Arguments

Positional arguments are identified by their **position** in the command line, not by a flag name.

### NextArg — Sequential Consumption

Each `NextArg` field consumes the next positional argument in declaration order.

```go
type CopyEntries struct {
    Src  string `type:"NextArg" description:"source file path"`
    Dest string `type:"NextArg" description:"destination file path"`
}
```

**Usage:**
```bash
myapp copy readme.md /tmp/readme-backup.md
# Src = "readme.md", Dest = "/tmp/readme-backup.md"
```

#### Key Points

- **Order matters** — Field declaration order determines consumption order.
- **Optional** — Use `required:"false"` to allow omission; `default:"<value>"` for a fallback.
- **Type coercion** — Parsed according to field type (string, int, float64, etc.).

```go
type GreetEntries struct {
    Name string `type:"NextArg" description:"person's name"`
    Age  int    `type:"NextArg" required:"false" default:"0" 
                 description:"person's age (optional)"`
}
```

---

### Arg — Fixed Position

An **Arg** binds to a specific positional index via the `position` tag.

```go
type NavigateEntries struct {
    Filename string `type:"Arg" position:"0" description:"file path"`
    LineNum  int    `type:"Arg" position:"1" description:"line number"`
    ColNum   int    `type:"Arg" position:"2" required:"false" 
                     description:"column number (optional)"`
}
```

**Usage:**
```bash
myapp goto main.go 42        # Filename="main.go", LineNum=42, ColNum=0
myapp goto main.go 42 10     # Filename="main.go", LineNum=42, ColNum=10
```

#### Key Points

- **Fixed indexing** — `position` is relative to the command's positional args (0-based).
- **Reorderable** — Unlike `NextArg`, you can skip positions or reorder fields.
- **Optional positions** — Use `required:"false"` for optional fixed arguments.

---

### ArrayArg — Range of Positionals

An **array arg** collects a contiguous range of positional arguments into a slice. Use `start` and `end` to define bounds; `-1` = to the end.

```go
type MergeEntries struct {
    Files []string `type:"ArrayArg" start:"0" end:"-1" min_size:"2"
                     description:"list of files to merge (at least 2)"`
}
```

**Usage:**
```bash
myapp merge file1.txt file2.txt file3.txt
# Files = ["file1.txt", "file2.txt", "file3.txt"]
```

#### Bounded Windows

Capture only specific positional indices:

```go
type SwapEntries struct {
    Pair []string `type:"ArrayArg" start:"0" end:"2" min_size:"2" max_size:"2"
                    description:"two files to swap"`
}
```

**Usage:**
```bash
myapp swap left.txt right.txt extra.txt
# Pair = ["left.txt", "right.txt"]  (end:"2" stops after index 1)
```

#### Constraints on ArrayArg

- **`min_size`** — Minimum element count. Default is 0.
- **`max_size`** — Maximum element count. `-1` or absent = unbounded.

```go
type ProcessOptions struct {
    Inputs []string `type:"ArrayArg" start:"0" end:"-1" min_size:"1" max_size:"10"
                      description:"input files (1-10)"`
}
```

---

## Combining Flags and Positional Arguments

Flags and positional arguments can coexist in the same struct. **Flags are extracted first**, then remaining tokens are treated as positional.

```go
type ExtractEntries struct {
    Archive  string   `type:"NextArg" description:"archive file to extract from"`
    Files    []string `type:"ArrayArg" start:"1" end:"-1" 
                       description:"files to extract (optional)"`
    Output   string   `type:"Flag" identifiers:"-o,--output" default:"."
                       description:"output directory"`
    Verbose  bool     `type:"Flag" identifiers:"-v,--verbose"
                       description:"print extraction details"`
}
```

**Usage:**
```bash
myapp extract archive.tar.gz file1.txt file2.txt -o /tmp -v
# Archive = "archive.tar.gz"
# Files = ["file1.txt", "file2.txt"]
# Output = "/tmp"
# Verbose = true
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
Host     string `type:"Flag" identifiers:"--host" 
                  description:"hostname or IP to bind to"`
Port     int    `type:"Flag" identifiers:"--port" default:"8080"
                  description:"port number (default: 8080)"`
Tags     []string `type:"ArrayFlag" identifiers:"-t,--tag"
                    description:"labels to apply (can be repeated)"`
Src      string `type:"NextArg"
                  description:"source file path"`
```

#### Descriptions to Avoid

```go
// Too vague
Port int `type:"Flag" identifiers:"--port" description:"the port"`

// Redundant with flag name
Host string `type:"Flag" identifiers:"--host" description:"host flag"`

// Unclear constraints
Tags []string `type:"ArrayFlag" identifiers:"-t,--tag" 
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
type ServeEntries struct {
    Host      string `type:"Flag" identifiers:"-h,--host" default:"localhost"
                       description:"hostname to bind to"`
    Port      int    `type:"Flag" identifiers:"-p,--port" default:"8080"
                       description:"port number (default: 8080)"`
    TLS       bool   `type:"Flag" identifiers:"--tls"
                       description:"enable HTTPS"`
    CertFile  string `type:"Flag" identifiers:"--cert-file" required:"false"
                       description:"path to SSL certificate (required if --tls)"`
}
```

### File Processing with Output Flag

```go
type ProcessEntries struct {
    Input    string `type:"NextArg" description:"input file path"`
    Output   string `type:"Flag" identifiers:"-o,--output" required:"false"
                      description:"output file path (default: stdout)"`
    Parallel bool   `type:"Flag" identifiers:"-p,--parallel"
                      description:"process in parallel"`
}
```

### Multi-File Operations

```go
type MergeEntries struct {
    Output   string   `type:"Flag" identifiers:"-o,--output" default:"merged.txt"
                        description:"output file path"`
    Format   string   `type:"Flag" identifiers:"-f,--format" default:"text"
                        description:"output format: text, json, csv"`
    Files    []string `type:"ArrayArg" start:"0" end:"-1" min_size:"2"
                        description:"input files to merge (minimum 2)"`
}

// Usage: merge file1 file2 file3 -o output.txt -f json
```

---

## See Also

- [Entries Reference](entries.md) — Full tag system documentation
- [Custom Messages](msg_format.md) — Customizing error messages and help text
- [Samples](../samples) — Working examples of flags and positional arguments

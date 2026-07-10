# CLI Entry Binding — Public API Reference

Command callbacks receive a single **entries struct**. Each field is bound to a CLI
element (positional argument or flag) through its **struct tag**. The parser reads
these tags to know *what* to extract, *where* to find it, and *how* to validate it.

> **Tag syntax matters.** Tags must follow Go's canonical form: `key:"value"` with
> **no space** after the colon and single spaces between pairs. `type: "Flag"` (with a
> space) is silently ignored by `reflect.StructTag.Get` and flagged by `go vet`. Use
> `type:"Flag"`.

```go
type Entries struct {
    Field Type `type:"..." <attr>:"..." ...`
}
```

---

## Entry types

The `type` attribute selects how the field is populated.

| `type`     | Source                                         | Field kind      |
|------------|------------------------------------------------|-----------------|
| `Arg`      | Positional argument at a fixed index           | scalar          |
| `NextArg`  | Next unconsumed, unflagged positional argument | scalar          |
| `ArrayArg` | A slice of positional arguments                | slice           |
| `Flag`     | Named flag with a value (`--port 8080`)        | scalar          |
| `ArrayFlag`| Named flag repeated to build a slice           | slice           |

A `Flag` bound to a `bool` field is a **presence flag** — no value is read; the field
becomes `true` when any identifier appears. See *Required and defaults*.

## Tag attributes

| Attribute     | Applies to                | Meaning                                                       |
|---------------|---------------------------|--------------------------------------------------------------|
| `type`        | all (required)            | Entry type from the table above                              |
| `position`    | `Arg`                     | Index among positional args (see *Open questions*)           |
| `identifiers` | `Flag`, `ArrayFlag`       | Comma-separated aliases, e.g. `"-p,--port"`                   |
| `required`    | all                       | `"true"` / `"false"` — **defaults to `"true"`**              |
| `default`     | scalars                   | Fallback value when the entry is absent; implies optional    |
| `start`,`end` | `ArrayArg`                | Slice bounds over positional args; `-1` = to the end         |
| `min_size`    | `ArrayArg`, `ArrayFlag`   | Minimum element count for validation                         |
| `max_size`    | `ArrayArg`, `ArrayFlag`   | Maximum element count; `-1` = unbounded                      |
| `help`        | all                       | **Deprecated.** Use `description` instead.                  |
| `description` | all                       | Description text (used in generated `--help` and errors)     |

## Supported field types

Scalars: `string`, `int`, `int64`, `float64`, `bool`.
Slices of any scalar for the `Array*` types: `[]string`, `[]int`, `[]float64`, …

Parsing failures (e.g. a non-numeric value bound to `float64`) should surface as a
usage error rather than a panic.

---

## Required and defaults

Every entry is **required by default** (`required:"true"`). A missing required entry
produces a usage error.

An entry becomes **optional** in either of two ways:

- `required:"false"` — absence is allowed; the field takes its zero value.
- `default:"<value>"` — absence is allowed and the field takes `<value>`. Declaring a
  `default` **implies `required:"false"`**, so you never need both.

`bool` fields bound with `type:"Flag"` are presence flags and are therefore always
optional: present → `true`, absent → `false`.

```go
type ServeEntries struct {
    Host string `type:"Flag" identifiers:"--host" description:"hostname to bind to"`
    Port int    `type:"Flag" identifiers:"-p,--port" default:"8080" description:"port number (default: 8080)"`
    TLS  bool   `type:"Flag" identifiers:"--tls" description:"enable HTTPS"`
}
```

```sh
serve --host 0.0.0.0
# Host="0.0.0.0", Port=8080, TLS=false

serve --host 0.0.0.0 --port 9090 --tls
# Host="0.0.0.0", Port=9090, TLS=true
```

---

## Positional arguments

### `NextArg` — sequential consumption

Each `NextArg` field consumes the next positional argument that hasn't been claimed
by a flag or a previous entry, in declaration order.

```go
type AddEntries struct {
    A float64 `type:"NextArg" description:"first number"`
    B float64 `type:"NextArg" description:"second number"`
}

func sum(e AddEntries, deps argus_dep.Deps) int {
    deps.Print(fmt.Sprintf("%v\n", e.A + e.B))
    return 0
}
```

```sh
calc add 10 20
# A=10, B=20  -> 30
```

### `Arg` — fixed position

Binds to a specific positional index. Useful when order is fixed but you want to
skip or reorder.

```go
type SumEntries struct {
    A float64 `type:"Arg" position:"0" description:"first number"`
    B float64 `type:"Arg" position:"1" description:"second number"`
    C float64 `type:"Arg" position:"2" description:"third number"`
}

func sum(e SumEntries, deps argus_dep.Deps) int {
    deps.Print(fmt.Sprintf("%v\n", e.A + e.B + e.C))
    return 0
}
```

```sh
calc add 10 20 39
# A=10, B=20, C=39  -> 69
```

> `position` here is shown **relative to the command's positional args** (0-based).
> The original draft used absolute `argv` indices starting at `2`; see *Open questions*.

### `ArrayArg` — slice of positionals

Collects a contiguous range of positional args as a slice. `start`/`end` behave like
a Go slice bound; `end:"-1"` means "until the last positional".

```go
type PrintEntries struct {
    Nums []int `type:"ArrayArg" start:"0" end:"-1" min_size:"1" description:"numbers to print"`
}

func printAll(e PrintEntries, deps argus_dep.Deps) int {
    for _, n := range e.Nums {
        deps.Print(fmt.Sprintf("%v\n", n))
    }
    return 0
}
```

```sh
calc print 3 7 11 42
# Nums = [3 7 11 42]
```

A bounded window (`start:"0" end:"2"`) would capture only the first two positionals.

---

## Flags

### `Flag` — named value

```go
type AddEntries struct {
    A float64 `type:"Flag" identifiers:"-a,--a" description:"first number"`
    B float64 `type:"Flag" identifiers:"-b,--b" description:"second number"`
}

func sum(e AddEntries, deps argus_dep.Deps) int {
    deps.Print(fmt.Sprintf("%v\n", e.A + e.B))
    return 0
}
```

```sh
calc add --a 10 --b 20
# also accepted: calc add -a 10 -b 20
```

### Boolean (presence) flags

A `bool` field with `type:"Flag"` reads no value — it is `true` when present. Because
absence maps to `false`, these are always optional.

```go
type BuildEntries struct {
    Verbose bool   `type:"Flag" identifiers:"-v,--verbose" description:"verbose output"`
    Output  string `type:"Flag" identifiers:"-o,--output" default:"a.out" description:"output file path (default: a.out)"`
}

func build(e BuildEntries, deps argus_dep.Deps) int {
    if e.Verbose {
        deps.Print(fmt.Sprintf("building -> %s\n", e.Output))
    }
    return 0
}
```

```sh
myc build --verbose -o bin/app
# Verbose=true, Output="bin/app"
myc build
# Verbose=false, Output="a.out"
```

### `ArrayFlag` — repeated flag

The flag may appear multiple times; each occurrence appends one element.

```go
type CollectEntries struct {
    Nums []float64 `type:"ArrayFlag" identifiers:"-a,--a" min_size:"1" max_size:"-1" description:"numbers to collect (can be repeated)"`
}

func collect(e CollectEntries, deps argus_dep.Deps) int {
    for _, n := range e.Nums {
        deps.Print(fmt.Sprintf("%v\n", n))
    }
    return 0
}
```

```sh
test -a 1 -a 2 -a 3 -a 4 -a 5
# Nums = [1 2 3 4 5]
```

---

## Combining args and flags

Positional and flag entries can coexist in one struct. Flags are extracted first,
then the remaining tokens are treated as positionals.

```go
type CopyEntries struct {
    Src   string `type:"NextArg" description:"source file path"`
    Dst   string `type:"NextArg" description:"destination file path"`
    Force bool   `type:"Flag" identifiers:"-f,--force" description:"overwrite without prompting"`
}
```

```sh
mytool copy src.txt dst.txt --force
# Src="src.txt", Dst="dst.txt", Force=true
```

---

## Return value

The callback returns an `int` used as the process **exit code** (`0` = success).
Non-zero values propagate to the shell.

---

## Validation summary

- Entries are **required by default**; a missing required entry → usage error.
- `required:"false"` or a `default` makes an entry optional.
- Optional entry absent: with `default` → default applied; without → zero value
  (`bool` → `false`).
- Value fails to parse into the field type → usage error.
- `min_size` / `max_size` violated on an array entry → usage error.

---

## Open questions / design decisions

These are unresolved in the current draft and worth pinning down before locking the API:

1. **`position` base.** The draft used absolute `argv` indices (`2`, `3`, `4` for
   `calc add 10 20 39`), which couples the index to the subcommand depth. Recommend
   making `position` **relative to the current command's positional args** (0- or
   1-based) so the same struct works regardless of nesting.
2. **Subcommand mapping.** Nothing here shows how `calc add` resolves to `sum`. Where
   is the command/subcommand tree declared, and how do callbacks attach to it?
3. **Flag value syntax.** Are both `--port 8080` and `--port=8080` accepted? Document
   the supported forms.
4. **Negative numbers vs flags.** With `-a -5`, is `-5` a value or an unknown flag?
   Define the disambiguation rule (e.g. treat `-<digit>` as a value).
5. **Unexported fields.** The examples use lowercase fields. Reflection can't read
   these; AST/codegen can. Confirm which path the parser takes, since it dictates
   whether fields must be exported.
6. **Repeated-flag vs space-separated arrays.** `ArrayFlag` shows `-a 1 -a 2`. Is
   `-a 1 2 3` also valid, or is repetition the only accepted form?
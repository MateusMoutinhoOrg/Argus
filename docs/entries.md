# CLI Entry Binding — Public API Reference

Command callbacks receive a single **entries struct**. That struct may contain up
to three things:

- an `Args` field (a struct) — positional arguments
- a `Flags` field (a struct) — named flags
- a `deps.Deps` field (exported or not) — auto-injected by Argus

Each field *inside* `Args` or `Flags` is bound to a CLI element through its
**struct tag**. There is no `type` tag anymore — Argus **infers** what kind of
entry a field is from where it lives (`Args` vs `Flags`) and from its shape
(slice vs scalar) and tags (`position` vs none).

> **Tag syntax matters.** Tags must follow Go's canonical form: `key:"value"` with
> **no space** after the colon and single spaces between pairs. `identifiers: "-p"` (with
> a space) is silently ignored by `reflect.StructTag.Get` and flagged by `go vet`. Use
> `identifiers:"-p"`.

```go
type FooArgs struct {
    Field Type `<attr>:"..." ...`
}
type FooFlags struct {
    Field Type `<attr>:"..." ...`
}
type FooEntries struct {
    Args  FooArgs
    Flags FooFlags
}
```

---

## Entry types (inferred, not declared)

| Kind       | Lives in | Rule                                            | Field kind |
|------------|----------|--------------------------------------------------|------------|
| `Arg`      | `Args`   | non-slice field **with** a `position` tag        | scalar     |
| `NextArg`  | `Args`   | non-slice field **without** a `position` tag     | scalar     |
| `ArrayArg` | `Args`   | slice field (regardless of `position`)           | slice      |
| `Flag`     | `Flags`  | non-slice field (requires `identifiers`)         | scalar     |
| `ArrayFlag`| `Flags`  | slice field (requires `identifiers`)             | slice      |

A `Flag` bound to a `bool` field is a **presence flag** — no value is read; the field
becomes `true` when any identifier appears. See *Required and defaults*.

Both `Args` and `Flags` are optional — a command with only flags omits `Args`
entirely, and vice versa.

## Tag attributes

| Attribute     | Applies to                | Meaning                                                       |
|---------------|---------------------------|--------------------------------------------------------------|
| `position`    | fields in `Args`          | Presence makes the field an `Arg`; its value is the index    |
| `identifiers` | fields in `Flags`         | Comma-separated aliases, e.g. `"-p,--port"` — **required**    |
| `required`    | all                       | `"true"` / `"false"` — **defaults to `"true"`**              |
| `default`     | scalars                   | Fallback value when the entry is absent; implies optional    |
| `start`,`end` | slice fields in `Args`    | Slice bounds over positional args; `-1` = to the end         |
| `min_size`    | slice fields              | Minimum element count for validation                         |
| `max_size`    | slice fields              | Maximum element count; `-1` = unbounded                      |
| `help`        | all                       | **Deprecated.** Use `description` instead.                  |
| `description` | all                       | Description text (used in generated `--help` and errors)     |

## Supported field types

Scalars: `string`, `int`, `int64`, `float64`, `bool`.
Slices of any scalar for array entries: `[]string`, `[]int`, `[]float64`, …

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

`bool` fields in `Flags` are presence flags and are therefore always
optional: present → `true`, absent → `false`.

```go
type ServeFlags struct {
    Host string `identifiers:"--host" description:"hostname to bind to"`
    Port int    `identifiers:"-p,--port" default:"8080" description:"port number (default: 8080)"`
    TLS  bool   `identifiers:"--tls" description:"enable HTTPS"`
}

type ServeEntries struct {
    Flags ServeFlags
}
```

```sh
serve --host 0.0.0.0
# Flags.Host="0.0.0.0", Flags.Port=8080, Flags.TLS=false

serve --host 0.0.0.0 --port 9090 --tls
# Flags.Host="0.0.0.0", Flags.Port=9090, Flags.TLS=true
```

---

## Positional arguments

Positional fields live in an `Args` sub-struct.

### `NextArg` — sequential consumption

Each field in `Args` **without** a `position` tag consumes the next positional
argument that hasn't been claimed by a flag or a previous entry, in declaration
order.

```go
type AddArgs struct {
    A float64 `description:"first number"`
    B float64 `description:"second number"`
}
type AddEntries struct {
    Args AddArgs
}

func sum(e AddEntries) int {
    fmt.Println(e.Args.A + e.Args.B)
    return 0
}
```

```sh
calc add 10 20
# Args.A=10, Args.B=20  -> 30
```

### `Arg` — fixed position

A `position` tag on a field in `Args` binds it to a specific positional index.
Useful when order is fixed but you want to skip or reorder.

```go
type SumArgs struct {
    A float64 `position:"0" description:"first number"`
    B float64 `position:"1" description:"second number"`
    C float64 `position:"2" description:"third number"`
}
type SumEntries struct {
    Args SumArgs
}

func sum(e SumEntries) int {
    fmt.Println(e.Args.A + e.Args.B + e.Args.C)
    return 0
}
```

```sh
calc add 10 20 39
# Args.A=10, Args.B=20, Args.C=39  -> 69
```

> `position` here is shown **relative to the command's positional args** (0-based).
> The original draft used absolute `argv` indices starting at `2`; see *Open questions*.

### `ArrayArg` — slice of positionals

Any **slice-typed** field in `Args` collects a contiguous range of positional args.
`start`/`end` behave like a Go slice bound; `end:"-1"` means "until the last
positional".

```go
type PrintArgs struct {
    Nums []int `start:"0" end:"-1" min_size:"1" description:"numbers to print"`
}
type PrintEntries struct {
    Args PrintArgs
}

func printAll(e PrintEntries) int {
    for _, n := range e.Args.Nums {
        fmt.Println(n)
    }
    return 0
}
```

```sh
calc print 3 7 11 42
# Args.Nums = [3 7 11 42]
```

A bounded window (`start:"0" end:"2"`) would capture only the first two positionals.

---

## Flags

Flag fields live in a `Flags` sub-struct and always require an `identifiers` tag.

### `Flag` — named value

Any **non-slice** field in `Flags` is a `Flag`.

```go
type AddFlags struct {
    A float64 `identifiers:"-a,--a" description:"first number"`
    B float64 `identifiers:"-b,--b" description:"second number"`
}
type AddEntries struct {
    Flags AddFlags
}

func sum(e AddEntries) int {
    fmt.Println(e.Flags.A + e.Flags.B)
    return 0
}
```

```sh
calc add --a 10 --b 20
# also accepted: calc add -a 10 -b 20
```

### Boolean (presence) flags

A `bool` field in `Flags` reads no value — it is `true` when present. Because
absence maps to `false`, these are always optional.

```go
type BuildFlags struct {
    Verbose bool   `identifiers:"-v,--verbose" description:"verbose output"`
    Output  string `identifiers:"-o,--output" default:"a.out" description:"output file path (default: a.out)"`
}
type BuildEntries struct {
    Flags BuildFlags
}

func build(e BuildEntries) int {
    if e.Flags.Verbose {
        fmt.Println("building ->", e.Flags.Output)
    }
    return 0
}
```

```sh
myc build --verbose -o bin/app
# Flags.Verbose=true, Flags.Output="bin/app"
myc build
# Flags.Verbose=false, Flags.Output="a.out"
```

### `ArrayFlag` — repeated flag

Any **slice-typed** field in `Flags` is an `ArrayFlag`. The flag may appear
multiple times; each occurrence appends one element.

```go
type CollectFlags struct {
    Nums []float64 `identifiers:"-a,--a" min_size:"1" max_size:"-1" description:"numbers to collect (can be repeated)"`
}
type CollectEntries struct {
    Flags CollectFlags
}

func collect(e CollectEntries) int {
    for _, n := range e.Flags.Nums {
        fmt.Println(n)
    }
    return 0
}
```

```sh
test -a 1 -a 2 -a 3 -a 4 -a 5
# Flags.Nums = [1 2 3 4 5]
```

---

## Combining args and flags

`Args` and `Flags` can coexist on the same entries struct. Flags are extracted first,
then the remaining tokens are treated as positionals.

```go
type CopyArgs struct {
    Src string `description:"source file path"`
    Dst string `description:"destination file path"`
}
type CopyFlags struct {
    Force bool `identifiers:"-f,--force" description:"overwrite without prompting"`
}
type CopyEntries struct {
    Args  CopyArgs
    Flags CopyFlags
}
```

```sh
mytool copy src.txt dst.txt --force
# Args.Src="src.txt", Args.Dst="dst.txt", Flags.Force=true
```

---

## Dependency injection

Declare a field of type `deps.Deps` anywhere on the entries struct — exported or
not — and Argus populates it automatically before invoking the callback. This
gives the callback access to `Print` (and the raw `Args`) without importing
`fmt`/`os` directly.

```go
import "github.com/MateusMoutinhoOrg/Argus/pkg/deps"

type CommitFlags struct {
    Message string `identifiers:"-m,--message" description:"commit message"`
}
type CommitEntries struct {
    Flags CommitFlags
    deps  deps.Deps
}

func commit(e CommitEntries) int {
    e.deps.Print("committed: " + e.Flags.Message)
    return 0
}
```

Unexported fields normally can't be set through reflection; Argus bypasses this
specifically for `deps.Deps` fields so the field can stay unexported by
convention. See [Dependency Injection & Testing](deps.md).

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
- Fields in `Args`/`Flags` must be exported. `Flags` fields must carry an
  `identifiers` tag. `Arg` fields (those with a `position` tag) must have a
  parseable integer position.
- Any top-level field on the entries struct other than `Args`, `Flags`, or a
  `deps.Deps` field is a configuration error, caught upfront when `HandleCli` runs.

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
5. **Repeated-flag vs space-separated arrays.** `ArrayFlag` shows `-a 1 -a 2`. Is
   `-a 1 2 3` also valid, or is repetition the only accepted form?

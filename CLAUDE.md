# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Argus** is a Go command-line argument parser library that uses reflection to bind CLI arguments and flags to struct fields through struct tags. It powers multi-command CLI applications with automatic help generation and structured error handling.

## Architecture

### Core Components

- **`pkg/Argus/`** — Core parsing engine
  - `handle.go` — Main CLI parsing logic; uses reflection to populate structs from CLI args, process flags (including ArrayFlag support), handle positional arguments (Arg/NextArg/ArrayArg), and generate help output
  - `new.go` — Factory for creating the Argus library instance
  - `errors.go` — Error message templates (being refactored to `messages.go`)

- **`pkg/deps/`** — Dependency injection
  - `deps.go` — Defines the `Deps` interface with `Args` (command-line arguments) and `Print` (output function). Allows mocking in tests and alternative implementations.

- **`adapters/native/`** — OS integration adapter
  - Implements `Deps` for real OS interaction (reads `os.Args`, prints to stdout)

- **`samples/`** — Reference implementations
  - `flags/` — Flag-based arguments and defaults
  - `positional/` — Positional argument handling (NextArg, Arg)
  - `arrays/` — Array-type arguments (ArrayFlag, ArrayArg)
  - `mixed/` — Combined flags and positional args
  - `gitlike/` — Subcommand pattern (e.g., `git commit`)
  - `types/` — Type conversions (int, float64, bool)
  - `custom_errors/` — Custom error messages via Messages struct

### Parsing Flow

1. **Struct shape inspection** — Reflects on the callback's parameter struct, which may contain an `Args` sub-struct (positional args), a `Flags` sub-struct (named flags), and/or a `deps.Deps` field. There is no `type:` tag; entry kind is **inferred** from where a field lives, whether it's a slice, and whether it has a `position:` tag.
2. **Deps injection** — Any `deps.Deps` field (exported or not) is populated with the Lib's Deps via reflection (unexported fields are set through an `unsafe` bypass).
3. **Flag extraction** — First pass collects named flags from the `Flags` sub-struct (Flag/ArrayFlag), marking consumed args
4. **Positional population** — Second pass fills the `Args` sub-struct (Arg/NextArg/ArrayArg) from remaining args
5. **Validation** — Checks required fields, applies defaults, validates array sizes
6. **Callback invocation** — Calls user function with populated struct; captures exit code

### Tag System

There is no `type:` tag. Argus **infers** how a field is populated from where it
lives and its shape:

| Kind        | Lives in | Inferred when                            |
|-------------|----------|--------------------------------------------|
| `Flag`      | `Flags`  | non-slice field (requires `identifiers:`)  |
| `ArrayFlag` | `Flags`  | slice field (requires `identifiers:`)      |
| `NextArg`   | `Args`   | non-slice field, no `position:` tag        |
| `Arg`       | `Args`   | non-slice field, has a `position:` tag     |
| `ArrayArg`  | `Args`   | slice field                                |

bool fields in `Flags` are presence-only flags. `Args` and `Flags` are both
optional on the top-level callback struct — declare whichever you need.

Modifier tags:
- `identifiers:"-p,--port"` — Flag aliases (required on every `Flags` field)
- `required:"false"` — Optional; uses zero value if missing
- `default:"8080"` — Optional with fallback; implies `required:"false"`
- `description:"..."` — Description for help text and error messages (user-facing docs)
- `help:"Description"` — **Deprecated.** Use `description` instead.
- `min_size:`, `max_size:` — Array validation

A `deps.Deps` field (any name, exported or not) is auto-injected by Argus so a
callback can call `e.deps.Print(...)` instead of using `fmt`/`os` directly.

See `docs/entries.md` for the full API reference and `docs/flags_and_args.md` for comprehensive patterns.

## Commands

### Running Samples

```bash
# Flags-based command with defaults and presence flags
go run samples/flags/flags.go serve --host 0.0.0.0
go run samples/flags/flags.go serve -h 127.0.0.1 -p 9090 --tls

# Positional arguments
go run samples/positional/positional.go

# Array arguments
go run samples/arrays/arrays.go

# Mixed flags and positional
go run samples/mixed/mixed.go

# Git-like subcommands
go run samples/gitlike/gitlike.go

# Type conversions
go run samples/types/types.go

# Custom error messages
go run samples/custom_errors/custom_errors.go
```

### Building

```bash
# Build a sample
go build -o bin/flags samples/flags/flags.go
./bin/flags serve --host 0.0.0.0 -p 9090
```

### Testing

This project currently has no tests. To validate changes:
1. Run affected samples via `go run samples/*/...go`
2. Check that help output is generated correctly (`go run samples/flags/flags.go help`)
3. Verify command-specific help (`go run samples/flags/flags.go help serve`)
4. Verify error messages appear correctly for missing required args/flags

## Common Development Tasks

### Adding a New Entry Type

1. Add the inference rule to `classifyArgField`/`classifyFlagField` in `handle.go`, and wire the new `entryKind` case into `populateFlags`/`populateArgs`/`validateRequired`/`printCommandHelp`
2. Add validation support in `validateArgsStruct`/`validateFlagsStruct`
3. Create a sample in `samples/` demonstrating the feature
4. Update `docs/entries.md` with tag syntax and examples

### Customizing Error Messages

Error messages are now built via the `Messages` struct (replacing the legacy `Errors` struct). Users pass a custom `Messages` via `GenerationProps.Errors`. See `samples/custom_errors/` for usage.

### Generating Help Text

Help is auto-generated from callback descriptions and field description tags:
- `Callback.Description` — Command summary shown in global help
- `description:"..."` tag — Detailed field description in command-specific help and error messages

Help layout:
- Global help: app name, description, list of commands, example invocation
- Command help: usage, arguments section, flags section (each with descriptions)

See `docs/flags_and_args.md` for best practices on writing descriptions.

## Key Design Decisions

- **Reflection over codegen** — Struct tags and reflection enable declarative binding without build-time code generation
- **Upfront validation** — Callback signatures are validated when `HandleCli()` is called (dev-time errors), not at parse time
- **Dependency injection via adapters** — `Deps` allows testing without `os.Args` and mocking output; the same `Deps` can also be auto-injected into a callback's `deps.Deps` field, exported or not
- **No implicit subcommand nesting** — The library handles one level; samples show patterns for deeper hierarchies
- **Arrays are open-ended by default** — `ArrayArg` and `ArrayFlag` consume all available args; use `start:`, `end:`, `min_size:`, `max_size:` to constrain

## File Organization

```
pkg/Argus/              # Core parsing logic (reflection, struct tag inspection)
pkg/deps/               # Dependency injection (Args, Print)
adapters/native/        # OS integration (os.Args, stdout)
samples/                # Reference CLI applications demonstrating features
samples/README.md       # Walkthroughs of each sample with descriptions
docs/
  entries.md            # Complete API reference for struct tags and types
  flags_and_args.md     # Comprehensive guide to flags, args, and descriptions
  deps.md               # Testing and dependency injection patterns
  msg_format.md         # Custom messages and localization
  quick_start.md        # Getting started guide
  glossary.md           # Troubleshooting and terminology
```

## Next Steps / Known Gaps

- **No automated tests** — Consider adding test suite once API stabilizes
- **Messages refactor** — `errors.go` will be renamed to `messages.go` to better reflect its role in providing user-facing output (error messages, help text, etc.)
- **Subcommand nesting** — Currently single-level; deeper trees require wrapper code in samples
- **Flag value syntax** — Currently only `--flag value` (space-separated); `--flag=value` not yet supported
- **Negative numbers** — `-5` in `int` field context may be ambiguous with flag syntax

# Glossary & Troubleshooting

## Key Terms

### Entries Struct
A Go struct whose fields define what CLI arguments a command accepts. Each field is tagged with parsing instructions (`type:"Flag"`, `type:"Arg"`, etc.). The struct is passed to a callback function when the command is invoked.

```go
type ServeEntries struct {
	Host string `type:"Flag" identifiers:"-h,--host"`
	Port int    `type:"Flag" identifiers:"-p,--port"`
}
```

### Callback
A user-defined function that executes when a command is matched. It receives an entries struct populated with parsed CLI arguments as the first parameter and an optional `deps` (dependency injection) as the second parameter, and returns an int exit code.

```go
func serve(e ServeEntries, deps argus_dep.Deps) int {
	deps.Print(fmt.Sprintf("Server on %s:%d\n", e.Host, e.Port))
	return 0 // exit code
}
```

### Flag
A named argument preceded by a dash (e.g., `--port` or `-p`). Flags accept optional values; boolean flags are presence-only (no value needed).

```bash
# Flag with value
myapp serve --port 8080

# Boolean flag (presence-only)
myapp build --verbose
```

### Positional Argument
An argument without a flag prefix, consumed in order. Used for required or sequential data.

```bash
# Three positional arguments
myapp copy src.txt dst.txt backup.txt
```

### Struct Tag
A metadata annotation on a struct field that tells Argus how to parse and validate that field.

```go
Port int `type:"Flag" identifiers:"-p,--port" default:"8080"`
//  ╭──────────────────────────────────────────────────╯
//  └─ Struct tag: tells Argus this is a flag named "-p" or "--port"
```

### Deps (Dependency Injection)
An abstraction over CLI input/output. In production, `native.New()` reads `os.Args` and writes to stdout. In tests, you inject custom `Deps` to mock input and capture output.

```go
// Production
a := argus.New(native.New())

// Testing
testDeps := deps.Deps{
	Args:  []string{"serve", "--port", "9090"},
	Print: func(s string) { /* capture */ },
}
a := argus.New(&testDeps)
```

### Message
A customizable, user-facing text output (error, help, usage). All messages are defined as functions in the `Messages` struct, allowing localization and branding.

## Common Issues & Solutions

### "Field must be exported" Error

**Problem:** Struct fields are lowercase (unexported).

```go
type ServeEntries struct {
	port int `type:"Flag" identifiers:"-p,--port"` // ✗ lowercase
}
```

**Solution:** Capitalize field names; Go reflection can only access exported (public) fields.

```go
type ServeEntries struct {
	Port int `type:"Flag" identifiers:"-p,--port"` // ✓ uppercase
}
```

---

### "Unknown flag" Error

**Problem:** A flag name isn't recognized.

```bash
myapp serve --p 8080
# error: unknown argument '--p'
```

**Solution:** Check the struct tag for the correct flag name.

```go
Port int `type:"Flag" identifiers:"-p,--port"` // ✓ recognizes -p or --port
```

---

### "Required flag not provided" Error

**Problem:** A field is required but the user didn't supply it.

```bash
myapp serve --port 8080
# error: required flag '--host' not provided
```

**Solution:** Either provide the flag or make it optional with `required:"false"` or `default:"value"`.

```go
Host string `type:"Flag" identifiers:"-h,--host" required:"false"` // now optional
// or
Host string `type:"Flag" identifiers:"-h,--host" default:"localhost"` // optional + default
```

---

### "NaN" (Not a Number) Error

**Problem:** A numeric field received a non-numeric value.

```bash
myapp serve --port abc
# error: flag '--port' is not a number
```

**Solution:** Ensure the provided value can be parsed as the field type.

```bash
myapp serve --port 8080  # ✓ valid integer
```

---

### "Positional argument missing" Error

**Problem:** A required positional argument wasn't provided.

```bash
myapp copy src.txt
# error: required argument 'Dst' not provided
```

**Solution:** Provide all required positional arguments in order, or make them optional.

```go
Dst string `type:"NextArg" required:"false"` // now optional
```

---

### Array Arguments Not Capturing Multiple Values

**Problem:** `ArrayFlag` or `ArrayArg` isn't collecting values.

```bash
myapp tag file1 file2 file3
# Expected: Files=[file1, file2, file3]
# Actual: Files=[file1]
```

**Solution:** Ensure `ArrayArg` has proper bounds, and check that flags are repeated for `ArrayFlag`.

```go
// For ArrayArg, specify the range:
Files []string `type:"ArrayArg" start:"0" end:"-1"`

// For ArrayFlag, repeat the flag:
Tags []string `type:"ArrayFlag" identifiers:"-t,--tag"`
```

Usage:

```bash
myapp tag file1 file2 file3          # ✓ ArrayArg captures all positionals
myapp -t bug -t feature -t urgent    # ✓ ArrayFlag repeats the flag
```

---

### "Invalid struct tag syntax" (silent failure)

**Problem:** Struct tag has a space after the colon; Go's `reflect.StructTag.Get()` ignores it.

```go
// ✗ Space after colon — tag is ignored
Port int `type: "Flag" identifiers: "-p,--port"`

// ✓ No spaces — correct syntax
Port int `type:"Flag" identifiers:"-p,--port"`
```

**Solution:** Use Go's canonical struct tag format: `key:"value"` with no spaces.

---

### Tests Can't Capture Output

**Problem:** Print function isn't being called; output goes directly to stdout.

```go
var output strings.Builder
testDeps := deps.Deps{
	Args: []string{"serve"},
	Print: func(s string) {
		output.WriteString(s) // Not called
	},
}
a := argus.New(&testDeps)
a.HandleCli(props)
fmt.Println(output.String()) // Empty!
```

**Solution:** Ensure you're actually using the testDeps. If using `native.New()`, it calls `fmt.Println` directly (not the `Print` callback).

```go
// ✓ Inject test deps, not native adapter
a := argus.New(&testDeps) // Correct
a := argus.New(native.New()) // ✗ Uses os, not your Print func
```

---

### Custom Messages Not Appearing

**Problem:** Error messages still show defaults.

```go
props := argus.GenerationProps{
	// Messages field accidentally omitted
	Callbacks: /* ... */,
}
```

**Solution:** Pass the `Messages` struct in `GenerationProps`.

```go
props := argus.GenerationProps{
	Messages: argus.Messages{
		MissingFlag: func(flag, description string) string {
			return fmt.Sprintf("⚠️  Flag '%s' is required", flag)
		},
	},
	Callbacks: /* ... */,
}
```

---

### Flag Value Not Being Parsed

**Problem:** Boolean flag always reads a value (consumes the next arg).

```go
// ✗ String field with Flag type reads a value
Debug string `type:"Flag" identifiers:"--debug"`

// Usage: myapp --debug true
// 'true' is the value; myapp --debug is incomplete
```

**Solution:** For presence-only flags, use `bool` type. For flags with values, use `string`, `int`, etc.

```go
// ✓ Boolean flag — presence-only, no value consumed
Debug bool `type:"Flag" identifiers:"--debug"`

// ✓ String flag — consumes next arg as value
Output string `type:"Flag" identifiers:"-o,--output"`
```

Usage:

```bash
myapp build --debug           # ✓ Debug=true
myapp build --debug -o bin    # ✓ Debug=true, Output="bin"
myapp build -o bin --verbose  # ✓ Output="bin", Verbose=true
```

---

### Negative Numbers Confused with Flags

**Problem:** `-5` is interpreted as a flag, not a number.

```bash
myapp add -5 10
# error: unknown flag '-5'
```

**Solution:** This is a known limitation. Workaround: use named flags for negative numbers.

```bash
myapp add --a -5 --b 10  # ✓ Using flags makes it clear
```

See [docs/entries.md](entries.md) for open design questions about negative number handling.

---

## Type Conversion Reference

Argus automatically converts CLI string arguments to the field type:

| Field Type | Example Input | Parses To | Notes |
|------------|--------------|-----------|-------|
| `string` | `"hello"` | `"hello"` | No conversion |
| `int` | `"42"` | `42` | Must be valid decimal; fails on `"3.14"` |
| `int64` | `"9999999999"` | `9999999999` | Larger range than `int` |
| `float64` | `"3.14"` | `3.14` | Accepts decimals and integers |
| `bool` | (none for flag) | `true` | Flags only; presence = `true` |
| `[]string` | `"a"`, `"b"`, `"c"` | `["a","b","c"]` | Multiple inputs collected |
| `[]int` | `"1"`, `"2"`, `"3"` | `[1,2,3]` | Each must parse as int |

---

## Best Practices

### 1. **Make Required Fields Explicit**

Avoid ambiguous optionality. If a field is required, don't set a default:

```go
// ✓ Clear: port is optional with default
Port int `type:"Flag" identifiers:"-p,--port" default:"8080"`

// ✓ Clear: host is required
Host string `type:"Flag" identifiers:"-h,--host"`

// ✗ Confusing: required:true is the default anyway
Port int `type:"Flag" identifiers:"-p,--port" required:"true"`
```

### 2. **Use Help Text for Context**

Every field deserves a `help` tag to guide users:

```go
Port int `type:"Flag" identifiers:"-p,--port" default:"8080" help:"TCP port to listen on"`
Host string `type:"Flag" identifiers:"-h,--host" help:"Bind address (e.g., 0.0.0.0 or localhost)"`
```

### 3. **Order Fields Logically**

In the struct, group related fields together:

```go
type ServeEntries struct {
	// Connection settings
	Host string `type:"Flag" identifiers:"-h,--host" default:"localhost"`
	Port int    `type:"Flag" identifiers:"-p,--port" default:"8080"`
	TLS  bool   `type:"Flag" identifiers:"--tls"`

	// Logging settings
	LogLevel string `type:"Flag" identifiers:"-l,--log-level" default:"info"`
	Verbose  bool   `type:"Flag" identifiers:"-v,--verbose"`
}
```

### 4. **Test Both Success and Failure Cases**

```go
// Test success path
exitCode, _ := runCLI([]string{"serve", "--port", "9090"})
if exitCode != 0 { t.Error("expected exit 0") }

// Test error path
exitCode, _ := runCLI([]string{"serve"}) // Missing required --host
if exitCode == 0 { t.Error("expected non-zero exit") }
```

### 5. **Provide Consistent Error Messages**

Localize or brand all messages together in one place:

```go
var messages = argus.Messages{
	MissingFlag: func(flag, description string) string {
		return fmt.Sprintf("❌ Flag '%s' is required: %s", flag, description)
	},
	UnknownAction: func(action string) string {
		if action == "" {
			return "❌ No command specified. Run 'help' for available commands."
		}
		return fmt.Sprintf("❌ Unknown command '%s'. Run 'help' for available commands.", action)
	},
}

// Use across multiple apps
props := argus.GenerationProps{
	Messages:  messages,
	Callbacks: /* ... */,
}
```

---

## Quick Reference

| Task | Example |
|------|---------|
| **Create a required flag** | `Host string` \` `type:"Flag" identifiers:"-h,--host"` \` |
| **Create an optional flag with default** | `Port int` \` `type:"Flag" identifiers:"-p" default:"8080"` \` |
| **Create a boolean (presence) flag** | `Verbose bool` \` `type:"Flag" identifiers:"-v"` \` |
| **Create a required positional arg** | `Name string` \` `type:"NextArg"` \` |
| **Create an optional positional arg** | `Name string` \` `type:"NextArg" required:"false"` \` |
| **Capture multiple flags** | `Tags []string` \` `type:"ArrayFlag" identifiers:"-t"` \` |
| **Capture multiple positional args** | `Files []string` \` `type:"ArrayArg" start:"0" end:"-1"` \` |
| **Test with custom input** | `deps.Deps{Args: []string{...}, Print: func(s string) {...}}` |
| **Suppress output in tests** | `Print: func(s string) {}` (no-op) |
| **Localize messages** | Create `argus.Messages` with translated functions |

---

For more details, see:
- [Quick Start](quick_start.md) — Getting started guide
- [Entries Reference](entries.md) — Complete struct tag API
- [Dependency Injection](deps.md) — Testing patterns
- [Custom Messages](msg_format.md) — Localization and branding

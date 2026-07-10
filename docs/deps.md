# Dependency Injection & Testing

Argus uses **dependency injection** to separate command logic from CLI mechanics, making it easy to test without parsing real command-line arguments.

## The `Deps` Interface

The core interface is minimal:

```go
type Deps interface {
	GetArgs() []string         // Command-line arguments
	Print(s string)            // Output function (respects quiet mode)
	SetQuiet()                 // Set the application in quiet mode
}
```

- **`Args`** — The list of CLI arguments (like `os.Args[1:]`)
- **`Print`** — A function called to output text (help, errors, etc.)

## Native Adapter

For real applications, use the native adapter that reads from the OS:

```go
import "github.com/MateusMoutinhoOrg/Argus/adapters/native"

func main() {
	a := argus.New(native.New())
	// ...
}
```

The native adapter:
- Reads `os.Args[1:]` as the command arguments
- Calls `fmt.Println()` for output

## Testing with Custom Deps

For unit tests, inject custom `Deps` to mock input and capture output:

```go
package main

import (
	"strings"
	"testing"

	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
	"github.com/MateusMoutinhoOrg/Argus/pkg/deps"
)

func TestGreet(t *testing.T) {
	// Arrange: Create a test deps with pre-defined arguments
	var output strings.Builder
	testDeps := deps.Deps{
		Args: []string{"greet", "Alice"},
		Print: func(s string) {
			output.WriteString(s)
		},
	}

	a := argus.New(&testDeps)
	
	props := argus.GenerationProps{
		Callbacks: []argus.Callback{
			{
				Starter:  "greet",
				Callback: greet,
				Description: "Greet someone",
			},
		},
	}

	// Act: Run the CLI handler
	exitCode, err := a.HandleCli(props)

	// Assert: Verify results
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(output.String(), "Hello, Alice!") {
		t.Errorf("output doesn't contain greeting: %s", output.String())
	}
}
```

## Quiet Mode: Suppressing Output

Argus has a built-in quiet system. When the user passes `--quiet` or `-q` on the command line, Argus calls `SetQuiet()` on your `Deps` before parsing, and every implementation of `Print` is expected to become a no-op from that point on. The native adapter already does this.

The quiet flag is stripped from the arguments before parsing, so it never interferes with your flags or positional arguments.

Configure it through `GenerationProps`:

```go
props := argus.GenerationProps{
	DisableQuiet:     false,                        // true disables the quiet system (default: false)
	QuietIdentifiers: []string{"--quiet", "-q"},    // flags that trigger quiet mode (this is the default)
	// ...
}
```

When implementing your own `Deps`, honor the contract:

```go
type myDeps struct {
	quiet bool
}

func (d *myDeps) SetQuiet()      { d.quiet = true }
func (d *myDeps) Print(s string) {
	if !d.quiet {
		fmt.Println(s)
	}
}
```

In tests you can also silence output unconditionally with a no-op `Print` — useful when:
- Testing only the exit code or return behavior
- You don't care about console output
- You want cleaner test logs

## Capturing Output

Store output in a variable for assertion:

```go
var output strings.Builder

testDeps := deps.Deps{
	Args: []string{"version"},
	Print: func(s string) {
		output.WriteString(s)
	},
}

// After running the CLI...
finalOutput := output.String()
if !strings.Contains(finalOutput, "v1.0.0") {
	t.Errorf("version not in output: %s", finalOutput)
}
```

## Testing Multiple Scenarios

Create a helper function to reduce boilerplate:

```go
func runCLI(args []string) (int, string, error) {
	var output strings.Builder
	testDeps := deps.Deps{
		Args: args,
		Print: func(s string) {
			output.WriteString(s)
		},
	}

	a := argus.New(&testDeps)
	exitCode, err := a.HandleCli(props)
	return exitCode, output.String(), err
}

// Use in tests:
exitCode, output, err := runCLI([]string{"greet", "Bob"})
if exitCode != 0 {
	t.Errorf("greet failed with exit code %d: %s", exitCode, output)
}
```

## Common Test Patterns

### Test a Command with Flags

```go
func TestServeWithFlags(t *testing.T) {
	exitCode, output, _ := runCLI([]string{
		"serve",
		"--host", "0.0.0.0",
		"--port", "3000",
	})
	
	if exitCode != 0 {
		t.Errorf("serve failed: %s", output)
	}
}
```

### Test Error Handling

```go
func TestMissingRequiredFlag(t *testing.T) {
	// Omit the required --port flag
	exitCode, output, _ := runCLI([]string{"serve", "--host", "localhost"})
	
	if exitCode == 0 {
		t.Error("expected non-zero exit code for missing required flag")
	}
	if !strings.Contains(output, "required") {
		t.Errorf("expected error message about required flag, got: %s", output)
	}
}
```

### Test with Array Arguments

```go
func TestCollectWithArrays(t *testing.T) {
	exitCode, _, _ := runCLI([]string{
		"collect",
		"-t", "bug",
		"-t", "feature",
		"file1.txt",
		"file2.txt",
	})
	
	if exitCode != 0 {
		t.Error("collect command failed")
	}
}
```

## Debugging Tips

### Inspect What Arguments Were Parsed

Print the `Deps.Args` before running:

```go
testDeps := deps.Deps{
	Args: []string{"serve", "--port", "8080"},
	Print: func(s string) { /* ... */ },
}
fmt.Printf("Args: %v\n", testDeps.Args)

a := argus.New(&testDeps)
exitCode, _ := a.HandleCli(props)
```

### Capture All Output for Inspection

```go
var output strings.Builder
testDeps := deps.Deps{
	Args: args,
	Print: func(s string) {
		output.WriteString(s)
		output.WriteString("\n") // Add newlines if needed
	},
}

// ...

fmt.Println("Captured output:")
fmt.Println(output.String())
```

### Verify Exit Code and Output Together

```go
exitCode, output, err := runCLI(args)

t.Logf("Exit code: %d", exitCode)
t.Logf("Output:\n%s", output)
if err != nil {
	t.Logf("Error: %v", err)
}
```

## Summary

| Task | Pattern |
|------|---------|
| **Test with custom args** | Create `Deps` with `Args` set to your test input |
| **Capture output** | Use `strings.Builder` in the `Print` function |
| **Suppress output** | Pass a no-op function: `func(s string) {}` |
| **Debug parsing** | Inspect `testDeps.Args` and `output` before/after |
| **Reduce boilerplate** | Write a `runCLI(args)` helper function |

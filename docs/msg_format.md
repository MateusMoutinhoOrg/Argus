# Custom Messages & Localization

Argus allows you to customize all user-facing messages: error text, help output, and usage hints. This is essential for **localization** (supporting multiple languages) or **custom branding**.

## The `Messages` Struct

The `Messages` struct contains **function fields**, each responsible for formatting a specific message type:

```go
type Messages struct {
	MissingFlag    func(flag, description string) string
	MissingArg     func(arg, description, position string) string
	UnknownAction  func(action string) string
	UnknownArg     func(arg string) string
	NaN            func(flag string) string
	// More fields available...
}
```

Each function:
- Takes **contextual parameters** (field name, description, etc.)
- Returns a **formatted string** to display to the user
- Lets you control tone, language, and layout

## Default Messages

If you don't provide custom messages, Argus uses sensible English defaults. These appear in error conditions:

```go
props := argus.GenerationProps{
	Callbacks: /* ... */,
	// Messages field omitted → uses defaults
}
```

## Custom English Messages

Override specific messages for clarity or branding:

```go
props := argus.GenerationProps{
	Messages: argus.Messages{
		MissingFlag: func(flag, description string) string {
			return fmt.Sprintf("⚠️  The '%s' flag is required.\n%s", flag, description)
		},
		UnknownAction: func(action string) string {
			if action == "" {
				return "Please specify a command (e.g., 'serve', 'version')"
			}
			return fmt.Sprintf("Unknown command '%s'. Run with --help for available commands.", action)
		},
	},
	Callbacks: /* ... */,
}
```

## Localization Example: Portuguese

Here's a complete example localizing to Portuguese:

```go
errosPt := argus.Messages{
	MissingFlag: func(flag, description string) string {
		msg := fmt.Sprintf("erro: flag obrigatória '%s' não foi informada", flag)
		if description != "" {
			msg += fmt.Sprintf("\n  %s", description)
		}
		return msg
	},
	MissingArg: func(arg, description, position string) string {
		msg := fmt.Sprintf("erro: argumento '%s' não foi informado", arg)
		if position != "" {
			msg += fmt.Sprintf(" (posição %s)", position)
		}
		if description != "" {
			msg += fmt.Sprintf("\n  %s", description)
		}
		return msg
	},
	UnknownAction: func(action string) string {
		if action == "" {
			return "erro: comando não especificado"
		}
		return fmt.Sprintf("erro: comando '%s' não reconhecido", action)
	},
	UnknownArg: func(arg string) string {
		return fmt.Sprintf("erro: argumento inválido '%s'", arg)
	},
	NaN: func(flag string) string {
		return fmt.Sprintf("erro: '%s' não é um número válido", flag)
	},
}

props := argus.GenerationProps{
	Messages:  errosPt,
	Callbacks: /* ... */,
}
```

## All Available Message Types

Here's the complete list of customizable messages:

| Function | Parameters | Triggered When |
|----------|-----------|-----------------|
| `MissingFlag` | `flag, description` | A required flag is not provided |
| `MissingArg` | `arg, description, position` | A required positional arg is missing |
| `UnknownAction` | `action` | The command name is unrecognized or missing |
| `UnknownArg` | `arg` | An unexpected positional argument appears |
| `NaN` | `flag` | A flag value can't be parsed as its type (e.g., non-numeric for `int`) |

## Practical Example: Build Command with Validation Messages

```go
package main

import (
	"fmt"
	"os"

	"github.com/MateusMoutinhoOrg/Argus/adapters/native"
	"github.com/MateusMoutinhoOrg/Argus/pkg/argus"
)

type BuildArgs struct {
	Target string `description:"Build target (e.g., 'linux', 'windows')"`
}
type BuildFlags struct {
	Output string `identifiers:"-o,--output" description:"Output file path"`
	Debug  bool   `identifiers:"-d,--debug" description:"Enable debug symbols"`
}
type BuildEntries struct {
	Args  BuildArgs
	Flags BuildFlags
}

func build(e BuildEntries) int {
	fmt.Printf("Building for %s → %s (debug=%v)\n", e.Args.Target, e.Flags.Output, e.Flags.Debug)
	return 0
}

func main() {
	a := argus.New(native.New())

	buildMessages := argus.Messages{
		MissingArg: func(arg, description, position string) string {
			return fmt.Sprintf(
				"⛔ Build target is required\n%s\n\nUsage: myapp build <target> -o <output>",
				description,
			)
		},
		MissingFlag: func(flag, description string) string {
			return fmt.Sprintf(
				"⛔ The '%s' flag is required\n%s",
				flag, description,
			)
		},
		UnknownArg: func(arg string) string {
			return fmt.Sprintf(
				"⛔ Unknown argument '%s'\nValid targets: linux, windows, macos",
				arg,
			)
		},
	}

	props := argus.GenerationProps{
		Messages: buildMessages,
		Callbacks: []argus.Callback{
			{
				Starter:     "build",
				Callback:    build,
				Description: "Build the project",
			},
		},
	}

	exitCode, err := a.HandleCli(props)
	if err != nil {
		fmt.Println("Configuration error:", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
```

Error output with custom messages:

```bash
$ myapp build
⛔ Build target is required
Build target (e.g., 'linux', 'windows')

Usage: myapp build <target> -o <output>
```

## Tips for Good Messages

### 1. **Be Consistent**

Use a consistent tone across all messages (formal, friendly, minimal, etc.).

```go
// ✓ Consistent style
MissingFlag: func(flag, description string) string {
	return fmt.Sprintf("error: flag '%s' is required", flag)
},
MissingArg: func(arg, description, position string) string {
	return fmt.Sprintf("error: argument '%s' is required", arg)
},
```

### 2. **Include Context**

Provide enough info for users to understand what went wrong:

```go
// ✓ Helpful
NaN: func(flag string) string {
	return fmt.Sprintf("error: flag '%s' expects a number, got '%s'", flag, /* actual value */)
},

// ✗ Confusing
NaN: func(flag string) string {
	return "invalid input"
},
```

### 3. **Suggest Next Steps**

When possible, tell users how to fix the problem:

```go
UnknownAction: func(action string) string {
	return fmt.Sprintf(
		"unknown command '%s'\nRun 'myapp help' to see available commands",
		action,
	)
},
```

### 4. **Respect the Description Field**

Use the `description` parameter when available:

```go
MissingFlag: func(flag, description string) string {
	msg := fmt.Sprintf("error: required flag '%s'", flag)
	if description != "" {
		msg += fmt.Sprintf("\n  %s", description)
	}
	return msg
},
```

## Testing Custom Messages

When testing, verify that your messages appear correctly:

```go
func TestCustomErrorMessages(t *testing.T) {
	var output strings.Builder
	testDeps := deps.Deps{
		Args: []string{"serve"}, // Missing required flags
		Print: func(s string) {
			output.WriteString(s)
		},
	}

	messages := argus.Messages{
		MissingFlag: func(flag, description string) string {
			return fmt.Sprintf("MISSING: %s", flag)
		},
	}

	props := argus.GenerationProps{
		Messages:  messages,
		Callbacks: /* ... */,
	}

	a := argus.New(&testDeps)
	a.HandleCli(props)

	// Verify custom message appeared
	if !strings.Contains(output.String(), "MISSING:") {
		t.Error("custom error message not found")
	}
}
```

## Summary

| Task | How |
|------|-----|
| **Localize to another language** | Create a `Messages` struct with translations for each function |
| **Customize error tone** | Override message functions with your preferred wording |
| **Add branding** | Include emojis, logos, or brand-specific formatting |
| **Debug message calls** | Use a helper function to log what messages are called |
| **Reuse messages** | Define a messages package and import across multiple apps |

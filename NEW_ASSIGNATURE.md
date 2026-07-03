
change api for:
### IMPORTANT:
- the type will not be required anymore.
aply the following rules to identify types.
- Arg (be in the Args struct and have a position)
- NextArg (be in the Args struct and not have a position)
- ArrayArg (be in the Args struct and have a slice type)
- Flag (be in the Flags struct and have a identifiers)
- ArrayFlag (be in the Flags struct and have a slice type)

Change all the samples for this new assignature.
Change the docs to show this new assignature.


```go

type CommitFlagsEntries struct {
	a string ` identifiers:"-a,--a" description:"commit message"`
	b []string ` identifiers:"-b,--b" description:"commit message"`
}

type CommitArgsEntries struct {
    first string `position:"0" description:"positional arguments"`
    second string `description:"positional arguments"`
    rest []string `start:"2" end:"-1" description:"positional arguments"`

}
type CallbackEntries struct {
    Flags CommitFlagsEntries
    Args  CommitArgsEntries
    deps deps.Deps
}

func my_callback(e CallbackEntries) int {
	

    e.deps.Print(e.Args.first)
    e.deps.Print(e.Args.second)
    for _, arg := range e.Args.rest {
        e.deps.Print(arg)
    }
    e.deps.Print(fmt.Sprintf("a = %v", e.Flags.b))
    for i, flag := range e.Flags.b {
        e.deps.Print(fmt.Sprintf("b(%d) = %v", i, flag))
    }

	return 0
}

```
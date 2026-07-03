package deps

type Deps struct {
	Args  []string
	Print func(s string)
	// Quiet points to the shared quiet flag consulted by Print implementations
	// before writing output. It's a pointer so toggling it (e.g. from HandleCli
	// after seeing --quiet/-q) is visible through every copy of Deps.
	Quiet *bool
}

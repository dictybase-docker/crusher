package build

// Request holds the validated build parameters extracted from CLI flags.
type Request struct {
	File string
	Tags []string
}

// CommandSpec holds the resolved executable name and argv slice.
type CommandSpec struct {
	Name string
	Args []string
}

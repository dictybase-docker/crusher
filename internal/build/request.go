package build

import "context"

const containerBinary = "container"

// Request holds the build parameters throughout the pipeline.
// It carries the CLI inputs, execution context, and resolved command spec.
type Request struct {
	File string
	Tags []string
	Ctx  context.Context
	CommandSpec
}

// CommandSpec holds the resolved executable name and argv slice.
type CommandSpec struct {
	Name string
	Args []string
}

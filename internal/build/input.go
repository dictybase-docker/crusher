package build

import "context"

const containerBinary = "container"

// Input holds the build parameters throughout the pipeline.
// It carries the CLI inputs, execution context, and resolved command spec.
type Input struct {
	File string
	Name string
	Tags []string
	Ctx  context.Context
	CommandSpec
}

// CommandSpec holds the resolved executable binary and argv slice.
type CommandSpec struct {
	Bin  string
	Args []string
}

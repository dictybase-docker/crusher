// Package build provides the logic to render the command line arguments for the
// build command.
package build

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	S "github.com/IBM/fp-go/v2/string"
)

var nameTag = S.IntersperseSemigroup(":")

func renderTagArgs(r Input) []string {
	return F.Pipe1(
		r.Tags,
		A.Chain(func(tag string) []string {
			return A.From("--tag", nameTag.Concat(r.Name, tag))
		}),
	)
}

// RenderCommand is a pure function that builds a CommandSpec from an Input
// and a resolved Dockerfile path. Called inside Execute after the
// DockerfileResource is acquired.
func RenderCommand(r Input, path string) CommandSpec {
	return CommandSpec{
		Bin: containerBinary,
		Args: A.ArrayConcatAll(
			[]string{"build", "--file", path},
			renderTagArgs(r),
			[]string{"."},
		),
	}
}

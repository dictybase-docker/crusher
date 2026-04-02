// Package containerbuild provides the implementation of the ContainerBuild
// command, which builds a container image from a Dockerfile and tags it with
// the provided name and tags.
package containerbuild

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	S "github.com/IBM/fp-go/v2/string"
)

// nameTag is a Semigroup that concatenates the name and tag with a ":" in
// between.
var nameTag = S.IntersperseSemigroup(":")

// renderTagArgs takes the tags from the Input and renders them into an array of
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

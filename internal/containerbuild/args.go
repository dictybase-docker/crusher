// Package containerbuild provides the implementation of the ContainerBuild
// command, which builds a container image from a Dockerfile and tags it with
// the provided name and tags.
package containerbuild

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	R "github.com/IBM/fp-go/v2/record"
	S "github.com/IBM/fp-go/v2/string"
)

// nameTag is a Semigroup that concatenates the name and tag with a ":" in
// between.
var nameTag = S.IntersperseSemigroup(":")

// renderBuildArgs converts the BuildArgs map into an array of
// "--build-arg" "KEY=VALUE" pairs using functional composition.
// Keys are sorted alphabetically for deterministic output.
func renderBuildArgs(buildArgs map[string]string) []string {
	return F.Pipe2(
		R.Keys(buildArgs),
		A.Sort(S.Ord),
		A.Chain(func(key string) []string {
			return []string{"--build-arg", fmt.Sprintf("%s=%s", key, buildArgs[key])}
		}),
	)
}

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
			renderBuildArgs(r.BuildArgs),
			[]string{"."},
		),
	}
}

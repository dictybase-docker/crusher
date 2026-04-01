package build

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
)

func renderTagArgs(tag string) []string {
	return []string{"--tag", tag}
}

func RenderCommand(r Input) Input {
	return Input{
		File: r.File,
		Tags: r.Tags,
		Ctx:  r.Ctx,
		CommandSpec: CommandSpec{
			Name: containerBinary,
			Args: A.ArrayConcatAll(
				[]string{"build", "--file", r.File},
				F.Pipe1(r.Tags, A.Chain(renderTagArgs)),
				[]string{"."},
			),
		},
	}
}

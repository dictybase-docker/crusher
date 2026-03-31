package build

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
)

func repeated(flag string) func([]string) []string {
	return A.Chain(func(value string) []string {
		return []string{flag, value}
	})
}

func RenderCommand(r Request) Request {
	return Request{
		File: r.File,
		Tags: r.Tags,
		Ctx:  r.Ctx,
		CommandSpec: CommandSpec{
			Name: containerBinary,
			Args: F.Pipe1(
				[][]string{
					{"build"},
					{"--file", r.File},
					F.Pipe1(r.Tags, repeated("--tag")),
					{"."},
				},
				A.Flatten[string],
			),
		},
	}
}

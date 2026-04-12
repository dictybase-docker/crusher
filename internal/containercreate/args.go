package containercreate

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
)

const containerBinary = "docker"

// RenderCommand builds the CommandSpec for "container create".
func RenderCommand(rinput ResolvedInput) CommandSpec {
	return F.Pipe8(
		A.Of("create"),
		A.Concat([]string{
			"--name",
			rinput.ContainerName,
		}),
		A.Concat(F.Pipe1(
			rinput.Mounts,
			A.Chain(renderMount),
		)),
		A.Concat(renderEnvVars(rinput.APIKey)),
		A.Concat([]string{"--workdir", rinput.Workdir}),
		A.Concat([]string{"--dns", "8.8.8.8"}),
		A.Concat([]string{"--interactive", "--tty"}),
		A.Push(rinput.ImageName),
		func(args []string) CommandSpec {
			return CommandSpec{Bin: containerBinary, Args: args}
		},
	)
}

package containercreate

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

const containerBinary = "container"

// RenderCommand builds the CommandSpec for "container create".
func RenderCommand(r ResolvedInput) CommandSpec {
	return F.Pipe6(
		[]string{"create"},
		A.Concat([]string{"--name", r.ContainerName}),
		A.Concat(renderAllMounts(r.Mounts)),
		A.Concat(renderEnvVars()),
		A.Concat(F.Pipe1(
			O.FromPredicate(func(s string) bool { return s != "" })(r.Workdir),
			O.Fold(
				func() []string { return []string{} },
				func(w string) []string { return []string{"--workdir", w} },
			),
		)),
		A.Concat([]string{r.ImageName}),
		func(args []string) CommandSpec {
			return CommandSpec{Bin: containerBinary, Args: args}
		},
	)
}

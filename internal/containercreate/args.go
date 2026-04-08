package containercreate

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

const containerBinary = "container"

// RenderCommand builds the CommandSpec for "container create".
func RenderCommand(rinput ResolvedInput) CommandSpec {
	return F.Pipe7(
		A.Of("create"),
		A.Concat([]string{
			"--name",
			rinput.ContainerName,
		}),
		A.Concat(renderAllMounts(rinput.Mounts)),
		A.Concat([]string{
			"--env",
			fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
		}),
		A.Concat([]string{
			"--env",
			fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
		}),
		A.Concat(F.Pipe1(
			O.FromPredicate(func(s string) bool { return s != "" })(rinput.Workdir),
			O.Fold(
				func() []string { return []string{} },
				func(w string) []string { return []string{"--workdir", w} },
			),
		)),
		A.Push(rinput.ImageName),
		func(args []string) CommandSpec {
			return CommandSpec{Bin: containerBinary, Args: args}
		},
	)
}

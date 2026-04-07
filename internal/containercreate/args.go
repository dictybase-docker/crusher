package containercreate

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

const containerBinary = "container"

// RenderCommand builds the CommandSpec for "container create".
func RenderCommand(r ResolvedInput) CommandSpec {
	return CommandSpec{
		Bin:  containerBinary,
		Args: buildArgs(r),
	}
}

// buildArgs constructs the full argument list using fp-go array operations.
func buildArgs(r ResolvedInput) []string {
	return A.ArrayConcatAll(
		[]string{"create"},
		buildNameArgs(r.ContainerName),
		renderAllMounts(r.Mounts),
		renderEnvVars(),
		buildWorkdirArgs(r.Workdir),
		[]string{r.ImageName},
	)
}

// buildNameArgs constructs the --name argument.
func buildNameArgs(name string) []string {
	return []string{"--name", name}
}

// buildWorkdirArgs constructs the --workdir argument if specified.
func buildWorkdirArgs(workdir string) []string {
	return F.Pipe1(
		O.FromPredicate(func(s string) bool { return s != "" })(workdir),
		O.Fold(
			func() []string { return []string{} },
			func(w string) []string { return []string{"--workdir", w} },
		),
	)
}

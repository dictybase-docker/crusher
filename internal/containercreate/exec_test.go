package containercreate

import (
	"errors"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/stretchr/testify/require"
)

func fakeRunner(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Of[error](F.VOID)
}

func fakeRunnerFail(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Left[F.Void](errors.New("container command failed"))
}

func TestExecuteWith_HappyPath(t *testing.T) {
	require := require.New(t)
	rinput := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "my-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: WorkspaceTarget,
		APIKey:  "test-key",
	}

	either := executeWith(fakeRunner, rinput)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) ContainerResult { return ContainerResult{} },
		F.Identity[ContainerResult],
	)(either)
	require.Equal("my-container", result.Name)
}

func TestExecuteWith_RunnerError(t *testing.T) {
	require := require.New(t)
	rinput := ResolvedInput{
		ImageName:     "crusher:latest",
		ContainerName: "my-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: WorkspaceTarget,
		APIKey:  "test-key",
	}

	either := executeWith(fakeRunnerFail, rinput)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(ContainerResult) error { return nil },
	)(either)
	require.EqualError(err, "container command failed")
}

func TestStartContainerWith_HappyPath(t *testing.T) {
	require := require.New(t)
	result := ContainerResult{Name: "my-container"}

	either := startContainerWith(fakeRunner, result)()
	require.True(E.IsRight(either))

	out := E.Fold(
		func(_ error) ContainerResult { return ContainerResult{} },
		F.Identity[ContainerResult],
	)(either)
	require.Equal("my-container", out.Name)
}

func TestStartContainerWith_RunnerError(t *testing.T) {
	require := require.New(t)
	result := ContainerResult{Name: "my-container"}

	either := startContainerWith(fakeRunnerFail, result)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(ContainerResult) error { return nil },
	)(either)
	require.EqualError(err, "container command failed")
}

package containerbuild

import (
	"context"
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
	return IOE.Left[F.Void](errors.New("build failed"))
}

func fakeDockerfileSource(path string) IOE.IOEither[error, DockerfileResource] {
	return IOE.Of[error](DockerfileResource{Path: path, Release: IOE.Of[error]("")})
}

func TestExecuteWith_HappyPath(t *testing.T) {
	require := require.New(t)
	input := Input{
		Name:             "testimage",
		Tags:             []string{"latest"},
		BuildArgs:        map[string]string{},
		Ctx:              context.Background(),
		DockerfileSource: fakeDockerfileSource("/fake/Dockerfile"),
	}

	either := executeWith(fakeRunner, input)()
	require.True(E.IsRight(either))
}

func TestExecuteWith_RunnerError(t *testing.T) {
	require := require.New(t)
	input := Input{
		Name:             "testimage",
		Tags:             []string{"latest"},
		BuildArgs:        map[string]string{},
		Ctx:              context.Background(),
		DockerfileSource: fakeDockerfileSource("/fake/Dockerfile"),
	}

	either := executeWith(fakeRunnerFail, input)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(F.Void) error { return nil },
	)(either)
	require.EqualError(err, "build failed")
}

func TestExecuteWith_FileSource(t *testing.T) {
	require := require.New(t)

	var captured CommandSpec
	fakeCapture := func(spec CommandSpec) IOE.IOEither[error, F.Void] {
		captured = spec
		return IOE.Of[error](F.VOID)
	}

	input := Input{
		Name:             "testimage",
		Tags:             []string{"latest"},
		BuildArgs:        map[string]string{},
		Ctx:              context.Background(),
		DockerfileSource: fakeDockerfileSource("/custom/Dockerfile"),
	}

	_ = executeWith(fakeCapture, input)()

	require.Equal(containerBinary, captured.Bin)
	require.Contains(captured.Args, "--file")
	require.Contains(captured.Args, "/custom/Dockerfile")
}

package containersbx

import (
	"errors"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/stretchr/testify/require"
)

func fakeSbxRunner(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Of[error](F.VOID)
}

func fakeSbxRunnerFail(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Left[F.Void](errors.New("sbx command failed"))
}

func TestBuildCreateArgs_WithoutSkillsPath(t *testing.T) {
	require := require.New(t)
	state := execState{
		Input: Input{
			SkillsAbsPath: "",
		},
		KitName:    "my-kit",
		OutputPath: "/tmp/out.zip",
	}

	args := buildCreateArgs(state)

	require.Len(args, 4)
	require.Equal("create", args[0])
	require.Equal("my-kit", args[1])
	require.Equal("--kit", args[2])
	require.Equal("/tmp/out.zip", args[3])
}

func TestBuildCreateArgs_WithSkillsPath(t *testing.T) {
	require := require.New(t)
	state := execState{
		Input: Input{
			SkillsAbsPath: "/abs/skills",
		},
		KitName:    "my-kit",
		OutputPath: "/tmp/out.zip",
	}

	args := buildCreateArgs(state)

	require.Len(args, 5)
	require.Equal("create", args[0])
	require.Equal("my-kit", args[1])
	require.Equal("--kit", args[2])
	require.Equal("/tmp/out.zip", args[3])
	require.Equal("/abs/skills:ro", args[4])
}

func TestResolveSkillsPath_EmptyPath(t *testing.T) {
	require := require.New(t)
	input := Input{SkillsPath: ""}

	either := resolveSkillsPath(input)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) Input { return Input{} },
		F.Identity[Input],
	)(either)
	require.Empty(result.SkillsAbsPath)
}

func TestValidateKitWith_Success(t *testing.T) {
	require := require.New(t)
	state := execState{
		TempDir: "/tmp/sbx-test",
	}

	either := validateKitWith(fakeSbxRunner)(state)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) execState { return execState{} },
		F.Identity[execState],
	)(either)
	require.Equal("/tmp/sbx-test", result.TempDir)
}

func TestValidateKitWith_Failure(t *testing.T) {
	require := require.New(t)
	state := execState{
		TempDir: "/tmp/sbx-test",
	}

	either := validateKitWith(fakeSbxRunnerFail)(state)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(execState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestStoreSecretWith_Success(t *testing.T) {
	require := require.New(t)
	state := execState{
		APIKey: "sk-abc123",
	}

	either := storeSecretWith(fakeSbxRunner)(state)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) execState { return execState{} },
		F.Identity[execState],
	)(either)
	require.Equal("sk-abc123", result.APIKey)
}

func TestStoreSecretWith_Failure(t *testing.T) {
	require := require.New(t)
	state := execState{
		APIKey: "sk-abc123",
	}

	either := storeSecretWith(fakeSbxRunnerFail)(state)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(execState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestPackKitWith_Success(t *testing.T) {
	require := require.New(t)
	state := execState{
		TempDir:    "/tmp/sbx-test",
		OutputPath: "/tmp/kit.zip",
		KitName:    "test-kit",
	}

	either := packKitWith(fakeSbxRunner)(state)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) execState { return execState{} },
		F.Identity[execState],
	)(either)
	require.Equal("/tmp/kit.zip", result.Result.OutputPath)
	require.Equal("test-kit", result.Result.KitName)
}

func TestPackKitWith_Failure(t *testing.T) {
	require := require.New(t)
	state := execState{
		TempDir:    "/tmp/sbx-test",
		OutputPath: "/tmp/kit.zip",
		KitName:    "test-kit",
	}

	either := packKitWith(fakeSbxRunnerFail)(state)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(execState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestCreateSandboxOrSkipWith_Skip(t *testing.T) {
	require := require.New(t)
	state := execState{
		Input: Input{
			ShouldCreate: false,
		},
		KitName:    "test-kit",
		OutputPath: "/tmp/kit.zip",
	}

	either := createSandboxOrSkipWith(fakeSbxRunner)(state)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) execState { return execState{} },
		F.Identity[execState],
	)(either)
	require.False(result.Result.Created)
}

func TestCreateSandboxOrSkipWith_Create(t *testing.T) {
	require := require.New(t)
	state := execState{
		Input: Input{
			ShouldCreate: true,
		},
		KitName:    "test-kit",
		OutputPath: "/tmp/kit.zip",
	}

	either := createSandboxOrSkipWith(fakeSbxRunner)(state)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(e error) execState { return execState{} },
		F.Identity[execState],
	)(either)
	require.True(result.Result.Created)
}

func TestCreateSandboxOrSkipWith_CreateFail(t *testing.T) {
	require := require.New(t)
	state := execState{
		Input: Input{
			ShouldCreate: true,
		},
		KitName:    "test-kit",
		OutputPath: "/tmp/kit.zip",
	}

	either := createSandboxOrSkipWith(fakeSbxRunnerFail)(state)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(execState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}
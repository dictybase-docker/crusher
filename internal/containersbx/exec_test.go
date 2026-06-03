package containersbx

import (
	"errors"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/stretchr/testify/require"
)

func fakeSbxRunner(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Of[error](F.VOID)
}

func fakeSbxRunnerFail(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Left[F.Void](errors.New("sbx command failed"))
}

func TestBuildCreateArgs_WithoutSkillsPath(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input: Input{
				SkillsAbsPath: "",
			},
			KitName:    "my-kit",
			OutputPath: "/tmp/out.zip",
		},
	}

	args := buildCreateArgs(ss)

	require.Len(args, 4)
	require.Equal("create", args[0])
	require.Equal("my-kit", args[1])
	require.Equal("--kit", args[2])
	require.Equal("/tmp/out.zip", args[3])
}

func TestBuildCreateArgs_WithSkillsPath(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input: Input{
				SkillsAbsPath: "/abs/skills",
			},
			KitName:    "my-kit",
			OutputPath: "/tmp/out.zip",
		},
	}

	args := buildCreateArgs(ss)

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
		func(_ error) Input { return Input{} },
		F.Identity[Input],
	)(either)
	require.Empty(result.SkillsAbsPath)
}

func TestResolveSkillsPath_NonEmptyPath(t *testing.T) {
	require := require.New(t)
	tmpDir := t.TempDir()
	input := Input{SkillsPath: tmpDir}

	either := resolveSkillsPath(input)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) Input { return Input{} },
		F.Identity[Input],
	)(either)
	require.Equal(tmpDir, result.SkillsAbsPath)
}

func TestValidateKit_Success(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{TempDir: "/tmp/sbx-test"},
		Run:   fakeSbxRunner,
	}

	either := validateKit(ss)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.Equal("/tmp/sbx-test", result.State.TempDir)
}

func TestValidateKit_Failure(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{TempDir: "/tmp/sbx-test"},
		Run:   fakeSbxRunnerFail,
	}

	either := validateKit(ss)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(stepState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestStoreSecret_Success(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{APIKey: "sk-abc123"},
		Run:   fakeSbxRunner,
	}

	either := storeSecret(ss)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.Equal("sk-abc123", result.State.APIKey)
}

func TestStoreSecret_Failure(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{APIKey: "sk-abc123"},
		Run:   fakeSbxRunnerFail,
	}

	either := storeSecret(ss)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(stepState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestPackKit_Success(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			TempDir:    "/tmp/sbx-test",
			OutputPath: "/tmp/kit.zip",
			KitName:    "test-kit",
		},
		Run: fakeSbxRunner,
	}

	either := packKit(ss)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.Equal("/tmp/kit.zip", result.State.Result.OutputPath)
	require.Equal("test-kit", result.State.Result.KitName)
}

func TestPackKit_Failure(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			TempDir:    "/tmp/sbx-test",
			OutputPath: "/tmp/kit.zip",
			KitName:    "test-kit",
		},
		Run: fakeSbxRunnerFail,
	}

	either := packKit(ss)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(stepState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

func TestCreateSandboxOrSkip_Skip(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input: Input{
				ShouldCreate: false,
			},
			KitName:    "test-kit",
			OutputPath: "/tmp/kit.zip",
		},
		Run: fakeSbxRunner,
	}

	either := createSandboxOrSkip(ss)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.False(result.State.Result.Created)
}

func TestCreateSandboxOrSkip_Create(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input: Input{
				ShouldCreate: true,
			},
			KitName:    "test-kit",
			OutputPath: "/tmp/kit.zip",
		},
		Run: fakeSbxRunner,
	}

	either := createSandboxOrSkip(ss)()
	require.True(E.IsRight(either))

	result := E.Fold(
		func(_ error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.True(result.State.Result.Created)
}

func TestCreateSandboxOrSkip_CreateFail(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input: Input{
				ShouldCreate: true,
			},
			KitName:    "test-kit",
			OutputPath: "/tmp/kit.zip",
		},
		Run: fakeSbxRunnerFail,
	}

	either := createSandboxOrSkip(ss)()
	require.True(E.IsLeft(either))

	err := E.Fold(
		F.Identity[error],
		func(stepState) error { return nil },
	)(either)
	require.EqualError(err, "sbx command failed")
}

package containeropencodebx

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	O "github.com/IBM/fp-go/v2/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tempDirPath = "/tmp/opencode-sbx-test"

func fakeSbxRunner(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Of[error](F.VOID)
}

func fakeSbxRunnerFail(_ CommandSpec) IOE.IOEither[error, F.Void] {
	return IOE.Left[F.Void](errors.New("sbx command failed"))
}

// fakeSbxRunnerFailOnSecond returns a runner that succeeds on call 1 and fails
// on call 2 (used to test createWithSecret secret-store failure).
func fakeSbxRunnerFailOnSecond() processRunner {
	calls := 0

	return func(_ CommandSpec) IOE.IOEither[error, F.Void] {
		calls++

		return F.Pipe2(
			calls,
			O.FromPredicate(func(n int) bool { return n < 2 }),
			O.Fold(
				func() IOE.IOEither[error, F.Void] {
					return IOE.Left[F.Void](errors.New("sbx command failed"))
				},
				func(int) IOE.IOEither[error, F.Void] {
					return IOE.Of[error](F.VOID)
				},
			),
		)
	}
}

// capturingRunner records every CommandSpec it receives while succeeding.
func capturingRunner(seen *[]CommandSpec) processRunner {
	return func(spec CommandSpec) IOE.IOEither[error, F.Void] {
		*seen = append(*seen, spec)
		return IOE.Of[error](F.VOID)
	}
}

func TestGlobEmbedded_SkillPattern(t *testing.T) {
	either := globEmbedded("global/skills/*/SKILL.md")()
	require.True(t, E.IsRight(either))
	paths := E.Fold(
		func(error) []string { return nil },
		F.Identity[[]string],
	)(either)
	assert.Contains(t, paths, "global/skills/git-commit/SKILL.md")
}

func TestGlobalFilePaths_ReturnsOnlyFiles(t *testing.T) {
	either := globalFilePaths()()
	require.True(t, E.IsRight(either))
	paths := E.Fold(
		func(error) []string { return nil },
		F.Identity[[]string],
	)(either)
	require.NotEmpty(t, paths)

	for _, p := range paths {
		assert.False(t, endsWith(p, "/"))
		assert.NotEqual(t, "global/skills", p)
		assert.NotEqual(t, "global/agents", p)
	}
}

func TestGlobalFilePaths_ContainsExpectedSkill(t *testing.T) {
	either := globalFilePaths()()
	require.True(t, E.IsRight(either))
	paths := E.Fold(
		func(error) []string { return nil },
		F.Identity[[]string],
	)(either)
	assert.Contains(t, paths, "global/skills/git-commit/SKILL.md")
}

func TestGlobalFilePaths_ExcludesKnownDirectory(t *testing.T) {
	either := globalFilePaths()()
	require.True(t, E.IsRight(either))
	paths := E.Fold(
		func(error) []string { return nil },
		F.Identity[[]string],
	)(either)
	assert.NotContains(t, paths, "global/skills")
	assert.NotContains(t, paths, "global/agents")
}

func TestWriteOneGlobalFile_CreatesFileAtCorrectPath(t *testing.T) {
	require := require.New(t)
	kitDir := t.TempDir()
	either := writeOneGlobalFile(kitDir)("global/skills/git-commit/SKILL.md")()
	require.True(E.IsRight(either))

	dest := filepath.Join(
		kitDir,
		"files",
		"home",
		".config",
		"opencode",
		"skills",
		"git-commit",
		"SKILL.md",
	)
	_, err := os.Stat(dest)
	require.NoError(err)
}

func TestWriteGlobalFiles_AllFilesPresent(t *testing.T) {
	require := require.New(t)
	tmpDir := t.TempDir()
	either := writeGlobalFiles(genState{tempDir: tmpDir})()
	require.True(E.IsRight(either))

	skillsDir := filepath.Join(tmpDir, "files", "home", ".config", "opencode", "skills")
	entries, err := os.ReadDir(skillsDir)
	require.NoError(err)
	require.NotEmpty(entries)

	found := false

	for _, e := range entries {
		if e.IsDir() {
			_, err := os.Stat(filepath.Join(skillsDir, e.Name(), "SKILL.md"))
			if err == nil {
				found = true
				break
			}
		}
	}

	assert.True(t, found, "expected at least one SKILL.md under skills/")
}

func TestValidateKit_Success(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{TempDir: tempDirPath},
		Run:   fakeSbxRunner,
	}
	either := validateKit(ss)()
	require.True(E.IsRight(either))
	result := E.Fold(
		func(error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.Equal(tempDirPath, result.State.TempDir)
}

func TestValidateKit_Failure(t *testing.T) {
	ss := stepState{
		State: execState{TempDir: tempDirPath},
		Run:   fakeSbxRunnerFail,
	}
	either := validateKit(ss)()
	require.True(t, E.IsLeft(either))
}

func TestPackKit_SetsOutputPath(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			TempDir:    tempDirPath,
			OutputPath: testOutputPath,
		},
		Run: fakeSbxRunner,
	}
	either := packKit(ss)()
	require.True(E.IsRight(either))
	result := E.Fold(
		func(error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.Equal(testOutputPath, result.State.Result.OutputPath)
	require.Equal(agentKitName, result.State.Result.KitName)
}

func TestCreateWithSecret_ShouldCreateFalse(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{Input: Input{ShouldCreate: false}},
		Run:   fakeSbxRunnerFail,
	}
	either := createWithSecret(ss)()
	require.True(E.IsRight(either))
	result := E.Fold(
		func(error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.False(result.State.Result.Created)
}

func TestCreateWithSecret_UsesProviderAsSecretService(t *testing.T) {
	require := require.New(t)

	var seen []CommandSpec

	ss := stepState{
		State: execState{
			Input:      Input{ShouldCreate: true, Provider: providerAnthropic},
			APIKey:     testAPIKey,
			OutputPath: testOutputPath,
		},
		Run: capturingRunner(&seen),
	}
	either := createWithSecret(ss)()
	require.True(E.IsRight(either))
	require.NotEmpty(seen)
	// First call is the secret set: args = ["secret","set","-g",<provider>]
	require.Equal("anthropic", seen[0].Args[3])
}

func TestCreateWithSecret_SetCreatedTrue(t *testing.T) {
	require := require.New(t)
	ss := stepState{
		State: execState{
			Input:      Input{ShouldCreate: true},
			APIKey:     testAPIKey,
			OutputPath: testOutputPath,
		},
		Run: fakeSbxRunner,
	}
	either := createWithSecret(ss)()
	require.True(E.IsRight(either))
	result := E.Fold(
		func(error) stepState { return stepState{} },
		F.Identity[stepState],
	)(either)
	require.True(result.State.Result.Created)
}

func TestCreateWithSecret_StoreSecretFails(t *testing.T) {
	ss := stepState{
		State: execState{Input: Input{ShouldCreate: true}, APIKey: testAPIKey},
		Run:   fakeSbxRunnerFail,
	}
	either := createWithSecret(ss)()
	require.True(t, E.IsLeft(either))
}

func TestCreateWithSecret_CreateFails(t *testing.T) {
	ss := stepState{
		State: execState{Input: Input{ShouldCreate: true}, APIKey: testAPIKey},
		Run:   fakeSbxRunnerFailOnSecond(),
	}
	either := createWithSecret(ss)()
	require.True(t, E.IsLeft(either))
}

// endsWith is a tiny pure helper used by TestGlobalFilePaths_ReturnsOnlyFiles.
// Defined locally to avoid pulling in an extra import for a one-off check.
func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

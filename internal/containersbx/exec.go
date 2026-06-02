package containersbx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/IBM/fp-go/v2/file"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FILE "github.com/IBM/fp-go/v2/ioeither/file"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	Str "github.com/IBM/fp-go/v2/string"
)

// Execute runs the full pipeline: generate → validate → pack → optionally create → cleanup.
func Execute(input Input) IOE.IOEither[error, KitResult] {
	return F.Pipe7(
		input,
		generateToTempDir,
		IOE.Chain(validateKit),
		IOE.Chain(storeSecret),
		IOE.Chain(packKit),
		IOE.Chain(createSandboxOrSkip),
		IOE.Chain(cleanupTempDir),
		IOE.Map[error](func(state execState) KitResult {
			return state.Result
		}),
	)
}

// genState carries intermediate values through the generateToTempDir pipeline.
type genState struct {
	input         Input
	configContent string // set by ReadConfig, consumed by GenerateSpec
	spec          string // set by GenerateSpec
	tempDir       string // set by makeTempDirWithSpec
}

// skillsAbsPathL is a lens for the SkillsAbsPath field in Input.
var skillsAbsPathL = L.MakeLens(
	func(i Input) string { return i.SkillsAbsPath },
	func(i Input, s string) Input { i.SkillsAbsPath = s; return i },
)

// resolveSkillsPath resolves input.SkillsPath to an absolute path if non-empty.
func resolveSkillsPath(input Input) IOE.IOEither[error, Input] {
	return F.Pipe2(
		input.SkillsPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() IOE.IOEither[error, Input] {
				return IOE.Of[error](input)
			},
			func(skillsPath string) IOE.IOEither[error, Input] {
				return F.Pipe2(
					IOE.TryCatchError(func() (string, error) {
						return filepath.Abs(skillsPath)
					}),
					IOE.MapLeft[string](func(err error) error {
						return fmt.Errorf(
							"failed to resolve skills path: %w",
							err,
						)
					}),
					IOE.Map[error](func(absPath string) Input {
						return F.Pipe1(input, skillsAbsPathL.Set(absPath))
					}),
				)
			},
		),
	)
}

// makeTempDirWithSpec creates a temp directory and stores it in genState.
func makeTempDirWithSpec(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return os.MkdirTemp("", "crush-sbx-*")
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}),
		IOE.Map[error](func(tempDir string) genState {
			return genState{
				input:         gs.input,
				configContent: gs.configContent,
				spec:          gs.spec,
				tempDir:       tempDir,
			}
		}),
	)
}

// writeSpecAndBuild writes spec.yaml to the temp dir and builds execState.
func writeSpecAndBuild(gs genState) IOE.IOEither[error, execState] {
	specPath := file.Join("spec.yaml")(gs.tempDir)
	return F.Pipe3(
		[]byte(gs.spec),
		FILE.WriteFile(specPath, filePerm),
		IOE.MapLeft[[]byte](func(err error) error {
			_ = os.RemoveAll(gs.tempDir)
			return fmt.Errorf("failed to write spec.yaml: %w", err)
		}),
		IOE.Chain(func(_ []byte) IOE.IOEither[error, execState] {
			return buildExecState(gs.input, gs.tempDir)
		}),
	)
}

// buildExecState resolves the output path and constructs the final execState.
func buildExecState(input Input, tempDir string) IOE.IOEither[error, execState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return filepath.Abs(input.OutputPath)
		}),
		IOE.MapLeft[string](func(err error) error {
			_ = os.RemoveAll(tempDir)
			return fmt.Errorf("failed to resolve output path: %w", err)
		}),
		IOE.Map[error](func(absOutput string) execState {
			return execState{
				Input:      input,
				TempDir:    tempDir,
				OutputPath: absOutput,
				KitName:    input.KitName,
				APIKey:     input.APIKey,
			}
		}),
	)
}

// generateToTempDir reads config, resolves skills path, renders spec, writes to os.MkdirTemp.
func generateToTempDir(input Input) IOE.IOEither[error, execState] {
	return F.Pipe4(
		resolveSkillsPath(input),
		IOE.Chain(ReadConfig),
		IOE.Chain(GenerateSpec),
		IOE.Chain(makeTempDirWithSpec),
		IOE.Chain(writeSpecAndBuild),
	)
}

// validateKit runs "sbx kit validate <tempDir>".
func validateKit(state execState) IOE.IOEither[error, execState] {
	return F.Pipe1(
		runSbxCommand(CommandSpec{
			Bin:  sbxBinary,
			Args: []string{"kit", "validate", state.TempDir},
		}),
		IOE.Map[error](func(F.Void) execState { return state }),
	)
}

// storeSecret runs "sbx secret set openrouter" using the API key via stdin.
func storeSecret(state execState) IOE.IOEither[error, execState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (F.Void, error) {
			bin, err := exec.LookPath(sbxBinary)
			if err != nil {
				return F.VOID, fmt.Errorf(
					"sbx not found in PATH; install it from https://docs.docker.com/sbx/: %w",
					err,
				)
			}
			cmd := &exec.Cmd{
				Path:   bin,
				Args:   []string{bin, "secret", "set", "openrouter"},
				Stdin:  strings.NewReader(state.APIKey + "\n"),
				Stdout: os.Stdout,
				Stderr: os.Stderr,
			}
			return F.VOID, cmd.Run()
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx secret set failed: %w", err)
		}),
		IOE.Map[error](func(F.Void) execState { return state }),
	)
}

// packKit runs "sbx kit pack <tempDir> -o <outputPath>".
func packKit(state execState) IOE.IOEither[error, execState] {
	return F.Pipe1(
		runSbxCommand(CommandSpec{
			Bin:  sbxBinary,
			Args: []string{"kit", "pack", state.TempDir, "-o", state.OutputPath},
		}),
		IOE.Map[error](func(F.Void) execState {
			state.Result.OutputPath = state.OutputPath
			state.Result.KitName = state.KitName
			return state
		}),
	)
}

// createSandboxOrSkip runs "sbx create <name> --kit <outputPath>" if ShouldCreate.
// If SkillsAbsPath is set, it's appended as a read-only workspace.
func createSandboxOrSkip(state execState) IOE.IOEither[error, execState] {
	if !state.ShouldCreate {
		return IOE.Of[error](state)
	}
	args := []string{
		"create", state.KitName,
		"--kit", state.OutputPath,
	}
	if state.SkillsAbsPath != "" {
		args = append(args, state.SkillsAbsPath+":ro")
	}
	return F.Pipe1(
		runSbxCommand(CommandSpec{
			Bin:  sbxBinary,
			Args: args,
		}),
		IOE.Map[error](func(F.Void) execState {
			state.Result.Created = true
			return state
		}),
	)
}

// cleanupTempDir removes the temp directory.
func cleanupTempDir(state execState) IOE.IOEither[error, execState] {
	return IOE.TryCatchError(func() (execState, error) {
		_ = os.RemoveAll(state.TempDir)
		return state, nil
	})
}

// runSbxCommand executes an sbx CLI command.
func runSbxCommand(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return exec.LookPath(spec.Bin)
		}),
		IOE.Chain(func(bin string) IOE.IOEither[error, F.Void] {
			return IOE.TryCatchError(func() (F.Void, error) {
				cmd := &exec.Cmd{
					Path:   bin,
					Args:   append([]string{bin}, spec.Args...),
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				}
				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx command failed: %w", err)
		}),
	)
}

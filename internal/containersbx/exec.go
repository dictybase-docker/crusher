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

// processRunner is a type alias for an sbx subprocess runner, enabling injection
// of test doubles.
type processRunner func(spec CommandSpec) IOE.IOEither[error, F.Void]

// Execute runs the full pipeline: generate → validate → pack → optionally create → cleanup.
func Execute(input Input) IOE.IOEither[error, KitResult] {
	return executeWith(runSbxCommand, input)
}

// executeWith is the internal parameterized variant that accepts a
// processRunner, enabling unit tests to inject stubs for each stage.
func executeWith(run processRunner, input Input) IOE.IOEither[error, KitResult] {
	return F.Pipe8(
		input,
		generateToTempDir,
		IOE.Chain(validateKitWith(run)),
		IOE.Chain(storeSecretWith(run)),
		IOE.Chain(packKitWith(run)),
		IOE.Chain(createSandboxOrSkipWith(run)),
		IOE.Chain(func(state execState) IOE.IOEither[error, execState] {
			return IOE.TryCatchError(func() (execState, error) {
				err := os.RemoveAll(state.TempDir)
				return state, err
			})
		}),
		IOE.MapLeft[execState](func(err error) error {
			return fmt.Errorf("failed to cleanup temp dir: %w", err)
		}),
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
	return F.Pipe7(
		gs.tempDir,
		file.Join("spec.yaml"),
		FILE.Create,
		FILE.WriteAll[*os.File]([]byte(gs.spec)),
		IOE.MapLeft[[]byte](func(err error) error {
			return fmt.Errorf("failed to write spec.yaml: %w", err)
		}),
		IOE.Chain(func(_ []byte) IOE.IOEither[error, string] {
			return IOE.TryCatchError(func() (string, error) {
				return filepath.Abs(gs.input.OutputPath)
			})
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("failed to resolve output path: %w", err)
		}),
		IOE.Map[error](func(absOutput string) execState {
			return execState{
				Input:      gs.input,
				TempDir:    gs.tempDir,
				OutputPath: absOutput,
				KitName:    gs.input.KitName,
				APIKey:     gs.input.APIKey,
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
	return validateKitWith(runSbxCommand)(state)
}

// validateKitWith returns a Kleisli arrow for kit validation.
func validateKitWith(run processRunner) func(execState) IOE.IOEither[error, execState] {
	return func(state execState) IOE.IOEither[error, execState] {
		return F.Pipe1(
			run(CommandSpec{
				Bin:  sbxBinary,
				Args: []string{"kit", "validate", state.TempDir},
			}),
			IOE.Map[error](func(F.Void) execState { return state }),
		)
	}
}

// storeSecret runs "sbx secret set openrouter" using the API key via stdin.
func storeSecret(state execState) IOE.IOEither[error, execState] {
	return F.Pipe3(
		IOE.TryCatchError(func() (string, error) {
			return exec.LookPath(sbxBinary)
		}),
		IOE.Chain(func(bin string) IOE.IOEither[error, F.Void] {
			return IOE.TryCatchError(func() (F.Void, error) {
				cmd := &exec.Cmd{
					Path:   bin,
					Args:   []string{bin, "secret", "set", "openrouter"},
					Stdin:  strings.NewReader(state.APIKey + "\n"),
					Stdout: os.Stdout,
					Stderr: os.Stderr,
				}
				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx command failed: %w", err)
		}),
		IOE.Map[error](func(F.Void) execState { return state }),
	)
}

// storeSecretWith returns a Kleisli arrow for secret storage (testable variant).
func storeSecretWith(run processRunner) func(execState) IOE.IOEither[error, execState] {
	return func(state execState) IOE.IOEither[error, execState] {
		return F.Pipe1(
			run(CommandSpec{
				Bin:  sbxBinary,
				Args: []string{"secret", "set", "openrouter"},
			}),
			IOE.Map[error](func(F.Void) execState { return state }),
		)
	}
}

// packKit runs "sbx kit pack <tempDir> -o <outputPath>".
func packKit(state execState) IOE.IOEither[error, execState] {
	return packKitWith(runSbxCommand)(state)
}

// packKitWith returns a Kleisli arrow for kit packing.
func packKitWith(run processRunner) func(execState) IOE.IOEither[error, execState] {
	return func(state execState) IOE.IOEither[error, execState] {
		return F.Pipe1(
			run(CommandSpec{
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
}

// buildCreateArgs renders the args slice for "sbx create" with an optional --workspace flag.
func buildCreateArgs(state execState) []string {
	return F.Pipe2(
		state.SkillsAbsPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() []string {
				return []string{"create", state.KitName, "--kit", state.OutputPath}
			},
			func(p string) []string {
				return []string{"create", state.KitName, "--kit", state.OutputPath, p + ":ro"}
			},
		),
	)
}

// createSandboxOrSkip runs "sbx create <name> --kit <outputPath>" if ShouldCreate.
func createSandboxOrSkip(state execState) IOE.IOEither[error, execState] {
	return createSandboxOrSkipWith(runSbxCommand)(state)
}

// createSandboxOrSkipWith returns a Kleisli arrow for sandbox creation.
func createSandboxOrSkipWith(run processRunner) func(execState) IOE.IOEither[error, execState] {
	return func(state execState) IOE.IOEither[error, execState] {
		return F.Pipe2(
			state.ShouldCreate,
			O.FromPredicate(F.Identity[bool]),
			O.Fold(
				func() IOE.IOEither[error, execState] {
					return IOE.Of[error](state)
				},
				func(_ bool) IOE.IOEither[error, execState] {
					return F.Pipe1(
						run(CommandSpec{
							Bin:  sbxBinary,
							Args: buildCreateArgs(state),
						}),
						IOE.Map[error](func(F.Void) execState {
							state.Result.Created = true
							return state
						}),
					)
				},
			),
		)
	}
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
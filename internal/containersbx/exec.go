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

// skillsAbsPathL is a lens for the SkillsAbsPath field in Input.
var skillsAbsPathL = L.MakeLens(
	func(i Input) string { return i.SkillsAbsPath },
	func(i Input, s string) Input { i.SkillsAbsPath = s; return i },
)

// processRunner is a type alias for an sbx subprocess runner, enabling injection
// of test doubles.
type processRunner func(spec CommandSpec) IOE.IOEither[error, F.Void]

// stepState pairs the domain state with the process runner, enabling a fully
// univariate pipeline where every handler is stepState → IOE[stepState].
type stepState struct {
	State execState
	Run   processRunner
}

// Execute runs the full pipeline: generate → validate → pack → optionally store secret + create.
func Execute(input Input) IOE.IOEither[error, KitResult] {
	return F.Pipe6(
		input,
		generateToTempDir,
		IOE.Map[error](func(es execState) stepState {
			return stepState{State: es, Run: runSbxCommand}
		}),
		IOE.Chain(validateKit),
		IOE.Chain(packKit),
		IOE.Chain(createWithSecret),
		IOE.Map[error](func(ss stepState) KitResult {
			return ss.State.Result
		}),
	)
}

// runSbxCommand executes an sbx CLI command, optionally piping stdin content.
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

				if spec.Stdin != "" {
					cmd.Stdin = strings.NewReader(spec.Stdin)
				}

				return F.VOID, cmd.Run()
			})
		}),
		IOE.MapLeft[F.Void](func(err error) error {
			return fmt.Errorf("sbx command failed: %w", err)
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

// validateKit runs "sbx kit validate" against the temp dir.
func validateKit(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe1(
		ss.Run(CommandSpec{
			Bin:  sbxBinary,
			Args: []string{"kit", "validate", ss.State.TempDir},
		}),
		IOE.Map[error](func(F.Void) stepState { return ss }),
	)
}

// packKit runs "sbx kit pack" to produce the output zip.
func packKit(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe1(
		ss.Run(CommandSpec{
			Bin:  sbxBinary,
			Args: []string{"kit", "pack", ss.State.TempDir, "-o", ss.State.OutputPath},
		}),
		IOE.Map[error](func(F.Void) stepState {
			ss.State.Result.OutputPath = ss.State.OutputPath
			ss.State.Result.KitName = agentKitName

			return ss
		}),
	)
}

// buildCreateArgs renders the args slice for "sbx create" with an optional --workspace flag.
func buildCreateArgs(ss stepState) []string {
	return F.Pipe2(
		ss.State.SkillsAbsPath,
		O.FromPredicate(Str.IsNonEmpty),
		O.Fold(
			func() []string {
				return []string{
					createCmd,
					agentKitName,
					"--kit",
					ss.State.OutputPath,
				}
			},
			func(p string) []string {
				return []string{
					createCmd,
					agentKitName,
					"--kit",
					ss.State.OutputPath,
					p + ":ro",
				}
			},
		),
	)
}

// createWithSecret conditionally stores the API-key secret then creates the
// sandbox. Both actions share the same ShouldCreate gate — storing a global
// secret only makes sense when a sandbox is about to consume it.
func createWithSecret(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe2(
		ss.State.ShouldCreate,
		O.FromPredicate(F.Identity[bool]),
		O.Fold(
			func() IOE.IOEither[error, stepState] {
				return IOE.Of[error](ss)
			},
			func(_ bool) IOE.IOEither[error, stepState] {
				return F.Pipe2(
					ss.Run(CommandSpec{
						Bin:   sbxBinary,
						Args:  []string{"secret", "set", "-g", "openrouter"},
						Stdin: ss.State.APIKey + "\n",
					}),
					IOE.Chain(func(F.Void) IOE.IOEither[error, F.Void] {
						return ss.Run(CommandSpec{
							Bin:  sbxBinary,
							Args: buildCreateArgs(ss),
						})
					}),
					IOE.Map[error](func(F.Void) stepState {
						ss.State.Result.Created = true
						return ss
					}),
				)
			},
		),
	)
}

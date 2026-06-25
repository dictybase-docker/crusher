package containeropencodebx

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	file "github.com/IBM/fp-go/v2/file"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FILE "github.com/IBM/fp-go/v2/ioeither/file"
	O "github.com/IBM/fp-go/v2/option"

	"github.com/dictybase-docker/crusher/internal/sbxexec"
)

// trimGlobalPrefix strips the leading "global/" segment from an embedded path
// so it can be relocated under ~/.config/opencode/.
func trimGlobalPrefix(path string) string {
	return strings.TrimPrefix(path, "global/")
}

// globalPatterns enumerates every embedded file shape under global/.
var globalPatterns = []string{
	"global/skills/*/SKILL.md",
	"global/agents/*.md",
	"global/commands/*.md",
	"global/plugins/*.ts",
	"global/plugins/*.js",
}

// globEmbedded matches a single pattern against the embedded FS.
func globEmbedded(pattern string) IOE.IOEither[error, []string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([]string, error) {
			return fs.Glob(globalFS, pattern)
		}),
		IOE.MapLeft[[]string](func(err error) error {
			return fmt.Errorf("glob embedded %s: %w", pattern, err)
		}),
	)
}

// globalFilePaths collects every embedded global file path, flattened.
func globalFilePaths() IOE.IOEither[error, []string] {
	return F.Pipe2(
		globalPatterns,
		IOE.TraverseArray[error](globEmbedded),
		IOE.Map[error](func(paths [][]string) []string {
			return A.Flatten(paths)
		}),
	)
}

// writeOneGlobalFile returns a Kleisli arrow that reads one embedded global
// file and writes it into the kit's files/home/.config/opencode/ subtree.
func writeOneGlobalFile(kitDir string) func(string) IOE.IOEither[error, []byte] {
	return func(globalPath string) IOE.IOEither[error, []byte] {
		destAbs := F.Pipe1(
			kitDir,
			file.Join(
				filepath.Join("files", "home", ".config", "opencode", trimGlobalPrefix(globalPath)),
			),
		)

		return F.Pipe2(
			IOE.TryCatchError(func() ([]byte, error) {
				return fs.ReadFile(globalFS, globalPath)
			}),
			IOE.MapLeft[[]byte](func(err error) error {
				return fmt.Errorf("read embedded %s: %w", globalPath, err)
			}),
			IOE.Chain(func(content []byte) IOE.IOEither[error, []byte] {
				return F.Pipe1(
					FILE.MkdirAll(filepath.Dir(destAbs), dirPerm),
					IOE.ChainTo[string](F.Pipe1(content, FILE.WriteFile(destAbs, filePerm))),
				)
			}),
		)
	}
}

// writeGlobalFiles mirrors every embedded global file into the kit temp dir,
// threading genState through unchanged on success.
func writeGlobalFiles(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe2(
		globalFilePaths(),
		IOE.Chain(IOE.TraverseArray[error](writeOneGlobalFile(gs.tempDir))),
		IOE.Map[error](F.Constant1[[][]byte](gs)),
	)
}

// makeTempDir creates a temp directory and records it in genState.
func makeTempDir(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return os.MkdirTemp("", "opencode-sbx-*")
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}),
		IOE.Map[error](func(tempDir string) genState {
			return genState{input: gs.input, spec: gs.spec, tempDir: tempDir}
		}),
	)
}

// writeSpecFile writes the rendered spec.yaml into the temp dir.
func writeSpecFile(gs genState) IOE.IOEither[error, genState] {
	return F.Pipe2(
		[]byte(gs.spec),
		FILE.WriteFile(F.Pipe1(gs.tempDir, file.Join("spec.yaml")), filePerm),
		IOE.Map[error](F.Constant1[[]byte](gs)),
	)
}

// buildExecState resolves OutputPath to absolute and constructs execState.
func buildExecState(gs genState) IOE.IOEither[error, execState] {
	return F.Pipe2(
		IOE.TryCatchError(func() (string, error) {
			return filepath.Abs(gs.input.OutputPath)
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("resolve output path: %w", err)
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

// generateToTempDir renders the spec, creates a temp dir, writes spec.yaml and
// every embedded global file, then assembles the execState.
func generateToTempDir(input Input) IOE.IOEither[error, execState] {
	return F.Pipe5(
		input,
		GenerateSpec,
		IOE.Chain(makeTempDir),
		IOE.Chain(writeSpecFile),
		IOE.Chain(writeGlobalFiles),
		IOE.Chain(buildExecState),
	)
}

// runSbxCommand delegates to the shared sbxexec runner, adapting the local
// CommandSpec type to the shared one.
func runSbxCommand(spec CommandSpec) IOE.IOEither[error, F.Void] {
	return sbxexec.RunSbxCommand(sbxexec.CommandSpec{
		Ctx:   spec.Ctx,
		Bin:   spec.Bin,
		Args:  spec.Args,
		Stdin: spec.Stdin,
	})
}

// validateKit runs "sbx kit validate" against the temp dir.
func validateKit(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe1(
		ss.Run(CommandSpec{
			Ctx:  ss.State.Ctx,
			Bin:  sbxBinary,
			Args: []string{"kit", "validate", ss.State.TempDir},
		}),
		IOE.Map[error](F.Constant1[F.Void](ss)),
	)
}

// withPackedResult returns a new stepState with the packed kit result recorded.
func withPackedResult(ss stepState) stepState {
	return stepState{
		Run: ss.Run,
		State: execState{
			Input:      ss.State.Input,
			TempDir:    ss.State.TempDir,
			OutputPath: ss.State.OutputPath,
			KitName:    ss.State.KitName,
			APIKey:     ss.State.APIKey,
			Result: KitResult{
				OutputPath: ss.State.OutputPath,
				KitName:    agentKitName,
				Created:    ss.State.Result.Created,
			},
		},
	}
}

// markCreated returns a new stepState with Result.Created set true.
func markCreated(ss stepState) stepState {
	return stepState{
		Run: ss.Run,
		State: execState{
			Input:      ss.State.Input,
			TempDir:    ss.State.TempDir,
			OutputPath: ss.State.OutputPath,
			KitName:    ss.State.KitName,
			APIKey:     ss.State.APIKey,
			Result: KitResult{
				OutputPath: ss.State.Result.OutputPath,
				KitName:    ss.State.Result.KitName,
				Created:    true,
			},
		},
	}
}

// packKit runs "sbx kit pack" to produce the output zip.
func packKit(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe1(
		ss.Run(CommandSpec{
			Ctx:  ss.State.Ctx,
			Bin:  sbxBinary,
			Args: []string{"kit", "pack", ss.State.TempDir, "-o", ss.State.OutputPath},
		}),
		IOE.Map[error](F.Constant1[F.Void](withPackedResult(ss))),
	)
}

// createWithSecret gates on ShouldCreate; when true it stores the API key
// secret under the provider ID, then creates the sandbox instance.
func createWithSecret(ss stepState) IOE.IOEither[error, stepState] {
	return F.Pipe2(
		ss.State.ShouldCreate,
		O.FromPredicate(F.Identity[bool]),
		O.Fold(
			func() IOE.IOEither[error, stepState] { return IOE.Of[error](ss) },
			func(_ bool) IOE.IOEither[error, stepState] {
				return F.Pipe2(
					ss.Run(CommandSpec{
						Ctx:   ss.State.Ctx,
						Bin:   sbxBinary,
						Args:  []string{"secret", "set", "-g", ss.State.Provider},
						Stdin: ss.State.APIKey + "\n",
					}),
					IOE.Chain(func(F.Void) IOE.IOEither[error, F.Void] {
						return ss.Run(CommandSpec{
							Ctx:  ss.State.Ctx,
							Bin:  sbxBinary,
							Args: []string{createCmd, agentKitName, "--kit", ss.State.OutputPath},
						})
					}),
					IOE.Map[error](F.Constant1[F.Void](markCreated(withPackedResult(ss)))),
				)
			},
		),
	)
}

// toStepState is a curried constructor converting execState to stepState.
var toStepState = F.Curry2(func(run processRunner, es execState) stepState {
	return stepState{State: es, Run: run}
})

// Execute runs the full pipeline: generate → validate → pack → optionally
// store secret + create.
func Execute(input Input) IOE.IOEither[error, KitResult] {
	return F.Pipe6(
		input,
		generateToTempDir,
		IOE.Map[error](toStepState(runSbxCommand)),
		IOE.Chain(validateKit),
		IOE.Chain(packKit),
		IOE.Chain(createWithSecret),
		IOE.Map[error](func(ss stepState) KitResult { return ss.State.Result }),
	)
}

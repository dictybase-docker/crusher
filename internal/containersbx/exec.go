package containersbx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// Execute runs the full pipeline: generate → validate → pack → optionally create → cleanup.
func Execute(input Input) IOE.IOEither[error, KitResult] {
	return F.Pipe6(
		generateToTempDir(input),
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

// generateToTempDir reads config/skills, renders spec, writes to os.MkdirTemp.
func generateToTempDir(input Input) IOE.IOEither[error, execState] {
	return F.Pipe1(
		ReadConfig(input.ConfigPath),
		IOE.Chain(func(configContent string) IOE.IOEither[error, execState] {
			return F.Pipe1(
				ReadSkills(input.SkillsPath),
				IOE.Chain(func(skills map[string]string) IOE.IOEither[error, execState] {
					return F.Pipe1(
						IOE.TryCatchError(func() (string, error) {
							return GenerateSpec(input, configContent, skills)
						}),
						IOE.Chain(func(spec string) IOE.IOEither[error, execState] {
							return IOE.TryCatchError(func() (execState, error) {
								tempDir, err := os.MkdirTemp("", "crush-sbx-*")
								if err != nil {
									return execState{}, fmt.Errorf(
										"failed to create temp dir: %w",
										err,
									)
								}
								specPath := filepath.Join(tempDir, "spec.yaml")
								writeSpec := func() error {
									return os.WriteFile(specPath, []byte(spec), filePerm)
								}
								if err := writeSpec(); err != nil {
									os.RemoveAll(tempDir)
									return execState{}, fmt.Errorf(
										"failed to write spec.yaml: %w",
										err,
									)
								}
								absOutput, err := filepath.Abs(input.OutputPath)
								if err != nil {
									os.RemoveAll(tempDir)
									return execState{}, fmt.Errorf(
										"failed to resolve output path: %w",
										err,
									)
								}
								return execState{
									Input:      input,
									TempDir:    tempDir,
									OutputPath: absOutput,
									KitName:    input.KitName,
									APIKey:     input.APIKey,
								}, nil
							})
						}),
					)
				}),
			)
		}),
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
func createSandboxOrSkip(state execState) IOE.IOEither[error, execState] {
	if !state.ShouldCreate {
		return IOE.Of[error](state)
	}
	return F.Pipe1(
		runSbxCommand(CommandSpec{
			Bin: sbxBinary,
			Args: []string{
				"create", state.KitName,
				"--kit", state.OutputPath,
			},
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

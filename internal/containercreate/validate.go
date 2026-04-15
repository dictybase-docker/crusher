package containercreate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	Eq "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	iof "github.com/IBM/fp-go/v2/ioeither/file"
	O "github.com/IBM/fp-go/v2/option"
	Pred "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"

	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
)

var (
	// isBlank is true when the input becomes empty after trimming whitespace.
	isBlank = F.Pipe1(
		Str.IsEmpty,
		Pred.ContraMap(strings.TrimSpace),
	)

	// isNonBlank is the negation of isBlank.
	isNonBlank = Pred.Not(isBlank)

	// validBasenameRegex ensures basename has at least one letter or digit.
	validBasenameRegex = regexp.MustCompile(`[a-zA-Z0-9]`)

	// EqString is the Eq instance for strings.
	EqString = Eq.FromStrictEquals[string]()
)

// ============================================================================
// Reserved basenames
// ============================================================================

// isReservedBasename checks if basename is reserved.
func isReservedBasename(basename string) bool {
	return F.Pipe2(
		A.From("config", "data", "crush"),
		A.FindFirst(func(s string) bool {
			return EqString.Equals(s, basename)
		}),
		O.IsSome,
	)
}

// ============================================================================
// Normalization (pure, resolves defaults before validation)
// ============================================================================

// NormalizeInput resolves all defaults, producing a fully-specified Input.
// After this, ImageName, ContainerName, and WorkspacePath are guaranteed non-blank.
func NormalizeInput(input Input) Input {
	return Input{
		ImageName: F.Pipe2(
			input.ImageName,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return DefaultImageName }),
		),
		ContainerName: F.Pipe2(
			input.ContainerName,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(GenerateName),
		),
		WorkspacePath: F.Pipe2(
			input.WorkspacePath,
			O.FromPredicate(isNonBlank),
			O.GetOrElse(func() string { return "." }),
		),
		ConfigPath: input.ConfigPath,
		DataPath:   input.DataPath,
		APIKey:      input.APIKey,
		GitHubToken: input.GitHubToken,
		Volumes:     input.Volumes,
		Ctx:         input.Ctx,
	}
}

// ============================================================================
// Validation Functions (pure, using fp-go Either API)
// ============================================================================

// ValidateInput normalizes defaults, validates paths, and resolves to absolute form.
// Returns Either[error, ResolvedInput].
func ValidateInput(input Input) E.Either[error, ResolvedInput] {
	return F.Pipe6(
		E.Of[error](input),
		E.Map[error](NormalizeInput),
		E.Chain(resolveConfigPath),
		E.Chain(resolveDataPath),
		E.Chain(resolveWorkspace),
		E.Chain(validateVolumes),
		E.Map[error](buildResolvedInput),
	)
}

// resolveConfigPath resolves the config path to absolute.
func resolveConfigPath(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.ConfigPath,
		resolveAbsolutePath,
		E.Map[error](func(p string) Input {
			return Input{
				ImageName:     input.ImageName,
				ContainerName: input.ContainerName,
				ConfigPath:    p,
				DataPath:      input.DataPath,
				APIKey:        input.APIKey,
				GitHubToken:   input.GitHubToken,
				WorkspacePath: input.WorkspacePath,
				Volumes:       input.Volumes,
				Ctx:           input.Ctx,
			}
		}),
	)
}

// resolveDataPath resolves the data path to absolute.
func resolveDataPath(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.DataPath,
		resolveAbsolutePath,
		E.Map[error](func(p string) Input {
			return Input{
				ImageName:     input.ImageName,
				ContainerName: input.ContainerName,
				ConfigPath:    input.ConfigPath,
				DataPath:      p,
				APIKey:        input.APIKey,
				GitHubToken:   input.GitHubToken,
				WorkspacePath: input.WorkspacePath,
				Volumes:       input.Volumes,
				Ctx:           input.Ctx,
			}
		}),
	)
}

// resolveWorkspace resolves the workspace path to absolute.
func resolveWorkspace(input Input) E.Either[error, Input] {
	return F.Pipe1(
		resolveAbsolutePath(input.WorkspacePath),
		E.Map[error](func(workspace string) Input {
			return Input{
				ImageName:     input.ImageName,
				ContainerName: input.ContainerName,
				ConfigPath:    input.ConfigPath,
				DataPath:      input.DataPath,
				APIKey:        input.APIKey,
				GitHubToken:   input.GitHubToken,
				WorkspacePath: workspace,
				Volumes:       input.Volumes,
				Ctx:           input.Ctx,
			}
		}),
	)
}

// validateVolumes validates and resolves all additional volume paths.
// Skips validation entirely when the volumes slice is empty.
func validateVolumes(input Input) E.Either[error, Input] {
	return F.Pipe2(
		input.Volumes,
		O.FromPredicate(func(vols []string) bool { return len(vols) > 0 }),
		O.Fold(
			func() E.Either[error, Input] { return E.Of[error](input) },
			func(vols []string) E.Either[error, Input] {
				return F.Pipe3(
					vols,
					A.Map(validateVolumePath),
					E.SequenceArray[error, string],
					E.Map[error](func(volumes []string) Input {
						return Input{
							ImageName:     input.ImageName,
							ContainerName: input.ContainerName,
							ConfigPath:    input.ConfigPath,
							DataPath:      input.DataPath,
							APIKey:        input.APIKey,
							GitHubToken:   input.GitHubToken,
							WorkspacePath: input.WorkspacePath,
							Volumes:       volumes,
							Ctx:           input.Ctx,
						}
					}),
				)
			},
		),
	)
}

// validateVolumePath validates a single volume path.
func validateVolumePath(vol string) E.Either[error, string] {
	return F.Pipe3(
		vol,
		E.FromPredicate(
			isNonBlank,
			func(string) error {
				return errors.New("volume path cannot be blank")
			},
		),
		E.Chain(resolveAbsolutePath),
		E.Chain(E.FromPredicate(
			F.Pipe2(
				Pred.Not(isReservedBasename),
				Pred.And(validBasenameRegex.MatchString),
				Pred.ContraMap(filepath.Base),
			),
			func(bp string) error {
				return fmt.Errorf(
					"volume basename '%s' is reserved or invalid",
					filepath.Base(bp),
				)
			},
		)),
	)
}

// ============================================================================
// Path Resolution Helpers (pure functions)
// ============================================================================

// resolveAbsolutePath resolves a path to absolute form and validates it exists.
func resolveAbsolutePath(path string) E.Either[error, string] {
	return F.Pipe3(
		IOE.TryCatchError(func() (string, error) {
			return filepath.Abs(path)
		}),
		IOE.Chain(func(abs string) IOE.IOEither[error, string] {
			return F.Pipe1(
				iof.Stat(abs),
				IOE.Map[error](func(_ os.FileInfo) string {
					return abs
				}),
			)
		}),
		IOE.MapLeft[string](func(err error) error {
			return fmt.Errorf("path validation failed: %w", err)
		}),
		FP.ToEither[error, string],
	)
}

// ============================================================================
// Build ResolvedInput (pure transformation)
// ============================================================================

func buildResolvedInput(input Input) ResolvedInput {
	return F.Pipe5(
		input.Volumes,
		A.Map(func(vol string) MountSpec {
			return MountSpec{
				HostPath:   vol,
				TargetPath: pathJoin.Concat(ContainerHome, filepath.Base(vol)),
				Readonly:   true,
			}
		}),
		A.Push(MountSpec{
			HostPath:   input.ConfigPath,
			TargetPath: ConfigTarget,
			Readonly:   false,
		}),
		A.Push(MountSpec{
			HostPath:   input.DataPath,
			TargetPath: DataTarget,
			Readonly:   false,
		}),
		A.Push(MountSpec{
			HostPath:   input.WorkspacePath,
			TargetPath: WorkspaceTarget,
			Readonly:   false,
		}),
		func(mspec []MountSpec) ResolvedInput {
			return ResolvedInput{
				ImageName:     input.ImageName,
				ContainerName: input.ContainerName,
				Mounts:        mspec,
				Workdir:       WorkspaceTarget,
				APIKey:        input.APIKey,
				GitHubToken:   input.GitHubToken,
			}
		},
	)
}

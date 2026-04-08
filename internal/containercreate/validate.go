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
	Ord "github.com/IBM/fp-go/v2/ord"
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

	// OrdMountSpecByTarget is the Ord instance for MountSpec sorted by TargetPath.
	OrdMountSpecByTarget = Ord.Contramap(
		func(m MountSpec) string { return m.TargetPath },
	)(
		Str.Ord,
	)
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
		Volumes:    input.Volumes,
		Ctx:        input.Ctx,
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

// buildResolvedInput constructs the ResolvedInput with mount specifications.
func buildResolvedInput(input Input) ResolvedInput {
	return ResolvedInput{
		ImageName:     input.ImageName,
		ContainerName: input.ContainerName,
		Mounts:        buildMounts(input),
		Workdir:       buildWorkdir(),
	}
}

// buildMounts constructs the mount specifications in order.
func buildMounts(input Input) []MountSpec {
	return sortAdditionalMounts(A.ArrayConcatAll(
		buildCoreMounts(input),
		buildWorkspaceMount(input.WorkspacePath),
		buildVolumeMounts(input.Volumes),
	))
}

// buildCoreMounts constructs config and data mounts.
func buildCoreMounts(input Input) []MountSpec {
	return []MountSpec{
		{HostPath: input.ConfigPath, TargetPath: ConfigTarget, Readonly: true},
		{HostPath: input.DataPath, TargetPath: DataTarget, Readonly: false},
	}
}

// buildWorkspaceMount constructs workspace mount (always present after normalization).
func buildWorkspaceMount(workspacePath string) []MountSpec {
	return []MountSpec{{
		HostPath:   workspacePath,
		TargetPath: WorkspaceTarget,
		Readonly:   false,
	}}
}

// buildVolumeMounts constructs additional volume mounts (all read-only).
func buildVolumeMounts(volumes []string) []MountSpec {
	return F.Pipe2(
		volumes,
		A.Filter(isNonBlank),
		A.Map(func(vol string) MountSpec {
			return MountSpec{
				HostPath:   vol,
				TargetPath: ContainerHome + "/" + filepath.Base(vol),
				Readonly:   true,
			}
		}),
	)
}

// sortAdditionalMounts sorts mounts by TargetPath.
func sortAdditionalMounts(mounts []MountSpec) []MountSpec {
	return A.Sort(OrdMountSpecByTarget)(mounts)
}

// buildWorkdir returns the working directory (always WorkspaceTarget after normalization).
func buildWorkdir() string {
	return WorkspaceTarget
}

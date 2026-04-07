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
	P "github.com/IBM/fp-go/v2/pair"
	Pred "github.com/IBM/fp-go/v2/predicate"
	R "github.com/IBM/fp-go/v2/record"
	Str "github.com/IBM/fp-go/v2/string"

	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
)

// ============================================================================
// Predicates (using fp-go predicate API)
// ============================================================================

// isBlank is true when the input becomes empty after trimming whitespace.
var isBlank = F.Pipe1(
	Str.IsEmpty,
	Pred.ContraMap(strings.TrimSpace),
)

// isNonBlank is the negation of isBlank.
var isNonBlank = Pred.Not(isBlank)

// containerNameRegex validates container names.
// Must start with a letter, followed by letters, digits, dashes, or underscores.
var containerNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// isValidContainerName checks if name matches container naming conventions.
var isValidContainerName = F.Pipe1(
	Pred.Not(isBlank),
	Pred.And(func(name string) bool {
		return containerNameRegex.MatchString(name)
	}),
)

// validBasenameRegex ensures basename has at least one letter or digit.
var validBasenameRegex = regexp.MustCompile(`[a-zA-Z0-9]`)

// hasValidBasename checks if basename contains at least one letter or digit.
var hasValidBasename = F.Pipe1(
	Pred.Not(isBlank),
	Pred.And(func(basename string) bool {
		return validBasenameRegex.MatchString(basename)
	}),
)

// ============================================================================
// Eq and Ord instances (using fp-go Eq and Ord API)
// ============================================================================

// EqString is the Eq instance for strings.
var EqString = Eq.FromStrictEquals[string]()

// OrdString is the Ord instance for strings (lexicographic ordering).
var OrdString = Str.Ord

// OrdMountSpecByTarget is the Ord instance for MountSpec sorted by TargetPath.
var OrdMountSpecByTarget = Ord.Contramap(
	func(m MountSpec) string { return m.TargetPath },
)(
	OrdString,
)

// ============================================================================
// Reserved basenames (using fp-go Record API for lookup)
// ============================================================================

// reservedBasenames is a Record of reserved mount target basenames.
var reservedBasenames = F.Pipe2(
	[]string{"config", "data", "crush"},
	A.Map(func(s string) P.Pair[string, bool] {
		return P.MakePair(s, true)
	}),
	R.FromEntries[string, bool],
)

// isReservedBasename checks if basename is reserved using Record lookup.
func isReservedBasename(basename string) bool {
	return F.Pipe2(
		reservedBasenames,
		R.Lookup[bool](basename),
		O.IsSome,
	)
}

// isValidVolumeBasename checks if basename is valid for volume mount.
var isValidVolumeBasename = F.Pipe1(
	Pred.Not(isReservedBasename),
	Pred.And(hasValidBasename),
)

// ============================================================================
// Validation Functions (pure, using fp-go Either API)
// ============================================================================

// ValidateInput validates the Input and resolves all paths to absolute form.
// Returns Either[error, ResolvedInput].
func ValidateInput(input Input) E.Either[error, ResolvedInput] {
	return F.Pipe5(
		E.Of[error](input),
		E.Chain(resolveConfigPath),
		E.Chain(resolveDataPath),
		E.Chain(resolveWorkspaceAndName),
		E.Chain(validateVolumes),
		E.Map[error](buildResolvedInput),
	)
}

// resolveConfigPath resolves the config path to absolute.
func resolveConfigPath(input Input) E.Either[error, Input] {
	return F.Pipe1(
		resolveAbsolutePath(input.ConfigPath),
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
	return F.Pipe1(
		resolveAbsolutePath(input.DataPath),
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

// resolveWorkspaceAndName resolves workspace path and validates container name.
func resolveWorkspaceAndName(input Input) E.Either[error, Input] {
	return F.Pipe1(
		validateContainerName(input.ContainerName),
		E.Chain(func(string) E.Either[error, Input] {
			return F.Pipe1(
				resolveOptionalPath(input.WorkspacePath),
				E.Map[error](func(workspace string) Input {
					return Input{
						ImageName:     resolveImageName(input.ImageName),
						ContainerName: input.ContainerName,
						ConfigPath:    input.ConfigPath,
						DataPath:      input.DataPath,
						WorkspacePath: workspace,
						Volumes:       input.Volumes,
						Ctx:           input.Ctx,
					}
				}),
			)
		}),
	)
}

// validateVolumes validates and resolves all additional volume paths.
func validateVolumes(input Input) E.Either[error, Input] {
	return F.Pipe3(
		A.Map(validateVolumePath)(input.Volumes),
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
		F.Identity[E.Either[error, Input]],
	)
}

// validateVolumePath validates a single volume path.
func validateVolumePath(vol string) E.Either[error, string] {
	return F.Pipe3(
		E.Of[error](vol),
		E.Chain(func(v string) E.Either[error, string] {
			return E.FromPredicate(
				isNonBlank,
				func(string) error { return errors.New("volume path cannot be blank") },
			)(v)
		}),
		E.Chain(resolveAbsolutePath),
		E.Chain(validateVolumeBasename),
	)
}

// validateVolumeBasename validates the basename of a volume path.
func validateVolumeBasename(absPath string) E.Either[error, string] {
	basename := filepath.Base(absPath)
	return F.Pipe1(
		E.FromPredicate(
			isValidVolumeBasename,
			func(string) error {
				return errors.New("volume basename '" + basename + "' is reserved or invalid")
			},
		)(basename),
		E.Map[error](func(string) string { return absPath }),
	)
}

// validateContainerName validates container name if provided.
func validateContainerName(name string) E.Either[error, string] {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(name),
		O.Fold(
			func() E.Either[error, string] { return E.Of[error](name) },
			func(n string) E.Either[error, string] {
				return E.FromPredicate(
					isValidContainerName,
					func(string) error {
						return errors.New(
							"container name must start with a letter and contain only letters, digits, dashes, or underscores",
						)
					},
				)(n)
			},
		),
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

// resolveOptionalPath resolves an optional path (blank = skip).
func resolveOptionalPath(path string) E.Either[error, string] {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(path),
		O.Fold(
			func() E.Either[error, string] { return E.Of[error]("") },
			resolveAbsolutePath,
		),
	)
}

// resolveImageName returns the default image name if blank.
func resolveImageName(name string) string {
	return F.Pipe2(
		O.FromPredicate(isNonBlank)(name),
		O.GetOrElse(func() string { return DefaultImageName }),
		F.Identity[string],
	)
}

// ============================================================================
// Build ResolvedInput (pure transformation)
// ============================================================================

// buildResolvedInput constructs the ResolvedInput with mount specifications.
func buildResolvedInput(input Input) ResolvedInput {
	return ResolvedInput{
		ImageName:     input.ImageName,
		ContainerName: resolveContainerName(input.ContainerName),
		Mounts:        buildMounts(input),
		Workdir:       buildWorkdir(input.WorkspacePath),
	}
}

// resolveContainerName returns the provided name or generates a new one.
func resolveContainerName(name string) string {
	return F.Pipe2(
		O.FromPredicate(isNonBlank)(name),
		O.GetOrElse(GenerateName),
		F.Identity[string],
	)
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

// buildWorkspaceMount constructs workspace mount if path is provided.
func buildWorkspaceMount(workspacePath string) []MountSpec {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(workspacePath),
		O.Fold(
			func() []MountSpec { return []MountSpec{} },
			func(p string) []MountSpec {
				return []MountSpec{{
					HostPath:   p,
					TargetPath: WorkspaceTarget,
					Readonly:   false,
				}}
			},
		),
	)
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

// buildWorkdir returns the working directory if workspace is mounted.
func buildWorkdir(workspacePath string) string {
	return F.Pipe3(
		O.FromPredicate(isNonBlank)(workspacePath),
		O.Map(func(string) string { return WorkspaceTarget }),
		O.GetOrElse(func() string { return "" }),
		F.Identity[string],
	)
}

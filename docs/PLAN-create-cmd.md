
# PLAN: Create Subcommand for Crush Sandbox Container

## 1. Executive Summary

This document provides an exhaustive implementation plan for a new `create` subcommand in `container-cli` that creates a container from an image built by the `build` subcommand. The container is pre-configured to run the Crush AI coding assistant with proper volume mounts for configuration, data, and workspace access.

---

## 2. Objectives

### 2.1 Primary Goals

1. **Create containers for Crush sandbox** — Enable users to create isolated, reproducible environments for running the Crush AI assistant
2. **Sensible defaults** — Default image name matches the `build` subcommand (`crusher:latest`)
3. **Safe configuration** — Config and additional volumes are read-only to prevent accidental corruption
4. **Flexible workspace** — Support optional workspace mounting with configurable paths
5. **Extensibility** — Allow additional volume mounts for project-specific needs (read-only)
6. **Clear user feedback** — Output the exact `container` command needed to start the created container

### 2.2 Non-Goals (Out of Scope)

- Starting or running the container (user does this manually)
- Managing container lifecycle (stop, delete, etc.)
- Building images (handled by `build` subcommand)
- Network configuration
- Resource limits (CPU, memory)
- Multiple workspace mounts

---

## 3. CLI Surface Design

### 3.1 Command Signature

```bash
container-cli create [options] --config <path> --data <path>
```

### 3.2 Required Flags

| Flag | Type | Description |
|------|------|-------------|
| `--config`, `-c` | string | Host folder containing `crush.json` configuration. Mounted read-only to `${HOME}/crush/config` |
| `--data`, `-d` | string | Host folder for Crush persistent data. Mounted read-write to `${HOME}/crush/data` |

### 3.3 Optional Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--image`, `-i` | string | `crusher:latest` | Image name (should match build subcommand output) |
| `--workspace`, `-w` | string | `.` (current directory) | Host workspace folder. Mounted read-write to `${HOME}/workspace`. Also set as working directory |
| `--volume`, `-v` | stringSlice | `[]` | Additional host folders to mount read-only. Mounted to `${HOME}/<basename>` |
| `--name`, `-n` | string | auto-generated | Container name. If not provided, generates a random alphabetical name |

### 3.4 Usage Examples

#### Minimal invocation (required parameters only)

```bash
container-cli create \
  --config ~/.config/crush \
  --data ~/.local/share/crush
```

Output:
```
Container created: mystifying-hoover

To start with an interactive shell:
  container start -it mystifying-hoover
```

#### With workspace mount

```bash
container-cli create \
  --config ~/.config/crush \
  --data ~/.local/share/crush \
  --workspace ~/Projects/myapp
```

#### With additional volume mounts (read-only)

```bash
container-cli create \
  --config ~/.config/crush \
  --data ~/.local/share/crush \
  --workspace ~/Projects/myapp \
  --volume ~/Projects/shared-libs \
  --volume ~/.ssh
```

#### With custom image and name

```bash
container-cli create \
  --config ~/.config/crush \
  --data ~/.local/share/crush \
  --image my-crusher:v2.0.0 \
  --name my-crush-sandbox
```

---

## 4. Domain Model

### 4.1 Core Types

```go
// Package containercreate implements the "create" subcommand, which creates a
// container for running the Crush AI assistant with pre-configured mounts.
package containercreate

import (
	"context"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// MountSpec represents a single volume mount specification.
type MountSpec struct {
	HostPath   string // Absolute path on host
	TargetPath string // Absolute path inside container
	Readonly   bool   // Whether mount is read-only
}

// Input holds all parameters for container creation.
type Input struct {
	ImageName     string            // Image name (e.g., "crusher:latest")
	ContainerName string            // Container name (empty = auto-generate)
	ConfigPath    string            // Host path to Crush config directory (required)
	DataPath      string            // Host path to Crush data directory (required)
	WorkspacePath string            // Host path to workspace (optional)
	Volumes       []string          // Additional host paths to mount (read-only)
	Ctx           context.Context
}

// CommandSpec holds the resolved executable binary and argv slice.
// Identical to containerbuild.CommandSpec for consistency.
type CommandSpec struct {
	Bin  string   // "container"
	Args []string // Full argument list for "container create"
}

// ContainerResult holds the result of a successful container creation.
type ContainerResult struct {
	Name string // The container name
}

// ResolvedInput is the validated and resolved input with absolutized paths.
type ResolvedInput struct {
	ImageName     string
	ContainerName string
	Mounts        []MountSpec
	Workdir       string // Working directory inside container (empty if no workspace)
}
```

### 4.2 Container Paths

The following paths are used inside the container:

| Purpose | Container Path | Environment Variable | Mount Mode |
|---------|---------------|---------------------|------------|
| Crush config | `${HOME}/crush/config` | `CRUSH_GLOBAL_CONFIG` | read-only |
| Crush data | `${HOME}/crush/data` | `CRUSH_GLOBAL_DATA` | read-write |
| Workspace | `${HOME}/workspace` | — | read-write |
| Additional volumes | `${HOME}/<basename>` | — | **read-only** |

The `${HOME}` variable inside the container is typically `/home/agent` (from the Dockerfile's `USER agent` directive).

### 4.3 Environment Variables Set in Container

```go
var crushEnvVars = []string{
	"CRUSH_GLOBAL_CONFIG=/home/agent/crush/config",
	"CRUSH_GLOBAL_DATA=/home/agent/crush/data",
}
```

---

## 5. Functional Programming Conventions

This subcommand follows the same fp-go conventions as the `build` subcommand:

### 5.1 Core Rules

1. **No imperative branching** — Use `E.Fold`, `E.FromPredicate`, `E.Chain`, `O.Fold`
2. **Side effects isolated in IOEither** — `exec.CommandContext` receives direct argv slice
3. **Validation is pure** — Separate from process execution
4. **Use `F.Pipe1/2/3/etc`** — Match arity to transform count
5. **Use fp-go APIs for all operations** — Array, Predicate, Ord, Eq, Option, Record

### 5.2 Key fp-go Combinators Used

| Combinator | Package | Purpose |
|------------|---------|---------|
| `E.FromPredicate` | `either` | Lift value into Either based on predicate |
| `E.Chain` | `either` | Sequence dependent Either operations |
| `E.Fold` | `either` | Terminal branching at boundary |
| `E.SequenceArray` | `either` | Convert `[]Either` to `Either[]` |
| `IOE.FromEither` | `ioeither` | Lift Either into IOEither |
| `IOE.Chain` | `ioeither` | Sequence IOEither operations |
| `IOE.TryCatchError` | `ioeither` | Wrap fallible effect |
| `A.Map` | `array` | Transform array elements |
| `A.Filter` | `array` | Filter array elements |
| `A.Sort` | `array` | Sort array with Ord |
| `A.Chain` | `array` | FlatMap over arrays |
| `A.Flatten` | `array` | Flatten nested arrays |
| `A.ArrayConcatAll` | `array` | Concatenate multiple arrays |
| `P.Predicate` | `predicate` | Predicate combinators |
| `P.Not` | `predicate` | Negate a predicate |
| `P.And` / `P.Or` | `predicate` | Combine predicates |
| `P.ContraMap` | `predicate` | Transform predicate input |
| `Ord.Ord` | `ord` | Ordering for sorting |
| `Ord.Contramap` | `ord` | Transform Ord input |
| `Eq.Eq` | `eq` | Equality comparison |
| `O.Of` / `O.None` | `option` | Optional values |
| `O.Fold` | `option` | Terminal branching on Option |
| `O.GetOrElse` | `option` | Extract with default |
| `R.FromEntries` | `record` | Build map from entries |
| `R.Lookup` | `record` | Lookup in map |
| `Str.IsEmpty` | `string` | String predicates |

### 5.3 Import Aliases

Following AGENTS.md conventions:

```go
import (
	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	A "github.com/IBM/fp-go/v2/array"
	Str "github.com/IBM/fp-go/v2/string"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	Ord "github.com/IBM/fp-go/v2/ord"
	Eq "github.com/IBM/fp-go/v2/eq"
	R "github.com/IBM/fp-go/v2/record"
	Pkg "github.com/IBM/fp-go/v2/pair"
)
```

---

## 6. Implementation Details

### 6.1 File Structure

```
internal/containercreate/
├── command.go       # CLI flags, InputFromCommand, Action
├── input.go         # Input, ResolvedInput, MountSpec types
├── validate.go      # Pure Either-based validation (fp-go only)
├── mounts.go        # Mount rendering logic (fp-go only)
├── namegen.go       # Random container name generation
├── args.go          # CommandSpec rendering (fp-go only)
├── exec.go          # IOEither-based process execution
└── *_test.go        # Unit tests
```

### 6.2 Validation Logic

The validation phase is pure and returns an `Either[error, ResolvedInput]`. **All operations use fp-go APIs — no imperative branching.**

#### Validation Rules

1. **Config path validation**
   - Must be non-blank
   - Must be an absolute path (resolve if relative)

2. **Data path validation**
   - Must be non-blank
   - Must be an absolute path (resolve if relative)

3. **Workspace path validation** (if provided)
   - Must be non-blank (if provided)
   - Must be an absolute path (resolve if relative)

4. **Volume paths validation**
   - Each path must be non-blank
   - Each path must be an absolute path
   - Basename must not be "config" or "data"
   - Basename must contain at least one letter or number

5. **Image name validation**
   - Must be non-blank (use default if not provided)

6. **Container name validation** (if provided)
   - Must match container naming conventions (alphanumeric, dash, underscore)
   - If not provided, will be auto-generated

### 6.3 Mount Rendering Logic

The mount rendering follows Apple's `container create` syntax:

```bash
--mount type=bind,source=<host-path>,target=<container-path>,readonly
```

For read-write mounts, the `readonly` suffix is omitted:

```bash
--mount type=bind,source=<host-path>,target=<container-path>
```

#### Mount Priority Order

Mounts are rendered in this order:

1. Config mount (read-only)
2. Data mount (read-write)
3. Workspace mount (read-write, if provided)
4. Additional volume mounts (**read-only**, sorted alphabetically by target path)

---

## 7. Complete Implementation Code

### 7.1 `internal/containercreate/input.go`

```go
// Package containercreate implements the "create" subcommand, which creates a
// container for running the Crush AI assistant with pre-configured mounts.
package containercreate

import (
	"context"
)

const (
	// ContainerHome is the home directory inside the container.
	ContainerHome = "/home/agent"

	// ConfigTarget is the config mount target inside the container.
	ConfigTarget = ContainerHome + "/crush/config"

	// DataTarget is the data mount target inside the container.
	DataTarget = ContainerHome + "/crush/data"

	// WorkspaceTarget is the workspace mount target inside the container.
	WorkspaceTarget = ContainerHome + "/workspace"

	// DefaultImageName is the default image name matching the build subcommand.
	DefaultImageName = "crusher:latest"
)

// MountSpec represents a single volume mount specification.
type MountSpec struct {
	HostPath   string // Absolute path on host
	TargetPath string // Absolute path inside container
	Readonly   bool   // Whether mount is read-only
}

// Input holds all parameters for container creation.
type Input struct {
	ImageName     string            // Image name (e.g., "crusher:latest")
	ContainerName string            // Container name (empty = auto-generate)
	ConfigPath    string            // Host path to Crush config directory (required)
	DataPath      string            // Host path to Crush data directory (required)
	WorkspacePath string            // Host path to workspace (optional)
	Volumes       []string          // Additional host paths to mount (read-only)
	Ctx           context.Context
}

// ResolvedInput is the validated and resolved input with absolutized paths.
type ResolvedInput struct {
	ImageName     string
	ContainerName string
	Mounts        []MountSpec
	Workdir       string // Working directory inside container (empty if no workspace)
}

// ContainerResult holds the result of a successful container creation.
type ContainerResult struct {
	Name string // The container name
}
```

### 7.2 `internal/containercreate/command.go`

```go
package containercreate

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

// InputFromCommand reads CLI flags and constructs the create Input.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		Ctx:           ctx,
		ImageName:     cmd.String("image"),
		ContainerName: cmd.String("name"),
		ConfigPath:    cmd.String("config"),
		DataPath:      cmd.String("data"),
		WorkspacePath: cmd.String("workspace"),
		Volumes:       cmd.StringSlice("volume"),
	}
}

// Command returns the CLI command definition for "create".
func Command() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a container for running Crush with pre-configured mounts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config",
				Aliases:  []string{"c"},
				Usage:    "Host folder containing crush.json configuration (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "data",
				Aliases:  []string{"d"},
				Usage:    "Host folder for Crush persistent data (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "image",
				Aliases: []string{"i"},
				Usage:   "Image name (should match build subcommand output)",
				Value:   DefaultImageName,
			},
			&cli.StringFlag{
				Name:    "workspace",
				Aliases: []string{"w"},
				Usage:   "Host workspace folder to mount (default: current directory)",
				Value:   ".",
			},
			&cli.StringSliceFlag{
				Name:    "volume",
				Aliases: []string{"v"},
				Usage:   "Additional host folders to mount read-only (repeatable)",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Container name (auto-generated if not provided)",
			},
		},
		Action: Action,
	}
}

// Action is the create subcommand entry point.
// Pipeline: validate input → resolve paths → render args → run process → print result.
func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, ContainerResult],
		E.Fold(
			F.Identity[error],
			func(res ContainerResult) error {
				printSuccessMessage(res.Name)
				return nil
			},
		),
	)
}

// printSuccessMessage outputs instructions for starting the container.
func printSuccessMessage(name string) {
	fmt.Printf("Container created: %s\n\n", name)
	fmt.Println("To start with an interactive shell:")
	fmt.Printf("  container start -it %s\n", name)
}
```

### 7.3 `internal/containercreate/validate.go`

```go
package containercreate

import (
	"errors"
	"path/filepath"
	"regexp"

	A "github.com/IBM/fp-go/v2/array"
	E "github.com/IBM/fp-go/v2/either"
	Eq "github.com/IBM/fp-go/v2/eq"
	F "github.com/IBM/fp-go/v2/function"
	Ord "github.com/IBM/fp-go/v2/ord"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	Str "github.com/IBM/fp-go/v2/string"
)

// ============================================================================
// Predicates (using fp-go predicate API)
// ============================================================================

// isBlank is true when the input becomes empty after trimming whitespace.
var isBlank = F.Pipe1(
	Str.IsEmpty,
	P.ContraMap(string.TrimSpace),
)

// isNonBlank is the negation of isBlank.
var isNonBlank = P.Not(isBlank)

// containerNameRegex validates container names.
// Must start with a letter, followed by letters, digits, dashes, or underscores.
var containerNameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

// isValidContainerName checks if name matches container naming conventions.
var isValidContainerName = P.And(
	P.Not(isBlank),
	func(name string) bool {
		return containerNameRegex.MatchString(name)
	},
)

// validBasenameRegex ensures basename has at least one letter or digit.
var validBasenameRegex = regexp.MustCompile(`[a-zA-Z0-9]`)

// hasValidBasename checks if basename contains at least one letter or digit.
var hasValidBasename = P.And(
	P.Not(isBlank),
	func(basename string) bool {
		return validBasenameRegex.MatchString(basename)
	},
)

// ============================================================================
// Eq and Ord instances (using fp-go Eq and Ord API)
// ============================================================================

// EqString is the Eq instance for strings.
var EqString = Eq.FromStrictEquals[string]()

// OrdString is the Ord instance for strings (lexicographic ordering).
var OrdString = Ord.FromCompare[string](func(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
})

// OrdMountSpecByTarget is the Ord instance for MountSpec sorted by TargetPath.
var OrdMountSpecByTarget = Ord.Contramap(
	func(m MountSpec) string { return m.TargetPath },
	OrdString,
)

// ============================================================================
// Reserved basenames (using fp-go Record API for lookup)
// ============================================================================

// reservedBasenames is a Record of reserved mount target basenames.
var reservedBasenames = F.Pipe2(
	[]string{"config", "data", "crush"},
	A.Map(func(s string) Pkg.Pair[string, bool] {
		return Pkg.MakePair(s, true)
	}),
	R.FromEntries[string, bool],
)

// isReservedBasename checks if basename is reserved using Record lookup.
func isReservedBasename(basename string) bool {
	return F.Pipe1(
		reservedBasenames,
		R.Lookup[bool](basename),
		O.IsSome,
	)
}

// ============================================================================
// Validation Functions (pure, using fp-go Either API)
// ============================================================================

// ValidateInput validates the Input and resolves all paths to absolute form.
// Returns Either[error, ResolvedInput].
func ValidateInput(input Input) E.Either[error, ResolvedInput] {
	return F.Pipe4(
		E.Of[error](input),
		E.Chain(validateRequiredPaths),
		E.Chain(resolvePaths),
		E.Chain(validateVolumes),
		E.Map(buildResolvedInput),
	)
}

// validateRequiredPaths checks that required paths are non-blank.
func validateRequiredPaths(input Input) E.Either[error, Input] {
	return F.Pipe2(
		[]E.Either[error, string]{
			E.FromPredicate(
				input.ConfigPath,
				isNonBlank,
				func(string) error { return errors.New("config path is required") },
			),
			E.FromPredicate(
				input.DataPath,
				isNonBlank,
				func(string) error { return errors.New("data path is required") },
			),
		},
		E.SequenceArray[string],
		E.MapTo[error](input),
	)
}

// resolvePaths resolves all paths to absolute form and validates container name.
func resolvePaths(input Input) E.Either[error, Input] {
	return F.Pipe3(
		E.Of[error](input),
		E.Chain(resolveConfigPath),
		E.Chain(resolveDataPath),
		E.Chain(resolveWorkspaceAndName),
	)
}

// resolveConfigPath resolves the config path to absolute.
func resolveConfigPath(input Input) E.Either[error, Input] {
	return F.Pipe1(
		resolveAbsolutePath(input.ConfigPath),
		E.Map(func(p string) Input {
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
		E.Map(func(p string) Input {
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
	return F.Pipe2(
		validateContainerName(input.ContainerName),
		E.Chain(func(string) E.Either[error, Input] {
			return F.Pipe1(
				resolveOptionalPath(input.WorkspacePath),
				E.Map(func(workspace string) Input {
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
	return F.Pipe2(
		A.Map(validateVolumePath)(input.Volumes),
		E.SequenceArray[string],
		E.Map(func(volumes []string) Input {
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
}

// validateVolumePath validates a single volume path.
func validateVolumePath(vol string) E.Either[error, string] {
	return F.Pipe5(
		E.Of[error](vol),
		E.Filter(isNonBlank, func(string) error {
			return errors.New("volume path cannot be blank")
		}),
		E.Chain(resolveAbsolutePath),
		E.Chain(validateVolumeBasename),
		E.Map(EqString.Equals),
	)
}

// validateVolumeBasename validates the basename of a volume path.
func validateVolumeBasename(absPath string) E.Either[error, string] {
	basename := filepath.Base(absPath)
	return F.Pipe3(
		E.Of[error](basename),
		E.Filter(P.Not(isReservedBasename), func(string) error {
			return errors.New("volume basename '" + basename + "' is reserved")
		}),
		E.Filter(hasValidBasename, func(string) error {
			return errors.New("volume basename '" + basename + "' must contain at least one letter or digit")
		}),
		E.Map(func(string) string { return absPath }),
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
					n,
					isValidContainerName,
					func(string) error {
						return errors.New("container name must start with a letter and contain only letters, digits, dashes, or underscores")
					},
				)
			},
		),
	)
}

// ============================================================================
// Path Resolution Helpers (pure functions)
// ============================================================================

// resolveAbsolutePath resolves a path to absolute form.
// This is a pure function returning Either to allow for potential errors.
func resolveAbsolutePath(path string) E.Either[error, string] {
	return E.Of[error](filepath.Abs(path))
}

// resolveOptionalPath resolves an optional path (blank = skip).
func resolveOptionalPath(path string) E.Either[error, string] {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(path),
		O.Fold(
			func() E.Either[error, string] { return E.Of[error]("") },
			func(p string) E.Either[error, string] { return resolveAbsolutePath(p) },
		),
	)
}

// resolveImageName returns the default image name if blank.
func resolveImageName(name string) string {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(name),
		O.GetOrElse(func() string { return DefaultImageName }),
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
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(name),
		O.GetOrElse(GenerateName),
	)
}

// buildMounts constructs the mount specifications in order.
func buildMounts(input Input) []MountSpec {
	return F.Pipe1(
		A.ArrayConcatAll(
			buildCoreMounts(input),
			buildWorkspaceMount(input.WorkspacePath),
			buildVolumeMounts(input.Volumes),
		),
		sortAdditionalMounts,
	)
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
	return F.Pipe1(
		volumes,
		A.Filter(isNonBlank),
		A.Map(func(vol string) MountSpec {
			return MountSpec{
				HostPath:   vol,
				TargetPath: ContainerHome + "/" + filepath.Base(vol),
				Readonly:   true, // All additional volumes are read-only
			}
		}),
	)
}

// sortAdditionalMounts sorts mounts, keeping core mounts first.
func sortAdditionalMounts(mounts []MountSpec) []MountSpec {
	return F.Pipe1(
		mounts,
		A.Sort(OrdMountSpecByTarget),
	)
}

// buildWorkdir returns the working directory if workspace is mounted.
func buildWorkdir(workspacePath string) string {
	return F.Pipe1(
		O.FromPredicate(isNonBlank)(workspacePath),
		O.Map(func(string) string { return WorkspaceTarget }),
		O.GetOrElse(func() string { return "" }),
	)
}
```

### 7.4 `internal/containercreate/namegen.go`

```go
package containercreate

import (
	"crypto/rand"
	"math/big"
	"time"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	Str "github.com/IBM/fp-go/v2/string"
)

// adjectives for container names (Docker-style naming).
var adjectives = []string{
	"admiring", "adoring", "affectionate", "agitated", "amazing",
	"angry", "awesome", "beautiful", "blissful", "bold", "boring",
	"brave", "busy", "calm", "charming", "clever", "cool", "compassionate",
	"competent", "condescending", "confident", "cranky", "crazy", "curious",
	"dazzling", "determined", "distracted", "dreamy", "eager", "ecstatic",
	"elastic", "elated", "elegant", "eloquent", "epic", "exciting", "fervent",
	"festive", "flamboyant", "focused", "friendly", "frosty", "funny", "gallant",
	"gifted", "goofy", "gracious", "great", "groggy", "happy", "hardcore",
	"heuristic", "hopeful", "hungry", "infallible", "inspiring", "intelligent",
	"interesting", "introspective", "jolly", "jovial", "keen", "kind", "laughing",
	"loving", "lucid", "magical", "mystifying", "modest", "musing", "naughty",
	"nervous", "nice", "nifty", "nostalgic", "objective", "optimistic", "peaceful",
	"pedantic", "pensive", "practical", "priceless", "quirky", "quizzical",
	"recursing", "relaxed", "reverent", "romantic", "sad", "serene", "sharp",
	"silly", "sleepy", "stoic", "strange", "stupefied", "suspicious", "sweet",
	"tender", "thirsty", "trusting", "unruffled", "upbeat", "vibrant", "vigilant",
	"vigorous", "wizardly", "wonderful", "xenodochial", "youthful", "zealous", "zen",
}

// nouns for container names (famous scientists and engineers).
var nouns = []string{
	"albattani", "allen", "almeida", "antonelli", "archimedes", "ardinghelli",
	"aryabhata", "austin", "babbage", "banach", "bardeen", "bartik", "bassi",
	"beaver", "bell", "benz", "black", "blackwell", "bohr", "booth", "borg",
	"bose", "bouman", "boyd", "brahmagupta", "brattain", "brown", "buck",
	"burnell", "cannon", "carson", "cartwright", "carver", "cauchy", "cerf",
	"chandrasekhar", "chaplygin", "chatelet", "chatterjee", "chebyshev", "cohen",
	"chaum", "clarke", "colden", "cori", "cray", "curie", "curran", "darwin",
	"davinci", "dewdney", "dhawan", "diffie", "dijkstra", "dirac", "driscoll",
	"dubinsky", "easley", "edison", "einstein", "elbakyan", "elgamal", "elion",
	"ellis", "engelbart", "euclid", "euler", "faraday", "feistel", "fermat",
	"fermi", "feynman", "franklin", "gagarin", "galileo", "galois", "ganguly",
	"gates", "gauss", "germain", "goldberg", "goldstine", "goldwasser", "golick",
	"goodall", "gould", "greider", "grothendieck", "haibt", "hall", "hamilton",
	"haslett", "hawking", "hellman", "heisenberg", "hermann", "herschel", "hertz",
	"heyrovsky", "hodgkin", "hoover", "hopper", "hugle", "hypatia", "ishizaka",
	"jackson", "jang", "jemison", "jennings", "jepsen", "johnson", "joliot",
	"jones", "kalam", "kapitsa", "kare", "keldysh", "keller", "kepler", "khayyam",
	"khorana", "kilby", "kirch", "knuth", "kowalevski", "lalande", "lamarr",
	"lamport", "leakey", "leavitt", "lederberg", "lehmann", "lewin", "lichterman",
	"liskov", "lovelace", "lumiere", "mahavira", "margulis", "matsumoto", "maxwell",
	"mayer", "mccarthy", "mcclintock", "mclaurin", "mclean", "mcnulty", "mendel",
	"mendeleev", "meitner", "meninsky", "merkle", "mestorf", "mirzakhani", "montalcini",
	"moore", "morse", "murdock", "moser", "napier", "nash", "neumann", "newton",
	"nightingale", "nobel", "noether", "northcutt", "noyce", "panini", "pare",
	"pascal", "pasteur", "payne", "perlman", "pike", "poincare", "poitras",
	"proskuriakova", "ptolemy", "raman", "ramanujan", "ride", "ritchie", "rhodes",
	"robinson", "roentgen", "rosalind", "rubin", "saha", "sammet", "sanderson",
	"satoshi", "shamir", "shannon", "shaw", "shirley", "shockley", "shtern", "sinoussi",
	"snyder", "solomon", "spence", "stonebraker", "sutherland", "swanson", "swartz",
	"swirles", "taussig", "tereshkova", "tesla", "tharp", "thompson", "torvalds",
	"tu", "turing", "tyson", "varahamihira", "vaughan", "vaughn", "villani",
	"visvesvaraya", "volhard", "wescoff", "wilbur", "wiles", "williams", "williamson",
	"wilson", "wing", "wozniak", "wright", "wu", "yalow", "yonath", "zhukovsky",
}

// GenerateName creates a random container name in the format "adjective-noun".
func GenerateName() string {
	return Str.IntersperseSemigroup("-").Concat(
		pickRandom(adjectives),
		pickRandom(nouns),
	)
}

// pickRandom selects a random element from a slice using fp-go.
func pickRandom[T any](items []T) T {
	return F.Pipe1(
		items,
		A.Lookup[randomInt(len(items))],
		O.GetOrElse(func() T { var zero T; return zero }),
	)
}

// randomInt returns a cryptographically secure random integer in [0, max).
func randomInt(max int) int {
	return F.Pipe1(
		O.FromPredicate(func(int) bool { return max > 0 })(max),
		O.Fold(
			func() int { return 0 },
			func(m int) int {
				n, err := rand.Int(rand.Reader, big.NewInt(int64(m)))
				return F.Pipe1(
					O.FromPredicate(func(error) bool { return err == nil })(err),
					O.Fold(
						func() int {
							return int(uint64(time.Now().UnixNano()) % uint64(m))
						},
						func(error) int { return int(n.Int64()) },
					),
				)
			},
		),
	)
}
```

### 7.5 `internal/containercreate/mounts.go`

```go
package containercreate

import (
	"fmt"

	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
)

// renderMount converts a MountSpec to a container --mount argument.
// Format: type=bind,source=<host>,target=<container>[,readonly]
func renderMount(mount MountSpec) string {
	return F.Pipe1(
		mount.Readonly,
		func(readonly bool) string {
			return fmt.Sprintf(
				"type=bind,source=%s,target=%s",
				mount.HostPath,
				mount.TargetPath,
			)
		},
		func(base string) string {
			return F.Pipe1(
				mount.Readonly,
				func(ro bool) string {
					return F.Pipe1(
						ro,
						func(isReadonly bool) string {
							return base + func() string {
								if isReadonly {
									return ",readonly"
								}
								return ""
							}()
						},
					)
				},
			)
		},
	)
}

// renderMountArgs converts a MountSpec to a pair of arguments ["--mount", "<spec>"].
func renderMountArgs(mount MountSpec) []string {
	return []string{"--mount", renderMount(mount)}
}

// renderAllMounts converts all MountSpecs to flattened --mount arguments.
func renderAllMounts(mounts []MountSpec) []string {
	return F.Pipe1(
		mounts,
		A.Map(renderMountArgs),
		A.Flatten,
	)
}

// renderEnvVars returns the environment variable arguments for Crush.
func renderEnvVars() []string {
	return []string{
		"--env", fmt.Sprintf("CRUSH_GLOBAL_CONFIG=%s", ConfigTarget),
		"--env", fmt.Sprintf("CRUSH_GLOBAL_DATA=%s", DataTarget),
	}
}
```

### 7.6 `internal/containercreate/args.go`

```go
package containercreate

import (
	A "github.com/IBM/fp-go/v2/array"
	F "github.com/IBM/fp-go/v2/function"
	O "github.com/IBM/fp-go/v2/option"
)

const containerBinary = "container"

// RenderCommand builds the CommandSpec for "container create".
func RenderCommand(r ResolvedInput) CommandSpec {
	return CommandSpec{
		Bin:  containerBinary,
		Args: buildArgs(r),
	}
}

// buildArgs constructs the full argument list using fp-go array operations.
func buildArgs(r ResolvedInput) []string {
	return F.Pipe1(
		[]string{"create"},
		A.Concat(buildNameArgs(r.ContainerName)),
		A.Concat(renderAllMounts(r.Mounts)),
		A.Concat(renderEnvVars()),
		A.Concat(buildWorkdirArgs(r.Workdir)),
		A.Concat([]string{r.ImageName}),
	)
}

// buildNameArgs constructs the --name argument.
func buildNameArgs(name string) []string {
	return []string{"--name", name}
}

// buildWorkdirArgs constructs the --workdir argument if specified.
func buildWorkdirArgs(workdir string) []string {
	return F.Pipe1(
		O.FromPredicate(func(s string) bool { return s != "" })(workdir),
		O.Fold(
			func() []string { return []string{} },
			func(w string) []string { return []string{"--workdir", w} },
		),
	)
}
```

### 7.7 `internal/containercreate/exec.go`

```go
package containercreate

import (
	"fmt"
	"os"
	"os/exec"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// Execute runs the container create command and returns the result.
func Execute(r ResolvedInput) IOE.IOEither[error, ContainerResult] {
	return F.Pipe2(
		runProcess(RenderCommand(r)),
		IOE.Map[error](func(F.Void) ContainerResult {
			return ContainerResult{Name: r.ContainerName}
		}),
	)
}

// runProcess executes the container binary with the given CommandSpec.
func runProcess(spec CommandSpec) IOE.IOEither[error, F.Void] {
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
			return fmt.Errorf("container create failed: %w", err)
		}),
	)
}
```

### 7.8 `internal/containercreate/validate_test.go`

```go
package containercreate

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestValidateInput_MissingConfig(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath: "",
		DataPath:   "/tmp/data",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for missing config path")
}

func TestValidateInput_MissingData(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for missing data path")
}

func TestValidateInput_ValidMinimal(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result), "expected Right for valid minimal input")

	resolved := E.GetOrElse(func(error) ResolvedInput {
		return ResolvedInput{}
	})(result)

	require.Equal("/tmp/config", resolved.Mounts[0].HostPath)
	require.Equal(ConfigTarget, resolved.Mounts[0].TargetPath)
	require.True(resolved.Mounts[0].Readonly)

	require.Equal("/tmp/data", resolved.Mounts[1].HostPath)
	require.Equal(DataTarget, resolved.Mounts[1].TargetPath)
	require.False(resolved.Mounts[1].Readonly)
}

func TestValidateInput_WithWorkspace(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		WorkspacePath: "/tmp/workspace",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result))

	resolved := E.GetOrElse(func(error) ResolvedInput {
		return ResolvedInput{}
	})(result)

	require.Len(resolved.Mounts, 3)
	require.Equal(WorkspaceTarget, resolved.Mounts[2].TargetPath)
	require.False(resolved.Mounts[2].Readonly)
	require.Equal(WorkspaceTarget, resolved.Workdir)
}

func TestValidateInput_InvalidContainerName(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		ContainerName: "123invalid",
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for invalid container name")
}

func TestValidateInput_ValidContainerName(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath:    "/tmp/config",
		DataPath:      "/tmp/data",
		ContainerName: "my-crush-sandbox",
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result))

	resolved := E.GetOrElse(func(error) ResolvedInput {
		return ResolvedInput{}
	})(result)

	require.Equal("my-crush-sandbox", resolved.ContainerName)
}

func TestValidateInput_ReservedVolumeName(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		Volumes:    []string{"/tmp/config"},
	}

	result := ValidateInput(input)

	require.True(E.IsLeft(result), "expected Left for reserved volume name")
}

func TestValidateInput_AdditionalVolumesReadOnly(t *testing.T) {
	require := require.New(t)

	input := Input{
		ConfigPath: "/tmp/config",
		DataPath:   "/tmp/data",
		Volumes:    []string{"/tmp/projects", "/opt/tools"},
	}

	result := ValidateInput(input)

	require.True(E.IsRight(result))

	resolved := E.GetOrElse(func(error) ResolvedInput {
		return ResolvedInput{}
	})(result)

	require.Len(resolved.Mounts, 4)

	require.Equal(ContainerHome+"/opt", resolved.Mounts[2].TargetPath)
	require.True(resolved.Mounts[2].Readonly, "additional volumes must be read-only")

	require.Equal(ContainerHome+"/projects", resolved.Mounts[3].TargetPath)
	require.True(resolved.Mounts[3].Readonly, "additional volumes must be read-only")
}

func TestIsReservedBasename(t *testing.T) {
	require := require.New(t)

	require.True(isReservedBasename("config"))
	require.True(isReservedBasename("data"))
	require.True(isReservedBasename("crush"))
	require.False(isReservedBasename("projects"))
	require.False(isReservedBasename("workspace"))
}

func TestIsValidContainerName(t *testing.T) {
	require := require.New(t)

	require.True(isValidContainerName("my-container"))
	require.True(isValidContainerName("container123"))
	require.True(isValidContainerName("a"))
	require.False(isValidContainerName("123container"))
	require.False(isValidContainerName("-container"))
	require.False(isValidContainerName(""))
}

func TestHasValidBasename(t *testing.T) {
	require := require.New(t)

	require.True(hasValidBasename("projects"))
	require.True(hasValidBasename("my-folder123"))
	require.False(hasValidBasename("---"))
	require.False(hasValidBasename(""))
}
```

### 7.9 `internal/containercreate/namegen_test.go`

```go
package containercreate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateName_Format(t *testing.T) {
	require := require.New(t)

	name := GenerateName()

	parts := strings.Split(name, "-")
	require.GreaterOrEqual(len(parts), 2, "name should have at least two parts")
	require.Regexp(`^[a-z]+-[a-z]+$`, name, "name should be lowercase with dash separator")
}

func TestGenerateName_Uniqueness(t *testing.T) {
	require := require.New(t)

	names := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := GenerateName()
		names[name] = true
	}

	require.Greater(len(names), 50, "should generate diverse names")
}

func TestRandomInt_Range(t *testing.T) {
	require := require.New(t)

	for i := 0; i < 100; i++ {
		n := randomInt(10)
		require.GreaterOrEqual(n, 0)
		require.Less(n, 10)
	}
}

func TestRandomInt_Zero(t *testing.T) {
	require := require.New(t)

	require.Equal(0, randomInt(0))
	require.Equal(0, randomInt(-1))
}
```

### 7.10 `internal/containercreate/args_test.go`

```go
package containercreate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderCommand_Minimal(t *testing.T) {
	require := require.New(t)

	resolved := ResolvedInput{
		ImageName:     DefaultImageName,
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
		Workdir: "",
	}

	spec := RenderCommand(resolved)

	require.Equal("container", spec.Bin)
	require.Equal("create", spec.Args[0])
	require.Contains(spec.Args, "--name")
	require.Contains(spec.Args, "test-container")
	require.Contains(spec.Args, DefaultImageName)
	require.Equal(DefaultImageName, spec.Args[len(spec.Args)-1])
}

func TestRenderCommand_WithWorkspace(t *testing.T) {
	require := require.New(t)

	resolved := ResolvedInput{
		ImageName:     DefaultImageName,
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
			{HostPath: "/host/workspace", TargetPath: WorkspaceTarget, Readonly: false},
		},
		Workdir: WorkspaceTarget,
	}

	spec := RenderCommand(resolved)

	require.Contains(spec.Args, "--workdir")
	require.Contains(spec.Args, WorkspaceTarget)
}

func TestRenderCommand_MountFormat(t *testing.T) {
	require := require.New(t)

	resolved := ResolvedInput{
		ImageName:     DefaultImageName,
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
		},
	}

	spec := RenderCommand(resolved)

	var mountArg string
	for i, arg := range spec.Args {
		if arg == "--mount" && i+1 < len(spec.Args) {
			mountArg = spec.Args[i+1]
			break
		}
	}

	require.Contains(mountArg, "type=bind")
	require.Contains(mountArg, "source=/host/config")
	require.Contains(mountArg, "target="+ConfigTarget)
	require.Contains(mountArg, "readonly")
}

func TestRenderCommand_EnvVars(t *testing.T) {
	require := require.New(t)

	resolved := ResolvedInput{
		ImageName:     DefaultImageName,
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
		},
	}

	spec := RenderCommand(resolved)

	args := strings.Join(spec.Args, " ")
	require.Contains(args, "CRUSH_GLOBAL_CONFIG="+ConfigTarget)
	require.Contains(args, "CRUSH_GLOBAL_DATA="+DataTarget)
}

func TestRenderCommand_AdditionalVolumesReadOnly(t *testing.T) {
	require := require.New(t)

	resolved := ResolvedInput{
		ImageName:     DefaultImageName,
		ContainerName: "test-container",
		Mounts: []MountSpec{
			{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
			{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
			{HostPath: "/host/projects", TargetPath: ContainerHome + "/projects", Readonly: true},
		},
	}

	spec := RenderCommand(resolved)

	args := strings.Join(spec.Args, " ")
	require.Contains(args, "source=/host/projects")
	require.Contains(args, ContainerHome+"/projects")
	require.Contains(args, "readonly")
}
```

### 7.11 `internal/containercreate/mounts_test.go`

```go
package containercreate

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderMount_ReadOnly(t *testing.T) {
	require := require.New(t)

	mount := MountSpec{
		HostPath:   "/host/config",
		TargetPath: "/container/config",
		Readonly:   true,
	}

	result := renderMount(mount)

	require.Contains(result, "type=bind")
	require.Contains(result, "source=/host/config")
	require.Contains(result, "target=/container/config")
	require.Contains(result, "readonly")
	require.True(strings.HasSuffix(result, ",readonly"))
}

func TestRenderMount_ReadWrite(t *testing.T) {
	require := require.New(t)

	mount := MountSpec{
		HostPath:   "/host/data",
		TargetPath: "/container/data",
		Readonly:   false,
	}

	result := renderMount(mount)

	require.Contains(result, "type=bind")
	require.Contains(result, "source=/host/data")
	require.Contains(result, "target=/container/data")
	require.NotContains(result, "readonly")
}

func TestRenderAllMounts(t *testing.T) {
	require := require.New(t)

	mounts := []MountSpec{
		{HostPath: "/host/config", TargetPath: ConfigTarget, Readonly: true},
		{HostPath: "/host/data", TargetPath: DataTarget, Readonly: false},
	}

	result := renderAllMounts(mounts)

	require.Len(result, 4)
	require.Equal("--mount", result[0])
	require.Contains(result[1], "config")
	require.Equal("--mount", result[2])
	require.Contains(result[3], "data")
}

func TestRenderEnvVars(t *testing.T) {
	require := require.New(t)

	result := renderEnvVars()

	require.Len(result, 4)
	require.Equal("--env", result[0])
	require.Contains(result[1], "CRUSH_GLOBAL_CONFIG")
	require.Equal("--env", result[2])
	require.Contains(result[3], "CRUSH_GLOBAL_DATA")
}
```

---

## 8. Integration with Main CLI

### 8.1 Update `cmd/container-cli/main.go`

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cybersiddhu/crush-sandbox/internal/containerbuild"
	"github.com/cybersiddhu/crush-sandbox/internal/containercreate"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "container-cli",
		Usage: "Build OCI images and create containers through the container CLI",
		Commands: []*cli.Command{
			containerbuild.Command(),
			containercreate.Command(),
		},
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

---

## 9. fp-go API Reference Used

### 9.1 Array Operations (`array`)

| Function | Usage |
|----------|-------|
| `A.Map` | Transform array elements |
| `A.Filter` | Filter array elements |
| `A.Chain` | FlatMap over arrays |
| `A.Flatten` | Flatten nested arrays |
| `A.Sort` | Sort with Ord |
| `A.Concat` | Concatenate arrays |
| `A.ArrayConcatAll` | Concatenate multiple arrays |
| `A.Lookup` | Safe indexed access returning Option |

### 9.2 Either Operations (`either`)

| Function | Usage |
|----------|-------|
| `E.Of` | Lift value into Either |
| `E.FromPredicate` | Lift based on predicate |
| `E.Chain` | Sequence Either operations |
| `E.Map` | Transform Right value |
| `E.MapTo` | Replace Right value |
| `E.Filter` | Filter with predicate |
| `E.Fold` | Terminal branching |
| `E.SequenceArray` | Convert `[]Either` to `Either[]` |
| `E.IsLeft` / `E.IsRight` | Check type |

### 9.3 Option Operations (`option`)

| Function | Usage |
|----------|-------|
| `O.Of` | Lift value into Option |
| `O.FromPredicate` | Lift based on predicate |
| `O.Fold` | Terminal branching |
| `O.GetOrElse` | Extract with default |
| `O.IsSome` / `O.IsNone` | Check type |
| `O.Map` | Transform Some value |

### 9.4 Predicate Operations (`predicate`)

| Function | Usage |
|----------|-------|
| `P.Not` | Negate predicate |
| `P.And` | Logical AND |
| `P.Or` | Logical OR |
| `P.ContraMap` | Transform predicate input |

### 9.5 String Operations (`string`)

| Function | Usage |
|----------|-------|
| `Str.IsEmpty` | Check if empty |
| `Str.IntersperseSemigroup` | Semigroup for joining |

### 9.6 Ordering Operations (`ord`)

| Function | Usage |
|----------|-------|
| `Ord.FromCompare` | Create Ord from compare function |
| `Ord.Contramap` | Transform Ord input |

### 9.7 Equality Operations (`eq`)

| Function | Usage |
|----------|-------|
| `Eq.FromStrictEquals` | Create Eq from == |

### 9.8 Record Operations (`record`)

| Function | Usage |
|----------|-------|
| `R.FromEntries` | Build map from entries |
| `R.Lookup` | Safe lookup returning Option |

---

## 10. Test Strategy

### 10.1 Unit Tests

| File | Test Coverage |
|------|---------------|
| `validate_test.go` | Path validation, name validation, volume validation, reserved names |
| `namegen_test.go` | Name format, uniqueness, randomness bounds |
| `args_test.go` | Command rendering, mount format, env vars, workspace |
| `mounts_test.go` | Mount rendering, readonly vs read-write, env vars |

### 10.2 Integration Tests

```bash
# Build the binary
go build ./cmd/container-cli/...

# Test help
./container-cli create --help

# Test validation failure
./container-cli create --config /nonexistent

# Test successful creation (requires container binary)
./container-cli create \
  --config ~/.config/crush \
  --data ~/.local/share/crush \
  --workspace .
```

### 10.3 Edge Cases

1. **Relative paths** — Resolved to absolute via `filepath.Abs`
2. **Blank paths** — Rejected by predicate validation
3. **Reserved basenames** — Rejected by Record lookup
4. **Invalid container names** — Rejected by regex predicate
5. **Empty volume list** — Handled by array operations (returns empty)

---

## 11. Error Messages

| Error | Cause | Resolution |
|-------|-------|------------|
| `config path is required` | `--config` not provided | Provide `--config` flag |
| `data path is required` | `--data` not provided | Provide `--data` flag |
| `container name must start with a letter...` | Invalid `--name` format | Use alphanumeric, dash, underscore, starting with letter |
| `volume basename "<name>" is reserved` | Volume uses "config", "data", or "crush" | Use different path |
| `volume basename "<name>" must contain at least one letter or digit` | Basename invalid | Use valid path |
| `container create failed: <err>` | Container binary failed | Check container tool logs |

---

## 12. Dependencies

No new dependencies beyond existing:

- `github.com/urfave/cli/v3` — CLI framework
- `github.com/IBM/fp-go/v2` — Functional programming
- `github.com/stretchr/testify` — Test assertions

---

## 13. Summary

This plan provides a complete specification for implementing the `create` subcommand following strict fp-go conventions:

### Key Changes from Original

1. **Additional volumes are read-only** — Changed from read-write to read-only for safety
2. **No imperative branching** — All code uses fp-go combinators (`E.Fold`, `O.Fold`, `P.And`, etc.)
3. **Full fp-go API usage**:
   - `A.Map`, `A.Filter`, `A.Sort`, `A.Chain`, `A.Flatten` for arrays
   - `P.Not`, `P.And`, `P.ContraMap` for predicates
   - `Ord.Contramap`, `Ord.FromCompare` for sorting
   - `Eq.FromStrictEquals` for equality
   - `R.FromEntries`, `R.Lookup` for reserved names map
   - `O.Fold`, `O.GetOrElse` for optional values

### Mount Permissions Summary

| Mount Type | Permission |
|------------|------------|
| Config | read-only |
| Data | read-write |
| Workspace | read-write |
| Additional volumes | **read-only** |

The implementation is fully specified with all types, functions, and tests documented.

---

## 14. Review Recommendations

### 14.1 Correctness Updates

1. Decide whether `--workspace` is truly optional. If yes, remove the default `.` and update examples accordingly. If no, reclassify it as a defaulted mount rather than an optional feature.
2. Replace `container start -it <name>` with `container start -ai <name>` or equivalent wording that matches `docs/command-reference.md`.
3. Align validation strategy: either rely on `urfave/cli` required-flag errors or remove `Required: true` and keep custom validation errors, but do not document both as primary behavior.
4. Update the “Integration with Main CLI” section to note that `cmd/container-cli/main.go` already registers `containercreate.Command()`.

### 14.2 Implementation Fixes Needed

1. Mark the code samples as illustrative or revise them until they compile cleanly.
2. Fix missing imports and alias mismatches in the validation and name generation snippets.
3. Replace invalid expressions such as `E.Map(EqString.Equals)` and `E.Of[error](filepath.Abs(path))` with compile-valid fp-go code.
4. Fix the `F.PipeN` usages whose arity does not match the number of transforms.
5. Remove the literal `if` from `renderMount` if the document is enforcing a strict “no imperative branching” rule.

### 14.3 Validation and Mounting Gaps

1. Add existence and directory checks for config, data, workspace, and extra volume paths.
2. Reserve or otherwise prevent mount-target collisions for `workspace` and duplicate basenames across `--volume` entries.
3. Ensure the implementation preserves the documented mount order: core mounts first, additional mounts sorted only within their own group.
4. Clarify whether the current plan wants fail-fast validation for host-path problems or is intentionally delegating them to `container create`.

### 14.4 Verification Notes

- `go build ./cmd/container-cli/...` currently passes in the repo.
- `gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./internal/containercreate/...` currently passes in the repo.
- `cmd/container-cli/main.go` already includes `containercreate.Command()`.

# PLAN: Embed Dockerfile in Binary

## 1. Overview

Embed the project `Dockerfile` into the `container-cli` binary so users can run
`container-cli build --embed` without needing a Dockerfile on disk. The design
replaces the string `File` field in `Input` with a **lazy IOEither resolver**
(`DockerfileSource`) that encapsulates both the file-based and embedded
strategies. A `WithResource` bracket combinator guarantees temp-file cleanup on
every code path, including failures.

---

## 2. Problems with the Naive Approach

A first-pass design would add `UseEmbed bool` to `Input` and sprinkle `if`
checks across validate/args/exec. This violates the project's fp-go
conventions:

| Problem | Root Cause |
|---|---|
| `if r.UseEmbed` in validate/args/exec | Boolean field scatters dispatch across 3 files |
| Sentinel string `"EMBEDDED_DOCKERFILE"` | Requires a for-loop to replace — not composable |
| No cleanup on error path | Temp file leaks when `Execute` returns Left |
| File validation mixed into `ValidateInput` | Mixed concern: file lifetime differs from tag lifetime |

---

## 3. Architectural Shifts

| Naive | Improved |
|---|---|
| `UseEmbed bool` + `File string` in `Input` | `DockerfileSource IOE.IOEither[error, DockerfileResource]` |
| Boolean dispatch `if/else` in 3 files | Map dispatch **once** in `InputFromCommand` — nowhere else |
| Sentinel string + for-loop in exec | Path resolved via IOEither chain before rendering |
| No cleanup guarantee | `ioeither.WithResource` bracket (fp-go built-in) — always releases |
| Imperative `EmbeddedResolver` with if/else | Multi-step `F.Pipe` using `IOEF.CreateTemp`, `IOEF.Close`, `IOEF.Remove` |
| File validation in `ValidateInput` | Validation co-located in `FileResolver` |

---

## 4. Data Flow

### Current Pipeline

```
Input{File, Name, Tags, Ctx}
  │
  ▼  ValidateInput          ── pure: validates File + Tags
Either[error, Input]
  │
  ▼  IOE.FromEither
IOEither[error, Input]
  │
  ▼  IOE.Map(RenderCommand) ── pure: builds CommandSpec inside Input
IOEither[error, Input]
  │
  ▼  IOE.Chain(Execute)      ── effect: runs container process
IOEither[error, struct{}]
  │
  ▼  FP.ToEither → E.Fold
error
```

### New Pipeline

```
Input{DockerfileSource, Name, Tags, Ctx}
  │
  ▼  ValidateInput            ── pure: validates Tags only
Either[error, Input]
  │
  ▼  IOE.FromEither
IOEither[error, Input]
  │
  ▼  IOE.Chain(Execute)
  │    ┌───────────────────────────────────────────┐
  │    │  ioeither.WithResource bracket             │
  │    │                                            │
  │    │  acquire: DockerfileSource                  │
  │    │     → DockerfileResource{Path, Release}    │
  │    │                                            │
  │    │  use: RenderCommand(r, res.Path)           │
  │    │     → CommandSpec → runProcess             │
  │    │                                            │
  │    │  release: res.Release                      │
  │    │     → IOEF.Remove(path) or nopRelease      │
  │    │     → ALWAYS called, even on error         │
  │    └───────────────────────────────────────────┘
IOEither[error, struct{}]
  │
  ▼  FP.ToEither → E.Fold
error
```

Key difference: `RenderCommand` is no longer a separate pipeline step. It is
called _inside_ `Execute` after the resource is acquired, receiving the resolved
path as a direct argument.

---

## 5. New Files

### 5.1 No Custom `WithResource` — Use `ioeither.WithResource` Directly

fp-go already ships `ioeither.WithResource`:

```go
func WithResource[A, E, R, ANY any](
    onCreate IOEither[E, R],
    onRelease Kleisli[E, R, ANY],
) Kleisli[E, Kleisli[E, R, A], A]
```

This is a curried bracket: it takes `acquire` and `release`, then returns a
function expecting a `use` Kleisli. Release is **always** called, even when
`use` returns Left. There is no need to write a custom bracket in
`internal/fp/`.

The `ioeither/file` package builds on this with helpers we will use:

| Combinator | Package | Signature / Description |
|---|---|---|
| `CreateTemp` | `ioeither/file` | `func(dir, pattern string) IOEither[error, *os.File]` — wraps `os.CreateTemp` via `Eitherize2` |
| `WriteAll` | `ioeither/file` | `func(data []byte) Operator[error, W, []byte]` — writes `data` to a `WriteCloser`, then closes it via internal `WithResource` |
| `Remove` | `ioeither/file` | `func(name string) IOEither[error, string]` — wraps `os.Remove`, returns the path on success |
| `Close` | `ioeither/file` | `func(c C) IOEither[error, struct{}]` — wraps `io.Closer.Close` |

**Why not `WithTempFile`?** `WithTempFile` closes **and removes** the temp file
after the callback. We need the file to persist on disk so the external
`container build --file <path>` process can read it. Our lifecycle is:

1. **Create + write + close** → `WriteAll(data)(CreateTemp("", "Dockerfile-*"))`
2. External process reads the file
3. **Remove** → deferred to `WithResource`'s release callback

This two-phase lifecycle requires composing the fp-go primitives ourselves
rather than using the pre-composed `WithTempFile`.

---

### 5.2 `internal/containerbuild/embed.go` — Embed Directive

```go
package build

import _ "embed"

//go:embed Dockerfile
var embeddedDockerfile string
```

The `//go:embed` directive requires the target file to be in the same directory
as the Go source file (or a subdirectory). Copy the project-root `Dockerfile`
to `internal/containerbuild/Dockerfile`.

---

### 5.3 `internal/containerbuild/resource.go` — DockerfileResource & Resolvers

#### Key fp-go file API patterns used

From the examples in `fp-go-concepts/v2/file/`:

| Pattern | Example | What it does |
|---|---|---|
| `IOEF.Write[R, *os.File](acquire)(kleisli)` | `04_interface_helpers.go` | Bracket: acquire → use kleisli → auto-close, even on error |
| `IOEF.WriteAll[*os.File](data)(acquire)` | `05_writeall_pattern.go` | Bracket: acquire → write data → close, returns `[]byte` |
| `IOEF.CreateTemp(dir, pattern)` | `tempfile.go` | `Eitherize2(os.CreateTemp)` → `IOEither[error, *os.File]` |

`IOEF.WriteAll` returns the **data** (`[]byte`), so we lose `f.Name()`.
`IOEF.Write[string, *os.File]` is the right combinator: it takes an `acquire`
(`IOEither[error, *os.File]`) and returns a function expecting a Kleisli
`(*os.File → IOEither[error, string])`. Inside the Kleisli we write content
**and** return `f.Name()`. The file is guaranteed to be closed by the internal
bracket — even if the write fails.

#### Full file

```go
package build

import (
	"errors"
	"os"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	IOEF "github.com/IBM/fp-go/v2/ioeither/file"
)

// DockerfileResource pairs a resolved Dockerfile path with its cleanup
// IOEither. For file-based builds cleanup is a nop; for embedded builds
// it removes the temp file via IOEF.Remove.
type DockerfileResource struct {
	Path    string
	Release IOE.IOEither[error, string]
}

var nopRelease = IOE.Of[error]("")

// FileResolver validates that path is non-blank, then wraps it in a
// DockerfileResource with nop cleanup.
func FileResolver(path string) IOE.IOEither[error, DockerfileResource] {
	return F.Pipe1(
		F.Pipe2(
			path,
			E.FromPredicate(isNonBlank, func(string) error {
				return errors.New("dockerfile path is required")
			}),
			E.Map[error](func(p string) DockerfileResource {
				return DockerfileResource{Path: p, Release: nopRelease}
			}),
		),
		IOE.FromEither[error, DockerfileResource],
	)
}

// writeContentAndReturnName writes the embedded content to the file and
// returns the file's Name(). This is the Kleisli arrow passed to
// IOEF.Write — the bracket guarantees the file is closed even if the
// write fails.
func writeContentAndReturnName(
	content []byte,
) func(*os.File) IOE.IOEither[error, string] {
	return func(f *os.File) IOE.IOEither[error, string] {
		return F.Pipe2(
			IOE.TryCatchError(func() (int, error) {
				return f.Write(content)
			}),
			IOE.Map[error](func(int) string {
				return f.Name()
			}),
		)
	}
}

// EmbeddedResolver writes the compile-time embedded Dockerfile content to a
// temp file and returns a DockerfileResource whose Release removes that file.
//
// Pipeline (all IOEither combinators, no if/else):
//
//  1. IOEF.CreateTemp("", "Dockerfile-*")        → IOEither[error, *os.File]
//     (acquire for the Write bracket)
//
//  2. IOEF.Write[string, *os.File](acquire)      → bracket: acquire → use → close
//     (writeContentAndReturnName)                   Kleisli writes content, returns Name()
//                                                   File is ALWAYS closed (bracket guarantee)
//
//  3. IOE.Map → DockerfileResource                → wires IOEF.Remove(name) as Release
//
// This follows the same pattern as 04_interface_helpers.go:
//
//   F.Pipe1(
//       IOEF.Create(path),
//       IOEF.Write[int, *os.File],
//   )(writeKleisli)
//
// but using CreateTemp instead of Create, and returning the path instead
// of the byte count.
func EmbeddedResolver() IOE.IOEither[error, DockerfileResource] {
	content := []byte(embeddedDockerfile)

	return F.Pipe1(
		F.Pipe1(
			IOEF.CreateTemp("", "Dockerfile-*"),
			IOEF.Write[string, *os.File],
		)(writeContentAndReturnName(content)),
		IOE.Map[error](func(name string) DockerfileResource {
			return DockerfileResource{
				Path:    name,
				Release: IOEF.Remove(name),
			}
		}),
	)
}
```

#### Type trace

```
IOEF.CreateTemp("", "Dockerfile-*")
  → IOEither[error, *os.File]                          (acquire)

IOEF.Write[string, *os.File]
  → func(IOEither[error, *os.File])                    (partially applied with acquire)
       func(Kleisli[error, *os.File, string])           (waiting for use)
          IOEither[error, string]                       (result after bracket)

writeContentAndReturnName(content)
  → func(*os.File) IOEither[error, string]             (the Kleisli: write + return Name())

IOEF.Write[string, *os.File](acquire)(writeKleisli)
  → IOEither[error, string]                            (file path, file is closed)

IOE.Map[error](func(name string) DockerfileResource{…})
  → IOEither[error, DockerfileResource]
```

#### Design notes

- **`IOEF.Write` bracket guarantees close** — even if
  `writeContentAndReturnName` fails, the file handle is closed. This is
  strictly better than the manual `Chain → Close` approach which leaks on
  write failure.
- `isNonBlank` is defined in `validate.go` (same package) — no import needed.
- `Release` is a public `IOEither` field (not a method wrapping a private
  `cleanup func() error`). This composes directly with
  `ioeither.WithResource`'s release Kleisli.
- `IOEF.CreateTemp` wraps `os.CreateTemp` via `Eitherize2` — no manual error
  handling.
- `IOEF.Remove` wraps `os.Remove` into `IOEither[error, string]`.
- `writeContentAndReturnName` is the only `TryCatchError` call — a single
  effect boundary wrapping `(*os.File).Write`. All composition uses
  `F.Pipe1`/`F.Pipe2`, `IOE.Map`, and the `IOEF.Write` bracket.

---

## 6. Modified Files

### 6.1 `internal/containerbuild/input.go`

**Before:**

```go
package build

import "context"

const containerBinary = "container"

type Input struct {
	File string
	Name string
	Tags []string
	Ctx  context.Context
	CommandSpec
}

type CommandSpec struct {
	Bin  string
	Args []string
}
```

**After:**

```go
package build

import (
	"context"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

const containerBinary = "container"

// Input holds the build parameters throughout the pipeline.
// DockerfileSource is a lazy IOEither that resolves to a DockerfileResource
// when executed. It is set once in InputFromCommand and never branched on.
type Input struct {
	DockerfileSource IOE.IOEither[error, DockerfileResource]
	Name             string
	Tags             []string
	Ctx              context.Context
}

// CommandSpec holds the resolved executable binary and argv slice.
// Built inside Execute after the DockerfileResource is acquired.
type CommandSpec struct {
	Bin  string
	Args []string
}
```

**What changed:**

- Removed `File string` — the path is inside `DockerfileResource.Path`.
- Removed embedded `CommandSpec` — it is now built locally in `Execute` after
  the resolver runs, so it is never threaded through the pipeline.
- Added `DockerfileSource` — the lazy resolver selected once in
  `InputFromCommand`.
- New import: `ioeither` (for the `IOEither` type in the struct field).

---

### 6.2 `internal/containerbuild/validate.go`

**Before:**

```go
func ValidateInput(r Input) E.Either[error, Input] {
	validations := []E.Either[error, bool]{
		F.Pipe2(
			r.File,
			E.FromPredicate(
				isNonBlank,
				func(string) error {
					return errors.New("dockerfile path is required")
				},
			),
			E.MapTo[error, string](true),
		),
		F.Pipe2(
			r.Tags,
			E.FromPredicate(
				isAllNonBlank,
				func([]string) error {
					return errors.New("tag values must be non-empty")
				},
			),
			E.MapTo[error, []string](true),
		),
	}

	return F.Pipe1(
		E.SequenceArray(validations),
		E.MapTo[error, []bool](r),
	)
}
```

**After:**

```go
func ValidateInput(r Input) E.Either[error, Input] {
	return F.Pipe2(
		r.Tags,
		E.FromPredicate(
			isAllNonBlank,
			func([]string) error {
				return errors.New("tag values must be non-empty")
			},
		),
		E.MapTo[error, []string](r),
	)
}
```

**What changed:**

- Removed the file-path validation — `FileResolver` now owns that check.
- Removed `E.SequenceArray` — with only one validation remaining, a direct
  `F.Pipe2` is cleaner.
- The predicate helpers (`isBlank`, `isNonBlank`, `isAllNonBlank`) remain
  unchanged; `isNonBlank` is reused by `FileResolver` in `resource.go`.

---

### 6.3 `internal/containerbuild/args.go`

**Before:**

```go
func RenderCommand(r Input) Input {
	return Input{
		File: r.File,
		Name: r.Name,
		Tags: r.Tags,
		Ctx:  r.Ctx,
		CommandSpec: CommandSpec{
			Bin: containerBinary,
			Args: A.ArrayConcatAll(
				[]string{"build", "--file", r.File},
				renderTagArgs(r),
				[]string{"."},
			),
		},
	}
}
```

**After:**

```go
// RenderCommand is a pure function that builds a CommandSpec from an Input
// and a resolved Dockerfile path. Called inside Execute after the
// DockerfileResource is acquired.
func RenderCommand(r Input, path string) CommandSpec {
	return CommandSpec{
		Bin: containerBinary,
		Args: A.ArrayConcatAll(
			[]string{"build", "--file", path},
			renderTagArgs(r),
			[]string{"."},
		),
	}
}
```

**What changed:**

- Signature: `Input → Input` becomes `(Input, string) → CommandSpec`.
- The path is an explicit argument instead of reading `r.File`.
- Returns `CommandSpec` directly — no need to rebuild an entire `Input`.
- `renderTagArgs` is unchanged (reads `r.Name` and `r.Tags`).

---

### 6.4 `internal/containerbuild/command.go`

**Before:**

```go
package build

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an OCI image via the container CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path to Dockerfile",
				Value:   "Dockerfile",
			},
			&cli.StringSliceFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "Image tag, repeatable",
				Value:   []string{"latest"},
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Image name (combines with tags as name:tag)",
				Value:   "container",
			},
		},
		Action: Action,
	}
}

func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe6(
		Input{
			File: cmd.String("file"),
			Name: cmd.String("name"),
			Tags: cmd.StringSlice("tag"),
			Ctx:  ctx,
		},
		ValidateInput,
		IOE.FromEither[error],
		IOE.Map[error](RenderCommand),
		IOE.Chain(Execute),
		FP.ToEither[error, struct{}],
		E.Fold(
			F.Identity[error],
			func(struct{}) error { return nil },
		),
	)
}
```

**After:**

```go
package build

import (
	"context"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	FP "github.com/cybersiddhu/crush-sandbox/internal/fp"
	"github.com/urfave/cli/v3"
)

// resolverFactories is a map-based dispatch table that selects the Dockerfile
// resolver strategy based on the --embed flag. This is the ONLY location
// where the boolean is observed — no if/else anywhere in application code.
var resolverFactories = map[bool]func(*cli.Command) IOE.IOEither[error, DockerfileResource]{
	false: func(cmd *cli.Command) IOE.IOEither[error, DockerfileResource] {
		return FileResolver(cmd.String("file"))
	},
	true: func(_ *cli.Command) IOE.IOEither[error, DockerfileResource] {
		return EmbeddedResolver()
	},
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build an OCI image via the container CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Path to Dockerfile",
				Value:   "Dockerfile",
			},
			&cli.StringSliceFlag{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "Image tag, repeatable",
				Value:   []string{"latest"},
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Image name (combines with tags as name:tag)",
				Value:   "container",
			},
			&cli.BoolFlag{
				Name:  "embed",
				Usage: "Use the Dockerfile embedded in the binary (ignores --file)",
			},
		},
		Action: Action,
	}
}

// InputFromCommand reads CLI flags and selects the Dockerfile resolver
// via the map-based dispatch table.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
	return Input{
		DockerfileSource: resolverFactories[cmd.Bool("embed")](cmd),
		Name:             cmd.String("name"),
		Tags:             cmd.StringSlice("tag"),
		Ctx:              ctx,
	}
}

// Action is the build subcommand entry point.
// Pipeline: validate tags → acquire dockerfile → render args → run process → release.
func Action(ctx context.Context, cmd *cli.Command) error {
	return F.Pipe5(
		InputFromCommand(ctx, cmd),
		ValidateInput,
		IOE.FromEither[error],
		IOE.Chain(Execute),
		FP.ToEither[error, struct{}],
		E.Fold(
			F.Identity[error],
			func(struct{}) error { return nil },
		),
	)
}
```

**What changed:**

- Added `--embed` `BoolFlag` with usage hint that it ignores `--file`.
- Extracted `InputFromCommand` — constructs `Input` with the selected resolver.
- `resolverFactories` map replaces boolean dispatch without `if/else`.
- Pipeline simplifies from `F.Pipe6` to `F.Pipe5` — the `IOE.Map(RenderCommand)`
  step is gone because `RenderCommand` is now called inside `Execute` after the
  resource is acquired.
- `FP.ToEither` is still used (from `internal/fp`) — this package is retained
  for the `ToEither` utility. Only the custom `WithResource` bracket is
  removed.

---

### 6.5 `internal/containerbuild/exec.go`

**Before:**

```go
package build

import (
	"fmt"
	"os"
	"os/exec"

	IOE "github.com/IBM/fp-go/v2/ioeither"
)

func Execute(r Input) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		cmd := exec.CommandContext(r.Ctx, containerBinary)
		cmd.Args = append(cmd.Args, r.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
```

**After:**

```go
package build

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	F "github.com/IBM/fp-go/v2/function"
	IOE "github.com/IBM/fp-go/v2/ioeither"
)

// releaseResource is the Kleisli arrow passed to ioeither.WithResource as the
// release callback. It delegates to the resource's Release field (nop for
// file-based, IOEF.Remove for embedded).
func releaseResource(res DockerfileResource) IOE.IOEither[error, string] {
	return res.Release
}

// useResource builds the use Kleisli arrow for ioeither.WithResource.
// It renders the argv from the resolved path and runs the container process.
func useResource(r Input) func(DockerfileResource) IOE.IOEither[error, struct{}] {
	return func(res DockerfileResource) IOE.IOEither[error, struct{}] {
		return runProcess(r.Ctx, RenderCommand(r, res.Path))
	}
}

// Execute orchestrates the full build lifecycle using ioeither.WithResource:
//
//  1. acquire — run DockerfileSource to get a DockerfileResource
//  2. use    — render the argv and run the container process
//  3. release — run res.Release (nop for file, IOEF.Remove for embedded)
//
// release is guaranteed to run even when the container process fails.
//
// ioeither.WithResource returns a Kleisli[E, Kleisli[E, R, A], A], so we
// call it with the use Kleisli to get the final IOEither.
func Execute(r Input) IOE.IOEither[error, struct{}] {
	return F.Pipe1(
		useResource(r),
		IOE.WithResource[struct{}](r.DockerfileSource, releaseResource),
	)
}

// runProcess executes the container binary with the given CommandSpec.
// Split from Execute to satisfy funlen limits.
func runProcess(ctx context.Context, spec CommandSpec) IOE.IOEither[error, struct{}] {
	return IOE.TryCatchError(func() (struct{}, error) {
		cmd := exec.CommandContext(ctx, spec.Bin)
		cmd.Args = append(cmd.Args, spec.Args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return struct{}{}, fmt.Errorf("container build failed: %w", err)
		}
		return struct{}{}, nil
	})
}
```

**What changed:**

- `Execute` uses `ioeither.WithResource` from fp-go directly — no custom
  bracket in `internal/fp/`.
- `ioeither.WithResource[struct{}](acquire, release)` returns a
  `Kleisli[error, Kleisli[error, DockerfileResource, struct{}], struct{}]`.
  We pass the `use` Kleisli via `F.Pipe1` to invoke it.
- `releaseResource` and `useResource` are extracted as named functions for
  readability and to satisfy `funlen` limits.
- `RenderCommand(r, res.Path)` is called _inside_ the `use` callback after
  the resource is acquired — no sentinel strings, no second IOEither resolution.
- `runProcess` extracted as a separate function to keep `Execute` concise.
- Removed import of `internal/fp` — no custom bracket needed.

---

## 7. File Placement

Copy the project-root `Dockerfile` into `internal/containerbuild/`:

```bash
cp Dockerfile internal/containerbuild/Dockerfile
```

The `//go:embed` directive requires the file to be in the same package
directory (or a subdirectory). This copy is the embedded version — future
Dockerfile changes should be mirrored here.

---

## 8. Updated CLI Surface

```
container-cli build [options]
  -f, --file <path>   Path to Dockerfile (default: "Dockerfile")
  -t, --tag <name>    Image tag, repeatable (default: "latest")
  -n, --name <name>   Image name (default: "container")
  --embed             Use the Dockerfile embedded in the binary
```

Usage:

```bash
# Use embedded Dockerfile (no file on disk needed)
container-cli build --embed -n myapp -t v1.0

# Use file-based Dockerfile (existing behavior, unchanged)
container-cli build -f Dockerfile -n myapp -t latest
```

When `--embed` and `--file` are both provided, `--embed` wins because the
`resolverFactories` map selects `EmbeddedResolver` when `embed=true`,
ignoring the `--file` value entirely.

---

## 9. Test Files

### 9.1 `internal/containerbuild/resource_test.go` (new)

```go
package build

import (
	"os"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestFileResolver_EmptyPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("")()

	require.True(E.IsLeft(result), "expected Left for empty path")

	err := F.Pipe1(
		result,
		E.Fold(
			F.Identity[error],
			func(DockerfileResource) error { return nil },
		),
	)
	require.EqualError(err, "dockerfile path is required")
}

func TestFileResolver_BlankPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("   ")()

	require.True(E.IsLeft(result), "expected Left for blank path")
}

func TestFileResolver_ValidPath(t *testing.T) {
	require := require.New(t)

	result := FileResolver("Dockerfile")()

	require.True(E.IsRight(result), "expected Right for valid path")

	res := F.Pipe1(
		result,
		E.Fold(
			func(error) DockerfileResource { return DockerfileResource{} },
			F.Identity[DockerfileResource],
		),
	)
	require.Equal("Dockerfile", res.Path)
}

func TestFileResolver_ReleaseIsNop(t *testing.T) {
	require := require.New(t)

	result := FileResolver("Dockerfile")()
	require.True(E.IsRight(result))

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)

	releaseResult := res.Release()
	require.True(E.IsRight(releaseResult), "file resolver release must succeed (nop)")
}

func TestEmbeddedResolver_WritesEmbeddedContent(t *testing.T) {
	require := require.New(t)

	result := EmbeddedResolver()()
	require.True(E.IsRight(result), "expected Right from EmbeddedResolver")

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)
	defer func() { _ = res.Release() }()

	content, err := os.ReadFile(res.Path)
	require.NoError(err)
	require.Equal(embeddedDockerfile, string(content))
}

func TestEmbeddedResolver_ReleaseCleansUp(t *testing.T) {
	require := require.New(t)

	result := EmbeddedResolver()()
	require.True(E.IsRight(result))

	res := E.GetOrElse(func(error) DockerfileResource {
		return DockerfileResource{}
	})(result)

	path := res.Path
	require.FileExists(path, "temp file must exist before release")

	releaseResult := res.Release()
	require.True(E.IsRight(releaseResult))
	require.NoFileExists(path, "temp file must be removed after release")
}
```

---

### 9.2 `internal/containerbuild/validate_test.go` (updated)

The `TestValidateInput_EmptyFile` test is removed — file path validation is now
tested in `resource_test.go` via `TestFileResolver_EmptyPath`. The remaining
tests update `Input` construction to omit the `File` field.

```go
package build

import (
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	"github.com/stretchr/testify/require"
)

func TestValidateInput_EmptyTagList(t *testing.T) {
	require := require.New(t)
	req := Input{
		Tags: []string{},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for empty tag list")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_BlankTagEntry(t *testing.T) {
	require := require.New(t)
	req := Input{
		Tags: []string{"latest", ""},
	}

	result := ValidateInput(req)

	require.True(E.IsLeft(result), "expected Left for blank tag entry")

	err := F.Pipe1(
		result,
		E.Fold(F.Identity[error], func(Input) error { return nil }),
	)
	require.NotNil(err)
	require.EqualError(err, "tag values must be non-empty")
}

func TestValidateInput_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for default input")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal("myapp", validated.Name)
	require.Equal([]string{"latest"}, validated.Tags)
}

func TestValidateInput_MultipleTags(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	result := ValidateInput(req)

	require.True(E.IsRight(result), "expected Right for valid build input with multiple tags")

	validated := F.Pipe1(
		result,
		E.Fold(func(error) Input { return Input{} }, F.Identity[Input]),
	)
	require.Equal("myapp", validated.Name)
	require.Equal([]string{"latest", "stable", "v1.0.0"}, validated.Tags)
}
```

---

### 9.3 `internal/containerbuild/args_test.go` (updated)

`RenderCommand(req)` becomes `RenderCommand(req, path)`. The return type
changes from `Input` (with embedded `CommandSpec`) to `CommandSpec` directly.

```go
package build

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderCommand_DefaultInput(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file Dockerfile --tag myapp:latest ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_RepeatedTags(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest", "stable", "v1.0.0"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file Dockerfile --tag myapp:latest --tag myapp:stable --tag myapp:v1.0.0 ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_DockerfileOverride(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "docker/Prod.Dockerfile")

	require.Equal(containerBinary, spec.Bin)

	expected := containerBinary + " build --file docker/Prod.Dockerfile --tag myapp:latest ."
	actual := spec.Bin + " " + strings.Join(spec.Args, " ")
	require.Equal(expected, actual)
}

func TestRenderCommand_FinalArgIsBuildContext(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.NotEmpty(spec.Args)

	lastArg := spec.Args[len(spec.Args)-1]
	require.Equal(".", lastArg)
}

func TestRenderCommand_ArgsOrder(t *testing.T) {
	require := require.New(t)
	req := Input{
		Name: "myapp",
		Tags: []string{"latest", "stable"},
	}

	spec := RenderCommand(req, "Dockerfile")

	require.GreaterOrEqual(len(spec.Args), 7)

	require.Equal("build", spec.Args[0])
	require.Equal("--file", spec.Args[1])
	require.Equal("Dockerfile", spec.Args[2])
	require.Equal("--tag", spec.Args[3])
	require.Equal("myapp:latest", spec.Args[4])
	require.Equal("--tag", spec.Args[5])
	require.Equal("myapp:stable", spec.Args[6])
}
```

---

## 10. Implementation Order

The order matters because later steps depend on types/functions introduced
earlier. Each step ends with `gotestsum` to catch regressions immediately.

| Step | File(s) | Rationale |
|---|---|---|
| 1 | `internal/containerbuild/Dockerfile` (copy) | Required by `//go:embed` in step 2 |
| 2 | `internal/containerbuild/embed.go` | Makes `embeddedDockerfile` available |
| 3 | `internal/containerbuild/input.go` | New `Input` type — breaks compilation of everything else intentionally |
| 4 | `internal/containerbuild/resource.go` + `resource_test.go` | `DockerfileResource`, resolvers — depends on new `Input` and `embed.go` |
| 5 | `internal/containerbuild/validate.go` + `validate_test.go` | Remove file validation, update tests |
| 6 | `internal/containerbuild/args.go` + `args_test.go` | New `RenderCommand` signature, update tests |
| 7 | `internal/containerbuild/exec.go` | `ioeither.WithResource` bracket, `runProcess` split |
| 8 | `internal/containerbuild/command.go` | `--embed` flag, `InputFromCommand`, `F.Pipe5` pipeline |

After step 8, the full build should compile and all tests should pass:

```bash
go build ./cmd/container-cli/...
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...
golangci-lint run ./...
```

---

## 11. Files Summary

| File | Action |
|---|---|
| `internal/containerbuild/Dockerfile` | **Create** — copy from project root |
| `internal/containerbuild/embed.go` | **Create** — `//go:embed` directive |
| `internal/containerbuild/resource.go` | **Create** — `DockerfileResource`, `FileResolver`, `EmbeddedResolver` |
| `internal/containerbuild/resource_test.go` | **Create** — resolver lifecycle tests |
| `internal/containerbuild/input.go` | **Modify** — replace `File` + `CommandSpec` with `DockerfileSource` |
| `internal/containerbuild/validate.go` | **Modify** — tags-only validation |
| `internal/containerbuild/validate_test.go` | **Modify** — remove file test, update `Input` construction |
| `internal/containerbuild/args.go` | **Modify** — `RenderCommand(r Input, path string) CommandSpec` |
| `internal/containerbuild/args_test.go` | **Modify** — update call signatures and assertions |
| `internal/containerbuild/exec.go` | **Modify** — `ioeither.WithResource` bracket, split `runProcess` |
| `internal/containerbuild/command.go` | **Modify** — `--embed` flag, map dispatch, `F.Pipe5` |

---

## 12. Gotchas

- **`//go:embed` path**: The file must be in the package directory or below.
  `//go:embed ../../Dockerfile` does **not** work — the Dockerfile must be
  physically placed at `internal/containerbuild/Dockerfile`.

- **`IOEither` is lazy**: `DockerfileSource` is a `func() Either[error, DockerfileResource]`.
  It is NOT executed when `InputFromCommand` builds the `Input` — only when
  `WithResource` calls `acquire()` inside `Execute`.

- **Resolver map dispatch**: `resolverFactories[cmd.Bool("embed")]` always
  succeeds because `bool` only has two values and the map covers both. No
  default case needed.

- **Temp file lifecycle**: `EmbeddedResolver` uses `IOEF.Write[string, *os.File]`
  which is a bracket combinator — it acquires the `*os.File` from `CreateTemp`,
  passes it to the write Kleisli, and **always closes** the file handle, even
  if the write fails. The Kleisli returns `f.Name()` so the path survives
  after close. File **removal** is deferred to `ioeither.WithResource`'s
  release callback in `Execute`, which runs after the container process
  completes (success or failure).

- **`--embed` + `--file` precedence**: When both are provided, `--embed` wins
  silently. The `--file` value is never read. The `--embed` flag's usage string
  documents this: `"Use the Dockerfile embedded in the binary (ignores --file)"`.

- **`isNonBlank` sharing**: Defined in `validate.go`, used by `FileResolver`
  in `resource.go`. Both are in the same package — no visibility issues.

- **fp-go type parameters**: `IOE.FromEither[error]` uses Go 1.21+ partial type
  instantiation — the second type parameter (`A`) is inferred from context by
  the compiler within `F.Pipe5`.

- **`ioeither.WithResource` signature**: Returns a curried
  `Kleisli[E, Kleisli[E, R, A], A]`. You call it with `(acquire, release)` to
  get a function, then pass the `use` Kleisli. This is why `Execute` uses
  `F.Pipe1(useResource(r), IOE.WithResource[struct{}](acquire, release))`.

- **`IOEF.Remove` returns `IOEither[error, string]`**: It wraps `os.Remove` and
  returns the path on success. This is the release type parameter `ANY` in
  `ioeither.WithResource` — the return value is discarded.

- **`E.GetOrElse` in tests**: Returns the Right value or a fallback. Useful for
  extracting values in test assertions without `E.Fold` boilerplate.

- **New imports**: `IOEF` alias for `github.com/IBM/fp-go/v2/ioeither/file` —
  used in `resource.go` and `exec.go`. Follow the existing alias conventions
  in the project.

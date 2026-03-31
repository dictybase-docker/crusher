# Functional CLI App Plan

## 1. Objective

Build a standalone planning brief that can be handed to any agentic AI system to implement a Go CLI application in this repository. The brief must capture all user feedback given so far and keep the implementation constrained to a `container build` wrapper built with `github.com/urfave/cli/v3` and `github.com/IBM/fp-go/v2`.

The objective now includes all of the following:

- current narrowed CLI surface:
  - Dockerfile path, default `Dockerfile`
  - repeatable tag list, default `latest`
  - fixed build context `.` with no user-facing flag
- source paths explicitly requested for fp-go code and semantics review:
  - `/Users/sba964/Projects/devenv/golang/learn-golang/grpc/plasmid/goldenbraid`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/fp-go-concepts/v2`
  - `/Users/sba964/Projects/devenv/golang/learn-golang/llm/adk/devops-engineer`
- strict functional-style instructions that must govern the implementation:
  - no imperative branching in application code samples
  - validation, defaults, optional behavior, and terminal control expressed through `Either`, `Option`, and `IOEither` combinators
  - effects isolated inside `IOEither`
  - direct argv construction for `exec.CommandContext`, never shell-string construction
- important findings discovered while developing the plan:
  - `command-reference.md` is the local source of truth for `container build`
  - `Dockerfile` already exists at the repository root
  - the scanned Either examples rely on combinators such as `E.FromPredicate`, `E.SequenceArray`, `E.FromOption`, `E.GetOrElse`, and `E.Fold`
  - planning guidance is grounded in inspected local code and local command documentation
  - the current repository still has no Go source files yet

The concrete command in scope remains a build command that drives `container build`.

## 2. Confirmed Inputs

### 2.1 Repository facts

- Repository root files currently present:
  - `Dockerfile`
  - `README.md`
  - `command-reference.md`
  - `go.mod`
- Current module path: `github.com/cybersiddhu/crush-sandbox` (`go.mod:1`)
- Current Go version: `1.25.7` (`go.mod:3`)
- Current repository has no Go source files yet, so project diagnostics report `no go files to analyze`
- `Dockerfile:1-31` already exists at repository root

### 2.2 Current narrowed CLI requirements

At this stage the CLI should expose only these user-facing options for the build command:

1. path to Dockerfile, default `Dockerfile` in the current folder
2. list of tag names, default `latest`

The build context remains the current folder and is rendered as `.` with no user-facing option.

### 2.3 CLI structure references already inspected

- `grpc/plasmid/goldenbraid/cmd/goldenbraid-list/main.go:14-112`
  - confirms `urfave/cli/v3` root command with multiple subcommands
- `llm/adk/devops-engineer/cmd/agent/main.go:29-40`
  - confirms minimal `cli.Command` root wiring and `app.Run(context.Background(), os.Args)`

### 2.4 Either-style scan across `fp-go-concepts/v2/either`

Scanned files:

- `fp-go-concepts/v2/either/01_fundamentals.go`
- `fp-go-concepts/v2/either/03_combinations.go`
- `fp-go-concepts/v2/either/04_real_world_patterns.go`
- `fp-go-concepts/v2/either/README.md`

Discrete findings:

- `fp-go-concepts/v2/either/01_fundamentals.go:38-49` uses `E.FromPredicate` plus `E.Map` for safe transformation
- `fp-go-concepts/v2/either/01_fundamentals.go:54-68` uses chained `E.FromPredicate` for range validation
- `fp-go-concepts/v2/either/01_fundamentals.go:72-83` uses `E.Fold` for terminal branching
- `fp-go-concepts/v2/either/03_combinations.go:43-65` uses `[]Either` plus `E.SequenceArray` plus `E.MapTo` for fail-fast batch validation
- `fp-go-concepts/v2/either/03_combinations.go:72-90` uses sequential `E.Chain` for dependent validation
- `fp-go-concepts/v2/either/04_real_world_patterns.go:76-94` uses `R.Lookup` plus `E.FromOption` for lookup validation
- `fp-go-concepts/v2/either/04_real_world_patterns.go:243-296` uses `E.GetOrElse` for defaults
- `fp-go-concepts/v2/either/04_real_world_patterns.go:443-456` uses `E.Fold` to model optional behavior
- `fp-go-concepts/v2/either/04_real_world_patterns.go:548-564` uses `E.Fold` plus `A.FoldMap` for partial-success batch processing

### 2.5 fp-go style references outside `either`

- `grpc/plasmid/goldenbraid/internal/wait/action.go:48-66`
  - confirms `IOE.Of` → `IOE.Bind` → `IOE.Let` → `IOE.Chain` → `ToEither` → `E.Fold`
- `grpc/plasmid/goldenbraid/internal/wait/poller.go:76-86`
  - confirms point-free `F.PipeN` orchestration around `IOEither`
- `grpc/plasmid/goldenbraid/internal/fputil/convert.go:12-17`
  - confirms a simple `ToEither` helper pattern
- `fp-go-concepts/v2/do-bind/ioeither_examples.go:171-186`
  - confirms `IOE.Bind` plus `IOE.Let` do-notation style

### 2.6 urfave/cli v3 API facts confirmed from local module cache

Confirmed from `github.com/urfave/cli/v3@v3.6.2` in the local module cache:

- `cli.Command` supports `Commands []*Command`, `Flags []Flag`, `Action ActionFunc`, and `Run(context.Context, []string)` (`command.go:20-68`)
- `cli.StringFlag` is available (`flag_string.go:8-14`)
- `cli.StringSliceFlag` is available (`flag_string_slice.go:3-8`)
- `(*cli.Command).String(name)` is available (`flag_string.go:56-64`)
- `(*cli.Command).StringSlice(name)` is available (`flag_string_slice.go:10-20`)
- `FlagBase` supports `Value`, so `StringSliceFlag` can carry a concrete default slice (`flag_impl.go:57-74`)

### 2.7 fp-go API facts confirmed from local module cache

Confirmed from `github.com/IBM/fp-go/v2@v2.2.6` in the local module cache:

- `IOE.Do[E, S any](empty S)` exists (`ioeither/bind.go:26-40`)
- `IOE.Bind` exists (`ioeither/bind.go:79-89`)
- `IOE.Let` exists (`ioeither/bind.go:91-101`)
- `A.Reduce` exists (`array/array.go:198-208`)
- `A.Append` exists (`array/array.go:241-255`)
- `A.IsNonEmpty` exists (`array/array.go:265-268`)
- `A.Empty` exists (`array/array.go:270-275`)
- `A.Chain` exists (`array/array.go:299-309`)
- `A.Flatten` exists (`array/array.go:419`)
- `E.TryCatchError` exists (`either/either.go:294`)
- `F.Identity` exists (`function/function.go:34`)

### 2.8 Container command reference facts

From `command-reference.md`:

- `container build` usage is:

```bash
container build [<options>] [<context-dir>]
```

- build context defaults to `.` (`command-reference.md:128-130`)
- documented options include `-f, --file <path>` and `-t, --tag <name>`
- documented build examples include:

```bash
container build -t my-app:latest .
container build -f docker/Dockerfile.prod -t my-app:prod .
container build -t my-app:latest -t my-app:v1.0.0 -t my-app:stable .
```

## 3. Non-Negotiable Functional Rules

The implementation plan follows the style proven by the scanned Either examples:

- no imperative branching in application code samples
- validation through `E.FromPredicate`, `E.Chain`, `E.SequenceArray`, `E.MapTo`, `E.FromOption`, and `E.GetOrElse`
- terminal branching through `E.Fold`
- side effects isolated inside `IOEither`
- `exec.CommandContext` receives a direct argv slice, never a shell string

## 4. Scope Decision

### 4.1 Phase 1 in scope

Implement a CLI with:

- a root command
- a `build` subcommand
- exactly two build options:
  - Dockerfile path
  - repeatable tag list
- default Dockerfile path of `Dockerfile`
- default tag list of `latest`
- fixed build context of `.`
- fp-go based validation, transformation, rendering, and execution pipeline
- unit tests for request validation and argv rendering

### 4.2 Out of scope

- any additional build flags
- registry flows
- runtime container flows
- user-facing build context option

## 5. Exact Dependency Plan

Pin these during implementation:

- `github.com/urfave/cli/v3 v3.6.2`
- `github.com/IBM/fp-go/v2 v2.2.6`

Grounding:

- `urfave/cli/v3 v3.6.2` is already used in `goldenbraid/go.mod:9` and `devops-engineer/go.mod:8`
- `fp-go/v2 v2.2.6` is already used in `goldenbraid/go.mod:6`
- the exact fp-go APIs listed above were confirmed from the local module cache for `v2.2.6`

Expected standard library imports:

- `context`
- `errors`
- `fmt`
- `os`
- `os/exec`
- `strings`

## 6. Planned Project Layout

```text
cmd/
  container-cli/
    main.go
internal/
  build/
    request.go
    validate.go
    args.go
    exec.go
    command.go
    validate_test.go
    args_test.go
```

Reasons:

- `cmd/.../main.go` matches the inspected CLI references
- `internal/build` isolates the single concrete subcommand
- `request.go` holds the domain type; `command.go` holds the CLI wiring and action pipeline
- no `internal/fputil` or `internal/app` — the pipeline is flat enough to not need conversion helpers or a separate app package

## 7. Root Command Contract

The root command must:

- use `urfave/cli/v3`
- expose `build` as the first concrete subcommand
- surface terminal errors through `stderr`
- produce exit code `1` on failure and `0` on success

### Root command code sample

Follows the `goldenbraid-list/main.go` entry-point pattern exactly:

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/cybersiddhu/crush-sandbox/internal/build"
    "github.com/urfave/cli/v3"
)

func main() {
    app := &cli.Command{
        Name:  "container-cli",
        Usage: "Build OCI images through the container CLI",
        Commands: []*cli.Command{
            build.Command(),
        },
    }
    if err := app.Run(context.Background(), os.Args); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

The entry point uses idiomatic Go error handling intentionally — fp-go
combinators govern the *application* code inside `internal/build`, not the
process bootstrap. This matches every inspected reference (`goldenbraid-list/main.go`,
`devops-engineer/cmd/agent/main.go`).

## 8. Build Subcommand Surface

### 8.1 Required behavior

The `build` subcommand must construct and execute a `container build ...` command.

### 8.2 User-facing flags

| Wrapper flag | Container mapping | Default | Notes |
| --- | --- | --- | --- |
| `--file`, `-f` | `-f, --file <path>` | `Dockerfile` | path to Dockerfile in current folder by default |
| `--tag`, `-t` | `-t, --tag <name>` | `latest` | repeatable |

The build context is always rendered as `.`.

### 8.3 urfave/cli v3 flag definition sample

```go
package build

import "github.com/urfave/cli/v3"

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
        },
        Action: Action,
    }
}
```

## 9. Domain Model Plan

```go
package build

// Request holds the validated build parameters extracted from CLI flags.
type Request struct {
    File string
    Tags []string
}

// CommandSpec holds the resolved executable name and argv slice.
type CommandSpec struct {
    Name string
    Args []string
}
```

Only two types are needed. The pipeline threads `Request → CommandSpec → unit`
through a flat `F.Pipe` chain — no accumulator struct required.

## 10. Validation Plan

### 10.1 Required validations

- `File` must be non-empty
- tag list must be non-empty
- every tag value must be non-empty

### 10.2 Validator sample aligned with the scanned Either patterns

```go
package build

import (
    "errors"
    "strings"

    A "github.com/IBM/fp-go/v2/array"
    E "github.com/IBM/fp-go/v2/either"
    F "github.com/IBM/fp-go/v2/function"
    Str "github.com/IBM/fp-go/v2/string"
)

var nonEmpty = F.Flow2(strings.TrimSpace, Str.IsNonEmpty)

func allNonEmpty(values []string) bool {
    return F.Pipe1(
        values,
        A.Reduce(func(acc bool, item string) bool {
            return acc && nonEmpty(item)
        }, true),
    )
}

func ValidateRequest(r Request) E.Either[error, Request] {
    validations := []E.Either[error, bool]{
        F.Pipe2(
            r.File,
            E.FromPredicate(
                nonEmpty,
                func(string) error { return errors.New("Dockerfile path is required") },
            ),
            E.MapTo[error, string](true),
        ),
        F.Pipe2(
            r.Tags,
            E.FromPredicate(
                A.IsNonEmpty[string],
                func([]string) error { return errors.New("at least one tag is required") },
            ),
            E.MapTo[error, []string](true),
        ),
        F.Pipe2(
            r.Tags,
            E.FromPredicate(
                allNonEmpty,
                func([]string) error { return errors.New("tag values must be non-empty") },
            ),
            E.MapTo[error, []string](true),
        ),
    }

    return F.Pipe2(
        E.SequenceArray(validations),
        E.MapTo[error, []bool](r),
    )
}
```

Grounding for this style:

- chained and predicate-based validation: `fp-go-concepts/v2/either/01_fundamentals.go:54-68`
- batch fail-fast validation: `fp-go-concepts/v2/either/03_combinations.go:43-65`

## 11. CLI Extraction Plan

```go
package build

import "github.com/urfave/cli/v3"

func RequestFromCommand(cmd *cli.Command) Request {
    return Request{
        File: cmd.String("file"),
        Tags: cmd.StringSlice("tag"),
    }
}
```

## 12. Argument Rendering Plan

The wrapper must avoid shell string construction. It must build an argv slice and call `exec.CommandContext` directly.

### 12.1 Rendering rule

Given a valid `Request`, emit:

```text
container build --file <path> --tag <value> ... .
```

### 12.2 Rendering sample

```go
package build

import (
    A "github.com/IBM/fp-go/v2/array"
    F "github.com/IBM/fp-go/v2/function"
)

func repeated(flag string) func([]string) []string {
    return A.Chain(func(value string) []string {
        return []string{flag, value}
    })
}

func RenderCommand(r Request) CommandSpec {
    return CommandSpec{
        Name: "container",
        Args: F.Pipe1(
            [][]string{
                {"build"},
                {"--file", r.File},
                F.Pipe1(r.Tags, repeated("--tag")),
                {"."},
            },
            A.Flatten[string],
        ),
    }
}
```

Expected minimal rendering with defaults:

```text
container build --file Dockerfile --tag latest .
```

## 13. Execution Plan

### 13.1 Runtime behavior

- invoke `container` from `PATH`
- stream stdout and stderr directly to the terminal
- return the underlying execution failure through `Left`
- avoid shell execution completely

### 13.2 Executor sample

Uses the canonical `IOE.TryCatchError` pattern from
`ioeither/01_fundamentals.go` — no manual `func() Either` construction:

```go
package build

import (
    "context"
    "fmt"
    "os"
    "os/exec"

    IOE "github.com/IBM/fp-go/v2/ioeither"
)

func Execute(ctx context.Context, spec CommandSpec) IOE.IOEither[error, struct{}] {
    return IOE.TryCatchError(func() (struct{}, error) {
        cmd := exec.CommandContext(ctx, spec.Name, spec.Args...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return struct{}{}, fmt.Errorf("container build failed: %w", err)
        }
        return struct{}{}, nil
    })
}
```

Grounding: `IOE.TryCatchError(func() (A, error) {...})` is the exact signature
used in `ioeither/01_fundamentals.go` (`fetchUser`) and
`devops-engineer/internal/runner/runner.go` (`createModel`).

## 14. fp-go Action Pipeline Plan

The pipeline uses a **flat `F.Pipe` chain** with `IOE.ChainEitherK` and
`IOE.Chain` — no accumulator struct, no setter functions, no `fputil.ToEither`.

```go
package build

import (
    "context"

    E "github.com/IBM/fp-go/v2/either"
    F "github.com/IBM/fp-go/v2/function"
    IOE "github.com/IBM/fp-go/v2/ioeither"
    "github.com/urfave/cli/v3"
)

func Action(ctx context.Context, cmd *cli.Command) error {
    req := RequestFromCommand(cmd)

    program := F.Pipe3(
        IOE.FromEither[error](ValidateRequest(req)),
        IOE.Map[error](RenderCommand),
        IOE.Chain(func(spec CommandSpec) IOE.IOEither[error, struct{}] {
            return Execute(ctx, spec)
        }),
    )

    return F.Pipe1(
        program(),
        E.Fold(
            F.Identity[error],
            func(struct{}) error { return nil },
        ),
    )
}
```

### Pipeline shape

```text
Request
  → ValidateRequest        (pure: Either[error, Request])
  → IOE.FromEither         (lift into IOEither)
  → IOE.Map(RenderCommand) (pure: Request → CommandSpec)
  → IOE.Chain(Execute)     (effect: CommandSpec → IOEither[error, struct{}])
  → program()              (run the lazy IO)
  → E.Fold(…)              (terminal: Either → error)
```

Three pipeline steps, zero helper types, zero setter functions.

Grounding:

- `IOE.FromEither` + `IOE.Chain` flat pipe: `ioeither/01_fundamentals.go` (`getUserTotal`)
- `IOE.ChainEitherK` for pure validation in IOEither pipe: `ioeither/05_file_io_patterns.go` (`loadConfig`)
- `E.Fold` at the terminal boundary: `goldenbraid/internal/wait/action.go:62-66`
- `program()` to execute lazy IO: `ioeither/01_fundamentals.go` (all examples)

## 16. Testing Plan

### 16.1 Unit tests to write

`internal/build/args_test.go` must cover:

- default request renders `container build --file Dockerfile --tag latest .`
- repeated `--tag` values preserve order
- Dockerfile override renders the supplied path
- final argv item remains `.`

`internal/build/validate_test.go` must cover:

- empty Dockerfile path returns `Left`
- empty tag list returns `Left`
- blank tag entry returns `Left`
- default request returns `Right`

`internal/build/action_test.go` must keep process execution injectable so tests can verify constructed `CommandSpec` without executing a real `container` binary.

### 16.2 Testability

Since `Execute` accepts a `CommandSpec`, tests can verify the full pipeline
up to the point of execution by calling `ValidateRequest` and `RenderCommand`
directly — both are pure functions. Integration tests that need to mock
process execution can inject a replacement `Execute` function:

```go
type Executor func(context.Context, CommandSpec) IOE.IOEither[error, struct{}]
```

### 16.3 Test command

Once source files exist:

```bash
go test ./...
```

## 17. Implementation Order

1. Update `go.mod` with `urfave/cli/v3` and `fp-go/v2`
2. Add `cmd/container-cli/main.go`
3. Add `internal/build/request.go` (Request, CommandSpec types)
4. Add `internal/build/validate.go`
5. Add `internal/build/args.go`
6. Add `internal/build/exec.go`
7. Add `internal/build/command.go` (Command(), Action, RequestFromCommand)
8. Add `internal/build/validate_test.go`
9. Add `internal/build/args_test.go`
10. Run `go test ./...`

## 18. Concrete Acceptance Criteria

The implementation is complete when all of the following are true:

- running the binary with `--help` shows a root command and a `build` subcommand
- running the binary with `build --help` shows only Dockerfile path and tag list flags
- running `build` with no flags renders and executes `container build --file Dockerfile --tag latest .`
- running `build -f docker/Prod.Dockerfile -t latest -t stable` renders `container build --file docker/Prod.Dockerfile --tag latest --tag stable .`
- validation and argv rendering stay within Either and IOEither combinators
- unit tests pass with `go test ./...`

## 19. Drift Guards

- do not bypass `container build`
- do not build shell strings
- keep build context fixed to `.`
- keep validation pure and separate from process execution
- keep side effects inside `IOEither`
- keep the public build surface limited to Dockerfile path and repeatable tags until requirements expand

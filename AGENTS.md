# AGENTS.md

## Project Overview

Go CLI application (`container-cli`) that wraps Apple's [`container build`](https://github.com/apple/container) command for building OCI images. The entire application layer uses **functional programming** via `github.com/IBM/fp-go/v2` — no imperative branching in application code.

- **Module**: `github.com/cybersiddhu/crush-sandbox`
- **Go version**: `1.25.7`
- **Binary name**: `container-cli`

---

## Essential Commands

```bash
# Build
go build ./cmd/container-cli/...

# Test
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Test (verbose, per-test detail)
gotestsum --format testdox --format-hide-empty-pkg -- ./...

# Watch mode (re-runs on file change)
gotestsum --watch --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Lint
golangci-lint run ./...

# Format
golangci-lint fmt

# Run
go run ./cmd/container-cli/... --help
go run ./cmd/container-cli/... build --help
go run ./cmd/container-cli/... build -f Dockerfile -t myapp:latest

# Package discovery
go list -f '{{.ImportPath}} => {{.Dir}}' ./...
```

---

## Project Layout

```
cmd/
  container-cli/
    main.go              # Entry point – root cli.Command, wires build.Command()
internal/
  build/
    input.go           # Input and CommandSpec domain types
    validate.go          # Pure Either-based validation
    args.go              # Pure argv rendering (Input → CommandSpec)
    exec.go              # IOEither-based process execution
    command.go           # cli.Command(), Action, InputFromCommand
    validate_test.go     # Unit tests for ValidateInput
    args_test.go         # Unit tests for RenderCommand
docs/
  PLAN-build-cmd.md      # Detailed implementation plan (ground truth for architecture)
  command-reference.md   # Apple container CLI reference (source of truth for flags)
  PROMPT-build-cmd.md    # Original prompt used to generate the build subcommand
  crush-readme.md
Dockerfile               # Docker sandbox image (gopls, golangci-lint, gotestsum, ripgrep)
go.mod / go.sum
```

---

## Dependencies

| Package | Version | Role |
|---|---|---|
| `github.com/urfave/cli/v3` | v3.6.2 | CLI framework |
| `github.com/IBM/fp-go/v2` | v2.2.6 | Functional programming (Either, IOEither, Array, Function) |
| `github.com/stretchr/testify` | v1.11.1 | Test assertions |

---

## CLI Surface (build subcommand)

```
container-cli build [options]
  -f, --file <path>   Path to Dockerfile (default: "Dockerfile")
  -t, --tag <name>    Image tag, repeatable (default: "latest")
```

Build context is always fixed to `.` — no user-facing flag for it.

Resulting `container` invocation:

```
container build --file <path> --tag <val> [--tag <val> ...] .
```

---

## Functional Programming Conventions

These rules are **non-negotiable** and apply to all code in `internal/`:

### No imperative branching

- Never use `if`/`else` for application logic — use `E.Fold`, `E.FromPredicate`, `E.Chain`
- Side effects (process execution) must be isolated inside `IOEither`
- `exec.CommandContext` receives a direct argv slice — **never** build a shell string
- Validation must be pure and separate from process execution
- Use `F.Pipe` chains for composition

### Key fp-go combinators

| Combinator | Purpose |
|---|---|
| `E.FromPredicate(pred, errFn)` | Lift a value into Either based on a predicate |
| `E.Chain` | Sequence dependent Either operations |
| `E.SequenceArray` | Fail-fast batch validation over `[]Either` |
| `E.MapTo[E, A](b)` | Replace the Right value (discard A, keep B) |
| `E.Fold(onLeft, onRight)` | Terminal branching at the boundary |
| `E.GetOrElse(default)` | Extract value from Either with a fallback |
| `IOE.TryCatchError(func() (A, error))` | Wrap a fallible effect into IOEither |
| `IOE.FromEither` | Lift a pure Either into IOEither |
| `IOE.Map` | Pure transform inside IOEither |
| `IOE.Chain` | Sequence effectful IOEither operations |
| `A.Chain` | FlatMap over arrays (used for repeated flag expansion) |
| `A.Flatten` | Flatten `[][]string` to `[]string` for argv construction |
| `A.IsNonEmpty` | Predicate: array has at least one element |
| `F.Pipe1/2/3` | Pipe a value through N transforms |

### Import alias conventions

| Alias | Package |
|---|---|
| `E` | `either` |
| `IOE` | `ioeither` |
| `A` | `array` |
| `F` | `function` |
| `Str` | `string` |

### Canonical pipeline shape (Action)

```go
func Action(ctx context.Context, cmd *cli.Command) error {
    req := InputFromCommand(cmd)

    program := F.Pipe3(
        IOE.FromEither[error](ValidateInput(req)),
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

### Domain types

```go
type Input struct {
    File string   // Dockerfile path, validated non-empty
    Tags []string // Tag list, validated non-empty, all items non-empty
}

type CommandSpec struct {
    Name string   // "container"
    Args []string // ["build", "--file", <path>, "--tag", <val>, ..., "."]
}
```

### Validation style

```go
func ValidateInput(r Input) E.Either[error, Input] {
    validations := []E.Either[error, bool]{
        F.Pipe2(r.File,
            E.FromPredicate(nonEmpty, func(string) error { return errors.New("...") }),
            E.MapTo[error, string](true),
        ),
        // ... more checks
    }
    return F.Pipe2(E.SequenceArray(validations), E.MapTo[error, []bool](r))
}
```

### Argv rendering style (repeated flags)

```go
func repeated(flag string) func([]string) []string {
    return A.Chain(func(value string) []string { return []string{flag, value} })
}

func RenderCommand(r Input) CommandSpec {
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

### Entrypoint pattern

Entry point uses idiomatic Go error handling; fp-go governs application code inside `internal/`:

```go
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

---

## Naming and Style Conventions

- Package names are short/lowercase (`build`)
- Exported domain types are nouns: `Input`, `CommandSpec`
- Validation returns `E.Either[error, Input]` rather than `error` + mutation
- Effects are isolated in `Execute` returning `IOE.IOEither[error, struct{}]`
- Process execution always uses direct argv construction with `exec.CommandContext`, never shell strings

---

## Testing Conventions

- Tests live alongside source in `internal/build/`
- Pure functions (`ValidateInput`, `RenderCommand`) are tested directly — no mocking needed
- Tests that need to verify the full pipeline without running a real `container` binary inject a replacement `Execute` function via `type Executor func(context.Context, CommandSpec) IOE.IOEither[error, struct{}]`
- Prefer adding/changing unit tests in `internal/build/*_test.go` with any behavior changes
- Verify with `gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...` after every modification

### Required test coverage

**`validate_test.go`**:

- Empty Dockerfile path → `Left`
- Empty tag list → `Left`
- Blank tag entry → `Left`
- Default request (non-empty file, non-empty tags) → `Right`

**`args_test.go`**:

- Default request renders `container build --file Dockerfile --tag latest .`
- Multiple tags preserve insertion order
- Custom Dockerfile path is reflected in argv
- Final argv element is always `.`

---

## Linting

The project uses `golangci-lint` v2 with strict configuration (`.golangci.yml`):

### Enabled linters

- `gosec` — Security
- `staticcheck` — Static analysis
- `revive` — Style
- `govet` — Go vet
- `gocyclo`, `cyclop` — Complexity
- `funlen` — Function length (80 lines, 50 statements)
- `gocognit` — Cognitive complexity
- `ineffassign` — Ineffective assignments
- `unconvert` — Unnecessary conversions
- `unparam` — Unused parameters
- `unused` — Unused code

### Formatters

- `gofumpt`
- `goimports`
- `golines`

---

## Environment (Dockerfile)

The Docker sandbox image provides:

- `gopls` — Go LSP
- `golangci-lint v2.11.4` — Linter
- `gotestsum` — Enhanced test runner
- `ripgrep`, `fd-find` — Fast file search tools

---

## Workflow for Agents

1. Start from `cmd/container-cli/main.go` and `internal/build/*` to understand command flow.
2. Use `docs/command-reference.md` for `container build` flag/behavior truth.
3. Use `docs/PLAN-build-cmd.md` as ground truth for architecture decisions.
4. Preserve fp-go functional composition style in `internal/build`.
5. Prefer adding/changing unit tests in `internal/build/*_test.go` with any behavior changes.
6. Verify with `gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...` after modifications.
7. Run `golangci-lint run ./...` to check for lint violations.

---

## Gotchas

- **Module path** is `github.com/cybersiddhu/crush-sandbox`.
- **`container` binary** must be available in PATH at runtime.
- **Build context is always `.`** — never expose this to users.
- **`command-reference.md`** is the local source of truth for `container build` flags — consult it before adding new options.
- **`urfave/cli/v3` API**: use `cmd.String("flag-name")` and `cmd.StringSlice("flag-name")` inside Action to read flag values; flag defaults are set on `StringFlag.Value` and `StringSliceFlag.Value`.
- **fp-go generics**: type parameters are often required explicitly (e.g. `E.MapTo[error, string](true)`) because Go cannot infer them from the discard semantics.
- **`IOEither` is lazy**: `program` is a `func() Either[...]` — you must call `program()` to trigger execution.
- **Never use `F.Pipe` with more than its declared arity** — use `F.Pipe1`, `F.Pipe2`, `F.Pipe3`, etc. matching the number of transform arguments.
- **Use `E.Fold` for terminal branching**, not `if` statements in application code.
- **LSP diagnostics** may report `no go files to analyze` at the module root when `internal/` is empty — this is expected.
- **Planning docs** (`PLAN-build-cmd.md`, `PROMPT-build-cmd.md`) contain planning history; parts may be stale relative to the current source.

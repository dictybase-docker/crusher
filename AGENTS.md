# AGENTS.md

## Project Overview

Go CLI application (`container-cli`) that wraps Apple's [`container build`](https://github.com/apple/container) command for building OCI images. The application layer uses **functional programming** via `github.com/IBM/fp-go/v2` — no imperative branching in application code.

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

# Test (verbose)
gotestsum --format testdox --format-hide-empty-pkg -- ./...

# Watch mode
gotestsum --watch --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Lint
golangci-lint run ./...

# Format
golangci-lint fmt

# Run
go run ./cmd/container-cli/... build --help
```

---

## Project Layout

```
cmd/container-cli/main.go          # Entry point
internal/
  fp/conversion.go                 # FP utilities (ToEither helper)
  containerbuild/
    command.go                     # CLI flags, InputFromCommand, Action
    input.go                       # Input, CommandSpec, DockerfileResource types
    validate.go                    # Pure Either-based validation
    args.go                        # Pure argv rendering
    exec.go                        # IOEither-based process execution with resource management
    resource.go                    # FileResolver, EmbeddedResolver
    embed.go                       //go:embed Dockerfile
    Dockerfile                     # Embedded Dockerfile (default image)
    *_test.go                      # Unit tests
docs/
  command-reference.md             # Apple container CLI reference
```

---

## Dependencies

| Package | Version | Role |
|---|---|---|
| `github.com/urfave/cli/v3` | v3.6.2 | CLI framework |
| `github.com/IBM/fp-go/v2` | v2.2.6 | Functional programming |
| `github.com/stretchr/testify` | v1.11.1 | Test assertions |

---

## CLI Surface

```
container-cli build [options]
  -n, --name <string>              Image name (default: "crusher")
  -t, --tag <string>               Image tag, repeatable (default: "latest")
  -f, --file <path>                Path to Dockerfile (default: "Dockerfile")
  --embed                          Use embedded Dockerfile (ignores --file)
  --golangci-lint-version <ver>    golangci-lint version (default: "2.11.4")
  --crush-version <ver>            crush version (default: "latest")
  --gotestsum-version <ver>        gotestsum version (default: "latest")
```

Build context is always `.`. Resulting `container` invocation:

```
container build --file <path> --tag <name>:<tag> [--build-arg KEY=VAL ...] .
```

---

## Domain Types

```go
type Input struct {
    DockerfileSource IOE.IOEither[error, DockerfileResource]  // Lazy resolver
    Name             string                                    // Image name
    Tags             []string                                  // Tags (validated non-empty)
    BuildArgs        map[string]string                         // Build arguments
    Ctx              context.Context
}

type CommandSpec struct {
    Bin  string      // "container"
    Args []string    // ["build", "--file", <path>, "--tag", <name>:<tag>, ..., "."]
}

type DockerfileResource struct {
    Path    string                         // Resolved Dockerfile path
    Release IOE.IOEither[error, string]    // Cleanup (nop for file, Remove for embedded)
}
```

---

## Functional Programming Conventions

### Core rules

- **No imperative branching** — use `E.Fold`, `E.FromPredicate`, `E.Chain`
- **Side effects isolated in IOEither** — `exec.CommandContext` receives direct argv slice, never shell strings
- **Validation is pure** — separate from process execution
- **Use `F.Pipe1/2/3/etc`** — match arity to transform count

### Key fp-go combinators

| Combinator | Purpose |
|---|---|
| `E.FromPredicate(pred, errFn)` | Lift value into Either based on predicate |
| `E.Chain` | Sequence dependent Either operations |
| `E.Fold(onLeft, onRight)` | Terminal branching at boundary |
| `E.MapTo[E, A](b)` | Replace Right value |
| `IOE.FromEither` | Lift Either into IOEither |
| `IOE.Chain` | Sequence IOEither operations |
| `IOE.WithResource` | Acquire/use/release pattern |
| `IOE.TryCatchError(func() (A, error))` | Wrap fallible effect |
| `A.Chain` | FlatMap over arrays |
| `A.Flatten` | Flatten `[][]string` to `[]string` |
| `R.FromEntries` | Build Record from key-value pairs |
| `R.Lookup` | Lookup value in Record |
| `P.MakePair` | Create Pair for Record entries |
| `F.Void` | Empty struct for side-effect-only results |

### Import aliases

| Alias | Package |
|---|---|
| `E` | `either` |
| `IOE` | `ioeither` |
| `IOEF` | `ioeither/file` |
| `A` | `array` |
| `F` | `function` |
| `O` | `option` |
| `P` | `pair` |
| `R` | `record` |
| `S` / `Str` | `string` |

### Action pipeline shape

```go
func Action(ctx context.Context, cmd *cli.Command) error {
    return F.Pipe5(
        InputFromCommand(ctx, cmd),
        ValidateInput,
        IOE.FromEither[error],
        IOE.Chain(Execute),
        FP.ToEither[error, F.Void],
        E.Fold(
            F.Identity[error],
            func(F.Void) error { return nil },
        ),
    )
}
```


## Testing Conventions

- Tests live alongside source in `internal/containerbuild/`
- Pure functions tested directly — no mocking needed
- Run tests after every modification: `gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...`


## Gotchas

- **`container` binary** must be in PATH at runtime
- **Build context is always `.`** — never exposed to users
- **`urfave/cli/v3`**: use `cmd.String("flag")`, `cmd.StringSlice("flag")`, `cmd.Bool("flag")` inside Action
- **fp-go generics**: type parameters often required explicitly (e.g., `E.MapTo[error, string](true)`)
- **`IOEither` is lazy**: must call `program()` to trigger execution
- **Use `E.Fold` for terminal branching**, not `if` statements in application code
- **`command-reference.md`** is local source of truth for `container build` flags

# Implementation Plan: Dynamic Tool Versioning for `container-cli`

## Executive Summary

This document outlines the complete implementation plan for extending `container-cli build` to support dynamic versioning of tools installed in the Docker image. The implementation follows strict functional programming principles using `github.com/IBM/fp-go/v2`.

---

## 1. Architecture Overview

### 1.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLI Flags                                       │
│        --golangci-lint-version    --crush-version    --gotestsum-version    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         InputFromCommand()                                   │
│  Extracts flags → constructs BuildArgs map[string]string                    │
│  Maps to: GOLANGCI_LINT_VERSION, CRUSH_VERSION, GOTESTSUM_VERSION          │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ValidateInput()                                    │
│  Pure Either validation (no mutations, no side effects)                     │
│  No validation for build args — they are passed through as-is               │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          RenderCommand()                                     │
│  Pure transformation: BuildArgs → []string{"--build-arg", "KEY=VALUE"}     │
│  Uses A.Chain for functional composition                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            Execute()                                         │
│  IOEither-based process execution                                           │
│  Passes resolved argv to `container build`                                  │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          Dockerfile                                          │
│  ARG GOLANGCI_LINT_VERSION=2.11.4                                           │
│  ARG CRUSH_VERSION=latest                                                   │
│  ARG GOTESTSUM_VERSION=latest                                               │
│  Uses ${VAR} in RUN instructions                                            │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Type Signature Changes

```go
// BEFORE
type Input struct {
    DockerfileSource IOE.IOEither[error, DockerfileResource]
    Name             string
    Tags             []string
    Ctx              context.Context
}

// AFTER
type Input struct {
    DockerfileSource IOE.IOEither[error, DockerfileResource]
    Name             string
    Tags             []string
    BuildArgs        map[string]string  // NEW: tool version overrides
    Ctx              context.Context
}
```

---

## 2. Implementation Details

### 2.1 Domain Model (`internal/containerbuild/input.go`)

**Changes Required:**
- Add `BuildArgs map[string]string` field to `Input` struct

**Implementation:**

```go
// Input holds the build parameters throughout the pipeline.
type Input struct {
    DockerfileSource IOE.IOEither[error, DockerfileResource]
    Name             string
    Tags             []string
    BuildArgs        map[string]string // Tool version overrides for Docker build
    Ctx              context.Context
}
```

**Design Rationale:**
- Using `map[string]string` provides O(1) lookup and natural iteration
- Named version flags map directly to canonical Dockerfile ARG names
- Simple and explicit — no parsing required

---

### 2.2 CLI Definition (`internal/containerbuild/command.go`)

**Changes Required:**
- Add three new string flags to the `build` command
- Update `InputFromCommand` to extract versions and construct `BuildArgs`

**Flag Specifications:**

| Flag | Type | Default | ARG Name |
|------|------|---------|----------|
| `--golangci-lint-version` | string | "2.11.4" | `GOLANGCI_LINT_VERSION` |
| `--crush-version` | string | "latest" | `CRUSH_VERSION` |
| `--gotestsum-version` | string | "latest" | `GOTESTSUM_VERSION` |

**Implementation:**

```go
func Command() *cli.Command {
    return &cli.Command{
        Name:  "build",
        Usage: "Build an OCI image via the container CLI",
        Flags: []cli.Flag{
            // Existing flags...
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
                Value:   "crusher",
            },
            &cli.BoolFlag{
                Name:  "embed",
                Usage: "Use the Dockerfile embedded in the binary (ignores --file)",
            },
            // NEW: Tool version flags
            &cli.StringFlag{
                Name:  "golangci-lint-version",
                Usage: "golangci-lint version",
                Value: "2.11.4",
            },
            &cli.StringFlag{
                Name:  "crush-version",
                Usage: "crush version",
                Value: "latest",
            },
            &cli.StringFlag{
                Name:  "gotestsum-version",
                Usage: "gotestsum version",
                Value: "latest",
            },
        },
        Action: Action,
    }
}
```

**`InputFromCommand` Implementation:**

```go
// InputFromCommand reads CLI flags and constructs the Input.
// BuildArgs are constructed directly from version flags — no parsing needed.
func InputFromCommand(ctx context.Context, cmd *cli.Command) Input {
    return Input{
        DockerfileSource: resolverFactories[cmd.Bool("embed")](cmd),
        Name:             cmd.String("name"),
        Tags:             cmd.StringSlice("tag"),
        BuildArgs: map[string]string{
            "GOLANGCI_LINT_VERSION": cmd.String("golangci-lint-version"),
            "CRUSH_VERSION":         cmd.String("crush-version"),
            "GOTESTSUM_VERSION":     cmd.String("gotestsum-version"),
        },
        Ctx: ctx,
    }
}
```

**Design Rationale:**
- Direct mapping from flags to map — simple and explicit
- No parsing logic required (unlike generic `--build-arg`)
- Default values are defined once in flag declarations
- Type-safe: all three are string flags

---

### 2.3 Argument Rendering (`internal/containerbuild/args.go`)

**Changes Required:**
- Add a function to render BuildArgs into `--build-arg` flags
- Integrate into `RenderCommand` function

**Implementation:**

```go
import (
    "fmt"
    "maps"
    "slices"

    A "github.com/IBM/fp-go/v2/array"
    F "github.com/IBM/fp-go/v2/function"
)

// renderBuildArgArgs converts the BuildArgs map into an array of
// "--build-arg" "KEY=VALUE" pairs using functional composition.
// Keys are sorted for deterministic output.
func renderBuildArgArgs(buildArgs map[string]string) []string {
    return F.Pipe2(
        maps.Keys(buildArgs),
        slices.Sort,
        A.Chain(func(key string) []string {
            return []string{"--build-arg", fmt.Sprintf("%s=%s", key, buildArgs[key])}
        }),
    )
}

// RenderCommand is a pure function that builds a CommandSpec from an Input
// and a resolved Dockerfile path.
func RenderCommand(r Input, path string) CommandSpec {
    return CommandSpec{
        Bin: containerBinary,
        Args: A.ArrayConcatAll(
            []string{"build", "--file", path},
            renderTagArgs(r),
            renderBuildArgArgs(r.BuildArgs), // NEW: build args
            []string{"."},
        ),
    }
}
```

**Alternative: Using `A.From` for clarity:**

```go
func renderBuildArgArgs(buildArgs map[string]string) []string {
    return F.Pipe2(
        maps.Keys(buildArgs),
        slices.Sort,
        A.Chain(func(key string) []string {
            return A.From("--build-arg", fmt.Sprintf("%s=%s", key, buildArgs[key]))
        }),
    )
}
```

**Design Rationale:**
- Sorting keys ensures deterministic command output (important for testing and caching)
- `A.Chain` naturally flattens `[][]string` to `[]string`
- Build args are positioned after tags, before the build context `.`
- Pure function — no side effects

---

### 2.4 Dockerfile (`Dockerfile`)

**Changes Required:**
- Add `ARG` declarations with default values
- Replace hardcoded versions with variable references

**Implementation:**

```dockerfile
FROM docker/sandbox-templates:shell

USER root

# Build arguments for tool versions (defaults match CLI defaults)
ARG GOLANGCI_LINT_VERSION=2.11.4
ARG CRUSH_VERSION=latest
ARG GOTESTSUM_VERSION=latest

# Install apt packages: ripgrep, gopls, fd-find, wget
RUN apt-get update && apt-get install -y \
    ripgrep \
    gopls \
    fd-find \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint from official .deb package
# Uses build argument for version
RUN set -eux; \
    GOLANGCI_DEB="golangci-lint-${GOLANGCI_LINT_VERSION}-linux-arm64.deb"; \
    GOLANGCI_URL="https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/${GOLANGCI_DEB}"; \
    wget -q "${GOLANGCI_URL}" -O "/tmp/${GOLANGCI_DEB}"; \
    dpkg -i "/tmp/${GOLANGCI_DEB}"; \
    rm -f "/tmp/${GOLANGCI_DEB}"

# Install Go tools to /usr/local/bin (system-wide)
# Uses build arguments for versions
RUN GOPATH=/tmp/go go install "github.com/charmbracelet/crush@${CRUSH_VERSION}" && \
    GOPATH=/tmp/go go install "gotest.tools/gotestsum@${GOTESTSUM_VERSION}" && \
    mv /tmp/go/bin/* /usr/local/bin/ && \
    rm -rf /tmp/go

USER agent
```

**Version Format Notes:**

- `golangci-lint`: Version number without `v` prefix (e.g., `2.11.4`, `2.12.0`)
- `crush` and `gotestsum`: Accept Go module version syntax (`latest`, `v1.2.3`, commit hash)

---

## 3. Testing Strategy

### 3.1 Unit Tests (`internal/containerbuild/args_test.go`)

**Test Cases:**

```go
func TestRenderCommand_BuildArgs_Defaults(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name: "myapp",
        Tags: []string{"latest"},
        BuildArgs: map[string]string{
            "GOLANGCI_LINT_VERSION": "2.11.4",
            "CRUSH_VERSION":         "latest",
            "GOTESTSUM_VERSION":     "latest",
        },
    }

    spec := RenderCommand(req, "Dockerfile")

    // Verify all three build args appear
    require.Contains(spec.Args, "GOLANGCI_LINT_VERSION=2.11.4")
    require.Contains(spec.Args, "CRUSH_VERSION=latest")
    require.Contains(spec.Args, "GOTESTSUM_VERSION=latest")
}

func TestRenderCommand_BuildArgs_CustomVersions(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name: "myapp",
        Tags: []string{"latest"},
        BuildArgs: map[string]string{
            "GOLANGCI_LINT_VERSION": "2.12.0",
            "CRUSH_VERSION":         "v1.0.0",
            "GOTESTSUM_VERSION":     "v1.10.0",
        },
    }

    spec := RenderCommand(req, "Dockerfile")

    require.Contains(spec.Args, "GOLANGCI_LINT_VERSION=2.12.0")
    require.Contains(spec.Args, "CRUSH_VERSION=v1.0.0")
    require.Contains(spec.Args, "GOTESTSUM_VERSION=v1.10.0")
}

func TestRenderCommand_BuildArgs_SortedOrder(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name: "myapp",
        Tags: []string{"latest"},
        BuildArgs: map[string]string{
            "GOLANGCI_LINT_VERSION": "2.11.4",
            "CRUSH_VERSION":         "latest",
            "GOTESTSUM_VERSION":     "latest",
        },
    }

    spec := RenderCommand(req, "Dockerfile")

    // Build args should be sorted alphabetically: CRUSH, GOLANGCI, GOTESTSUM
    var buildArgValues []string
    for i, arg := range spec.Args {
        if arg == "--build-arg" && i+1 < len(spec.Args) {
            buildArgValues = append(buildArgValues, spec.Args[i+1])
        }
    }

    require.Len(buildArgValues, 3)
    require.Equal("CRUSH_VERSION=latest", buildArgValues[0])
    require.Equal("GOLANGCI_LINT_VERSION=2.11.4", buildArgValues[1])
    require.Equal("GOTESTSUM_VERSION=latest", buildArgValues[2])
}

func TestRenderCommand_BuildArgs_Position(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name: "myapp",
        Tags: []string{"latest"},
        BuildArgs: map[string]string{
            "GOLANGCI_LINT_VERSION": "2.11.4",
        },
    }

    spec := RenderCommand(req, "Dockerfile")

    // Last element should always be "."
    require.Equal(".", spec.Args[len(spec.Args)-1])

    // Build args should come before "."
    buildArgIdx := -1
    for i, arg := range spec.Args {
        if arg == "GOLANGCI_LINT_VERSION=2.11.4" {
            buildArgIdx = i
            break
        }
    }
    require.Less(buildArgIdx, len(spec.Args)-1)
}

func TestRenderCommand_EmptyBuildArgs(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name:      "myapp",
        Tags:      []string{"latest"},
        BuildArgs: map[string]string{},
    }

    spec := RenderCommand(req, "Dockerfile")

    // No --build-arg should appear for empty map
    require.NotContains(spec.Args, "--build-arg")
}
```

### 3.2 Unit Tests (`internal/containerbuild/validate_test.go`)

**Test Cases for Build Args:**

```go
func TestValidateInput_BuildArgsNotValidated(t *testing.T) {
    // BuildArgs are not validated — they are passed through as-is
    require := require.New(t)
    req := Input{
        Name:      "myapp",
        Tags:      []string{"latest"},
        BuildArgs: map[string]string{"GOLANGCI_LINT_VERSION": "invalid-version"},
    }

    result := ValidateInput(req)
    require.True(E.IsRight(result))
}

func TestValidateInput_NilBuildArgs(t *testing.T) {
    require := require.New(t)
    req := Input{
        Name:      "myapp",
        Tags:      []string{"latest"},
        BuildArgs: nil,
    }

    result := ValidateInput(req)
    require.True(E.IsRight(result))
}
```

### 3.3 Integration Tests (Manual Verification)

```bash
# Test 1: Default versions (no version flags)
go run ./cmd/container-cli/... build -n test-image

# Test 2: Override single tool version
go run ./cmd/container-cli/... build -n test-image --golangci-lint-version 2.12.0

# Test 3: Override two tool versions
go run ./cmd/container-cli/... build -n test-image \
  --golangci-lint-version 2.12.0 \
  --crush-version v1.2.3

# Test 4: Override all tool versions
go run ./cmd/container-cli/... build -n test-image \
  --golangci-lint-version 2.12.0 \
  --crush-version v1.2.3 \
  --gotestsum-version v1.10.0

# Test 5: Verify generated command (dry-run with --help)
go run ./cmd/container-cli/... build --help
```

---

## 4. File Changes Summary

| File | Changes |
|------|---------|
| `internal/containerbuild/input.go` | Add `BuildArgs map[string]string` field to `Input` struct |
| `internal/containerbuild/command.go` | Add 3 new CLI flags, update `InputFromCommand` |
| `internal/containerbuild/args.go` | Add `renderBuildArgArgs` function, update `RenderCommand` |
| `internal/containerbuild/validate.go` | No changes required |
| `internal/containerbuild/args_test.go` | Add test cases for build arg rendering |
| `Dockerfile` | Add `ARG` declarations, use variables in `RUN` instructions |

---

## 5. Implementation Order

1. **Phase 1: Domain Model** (`input.go`)
   - Add `BuildArgs` field to `Input` struct

2. **Phase 2: CLI Layer** (`command.go`)
   - Add three version flags to `Command()`
   - Update `InputFromCommand` to construct `BuildArgs` map

3. **Phase 3: Argument Rendering** (`args.go`)
   - Implement `renderBuildArgArgs`
   - Update `RenderCommand`

4. **Phase 4: Dockerfile** (`Dockerfile`)
   - Add `ARG` declarations
   - Update `RUN` instructions to use variables

5. **Phase 5: Testing** (`args_test.go`, `validate_test.go`)
   - Add unit tests for build arg rendering
   - Run full test suite
   - Run linter

6. **Phase 6: Integration**
   - Build and test the CLI
   - Verify generated `container build` commands

---

## 6. Functional Programming Checklist

- [ ] No `if`/`else` statements in `internal/containerbuild`
- [ ] No `for` loops — use `A.Map`, `A.Chain`, `A.Fold`
- [ ] All validation returns `E.Either[error, T]`
- [ ] Side effects isolated in `IOEither` (process execution)
- [ ] Terminal branching uses `E.Fold`
- [ ] Composition uses `F.Pipe1`, `F.Pipe2`, etc.
- [ ] Array operations use `A.Chain`, `A.Flatten`, `A.Map`
- [ ] All functions are pure except those returning `IOEither`

---

## 7. Verification Commands

```bash
# Build
go build ./cmd/container-cli/...

# Test
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Lint
golangci-lint run ./...

# Format
golangci-lint fmt

# Verify no imperative branching
grep -E '\bif\b|\belse\b|\bfor\b' internal/containerbuild/*.go | grep -v '// '
```

---

## 8. Future Enhancements (Out of Scope)

1. **Version Validation**: Validate version format before building
2. **Version List Command**: Add `container-cli list-versions` to show available versions
3. **Shell Completion**: Add completion for known version patterns

---

## 9. References

- [AGENTS.md](../AGENTS.md) — Project conventions and functional programming guidelines
- [command-reference.md](./command-reference.md) — Apple `container` CLI documentation
- [PLAN-build-cmd.md](./PLAN-build-cmd.md) — Original build command implementation plan
- [fp-go Documentation](https://github.com/IBM/fp-go) — Functional programming library
- [urfave/cli/v3 Documentation](https://github.com/urfave/cli) — CLI framework

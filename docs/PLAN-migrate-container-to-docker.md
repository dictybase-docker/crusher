# Plan: Migrate container CLI to docker CLI

## Context

The crusher binary wraps Apple's `container` CLI to build and manage OCI containers. The goal is to migrate every invocation to use `docker` instead, so the tool works on systems where Docker is installed rather than (or in addition to) Apple's container CLI.

The two CLI tools share nearly identical flag interfaces for the subcommands in use (`build`, `create`, `start`, `exec`), so the migration is a targeted string-swap plus cosmetic string updates.

---

## Changes

### 1. `internal/containerbuild/input.go` — line 9
```go
// before
const containerBinary = "container"
// after
const containerBinary = "docker"
```

### 2. `internal/containercreate/args.go` — line 8
```go
// before
const containerBinary = "container"
// after
const containerBinary = "docker"
```

### 3. `internal/containercreate/command.go` — line 127 (hint message in `printResult`)
```go
// before
nord10.Printf("container exec -it %s /bin/sh", r.Name)
// after
nord10.Printf("docker exec -it %s /bin/sh", r.Name)
```

### 4. `cmd/crusher/main.go` — cosmetic app metadata
- `Name: "container-cli"` → `Name: "docker-cli"` (or just `"crusher"`)
- `Usage: "Build OCI images through the container CLI"` → `"Build OCI images through the docker CLI"`

### 5. `internal/containerbuild/command.go` — line 72 (usage string)
```
Usage: "Build an OCI image via the container CLI"
→
Usage: "Build an OCI image via the docker CLI"
```

---

## Flag Compatibility (no arg changes needed)

| container CLI flag | docker CLI flag | Status |
|--------------------|----------------|--------|
| `build --file` | `build --file` | ✓ identical |
| `build --tag` | `build --tag` | ✓ identical |
| `build --build-arg` | `build --build-arg` | ✓ identical |
| `create --name` | `create --name` | ✓ identical |
| `create --mount type=bind,...` | `create --mount type=bind,...` | ✓ identical |
| `create --env` | `create --env` | ✓ identical |
| `create --workdir` | `create --workdir` | ✓ identical |
| `create --dns` | `create --dns` | ✓ identical |
| `create --interactive --tty` | `create --interactive --tty` | ✓ identical |
| `start <name>` | `start <name>` | ✓ identical |

No changes to `args.go` argument construction beyond the binary name constant.

---

## Critical Files

- `internal/containerbuild/input.go` — `containerBinary` constant
- `internal/containercreate/args.go` — `containerBinary` constant
- `internal/containercreate/command.go` — hint message
- `cmd/crusher/main.go` — app name/usage strings
- `internal/containerbuild/command.go` — usage string

---

## Verification

```bash
# Build
go build ./...

# Tests
gotestsum --format-hide-empty-pkg --format dots

# Smoke check (requires docker installed)
go run ./cmd/crusher/ build --embed --name crusher
```

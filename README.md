# crusher

[![Go Reference](https://pkg.go.dev/badge/github.com/cybersiddhu/crush-sandbox.svg)](https://pkg.go.dev/github.com/cybersiddhu/crush-sandbox)
[![Go Report Card](https://goreportcard.com/badge/github.com/dictybase-docker/crusher)](https://goreportcard.com/report/github.com/dictybase-docker/crusher)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

CLI tool for building OCI images, creating Crush sandbox containers, and packing Docker Sandbox agent kits — all powered by functional programming combinators.

## Contents

- [Prerequisites](#prerequisites)
- [Install](#install)
- [Commands](#commands)
  - [build](#build)
  - [create](#create)
  - [sbx](#sbx)
- [Project Structure](#project-structure)
- [Development](#development)

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Docker](https://docs.docker.com/) — OCI image builds and container management
- [sbx](https://docker.com/sandbox) — Docker Sandbox CLI (for `sbx` subcommand)

## Install

```bash
go install github.com/cybersiddhu/crush-sandbox/cmd/crusher@latest
```

Or build from source:

```bash
git clone https://github.com/cybersiddhu/crush-sandbox.git
cd crush-sandbox
go build -o crusher ./cmd/crusher
```

## Commands

### build

Build an OCI image using the Docker CLI. Supports embeddable Dockerfile, custom build args for tool versions, and multiple tags.

```
crusher build [--name NAME] [--tag TAG] [--file PATH] [--embed] [--golangci-lint-version VER] [--crush-version VER] [--gotestsum-version VER] [--moxide-version VER] [--sem-version VER] [--rtk-version VER]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--name` / `-n` | `crusher` | Image name (combines with tags as `name:tag`) |
| `--tag` / `-t` | `[latest]` | Image tag, repeatable |
| `--file` / `-f` | `Dockerfile` | Path to Dockerfile |
| `--embed` | `false` | Use the Dockerfile embedded in the binary (ignores `--file`) |
| `--golangci-lint-version` | `2.11.4` | golangci-lint version to install |
| `--crush-version` | `latest` | Crush version to install |
| `--gotestsum-version` | `latest` | gotestsum version to install |
| `--moxide-version` | `latest` | markdown-oxide version to install |
| `--sem-version` | `latest` | sem version to install |
| `--rtk-version` | `latest` | rtk version to install |

Tool versions are passed as `--build-arg` to Docker. Build context is always the current directory (`.`).

**Example:**

```bash
# Build with defaults
crusher build

# Build with custom image name and tag
crusher build --name my-image --tag v1.0

# Build using embedded Dockerfile with custom tool versions
crusher build --embed --crush-version 1.2.3 --go-version 1.24.0
```

### create

Create and start a Crush sandbox container with pre-configured volume mounts for config, data, skills, and workspace directories.

```
crusher create --config PATH --data PATH --skills PATH --api-key KEY [--name NAME] [--image NAME] [--workspace PATH] [--github-token TOKEN] [--volume PATH]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--config` / `-c` | *(required)* | Host path to Crush config directory |
| `--data` / `-d` | *(required)* | Host path to Crush data directory |
| `--skills` / `-s` | *(required)* | Host path to Crush skills directory |
| `--api-key` / `-k` | *(required)* | API key for Crush |
| `--name` / `-n` | *(auto-generated)* | Container name |
| `--image` / `-i` | `crusher:latest` | Image name |
| `--workspace` / `-w` | *(current dir)* | Host path to workspace directory |
| `--github-token` / `-g` | *(none)* | GitHub personal access token |
| `--volume` / `-v` | *(none)* | Additional host path to mount (read-only, repeatable) |

Required mounts are mapped under `/home/agent/crush/` inside the container. Extra volumes are mounted read-only under `/home/agent/mount/`.

**Example:**

```bash
# Create a Crush container with directory mounts
crusher create \
  --config ~/.crush/config \
  --data ~/.crush/data \
  --skills ~/.crush/skills \
  --api-key "sk-..." \
  --workspace ~/my-project

# Create with additional read-only volume and GitHub token
crusher create \
  --config ~/.crush/config \
  --data ~/.crush/data \
  --skills ~/.crush/skills \
  --api-key "sk-..." \
  --volume /data/reference \
  --github-token "ghp_..."
```

### sbx

Generate, validate, and pack a Docker Sandbox agent kit. Reads an optional `crush.json` config, renders a `spec.yaml`, validates with the `sbx` CLI, stores secrets, packs into a zip, and optionally creates the sandbox instance.

```
crusher sbx --api-key KEY [--output PATH] [--config PATH] [--skills PATH] [--name NAME] [--create] [--crush-version VER] [--golangci-lint-version VER] [--go-version VER] [--gotestsum-version VER] [--moxide-version VER] [--sem-version VER] [--rtk-version VER]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` / `-k` | *(required)* | OpenRouter API key |
| `--output` / `-o` | `crush-sbx-kit.zip` | Path for the packed kit zip file |
| `--config` / `-c` | *(default config)* | Path to `crush.json` (default OpenRouter config if omitted) |
| `--skills` / `-s` | *(none)* | Path to skills directory |
| `--name` / `-n` | *(current dir basename)* | Sandbox display name |
| `--create` | `false` | Create the sandbox instance after packing |
| `--crush-version` | `latest` | Crush version for `go install` |
| `--golangci-lint-version` | `2.11.4` | golangci-lint version |
| `--go-version` | `1.25.4` | Go toolchain version |
| `--gotestsum-version` | `latest` | gotestsum version |
| `--moxide-version` | `latest` | markdown-oxide version |
| `--sem-version` | `latest` | sem version |
| `--rtk-version` | `latest` | rtk version |

The pipeline: generate spec → validate kit → store secret → pack kit → (optionally create sandbox) → cleanup temp dir.

**Example:**

```bash
# Pack a kit with default config
crusher sbx --api-key "sk-or-..."

# Pack with custom config and skills, then create the sandbox
crusher sbx \
  --api-key "sk-or-..." \
  --config ~/.crush/config/crush.json \
  --skills ~/.crush/skills \
  --name my-sandbox \
  --create \
  --output ./my-kit.zip
```

## Project Structure

```
.
├── cmd/
│   └── crusher/
│       └── main.go              # CLI entry point, subcommand registration
├── internal/
│   ├── containerbuild/
│   │   ├── command.go           # build subcommand: CLI flags, InputFromCommand, Action
│   │   ├── input.go             # Input, CommandSpec, DockerfileResource types
│   │   ├── validate.go          # Pure Either-based input validation
│   │   ├── args.go              # Pure argv rendering for container build
│   │   ├── exec.go              # IOEither-based process execution with resource cleanup
│   │   ├── resource.go          # FileResolver, EmbeddedResolver for Dockerfile
│   │   ├── embed.go             # //go:embed Dockerfile
│   │   └── Dockerfile           # Embedded default Dockerfile
│   ├── containercreate/
│   │   ├── command.go           # create subcommand: CLI flags, InputFromCommand, Action
│   │   ├── input.go             # Input, ResolvedInput, ContainerResult, MountSpec types
│   │   ├── validate.go          # Path validation and container name generation
│   │   ├── args.go              # Pure argv rendering for container create
│   │   ├── exec.go              # Process execution with container start
│   │   ├── mounts.go            # Volume mount resolution
│   │   └── namegen.go           # Container name auto-generation
│   ├── containersbx/
│   │   ├── command.go           # sbx subcommand: CLI flags, InputFromCommand, Action
│   │   ├── input.go             # Input, ResolvedInput, KitResult, execState types
│   │   ├── validate.go          # Input normalization and validation
│   │   ├── specgen.go           # spec.yaml template rendering
│   │   ├── config.go            # crush.json reader with OpenRouter default
│   │   ├── exec.go              # Multi-stage pipeline: generate → validate → pack → create
│   │   ├── embed.go             # //go:embed spec.yaml.tmpl
│   │   └── spec.yaml.tmpl       # Docker Sandbox agent spec template
│   └── fp/
│       ├── conversion.go        # ToEither helper for IOEither → Either conversion
│       └── conversion_test.go
```

### Packages

| Package | Responsibility |
|---------|---------------|
| `internal/containerbuild` | OCI image builds — Dockerfile resolution, validation, `docker build` execution |
| `internal/containercreate` | Container lifecycle — volume mount resolution, `docker create` + `docker start` |
| `internal/containersbx` | Sandbox agent kit — `spec.yaml` generation, `sbx` CLI orchestration (validate, pack, create) |
| `internal/fp` | Shared functional utilities — `ToEither` conversion combinator |

All packages are built with [fp-go](https://github.com/IBM/fp-go) functional programming combinators and use [urfave/cli](https://github.com/urfave/cli) for the CLI framework.

## Development

```bash
# Run tests
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./...

# Run tests with verbose output
gotestsum --format testdox --format-hide-empty-pkg -- ./...

# Lint
golangci-lint run ./...

# Format
golangci-lint fmt

# Build
go build -o crusher ./cmd/crusher
```
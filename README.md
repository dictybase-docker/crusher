# crusher

[![Go Reference](https://pkg.go.dev/badge/github.com/dictybase-docker/crusher.svg)](https://pkg.go.dev/github.com/dictybase-docker/crusher)
[![Go Report Card](https://goreportcard.com/badge/github.com/dictybase-docker/crusher)](https://goreportcard.com/report/github.com/dictybase-docker/crusher)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

CLI tool for building OCI images, creating Crush sandbox containers, and packing Docker Sandbox agent kits for Crush and Opencode — all powered by functional programming combinators.

## Contents

- [Prerequisites](#prerequisites)
- [Install](#install)
- [Commands](#commands)
  - [build](#build)
  - [create](#create)
  - [sbx](#sbx)
  - [opencode-sbx](#opencode-sbx)
- [Developing Kit Extensions](#developing-kit-extensions)
- [Project Structure](#project-structure)
- [Development](#development)

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Docker](https://docs.docker.com/) — OCI image builds and container management
- [sbx](https://docker.com/sandbox) — Docker Sandbox CLI (for `sbx` subcommand)

## Install

```bash
go install github.com/dictybase-docker/crusher/cmd/crusher@latest
```

Or build from source:

```bash
git clone https://github.com/dictybase-docker/crusher.git
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

Generate, validate, and pack a Docker Sandbox agent kit for Crush. Reads an optional `crush.json` config, renders a `spec.yaml`, validates with the `sbx` CLI, stores secrets, packs into a zip, and optionally creates the sandbox instance.

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

### opencode-sbx

Generate, validate, and pack a [Docker Sandbox](https://docs.docker.com/ai/sandboxes/) agent kit for [Opencode](https://opencode.ai). Unlike the Crush kit, all tooling is pre-baked into the base image — the kit is configuration-only with embedded global skills, agents, commands, and plugins. Supports multiple AI providers.

```
crusher opencode-sbx --api-key KEY [--output PATH] [--name NAME] [--provider NAME] [--image IMAGE] [--create] [--golangci-lint-version VER]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--api-key` / `-k` | *(required)* | AI provider API key |
| `--output` / `-o` | `opencode-sbx-kit.zip` | Path for the packed kit zip file |
| `--name` / `-n` | *(auto-generated)* | Sandbox display name |
| `--provider` / `-p` | `openrouter` | AI provider: `openrouter`, `anthropic`, `openai`, `google` |
| `--image` | `ghcr.io/dictybase/crusher:opencode-sbx-base` | Base Docker image for the sandbox agent |
| `--create` | `false` | Create the sandbox instance after packing |
| `--golangci-lint-version` | `2.11.4` | golangci-lint version surfaced in agent context |

The pipeline: generate spec → write spec.yaml and global files → validate kit → pack kit → (optionally store API key secret + create sandbox).

**Example:**

```bash
# Pack a kit with OpenRouter as the provider
crusher opencode-sbx --api-key "sk-or-..."

# Pack with Anthropic, create the sandbox, and use a custom image
crusher opencode-sbx \
  --api-key "sk-ant-..." \
  --provider anthropic \
  --name my-opencode \
  --create \
  --output ./my-opencode-kit.zip
```

#### Next Steps: Running the Sandbox

After the kit is packed, use the `sbx` CLI to run it. Make sure you are signed in and have your secrets stored:

```bash
# Sign in to Docker (required once)
sbx login

# Store your API key as a global secret (once per provider)
sbx secret set -g openrouter
# or: sbx secret set -g anthropic
# or: sbx secret set -g openai
# or: sbx secret set -g google

# Run OpenCode with your kit in the current workspace
sbx run opencode --kit ./opencode-sbx-kit.zip .

# Or run with a named sandbox for easier management
sbx run --name my-agent opencode --kit ./opencode-sbx-kit.zip .
```

Common follow-up commands:

```bash
# List all sandboxes and their status
sbx ls

# Get a shell inside the sandbox
sbx exec -it my-agent bash

# Stop the sandbox (state preserved, restartable)
sbx stop my-agent

# Restart and re-attach to an existing sandbox
sbx run --name my-agent

# Remove a sandbox entirely
sbx rm my-agent
```

See the [Docker Sandbox CLI reference](https://docs.docker.com/reference/cli/sbx/) for all available commands.

## Developing Kit Extensions

The `opencode-sbx` kit ships with embedded skills, agents, commands, and plugins that are mirrored into the sandbox at `~/.config/opencode/`. All extensions live under `internal/containeropencodebx/global/` and are compiled into the binary via `//go:embed`.

### Directory Layout

```
internal/containeropencodebx/global/
├── skills/              # OpenCode skills (one subdirectory per skill)
│   └── <name>/SKILL.md
├── agents/              # Subagent definitions
│   └── <name>.md
├── commands/            # Custom slash commands
│   └── <name>.md
├── plugins/             # TypeScript/JavaScript plugins loaded by Bun
│   └── <name>.ts | <name>.js
├── package.json         # Dependency manifest for plugins
├── bun.lock             # Bun lockfile (npm dependencies)
└── bun.lockb            # Bun binary lockfile
```

Files are mirrored to the sandbox with the `global/` prefix stripped:

| Source | Destination (in sandbox) |
|---|---|
| `global/skills/git-commit/SKILL.md` | `files/home/.config/opencode/skills/git-commit/SKILL.md` |
| `global/agents/reviewer.md` | `files/home/.config/opencode/agents/reviewer.md` |
| `global/commands/commit.md` | `files/home/.config/opencode/commands/commit.md` |
| `global/plugins/sem-tools.ts` | `files/home/.config/opencode/plugins/sem-tools.ts` |
| `global/package.json` | `files/home/.config/opencode/package.json` |

### How It Works

**Embedding**: `internal/containeropencodebx/embed.go` uses `//go:embed all:global` to compile the entire `global/` tree into the binary as an [`embed.FS`](https://pkg.go.dev/embed). No external files are needed at runtime.

**Pattern Matching**: `globalPatterns` in `exec.go` declares which file patterns to collect. The pipeline runs `fs.Glob` against the embedded filesystem, flattens the results, and writes each file to a temp directory at kit generation time.

```go
var globalPatterns = []string{
    "global/skills/*/SKILL.md",
    "global/agents/*.md",
    "global/commands/*.md",
    "global/plugins/*.ts",
    "global/plugins/*.js",
    "global/package.json",
    "global/bun.lock",
    "global/bun.lockb",
}
```

The `toOpencodeRoot` chain produces the destination prefix `files/home/.config/opencode/`, and `trimGlobalPrefix` strips the leading `global/` segment.

### Adding a Skill

Skills are defined as Markdown files with YAML frontmatter. OpenCode loads them from `~/.config/opencode/skills/<name>/SKILL.md`.

**Steps:**

1. Create the skill directory and file:
   ```bash
   mkdir -p internal/containeropencodebx/global/skills/my-skill
   cat > internal/containeropencodebx/global/skills/my-skill/SKILL.md << 'EOF'
   ---
   name: my-skill
   description: What this skill does and when to trigger it
   ---

   # my-skill

   Instructions for the agent...
   EOF
   ```

2. Run tests to confirm the glob picks it up:
   ```bash
   go test ./internal/containeropencodebx/ -run TestGlobalFilePaths
   ```

Skills are picked up automatically by the existing `global/skills/*/SKILL.md` glob pattern — no code changes needed.

### Adding an Agent

Agents are subagent definitions in Markdown format with YAML frontmatter.

**Steps:**

1. Create the agent file:
   ```bash
   cat > internal/containeropencodebx/global/agents/my-agent.md << 'EOF'
   ---
   description: Specialized agent for code reviews
   mode: subagent
   permission:
     edit: deny
   ---

   Review the provided code for correctness, security, and style.
   EOF
   ```

Agents are matched by the existing `global/agents/*.md` glob. No code changes needed.

### Adding a Command

Commands are Markdown files that define slash commands in OpenCode.

**Steps:**

1. Create the command file:
   ```bash
   cat > internal/containeropencodebx/global/commands/lint.md << 'EOF'
   ---
   description: Run linters and auto-fix issues
   ---

   Run golangci-lint and apply all auto-fixes:

   ```bash
   golangci-lint run --fix ./...
   ```
   EOF
   ```

Commands are matched by the existing `global/commands/*.md` glob. No code changes needed.

### Adding a Plugin

Plugins are TypeScript or JavaScript modules loaded by [Bun](https://bun.sh) at OpenCode startup. Use the `opencode` global to register tools.

**Steps:**

1. Create the plugin file:
   ```bash
   cat > internal/containeropencodebx/global/plugins/my-tool.ts << 'EOF'
   // Register a custom tool accessible by the agent
   export default {
     name: "my-tool",
     setup(opencode) {
       opencode.tool.register("greet", {
         description: "Greet the user by name",
         parameters: {
           name: { type: "string", description: "Name to greet" },
         },
         async handler({ name }) {
           return `Hello, ${name}!`;
         },
       });
     },
   };
   EOF
   ```

2. If the plugin has npm dependencies, add them to `global/package.json` and regenerate the lockfiles:
   ```bash
   cd internal/containeropencodebx/global/
   npm install some-package     # or: bun add some-package
   # bun.lock / bun.lockb will be updated automatically
   ```

3. Run tests to verify:
   ```bash
   go test ./internal/containeropencodebx/ -run TestWriteGlobalFiles
   ```

### Testing Extensions

The existing test suite verifies that embedded files are discoverable and mirrored correctly:

| Test | What It Verifies |
|---|---|
| `TestGlobalFilePaths_ContainsExpectedSkill` | Skills are found by glob |
| `TestWriteGlobalFiles_AllFilesPresent` | All files are mirrored to the kit output |
| `TestWriteGlobalFiles_MirrorsPackageJSON` | `package.json` is included |
| `TestWriteOneGlobalFile_CreatesFileAtCorrectPath` | A single file lands at the right destination |

Run the full suite after adding extensions:

```bash
gotestsum --format pkgname-and-test-fails --format-hide-empty-pkg -- ./internal/containeropencodebx/...
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
│   ├── containeropencodebx/
│   │   ├── command.go           # opencode-sbx subcommand: CLI flags, InputFromCommand, Action
│   │   ├── input.go             # Input, KitResult, execState types and provider configs
│   │   ├── validate.go          # Input normalization and validation
│   │   ├── specgen.go           # spec.yaml template rendering
│   │   ├── exec.go              # Pipeline: generate temp dir → validate → pack → create
│   │   ├── embed.go             # //go:embed spec.yaml.tmpl + global/ tree
│   │   ├── spec.yaml.tmpl       # Docker Sandbox opencode agent spec template
│   │   └── global/              # Embedded skills, agents, commands, and plugins
│   ├── sbxexec/
│   │   ├── exec.go              # Provider-agnostic sbx subprocess runner
│   │   └── types.go             # Shared CommandSpec type
│   └── fp/
│       ├── conversion.go        # ToEither helper for IOEither → Either conversion
│       └── conversion_test.go
```

### Packages

| Package | Responsibility |
|---------|---------------|
| `internal/containerbuild` | OCI image builds — Dockerfile resolution, validation, `docker build` execution |
| `internal/containercreate` | Container lifecycle — volume mount resolution, `docker create` + `docker start` |
| `internal/containersbx` | Crush sandbox agent kit — `spec.yaml` generation, `sbx` CLI orchestration (validate, pack, create) |
| `internal/containeropencodebx` | Opencode sandbox agent kit — configuration-only kit with embedded globals (skills, agents, commands, plugins) |
| `internal/sbxexec` | Shared `sbx` subprocess runner — binary lookup, stdin piping, context-driven cancellation |
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
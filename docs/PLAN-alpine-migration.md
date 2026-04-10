# Implementation Plan: Convert Dockerfile to Alpine

## Context

The current Dockerfile at `internal/containerbuild/Dockerfile` uses Debian-based images (`rust:1-slim-bookworm` for the builder stage and `docker/sandbox-templates:shell` for the runtime stage). This plan converts both stages to Alpine-based images to achieve:

- **Smaller image size**: Alpine images are ~30-50% smaller than Debian equivalents
- **Faster builds**: Less data to download
- **Better security posture**: Smaller attack surface
- **Industry standard**: Alpine is the de facto minimal container base

The Dockerfile is embedded in the Go binary via `internal/containerbuild/embed.go` and used by the `container-cli` command with either `--embed` flag (embedded) or default (external file).

## Implementation Approach

### Base Image Changes

**Builder Stage**
- Current: `rust:1-slim-bookworm`
- New: `rust:1-alpine3.21`
- Rationale: Official Rust image with Alpine support, includes musl toolchain, smaller size (~50MB vs ~200MB)

**Runtime Stage**
- Current: `docker/sandbox-templates:shell` (undocumented Debian-based image)
- New: `golang:1.25-alpine3.21`
- Rationale: Matches project's Go version (1.25.7 from go.mod), provides Go toolchain for `go install` commands, actively maintained

### Package Manager Migration

**Builder Stage (Debian → Alpine)**
```dockerfile
# Before
RUN apt-get update && apt-get install -y --no-install-recommends \
    git pkg-config libssl-dev cmake \
    && rm -rf /var/lib/apt/lists/*

# After
RUN apk add --no-cache git pkgconfig openssl-dev cmake musl-dev
```

**Runtime Stage (Debian → Alpine)**
```dockerfile
# Before
RUN apt-get update && apt-get install -y \
    ripgrep gopls fd-find wget ca-certificates libssl3 \
    && rm -rf /var/lib/apt/lists/*

# After
RUN apk add --no-cache ripgrep fd wget ca-certificates openssl
```

**Package Mappings:**
- `pkg-config` → `pkgconfig`
- `libssl-dev` → `openssl-dev` (builder)
- `libssl3` → `openssl` (runtime)
- `fd-find` → `fd`
- Add `musl-dev` (required for Rust builds on Alpine)
- `gopls` - removed from apt install (not in Alpine repos), moved to `go install`

### User Creation

The runtime image needs an `agent` user (referenced by `USER agent` at line 106):

```dockerfile
# Add after FROM golang:1.25-alpine3.21
RUN addgroup -g 1000 agent && \
    adduser -D -u 1000 -G agent agent
```

### golangci-lint Installation Change

**Current approach**: Downloads `.deb` package for Debian
**Alpine approach**: Download tar.gz and extract binary

```dockerfile
RUN set -eux; \
    case "$(uname -m)" in \
        x86_64) ARCH="amd64" ;; \
        aarch64|arm64) ARCH="arm64" ;; \
        *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;; \
    esac; \
    GOLANGCI_TAR="golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz"; \
    GOLANGCI_URL="https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/${GOLANGCI_TAR}"; \
    wget -q "${GOLANGCI_URL}" -O "/tmp/${GOLANGCI_TAR}"; \
    tar -xzf "/tmp/${GOLANGCI_TAR}" -C /tmp; \
    install -m 0755 "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}/golangci-lint" /usr/local/bin/golangci-lint; \
    rm -rf "/tmp/${GOLANGCI_TAR}" "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}"; \
    golangci-lint --version
```

### gopls Installation

Move from apt package to `go install`:

```dockerfile
# Add to the go install RUN block
GOPATH=/tmp/go go install "golang.org/x/tools/gopls@latest" && \
```

### fd-find Compatibility Symlink

Alpine's package is named `fd` (binary at `/usr/bin/fd`), but the codebase may expect `fd-find`:

```dockerfile
RUN ln -s /usr/bin/fd /usr/local/bin/fd-find
```

### ARG Variables

Keep all existing ARG declarations unchanged:
- `SEM_VERSION=latest`
- `GOLANGCI_LINT_VERSION=2.11.4`
- `CRUSH_VERSION=latest`
- `GOTESTSUM_VERSION=latest`
- `MOXIDE_VERSION=latest`

### Files to Modify

1. **`internal/containerbuild/Dockerfile`** (primary change)
   - Update both FROM statements
   - Replace apt-get with apk
   - Update package names
   - Add user creation
   - Replace golangci-lint .deb installation with tar.gz
   - Add gopls to go install block
   - Add fd-find symlink

2. **`docs/PLAN-dynamic-versions.md`** (optional documentation update)
   - Add note about Alpine migration in relevant sections

## Verification Strategy

### Build Verification

```bash
# Test builder stage in isolation
go run ./cmd/container-cli/main.go build --target sem-builder -n test-sem -t builder

# Test full build with embedded Dockerfile
go run ./cmd/container-cli/main.go build --embed -n crusher-alpine -t test

# Test full build with external Dockerfile
go run ./cmd/container-cli/main.go build -n crusher-alpine -t test
```

### Runtime Verification

```bash
# Verify all tools are installed and working
container run --rm crusher-alpine:test sh -c "
    ripgrep --version && \
    fd --version && \
    fd-find --version && \
    golangci-lint --version && \
    gopls version && \
    crush --version && \
    gotestsum --version && \
    markdown-oxide --version && \
    sem --version && \
    sem-mcp --version
"

# Verify user
container run --rm crusher-alpine:test id
# Expected: uid=1000(agent) gid=1000(agent)

# Verify SSL/TLS works
container run --rm crusher-alpine:test wget -O- https://api.github.com/
```

### Test Suite Verification

```bash
# Run existing tests (they shouldn't fail - tests don't reference base images)
gotestsum --format-hide-empty-pkg --format dots ./internal/containerbuild/...
```

### Size Comparison

```bash
# Compare image sizes
container images | grep crusher
```

Expected: Alpine image 30-50% smaller than the original Debian-based image.

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| musl vs glibc compatibility | Rust's musl support is production-ready; test all binaries after build |
| markdown-oxide may use glibc binary | Test thoroughly; if fails, switch to musl target in `install_github_binary` call |
| fd binary name difference | Create symlink `/usr/local/bin/fd-find` → `/usr/bin/fd` |
| OpenSSL version differences | OpenSSL 3.x has stable ABI; verify with wget HTTPS test |

## Critical Files

- `internal/containerbuild/Dockerfile` - Main implementation
- `internal/containerbuild/embed.go` - No changes needed
- `internal/containerbuild/args_test.go` - Tests ARG variables (no changes needed)
- `internal/containerbuild/resource_test.go` - Tests embedding (no changes needed)

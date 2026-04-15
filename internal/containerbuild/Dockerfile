# --- Stage 1: Build sem and sem-mcp from source ---
FROM rust:1-alpine3.21 AS sem-builder

ARG SEM_VERSION=latest

RUN apk add --no-cache git pkgconfig openssl-dev openssl-libs-static cmake musl-dev

RUN set -eux; \
    SEM_VER="${SEM_VERSION}"; \
    if [ "${SEM_VER}" = "latest" ]; then \
        SEM_VER="$(git ls-remote --tags --sort=-v:refname https://github.com/Ataraxy-Labs/sem.git | head -n1 | sed 's/.*refs\/tags\///')"; \
        if [ -z "${SEM_VER}" ]; then \
            echo "failed to resolve sem latest version" >&2; exit 1; \
        fi; \
    fi; \
    if [ "${SEM_VER#v}" = "${SEM_VER}" ]; then \
        SEM_VER="v${SEM_VER}"; \
    fi; \
    git clone --branch "${SEM_VER}" --depth 1 https://github.com/Ataraxy-Labs/sem.git /build && \
    cargo build --release -p sem-cli -p sem-mcp --manifest-path /build/crates/Cargo.toml

# --- Stage 2: Runtime image ---
FROM golang:1.25-alpine3.21

USER root

ARG GOLANGCI_LINT_VERSION=2.11.4
ARG CRUSH_VERSION=latest
ARG GOTESTSUM_VERSION=latest
ARG MOXIDE_VERSION=latest
ARG SEM_VERSION=latest

RUN addgroup -g 1000 agent && \
    adduser -D -u 1000 -G agent agent

RUN apk add --no-cache git ripgrep fd wget ca-certificates openssl gcompat github-cli

RUN ln -s /usr/bin/fd /usr/local/bin/fd-find

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

RUN CRUSH_VER="${CRUSH_VERSION}"; \
    GOTESTSUM_VER="${GOTESTSUM_VERSION}"; \
    if [ "${CRUSH_VER}" != "latest" ] && [ "${CRUSH_VER#v}" = "${CRUSH_VER}" ]; then \
        CRUSH_VER="v${CRUSH_VER}"; \
    fi; \
    if [ "${GOTESTSUM_VER}" != "latest" ] && [ "${GOTESTSUM_VER#v}" = "${GOTESTSUM_VER}" ]; then \
        GOTESTSUM_VER="v${GOTESTSUM_VER}"; \
    fi; \
    GOPATH=/tmp/go GOTOOLCHAIN=auto go install "github.com/charmbracelet/crush@${CRUSH_VER}" && \
    GOPATH=/tmp/go go install "gotest.tools/gotestsum@${GOTESTSUM_VER}" && \
    GOPATH=/tmp/go go install "golang.org/x/tools/gopls@latest" && \
    mv /tmp/go/bin/* /usr/local/bin/ && \
    rm -rf /tmp/go

RUN set -eux; \
    install_github_binary() { \
        repo=$1; ver=$2; asset_tpl=$3; name=$4; fallback=$5; arch_x86=$6; arch_arm64=$7; \
        if [ "${ver}" = "latest" ]; then \
            ver="$(wget -qO- "https://api.github.com/repos/${repo}/releases/latest" | awk -F\" '/tag_name/ {print $4; exit}')"; \
            if [ -z "${ver}" ]; then \
                echo "failed to resolve ${repo} latest version, falling back" >&2; \
                ver="${fallback}"; \
            fi; \
        fi; \
        if [ "${ver#v}" = "${ver}" ]; then \
            ver="v${ver}"; \
        fi; \
        case "$(uname -m)" in \
            x86_64) arch="${arch_x86}" ;; \
            aarch64|arm64) arch="${arch_arm64}" ;; \
            *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;; \
        esac; \
        asset="$(printf '%s' "${asset_tpl}" | sed "s/{VER}/${ver}/g; s/{ARCH}/${arch}/g")"; \
        url="https://github.com/${repo}/releases/download/${ver}/${asset}"; \
        wget -q --retry-connrefused --tries=3 --timeout=30 -O "/tmp/${asset}" "${url}"; \
        case "${asset}" in \
            *.tar.gz) \
                tar -xzf "/tmp/${asset}" -C /tmp; \
                install -m 0755 "/tmp/${name}" "/usr/local/bin/${name}"; \
                rm -f "/tmp/${asset}" "/tmp/${name}" ;; \
            *) \
                install -m 0755 "/tmp/${asset}" "/usr/local/bin/${name}"; \
                rm -f "/tmp/${asset}" ;; \
        esac; \
        "/usr/local/bin/${name}" --version >/dev/null 2>&1 || true; \
    }; \
    install_github_binary "Feel-ix-343/markdown-oxide" "${MOXIDE_VERSION}" \
        "markdown-oxide-{VER}-{ARCH}" markdown-oxide v0.25.10 \
        x86_64-unknown-linux-gnu aarch64-unknown-linux-gnu

COPY --from=sem-builder /build/crates/target/release/sem /usr/local/bin/sem
COPY --from=sem-builder /build/crates/target/release/sem-mcp /usr/local/bin/sem-mcp

USER agent

ENV CRUSH_SHORT_TOOL_DESCRIPTIONS=1
ENV PATH="/home/agent/.local/bin:${PATH}"

RUN sem setup && \
    sed -i 's|exec sem diff|exec sem diff --verbose|' /home/agent/.local/bin/sem-diff-wrapper

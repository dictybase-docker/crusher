FROM docker/sandbox-templates:shell

USER root

# Build arguments for tool versions (defaults match CLI defaults)
ARG GOLANGCI_LINT_VERSION=2.11.4
ARG CRUSH_VERSION=latest
ARG GOTESTSUM_VERSION=latest
ARG MOXIDE_VERSION=latest

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
# Both crush and gotestsum require 'v' prefix (e.g., v0.55.0), prepend if missing
RUN CRUSH_VER="${CRUSH_VERSION}"; \
    GOTESTSUM_VER="${GOTESTSUM_VERSION}"; \
    if [ "${CRUSH_VER}" != "latest" ] && [ "${CRUSH_VER#v}" = "${CRUSH_VER}" ]; then \
        CRUSH_VER="v${CRUSH_VER}"; \
    fi; \
    if [ "${GOTESTSUM_VER}" != "latest" ] && [ "${GOTESTSUM_VER#v}" = "${GOTESTSUM_VER}" ]; then \
        GOTESTSUM_VER="v${GOTESTSUM_VER}"; \
    fi; \
    GOPATH=/tmp/go go install "github.com/charmbracelet/crush@${CRUSH_VER}" && \
    GOPATH=/tmp/go go install "gotest.tools/gotestsum@${GOTESTSUM_VER}" && \
    mv /tmp/go/bin/* /usr/local/bin/ && \
    rm -rf /tmp/go

# Install markdown-oxide binary (supports version arg and arch detection)
RUN set -eux; \
    MOXIDE_VER="${MOXIDE_VERSION}"; \
    if [ "${MOXIDE_VER}" = "latest" ]; then \
        MOXIDE_VER="$(wget -qO- https://api.github.com/repos/Feel-ix-343/markdown-oxide/releases/latest | awk -F\" '/tag_name/ {print $4; exit}')"; \
        if [ -z "${MOXIDE_VER}" ]; then \
            echo "failed to resolve markdown-oxide latest version, falling back" >&2; \
            MOXIDE_VER="v0.25.10"; \
        fi; \
    fi; \
    if [ "${MOXIDE_VER#v}" = "${MOXIDE_VER}" ]; then \
        MOXIDE_VER="v${MOXIDE_VER}"; \
    fi; \
    case "$(uname -m)" in \
        x86_64) ARCH="x86_64-unknown-linux-gnu" ;; \
        aarch64|arm64) ARCH="aarch64-unknown-linux-gnu" ;; \
        *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;; \
    esac; \
    MOX_BIN="markdown-oxide-${MOXIDE_VER}-${ARCH}"; \
    MOX_URL="https://github.com/Feel-ix-343/markdown-oxide/releases/download/${MOXIDE_VER}/${MOX_BIN}"; \
    wget -q --retry-connrefused --tries=3 --timeout=30 -O "/tmp/${MOX_BIN}" "${MOX_URL}"; \
    install -m 0755 "/tmp/${MOX_BIN}" /usr/local/bin/markdown-oxide; \
    /usr/local/bin/markdown-oxide --version >/dev/null 2>&1 || true; \
    rm -f "/tmp/${MOX_BIN}"

USER agent

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

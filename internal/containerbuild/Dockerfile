FROM docker/sandbox-templates:shell

USER root

# Install apt packages: ripgrep, gopls, fd-find, wget
RUN apt-get update && apt-get install -y \
    ripgrep \
    gopls \
    fd-find \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install golangci-lint from official .deb package (v2.11.4)
# Download directly from GitHub releases and verify checksum
RUN set -eux; \
    GOLANGCI_VERSION="2.11.4"; \
    GOLANGCI_CHECKSUM="ec504050bee6e473d074d2a4b2fa57472390083b0a0bdd4435a838e922a21c81"; \
    GOLANGCI_DEB="golangci-lint-${GOLANGCI_VERSION}-linux-arm64.deb"; \
    GOLANGCI_URL="https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_VERSION}/${GOLANGCI_DEB}"; \
    wget -q "${GOLANGCI_URL}" -O "/tmp/${GOLANGCI_DEB}"; \
    echo "${GOLANGCI_CHECKSUM}  /tmp/${GOLANGCI_DEB}" | sha256sum -c -; \
    dpkg -i "/tmp/${GOLANGCI_DEB}"; \
    rm -f "/tmp/${GOLANGCI_DEB}"

# Install Go tools to /usr/local/bin (system-wide)
RUN GOPATH=/tmp/go go install github.com/charmbracelet/crush@latest && \
    GOPATH=/tmp/go go install gotest.tools/gotestsum@latest && \
    mv /tmp/go/bin/* /usr/local/bin/ && \
    rm -rf /tmp/go

USER agent

FROM docker/sandbox-templates:shell

USER root

# Build arguments for tool versions (defaults match CLI defaults)
ARG GOLANGCI_LINT_VERSION=2.11.4
ARG CRUSH_VERSION=latest
ARG GOTESTSUM_VERSION=latest
ARG MOXIDE_VERSION=latest
ARG SEM_VERSION=latest

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

# Install binary tools from GitHub releases
# install_github_binary(repo, version, asset_template, install_name, fallback_version, arch_x86, arch_arm64)
# Asset template uses {VER} and {ARCH} placeholders; .tar.gz assets are extracted automatically
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
        x86_64-unknown-linux-gnu aarch64-unknown-linux-gnu && \
    install_github_binary "Ataraxy-Labs/sem" "${SEM_VERSION}" \
        "sem-{ARCH}.tar.gz" sem v0.3.14 \
        linux-x86_64 linux-arm64

USER agent

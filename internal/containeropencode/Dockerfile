# --- Stage 1: Build sem, sem-mcp, and rtk from source ---
FROM rust:1-slim-bookworm AS sem-builder

ARG SEM_VERSION=latest
ARG RTK_VERSION=latest

RUN apt-get update && apt-get install -y --no-install-recommends \
    git pkg-config libssl-dev cmake musl-tools && rm -rf /var/lib/apt/lists/*

RUN set -eux; \
    SEM_VER="${SEM_VERSION}"; \
    if [ "${SEM_VER}" = "latest" ]; then \
        SEM_VER="$(git ls-remote --tags --sort=-v:refname https://github.com/Ataraxy-Labs/sem.git | head -n1 | sed 's/.*refs\/tags\///')"; \
        if [ -z "${SEM_VER}" ]; then echo "failed to resolve sem latest version" >&2; exit 1; fi; \
    fi; \
    if [ "${SEM_VER#v}" = "${SEM_VER}" ]; then SEM_VER="v${SEM_VER}"; fi; \
    git clone --branch "${SEM_VER}" --depth 1 https://github.com/Ataraxy-Labs/sem.git /build && \
    cargo build --release -p sem-cli -p sem-mcp --manifest-path /build/crates/Cargo.toml

RUN set -eux; \
    RTK_VER="${RTK_VERSION}"; \
    if [ "${RTK_VER}" = "latest" ]; then \
        RTK_VER="$(git ls-remote --tags --sort=-v:refname https://github.com/rtk-ai/rtk.git | head -n1 | sed 's/.*refs\/tags\///')"; \
        if [ -z "${RTK_VER}" ]; then echo "failed to resolve rtk latest version" >&2; exit 1; fi; \
    fi; \
    if [ "${RTK_VER#v}" = "${RTK_VER}" ]; then RTK_VER="v${RTK_VER}"; fi; \
    git clone --branch "${RTK_VER}" --depth 1 https://github.com/rtk-ai/rtk.git /rtk-build && \
    cargo build --release --manifest-path /rtk-build/Cargo.toml

# --- Stage 2: Runtime image extending the official opencode template ---
# docker/sandbox-templates:opencode is Ubuntu-based and already ships opencode,
# the agent user (UID 1000) with passwordless sudo, and all proxy wiring.
FROM docker/sandbox-templates:opencode

USER root

ARG GO_VERSION=1.26.3
ARG GOLANGCI_LINT_VERSION=2.11.4
ARG GOTESTSUM_VERSION=latest
ARG MOXIDE_VERSION=latest
ARG SEM_VERSION=latest
ARG RTK_VERSION=latest
ARG GREPAI_VERSION=latest

RUN apt-get update && apt-get install -y --no-install-recommends \
    ripgrep fd-find wget ca-certificates gh && \
    ln -s /usr/bin/fdfind /usr/local/bin/fd-find && \
    ln -s /usr/bin/fdfind /usr/local/bin/fd && \
    rm -rf /var/lib/apt/lists/*

RUN wget -q "https://go.dev/dl/go${GO_VERSION}.linux-$(dpkg --print-architecture).tar.gz" \
    -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz && \
    echo 'export PATH="/usr/local/go/bin:$PATH"' >> /etc/profile.d/golang.sh

RUN set -eux; \
    case "$(dpkg --print-architecture)" in \
        amd64) ARCH="amd64" ;; \
        arm64) ARCH="arm64" ;; \
        *) echo "unsupported architecture" >&2; exit 1 ;; \
    esac; \
    GOLANGCI_TAR="golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz"; \
    wget -q "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/${GOLANGCI_TAR}" \
        -O "/tmp/${GOLANGCI_TAR}"; \
    tar -xzf "/tmp/${GOLANGCI_TAR}" -C /tmp; \
    install -m 0755 "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}/golangci-lint" \
        /usr/local/bin/golangci-lint; \
    rm -rf "/tmp/${GOLANGCI_TAR}" "/tmp/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-${ARCH}"

RUN export PATH="/usr/local/go/bin:$PATH"; \
    GOTESTSUM_VER="${GOTESTSUM_VERSION}"; \
    if [ "${GOTESTSUM_VER}" != "latest" ] && [ "${GOTESTSUM_VER#v}" = "${GOTESTSUM_VER}" ]; then \
        GOTESTSUM_VER="v${GOTESTSUM_VER}"; \
    fi; \
    GOPATH=/tmp/go go install "gotest.tools/gotestsum@${GOTESTSUM_VER}" && \
    GOPATH=/tmp/go go install "golang.org/x/tools/gopls@latest" && \
    mv /tmp/go/bin/* /usr/local/bin/ && rm -rf /tmp/go

RUN set -eux; \
    install_github_binary() { \
        repo=$1; ver=$2; asset_tpl=$3; name=$4; fallback=$5; arch_x86=$6; arch_arm64=$7; \
        if [ "${ver}" = "latest" ]; then \
            ver="$(wget -qO- "https://api.github.com/repos/${repo}/releases/latest" | awk -F\" '/tag_name/ {print $4; exit}')"; \
            [ -z "${ver}" ] && ver="${fallback}"; \
        fi; \
        [ "${ver#v}" = "${ver}" ] && ver="v${ver}"; \
        case "$(dpkg --print-architecture)" in \
            amd64) arch="${arch_x86}" ;; \
            arm64) arch="${arch_arm64}" ;; \
        esac; \
        asset="$(printf '%s' "${asset_tpl}" | sed "s/{VER}/${ver}/g; s/{ARCH}/${arch}/g")"; \
        wget -q "https://github.com/${repo}/releases/download/${ver}/${asset}" -O "/tmp/${asset}"; \
        case "${asset}" in \
            *.tar.gz) tar -xzf "/tmp/${asset}" -C /tmp && install -m 0755 "/tmp/${name}" "/usr/local/bin/${name}" && rm -f "/tmp/${asset}" "/tmp/${name}" ;; \
            *) install -m 0755 "/tmp/${asset}" "/usr/local/bin/${name}" && rm -f "/tmp/${asset}" ;; \
        esac; \
    }; \
    install_github_binary "Feel-ix-343/markdown-oxide" "${MOXIDE_VERSION}" \
        "markdown-oxide-{VER}-{ARCH}" markdown-oxide v0.25.10 \
        x86_64-unknown-linux-gnu aarch64-unknown-linux-gnu

RUN set -eux; \
    GREPAI_VER="${GREPAI_VERSION}"; \
    if [ "${GREPAI_VER}" = "latest" ]; then \
        GREPAI_VER="$(wget -qO- "https://api.github.com/repos/yoanbernabeu/grepai/releases/latest" | awk -F\" '/tag_name/ {print $4; exit}')"; \
        [ -z "${GREPAI_VER}" ] && GREPAI_VER="v0.35.0"; \
    fi; \
    [ "${GREPAI_VER#v}" = "${GREPAI_VER}" ] && GREPAI_VER="v${GREPAI_VER}"; \
    case "$(dpkg --print-architecture)" in \
        amd64) ARCH="amd64" ;; \
        arm64) ARCH="arm64" ;; \
        *) echo "unsupported architecture" >&2; exit 1 ;; \
    esac; \
    ASSET="grepai_${GREPAI_VER#v}_linux_${ARCH}.tar.gz"; \
    wget -q "https://github.com/yoanbernabeu/grepai/releases/download/${GREPAI_VER}/${ASSET}" -O "/tmp/${ASSET}"; \
    tar -xzf "/tmp/${ASSET}" -C /tmp && install -m 0755 /tmp/grepai /usr/local/bin/grepai && rm -f "/tmp/${ASSET}" /tmp/grepai

COPY --from=sem-builder /build/crates/target/release/sem /usr/local/bin/sem
COPY --from=sem-builder /build/crates/target/release/sem-mcp /usr/local/bin/sem-mcp
COPY --from=sem-builder /rtk-build/target/release/rtk /usr/local/bin/rtk

USER agent

ENV PATH="/home/agent/.local/bin:/usr/local/go/bin:${PATH}"

RUN sem setup && \
    sed -i 's|exec sem diff|exec sem diff --verbose|' /home/agent/.local/bin/sem-diff-wrapper

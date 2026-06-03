# Build Stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download
# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build binary
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /bin/crusher ./cmd/crusher

# Runtime Stage
FROM gcr.io/distroless/static-debian12

LABEL org.opencontainers.image.source="https://github.com/dictybase-docker/crusher"

COPY --from=builder /bin/crusher /bin/crusher

ENTRYPOINT ["/bin/crusher"]
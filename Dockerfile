# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git for version info
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with optimizations
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETPLATFORM

RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -X github.com/shantoislamdev/kothaset/internal/cli.Version=${VERSION} -X github.com/shantoislamdev/kothaset/internal/cli.Commit=${COMMIT} -X github.com/shantoislamdev/kothaset/internal/cli.BuildDate=${BUILD_DATE}" \
    -o /kothaset ./cmd/kothaset

# Runtime stage
FROM alpine:3.19

# Add ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -u 1000 kothaset
USER kothaset

# Copy binary - for dockers_v2, GoReleaser puts binaries in $TARGETPLATFORM/
ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/kothaset /usr/local/bin/kothaset

# Default command
ENTRYPOINT ["kothaset"]
CMD ["--help"]

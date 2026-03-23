# Multi-stage Dockerfile for containertui - Go CLI application for Docker container management
# Builder stage: Compiles the Go binary
# Runtime stage: Minimal Alpine image with containertui and docker-cli

# Stage 1: Builder
FROM golang:1.25.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /build

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
# CGO_ENABLED=0 for static binary, -ldflags "-s -w" strips debugging info for smaller binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o containertui \
    ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    docker-cli \
    ca-certificates

# Create non-root user
RUN addgroup -g 1000 ctui && \
    adduser -D -u 1000 -G ctui ctui

# Set working directory
WORKDIR /home/ctui

# Copy binary from builder
COPY --from=builder /build/containertui /usr/local/bin/containertui

# Ensure binary is executable
RUN chmod +x /usr/local/bin/containertui

# Switch to non-root user
USER ctui

# Set entrypoint to containertui
ENTRYPOINT ["containertui"]

# Default command is --help
CMD ["--help"]

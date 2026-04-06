FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache \
  git \
  ca-certificates \
  tzdata

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary with optimizations
# CGO_ENABLED=0 for static binary, -ldflags "-s -w" strips debugging info for smaller binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-s -w" \
  -o containertui \
  ./cmd/main.go

FROM docker:dind

RUN apk add --no-cache \
  bash \
  ca-certificates \
  fontconfig \
  curl

# Install JetBrains Mono Nerd Font
RUN mkdir -p /usr/local/share/fonts/nerd-fonts && \
  curl -fsSL https://github.com/ryanoasis/nerd-fonts/releases/download/v3.0.2/JetBrainsMono.zip -o /tmp/JetBrainsMono.zip && \
  unzip -o /tmp/JetBrainsMono.zip -d /usr/local/share/fonts/nerd-fonts/ && \
  rm /tmp/JetBrainsMono.zip && \
  fc-cache -fv

# Set working directory
WORKDIR /demo

# Copy binary from builder
COPY --from=builder /build/containertui /usr/local/bin/containertui

# Ensure binary is executable
RUN chmod +x /usr/local/bin/containertui

COPY demos/setup.sh /demo/setup.sh
COPY demos/cleanup.sh /demo/cleanup.sh
COPY demos/demo-entrypoint.sh /demo/demo-entrypoint.sh

RUN chmod +x /demo/setup.sh /demo/cleanup.sh /demo/demo-entrypoint.sh

# Set entrypoint to dual-mode runtime launcher
ENTRYPOINT ["/demo/demo-entrypoint.sh"]

# Default runtime mode autodetects host socket, then falls back to DinD.
ENV CTUI_DEMO_MODE=auto
ENV CTUI_DEMO_SEED=1
ENV CTUI_DEMO_CLEANUP_ON_EXIT=1

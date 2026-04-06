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

FROM docker:cli

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
WORKDIR /workspace

# Copy binary from builder
COPY --from=builder /build/containertui /usr/local/bin/containertui

# Ensure binary is executable
RUN chmod +x /usr/local/bin/containertui

# Production image: run the TUI directly and use mounted host Docker socket.
ENTRYPOINT ["/usr/local/bin/containertui"]

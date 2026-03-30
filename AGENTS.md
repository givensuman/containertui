# AGENTS.md - Agent Guide for ContainerTUI

## Project Overview

ContainerTUI is a terminal-based user interface (TUI) for managing container lifecycles, built with Go using the Charm stack (Bubble Tea v2, Lipgloss v2). It provides seamless navigation between containers, images, volumes, networks, and services via Docker/Moby API.

**Status:** Under heavy development. Expect frequent breaking changes and incomplete features. Not ready for production use.

## Essential Commands

### Build & Run

```bash
# Build from source
go build -o containertui ./cmd

# Run directly
go run ./cmd

# Install locally
go install ./cmd

# Run tests
go test ./...

# Run specific package tests
go test ./internal/client/...
go test ./internal/config/...

# Run with race detection
go test -race ./...

# Build and run demo
docker build -f Dockerfile.demo -t containertui-demo .
```

### Docker

```bash
# Build and run in Docker
docker build -t containertui .
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock containertui

# Run the demo
docker run --rm -it ghcr.io/givensuman/containertui:latest
```

## Code Organization

```
containertui/
├── cmd/
│   └── main.go                # CLI entry point
├── internal/
│   ├── backend/               # State and configuration
│   ├── client/                # Docker/Moby client wrapper
│   │   ├── client.go          # Client initialization and operations
│   │   ├── constants.go       # Docker API constants
│   │   └── errors.go          # Error handling
│   ├── colors/                # Color and theme management
│   │   ├── colors.go          # Color definitions
│   │   ├── theme.go           # Theme application
│   │   └── parse.go           # Color parsing utilities
│   ├── config/                # User configuration
│   │   ├── config.go          # Config loading and defaults
│   │   ├── types.go           # Config types
│   │   └── theme.go           # Theme configuration
│   ├── registry/              # Command/action registry
│   ├── state/                 # Application state management
│   └── ui/                    # User interface components
│       ├── containers/        # Containers tab
│       │   ├── containers.go  # Main containers view
│       │   ├── item.go        # Container list item
│       │   ├── logs.go        # Container logs view
│       │   └── messages.go    # Event messages
│       ├── images/            # Images tab
│       │   ├── images.go      # Main images view
│       │   └── item.go        # Image list item
│       ├── networks/          # Networks tab
│       │   ├── networks.go    # Main networks view
│       │   └── item.go        # Network list item
│       ├── volumes/           # Volumes tab
│       │   ├── volumes.go     # Main volumes view
│       │   └── item.go        # Volume list item
│       └── services/          # Services tab
├── assets/                    # Logos and demo assets
├── demos/                     # Demo files
├── docs/                      # Documentation
├── scripts/                   # Build and utility scripts
└── tests/                     # Integration tests
```

## Architecture Patterns

### Bubble Tea MVU Pattern

ContainerTUI follows the Model-View-Update pattern:
- **Model**: Application state in `internal/state/` and `internal/backend/`
- **View**: UI components in `internal/ui/`
- **Update**: Event handling and state updates in component Update() methods

### Tabbed Interface

The main application uses a tabbed layout with these primary tabs:
1. **Containers**: Manage running containers
2. **Images**: Browse and inspect images
3. **Volumes**: List and inspect Docker volumes
4. **Networks**: View Docker networks
5. **Services**: Monitor Docker Compose services

Each tab is a self-contained component in `internal/ui/<resource>/`.

### Docker Client Abstraction

- Docker operations wrapped in `internal/client/` (Moby API wrapper)
- Error handling standardized in `client/errors.go`
- Constants for Docker operations in `client/constants.go`

### Theme & Color System

- Base colors defined in `internal/colors/colors.go`
- Theme application in `internal/colors/theme.go`
- User configuration in `internal/config/theme.go`
- Color parsing utilities in `internal/colors/parse.go`

## Key Dependencies

- **Bubble Tea v2** (`charm.land/bubbletea/v2`) - TUI framework
- **Bubbles v2** (`charm.land/bubbles/v2`) - UI components
- **Lipgloss v2** (`charm.land/lipgloss/v2`) - Styling
- **Moby/Docker** (`github.com/docker/docker`, `github.com/moby/moby`) - Container engine
- **Cobra** (`github.com/spf13/cobra`) - CLI commands
- **yaml** (`gopkg.in/yaml.v3`) - Configuration parsing

> **Note:** Uses modern Charm stack packages from `charm.land/*` module paths (post-2024 migration).

## Coding Conventions

### Go Style

- Follow standard Go conventions ([Effective Go](https://go.dev/doc/effective_go))
- Run `go fmt` before committing
- Package comments on all packages
- Meaningful variable names (avoid single letters except loop indices)

### Error Handling

- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Log warnings for non-fatal issues using standard library `log` package
- Return early on errors
- Use custom error types in `client/errors.go` for Docker-specific errors

### Documentation

- Package-level doc comments required
- Exported types/functions need doc comments
- Use godoc-style comments

### Testing

- Table-driven tests preferred
- Test file naming: `*_test.go`
- Benchmarks with `Benchmark*` prefix
- Use `t.Run()` for subtests

## Testing Approach

### Unit Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/client/...
go test ./internal/config/...
go test ./internal/colors/...

# With verbose output
go test -v ./internal/config/...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./tests/...
```

### Manual Testing Checklist

When testing UI/UX changes:
- [ ] Launch application and verify main tab loads
- [ ] Navigate between all tabs (Containers, Images, Volumes, Networks, Services)
- [ ] List containers and verify all columns display correctly
- [ ] Click/select items and verify details appear
- [ ] Test filtering/search if available
- [ ] Verify color theme applies correctly
- [ ] Test with different terminal sizes (resize terminal)
- [ ] Verify no panics or errors in logs
- [ ] Test quit/exit functionality (usually `q` key)

## Common Gotchas

### Bubble Tea v2 Specifics

- Use `tea.KeyPressMsg` not `tea.KeyMsg` (v2 change)
- Mouse events are separate types: `tea.MouseClickMsg`, `tea.MouseMotionMsg`
- `tea.WithFilter()` for event filtering

### Docker Client

- Docker socket path differs by platform:
  - Linux: `/var/run/docker.sock`
  - macOS: `~/.docker/run/docker.sock` or Docker Desktop socket
  - Windows: Named pipe `//./pipe/docker_engine`
- Ensure proper error handling for connection failures
- Set appropriate timeouts for long-running operations

### Terminal Rendering

- Always consider minimum terminal size (usually 80x24)
- Test with both nerd fonts and standard fonts
- Account for ANSI color support variations
- Use proper viewport/scrolling for long lists

### Configuration

- Config file location: `~/.config/containertui/config.yaml` (XDG standard on Linux)
- Ensure backward compatibility when adding config options
- Provide sensible defaults for all config values

## Important Files to Know

| Purpose | File |
|---------|------|
| Main entry point | `cmd/main.go` |
| Docker client | `internal/client/client.go` |
| Configuration | `internal/config/config.go` |
| Theme management | `internal/colors/colors.go` |
| Containers UI | `internal/ui/containers/containers.go` |
| Images UI | `internal/ui/images/images.go` |
| Volumes UI | `internal/ui/volumes/volumes.go` |
| Networks UI | `internal/ui/networks/networks.go` |
| Services UI | `internal/ui/services/services.go` |

## Commit Message Format

Use conventional commits:
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks
- `perf:` - Performance improvements

Examples:
```
feat: add container pause/unpause functionality
fix: handle Docker daemon connection errors gracefully
docs: update configuration documentation
refactor: extract volume filtering logic to helper function
test: add unit tests for color parsing
```

## Release Process

Releases are automated via GitHub Actions with GoReleaser:
- Tag format: `v*.*.*` (e.g., `v0.1.0`)
- Builds for Linux, macOS, Windows
- Publishes to GitHub Releases
- Publishes Docker image to GHCR

## Development Workflow

### Setting Up Local Development

1. Clone the repository
2. Ensure Docker daemon is running
3. Run `go mod download` to fetch dependencies
4. Build with `go build -o containertui ./cmd`
5. Run `./containertui` to start the TUI

### Making Changes

1. Create a feature branch from `main`
2. Make changes with appropriate tests
3. Run `go fmt ./...` to format code
4. Run `go test ./...` to verify tests pass
5. Create a pull request with descriptive title and body

### Debugging

- Enable verbose logging if available (check flags in main.go)
- Use Docker CLI directly to verify expected behavior
- Test against different Docker versions when possible
- Check Docker daemon logs for API-related issues

## Additional Resources

- **README**: `README.md` - Project overview and installation
- **Issues**: `ISSUES` - Known issues and feature requests
- **Dockerfile**: Multi-stage build for production container
- **Justfile**: Build recipes and common tasks
- **.goreleaser.yml**: Release configuration

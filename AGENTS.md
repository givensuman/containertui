# AGENTS.md - Agent Guide for CONTAINERTUI

## Project Overview

CONTAINERTUI is a terminal-based container manager built in Go using the Charm stack (Bubble Tea, Lipgloss). It aims to replicate the functionality of GUI-based container managers like Docker Desktop or Portainer, but entirely within the terminal. It supports as many backends as possible, including Docker, and Podman.

## Commands

Essential commands for development are container in our Justfile.

```

## Code Organization

```
containertui/
├── cmd/              # CLI entry point (main.go with cobra commands)
├── internal/
│   ├── app/                # TUI application core (Bubble Tea MVU)
│   │   ├── containers/     # Main view for managing containers
│   │   ├── images/         # View for managing images
│   │   ├── networks/       # View for managing networks
│   │   ├── volumes/        # View for managing volumes
│   │   ├── services/       # View for managing services (e.g. docker-compose outputs)
│   │   └── browse/         # View for remote repository browsing (Docker Hub, etc)
│   ├── config/             # Configuration and keybindings
│   │   ├── userconfig.go   # TOML config loading, defaults
│   │   └── validation.go   # Config validation
│   ├── theme/              # Color theming
│   ├── layout/             # Component layout logic (ratios, grids)
├── docs/                   # Documentation
    ├── ARCHITECTURE.md     # Technical architecture diagrams
    ├── KEYBINDINGS.md      # Complete keybinding reference
    ├── CONFIGURATION.md    # Config options
    └── CLI_REFERENCE.md    # CLI flags and commands

## Architecture Patterns

### Bubble Tea MVU Pattern

CONTAINERTUI follows Model-View-Update:

Where possible, children should have their state managed by pointer receivers:
  ```
  Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      m.child.DoSomething()
      return m, cmd
  }
  ```

If children want to influence the state of their parents (or any higher-level state), they should send messages up the chain:
  ```
  Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      switch msg := msg.(type) {
      case child.SomeEventMsg:
          m.doSomethingInResponse()
      }
      return m, nil
  }
  ```

## Key Dependencies

- **Bubble Tea v2** (`charm.land/bubbletea/v2`) - TUI framework
- Bubbles v2 (charm.land/bubbles/v2) - UI components
- **Lipgloss v2** (`charm.land/lipgloss/v2`) - Styling
- **Cobra** (`github.com/spf13/cobra`) - CLI commands

> **Note:** As of December 2025, the Charm stack packages have migrated from `github.com/charmbracelet/*` to `charm.land/*` module paths.

## Coding Conventions

### Go Style

- Follow standard Go conventions ([Effective Go](https://go.dev/doc/effective_go))
- Package comments on all packages (see existing `internal/*/` packages)
- Meaningful variable names (avoid single letters except loop indices)
- Prefer verbosity in variable and function names over brevity

### Error Handling

- Wrap errors with context: `fmt.Errorf("failed to X: %w", err)`
- Log warnings for non-fatal issues: `log.Printf("Warning: ...")`
- Return early on errors
- In non-fatal errors, send a notification to the user via the TUI

### Documentation

- Package-level doc comments required
- Exported types/functions need doc comments
- Use godoc-style comments

### Testing

- Test file naming: `*_test.go`
- Benchmarks with `Benchmark*` prefix
- Use `t.Run()` for subtests

## Testing Approach

### Unit Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/tape/...

# With verbose output
go test -v ./internal/config/...

# Run benchmarks
go test -bench=. ./internal/app/...
```

## Common Gotchas

### Bubble Tea v2 Specifics

- Use `tea.KeyPressMsg` not `tea.KeyMsg` (v2 change)
- Mouse events are separate types: `tea.MouseClickMsg`, `tea.MouseMotionMsg`, etc.
- `tea.WithFilter()` for event filtering (used for mouse motion filtering)

### Keybind Registry

- Check action descriptions in `internal/config/registry.go`
- Key normalization handles `opt+` → `alt+` conversion on macOS


## Additional Notes

Do not commit to `git`. Rely on the programmer to use git commits to mark working states during development.

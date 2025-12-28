# OCI-TUI Project Plan

## Overview

This plan outlines the creation of `oci-tui`, a Terminal User Interface (TUI) application for managing Docker containers, images, networks, and volumes. The application merges the sophisticated UI architecture of `gh-dash` (a GitHub dashboard TUI) with the Docker management functionality of `laboon` (a simple Docker container TUI). Key enhancements include:

- Tab-based navigation for different Docker entities (Containers, Images, Networks, Volumes)
- Three-panel layout: list (left), operations (middle), logs/metrics (right)
- Real-time ASCII graphs for system usage using `guptarohit/asciigraph`
- Comprehensive Docker operations with keyboard-driven interactions
- Modular, extensible architecture for future features

## Project Goals

- Provide a user-friendly TUI alternative to Docker Desktop
- Leverage existing open-source codebases for rapid development
- Ensure high code quality with proper testing and linting
- Maintain clear separation between borrowed code and new implementations

## Architecture

The application will follow a component-based architecture similar to `gh-dash`, using Bubble Tea for TUI management:

```
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── config/              # Configuration management (adapted from gh-dash)
│   ├── docker/              # Docker API client wrapper (adapted from laboon)
│   ├── tui/
│   │   ├── components/      # Reusable UI components
│   │   │   ├── tabs/        # Tab navigation (adapted from gh-dash carousel)
│   │   │   ├── list/        # Entity lists (adapted from gh-dash table)
│   │   │   ├── sidebar/     # Operations and details (adapted from gh-dash sidebar)
│   │   │   ├── logs/        # Logs panel (new, with real-time streaming)
│   │   │   └── graphs/      # Metrics graphs (new, using asciigraph)
│   │   ├── models/          # Data models for Docker entities (new)
│   │   ├── views/           # View controllers for each tab (adapted from gh-dash sections)
│   │   └── ui.go            # Main UI model and controller (adapted from gh-dash)
│   └── utils/               # Shared utilities
├── go.mod/go.sum            # Go module files
└── README.md                # Documentation
```

## Components to Adapt/Merge

### From gh-dash

#### Tab/Carousel System
- **Description**: The carousel component provides horizontal tab navigation with loading indicators, item counts, and overflow handling. Each tab represents a section with configurable filters and data sources.
- **Adaptation**: Rename tabs to "Containers", "Images", "Networks", "Volumes". Each tab will load the corresponding Docker entities and display counts (e.g., "Containers (5 running)"). Implement keyboard shortcuts for tab switching (e.g., 1-4 keys).
- **Implementation Steps**:
  1. Copy `carousel.go` from gh-dash
  2. Modify tab titles and icons (use Docker-related icons if available)
  3. Update loading states to show Docker daemon connection status
  4. Add tab-specific key bindings for quick access

#### Section Interface
- **Description**: Sections define data fetching, filtering, and display logic. Each section implements methods for fetching data, rendering rows, and handling user interactions.
- **Adaptation**: Create section interfaces for each Docker entity type. Implement data fetching using Docker API instead of GitHub GraphQL.
- **Implementation Steps**:
  1. Copy section interface and base implementation
  2. Create concrete sections: `ContainerSection`, `ImageSection`, etc.
  3. Implement data models for each entity type
  4. Add Docker-specific filtering (e.g., by status, image name)

#### Sidebar Architecture
- **Description**: The sidebar shows detailed information for selected items with tabbed views (e.g., summary, activity, files for PRs).
- **Adaptation**: Create detail views for each Docker entity with tabs for different information (e.g., for containers: summary, logs, inspect, stats).
- **Implementation Steps**:
  1. Copy sidebar component structure
  2. Implement container detail tabs: Summary (status, ports), Logs (streaming), Inspect (JSON), Stats (resource usage)
  3. Add operation buttons/panels for start/stop/pause actions
  4. Include search/filter within sidebar for logs

#### Table/List Component
- **Description**: Custom table component for displaying rows with sortable columns, selection, and pagination.
- **Adaptation**: Configure columns for Docker data (e.g., for containers: Name, Image, Status, Ports, Created). Add color coding for status (green for running, red for stopped).
- **Implementation Steps**:
  1. Copy table component with sorting and selection logic
  2. Define column configurations for each entity type
  3. Implement status-based styling and icons
  4. Add pagination for large lists (e.g., many images)

#### Configuration System
- **Description**: YAML-based configuration for sections, layouts, themes, and keybindings.
- **Adaptation**: Extend config for Docker-specific settings like default filters, refresh intervals, and Docker connection options.
- **Implementation Steps**:
  1. Copy configuration parser and structures
  2. Add Docker config fields (host, TLS settings, etc.)
  3. Implement default configs for each section
  4. Add validation for Docker-specific options

#### Key Binding System
- **Description**: Configurable keybindings with universal (quit, refresh) and context-specific actions.
- **Adaptation**: Adapt keymaps for Docker operations (start/stop containers, pull images) while preserving common actions.
- **Implementation Steps**:
  1. Copy keybinding structures and parsing
  2. Define Docker-specific keymaps (e.g., 's' for start, 'x' for stop)
  3. Implement context-aware bindings (different actions for different tabs)
  4. Add help overlay showing available actions

#### Styling/Theming
- **Description**: Lipgloss-based styling with configurable themes and adaptive layouts.
- **Adaptation**: Reuse color schemes but adapt for Docker context (e.g., use blue for running containers, red for errors).
- **Implementation Steps**:
  1. Copy theme and styling systems
  2. Define Docker-specific color palette
  3. Adapt component styles for three-panel layout
  4. Ensure responsive design for different terminal sizes

### From laboon

#### Docker Client Integration
- **Description**: DockerWrapper encapsulates the official Docker Go client for API interactions.
- **Adaptation**: Extend to support all Docker entities (containers, images, networks, volumes) and add error handling.
- **Implementation Steps**:
  1. Copy DockerWrapper and client initialization
  2. Add methods for images: ListImages, PullImage, RemoveImage, InspectImage
  3. Add methods for networks: ListNetworks, CreateNetwork, RemoveNetwork
  4. Add methods for volumes: ListVolumes, CreateVolume, RemoveVolume
  5. Improve error handling with user-friendly messages

#### Container Operations
- **Description**: Basic start, stop, pause, unpause operations with multi-selection support.
- **Adaptation**: Preserve core functionality while integrating with new UI architecture.
- **Implementation Steps**:
  1. Copy operation methods and selection logic
  2. Integrate with task system for async operations
  3. Add progress indicators for long-running operations
  4. Implement batch operations for multiple selections

#### State Management
- **Description**: Optimistic UI updates after operations without re-fetching data.
- **Adaptation**: Combine with gh-dash's reactive state management using Bubble Tea messages.
- **Implementation Steps**:
  1. Copy local state update logic
  2. Integrate with Bubble Tea message system
  3. Add periodic refresh for accurate data
  4. Implement conflict resolution for concurrent operations

### New Components

#### Metrics Graph Panel
- **Description**: ASCII graphs for system and container metrics using asciigraph.
- **Implementation Steps**:
  1. Integrate asciigraph package
  2. Add gopsutil for system metrics collection
  3. Create graph component with configurable series
  4. Implement periodic updates using Bubble Tea ticks
  5. Add CPU, memory, and disk usage graphs

#### Logs Viewer
- **Description**: Real-time container logs streaming in a scrollable panel.
- **Implementation Steps**:
  1. Implement Docker logs API integration
  2. Create scrollable text component
  3. Add log filtering and search
  4. Implement auto-scroll and pause functionality
  5. Handle large log volumes efficiently

#### Image/Network/Volume Management
- **Description**: Full CRUD operations for non-container Docker entities.
- **Implementation Steps**:
  1. Implement entity-specific operations
  2. Create detail views for each entity type
  3. Add creation wizards with form inputs
  4. Implement dependency checking (e.g., can't remove network with containers)

#### Real-time Updates
- **Description**: Periodic refresh of lists and metrics data.
- **Implementation Steps**:
  1. Implement configurable refresh intervals
  2. Add background data fetching
  3. Update UI components reactively
  4. Handle connection failures gracefully

## Implementation Phases

### Phase 1: Project Setup and Core Structure (1-2 weeks)
1. Initialize new Go module: `go mod init oci-tui`
2. Set up basic Bubble Tea application structure in `cmd/main.go`
3. Copy and adapt core UI components from gh-dash:
   - Create `internal/tui/ui.go` with main model and program context
   - Implement basic three-panel layout (list, operations, logs/metrics)
   - Set up tab system foundation
4. Integrate Docker client from laboon in `internal/docker/`
5. Create basic data models for all Docker entities in `internal/tui/models/`
6. Set up configuration system in `internal/config/`

### Phase 2: Container Management (Primary Focus) (2-3 weeks)
1. Implement Containers tab with list view:
   - Create `ContainerSection` implementing section interface
   - Configure table columns: Name, Image, Status, Ports, Created
   - Add status-based styling and icons
2. Add container operations:
   - Integrate start, stop, pause, unpause from laboon
   - Add multi-selection with spacebar
   - Implement optimistic UI updates
3. Create container details sidebar:
   - Summary tab: status, ports, image, created time
   - Logs tab: real-time streaming with scroll/pause
   - Inspect tab: formatted JSON details
   - Stats tab: resource usage graphs
4. Implement search and filtering:
   - Filter by status (running, stopped, paused)
   - Search by name or image
   - Persist filter state per tab

### Phase 3: Additional Entity Support (2-3 weeks)
1. Implement Images tab:
   - List all images with columns: Repository, Tag, ID, Size, Created
   - Operations: pull, remove, inspect
   - Details: layers, size, tags, history
2. Implement Networks tab:
   - List networks with columns: Name, Driver, Scope, Created
   - Operations: create, remove, connect/disconnect containers
   - Details: configuration, connected containers
3. Implement Volumes tab:
   - List volumes with columns: Name, Driver, Mountpoint, Created
   - Operations: create, remove, inspect
   - Details: usage, mount options

### Phase 4: Metrics and Monitoring (1-2 weeks)
1. Integrate asciigraph for system metrics:
   - Add dependency: `go get github.com/guptarohit/asciigraph`
   - Create graph component in `internal/tui/components/graphs/`
2. Add system metrics collection:
   - Integrate gopsutil for CPU, memory, disk
   - Implement periodic data collection (every 2-5 seconds)
3. Display graphs in right panel:
   - CPU usage percentage over time
   - Memory usage percentage over time
   - Disk usage for Docker data directory
4. Add container-specific metrics:
   - Per-container CPU/memory graphs when selected

### Phase 5: Polish and Testing (1-2 weeks)
1. Add comprehensive error handling:
   - Docker daemon connection errors
   - Operation failures with user feedback
   - Network timeouts and retries
2. Implement configuration file support:
   - YAML config for layout, themes, keybindings
   - Docker connection settings
   - Default filters and refresh intervals
3. Add help system and documentation:
   - In-app help overlay with key bindings
   - Update README with usage examples
   - Add tooltips for UI elements
4. Write tests:
   - Unit tests for Docker operations
   - Integration tests for UI interactions
   - Mock Docker client for testing
5. Performance optimization and linting:
   - Run `golangci-lint` and fix issues
   - Optimize rendering for large lists
   - Memory usage monitoring

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling and layout
- `github.com/charmbracelet/bubbles` - Bubble Tea components
- `github.com/docker/docker/client` - Docker API client
- `github.com/guptarohit/asciigraph` - ASCII graphs
- `github.com/shirou/gopsutil/v3` - System metrics
- `github.com/muesli/termenv` - Terminal environment detection
- `gopkg.in/yaml.v3` - Configuration parsing

## Credits and Licensing

All borrowed code must include appropriate credits in source files and documentation:

```go
// Core UI architecture adapted from gh-dash
// https://github.com/dlvhdr/gh-dash
// Copyright (c) 2023 Dolev Hadar
// Licensed under MIT License

// Docker client integration adapted from laboon
// https://github.com/redwebcreation/laboon
// Copyright (c) 2023 Red Web Creation
// Licensed under MIT License

// ASCII graph functionality from asciigraph
// https://github.com/guptarohit/asciigraph
// Copyright (c) 2018 Guptarohit
// Licensed under BSD-3-Clause License
```

## Risk Assessment

- **Complexity of Merge**: Combining two different architectural approaches may require significant refactoring. *Mitigation*: Start with core structure and incrementally add features.
- **API Compatibility**: Docker API changes could affect functionality. *Mitigation*: Use official Docker client library and add version compatibility checks.
- **Performance**: Real-time updates and graphs may impact terminal performance. *Mitigation*: Implement configurable refresh rates and optimize rendering.
- **Terminal Compatibility**: ASCII graphs and styling may not work consistently across all terminals. *Mitigation*: Add terminal detection and fallback modes.

## Success Criteria

- Functional TUI with all planned features working correctly
- Smooth navigation between tabs and panels with responsive updates
- Comprehensive Docker operations coverage (start/stop containers, manage images/networks/volumes)
- Real-time logs and metrics display
- Clean, maintainable codebase with proper credits and documentation
- Full test coverage for critical functionality
- Performance suitable for daily use (sub-second response times)

## Timeline Estimate

- Total Duration: 6-8 weeks
- Phase 1: 1-2 weeks
- Phase 2: 2-3 weeks  
- Phase 3: 2-3 weeks
- Phase 4: 1-2 weeks
- Phase 5: 1-2 weeks

## Next Steps

1. Confirm this detailed plan meets your requirements
2. Address any clarifying questions about scope, priorities, or technical details
3. Once approved, proceed with Phase 1 implementation
4. Set up development environment and initial commit</content>
<parameter name="filePath">PLAN.md
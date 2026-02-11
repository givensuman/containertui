# CONTAINERTUI Roadmap

This document outlines the planned features for CONTAINERTUI beyond Tier 1 implementation.

## Current Status

**Tier 1 (In Progress):** Essential features to match gomanagedocker functionality
- Prune operations
- Toggle show all containers
- Force delete options
- Volume/Network creation
- Image building
- Enhanced log viewing
- Image tagging
- Container renaming

## Tier 2: Enhanced Functionality (Medium Priority)

These features will be implemented after Tier 1 completion to provide advanced container management capabilities.

### 1. Live Container Stats Dashboard
**Priority:** HIGH | **Effort:** Medium (8-10 hours)

**Description:** Real-time monitoring of container resource usage with visual graphs.

**Features:**
- CPU usage percentage with sparkline graphs
- Memory usage/limit with visual bars
- Network I/O (Rx/Tx) with rate indicators
- Block I/O statistics
- Auto-refresh every 1-2 seconds
- Toggle between list view and detailed stats panel

**Technical Notes:**
- Backend `GetContainerStats()` already exists
- Use `charm.land/x/sparkline` for mini graphs
- Add stats toggle keybinding: `S` or `M`
- Option to view stats for selected container or all containers

**Dependencies:**
- github.com/charmbracelet/lipgloss for layout
- Existing stats backend in `internal/client/client.go`

---

### 2. Docker Compose Service Operations
**Priority:** HIGH | **Effort:** High (12-16 hours)

**Description:** Make the Services tab interactive with full compose lifecycle management.

**Features:**
- Start/stop individual services
- Restart services
- Scale services (adjust replica count)
- View service logs
- Pull service images
- Force recreate containers
- View service configuration

**Keybindings:**
- `s` - Start/stop service
- `r` - Restart service
- `S` - Scale service (opens dialog)
- `L` - View service logs
- `p` - Pull service images
- `f` - Force recreate

**Technical Approach:**
- Integration with docker-compose CLI or compose API
- Parse compose files to understand service structure
- Handle multi-container services
- Stream compose operation output

**Dependencies:**
- Requires docker-compose CLI or Docker Compose API
- May need `github.com/compose-spec/compose-go` for parsing

---

### 3. Volume File Browser
**Priority:** MEDIUM | **Effort:** High (10-12 hours)

**Description:** Browse and view files within Docker volumes.

**Features:**
- Tree view of volume contents
- Navigate directories
- View text file contents
- Show file metadata (size, permissions, modified date)
- Search files by name
- Copy file paths
- Read-only mode (no editing/deletion)

**Technical Approach:**
- Create temporary container with volume mounted
- Use container exec to list directory contents
- Stream file contents for viewing
- Clean up temporary container on exit

**Keybindings:**
- `b` - Browse selected volume
- `enter` - Open directory/view file
- `backspace` - Go to parent directory
- `/` - Search files
- `y` - Copy file path

---

### 4. Container Port Management
**Priority:** MEDIUM | **Effort:** Medium (6-8 hours)

**Description:** Dynamically add or remove port mappings on containers.

**Features:**
- View current port mappings in detail
- Add new port mappings (requires container restart)
- Remove port mappings
- Validate port availability
- Show port conflicts

**Technical Notes:**
- Requires stopping and recreating container with new port config
- Warn user about container restart requirement
- Preserve other container settings during recreation

---

### 5. Image Push to Registry
**Priority:** MEDIUM | **Effort:** Medium (8-10 hours)

**Description:** Push local images to Docker Hub or private registries.

**Features:**
- Push to Docker Hub
- Push to private registries
- Authentication handling
- Progress indication
- Tag images before pushing
- Support multiple tags

**Keybindings:**
- `P` (Shift+p) - Push image

**Technical Approach:**
- Form dialog for registry URL and credentials
- Use Docker SDK `ImagePush()` API
- Store registry credentials securely
- Show progress dialog with layer upload status

---

### 6. Network Connect/Disconnect
**Priority:** MEDIUM | **Effort:** Low (4-6 hours)

**Description:** Attach and detach containers from networks dynamically.

**Features:**
- Connect running containers to networks
- Disconnect containers from networks
- View current container network attachments
- Set network aliases
- Configure IP addresses

**Keybindings:**
- `c` - Connect container to network
- `d` - Disconnect container from network

---

### 7. Container Export/Save
**Priority:** LOW | **Effort:** Medium (6-8 hours)

**Description:** Export container filesystems as tarballs.

**Features:**
- Export container to tar archive
- Choose save location
- Show progress during export
- Validate disk space before export

**Technical Approach:**
- Use Docker SDK `ContainerExport()` API
- File dialog for save location
- Progress bar for export operation

---

### 8. Image Export/Import
**Priority:** LOW | **Effort:** Medium (6-8 hours)

**Description:** Save and load images as tar files without registry.

**Features:**
- Save image to tar file
- Load image from tar file
- Useful for offline transfers
- Progress indication

**Keybindings:**
- `e` - Export image
- `i` - Import image (if not conflicting with pull)

---

## Tier 3: Advanced Features (Future)

These are advanced features that will be considered after Tier 2 completion.

### 1. Docker Scout Security Scanning
**Priority:** MEDIUM | **Effort:** High (10-14 hours)

**Description:** Integrate Docker Scout for vulnerability scanning.

**Features:**
- Scan images for CVEs
- Display vulnerability severity levels
- Show recommended fixes
- Filter by severity
- Export scan results

**Requirements:**
- Docker Scout CLI must be installed
- Requires Docker Scout API access

**Keybindings:**
- `s` - Scan selected image (in Images tab)

**Technical Approach:**
- Execute `docker scout` CLI commands
- Parse JSON output
- Create vulnerability viewer component
- Color-code severity levels (critical, high, medium, low)

---

### 2. Multi-Registry Browser
**Priority:** MEDIUM | **Effort:** High (12-16 hours)

**Description:** Browse multiple container registries beyond Docker Hub.

**Features:**
- Add custom registry endpoints
- Search across multiple registries
- GitHub Container Registry support
- GitLab Container Registry support
- Private registry support
- Registry authentication
- Save registry configurations

**Registries to Support:**
- Docker Hub (already supported)
- GitHub Container Registry (ghcr.io)
- GitLab Container Registry
- Quay.io
- Custom private registries

---

### 3. Docker Events Stream
**Priority:** LOW | **Effort:** Medium (8-10 hours)

**Description:** Live feed of all Docker daemon events.

**Features:**
- Real-time event streaming
- Filter events by type (container, image, volume, network)
- Filter by action (create, start, stop, etc.)
- Searchable event log
- Export event log
- Event timestamps
- Color-coded event types

**Keybindings:**
- Add new tab: `Events`
- `/` - Filter events
- `e` - Export event log
- `c` - Clear event history

---

### 4. Container Templates
**Priority:** MEDIUM | **Effort:** High (10-12 hours)

**Description:** Save and reuse container configurations.

**Features:**
- Save container configuration as template
- Load template when creating new container
- Template library/management
- Share templates (export/import JSON)
- Template versioning
- Variables in templates (e.g., ${PORT})

**Use Cases:**
- Quickly spin up development environments
- Standardize container configurations across team
- Create pre-configured database containers

---

### 5. Bulk Configuration Editor
**Priority:** LOW | **Effort:** High (12-16 hours)

**Description:** Edit multiple containers' settings simultaneously.

**Features:**
- Multi-select containers
- Bulk edit environment variables
- Bulk add labels
- Bulk restart policy changes
- Bulk network attachments
- Preview changes before applying
- Atomic operations (all or nothing)

**Technical Challenges:**
- Requires stopping and recreating containers
- Must preserve unique settings per container
- Complex UI for bulk editing form

---

### 6. Resource Limits Management
**Priority:** MEDIUM | **Effort:** Medium (8-10 hours)

**Description:** Set and modify CPU and memory limits on containers.

**Features:**
- View current resource limits
- Set CPU limits (cores, shares)
- Set memory limits (soft/hard)
- Set swap limits
- Update limits on running containers (where possible)
- Show current resource usage vs limits

**Technical Notes:**
- Some limits require container restart
- Different behavior on different Docker versions
- Validate limits against system resources

---

### 7. Container Health Checks
**Priority:** MEDIUM | **Effort:** Medium (6-8 hours)

**Description:** Configure and monitor container health checks.

**Features:**
- View health check status
- Configure health check commands
- Set check intervals and timeouts
- View health check history
- Notifications on health check failures

---

### 8. Log Analysis Tools
**Priority:** LOW | **Effort:** High (12-14 hours)

**Description:** Advanced log viewing and analysis.

**Features:**
- Search logs with regex
- Filter by log level
- Highlight patterns
- Export filtered logs
- Log statistics (error counts, etc.)
- Compare logs from multiple containers
- Log timestamps and formatting options

---

### 9. Container Backup/Restore
**Priority:** MEDIUM | **Effort:** High (14-18 hours)

**Description:** Backup and restore container data.

**Features:**
- Snapshot container volumes
- Backup container configuration
- Full container backup (config + volumes)
- Restore from backup
- Scheduled backups
- Backup compression
- Incremental backups

---

### 10. Plugin System
**Priority:** LOW | **Effort:** Very High (20+ hours)

**Description:** Extensibility through plugins.

**Features:**
- Plugin architecture
- Custom actions
- Custom views/tabs
- Hook into events
- Plugin discovery and installation
- Plugin marketplace integration

**Technical Challenges:**
- Define stable plugin API
- Security considerations
- Plugin isolation
- Versioning and compatibility

---

## Configuration Features

These are ongoing improvements to configuration management:

### Planned Config Options
```toml
[general]
polling_interval = 500
notification_timeout = 2000
show_all_containers = true
confirm_force_delete = false

[logs]
auto_scroll = true
max_lines = 10000
show_timestamps = true
color_coded = true

[stats]
refresh_interval = 2000
show_sparklines = true

[registry]
default_registry = "docker.io"
registries = [
    { name = "Docker Hub", url = "docker.io" },
    { name = "GHCR", url = "ghcr.io" }
]

[theme]
# Custom color schemes
```

---

## Community Features

Features that benefit from community input:

1. **Custom Keybindings** - Allow users to remap keys
2. **Theme System** - User-defined color themes
3. **Macros/Scripting** - Automate common workflows
4. **Remote Docker Hosts** - Manage containers on remote machines
5. **Multi-Host Dashboard** - View multiple Docker hosts simultaneously

---

## Performance Improvements

Ongoing performance work:

1. **Caching** - Cache container/image lists between refreshes
2. **Lazy Loading** - Load data on-demand for large lists
3. **Pagination** - Handle hundreds/thousands of containers efficiently
4. **Reduced API Calls** - Batch operations and intelligent refresh

---

## Testing & Quality

Quality improvements throughout development:

1. **Integration Tests** - Test against real Docker daemon
2. **UI Tests** - Automated TUI testing
3. **Performance Benchmarks** - Track performance over time
4. **Error Handling** - Comprehensive error recovery
5. **Logging** - Better logging for troubleshooting

---

## Documentation Improvements

1. **Video Tutorials** - Screen recordings of features
2. **Interactive Tutorial** - First-run tutorial mode
3. **Keyboard Shortcuts Cheatsheet** - Quick reference guide
4. **Advanced Usage Guide** - Power user tips
5. **Troubleshooting Guide** - Common issues and solutions

---

## Platform Support

1. **Windows Native** - Better Windows support (currently via WSL)
2. **macOS Optimization** - Native macOS features
3. **ARM Support** - Raspberry Pi and ARM servers
4. **Podman Full Support** - Complete Podman parity with Docker

---

## Timeline Estimates

**Tier 2 Completion:** 10-12 weeks (assuming 1-2 features per week)
**Tier 3 Completion:** 16-20 weeks (more complex features)

Total estimated effort for all planned features: **200-250 hours**

---

## Contributing

Want to help implement these features? Check out:
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [AGENTS.md](AGENTS.md) - Architecture overview
- GitHub Issues - Pick a feature to work on

---

## Feedback

Have ideas for features not listed here? Open a discussion on GitHub!

---

**Last Updated:** February 10, 2026
**Current Phase:** Tier 1 Implementation

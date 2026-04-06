<div align="center">
  <img alt="oci-tui logo" src="./assets/logo.png" />
</div>

<br />

A terminal-based user interface (TUI) for managing container lifecycles, built on [Moby](https://github.com/moby/moby) and powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and excessive coffee consumption.

This repository is currently under heavy development. Expect frequent breaking changes, non-functional components, and incomplete features. This tool is not ready for production use.

## Installation

### Go Install

If you have Go 1.21 or later installed:

```bash
go install github.com/givensuman/containertui@latest
```

This will build and install containertui to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `$PATH`.

### Direct Binary Download

Download the pre-built binary for your platform from the [latest release](https://github.com/givensuman/containertui/releases/latest):

**macOS (Intel):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_darwin_amd64.tar.gz | tar xz
sudo mv containertui /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_darwin_arm64.tar.gz | tar xz
sudo mv containertui /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.tar.gz | tar xz
sudo mv containertui /usr/local/bin/
```

**Windows:**
Download the ZIP file from the releases page and extract `containertui.exe` to a directory in your `PATH`.

### Linux Package Managers

**Debian/Ubuntu (apt):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.deb -o containertui.deb
sudo dpkg -i containertui.deb
```

**Fedora/RHEL (dnf/rpm):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.rpm -o containertui.rpm
sudo rpm -i containertui.rpm
```

**Arch Linux (pacman):**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.tar.gz | tar xz
sudo mv containertui /usr/local/bin/
```

### Docker

Linux-only demo runtime (recommended first command):

```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest
```

This command uses your host Docker daemon through the mounted socket and auto-seeds demo containers/images/volumes/networks.

### Docker Modes (Host Socket and Simulated DinD)

The container image supports two demo modes with no dependency other than Docker.

**1) Host socket mode (default class demo command)**

```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/givensuman/containertui:latest
```

**2) Simulated environment via Docker-in-Docker (DinD)**

```bash
docker run --rm -it --privileged \
  -e CTUI_DEMO_MODE=dind \
  ghcr.io/givensuman/containertui:latest
```

Mode selection variables:

- `CTUI_DEMO_MODE=auto|socket|dind` (default: `auto`)
- `CTUI_DEMO_SEED=1|0` to enable/disable demo data seeding (default: `1`)
- `CTUI_DEMO_CLEANUP_ON_EXIT=1|0` to clean seeded resources on exit (default: `1`)

In `auto`, the container uses host socket mode when `/var/run/docker.sock` is mounted and falls back to DinD otherwise.

## Usage

### Launching Specific Tabs

You can launch containertui directly to a specific tab using subcommands:

```bash
containertui containers    # Launch to containers tab (default)
containertui images        # Launch to images tab
containertui volumes       # Launch to volumes tab
containertui networks      # Launch to networks tab
containertui services      # Launch to services tab
containertui browse        # Launch to browse tab
```

All existing flags continue to work with subcommands:

```bash
containertui images --config /path/to/config --no-nerd-fonts
```

You can also set a default startup tab in your config file:

```yaml
# ~/.config/containertui/config.yaml
startup-tab: images
```

### Linux-First Runtime Notes

This MVP targets Linux + Docker Engine first.

- Docker socket support currently assumes `/var/run/docker.sock` for local Linux usage.
- #TODO: Define Linux socket precedence when both `DOCKER_HOST` and default Unix socket are available.
- #TODO: Define shell fallback order for exec flows (`/bin/sh`, `/bin/bash`, and optional pager behavior such as `less`).
- #TODO: Define clipboard fallback behavior for Linux terminals without `xclip`/`wl-clipboard`.
- #TODO: Confirm rootless Docker support baseline and document tested limits.

### Mouse Support

containertui enables mouse support by default (cell-motion mode). This allows:

- Clicking list items to select them.
- Scrolling list/detail panes with the mouse wheel.
- Interacting with dialogs and overlays using mouse events where supported by the active component.

## Features

### Quick Overview
Navigate seamlessly between containers, images, volumes, networks, and services.

![Overview Demo](./assets/demo-overview.gif)

### Container Management
View, start, stop, inspect, and manage containers with ease.

![Containers Demo](./assets/demo-containers.gif)

### Image Management
Browse local images, view history, and inspect image details.

![Images Demo](./assets/demo-images.gif)

### Volume Management
List and inspect Docker volumes.

![Volumes Demo](./assets/demo-volumes.gif)

### Network Management
View and inspect Docker networks.

![Networks Demo](./assets/demo-networks.gif)

### Services View
Monitor Docker Compose services and container stacks.

![Services Demo](./assets/demo-services.gif)

<div align="center">
  <img alt="oci-tui logo" src="./assets/logo.png" />
</div>

<br />

A terminal-based user interface (TUI) for managing container lifecycles, built on [Moby](https://github.com/moby/moby) and powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and excessive coffee consumption.

This repository is currently under heavy development. Expect frequent breaking changes, non-functional components, and incomplete features. This tool is not ready for production use.

## Demo

You can run the following command to see containertui in action:

```bash
docker run --rm -it --privileged \
  ghcr.io/givensuman/containertui-demo:latest
```

Demo mode seeds a curated environment with images, networks, volumes, containers in different states, and a small Compose-backed service. Seeded resources are cleaned up automatically when you exit.

`--privileged` is required for the demo image because it runs Docker-in-Docker and needs kernel/network capabilities to create the default bridge network.

## Installation

### Go Install

If you have Go 1.21 or later installed:

```bash
go install github.com/givensuman/containertui@latest
```

This will build and install `containertui` to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `$PATH`.

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

**Docker:**
```bash
docker run --rm -it \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --group-add "$(stat -c '%g' /var/run/docker.sock)" \
  ghcr.io/givensuman/containertui:latest
```

This image talks to your host Docker daemon through the mounted socket (it does not start Docker-in-Docker).

For rootless Docker, use your rootless socket path and map it to `/var/run/docker.sock` inside the container:

```bash
docker run --rm -it \
  -v "$XDG_RUNTIME_DIR/docker.sock:/var/run/docker.sock" \
  ghcr.io/givensuman/containertui:latest
```

If `XDG_RUNTIME_DIR` is unset in your shell, try:

```bash
docker run --rm -it \
  -v "/run/user/$(id -u)/docker.sock:/var/run/docker.sock" \
  ghcr.io/givensuman/containertui:latest
```


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

# Installation Guide

containertui is available through multiple installation methods. Choose the one that best fits your workflow.

## Quick Start

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

Run containertui in a Docker container with your Docker socket mounted:

```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest
```

---

## Installation Methods

### Go Install

**Requirements:**
- Go 1.21 or later
- Git

**Installation:**
```bash
go install github.com/givensuman/containertui@latest
```

**Verify installation:**
```bash
containertui --version
```

This method compiles containertui from source on your system. The binary is installed to `$GOPATH/bin` (default: `~/go/bin`).

**Add to PATH:**
If containertui is not in your PATH, add it:

```bash
# Bash
echo 'export PATH=$PATH:~/go/bin' >> ~/.bashrc
source ~/.bashrc

# Zsh
echo 'export PATH=$PATH:~/go/bin' >> ~/.zshrc
source ~/.zshrc

# Fish
set -Ua fish_user_paths ~/go/bin
```

### Binary Download

Pre-built binaries are available for all major platforms.

**Supported platforms:**
- Linux: amd64, arm64, arm, 386
- macOS: amd64 (Intel), arm64 (Apple Silicon)
- Windows: amd64, arm64, arm, 386
- FreeBSD: amd64, arm64, arm, 386

**Download and install:**

1. Visit [releases page](https://github.com/givensuman/containertui/releases)
2. Download the archive for your platform
3. Extract the binary
4. Move to a directory in your PATH

**Example for Linux x86_64:**
```bash
# Download
wget https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.tar.gz

# Extract
tar xzf containertui_X.Y.Z_linux_amd64.tar.gz

# Install to system path
sudo mv containertui /usr/local/bin/

# Verify
containertui --version
```

**Example for macOS (Intel):**
```bash
# Download
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_darwin_amd64.tar.gz -o containertui.tar.gz

# Extract
tar xzf containertui.tar.gz

# Install
sudo mv containertui /usr/local/bin/

# Verify
containertui --version
```

**Example for Windows (PowerShell):**
```powershell
# Download (replace X.Y.Z with version)
$url = "https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile containertui.zip

# Extract
Expand-Archive -Path containertui.zip -DestinationPath .

# Add to a directory in PATH or use directly:
.\containertui.exe
```

### Debian/Ubuntu (apt)

**Direct deb package installation:**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.deb -o containertui.deb
sudo dpkg -i containertui.deb
```

**Verify installation:**
```bash
containertui --version
```

The package installs the binary to `/usr/local/bin/containertui` and is automatically available in your PATH.

### Fedora/RHEL (dnf/rpm)

**Direct rpm package installation:**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.rpm -o containertui.rpm
sudo rpm -i containertui.rpm
```

**Using dnf:**
```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.rpm -o containertui.rpm
sudo dnf install ./containertui.rpm
```

**Verify installation:**
```bash
containertui --version
```

### Arch Linux (pacman)

Download and extract the tar.gz archive:

```bash
curl -L https://github.com/givensuman/containertui/releases/download/vX.Y.Z/containertui_X.Y.Z_linux_amd64.tar.gz -o containertui.tar.gz
tar xzf containertui.tar.gz
sudo mv containertui /usr/local/bin/
```

**Verify installation:**
```bash
containertui --version
```

### Docker

Run containertui as a Docker container with your Docker socket mounted.

**Basic usage:**
```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest
```

**Docker socket mounting:**

containertui requires access to the Docker daemon through the Docker socket. The socket location depends on your system:

- **Linux:** `/var/run/docker.sock`
- **Docker Desktop (macOS/Windows):** `/var/run/docker.sock` or `~/Library/Containers/com.docker.docker/Data/docker.raw.sock` (macOS Intel) / `~/.docker/desktop/docker.sock` (newer Docker Desktop)

**On macOS with Docker Desktop:**
```bash
docker run --rm -it -v ~/.docker/desktop/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest
```

**Persistent configuration:**

Create an alias for easier access:

```bash
# Bash/Zsh
alias containertui='docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest'

# Then use:
containertui
```

Add this to your shell configuration file (`.bashrc`, `.zshrc`, etc.) to persist across sessions.

---

## Verification

To verify containertui is correctly installed:

```bash
# Check version
containertui --version

# Check help
containertui --help

# Test Docker connection (launch the UI)
containertui
```

The last command will start the containertui interface. If Docker is properly configured, you should see the container management dashboard. Press `q` to exit.

---

## Prerequisites

### Docker Daemon

containertui requires a running Docker daemon to manage containers. Ensure Docker is installed and running:

```bash
# Check Docker is running
docker --version
docker ps
```

### Docker Socket Access

containertui communicates with Docker through the Docker socket. You need appropriate permissions:

**Linux:**
```bash
# Add your user to the docker group (requires logout/login)
sudo usermod -aG docker $USER
newgrp docker

# Verify
docker ps
```

**macOS/Windows:**
Docker Desktop handles socket permissions automatically.

### Terminal Requirements

- **Terminal emulator** supporting VT100/ANSI escape sequences
- **Minimum size:** 80 columns × 24 rows (recommended: 120+ columns for best experience)
- **Color support:** 256-color or true color recommended

containertui works in:
- Terminal.app (macOS)
- iTerm2 (macOS)
- GNOME Terminal, Konsole, xterm (Linux)
- Windows Terminal, ConEmu, Git Bash (Windows)

---

## Troubleshooting

### Docker Socket Permission Denied

**Error:** `permission denied while trying to connect to Docker daemon socket`

**Solution:**

1. Verify Docker is running:
   ```bash
   docker ps
   ```

2. Add your user to the docker group (Linux):
   ```bash
   sudo usermod -aG docker $USER
   newgrp docker
   ```

3. Verify permissions:
   ```bash
   ls -l /var/run/docker.sock
   ```

### Command Not Found

**Error:** `containertui: command not found`

**Solution:**

1. Verify installation location:
   ```bash
   which containertui
   ```

2. If not found, add the installation directory to PATH:
   ```bash
   # For Go install
   export PATH=$PATH:~/go/bin
   
   # For manual binary installation
   export PATH=$PATH:/usr/local/bin
   ```

3. Reload shell configuration:
   ```bash
   source ~/.bashrc    # or ~/.zshrc for zsh
   ```

### Docker Socket Not Found (macOS)

**Error:** `cannot connect to Docker daemon`

**Solution:**

1. Verify Docker Desktop is running (check menu bar)

2. Find the correct socket path:
   ```bash
   # Modern Docker Desktop
   ls -la ~/.docker/desktop/docker.sock
   
   # Older Docker Desktop
   ls -la ~/Library/Containers/com.docker.docker/Data/docker.raw.sock
   ```

3. Use the correct path in Docker run command:
   ```bash
   docker run --rm -it -v ~/.docker/desktop/docker.sock:/var/run/docker.sock ghcr.io/givensuman/containertui:latest
   ```

### Windows Docker Desktop Issues

**Problem:** containertui won't run or can't connect to Docker

**Solution:**

1. Ensure Docker Desktop is running (system tray icon)

2. If using WSL 2 backend:
   ```powershell
   # Verify WSL is configured
   wsl --list --verbose
   
   # Docker socket should be available in WSL
   # Install containertui in WSL terminal for best experience
   ```

3. If using Hyper-V backend, ensure it's enabled:
   ```powershell
   # Run as Administrator
   systeminfo | findstr "Hyper-V"
   ```

### Display/Rendering Issues

**Problem:** Text appears corrupted or colors are wrong

**Solution:**

1. Verify terminal supports 256 colors:
   ```bash
   echo $TERM
   ```

2. Try setting explicit color support:
   ```bash
   TERM=xterm-256color containertui
   ```

3. Update terminal emulator to latest version

4. Ensure window is at least 80×24 characters

---

## Building from Source

To build containertui from source for development:

**Requirements:**
- Go 1.21 or later
- Git

**Clone and build:**
```bash
# Clone the repository
git clone https://github.com/givensuman/containertui.git
cd containertui

# Build
go build -o containertui ./cmd

# Run
./containertui
```

**Development build with version info:**
```bash
go build -ldflags="-X github.com/givensuman/containertui/internal/build.Version=dev" -o containertui ./cmd
```

**Run tests:**
```bash
go test ./...
```

For more information on contributing, see [CONTRIBUTING.md](../CONTRIBUTING.md).

---

## Additional Help

- **README:** See the [main README](../README.md) for features and overview
- **Issues:** Report bugs at [GitHub Issues](https://github.com/givensuman/containertui/issues)
- **Discussions:** Join discussions at [GitHub Discussions](https://github.com/givensuman/containertui/discussions)
- **Docker Hub:** [ghcr.io/givensuman/containertui](https://ghcr.io/givensuman/containertui)
- **GitHub Releases:** [Latest releases](https://github.com/givensuman/containertui/releases)

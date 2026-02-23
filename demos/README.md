# VHS Demo Generation

This directory contains VHS tape files and scripts for generating animated GIF demos of containertui functionality.

## Overview

The demo system uses [VHS](https://github.com/charmbracelet/vhs) to record terminal sessions and generate GIFs that showcase different features of containertui. These GIFs are automatically generated via GitHub Actions and embedded in the README.

## Files

- **`*.tape`** - VHS tape files defining the demo recordings
  - `overview.tape` - Navigation through all tabs
  - `containers.tape` - Container management operations
  - `images.tape` - Image browsing and inspection
  - `volumes.tape` - Volume management
  - `networks.tape` - Network viewing and inspection
  - `services.tape` - Services/compose view
  
- **`setup.sh`** - Script to populate Docker with test containers, images, volumes, and networks

- **`cleanup.sh`** - Script to remove test Docker resources

## Local Development

### Prerequisites

- [VHS](https://github.com/charmbracelet/vhs) installed (`brew install vhs` on macOS)
- Docker running locally
- Go development environment

### Generate All Demos

```bash
just demo
```

This will:
1. Run `setup.sh` to create test Docker resources
2. Generate all GIF demos from the tape files
3. Run `cleanup.sh` to remove test resources
4. Output GIFs to `../assets/`

### Generate a Single Demo

```bash
just demo-single <tape-name>
```

Example:
```bash
just demo-single containers
```

### Manual Setup/Cleanup

Set up test environment:
```bash
just demo-setup
# or
./demos/setup.sh
```

Clean up test environment:
```bash
just demo-cleanup
# or
./demos/cleanup.sh
```

## CI/CD

The `.github/workflows/vhs-demos.yml` workflow automatically:

1. Triggers on every push to `main` branch
2. Sets up a Docker-in-Docker environment
3. Installs VHS and dependencies
4. Runs `setup.sh` to populate test data
5. Generates all demo GIFs
6. Commits updated GIFs back to the repository (with `[skip ci]` to prevent loops)

The workflow uses the `[skip ci]` commit message to prevent triggering another CI run when pushing the generated GIFs.

## Modifying Demos

### Editing Tape Files

Each `.tape` file uses VHS syntax:

```tape
Output ../assets/demo-example.gif
Set Theme "Catppuccin Mocha"
Set Width 1200
Set Height 600
Set FontSize 14
Set Padding 10
Set TypingSpeed 50ms

Type "go run ./cmd"
Enter
Sleep 3s

# Your demo actions here
Type "j"  # Navigate down
Sleep 500ms

Type "q"  # Quit
Sleep 500ms
```

Key VHS commands:
- `Type "<text>"` - Type text
- `Enter` - Press Enter key
- `Sleep <duration>` - Wait (e.g., `500ms`, `2s`)
- `Set <setting> <value>` - Configure appearance

### Keyboard Navigation

Make sure tape files use the correct keybindings from containertui:

- `1-5` - Switch between tabs (containers, images, volumes, networks, services)
- `j/k` - Navigate down/up in lists
- `i` - Inspect selected item
- `l` - View logs (containers only)
- `h` - View history (images only)
- `Space` - Toggle container state
- `q` - Quit/back

### Test Environment

Edit `setup.sh` to modify the test containers, images, volumes, or networks created for demos. This is useful if you need specific scenarios (e.g., more containers, different image types, etc.).

## Tips

- Keep demos short (20-30 seconds) for quick overview
- Add strategic `Sleep` commands to let users see what's happening
- Use consistent timing across demos (e.g., `500ms` for navigation, `2s` for viewing content)
- Test locally before pushing to ensure demos work correctly
- GIF file sizes: aim for < 5MB per GIF (adjust resolution/duration if needed)

## Troubleshooting

**VHS fails with "command not found":**
- Ensure VHS is installed: `vhs --version`
- Install: `brew install vhs` (macOS) or download from [releases](https://github.com/charmbracelet/vhs/releases)

**Demos show wrong content:**
- Make sure Docker test environment is set up: `./demos/setup.sh`
- Verify containers are running: `docker ps`

**Timing issues (too fast/slow):**
- Adjust `Sleep` durations in tape files
- Modify `Set TypingSpeed` if typing appears too fast/slow

**Docker daemon not available:**
- Ensure Docker is running: `docker info`
- Check DOCKER_HOST environment variable if using remote Docker

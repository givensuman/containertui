# VHS Demo System - Quick Start

## What Was Implemented

A complete VHS-based GIF generation system for creating animated demos of containertui functionality.

## Structure

```
containertui/
├── demos/
│   ├── README.md                 # Comprehensive documentation
│   ├── setup.sh                  # Populate Docker with test data
│   ├── cleanup.sh                # Remove test Docker resources
│   ├── overview.tape             # Navigation demo
│   ├── containers.tape           # Container management demo
│   ├── images.tape               # Image browsing demo
│   ├── volumes.tape              # Volume management demo
│   ├── networks.tape             # Network viewing demo
│   └── services.tape             # Services/compose demo
├── assets/
│   └── demo-*.gif                # Generated GIFs (created by workflow)
├── .github/workflows/
│   └── vhs-demos.yml             # GitHub Actions workflow
└── Justfile                      # Added demo commands
```

## Quick Commands

### Generate All Demos Locally
```bash
just demo
```

### Generate Single Demo
```bash
just demo-single containers
```

### Manual Control
```bash
# Set up test environment
just demo-setup

# Generate a specific tape manually
cd demos && vhs containers.tape

# Clean up
just demo-cleanup
```

## CI/CD Integration

The GitHub Actions workflow (`.github/workflows/vhs-demos.yml`) will:

1. **Trigger**: On every push to `main` branch (or manual dispatch)
2. **Environment**: Ubuntu with Docker-in-Docker service
3. **Process**:
   - Install VHS and dependencies
   - Set up Go environment
   - Run `demos/setup.sh` to create test containers
   - Generate all 6 demo GIFs
   - Commit GIFs back to repo (with `[skip ci]`)
4. **Output**: Updated GIFs in `assets/` directory

## Testing Locally Before Push

Before pushing to trigger CI, test locally:

```bash
# 1. Ensure VHS is installed
vhs --version

# 2. Generate demos
just demo

# 3. Check generated files
ls -lh assets/demo-*.gif

# 4. View a GIF (macOS)
open assets/demo-overview.gif

# 5. View a GIF (Linux)
xdg-open assets/demo-overview.gif
```

## Customizing Demos

### Modify Tape Files

Edit `demos/*.tape` files to change:
- Recording length (adjust `Sleep` durations)
- Actions performed (add/remove `Type` commands)
- Visual appearance (`Set Theme`, `Set Width`, etc.)

### Modify Test Environment

Edit `demos/setup.sh` to:
- Create different test containers
- Pull different images
- Add more volumes/networks
- Set up specific scenarios

### Keybindings Used in Tapes

The tapes use these containertui keybindings:
- `1-5`: Switch tabs (containers, images, volumes, networks, services)
- `j/k`: Navigate lists
- `i`: Inspect
- `l`: Logs (containers)
- `h`: History (images)
- `Space`: Toggle container state
- `q`: Quit/back

## README Integration

The README.md now includes a "Features" section with all demo GIFs embedded:

```markdown
## Features

### Quick Overview
![Overview Demo](./assets/demo-overview.gif)

### Container Management
![Containers Demo](./assets/demo-containers.gif)

... etc
```

## Installation Requirements

### For Local Development
- **VHS**: `brew install vhs` (macOS) or see [releases](https://github.com/charmbracelet/vhs/releases)
- **Docker**: Running Docker daemon
- **Go**: For building containertui

### For CI/CD
- No manual setup needed
- GitHub Actions handles all dependencies
- Uses Docker-in-Docker for container management

## Expected GIF Specifications

- **Resolution**: 1200x600 pixels
- **Theme**: Catppuccin Mocha (matches typical terminal aesthetics)
- **Duration**: 20-30 seconds per demo
- **Font Size**: 14px
- **File Size**: Target < 5MB per GIF

## Workflow Details

### When GIFs Are Generated
- ✅ Every push to `main` branch
- ✅ Manual workflow dispatch
- ❌ Pull requests (not configured)
- ❌ Other branches (not configured)

### How GIFs Are Committed
- GitHub Actions bot commits with message: `"Update demo GIFs [skip ci]"`
- The `[skip ci]` tag prevents infinite loops
- Only commits if GIFs actually changed

### Potential Issues & Solutions

**Issue**: VHS fails in CI with "command not found"
- **Solution**: Workflow installs VHS from GitHub releases

**Issue**: Docker containers not available
- **Solution**: Uses Docker-in-Docker service with `--privileged` flag

**Issue**: Infinite CI loops
- **Solution**: Commit message includes `[skip ci]`

**Issue**: GIFs too large
- **Solution**: Adjust resolution/duration in tape files

## Next Steps

1. **Test locally**: Run `just demo` to generate GIFs locally
2. **Review output**: Check that GIFs look correct
3. **Push to main**: Workflow will trigger automatically
4. **Verify CI**: Check GitHub Actions for successful GIF generation
5. **View README**: Confirm GIFs appear in README

## Files Modified/Created

### Created
- `demos/*.tape` (6 files)
- `demos/setup.sh`
- `demos/cleanup.sh`
- `demos/README.md`
- `.github/workflows/vhs-demos.yml`
- `demos/QUICKSTART.md` (this file)

### Modified
- `Justfile` (added 4 demo commands)
- `README.md` (added Features section with GIF embeds)

## Resources

- [VHS Documentation](https://github.com/charmbracelet/vhs)
- [VHS Examples](https://github.com/charmbracelet/vhs/tree/main/examples)
- [GitHub Actions: Docker-in-Docker](https://docs.github.com/en/actions/using-containerized-services/about-service-containers)

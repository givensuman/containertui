#!/bin/bash
# Demo entrypoint script - initializes Docker-in-Docker environment with sample data
# and launches containertui for interactive demonstration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Trap signals for graceful cleanup
cleanup() {
	echo -e "\n${YELLOW}Cleaning up demo environment...${NC}"
	/demo/cleanup.sh || true
	exit 0
}

trap cleanup SIGINT SIGTERM EXIT

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  ContainerTUI Demo Environment${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

# Configure containertui with nerd fonts enabled
CONFIG_DIR="${HOME}/.config/containertui"
mkdir -p "${CONFIG_DIR}"
cat >"${CONFIG_DIR}/config.yaml" <<'EOF'
# Enable nerd fonts for proper icon rendering
no-nerd-fonts: false
EOF

echo -e "${YELLOW}Configuring nerd fonts...${NC}"
echo -e "${GREEN}✓ Nerd fonts enabled${NC}"
echo ""

# Start Docker daemon in the background
echo -e "${YELLOW}Starting Docker daemon...${NC}"
dockerd >/tmp/dockerd.log 2>&1 &
DOCKER_PID=$!

# Wait for Docker socket to be ready (more reliable than docker info command)
echo -e "${YELLOW}Waiting for Docker daemon...${NC}"
timeout 30 sh -c 'until [ -S /var/run/docker.sock ]; do sleep 1; done' || {
	echo -e "${RED}Docker daemon failed to start after 30 seconds${NC}"
	cat /tmp/dockerd.log
	exit 1
}
echo -e "${GREEN}✓ Docker daemon is ready${NC}"
echo ""

# Run setup script to populate demo environment
echo -e "${YELLOW}Setting up demo environment with sample containers, images, volumes, and networks...${NC}"
/demo/setup.sh
echo ""

# Display banner with instructions
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  Demo Ready!${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "You can now explore the demo environment:"
echo "  • Navigate between tabs: 1-5 (Containers, Images, Volumes, Networks, Services)"
echo "  • Inspect items: Press 'i' on selected item"
echo "  • View logs: Press 'l' on a running container"
echo "  • Quit: Press 'q' at any time"
echo ""
echo -e "${YELLOW}Launching ContainerTUI...${NC}"
echo ""

# Launch containertui
containertui

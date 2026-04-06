#!/bin/bash
# ContainerTUI runtime entrypoint supporting host-socket and DinD demo modes.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

CTUI_DEMO_MODE="${CTUI_DEMO_MODE:-auto}"
CTUI_DEMO_SEED="${CTUI_DEMO_SEED:-1}"
CTUI_DEMO_CLEANUP_ON_EXIT="${CTUI_DEMO_CLEANUP_ON_EXIT:-1}"
DOCKERD_LOG_FILE="/tmp/dockerd.log"
DOCKERD_STARTED=0
CLEANED_UP=0

print_banner() {
	echo -e "${GREEN}================================${NC}"
	echo -e "${GREEN}  ContainerTUI Demo Environment${NC}"
	echo -e "${GREEN}================================${NC}"
	echo ""
}

configure_user_settings() {
	local config_dir
	config_dir="${HOME}/.config/containertui"
	mkdir -p "${config_dir}"
	cat >"${config_dir}/config.yaml" <<'EOF'
no-nerd-fonts: false
EOF
}

wait_for_docker() {
	timeout 30 sh -c 'until docker info >/dev/null 2>&1; do sleep 1; done' || {
		echo -e "${RED}Docker daemon did not become ready within 30 seconds${NC}"
		if [ -f "${DOCKERD_LOG_FILE}" ]; then
			echo -e "${YELLOW}dockerd logs:${NC}"
			cat "${DOCKERD_LOG_FILE}"
		fi
		exit 1
	}
}

start_dind_if_needed() {
	echo -e "${YELLOW}Starting Docker-in-Docker daemon...${NC}"
	dockerd >"${DOCKERD_LOG_FILE}" 2>&1 &
	DOCKERD_STARTED=1
	wait_for_docker
	echo -e "${GREEN}✓ Docker-in-Docker daemon is ready${NC}"
}

has_host_socket() {
	[ -S /var/run/docker.sock ]
}

verify_host_socket_access() {
	docker info >/dev/null 2>&1
}

resolve_mode() {
	case "${CTUI_DEMO_MODE}" in
	socket)
		echo "socket"
		return
		;;
	dind)
		echo "dind"
		return
		;;
	auto)
		if has_host_socket; then
			echo "socket"
		else
			echo "dind"
		fi
		return
		;;
	*)
		echo -e "${RED}Invalid CTUI_DEMO_MODE: ${CTUI_DEMO_MODE}${NC}"
		echo "Valid values: auto, socket, dind"
		exit 1
		;;
	esac
}

cleanup() {
	if [ "${CLEANED_UP}" = "1" ]; then
		return
	fi
	CLEANED_UP=1

	if [ "${CTUI_DEMO_CLEANUP_ON_EXIT}" = "1" ]; then
		echo -e "\n${YELLOW}Cleaning up demo resources...${NC}"
		/demo/cleanup.sh || true
	fi
	if [ "${DOCKERD_STARTED}" = "1" ]; then
		pkill dockerd >/dev/null 2>&1 || true
	fi
}

trap cleanup EXIT
trap 'exit 130' SIGINT
trap 'exit 143' SIGTERM

print_banner
configure_user_settings

MODE="$(resolve_mode)"
if [ "${MODE}" = "socket" ]; then
	echo -e "${GREEN}Mode:${NC} Host Docker socket"
	if ! has_host_socket; then
		echo -e "${RED}Host Docker socket mode requested but /var/run/docker.sock is unavailable.${NC}"
		echo "Run with -v /var/run/docker.sock:/var/run/docker.sock or set CTUI_DEMO_MODE=dind with --privileged."
		exit 1
	fi
	if ! verify_host_socket_access; then
		echo -e "${RED}Host Docker socket is mounted but not accessible.${NC}"
		echo "Ensure your user can access Docker on the host and try again."
		exit 1
	fi
else
	echo -e "${GREEN}Mode:${NC} Docker-in-Docker"
	start_dind_if_needed
fi

if [ "${CTUI_DEMO_SEED}" = "1" ]; then
	echo -e "${YELLOW}Seeding demo resources...${NC}"
	/demo/cleanup.sh || true
	/demo/setup.sh
	if [ "${MODE}" = "socket" ]; then
		echo -e "${YELLOW}Note:${NC} Demo resources are created on your host Docker daemon."
	fi
else
	echo -e "${YELLOW}Skipping demo seed (CTUI_DEMO_SEED=${CTUI_DEMO_SEED}).${NC}"
fi

echo ""
echo "You can now explore the demo environment:"
echo "  - Navigate tabs: 1-6"
echo "  - Inspect item: i"
echo "  - View logs: l"
echo "  - Quit: q"
echo ""
echo -e "${YELLOW}Launching ContainerTUI...${NC}"

containertui "$@"

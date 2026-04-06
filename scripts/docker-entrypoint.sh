#!/usr/bin/env sh

set -eu

SOCKET_PATH="${DOCKER_SOCKET_PATH:-/var/run/docker.sock}"
DOCKER_ENDPOINT="${DOCKER_HOST:-unix://${SOCKET_PATH}}"

print_help_hints() {
	echo "hint (rootful): -v /var/run/docker.sock:/var/run/docker.sock" >&2
	echo "hint (rootless): -v \"\$XDG_RUNTIME_DIR/docker.sock:/var/run/docker.sock\" -e DOCKER_HOST=unix:///var/run/docker.sock" >&2
	echo "hint: for permission errors also add --group-add \"\$(stat -c '%g' /var/run/docker.sock)\"" >&2
}

case "${DOCKER_ENDPOINT}" in
unix://*)
	SOCKET_PATH="${DOCKER_ENDPOINT#unix://}"
	if [ -e "${SOCKET_PATH}" ] && [ ! -S "${SOCKET_PATH}" ]; then
		echo "error: ${SOCKET_PATH} exists but is not a unix socket" >&2
		print_help_hints
		exit 1
	fi
	if [ ! -S "${SOCKET_PATH}" ]; then
		echo "error: Docker socket not found at ${SOCKET_PATH}" >&2
		print_help_hints
		exit 1
	fi
	;;
esac

export DOCKER_HOST="${DOCKER_ENDPOINT}"

if ! DOCKER_ERROR_OUTPUT="$(docker version 2>&1 >/dev/null)"; then
	echo "error: cannot access Docker daemon using ${DOCKER_ENDPOINT}" >&2
	echo "docker version error: ${DOCKER_ERROR_OUTPUT}" >&2
	print_help_hints
	exit 1
fi

exec /usr/local/bin/containertui "$@"

#!/usr/bin/env sh

set -eu

SOCKET_PATH="${DOCKER_SOCKET_PATH:-/var/run/docker.sock}"

print_help_hints() {
	echo "hint: mount Docker socket: -v ${SOCKET_PATH}:${SOCKET_PATH}" >&2
	echo "hint: if you get permission errors, add host socket group: --group-add \"\$(stat -c '%g' ${SOCKET_PATH})\"" >&2
}

if [ ! -S "${SOCKET_PATH}" ]; then
	echo "error: Docker socket not found at ${SOCKET_PATH}" >&2
	print_help_hints
	exit 1
fi

if ! docker version >/dev/null 2>&1; then
	echo "error: cannot access Docker daemon via ${SOCKET_PATH}" >&2
	print_help_hints
	exit 1
fi

exec /usr/local/bin/containertui "$@"

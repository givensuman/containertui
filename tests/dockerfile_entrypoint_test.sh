#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="containertui:entrypoint-test"

echo "=== Dockerfile Entrypoint Test ==="

echo -n "Building production image... "
docker build -f "$PROJECT_ROOT/Dockerfile" -t "$IMAGE_NAME" "$PROJECT_ROOT" >/dev/null
echo "PASS"

ENTRYPOINT_JSON="$(docker image inspect "$IMAGE_NAME" --format '{{json .Config.Entrypoint}}')"

echo -n "Entrypoint uses production wrapper script... "
if [[ "$ENTRYPOINT_JSON" == '["/usr/local/bin/docker-entrypoint.sh"]' ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Entrypoint was: $ENTRYPOINT_JSON"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

echo -n "Wrapper script exists in image... "
if docker run --rm --entrypoint=test "$IMAGE_NAME" -f /usr/local/bin/docker-entrypoint.sh >/dev/null 2>&1; then
	echo "PASS"
else
	echo "FAIL"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

echo -n "Missing socket prints actionable hint... "
OUTPUT="$(docker run --rm "$IMAGE_NAME" 2>&1 || true)"
if [[ "$OUTPUT" == *"/var/run/docker.sock"* ]] && [[ "$OUTPUT" == *"--group-add"* ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Output was: $OUTPUT"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

echo -n "Directory bind on socket path gives rootless hint... "
TMPDIR_FOR_SOCKET_TEST="$(mktemp -d)"
OUTPUT="$(docker run --rm -v "${TMPDIR_FOR_SOCKET_TEST}:/var/run/docker.sock" "$IMAGE_NAME" 2>&1 || true)"
rm -rf "${TMPDIR_FOR_SOCKET_TEST}"
if [[ "$OUTPUT" == *"not a unix socket"* ]] && [[ "$OUTPUT" == *"rootless"* ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Output was: $OUTPUT"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

echo -n "Entrypoint is not demo runtime... "
if [[ "$ENTRYPOINT_JSON" == *"/demo/demo-entrypoint.sh"* ]]; then
	echo "FAIL"
	echo "Entrypoint unexpectedly points to demo runtime: $ENTRYPOINT_JSON"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi
echo "PASS"

docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true

echo "dockerfile_entrypoint_test.sh: PASS"

#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="containertui:demo-entrypoint-test"

echo "=== Dockerfile Demo Runtime Test ==="

echo -n "Building demo image... "
docker build -f "$PROJECT_ROOT/Dockerfile.demo" -t "$IMAGE_NAME" "$PROJECT_ROOT" >/dev/null
echo "PASS"

ENTRYPOINT_JSON="$(docker image inspect "$IMAGE_NAME" --format '{{json .Config.Entrypoint}}')"

echo -n "Demo image uses demo entrypoint... "
if [[ "$ENTRYPOINT_JSON" == '["/demo/demo-entrypoint.sh"]' ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Entrypoint was: $ENTRYPOINT_JSON"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

echo -n "Unprivileged run fails with privileged hint... "
OUTPUT="$(docker run --rm "$IMAGE_NAME" 2>&1 || true)"
if [[ "$OUTPUT" == *"operation not permitted"* ]] && [[ "$OUTPUT" == *"--privileged"* ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Output was: $OUTPUT"
	docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true
	exit 1
fi

docker image rm "$IMAGE_NAME" >/dev/null 2>&1 || true

echo "dockerfile_demo_test.sh: PASS"

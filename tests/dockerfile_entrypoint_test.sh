#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="containertui:entrypoint-test"

echo "=== Dockerfile Entrypoint Test ==="

echo -n "Building production image... "
docker build -f "$PROJECT_ROOT/Dockerfile" -t "$IMAGE_NAME" "$PROJECT_ROOT" >/dev/null
echo "PASS"

ENTRYPOINT_JSON="$(docker image inspect "$IMAGE_NAME" --format '{{json .Config.Entrypoint}}')"

echo -n "Entrypoint uses containertui binary... "
if [[ "$ENTRYPOINT_JSON" == '["containertui"]' ]] || [[ "$ENTRYPOINT_JSON" == '["/usr/local/bin/containertui"]' ]]; then
	echo "PASS"
else
	echo "FAIL"
	echo "Entrypoint was: $ENTRYPOINT_JSON"
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

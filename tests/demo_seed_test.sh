#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
NAMESPACE="containertui-demo-test"
LOG_FILE="$(mktemp)"
FAKE_BIN_DIR="$(mktemp -d)"

cleanup() {
	rm -f "$LOG_FILE"
	rm -rf "$FAKE_BIN_DIR"
}
trap cleanup EXIT

cat >"$FAKE_BIN_DIR/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DOCKER_FAKE_LOG:-}" ]]; then
  echo "DOCKER_FAKE_LOG is required" >&2
  exit 1
fi

echo "$*" >>"$DOCKER_FAKE_LOG"

if [[ "$1" == "compose" && "${2:-}" == "version" ]]; then
  exit 0
fi

if [[ "$1" == "images" ]]; then
  printf 'alpine:latest\nnginx:alpine\nbusybox:latest\n'
  exit 0
fi

if [[ "$1" == "ps" ]]; then
  printf 'demo-container\n'
  exit 0
fi

if [[ "$1" == "volume" && "${2:-}" == "ls" ]]; then
  printf 'demo-volume\n'
  exit 0
fi

if [[ "$1" == "network" && "${2:-}" == "ls" ]]; then
  printf '%s-network-1\n' "${CTUI_DEMO_NAMESPACE:-containertui-demo}"
  exit 0
fi

exit 0
EOF

chmod +x "$FAKE_BIN_DIR/docker"

assert_log_contains() {
	local expected="$1"
	if ! grep -Fq "$expected" "$LOG_FILE"; then
		echo "Missing expected docker command:" >&2
		echo "  $expected" >&2
		echo "Captured docker calls:" >&2
		cat "$LOG_FILE" >&2
		exit 1
	fi
}

echo "Running demo setup script with fake docker..."
PATH="$FAKE_BIN_DIR:$PATH" \
	DOCKER_FAKE_LOG="$LOG_FILE" \
	CTUI_DEMO_NAMESPACE="$NAMESPACE" \
	bash "$PROJECT_ROOT/demos/setup.sh" >/dev/null

assert_log_contains "image tag alpine:latest ${NAMESPACE}/demo-base:1.0"
assert_log_contains "image tag nginx:alpine ${NAMESPACE}/demo-web:1.0"
assert_log_contains "image tag busybox:latest ${NAMESPACE}/demo-tooling:1.0"

assert_log_contains "volume create --label ctui.demo=1 --label ctui.namespace=${NAMESPACE} ${NAMESPACE}-vol-1"
assert_log_contains "network create --label ctui.demo=1 --label ctui.namespace=${NAMESPACE} ${NAMESPACE}-network-1"
assert_log_contains "run --rm --label ctui.demo=1 --label ctui.namespace=${NAMESPACE} -v ${NAMESPACE}-data:/seed-data alpine sh -c"

echo "Running demo cleanup script with fake docker..."
PATH="$FAKE_BIN_DIR:$PATH" \
	DOCKER_FAKE_LOG="$LOG_FILE" \
	CTUI_DEMO_NAMESPACE="$NAMESPACE" \
	bash "$PROJECT_ROOT/demos/cleanup.sh" >/dev/null

assert_log_contains "image rm ${NAMESPACE}/demo-base:1.0 ${NAMESPACE}/demo-web:1.0 ${NAMESPACE}/demo-tooling:1.0"

echo "demo_seed_test.sh: PASS"

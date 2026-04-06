#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENTRYPOINT="$PROJECT_ROOT/demos/demo-entrypoint.sh"

assert_contains() {
	local needle="$1"
	if ! grep -Fq "$needle" "$ENTRYPOINT"; then
		echo "Expected to find: $needle" >&2
		exit 1
	fi
}

assert_not_contains() {
	local needle="$1"
	if grep -Fq "$needle" "$ENTRYPOINT"; then
		echo "Did not expect to find: $needle" >&2
		exit 1
	fi
}

assert_contains 'if [ "${1:-}" = "--demo" ]; then'
assert_contains 'Host Docker socket'
assert_contains 'Ephemeral demo (Docker-in-Docker)'
assert_not_contains 'CTUI_DEMO_MODE'
assert_not_contains 'CTUI_DEMO_SEED'
assert_not_contains 'CTUI_DEMO_CLEANUP_ON_EXIT'

echo "demo_entrypoint_usage_test.sh: PASS"

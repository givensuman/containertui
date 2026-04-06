#!/usr/bin/env bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENTRYPOINT="$PROJECT_ROOT/demos/demo-entrypoint.sh"

assert_contains() {
	local needle="$1"
	if ! grep -Fq -- "$needle" "$ENTRYPOINT"; then
		echo "Expected to find: $needle" >&2
		exit 1
	fi
}

assert_not_contains() {
	local needle="$1"
	if grep -Fq -- "$needle" "$ENTRYPOINT"; then
		echo "Did not expect to find: $needle" >&2
		exit 1
	fi
}

assert_contains 'Ephemeral demo (Docker-in-Docker)'
assert_contains 'start_dind'
assert_contains '/demo/setup.sh'
assert_contains 'dockerd --iptables=false --ip6tables=false'
assert_not_contains 'CTUI_DEMO_MODE'
assert_not_contains 'CTUI_DEMO_SEED'
assert_not_contains 'CTUI_DEMO_CLEANUP_ON_EXIT'
assert_not_contains '--demo'

echo "demo_entrypoint_usage_test.sh: PASS"

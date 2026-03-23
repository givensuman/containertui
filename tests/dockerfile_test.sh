#!/bin/bash
# Test script for Dockerfile - verifies image builds and runs correctly
# This is the RED phase of TDD - tests that should pass once Dockerfile is created

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE_NAME="containertui:test"
TEST_CONTAINER="containertui-test-$$"

echo "=== Dockerfile Test Suite ==="

# Test 1: Dockerfile exists
echo -n "Test 1: Dockerfile exists... "
if [ -f "$PROJECT_ROOT/Dockerfile" ]; then
	echo "PASS"
else
	echo "FAIL"
	exit 1
fi

# Test 2: Image builds successfully
echo -n "Test 2: Image builds successfully... "
if docker build -t "$IMAGE_NAME" "$PROJECT_ROOT" >/dev/null 2>&1; then
	echo "PASS"
else
	echo "FAIL"
	docker build -t "$IMAGE_NAME" "$PROJECT_ROOT"
	exit 1
fi

# Test 3: Image contains containertui binary
echo -n "Test 3: Binary exists in image... "
if docker run --rm --entrypoint=test "$IMAGE_NAME" -f /usr/local/bin/containertui >/dev/null 2>&1; then
	echo "PASS"
else
	echo "FAIL"
	exit 1
fi

# Test 4: Image runs with containertui as entrypoint
echo -n "Test 4: Default entrypoint executes binary with --help... "
OUTPUT=$(docker run --rm "$IMAGE_NAME" 2>&1 || true)
if echo "$OUTPUT" | grep -q "Usage:"; then
	echo "PASS"
else
	echo "FAIL"
	echo "Output was: $OUTPUT"
	exit 1
fi

# Test 5: Binary works and shows help output
echo -n "Test 5: Binary shows help/usage info... "
OUTPUT=$(docker run --rm "$IMAGE_NAME" 2>&1 || true)
if echo "$OUTPUT" | grep -qE "(Usage:|Flags:|Commands:)"; then
	echo "PASS"
else
	echo "FAIL"
	echo "Output was: $OUTPUT"
	exit 1
fi

# Test 6: docker-cli is installed in image
echo -n "Test 6: docker-cli installed... "
if docker run --rm --entrypoint=docker "$IMAGE_NAME" --version >/dev/null 2>&1; then
	echo "PASS"
else
	echo "FAIL"
	exit 1
fi

# Test 7: Container runs as non-root user
echo -n "Test 7: Runs as non-root user... "
USER_ID=$(docker run --rm --entrypoint=id "$IMAGE_NAME" -u)
if [ "$USER_ID" != "0" ]; then
	echo "PASS (uid: $USER_ID)"
else
	echo "FAIL - running as root"
	exit 1
fi

# Test 8: Can accept custom command arguments
echo -n "Test 8: Binary accepts custom commands... "
OUTPUT=$(docker run --rm "$IMAGE_NAME" --help 2>&1 || true)
if echo "$OUTPUT" | grep -qE "(Usage:|Flags:|Commands:)"; then
	echo "PASS"
else
	echo "FAIL"
	exit 1
fi

# Test 9: ca-certificates installed (for HTTPS)
echo -n "Test 9: ca-certificates installed... "
if docker run --rm --entrypoint=test "$IMAGE_NAME" -f /etc/ssl/certs/ca-certificates.crt >/dev/null 2>&1 ||
	docker run --rm --entrypoint=test "$IMAGE_NAME" -d /etc/ssl/certs >/dev/null 2>&1; then
	echo "PASS"
else
	echo "FAIL"
	exit 1
fi

# Cleanup
docker rmi "$IMAGE_NAME" >/dev/null 2>&1 || true

echo ""
echo "=== All tests PASSED ==="

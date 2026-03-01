#!/bin/bash
# Cleanup script to remove test Docker resources created by setup.sh

set -e

echo "Removing binary..."

rm -f containertui 2>/dev/null || true

echo "Cleaning up Docker test environment..."

# Stop and remove test containers
echo "Removing test containers..."
docker rm -f containertui-nginx 2>/dev/null || true
docker rm -f containertui-alpine-logger 2>/dev/null || true
docker rm -f containertui-busybox 2>/dev/null || true
docker rm -f containertui-alpine-stopped 2>/dev/null || true
docker rm -f containertui-busybox-created 2>/dev/null || true

# Remove test volumes
echo "Removing test volumes..."
docker volume rm containertui-test-vol-1 2>/dev/null || true
docker volume rm containertui-test-vol-2 2>/dev/null || true
docker volume rm containertui-data 2>/dev/null || true

# Remove test networks
echo "Removing test networks..."
docker network rm containertui-network-1 2>/dev/null || true
docker network rm containertui-network-2 2>/dev/null || true

# Optionally remove pulled images (commented out by default)
# Uncomment if you want to remove test images as well
# echo "Removing test images..."
# docker rmi alpine:latest 2>/dev/null || true
# docker rmi nginx:alpine 2>/dev/null || true
# docker rmi busybox:latest 2>/dev/null || true

echo ""
echo "Cleanup complete!"

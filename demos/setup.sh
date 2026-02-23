#!/bin/bash
# Setup script to populate Docker with test containers, images, volumes, and networks
# for VHS demo generation

set -e

echo "Setting up Docker test environment for VHS demos..."

# Wait for Docker daemon to be ready
echo "Waiting for Docker daemon..."
timeout 30 sh -c 'until docker info > /dev/null 2>&1; do sleep 1; done' || {
	echo "Docker daemon failed to start"
	exit 1
}
echo "Docker daemon is ready"

# Pull base images (these will show up in images view)
echo "Pulling base images..."
docker pull alpine:latest || true
docker pull nginx:alpine || true
docker pull busybox:latest || true

# Create some test volumes
echo "Creating test volumes..."
docker volume create containertui-test-vol-1 || true
docker volume create containertui-test-vol-2 || true
docker volume create containertui-data || true

# Create custom networks
echo "Creating test networks..."
docker network create containertui-network-1 || true
docker network create containertui-network-2 || true

# Create test containers (some running, some stopped)
echo "Creating test containers..."

# Running container 1: nginx web server
docker run -d \
	--name containertui-nginx \
	--network containertui-network-1 \
	-p 8080:80 \
	nginx:alpine || true

# Running container 2: alpine with continuous date output
docker run -d \
	--name containertui-alpine-logger \
	--network containertui-network-1 \
	-v containertui-test-vol-1:/data \
	alpine sh -c "while true; do date; sleep 2; done" || true

# Running container 3: busybox with simple loop
docker run -d \
	--name containertui-busybox \
	--network containertui-network-2 \
	-v containertui-test-vol-2:/app \
	busybox sh -c "while true; do echo 'Container running...'; sleep 5; done" || true

# Stopped container 1: alpine (exited)
docker run -d \
	--name containertui-alpine-stopped \
	alpine echo "This container has exited" || true
sleep 2
docker stop containertui-alpine-stopped 2>/dev/null || true

# Stopped container 2: busybox (created but not started)
docker create \
	--name containertui-busybox-created \
	busybox echo "This container was only created" || true

echo ""
echo "Docker environment setup complete!"
echo ""
echo "Summary:"
echo "  Images: $(docker images --format '{{.Repository}}:{{.Tag}}' | wc -l) images"
echo "  Containers: $(docker ps -a --format '{{.Names}}' | wc -l) containers ($(docker ps --format '{{.Names}}' | wc -l) running)"
echo "  Volumes: $(docker volume ls --format '{{.Name}}' | wc -l) volumes"
echo "  Networks: $(docker network ls --format '{{.Name}}' | grep containertui | wc -l) custom networks"
echo ""
echo "Ready for VHS demo recording!"

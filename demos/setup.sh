#!/bin/bash
# Setup script to populate Docker with demo containers, images, volumes, and networks.

set -e

NAMESPACE="${CTUI_DEMO_NAMESPACE:-containertui-demo}"
NETWORK_ONE="${NAMESPACE}-network-1"
NETWORK_TWO="${NAMESPACE}-network-2"
VOLUME_ONE="${NAMESPACE}-vol-1"
VOLUME_TWO="${NAMESPACE}-vol-2"
VOLUME_DATA="${NAMESPACE}-data"
CONTAINER_NGINX="${NAMESPACE}-nginx"
CONTAINER_LOGGER="${NAMESPACE}-alpine-logger"
CONTAINER_BUSYBOX="${NAMESPACE}-busybox"
CONTAINER_STOPPED="${NAMESPACE}-alpine-stopped"
CONTAINER_CREATED="${NAMESPACE}-busybox-created"
COMPOSE_PROJECT="${NAMESPACE}"
COMPOSE_FILE="/tmp/${NAMESPACE}-compose.yml"

echo "Setting up Docker demo environment for namespace: ${NAMESPACE}"

# Wait for Docker daemon to be ready
# echo "Waiting for Docker daemon..."
# timeout 30 sh -c 'until docker info > /dev/null 2>&1; do sleep 1; done' || {
# 	echo "Docker daemon failed to start"
# 	exit 1
# }
# echo "Docker daemon is ready"

# Pull base images (these show up in images view)
echo "Pulling base images..."
docker pull alpine:latest
docker pull nginx:alpine
docker pull busybox:latest

# Create some test volumes
echo "Creating test volumes..."
docker volume create "${VOLUME_ONE}" >/dev/null
docker volume create "${VOLUME_TWO}" >/dev/null
docker volume create "${VOLUME_DATA}" >/dev/null

# Create custom networks
echo "Creating test networks..."
docker network create "${NETWORK_ONE}" >/dev/null
docker network create "${NETWORK_TWO}" >/dev/null

# Create a demo compose project so Services tab has data
echo "Creating demo compose service..."
cat >"${COMPOSE_FILE}" <<EOF
services:
  demo-service:
    image: alpine:latest
    command: ["sh", "-c", "while true; do echo demo-service-alive; sleep 10; done"]
EOF

if docker compose version >/dev/null 2>&1; then
	docker compose -f "${COMPOSE_FILE}" -p "${COMPOSE_PROJECT}" up -d
else
	echo "docker compose plugin not available; skipping services demo seed"
fi

# Create test containers (some running, some stopped)
echo "Creating test containers..."

# Running container 1: nginx web server
docker run -d \
	--name "${CONTAINER_NGINX}" \
	--network "${NETWORK_ONE}" \
	nginx:alpine >/dev/null

# Running container 2: alpine with continuous date output
docker run -d \
	--name "${CONTAINER_LOGGER}" \
	--network "${NETWORK_ONE}" \
	-v "${VOLUME_ONE}:/data" \
	alpine sh -c "while true; do date; sleep 2; done" >/dev/null

# Running container 3: busybox with simple loop
docker run -d \
	--name "${CONTAINER_BUSYBOX}" \
	--network "${NETWORK_TWO}" \
	-v "${VOLUME_TWO}:/app" \
	busybox sh -c "while true; do echo 'Container running...'; sleep 5; done" >/dev/null

# Stopped container 1: alpine (exited)
docker run -d \
	--name "${CONTAINER_STOPPED}" \
	alpine echo "This container has exited" >/dev/null
sleep 2
docker stop "${CONTAINER_STOPPED}" >/dev/null 2>&1 || true

# Stopped container 2: busybox (created but not started)
docker create \
	--name "${CONTAINER_CREATED}" \
	busybox echo "This container was only created" >/dev/null

echo ""
echo "Docker environment setup complete!"
echo ""
echo "Summary:"
echo "  Images: $(docker images --format '{{.Repository}}:{{.Tag}}' | wc -l) images"
echo "  Containers: $(docker ps -a --format '{{.Names}}' | wc -l) containers ($(docker ps --format '{{.Names}}' | wc -l) running)"
echo "  Volumes: $(docker volume ls --format '{{.Name}}' | wc -l) volumes"
echo "  Networks: $(docker network ls --format '{{.Name}}' | grep "${NAMESPACE}" | wc -l) custom networks"
echo ""
echo "Demo seed complete."

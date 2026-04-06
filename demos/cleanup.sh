#!/bin/bash
# Cleanup script for demo environment - removes test Docker resources created by setup.sh

set -e

NAMESPACE="${CTUI_DEMO_NAMESPACE:-containertui-demo}"
IMAGE_BASE="${NAMESPACE}/demo-base:1.0"
IMAGE_WEB="${NAMESPACE}/demo-web:1.0"
IMAGE_TOOLING="${NAMESPACE}/demo-tooling:1.0"
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

echo "Cleaning up Docker demo environment for namespace: ${NAMESPACE}"

# Stop and remove test containers
echo "Removing test containers..."
docker rm -f "${CONTAINER_NGINX}" 2>/dev/null || true
docker rm -f "${CONTAINER_LOGGER}" 2>/dev/null || true
docker rm -f "${CONTAINER_BUSYBOX}" 2>/dev/null || true
docker rm -f "${CONTAINER_STOPPED}" 2>/dev/null || true
docker rm -f "${CONTAINER_CREATED}" 2>/dev/null || true

# Remove test volumes
echo "Removing test volumes..."
docker volume rm "${VOLUME_ONE}" 2>/dev/null || true
docker volume rm "${VOLUME_TWO}" 2>/dev/null || true
docker volume rm "${VOLUME_DATA}" 2>/dev/null || true

# Remove test networks
echo "Removing test networks..."
docker network rm "${NETWORK_ONE}" 2>/dev/null || true
docker network rm "${NETWORK_TWO}" 2>/dev/null || true

# Tear down demo compose project/service
echo "Removing demo compose service..."
docker compose -p "${COMPOSE_PROJECT}" down --remove-orphans 2>/dev/null || true
rm -f "${COMPOSE_FILE}" 2>/dev/null || true

# Remove tagged demo images
echo "Removing demo images..."
docker image rm "${IMAGE_BASE}" "${IMAGE_WEB}" "${IMAGE_TOOLING}" 2>/dev/null || true

echo "Cleanup complete!"

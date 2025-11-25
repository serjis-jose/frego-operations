#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
IMAGE_NAME=${IMAGE_NAME:-frego-operations}
IMAGE_TAG=${IMAGE_TAG:-latest}
PLATFORM=${PLATFORM:-linux/amd64}
REGISTRY=${REGISTRY:-}

if ! command -v docker > /dev/null 2>&1; then
	echo "docker command not found. Please install Docker." >&2
	exit 1
fi

echo "==> Building Docker image for Operations microservice"
echo "    Image: ${IMAGE_NAME}:${IMAGE_TAG}"
echo "    Platform: ${PLATFORM}"
echo "    Root: ${ROOT_DIR}"

cd "$ROOT_DIR"

echo "==> Building ${IMAGE_NAME}:${IMAGE_TAG} (${PLATFORM})"
docker build \
	--platform "${PLATFORM}" \
	--build-arg SQLC_VERSION=v1.30.0 \
	-t "${IMAGE_NAME}:${IMAGE_TAG}" \
	.

echo "==> Build complete: ${IMAGE_NAME}:${IMAGE_TAG}"

# Tag with registry if provided
if [ -n "$REGISTRY" ]; then
	FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
	echo "==> Tagging for registry: ${FULL_IMAGE}"
	docker tag "${IMAGE_NAME}:${IMAGE_TAG}" "${FULL_IMAGE}"
	
	# Optionally push if PUSH=true
	if [ "${PUSH:-false}" = "true" ]; then
		echo "==> Pushing to registry: ${FULL_IMAGE}"
		docker push "${FULL_IMAGE}"
		echo "==> Push complete"
	fi
fi

echo "==> Done!"

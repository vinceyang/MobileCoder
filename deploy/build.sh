#!/bin/bash
set -e

echo "=== Building Docker images ==="

# Build Cloud image
echo "Building cloud image..."
docker build -t agentapi/cloud:latest -f deploy/Dockerfile.cloud .

# Build Chat image
echo "Building chat image..."
docker build -t agentapi/chat:latest -f deploy/Dockerfile.chat .

echo "=== Build complete ==="
docker images | grep agentapi

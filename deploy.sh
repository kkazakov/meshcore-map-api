#!/bin/bash

set -e

PROJECT_NAME="meshcore-map-api"
CONTAINER_NAME="meshcore-map-api"
IMAGE_NAME="meshcore-map-api:latest"
PORT="8080"

echo "=== Starting deployment process ==="

if [ ! -f ".env" ]; then
    echo "Error: .env file not found in current directory"
    echo "Please create a .env file before running this script"
    exit 1
fi

echo "✓ .env file found"

echo ""
echo "=== Building Go project ==="
go build -o server
echo "✓ Go build completed"

echo ""
echo "=== Building Docker image ==="
docker build -t $IMAGE_NAME .
echo "✓ Docker image built: $IMAGE_NAME"

echo ""
echo "=== Stopping and removing existing container (if any) ==="
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
    echo "✓ Existing container removed"
else
    echo "✓ No existing container found"
fi

echo ""
echo "=== Starting new container ==="
docker run -d \
    --name $CONTAINER_NAME \
    --restart unless-stopped \
    -p $PORT:8080 \
    --env-file .env \
    $IMAGE_NAME

echo "✓ Container started successfully"

echo ""
echo "=== Deployment complete ==="
echo "Container name: $CONTAINER_NAME"
echo "Container status:"
docker ps --filter "name=$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "To view logs, run: docker logs -f $CONTAINER_NAME"
echo "To stop the container, run: docker stop $CONTAINER_NAME"

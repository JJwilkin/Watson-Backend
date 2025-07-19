#!/bin/bash

# Parse command line arguments
PUSH_IMAGES=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --push|-p)
            PUSH_IMAGES=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--push|-p]"
            exit 1
            ;;
    esac
done

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Go to the parent directory (project root)
cd "$SCRIPT_DIR/.."

echo "Building Docker images..."

# Build images locally from project root
docker build -f api/Dockerfile -t us-central1-docker.pkg.dev/watson-465400/watson-registry/watson-go-api:latest .
docker build -f background-worker/Dockerfile -t us-central1-docker.pkg.dev/watson-465400/watson-registry/watson-worker:latest .

# Ask for confirmation to push if no flag was provided
if [ "$PUSH_IMAGES" = false ]; then
    echo ""
    read -p "Do you want to push the images to the registry? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        PUSH_IMAGES=true
    fi
fi

# Push to Google Container Registry if confirmed
if [ "$PUSH_IMAGES" = true ]; then
    echo "Pushing images to registry..."
    docker push us-central1-docker.pkg.dev/watson-465400/watson-registry/watson-go-api:latest
    docker push us-central1-docker.pkg.dev/watson-465400/watson-registry/watson-worker:latest
    echo "Images pushed successfully!"
else
    echo "Images built locally. Use --push or -p to push to registry, or run again and answer 'y' when prompted."
fi
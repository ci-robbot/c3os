#!/bin/bash
set -e

echo "Building Kairos SLE Micro 5.5 image..."

# Check if repos directory exists
if [ ! -d "repos" ]; then
    echo "Error: 'repos' directory not found. Please create it with your SUSE repository files."
    echo "Expected: repos/SUSE* and repos/opensuse*"
    exit 1
fi

# Check if services directory exists
if [ ! -d "services" ]; then
    echo "Error: 'services' directory not found. Please create it with your zypper service files."
    exit 1
fi

# Build the image
IMAGE_NAME="kairos-slemicro-5.5"
docker build -t "$IMAGE_NAME" .

echo "Build complete! Image: $IMAGE_NAME"

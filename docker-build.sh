#!/bin/bash

set -e

function cleanup {
    docker rm mlem || true
    docker rmi mlem || true
}

cleanup
trap cleanup EXIT

# Build by way of building a Docker image.
docker build -t mlem -f Builderfile .

# Copy the built debs out of a new container.
docker create --name mlem mlem bash
docker cp mlem:/packages .
mv packages/*.deb ..
rmdir packages/ || true

# Exit so cleanup() cleans up the container & image.


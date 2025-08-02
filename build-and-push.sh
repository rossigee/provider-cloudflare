#!/bin/bash

set -e

PROVIDER="provider-cloudflare"
BASE_REGISTRY="ghcr.io/rossigee"

# Get version from git or default
VERSION=${VERSION:-$(git describe --tags --always --dirty)}
if [[ "$VERSION" == *-dirty ]]; then
    echo "Warning: Working directory is dirty, using dirty tag"
fi

echo "Building ${PROVIDER} version ${VERSION}"

# Build and publish to primary registry (GitHub Container Registry)
make build
make publish REGISTRY_ORGS="${BASE_REGISTRY}"

echo "Successfully built and published ${PROVIDER}:${VERSION} to ${BASE_REGISTRY}"
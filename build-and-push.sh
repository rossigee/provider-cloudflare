#!/bin/bash

set -e

PROVIDER="provider-cloudflare"
BASE_REGISTRY="ghcr.io/rossigee"

# Get version from VERSION file or default
if [[ -f VERSION ]]; then
    VERSION=${VERSION:-$(cat VERSION)}
else
    VERSION=${VERSION:-"v0.0.0-dev"}
fi

echo "Building ${PROVIDER} version ${VERSION}"

# Build and publish to primary registry (GitHub Container Registry)
make build
make publish REGISTRY_ORGS="${BASE_REGISTRY}"

echo "Successfully built and published ${PROVIDER}:${VERSION} to ${BASE_REGISTRY}"
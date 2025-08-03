# Publishing Guide for provider-cloudflare

This guide covers how to publish the provider-cloudflare package to various registries.

## Prerequisites

- Crossplane CLI (`up`) installed
- Docker running for building images
- Registry credentials configured

## Current Version

The current version is `v0.6.1` with the following registries:

- **Primary**: `ghcr.io/rossigee/provider-cloudflare:v0.6.1`
- **Upbound**: `xpkg.upbound.io/rossigee/provider-cloudflare:v0.6.1` (marketplace)

## Publishing Process

### 1. Build and Validate

```bash
# Build the package
make xpkg.build

# Validate package structure
make validate-package
```

### 2. Publish to GitHub Container Registry (Primary)

```bash
# This is handled automatically by CI/CD
make publish
```

### 3. Publish to Upbound Marketplace

```bash
# Publish to Upbound Marketplace
make publish-upbound
```

### 4. Manual Publishing to Upbound

If the make target fails, use the up CLI directly:

```bash
# Login to Upbound (if not already logged in)
up login

# Push to marketplace
up xpkg push package/provider-cloudflare-v0.6.1.xpkg xpkg.upbound.io/rossigee/provider-cloudflare:v0.6.1
```

## Package Structure

The package includes:

- **Metadata**: `package/crossplane.yaml` with marketplace annotations
- **Icon**: `.github/cloudflare-icon.svg` for marketplace listing
- **CRDs**: All generated Custom Resource Definitions
- **Documentation**: Comprehensive README and examples

## Marketplace Compliance

The package meets Upbound Marketplace requirements:

- ✅ Proper metadata annotations (maintainer, source, license)
- ✅ Marketplace-compliant icon
- ✅ Comprehensive documentation
- ✅ Category classification (Infrastructure)
- ✅ Keywords for discoverability
- ✅ Version constraints and dependencies

## Versioning Strategy

- Use semantic versioning (MAJOR.MINOR.PATCH)
- Tag releases in git: `git tag v0.6.1`
- Update version in relevant files when bumping
- Maintain changelog for marketplace visibility

## Registry Configuration

The Makefile supports multiple registries:

```bash
# Primary (default)
make publish

# Upbound Marketplace
make publish-upbound

# Harbor (if configured)
ENABLE_HARBOR_PUBLISH=true make publish XPKG_REG_ORGS=harbor.golder.lan/library
```

## Troubleshooting

### Package Build Issues
- Ensure all dependencies are up to date: `go mod tidy`
- Check build system: `make submodules`
- Verify CRD generation: `make generate`

### Publishing Issues
- Verify registry credentials
- Check network connectivity
- Confirm package validation passes

### Marketplace Issues
- Validate all required annotations are present
- Ensure icon file exists and is accessible
- Check marketplace-specific requirements

For support, see:
- [Upbound Marketplace Docs](https://docs.upbound.io/upbound-marketplace/)
- [Crossplane Docs](https://docs.crossplane.io/)
- [GitHub Issues](https://github.com/rossigee/provider-cloudflare/issues)
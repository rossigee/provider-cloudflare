# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.22.1] - 2026-11-12

### Changed
- Built with Go 1.27.4 (previously 1.27.3).

## [v0.22.0] - 2026-11-08

### Fixed
- **Connect failed for generated resource types**: `GetConfig` asserted that
  managed resources return `*xpv1.Reference` from `GetProviderConfigReference`.
  The generated methods return the spec's `ProviderConfigReference`, so the
  assertion never matched and every `Connect` failed with "providerConfigRef
  not set". Affected: Cache, Device, EmailRouting, Logpush, PageRules, Pages,
  R2, Rulesets, Security, SSL, Tunnel and Workers resources. This affected
  v0.21.2 through v0.21.5 (PR #36).

### Added
- **AI Gateway**: `Gateway` resource in `aigateway.cloudflare.m.crossplane.io/v1beta1`
  for AI Gateway settings (PR #37). Uses cloudflare-go v7 alongside the v0 SDK.
  Routes, logs, datasets, evaluations, billing, DLP, guardrails and Workers AI
  are not modeled.
- **Zone DNSSEC**: `dnssecEnabled` on Zone, with DS, key tag and algorithm reported in status.
  DNSSEC state is only read when `spec.forProvider.dnssecEnabled` is set, so zones that
  do not manage DNSSEC do not depend on that endpoint.
- **Universal SSL verification status**: `certificatePackCount` and `validationStatus`
  reported on UniversalSSL. A failed verification lookup is returned as an error.
- **Records, Spectrum, Rulesets, RateLimit and R2 test coverage** (PR #21).
- **Examples and docs**: Pages and PageRule documented; examples for UniversalSSL,
  TotalTLS, CertificatePack, origin Certificate, R2 Bucket, logpush Job and
  email routing Rule.

### Changed
- **Workers**: `secret_text` binding type supported (PR #33).
- **Namespaced references**: references are looked up in the managed resource's namespace (PR #32).
- **Dependencies**: crossplane-runtime pinned to v2.5.0-dev; golangci-lint v2.15.0 (PRs #34, #35).

## [v0.21.5] - 2026-11-03

### Fixed
- **ProviderConfig namespace**: the ProviderConfig is looked up in the managed
  resource's namespace (PR #31).
- **Pinned versions**: install and package versions updated (PR #27).
- **RBAC**: core-group RBAC is no longer aggregated into view and edit roles (PR #26).

### Changed
- **Lint config**: uses the build submodule's lint configuration (PR #28).
- **CI**: GitHub Actions pinned to commit SHAs (PR #29).
- **Repository hygiene**: stale status and investigation logs removed (PR #30).

## [v0.21.4] - 2026-10-26

### Added
- **Core version floor**: the package declares Crossplane core >= v2.5.0 (PR #22).

### Fixed
- **RBAC for events**: `events.k8s.io` permissions granted for the v2 EventRecorder (PR #18).
- **CacheSyncTimeout**: increased to 30 minutes for resource-constrained clusters (PR #20).
- **Rate limiting**: client-side QPS and burst for the Kubernetes API (PR #19).

### Changed
- **Build**: Go 1.27.3; crossplane-runtime v2.5.1; scanner actions pinned (PRs #23, #24, #25).

## [v0.21.3] - 2026-10-03

### Fixed
- **Controller names**: qualified with the API group (PR #17).
- **Release**: standardized release workflow applied (PR #16); image publishing neutralized (PR #14).

### Changed
- **Dependencies**: Crossplane APIs fork bumped (PR #15).

## [v0.21.2] - 2026-09-23

### Changed
- **Publish as xpkg-only (multi-platform)**: Release now publishes the
  Crossplane package (xpkg) directly via `crossplane xpkg push`, with
  multi-platform metadata for `linux/amd64` and `linux/arm64`. This is the
  current standardized pattern for Crossplane providers and avoids issues
  with per-arch container image tagging. Image publishing is neutralized
  via the Makefile overrides that other providers use (PR #13).

### Fixed
- **Regenerated deepcopy methods**: `zz_generated.deepcopy.go` files for
  Pages and Page Rules were regenerated with controller-gen to use plain
  (non-aliased) `k8s.io/apimachinery/pkg/runtime` imports. Fixes the
  `check-diff` CI step.

## [v0.21.1] - 2026-09-23

### Fixed
- **Multi-arch manifest**: Fixed image publish to produce a real multi-arch OCI index
  for `linux/amd64` and `linux/arm64`. The v0.21.0 release shipped a single-arch manifest.
  `cluster/images/provider-cloudflare/Makefile::img.publish` now builds and pushes
  per-arch images, then creates the multi-arch index via
  `docker buildx imagetools create`.

## [v0.21.0] - 2026-09-23

### Added
- **Cloudflare Pages Support**: Complete implementation for managing Pages projects, deployments, and custom domains
  - `pages.cloudflare.m.crossplane.io/v1beta1` API group
  - `Project`, `Deployment`, and `Domain` resources with full CRUD support
  - Account-scoped operations using ResourceContainer
- **Page Rules Support**: Full support for legacy Page Rules
  - `pagerules.cloudflare.m.crossplane.io/v1beta1` API group
  - `PageRule` resource with targets, actions, priority, and status handling
- **New Examples**: Comprehensive examples under `examples/pages/` and `examples/pagerules/`
- **Documentation**: New resource documentation in `docs/resources/` for Pages and PageRules
- **Tests**: Interface-based controller tests with mocks for the new resources

### Changed
- **Version**: Bumped to v0.21.0
- Improved controller interface usage for better testability across Pages and PageRules controllers

### Infrastructure
- All new resources follow the established v2 controller patterns
- CRDs generated for new resources

## [v0.13.0] - 2025-10-27

### Changed
- **Go Tooling**: Updated golangci-lint to version 2.5.0 (latest stable) for improved linting and Go 1.25.3 compatibility

### Infrastructure
- **Build System**: Enhanced code quality assurance with latest linting standards
- **CI/CD**: Consistent golangci-lint versioning between local development and CI environment

## [v0.12.2] - 2025-10-26

### Added
- **Scheme Verification**: Added runtime scheme verification during provider startup to catch registration issues early
- **Debug Logging**: Enhanced debugging capabilities for zone controller setup and API group handling

### Changed
- **Go Version**: Updated to Go 1.25.3 for improved performance and latest language features
- **CI Build Validation**: Fixed build validation target to use correct `make build.artifacts.platform` instead of non-existent `make docker.build`

### Fixed
- **API Group Compatibility**: Fixed API group name mismatches for proper CRD compatibility
- **Zone Settings**: Added defensive programming to zone settings loading to prevent runtime errors
- **CI Pipeline**: Corrected make target references in CI workflow to ensure reliable build validation

### Infrastructure
- **Go 1.25.3**: Updated all CI/CD workflows, go.mod, and documentation to use Go 1.25.3
- **Build System**: Improved CI reliability with correct make target usage
- **Debugging**: Enhanced troubleshooting capabilities for zone controller operations

## [v0.12.0] - 2025-10-24

### Fixed
- **Critical**: Added missing scheme registration for all API resource types to prevent provider startup panics
- **Release Workflow**: Fixed VERSION variable passing in GitHub Actions release workflow to ensure proper version tagging
- **CI/CD**: Corrected make variable precedence issues in release pipeline

### Added
- **Worker Examples**: Complete set of example configurations for all worker resources (CronTrigger, Domain, KVNamespace, Subdomain)
- **Testing**: Enhanced scheme registration verification tests to catch registration issues early

### Infrastructure
- **Scheme Registration**: Comprehensive verification of all API types properly registered with Kubernetes scheme
- **Release Process**: Improved reliability of automated release workflow with correct variable handling

## [v0.11.0] - 2025-10-20

### Added
- **Complete Worker Resources**: Full implementation of all Cloudflare Worker APIs including Cron Triggers, Domains, KV Namespaces, Routes, and Subdomains
- **Production Ready Workers**: Enterprise-grade serverless edge computing support with comprehensive API integration
- **Worker Ecosystem**: Complete worker management capabilities for script deployment, scheduled execution, storage, and custom domains

### Changed
- **BREAKING**: Migrated all resources to stable v1beta1 APIs for production stability
- **API Maturity**: All resource APIs now follow Kubernetes API conventions and stability guarantees
- **Documentation**: Comprehensive worker resource usage examples and deployment guides

### Fixed
- **Release Infrastructure**: Fixed GitHub Actions workflow VERSION variable handling (manual workaround applied for v0.11.0)
- **API Compatibility**: Ensured all resource types properly integrate with Crossplane v1.16.0+
- **Documentation**: Updated all examples and guides to reflect v1beta1 API changes

### Infrastructure
- **Crossplane Compatibility**: Full compatibility with Crossplane v1.16.0 and Kubernetes API standards
- **Build System**: Enhanced release workflow with proper version handling for future releases
- **Testing**: Comprehensive interface-based testing for all worker resource controllers

## [v0.9.1] - 2025-09-16

### Fixed
- Build system compatibility with Go 1.25.1 by updating golangci-lint to v2.4.0
- Build submodule initialization and standardization across all workflows
- CI/CD workflow version consistency (Go 1.25.1 across CI, Release, Security)
- Docker build target references in CI validation workflow
- Documentation version mismatches updated to reflect current v0.9.1 release

### Changed
- Updated all documentation to reference Go 1.25.1 and current v0.9.1 version
- Standardized golangci-lint version across Makefile and CI workflows
- Improved build system reliability through proper submodule management

### Infrastructure
- **Build System**: Full compatibility with Go 1.25.1 and modern toolchain
- **CI/CD**: Consistent versioning across all workflow files
- **Documentation**: Accurate version references and installation instructions

## [v0.9.0] - 2025-08-14

### Added
- Standardized CI/CD build system with "CI Builds, Release Publishes" pattern
- Comprehensive security scanning (govulncheck, gosec, CodeQL) in CI pipeline
- Parallel validation jobs for optimal CI performance
- Transform Rules resource for URL rewriting, header modification, and HTTP redirects
- Enhanced CRD display columns for Workers subdomain with full domain format

### Changed
- **BREAKING**: Migrated to standardized CI/CD workflows eliminating tag conflicts
- Updated CI workflow to validation-only (no publishing)
- Updated Release workflow as single source of truth for all publishing
- Registry standardization to `ghcr.io/rossigee/provider-cloudflare` for consistency
- Enhanced marketplace metadata with updated installation instructions
- Improved build validation with artifact verification

### Fixed
- CI/CD tag conflicts between workflows by implementing proper separation
- Build system reliability through standardized templates
- Security scanning integration with GitHub Security tab via SARIF uploads

### Infrastructure
- **CI Pipeline**: Build validation only with parallel job execution
- **Release Pipeline**: Single publishing source ensuring identical version/latest tags
- **Security**: Nightly security scans + per-commit vulnerability detection
- **Registry**: Primary ghcr.io/rossigee with consistent tagging strategy

## [v0.6.1] - 2025-01-08

### Added
- Comprehensive Upbound Marketplace support with proper metadata and icon
- Modern Ruleset resource with complete WAF integration
- Load Balancing resources (LoadBalancer, LoadBalancerPool, LoadBalancerMonitor)
- Cache Rules resource for advanced caching configurations
- Interface-based testing framework with comprehensive mock support
- Dedicated SRV record support with proper validation
- URI transformation parameters in Rulesets for advanced URL rewriting
- Publishing documentation and versioning strategy
- Marketplace-compliant package structure

### Fixed
- All linting issues including deprecated pointer usage and staticcheck warnings
- DNS controller interface tests with proper SRV record handling
- JumpStart documentation with clear usage guidelines and warnings
- Embedded field selector optimization across all clients
- Comprehensive test coverage improvements

### Changed
- Updated to Go 1.24.5 with modern dependencies
- Migrated from k8s.io/utils/pointer to k8s.io/utils/ptr
- Improved code quality standards and validation
- Enhanced error handling and client isolation

### Removed
- Unused test helper functions for cleaner codebase
- Deprecated pointer usage patterns
- TODO comments replaced with proper documentation

## [v0.6.0] - 2024-12-15

### Added
- Initial release with comprehensive Cloudflare API coverage
- Zone, Record, Rule, Filter, Application, Route, CustomHostname resources
- Complete test suite with mock implementations
- Crossplane runtime integration
- Docker containerization with distroless base

### Features
- DNS management with all record types
- Firewall rules and filters (legacy)
- Spectrum applications for TCP/UDP acceleration
- SSL for SaaS certificate management
- Worker route bindings
- Zone-level settings management

[v0.13.0]: https://github.com/rossigee/provider-cloudflare/compare/v0.12.2...v0.13.0
[v0.12.2]: https://github.com/rossigee/provider-cloudflare/compare/v0.12.0...v0.12.2
[v0.12.0]: https://github.com/rossigee/provider-cloudflare/compare/v0.11.0...v0.12.0
[v0.11.0]: https://github.com/rossigee/provider-cloudflare/compare/v0.9.1...v0.11.0
[v0.9.1]: https://github.com/rossigee/provider-cloudflare/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/rossigee/provider-cloudflare/compare/v0.6.1...v0.9.0
[v0.6.1]: https://github.com/rossigee/provider-cloudflare/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/rossigee/provider-cloudflare/releases/tag/v0.6.0
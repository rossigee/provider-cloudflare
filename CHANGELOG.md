# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[v0.6.1]: https://github.com/rossigee/provider-cloudflare/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/rossigee/provider-cloudflare/releases/tag/v0.6.0
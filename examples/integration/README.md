# Integration Test Examples

This directory contains comprehensive integration test scenarios that demonstrate real-world usage patterns for the Cloudflare provider.

## Overview

These examples showcase:
- **End-to-end workflows** - Complete infrastructure setups
- **Provider composition** - Multiple resources working together  
- **Testing patterns** - Validation and verification approaches
- **Real-world scenarios** - Production-like configurations

## Test Categories

### 🌐 Complete Website Setup
- DNS zone creation with comprehensive settings
- DNS records for web services (A, AAAA, CNAME, MX, TXT)
- SSL/TLS configuration and certificates
- Performance optimization (caching, load balancing)
- Security protection (WAF rules, firewall)

### 🚀 Edge Computing & Workers
- Worker route configurations
- Traffic steering and geographic routing
- Edge compute deployments
- API gateway patterns

### 🔒 Security & Compliance  
- Modern WAF ruleset deployments
- SSL for SaaS certificate management
- DDoS protection configuration
- Security policy enforcement

### ⚡ Performance & Reliability
- Load balancing with health checks
- Cache rule optimization
- Traffic steering strategies
- Failover configurations

### 🧪 Testing Strategies
- Resource validation patterns
- Dependency ordering
- Error handling scenarios
- Rollback procedures

## Usage Instructions

1. **Prerequisites**: Configure provider credentials
2. **Select scenario**: Choose appropriate example for your use case
3. **Customize values**: Update domain names and settings
4. **Apply incrementally**: Follow dependency order
5. **Validate results**: Verify resources in Cloudflare dashboard

## Best Practices

- Always test in development zones first
- Use resource dependencies to ensure proper ordering
- Implement health checks for critical services
- Monitor resource status and conditions
- Plan rollback strategies for production changes

## File Organization

```
integration/
├── README.md                    # This file
├── complete-website/            # Full website setup
├── edge-computing/              # Worker and edge examples
├── security-compliance/         # Security-focused configurations
├── performance-reliability/     # Performance optimization
└── testing-patterns/           # Testing and validation examples
```
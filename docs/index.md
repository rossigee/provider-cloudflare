# Provider Cloudflare Documentation

A Crossplane provider for managing Cloudflare resources.

## Quick Links

- [v2 Migration Guide](v2-migration-guide.md) — Migrating from v1 to v2
- [v2 Implementation Status](v2-implementation-status.md) — Current v2 support
- [v2 Test Status](v2-test-status.md) — Testing progress

## Resource Documentation

Resources are documented in the API types. Individual resource documentation will be added to the [resources/](resources/) folder.

### DNS & Traffic

| Resource | API Group | Description |
|----------|-----------|-------------|
| DNS | `dns.cloudflare.m.crossplane.io/v1beta1` | DNS records |
| LoadBalancing | `loadbalancing.cloudflare.m.crossplane.io/v1beta1` | Load balancers |

### Security

| Resource | API Group | Description |
|----------|-----------|-------------|
| Access | `access.cloudflare.m.crossplane.io/v1beta1` | Cloudflare Access policies |
| Firewall | `firewall.cloudflare.m.crossplane.io/v1beta1` | Firewall rules |

### Other Resources

See `apis/` directory for all available resources.

## Status

This provider is undergoing v2 expansion. See v2 implementation status for current resource coverage.

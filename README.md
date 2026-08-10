# provider-cloudflare

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-cloudflare/ci.yml?branch=master)][build]
[![Version](https://img.shields.io/github/v/release/rossigee/provider-cloudflare)][releases]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[build]: https://github.com/rossigee/provider-cloudflare/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-cloudflare/releases

## Overview

A Crossplane provider for managing Cloudflare resources via the Cloudflare v4 API — DNS, load balancing, WAF, caching, SSL, Zero Trust Access, tunneling, Workers, and device management.

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-cloudflare:v0.14.10`

## Features

- **DNS & Zones**: Zone settings and all DNS record types (A, AAAA, CNAME, MX, TXT, SRV, etc.)
- **Security & Firewall**: Modern Ruleset Engine WAF rules, plus legacy Rule/Filter support
- **Load Balancing**: Geographic traffic steering, origin pools, and health-check monitors
- **Performance**: Cache rules with custom TTL, bypass conditions, and cache-key customization
- **Workers**: Scripts, cron triggers, custom domains, KV namespaces, and routes for serverless edge computing
- **SSL/TLS**: SSL for SaaS custom hostname and fallback origin management
- **Zero Trust**: Access applications with authentication policies, and device posture rules
- **Tunneling**: Cloudflare Tunnel (Cloudflared) connections for secure remote access
- **Spectrum**: TCP/UDP traffic acceleration applications
- **Security**: Bot management, rate limiting, and Turnstile CAPTCHA
- **Email Routing**: Inbound email routing rules
- **R2 & Logpush**: Object storage buckets and log export jobs

## Getting Started

### Prerequisites

- Kubernetes with Crossplane installed
- A Cloudflare API token with the permissions needed for the resources you plan to manage (e.g. `Zone:Zone:Read`, `Zone:Zone:Edit`, `Zone:DNS:Edit`)

### Installation

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-cloudflare:v0.14.10
```

### Configuration

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  token: "your-cloudflare-api-token-here"
---
apiVersion: cloudflare.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: cloudflare-credentials
      key: token
```

## Usage

```yaml
apiVersion: dns.cloudflare.m.crossplane.io/v1beta1
kind: Record
metadata:
  name: example-a-record
  namespace: default
spec:
  forProvider:
    name: "app.example.com"
    type: "A"
    content: "203.0.113.10"
    ttl: 3600
    zone: "your-zone-id-here"
  providerConfigRef:
    name: default
  deletionPolicy: Delete
```

## Resource Types

| Resource | API Group | Description |
|----------|-----------|-------------|
| Zone | `zone.cloudflare.m.crossplane.io` | DNS zones and zone settings |
| Record | `dns.cloudflare.m.crossplane.io` | DNS records (all types) |
| Ruleset | `rulesets.cloudflare.m.crossplane.io` | Modern WAF rules |
| Rule / Filter | `firewall.cloudflare.m.crossplane.io` | Legacy firewall rules (deprecated, prefer Ruleset) |
| AccessApplication | `access.cloudflare.m.crossplane.io` | Zero Trust access applications |
| DevicePostureRule | `device.cloudflare.m.crossplane.io` | Endpoint security posture rules |
| LoadBalancer / LoadBalancerPool / LoadBalancerMonitor | `loadbalancing.cloudflare.m.crossplane.io` | Geographic load balancing |
| CacheRule | `cache.cloudflare.m.crossplane.io` | Cache TTL, bypass, and key rules |
| Rule (transform) | `transform.cloudflare.m.crossplane.io` | URL rewrites, header modification, redirects |
| Application | `spectrum.cloudflare.m.crossplane.io` | Spectrum TCP/UDP acceleration |
| Script / CronTrigger / Domain / KVNamespace / Route / Subdomain | `workers.cloudflare.m.crossplane.io` | Workers serverless edge computing |
| Tunnel | `tunnel.cloudflare.m.crossplane.io` | Cloudflare Tunnel (Cloudflared) |
| CustomHostname / FallbackOrigin | `sslsaas.cloudflare.m.crossplane.io` | SSL for SaaS |
| LogpushJob | `logpush.cloudflare.m.crossplane.io` | Logpush job configuration |
| R2Bucket | `r2.cloudflare.m.crossplane.io` | R2 object storage buckets |
| BotManagement / RateLimit / Turnstile | `security.cloudflare.m.crossplane.io` | Bot management, rate limiting, Turnstile CAPTCHA |
| Rule (email routing) | `emailrouting.cloudflare.m.crossplane.io` | Email routing rules |
| CertificatePack / TotalTLS / UniversalSSL | `ssl.cloudflare.m.crossplane.io` | SSL/TLS certificate management |
| Certificate (origin) | `originssl.cloudflare.m.crossplane.io` | Origin CA certificates |

See `examples/` for a full set of working manifests per resource type.

## Development

```bash
# Build
make build

# Test
make test

# Lint
make lint

# Generate
make generate
```

## Contributing

Issues and pull requests are welcome at [github.com/rossigee/provider-cloudflare](https://github.com/rossigee/provider-cloudflare).

## License

provider-cloudflare is under the Apache 2.0 license.

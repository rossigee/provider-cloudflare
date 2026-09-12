# Provider Cloudflare Documentation

A Crossplane v2 provider for managing Cloudflare resources. All managed resources are namespaced (`*.cloudflare.m.crossplane.io/v1beta1`) with full multi-tenancy support.

## Quick Links

- [Getting Started](getting-started.md) — Installation and first resources
- [Configuration](configuration.md) — Authentication and connection setup
- [Development](development.md) — Building, testing, and contributing

## Resource Documentation

### DNS & Traffic

| Resource | API Group | Description |
|----------|-----------|-------------|
| Zone | `zone.cloudflare.m.crossplane.io/v1beta1` | DNS zones |
| Record | `dns.cloudflare.m.crossplane.io/v1beta1` | DNS records (all types incl. SRV) |
| LoadBalancer, LoadBalancerPool, LoadBalancerMonitor | `loadbalancing.cloudflare.m.crossplane.io/v1beta1` | Geographic traffic steering |

### Security

| Resource | API Group | Description |
|----------|-----------|-------------|
| Ruleset | `rulesets.cloudflare.m.crossplane.io/v1beta1` | Modern WAF rulesets |
| Rule, Filter | `firewall.cloudflare.m.crossplane.io/v1beta1` | Legacy firewall (prefer Rulesets) |
| AccessApplication | `access.cloudflare.m.crossplane.io/v1beta1` | Zero Trust access apps |
| DevicePostureRule | `device.cloudflare.m.crossplane.io/v1beta1` | Endpoint posture checks |
| RateLimit, BotManagement, Turnstile | `security.cloudflare.m.crossplane.io/v1beta1` | Abuse controls |

### Performance & Edge

| Resource | API Group | Description |
|----------|-----------|-------------|
| CacheRule | `cache.cloudflare.m.crossplane.io/v1beta1` | Cache behavior |
| Rule (transform) | `transform.cloudflare.m.crossplane.io/v1beta1` | URL/header transforms, redirects |
| Application (Spectrum) | `spectrum.cloudflare.m.crossplane.io/v1beta1` | TCP/UDP acceleration |
| Bucket (R2) | `r2.cloudflare.m.crossplane.io/v1beta1` | R2 object storage buckets |

### Workers & Tunnels

| Resource | API Group | Description |
|----------|-----------|-------------|
| Script, CronTrigger, Domain, KVNamespace, Route, Subdomain | `workers.cloudflare.m.crossplane.io/v1beta1` | Worker ecosystem |
| Tunnel | `tunnel.cloudflare.m.crossplane.io/v1beta1` | Cloudflared tunnels |
| Job (Logpush) | `logpush.cloudflare.m.crossplane.io/v1beta1` | Log delivery jobs |
| Rule (Email Routing) | `emailrouting.cloudflare.m.crossplane.io/v1beta1` | Email routing rules |

### SSL/TLS

| Resource | API Group | Description |
|----------|-----------|-------------|
| UniversalSSL, TotalTLS, CertificatePack | `ssl.cloudflare.m.crossplane.io/v1beta1` | Zone certificate management |
| Certificate (Origin) | `originssl.cloudflare.m.crossplane.io/v1beta1` | Origin certificates |
| CustomHostname, FallbackOrigin | `sslsaas.cloudflare.m.crossplane.io/v1beta1` | SSL for SaaS |

### Provider

| Resource | API Group | Description |
|----------|-----------|-------------|
| ProviderConfig | `cloudflare.m.crossplane.io/v1beta1` | Credentials (cluster-scoped) |

## API Coverage Gaps

Cloudflare API surface not yet modeled: Pages projects/deployments, D1/KV read-write data operations (namespaces only), Queues, AI Gateway/Workers AI bindings, Waiting Rooms, Web Analytics rules, DNSSEC DS management beyond zone settings, account-level audit logs, and Spectrum origins beyond application tunnels.

# Cloudflare Provider Examples

This directory contains comprehensive examples for all Cloudflare provider resources. Each subdirectory focuses on a specific resource type or feature area.

## Directory Structure

### Core Resources

- **[provider/](provider/)** - ProviderConfig setup and authentication examples
- **[zone/](zone/)** - DNS zone management with settings configuration
- **[record/](record/)** - DNS record examples (A, AAAA, CNAME, MX, TXT, SRV)

### Security & Firewall

- **[rulesets/](rulesets/)** - Modern WAF rulesets with advanced rule matching
- **[firewall/](firewall/)** - Legacy firewall rules and filters (deprecated)
- **[transform/](transform/)** - URL transformation and rewriting rules

### Load Balancing & Traffic Management

- **[loadbalancing/](loadbalancing/)** - Complete load balancing examples including:
  - Health check monitors
  - Origin server pools 
  - Geographic load balancers with traffic steering
  - Advanced rules and session affinity

### Performance & Caching

- **[cache/](cache/)** - Advanced cache rules with:
  - Basic TTL configuration
  - Bypass conditions
  - Geographic caching strategies
  - Advanced eligibility criteria

### Applications & Services

- **[spectrum/](spectrum/)** - TCP/UDP traffic acceleration applications
- **[workers/](workers/)** - Cloudflare Worker route bindings

### SSL/TLS & Certificates

- **[custom_hostname/](custom_hostname/)** - SSL for SaaS custom hostname management
- **[fallback_origin/](fallback_origin/)** - SSL for SaaS fallback origin configuration

## Quick Start

1. **Setup Provider Configuration:**
   ```bash
   kubectl apply -f provider/provider-config.yaml
   ```

2. **Create a Zone:**
   ```bash
   kubectl apply -f zone/zone.yaml
   ```

3. **Add Load Balancing:**
   ```bash
   kubectl apply -f loadbalancing/full-example.yaml
   ```

4. **Configure Security Rules:**
   ```bash
   kubectl apply -f rulesets/basic-security-ruleset.yaml
   ```

## Example Highlights

### Geographic Load Balancing
The `loadbalancing/full-example.yaml` demonstrates a complete setup with:
- Health monitoring across multiple regions
- Intelligent traffic steering based on geography
- Session affinity and failover policies

### Advanced WAF Protection
The `rulesets/` examples show modern security configurations:
- Request filtering based on URI patterns
- Geographic blocking and allowlisting
- Rate limiting and DDoS protection

### Performance Optimization
The `cache/` examples demonstrate:
- Custom TTL policies for different content types
- Conditional caching based on request headers
- Geographic cache distribution strategies

## Resource Relationships

Many examples show how resources work together:

```
Zone
├── Records (DNS entries)
├── LoadBalancer
│   ├── LoadBalancerPool
│   └── LoadBalancerMonitor
├── Ruleset (Security rules)
├── CacheRule (Performance)
└── Application (Spectrum)
```

## Prerequisites

- Crossplane installed in your Kubernetes cluster
- Cloudflare provider installed and configured
- Valid Cloudflare API token with appropriate permissions
- Cloudflare zone(s) under your control

## Usage Notes

- Replace placeholder values (like `your-zone-id-here`) with actual values
- Ensure proper RBAC permissions for Crossplane service account
- Monitor resource status using `kubectl describe` for troubleshooting
- Check provider logs for detailed error information

For more detailed documentation, refer to the [Cloudflare API documentation](https://developers.cloudflare.com/api/).
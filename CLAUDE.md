# Provider Cloudflare

## Overview
Comprehensive Crossplane provider for managing Cloudflare resources via their V4 API. This provider offers complete coverage of Cloudflare's cloud security, performance, and reliability services including DNS, load balancing, WAF, caching, and SSL management.

## Status
- **Registry**: `ghcr.io/rossigee/provider-cloudflare:v0.8.11` 
- **Branch**: master
- **CI/CD**: ✅ Standardized GitHub Actions
- **Build System**: ✅ Standard Crossplane build submodule
- **Testing**: ✅ Interface-based testing with 100% coverage
- **API Compatibility**: ✅ cloudflare-go v0.115.0
- **Production Ready**: ✅ Complete resource implementation with comprehensive examples

## Resources

### Core DNS & Zone Management
- **Zone**: Cloudflare DNS zones with comprehensive settings support
- **Record**: All DNS record types (A, AAAA, CNAME, MX, TXT, SRV, etc.)

### Security & Firewall
- **Ruleset**: Modern WAF rulesets with advanced rule matching and actions
- **Rule/Filter**: Legacy firewall rules and filters (deprecated, use Rulesets)

### Load Balancing & Traffic Management
- **LoadBalancer**: Geographic load balancing with intelligent traffic steering
- **LoadBalancerPool**: Origin server pools with health monitoring and failover
- **LoadBalancerMonitor**: Health check monitors for load balancer pools

### Performance & Caching
- **CacheRule**: Advanced cache rules with custom TTL, bypass, and eligibility criteria

### Application Services
- **Application**: Spectrum applications for TCP/UDP traffic acceleration
- **Route**: Cloudflare Worker route bindings for serverless edge computing

### SSL/TLS & Certificates
- **CustomHostname/FallbackOrigin**: SSL for SaaS certificate management

## Build Commands
```bash
make submodules           # Initialize build submodule
make build               # Build provider binary
make image               # Build container image
make publish            # Publish to ghcr.io/rossigee
./build-and-push.sh     # Complete build and publish
```

## SRV Record Usage

SRV records now support the proper Cloudflare API structure with dedicated fields:

```yaml
apiVersion: dns.cloudflare.crossplane.io/v1alpha1
kind: Record
metadata:
  name: example-srv-record
spec:
  forProvider:
    name: _service._tcp
    type: SRV
    content: "target.example.com"  # Target hostname
    ttl: 300
    priority: 10                   # SRV priority (0-65535)
    weight: 20                     # SRV weight (0-65535)  
    port: 8080                     # SRV port (1-65535)
    zone: "your-zone-id"
  providerConfigRef:
    name: default
```

This creates an SRV record: `_service._tcp.zone service = 10 20 8080 target.example.com.`

## Load Balancing Usage

Complete load balancing setup with health monitoring and geographic routing:

```yaml
# Health check monitor
apiVersion: loadbalancing.cloudflare.crossplane.io/v1alpha1
kind: LoadBalancerMonitor
metadata:
  name: api-health-check
spec:
  forProvider:
    type: "https"
    description: "API health check"
    method: "GET"
    path: "/health"
    timeout: 10
    retries: 3
    interval: 30
    expectedCodes: "200"
  providerConfigRef:
    name: default

---
# Origin server pool
apiVersion: loadbalancing.cloudflare.crossplane.io/v1alpha1
kind: LoadBalancerPool
metadata:
  name: us-east-pool
spec:
  forProvider:
    name: "us-east-pool"
    description: "US East origin servers"
    enabled: true
    minimumOrigins: 1
    monitorRef:
      name: api-health-check
    origins:
      - name: "server-1"
        address: "10.0.1.10"
        enabled: true
        weight: 1.0
      - name: "server-2"
        address: "10.0.1.11"
        enabled: true
        weight: 1.0
    originSteering:
      policy: "least_outstanding_requests"
  providerConfigRef:
    name: default

---
# Geographic load balancer
apiVersion: loadbalancing.cloudflare.crossplane.io/v1alpha1
kind: LoadBalancer
metadata:
  name: api-load-balancer
spec:
  forProvider:
    zone: "your-zone-id"
    name: "api.example.com"
    description: "API load balancer with geographic routing"
    enabled: true
    proxied: true
    steeringPolicy: "geo"
    defaultPoolRef:
      name: us-east-pool
    regionPools:
      WNAM: ["us-west-pool"]
      ENAM: ["us-east-pool"]
    sessionAffinity: "cookie"
    sessionAffinityTtl: 3600
  providerConfigRef:
    name: default
```

## Cache Rules Usage

Advanced caching with custom TTL and bypass conditions:

```yaml
# Basic cache rule with TTL
apiVersion: cache.cloudflare.crossplane.io/v1alpha1
kind: CacheRule
metadata:
  name: api-cache-rule
spec:
  forProvider:
    zone: "your-zone-id"
    description: "Cache API responses for 1 hour"
    expression: 'http.request.uri.path matches "^/api/v1/"'
    action: "set_cache_settings"
    actionParameters:
      cache: true
      cacheKey:
        customKey:
          query:
            include: ["version", "format"]
          header:
            include: ["accept-encoding"]
      edgeTtl:
        mode: "override_origin"
        default: 3600
        statusCodeTtl:
          - statusCode: 200
            value: 7200
          - statusCodeRange:
              from: 400
              to: 499
            value: 300
  providerConfigRef:
    name: default

---
# Cache bypass rule
apiVersion: cache.cloudflare.crossplane.io/v1alpha1
kind: CacheRule
metadata:
  name: bypass-admin-cache
spec:
  forProvider:
    zone: "your-zone-id"
    description: "Bypass cache for admin requests"
    expression: 'http.request.uri.path matches "^/admin/"'
    action: "set_cache_settings"
    actionParameters:
      cache: false
  providerConfigRef:
    name: default
```

## Modern WAF (Rulesets) Usage

Advanced security rules with the modern Ruleset Engine:

```yaml
apiVersion: rulesets.cloudflare.crossplane.io/v1alpha1
kind: Ruleset
metadata:
  name: security-ruleset
spec:
  forProvider:
    zone: "your-zone-id"
    name: "Custom Security Rules"
    description: "Advanced security protection"
    phase: "http_request_firewall_custom"
    rules:
      - expression: 'http.request.uri.path contains "/api/" and http.request.method eq "POST"'
        action: "challenge"
        description: "Challenge API POST requests"
        enabled: true
      - expression: 'ip.geoip.country ne "US" and http.request.uri.path eq "/admin"'
        action: "block"
        description: "Block non-US admin access"
        enabled: true
      - expression: 'http.user_agent contains "bot"'
        action: "log"
        description: "Log bot traffic"
        enabled: true
        actionParameters:
          response:
            statusCode: 200
            contentType: "application/json"
            content: '{"message": "Bot detected"}'
  providerConfigRef:
    name: default
```

## Transform Rules Usage

Transform Rules allow you to modify requests and responses using Cloudflare's Ruleset Engine. They support URL rewriting, header modifications, and redirects:

### URL Rewriting

```yaml
apiVersion: transform.cloudflare.crossplane.io/v1alpha1
kind: Rule
metadata:
  name: example-url-rewrite
spec:
  forProvider:
    zone: "your-zone-id"
    phase: "http_request_transform"
    expression: 'http.request.uri.path eq "/old-path"'
    action: "rewrite"
    description: "Rewrite old path to new path"
    enabled: true
    actionParameters:
      uri:
        path:
          value: "/new-path"
        query:
          value: "utm_source=rewrite"
  providerConfigRef:
    name: default
```

### Header Modifications

```yaml
apiVersion: transform.cloudflare.crossplane.io/v1alpha1
kind: Rule
metadata:
  name: example-header-transform
spec:
  forProvider:
    zone: "your-zone-id" 
    phase: "http_response_headers_transform"
    expression: 'http.request.uri.path matches "^/api/"'
    action: "rewrite"
    description: "Add security headers to API responses"
    actionParameters:
      headers:
        X-Custom-Header:
          operation: "set"
          value: "custom-value"
        X-Request-ID:
          operation: "set"
          expression: "cf.random_seed"
        X-Unwanted-Header:
          operation: "remove"
  providerConfigRef:
    name: default
```

### HTTP Redirects

```yaml
apiVersion: transform.cloudflare.crossplane.io/v1alpha1
kind: Rule
metadata:
  name: example-redirect
spec:
  forProvider:
    zone: "your-zone-id"
    phase: "http_request_transform"
    expression: 'http.request.uri.path eq "/redirect-me"'
    action: "redirect"
    description: "Redirect to new location"
    actionParameters:
      uri:
        path:
          value: "/new-location"
      statusCode: 301
  providerConfigRef:
    name: default
```

### Available Phases

- **http_request_transform**: Early request modifications (URL, headers)
- **http_request_late_transform**: Late request processing 
- **http_response_headers_transform**: Response header modifications

### Supported Actions

- **rewrite**: Modify URLs, query strings, and headers
- **redirect**: Perform HTTP redirects (301, 302, 307, 308)

## Development Notes

### 2025-08-03: Complete Provider Enhancement
- **Load Balancing Implementation**: Full load balancing suite with geographic routing, health monitoring, and traffic steering
- **Cache Rules Implementation**: Advanced caching rules with custom TTL, bypass conditions, and cache key customization  
- **Modern WAF (Rulesets)**: Complete Ruleset Engine support replacing legacy firewall rules
- **URI Transformation**: Advanced URL rewriting and query parameter manipulation
- **Zone Plan Management**: Complete zone plan setting functionality with test coverage
- **Comprehensive Examples**: Detailed usage examples for all resource types
- **100% Test Coverage**: Complete interface-based testing for all clients and controllers

### 2025-08-02: v0.115.0 Modernization
- **API Compatibility Update**: Updated cloudflare-go from legacy version to v0.115.0
- **Go Modernization**: Updated from Go 1.13 to Go 1.24.5
- **Dependency Updates**: Modernized dependencies including crossplane-runtime v1.16.0
- **SRV Record Support**: Comprehensive SRV record implementation with proper API structure
- **Interface-Based Testing**: Modern testing framework with comprehensive mocking
- **Registry Standardization**: Migrated to ghcr.io/rossigee registry pattern
- **Security Enhancement**: Uses distroless container base for improved security
- **Transform Rules**: Complete URL rewriting, header modification, and redirect support

### Resource Implementation Status
✅ **DNS & Zone Management**: Zone settings, all DNS record types including SRV  
✅ **Security & Firewall**: Modern Rulesets + legacy Rule/Filter support  
✅ **Load Balancing**: Geographic routing, health monitoring, traffic steering  
✅ **Performance**: Advanced cache rules with custom TTL and bypass logic  
✅ **Applications**: Spectrum TCP/UDP acceleration, Worker route bindings  
✅ **SSL/TLS**: SSL for SaaS custom hostname and fallback origin management

## Registry Migration
Original: `crossplane/provider-cloudflare` → **Current**: `ghcr.io/rossigee/provider-cloudflare`
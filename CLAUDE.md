# Provider Cloudflare

## Overview
Comprehensive Crossplane provider for managing Cloudflare resources via their V4 API. This provider offers complete coverage of Cloudflare's cloud security, performance, and reliability services including DNS, load balancing, WAF, caching, SSL management, Zero Trust Access, tunneling, and device management.

## Status
- **Registry**: `ghcr.io/rossigee/provider-cloudflare:v0.13.0`
- **Branch**: master
- **CI/CD**: ✅ Standardized GitHub Actions with "CI Builds, Release Publishes" pattern
- **Build System**: ✅ Standard Crossplane build submodule
- **Testing**: ✅ Interface-based testing with comprehensive coverage
- **API Compatibility**: ✅ cloudflare-go v0.115.0
- **Production Ready**: ✅ Complete resource implementation with comprehensive examples

## Resources

### Core DNS & Zone Management
- **Zone**: Cloudflare DNS zones with comprehensive settings support
- **Record**: All DNS record types (A, AAAA, CNAME, MX, TXT, SRV, etc.)

### Security & Firewall
- **Ruleset**: Modern WAF rulesets with advanced rule matching and actions
- **Rule/Filter**: Legacy firewall rules and filters (deprecated, use Rulesets)
- **AccessApplication**: Zero Trust access applications with authentication policies
- **DevicePostureRule**: Device posture rules for endpoint security

### Load Balancing & Traffic Management
- **LoadBalancer**: Geographic load balancing with intelligent traffic steering
- **LoadBalancerPool**: Origin server pools with health monitoring and failover
- **LoadBalancerMonitor**: Health check monitors for load balancer pools

### Performance & Caching
- **CacheRule**: Advanced cache rules with custom TTL, bypass, and eligibility criteria

### Application Services
- **Application**: Spectrum applications for TCP/UDP traffic acceleration
- **Script**: Cloudflare Worker scripts for serverless edge computing
- **CronTrigger**: Scheduled execution triggers for Worker scripts
- **Domain**: Custom domain attachments for Workers
- **KVNamespace**: Key-Value storage namespaces for Workers
- **Route**: URL route bindings for Worker scripts
- **Subdomain**: Custom subdomain configuration for Workers
- **Tunnel**: Cloudflare Tunnel (Cloudflared) connections for secure access

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
apiVersion: dns.cloudflare.m.crossplane.io/v1beta1
kind: Record
metadata:
  namespace: default
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
apiVersion: loadbalancing.cloudflare.m.crossplane.io/v1beta1
kind: LoadBalancerMonitor
metadata:
  namespace: default
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
apiVersion: loadbalancing.cloudflare.m.crossplane.io/v1beta1
kind: LoadBalancerPool
metadata:
  namespace: default
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
apiVersion: loadbalancing.cloudflare.m.crossplane.io/v1beta1
kind: LoadBalancer
metadata:
  namespace: default
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
apiVersion: cache.cloudflare.m.crossplane.io/v1beta1
kind: CacheRule
metadata:
  namespace: default
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
apiVersion: cache.cloudflare.m.crossplane.io/v1beta1
kind: CacheRule
metadata:
  namespace: default
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
apiVersion: rulesets.cloudflare.m.crossplane.io/v1beta1
kind: Ruleset
metadata:
  namespace: default
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
apiVersion: transform.cloudflare.m.crossplane.io/v1beta1
kind: Rule
metadata:
  namespace: default
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
apiVersion: transform.cloudflare.m.crossplane.io/v1beta1
kind: Rule
metadata:
  namespace: default
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
apiVersion: transform.cloudflare.m.crossplane.io/v1beta1
kind: Rule
metadata:
  namespace: default
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

### 2025-10-20: Complete Worker Resources Implementation
- **Worker Controllers Enabled**: All worker resource controllers now fully functional
- **Cloudflare API Integration**: Real API calls for Cron Triggers, Domains, KV Namespaces, Routes, and Subdomains
- **Complete Worker Ecosystem**: Full support for serverless edge computing with all Cloudflare Worker APIs
- **Production Ready**: Worker resources ready for enterprise deployment

### 2026-01-05: Enterprise Features Implementation
- **Zero Trust Access**: Complete Cloudflare Access implementation with applications and policies
- **Cloudflare Tunnel**: Full Argo Tunnel management for secure remote access
- **Device Management**: Device posture rules for endpoint security and compliance
- **Enhanced Security**: Enterprise-grade security features for modern infrastructure

### 2025-08-03: Complete Provider Enhancement
- **Load Balancing Implementation**: Full load balancing suite with geographic routing, health monitoring, and traffic steering
- **Cache Rules Implementation**: Advanced caching rules with custom TTL, bypass conditions, and cache key customization
- **Modern WAF (Rulesets)**: Complete Ruleset Engine support replacing legacy firewall rules
- **URI Transformation**: Advanced URL rewriting and query parameter manipulation
- **Zone Plan Management**: Complete zone plan setting functionality with test coverage
- **Comprehensive Examples**: Detailed usage examples for all resource types
- **Interface-Based Testing**: Modern testing framework with comprehensive mocking

### 2025-08-02: v0.115.0 Modernization
- **API Compatibility Update**: Updated cloudflare-go from legacy version to v0.115.0
- **Go Modernization**: Updated from Go 1.13 to Go 1.25.3
- **Dependency Updates**: Modernized dependencies including crossplane-runtime v1.16.0
- **SRV Record Support**: Comprehensive SRV record implementation with proper API structure
- **Interface-Based Testing**: Modern testing framework with comprehensive mocking
- **Registry Standardization**: Migrated to ghcr.io/rossigee registry pattern
- **Security Enhancement**: Uses distroless container base for improved security
- **Transform Rules**: Complete URL rewriting, header modification, and redirect support

## Worker Resources Usage

### Worker Script

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Script
metadata:
  namespace: default
  name: my-worker-script
spec:
  forProvider:
    scriptName: "my-worker"
    script: |
      addEventListener('fetch', event => {
        event.respondWith(new Response('Hello from Cloudflare Worker!'))
      })
    bindings:
      - name: "MY_KV"
        type: "kv_namespace"
        namespaceId: "your-kv-namespace-id"
  providerConfigRef:
    name: default
```

### Worker Cron Trigger

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: CronTrigger
metadata:
  namespace: default
  name: daily-backup-trigger
spec:
  forProvider:
    scriptName: "backup-worker"
    cron: "0 2 * * *"  # Daily at 2 AM
  providerConfigRef:
    name: default
```

### Worker KV Namespace

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: KVNamespace
metadata:
  namespace: default
  name: app-config-namespace
spec:
  forProvider:
    title: "application-configuration"
  providerConfigRef:
    name: default
```

### Worker Route

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Route
metadata:
  namespace: default
  name: api-route
spec:
  forProvider:
    zone: "your-zone-id"
    pattern: "api.example.com/*"
    script: "my-api-worker"
  providerConfigRef:
    name: default
```

### Worker Domain

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Domain
metadata:
  namespace: default
  name: custom-worker-domain
spec:
  forProvider:
    accountId: "your-account-id"
    zoneId: "your-zone-id"
    hostname: "workers.example.com"
    service: "my-worker-script"
    environment: "production"
  providerConfigRef:
    name: default
```

### Worker Subdomain

```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Subdomain
metadata:
  namespace: default
  name: worker-subdomain
spec:
  forProvider:
    accountId: "your-account-id"
    name: "mycompany"
  providerConfigRef:
    name: default
```

## Access Application Usage

Zero Trust access applications for secure application access with authentication policies:

```yaml
apiVersion: access.cloudflare.m.crossplane.io/v1beta1
kind: AccessApplication
metadata:
  namespace: default
  name: admin-dashboard
spec:
  forProvider:
    accountId: "your-account-id"
    name: "Admin Dashboard"
    domain: "admin.example.com"
    type: "self_hosted"
    sessionDuration: "24h"
    allowedIdps: ["your-idp-id"]
    autoRedirectToIdentity: false
    enableBindingCookie: false
    httpOnlyCookieAttribute: true
    sameSiteCookieAttribute: "strict"
    skipInterstitial: false
    appLauncherVisible: true
    policies:
      - id: "policy-id-1"
        precedence: 1
      - id: "policy-id-2"
        precedence: 2
    corsHeaders:
      allowedMethods: ["GET", "POST"]
      allowedOrigins: ["https://trusted.example.com"]
      allowCredentials: true
      maxAge: 86400
    logoUrl: "https://example.com/logo.png"
    tags: ["admin", "internal"]
  providerConfigRef:
    name: default
```

## Device Posture Rule Usage

Device posture rules for endpoint security and compliance:

```yaml
apiVersion: device.cloudflare.m.crossplane.io/v1beta1
kind: DevicePostureRule
metadata:
  namespace: default
  name: corporate-device-rule
spec:
  forProvider:
    accountId: "your-account-id"
    name: "Corporate Devices"
    type: "os_version"
    description: "Ensure devices are running approved OS versions"
    schedule: "5m"
    expiration: "24h"
    input:
      os: "windows"
      operator: ">="
      version: "10.0.19043"
      osDistroName: "ubuntu"
      osDistroRevision: "20.04"
    match:
      platform: "windows"
  providerConfigRef:
    name: default
```

## Cloudflare Tunnel Usage

Argo Tunnel connections for secure remote access:

```yaml
apiVersion: tunnel.cloudflare.m.crossplane.io/v1beta1
kind: Tunnel
metadata:
  namespace: default
  name: office-tunnel
spec:
  forProvider:
    accountId: "your-account-id"
    name: "Office Tunnel"
    tunnelSecret: "base64-encoded-secret"
    configSrc: "local"
    metadata:
      "environment": "production"
      "location": "office"
  providerConfigRef:
    name: default
```

### Resource Implementation Status
✅ **DNS & Zone Management**: Zone settings, all DNS record types including SRV
✅ **Security & Firewall**: Modern Rulesets + legacy Rule/Filter support
✅ **Load Balancing**: Geographic routing, health monitoring, traffic steering
✅ **Performance**: Advanced cache rules with custom TTL and bypass logic
✅ **Applications**: Spectrum TCP/UDP acceleration
✅ **Workers**: Complete worker ecosystem with scripts, cron triggers, domains, KV storage, routes, and subdomains
✅ **SSL/TLS**: SSL for SaaS custom hostname and fallback origin management
✅ **Zero Trust Access**: Access applications with authentication policies
✅ **Device Management**: Device posture rules for endpoint security
✅ **Tunneling**: Cloudflare Tunnel connections for secure access

## Registry Migration
Original: `crossplane/provider-cloudflare` → **Current**: `ghcr.io/rossigee/provider-cloudflare`
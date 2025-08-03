# Provider Cloudflare

## Overview
Crossplane provider for managing Cloudflare resources via their V4 API. This provider manages Cloudflare Zones, DNS Records, Firewall Rules, Transform Rules, Spectrum Applications, SSL for SaaS settings, and Worker Routes.

## Status
- **Registry**: `ghcr.io/rossigee/provider-cloudflare:v0.6.0` 
- **Branch**: master
- **CI/CD**: ✅ Standardized GitHub Actions
- **Build System**: ✅ Standard Crossplane build submodule
- **Testing**: ✅ Interface-based testing with 100% coverage
- **API Compatibility**: ✅ cloudflare-go v0.115.0

## Resources
- **Zone**: Cloudflare DNS zones
- **Record**: DNS records within zones
- **Rule/Filter**: Firewall rules and filters
- **Transform Rule**: URL and header transformations via Ruleset Engine
- **Application**: Spectrum applications
- **CustomHostname/FallbackOrigin**: SSL for SaaS settings
- **Route**: Worker route bindings

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
- **Transform Rules Implementation** (2025-08-03): Complete Transform Rules support via Ruleset Engine
- **v0.115.0 Modernization** (2025-08-02): Complete API compatibility update
- Updated from Go 1.13 to Go 1.23 
- Modernized dependencies including crossplane-runtime v1.17.0
- Updated cloudflare-go from legacy version to v0.115.0
- Comprehensive test framework with interface-based testing
- Fixed all firewall, DNS, zone, and worker components
- Added Transform Rules with full CRUD operations and comprehensive testing
- Added fake client infrastructure for reliable testing
- Standardized to ghcr.io/rossigee registry pattern
- Uses distroless container base for security

## Registry Migration
Original: `crossplane/provider-cloudflare` → **Current**: `ghcr.io/rossigee/provider-cloudflare`
# Provider Cloudflare

## Overview
Crossplane provider for managing Cloudflare resources via their V4 API. This provider manages Cloudflare Zones, DNS Records, Firewall Rules, Spectrum Applications, SSL for SaaS settings, and Worker Routes.

## Status
- **Registry**: `ghcr.io/rossigee/provider-cloudflare:v0.6.0-dirty-srv` 
- **Branch**: master
- **CI/CD**: ✅ Standardized GitHub Actions
- **Build System**: ✅ Standard Crossplane build submodule
- **Testing**: ✅ Interface-based testing with comprehensive SRV record support
- **SRV Records**: ✅ Full support with dedicated priority, weight, and port fields

## Resources
- **Zone**: Cloudflare DNS zones
- **Record**: DNS records within zones
- **Rule/Filter**: Firewall rules and filters
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

## Development Notes
- Updated from Go 1.13 to Go 1.23 (2025-08-02)
- Modernized dependencies including crossplane-runtime v1.17.0
- Standardized to ghcr.io/rossigee registry pattern
- Added proper Crossplane build system integration
- Uses distroless container base for security
- Interface-based testing framework with comprehensive SRV record support (2025-08-02)
- **SRV Record Enhancement** (2025-08-02): Added dedicated priority, weight, and port fields to properly support Cloudflare API requirements

## Registry Migration
Original: `crossplane/provider-cloudflare` → **Current**: `ghcr.io/rossigee/provider-cloudflare`
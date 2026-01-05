# provider-cloudflare

`provider-cloudflare` is a [Crossplane](https://crossplane.io/) Provider
that manages Cloudflare resources via their V4 API (`cloudflare-go`). It provides
comprehensive coverage of Cloudflare's cloud security, performance, and reliability services
including enterprise Zero Trust, tunneling, and device management.

## Resources

### DNS & Zone Management
- **`Zone`** - Manages Cloudflare DNS zones with comprehensive settings support
- **`Record`** - Manages DNS records (A, AAAA, CNAME, MX, TXT, SRV, etc.) within zones

### Security & Firewall
- **`Ruleset`** - Modern WAF rulesets with advanced rule matching and actions (replaces legacy firewall rules)
- **`Rule`** & **`Filter`** - Legacy firewall rules and filters (deprecated, use Rulesets instead)
- **`AccessApplication`** - Zero Trust access applications with authentication policies
- **`DevicePostureRule`** - Device posture rules for endpoint security

### Load Balancing & Traffic Management  
- **`LoadBalancer`** - Geographic load balancing with intelligent traffic steering
- **`LoadBalancerPool`** - Origin server pools with health monitoring and failover
- **`LoadBalancerMonitor`** - Health check monitors for load balancer pools

### Performance & Caching
- **`CacheRule`** - Advanced cache rules with custom TTL, bypass, and eligibility criteria

### Applications & Services
- **`Application`** - Spectrum applications for TCP/UDP traffic acceleration
- **`Script`** - Cloudflare Worker scripts for serverless edge computing
- **`CronTrigger`** - Scheduled execution triggers for Worker scripts
- **`Domain`** - Custom domain attachments for Workers
- **`KVNamespace`** - Key-Value storage namespaces for Workers
- **`Route`** - URL route bindings for Worker scripts
- **`Subdomain`** - Custom subdomain configuration for Workers
- **`Tunnel`** - Cloudflare Tunnel (Cloudflared) connections for secure access

### SSL/TLS & Certificates
- **`CustomHostname`** & **`FallbackOrigin`** - SSL for SaaS certificate management

## Features

✅ **Crossplane v2 Native** - Full v2 migration complete with namespaced resources
✅ **v1beta1 APIs Only** - Modern namespaced APIs with `.m.` group naming (v1alpha1 removed)
✅ **Enterprise Features** - Zero Trust Access, Cloudflare Tunnel, and Device Management
✅ **Interface-Based Testing** - Modern testing framework with comprehensive mocking
✅ **Production Ready** - Used in production environments with proven reliability
✅ **Modern Go** - Updated to Go 1.25.5 with latest dependencies
✅ **Comprehensive Examples** - Detailed usage examples for all resource types
✅ **Advanced Capabilities** - Support for complex scenarios like geographic routing, traffic steering, and advanced caching

## Status

- **Registry**: `ghcr.io/rossigee/provider-cloudflare:v0.14.0`
- **Build Status**: ✅ `make lint` and `make test` passing
- **Crossplane v2**: ✅ Fully migrated to v2-native architecture
- **API Compatibility**: ✅ cloudflare-go v0.115.0, Go 1.25.5
- **Production Ready**: ✅ Complete resource implementation with v1beta1 namespaced resources

See [Current Status](docs/CURRENT-STATUS.md) for detailed technical information.

## Crossplane v2 Migration Complete

This provider has **fully migrated to Crossplane v2** with namespaced v1beta1 APIs only (v1alpha1 cluster-scoped APIs have been removed):

### ✅ Available Namespaced APIs (v1beta1)
- **DNS & Zones** - `dns.cloudflare.m.crossplane.io/v1beta1`, `zone.cloudflare.m.crossplane.io/v1beta1`
- **Load Balancing** - `loadbalancing.cloudflare.m.crossplane.io/v1beta1`
- **Security & Access** - `firewall.cloudflare.m.crossplane.io/v1beta1`, `security.cloudflare.m.crossplane.io/v1beta1`, `access.cloudflare.m.crossplane.io/v1beta1`, `device.cloudflare.m.crossplane.io/v1beta1`
- **Performance** - `cache.cloudflare.m.crossplane.io/v1beta1`
- **Edge Computing** - `workers.cloudflare.m.crossplane.io/v1beta1`, `spectrum.cloudflare.m.crossplane.io/v1beta1`, `tunnel.cloudflare.m.crossplane.io/v1beta1`
- **SSL/TLS** - `ssl.cloudflare.m.crossplane.io/v1beta1`, `sslsaas.cloudflare.m.crossplane.io/v1beta1`, `originssl.cloudflare.m.crossplane.io/v1beta1`
- **Advanced** - `transform.cloudflare.m.crossplane.io/v1beta1`, `logpush.cloudflare.m.crossplane.io/v1beta1`, `emailrouting.cloudflare.m.crossplane.io/v1beta1`, `r2.cloudflare.m.crossplane.io/v1beta1`

### v2 Benefits
- 🎯 **Namespace Isolation** - All resources scoped to Kubernetes namespaces
- 🔐 **Enhanced RBAC** - Namespace-level permissions and access control
- 🏗️ **Modern Patterns** - Full Crossplane v2 compliance and future-proof architecture
- ⚡ **API Group Naming** - Consistent `.m.` naming convention for namespaced resources

## Installation

Install the provider in your Crossplane cluster:

```bash
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-cloudflare
spec:
  package: ghcr.io/rossigee/provider-cloudflare:v0.13.0
EOF
```

## Configuration

Create a ProviderConfig with your Cloudflare credentials:

```bash
# Create secret with API token
kubectl create secret generic cloudflare-secret \
  --from-literal=token="your-cloudflare-api-token"

# Create ProviderConfig
kubectl apply -f - <<EOF
apiVersion: cloudflare.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: cloudflare-secret
      key: token
EOF
```

## Usage Examples

### DNS Zone Management

```yaml
apiVersion: zone.cloudflare.m.crossplane.io/v1beta1
kind: Zone
metadata:
  name: example-zone
  namespace: default  # Namespaced resource in v1beta1
spec:
  forProvider:
    name: "example.com"
    type: "full"  # full, partial, or secondary
    jumpStart: false
  providerConfigRef:
    name: default
```

### Load Balancer with Geographic Routing

```yaml
apiVersion: loadbalancing.cloudflare.m.crossplane.io/v1beta1
kind: LoadBalancer
metadata:
  name: api-load-balancer
  namespace: default  # Namespaced resource in v1beta1
spec:
  forProvider:
    zone: "your-zone-id"
    name: "api.example.com"
    enabled: true
    proxied: true
    steeringPolicy: "geo"
    regionPools:
      WNAM: ["us-west-pool"]
      ENAM: ["us-east-pool"]
  providerConfigRef:
    name: default
```

### DNS Record

```yaml
apiVersion: dns.cloudflare.m.crossplane.io/v1beta1
kind: Record
metadata:
  name: example-a-record
  namespace: default  # Namespaced resource in v1beta1
spec:
  forProvider:
    zone: "your-zone-id"
    name: "www"
    type: "A"
    content: "192.0.2.1"
    ttl: 300
    proxied: true
  providerConfigRef:
    name: default
```

For comprehensive examples covering all resource types including enterprise Zero Trust and device management, see the **[examples/](examples/)** directory with detailed usage scenarios.

## Developing

Run against a Kubernetes cluster:

```console
make run
```

Install `latest` into Kubernetes cluster where Crossplane is installed:

```console
make install
```

Install local build into [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
cluster where Crossplane is installed:

```console
make install-local
```

Build, push, and install:

```console
make all
```

Build image:

```console
make image
```

Push image:

```console
make push
```

Build binary:

```console
make build
```

## Testing

Run the full test suite:

```console
make test
```

Run tests with coverage:

```console
make test-coverage
```

Run linting:

```console
make lint
```

## Architecture

This provider follows Crossplane's provider architecture:

- **API Types** (`apis/`) - Define Kubernetes CRDs for Cloudflare resources
- **Controllers** (`internal/controller/`) - Reconcile desired state with Cloudflare API
- **Clients** (`internal/clients/`) - Abstracted Cloudflare API interactions
- **Examples** (`examples/`) - Real-world usage examples
- **Package** (`package/`) - Generated CRDs and metadata

## Supported Cloudflare APIs

- **Zones API** - Zone management and settings
- **DNS API** - All DNS record types including SRV records
- **Load Balancing API** - Geographic load balancing and health monitoring  
- **Rulesets API** - Modern WAF and transformation rules
- **Cache API** - Advanced cache rule configuration
- **Spectrum API** - TCP/UDP application acceleration
- **Workers API** - Serverless edge computing routes
- **Access API** - Zero Trust access applications and policies
- **Tunnel API** - Cloudflare Tunnel (Cloudflared) management
- **Device Posture API** - Endpoint security and device management
- **SSL for SaaS API** - Custom certificate management

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run `make test lint` to verify
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

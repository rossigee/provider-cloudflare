# Cloudflare Provider v2 Migration Guide

## Overview

The Cloudflare provider now supports **Crossplane v2** patterns with namespaced resources alongside existing cluster-scoped resources. This enables better multi-tenancy, RBAC isolation, and modern Kubernetes resource management.

## Current v2 Support Status

### ✅ Available v1beta1 Namespaced APIs (7/20)

| API Group | v1alpha1 (Cluster) | v1beta1 (Namespaced) | Migration Status |
|-----------|--------------------|-----------------------|------------------|
| **DNS** | ✅ | ✅ | Complete |
| **Load Balancing** | ✅ | ✅ | Complete |
| **Rulesets** | ✅ | ✅ | Complete |
| **Zone** | ✅ | ✅ | Complete |
| **Cache** | ✅ | ✅ | **NEW** |
| **Workers** | ✅ | ✅ | **NEW** |
| **Security** | ✅ | ✅ | **NEW** |

### 🔄 Remaining v1alpha1 Only APIs (13/20)

Email Routing, Firewall, Logpush, Origin SSL, R2, Spectrum, SSL, SSL SaaS, Transform, and others remain cluster-scoped only.

## Key Benefits of v1beta1 APIs

### 🎯 **Namespace Isolation**
- Resources scoped to Kubernetes namespaces
- Better multi-tenancy support
- Team/environment separation

### 🔐 **Enhanced RBAC**
- Namespace-level permissions
- Fine-grained access control
- Reduced blast radius

### 🏗️ **Modern Patterns**
- Crossplane v2 compliance
- Future-proof architecture
- Industry standard practices

## Migration Examples

### Cache Rules: v1alpha1 → v1beta1

**Before (Cluster-scoped):**
```yaml
apiVersion: cache.cloudflare.m.crossplane.io/v1beta1
kind: CacheRule
metadata:
  name: api-cache-rule  # No namespace
spec:
  forProvider:
    zone: "your-zone-id"
    expression: 'http.request.uri.path matches "^/api/"'
    action: "set_cache_settings"
```

**After (Namespaced):**
```yaml
apiVersion: cache.cloudflare.m.crossplane.io/v1beta1
kind: CacheRule
metadata:
  name: api-cache-rule
  namespace: production  # Namespace isolation
spec:
  forProvider:
    zone: "your-zone-id"
    expression: 'http.request.uri.path matches "^/api/"'
    action: "set_cache_settings"
    actionParameters:
      cache: true
      edgeTtl:
        mode: "override_origin"
        default: 3600
```

### Workers: v1alpha1 → v1beta1

**Before (Cluster-scoped):**
```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Script
metadata:
  name: edge-worker
spec:
  forProvider:
    scriptName: "edge-worker"
    script: "// Worker code here"
```

**After (Namespaced):**
```yaml
apiVersion: workers.cloudflare.m.crossplane.io/v1beta1
kind: Script
metadata:
  name: edge-worker
  namespace: edge-services
spec:
  forProvider:
    scriptName: "edge-worker"
    script: "// Worker code here"
    module: true
    compatibilityDate: "2025-01-01"
    bindings:
      - type: "kv_namespace"
        name: "CACHE"
        namespaceId: "kv-namespace-id"
```

### Security Rate Limits: v1alpha1 → v1beta1

**Before (Cluster-scoped):**
```yaml
apiVersion: security.cloudflare.m.crossplane.io/v1beta1
kind: RateLimit
metadata:
  name: api-limit
spec:
  forProvider:
    zone: "zone-id"
    threshold: 100
    period: 60
```

**After (Namespaced):**
```yaml
apiVersion: security.cloudflare.m.crossplane.io/v1beta1
kind: RateLimit
metadata:
  name: api-limit
  namespace: security-policies
spec:
  forProvider:
    zone: "zone-id"
    threshold: 100
    period: 60
    match:
      request:
        methods: ["POST", "PUT"]
        url: "*api*"
    action:
      mode: "challenge"
      timeout: 300
```

## API Differences

### Key Changes in v1beta1

1. **API Group**: `*.cloudflare.crossplane.io` → `*.cloudflare.m.crossplane.io`
2. **Scope**: `scope=Cluster` → `scope=Namespaced`
3. **Enhanced Features**: More comprehensive parameter support
4. **Validation**: Improved field validation and defaults

### Enhanced Features in v1beta1

**Cache API:**
- Advanced TTL controls (edge, browser, status-code specific)
- Custom cache keys (query, header, cookie, user variations)
- Serve stale configurations
- Cache deception armor

**Workers API:**
- ES module support with compatibility settings
- Enhanced bindings (KV, R2, Durable Objects, etc.)
- Smart placement controls
- Tail consumer configuration
- Log push integration

**Security API:**
- Traffic matching (request/response criteria)
- Bypass rules for trusted sources
- Comprehensive action modes
- Correlation settings

## Migration Strategy

### 1. **Gradual Migration** (Recommended)
- Create new v1beta1 resources alongside existing v1alpha1
- Test in non-production namespaces
- Gradually migrate production workloads
- Deprecate v1alpha1 resources when ready

### 2. **Namespace Organization**
```yaml
# Production workloads
namespace: production

# Development environment
namespace: development

# Team-specific resources
namespace: team-frontend
namespace: team-backend

# Security policies
namespace: security-policies
```

### 3. **RBAC Configuration**
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: production
  name: cloudflare-manager
rules:
- apiGroups: ["cache.cloudflare.m.crossplane.io"]
  resources: ["cacherules"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
- apiGroups: ["workers.cloudflare.m.crossplane.io"]
  resources: ["scripts"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
```

## Best Practices

### 🎯 **Resource Organization**
- Use meaningful namespace names
- Group related resources together
- Implement consistent naming conventions

### 🔐 **Security**
- Apply principle of least privilege
- Use namespace-scoped RBAC
- Audit resource access regularly

### 📊 **Monitoring**
- Monitor resources per namespace
- Set up alerts for resource limits
- Track cross-namespace dependencies

### 🚀 **Performance**
- Cache frequently accessed resources
- Use resource quotas appropriately
- Monitor resource utilization

## Backward Compatibility

### ✅ **Guaranteed**
- All existing v1alpha1 resources continue working
- No breaking changes to existing functionality
- Both APIs can run simultaneously

### 📋 **Migration Timeline**
- **Phase 1**: Test v1beta1 APIs in development
- **Phase 2**: Create new resources using v1beta1
- **Phase 3**: Migrate production workloads gradually
- **Phase 4**: Deprecate v1alpha1 usage (optional)

## Troubleshooting

### Common Issues

**1. API Group Confusion**
```bash
# Wrong - v1alpha1 API group
kubectl get cacherules

# Correct - v1beta1 API group
kubectl get cacherules.cache.cloudflare.m.crossplane.io -n production
```

**2. Namespace Requirements**
```bash
# v1beta1 resources MUST specify namespace
kubectl apply -f cacherule.yaml -n production
```

**3. RBAC Permissions**
```bash
# Ensure roles include new API groups
kubectl auth can-i create cacherules.cache.cloudflare.m.crossplane.io -n production
```

## Future Roadmap

### 🎯 **Planned v1beta1 Migrations**
- Email Routing APIs
- Firewall Rules
- SSL/TLS Management
- R2 Storage
- Transform Rules
- Additional security features

### 🚀 **Enhanced Features**
- Cross-namespace resource references
- Advanced validation rules
- Improved status reporting
- Enhanced observability

## Support

For questions or issues with v2 migration:
- Review existing examples in `examples/*/v1beta1/`
- Check provider documentation
- Consult Crossplane v2 migration guides
- Open issues for bugs or feature requests

---

**Provider Version**: v0.9.1+
**Crossplane Compatibility**: v1.14+
**Kubernetes Compatibility**: v1.25+
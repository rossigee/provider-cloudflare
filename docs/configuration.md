# Configuration

Guide for configuring the Cloudflare provider.

## ProviderConfig

Create a ProviderConfig to configure connection settings:

```yaml
apiVersion: cloudflare.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: cloudflare-credentials
      namespace: crossplane-system
      key: token
```

## Authentication

The provider uses Cloudflare API tokens. Create a secret with your credentials:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-credentials
  namespace: crossplane-system
type: Opaque
stringData:
  token: your-api-token
```

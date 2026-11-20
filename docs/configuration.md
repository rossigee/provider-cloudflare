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
  # Required when the API token can see more than one Cloudflare account.
  # Account-scoped resources (Workers scripts, KV namespaces, cron triggers,
  # domains, subdomains) are managed in this account. Cloudflare does not
  # guarantee account ordering, so the provider never guesses accounts[0].
  accountID: c1b74f148aee28025816e104a92622c5
  credentials:
    source: Secret
    secretRef:
      name: cloudflare-credentials
      namespace: crossplane-system
      key: token
```

If `accountID` is omitted, the provider uses the account list only when the
token can see exactly one account. With multiple accounts it fails the
reconcile rather than managing resources in an arbitrary account.

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

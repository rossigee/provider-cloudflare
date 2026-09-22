# Getting Started

Guide to getting started with the Cloudflare provider.

## Installation

Install the Cloudflare provider:

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-cloudflare:v0.21.0
```

## Prerequisites

- Kubernetes cluster with Crossplane installed
- Cloudflare account with API key

## Quick Start

1. Create a ProviderConfig:

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

2. Create a DNS record:

```yaml
apiVersion: dns.cloudflare.m.crossplane.io/v1beta1
kind: Record
metadata:
  name: example-record
spec:
  forProvider:
    zone: your-zone-id
    name: example
    type: A
    content: 192.0.2.1
    proxied: true
  providerConfigRef:
    name: default
```

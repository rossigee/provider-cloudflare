# Pages Domain

Attach a custom domain to a Cloudflare Pages project.

## Example

```yaml
apiVersion: pages.cloudflare.m.crossplane.io/v1beta1
kind: Domain
metadata:
  namespace: default
  name: my-pages-domain
spec:
  forProvider:
    accountId: "your-account-id"
    projectName: "my-static-site"
    domain: "www.example.com"
  providerConfigRef:
    name: default
```

## Parameters

- `accountId` (required)
- `projectName` (required)
- `domain` (required): Custom domain to attach.

## Status

- `status`, `customDomain`, `verificationErrors`.
- `createdOn`, `modifiedOn`.

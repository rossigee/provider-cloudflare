# Pages Project

Manage Cloudflare Pages projects.

## Example

```yaml
apiVersion: pages.cloudflare.m.crossplane.io/v1beta1
kind: Project
metadata:
  namespace: default
  name: my-pages-project
spec:
  forProvider:
    accountId: "your-account-id"
    name: "my-static-site"
    productionBranch: "main"
    buildCommand: "npm run build"
    destinationDir: "dist"
    gitRepository:
      url: "https://github.com/your-org/my-site.git"
      branch: "main"
  providerConfigRef:
    name: default
```

## Parameters

- `accountId` (required): Cloudflare account ID.
- `name` (required): Project name.
- `productionBranch`: Main branch for production deploys.
- `buildCommand`, `destinationDir`, `rootDir`: Build settings.
- `environmentVariables`: Map of env vars.
- `gitRepository`: Git source configuration.

## Status

- `subdomain`, `domains`, `customDomains`: Assigned domains.
- `createdOn`, `modifiedOn`.

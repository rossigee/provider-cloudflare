# Pages Deployment

Trigger or manage a deployment for a Pages project.

## Example

```yaml
apiVersion: pages.cloudflare.m.crossplane.io/v1beta1
kind: Deployment
metadata:
  namespace: default
  name: my-pages-deployment
spec:
  forProvider:
    accountId: "your-account-id"
    projectName: "my-static-site"
    branch: "main"
  providerConfigRef:
    name: default
```

## Parameters

- `accountId` (required)
- `projectName` (required): Name of the Pages project.
- `branch`: Git branch to deploy.
- `environmentVariables`, `aliases`: Optional overrides.

## Status

- `url`, `shortId`, `environment`, `latestStage`.
- `createdOn`, `modifiedOn`.

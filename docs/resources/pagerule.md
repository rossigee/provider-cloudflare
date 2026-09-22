# Page Rule

Legacy Page Rules (consider Rulesets for new use).

## Example

```yaml
apiVersion: pagerules.cloudflare.m.crossplane.io/v1beta1
kind: PageRule
metadata:
  namespace: default
  name: example-page-rule
spec:
  forProvider:
    zoneId: "your-zone-id"
    targets:
      - target: "url"
        constraint:
          operator: "matches"
          value: "https://example.com/*"
    actions:
      - id: "forwarding_url"
        value: '{"url": "https://new.example.com/$1", "status_code": 301}'
    priority: 1
    status: "active"
  providerConfigRef:
    name: default
```

## Parameters

- `zoneId` (required)
- `targets`: Array of target + constraint (operator + value).
- `actions`: Array of action ID + value.
- `priority`, `status`.

## Status

Observation mirrors parameters plus timestamps.

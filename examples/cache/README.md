# Cloudflare Cache Rules Examples

This directory contains examples for Cloudflare Cache Rules, which provide fine-grained control over caching behavior for your website or application.

## Prerequisites

Before using these examples, ensure you have:
1. A Cloudflare account with an active zone
2. A configured Crossplane provider for Cloudflare
3. Your Zone ID from the Cloudflare dashboard

## Examples Overview

### 1. Basic Cache Rule (`basic-cache-rule.yaml`)
A simple cache rule that caches static assets like images, CSS, and JavaScript files.

**Features:**
- Matches common static file extensions
- Sets edge TTL to 1 hour
- Sets browser TTL to 30 minutes
- Basic caching enabled

**Use case:** Perfect for websites wanting to cache static assets efficiently.

### 2. Advanced Cache Rule (`advanced-cache-rule.yaml`)
A comprehensive example showcasing advanced caching features for API endpoints.

**Features:**
- Status code-specific TTL settings
- Custom cache keys with query parameters, headers, and user attributes
- Cache by device type and geography
- Serve stale configuration
- Cache Reserve integration
- Custom cacheable ports

**Use case:** APIs with complex caching requirements, user-specific content, or geographic variations.

### 3. Bypass Cache Rule (`bypass-cache-rule.yaml`)
Demonstrates how to disable caching for sensitive areas like admin panels or authentication endpoints.

**Features:**
- Completely disables caching
- High priority to override other rules
- Matches admin areas and authentication paths
- Includes cookie-based conditions

**Use case:** Ensuring dynamic content like admin panels or user dashboards are never cached.

### 4. Geographic Cache Rule (`geo-specific-cache.yaml`)
Shows how to cache content with geographic and language variations.

**Features:**
- Geography-based cache keys
- Language and country-specific caching
- Header-based cache differentiation
- Optimized for localized content

**Use case:** Multi-language websites or applications serving region-specific content.

## Quick Start

1. **Update Zone ID**: Replace `your-zone-id-here` in each example with your actual Cloudflare Zone ID.

2. **Configure Provider**: Ensure you have a ProviderConfig named `cloudflare-provider-config`:
   ```yaml
   apiVersion: cloudflare.crossplane.io/v1alpha1
   kind: ProviderConfig
   metadata:
     name: cloudflare-provider-config
   spec:
     credentials:
       source: Secret
       secretRef:
         namespace: crossplane-system
         name: cloudflare-secret
         key: creds
   ```

3. **Apply Examples**: Deploy any example using kubectl:
   ```bash
   kubectl apply -f basic-cache-rule.yaml
   ```

## Cache Rule Expressions

Cache rules use Cloudflare's powerful expression language. Common patterns:

### File Extensions
```
(http.request.uri.path matches ".*\\.(css|js|png|jpg|jpeg|gif|svg|ico|woff|woff2)$")
```

### Path Matching
```
(http.request.uri.path matches "^/api/v1/.*")
```

### Header Conditions
```
(http.request.headers["content-type"][0] contains "application/json")
```

### Cookie Conditions
```
(http.cookie contains "admin_session")
```

### Geographic Conditions
```
(ip.geoip.country eq "US")
```

## Cache Settings Explained

### Edge TTL
Controls how long Cloudflare caches content at edge locations:
- `override_origin`: Override origin cache headers
- `respect_origin`: Respect origin cache headers
- `bypass`: Skip edge caching

### Browser TTL
Controls how long browsers cache content:
- `override_origin`: Set specific browser cache time
- `respect_origin`: Use origin cache headers
- `bypass`: No browser caching

### Cache Keys
Determines what makes a cached response unique:
- **Query Parameters**: Include/exclude specific parameters
- **Headers**: Cache based on specific headers
- **User Attributes**: Cache by device type, geography, language
- **Host**: Include resolved hostname

### Serve Stale
Allows serving cached content even after it expires:
- Improves performance during origin issues
- Reduces origin load
- Better user experience

## Best Practices

1. **Rule Priority**: Lower numbers = higher priority. Use priority carefully to ensure rules apply in the correct order.

2. **Expression Specificity**: Make expressions as specific as possible to avoid unintended matches.

3. **Testing**: Test cache rules in a staging environment before applying to production.

4. **Monitoring**: Monitor cache hit rates and origin load after implementing rules.

5. **Security**: Never cache sensitive content like authentication tokens or personal data.

## Troubleshooting

### Cache Rule Not Working
1. Check rule priority - higher priority rules may override
2. Verify expression syntax
3. Ensure zone ID is correct
4. Check Cloudflare dashboard for rule status

### Performance Issues
1. Review TTL settings
2. Check cache key complexity
3. Monitor cache hit ratios
4. Consider Cache Reserve for frequently accessed content

## Additional Resources

- [Cloudflare Cache Rules Documentation](https://developers.cloudflare.com/cache/how-to/cache-rules/)
- [Expression Language Reference](https://developers.cloudflare.com/ruleset-engine/rules-language/)
- [Cache Key Configuration](https://developers.cloudflare.com/cache/how-to/cache-keys/)
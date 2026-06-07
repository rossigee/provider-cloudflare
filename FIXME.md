# FIXME: Upgrade to crossplane-runtime v2.4.0-rc.0

## Issue

provider-cloudflare is currently on `crossplane-runtime v2.1.0` and needs to be upgraded to `v2.4.0-rc.0` for consistency with other providers.

## Problem

The code uses import path `github.com/crossplane/crossplane-runtime/v2/apis/common/v1` in 98+ locations. This package does NOT exist in v2.4.0-rc.0.

## Root Cause

In crossplane-runtime v2.2.0+, the common/v1 types were moved to a separate module:
- **Old (v2.1.0)**: `github.com/crossplane/crossplane-runtime/v2/apis/common/v1`
- **New (v2.2.0+)**: `github.com/crossplane/crossplane/apis/v2/core/v2`

## Fix Required

1. **Update go.mod**: Change `crossplane-runtime` from `v2.1.0` to `v2.4.0-rc.0`
2. **Add dependency**: Add `github.com/crossplane/crossplane/apis/v2` explicitly
3. **Replace imports**: Change all occurrences of:
   ```go
   import (
       xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
   )
   ```
   To:
   ```go
   import (
       xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
   )
   ```
4. **Run go mod tidy**: To resolve any remaining dependencies
5. **Build and test**: Verify all imports resolve correctly

## Files Affected

All files using the old import path need to be updated. Run this to find them:
```bash
grep -r "crossplane-runtime/v2/apis/common/v1" --include="*.go"
```

## References

- Crossplane runtime v2.2.0 release notes: https://github.com/crossplane/crossplane-runtime/releases/tag/v2.2.0
- Migration guide: Types moved from crossplane-runtime to crossplane/apis

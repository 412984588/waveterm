# PR: Fix repo-wide test baseline failures

## Background

This PR fixes pre-existing test failures in the repository that are **independent** of the security hardening PR (#1).

## Fixed Packages

### 1. pkg/aiusechat

**Tests Fixed:**

- `TestReadDirCallback`
- `TestReadDirSortBeforeTruncate`

**Root Cause:**
Tests used incorrect type assertions - expected `[]map[string]any` but actual return type is `[]fileutil.DirEntryOut`.

**Fix:**

- Added `fileutil` import
- Changed type assertions from `[]map[string]any` to `[]fileutil.DirEntryOut`
- Changed field access from `entry["is_dir"]` to `entry.Dir`

### 2. pkg/remote/connparse

**Tests Fixed:**

- `TestParseURI_WSHRemoteShorthand`
- `TestParseURI_WSHWSL`
- `TestParseUri_LocalWindowsAbsPath`

**Root Cause:**
`ParseURI()` didn't correctly handle URIs starting with `//` prefix, especially when combined with `wsl://` paths.

**Fix:**

- Handle `//` prefix before splitting by `://`
- Don't treat first part as scheme when URI starts with `//`
- Preserve full path for `//wsl://` style URIs

## Commits

1. `c5927beb` - [aiusechat] fix test type assertions for DirEntryOut
2. `79ff14f6` - [connparse] fix URI parsing for // prefix with wsl:// paths

## Verification

```bash
# Individual packages
go test ./pkg/aiusechat/...        # PASS
go test ./pkg/remote/connparse/... # PASS

# Full test suite
go test ./...                      # PASS (all tests)
```

## Relation to Other PRs

This PR is **independent** of PR #1 (security hardening). It can be merged separately.

## Risk Assessment

- **aiusechat fix**: Test-only change, no production code modified
- **connparse fix**: Production code change, but fixes existing bug in URI parsing

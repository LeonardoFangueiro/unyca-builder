# src/internal/manifest/verify.go — Blueprint Manifest Verification
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Validate `MANIFEST.json` integrity and compatibility before any execution.

## Responsibilities
- Parse `MANIFEST.json` from the blueprint directory.
- Check `min_engine`/`max_engine` against builder's version.
- Verify SHA256 for all listed files.

**Not responsible for:**
- Detached signature validation (can be added later).

## Public API / Invocation
Public function: `manifest.Verify(bpDir, engineVersion) error`.

## Data Contracts
**Inputs**
- `bpDir`: path to blueprint version folder.
- `engineVersion`: builder SemVer, e.g., `0.1.0`.

**Outputs**
- None (returns error on failure).

## Control Flow
Linear verification: read → semver compare → hash each file.

## Error Handling & Exit Codes
Clear error messages on semver mismatch or integrity mismatch.

## Logging & Observability
No logs; caller is expected to log errors. Hashing reads only.

## Security Considerations
Manifest may be shipped with signatures; ensure private keys remain outside the repo.

## Edge Cases & Limitations
If `MANIFEST.json` is missing or malformed, verification fails fast.

## Examples
```go
if err := manifest.Verify(bpDir, version.Version); err != nil { /* abort */ }
```

## Change Log (high-level)
- 2025-08-31: initial verification docs.

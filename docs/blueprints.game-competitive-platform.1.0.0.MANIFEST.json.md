# blueprints/game-competitive-platform/1.0.0/MANIFEST.json — Blueprint Manifest

**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** State blueprint version, engine compatibility, and file integrity hashes.

## Responsibilities
- List every relevant file with SHA256.
- Provide `min_engine`/`max_engine` bounds to guard compatibility.

**Not responsible for:**
- Storing private signing keys (never).

## Public API / Invocation
Consumed by: `manifest.Verify(bpDir, version.Version)`

## Data Contracts
**Fields**
- `version`: blueprint SemVer.
- `min_engine` / `max_engine`: compatible builder version range.
- `files`: map of relative file paths → `sha256:<hex>`.
- `signature`: optional signing info.
- `created_at`: ISO timestamp.

## Control Flow
Read JSON → semver compare → verify hashes for each listed file.

## Error Handling & Exit Codes
Errors if any file hash mismatches or engine constraints fail.

## Logging & Observability
No logs; integrity-only file.

## Security Considerations
If using detached signatures, keep secret keys in CI or vault, not in repo.

## Edge Cases & Limitations
Editing any tracked file requires regenerating this manifest.

## Examples
**Generate (Go CLI):**
```bash
./bin/unyca-builder manifest --bp blueprints/game-competitive-platform/1.0.1 --min-engine 0.1.0 --write
# or via Makefile
make manifest TYPE=game-competitive-platform VER=1.0.1
```

## Change Log (high-level)
- 2025-08-31: documented manifest format and purpose (Go CLI generator).

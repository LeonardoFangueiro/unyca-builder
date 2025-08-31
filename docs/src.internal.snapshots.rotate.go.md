# src/internal/snapshots/rotate.go — Snapshot Retention
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Keep the newest N snapshots; delete older ones.

## Responsibilities
- List `builds/<system>/snapshots/*` and sort by modtime desc.
- Delete entries with index ≥ keep.

**Not responsible for:**
- Creating snapshots; naming conventions (handled by caller).

## Public API / Invocation
Public function: `snapshots.Rotate(buildDir string, keep int) error`.

## Data Contracts
**Inputs**: build directory path, retention count.
**Outputs**: old snapshot directories removed.

## Control Flow
Simple list → sort → remove loop.

## Error Handling & Exit Codes
Errors on missing snapshots dir; caller can create it beforehand.

## Logging & Observability
No logging; caller may log after rotation.

## Security Considerations
Avoid retaining sensitive files inside snapshots; keep only metadata/config.

## Edge Cases & Limitations
Assumes snapshots are directories; non-standard files are ignored.

## Examples
```go
_ = snapshots.Rotate("builds/game-cp-01", 100)
```

## Change Log (high-level)
- 2025-08-31: retention documented.

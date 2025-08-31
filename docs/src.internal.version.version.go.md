# src/internal/version/version.go — Builder Version
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Expose builder SemVer for compatibility checks.

## Responsibilities
- Provide `Version` const used by manifest verification.

**Not responsible for:**
- Version negotiation logic.

## Public API / Invocation
Const: `const Version = "0.1.0"`.

## Data Contracts
**Inputs/Outputs**: N/A

## Control Flow
N/A

## Error Handling & Exit Codes
N/A

## Logging & Observability
N/A

## Security Considerations
N/A

## Edge Cases & Limitations
N/A

## Examples
```go
import "unyca-builder/src/internal/version" // version.Version
```

## Change Log (high-level)
- 2025-08-31: initial version doc.

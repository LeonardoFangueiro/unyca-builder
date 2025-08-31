# src/internal/schema/validate.go — Config Schema Validation
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Validate `config.json` against `schemas/config.schema.json`.

## Responsibilities
- Load JSON Schema from disk.
- Validate configuration file and surface errors clearly.

**Not responsible for:**
- Resolving blueprint versions; applying Ansible; building snapshots.

## Public API / Invocation
Public function: `schema.ValidateConfig(path, schemaPath string) error`.

## Data Contracts
**Inputs**: config path and schema path (file URLs).
**Outputs**: returns `nil` on success or an error listing schema violations.

## Control Flow
Direct call to gojsonschema: schema loader + document loader → validate.

## Error Handling & Exit Codes
All validation errors are returned in the error string; caller decides UX.

## Logging & Observability
No logs; deterministic behavior.

## Security Considerations
No secret material involved.

## Edge Cases & Limitations
Large configs are supported; performance depends on schema complexity.

## Examples
```go
must(schema.ValidateConfig("examples/config.json","schemas/config.schema.json"))
```

## Change Log (high-level)
- 2025-08-31: initial schema validation docs.

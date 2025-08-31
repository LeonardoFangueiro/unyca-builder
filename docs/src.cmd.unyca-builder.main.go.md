# src/cmd/unyca-builder/main.go — CLI Entry Point
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Provide CLI commands: validate | plan | build | run | snapshot.

## Responsibilities
- Parse flags and delegate to internal packages.
- Enforce order: validate → plan/build → run.
- Verify manifest before build/run; rotate snapshots after build/snapshot.

**Not responsible for:**
- Implementing Ansible logic, schema internals.

## Public API / Invocation
Commands:
- `validate <config.json> [-schema path]`
- `plan <config.json>`
- `build <config.json> [--upgrade]`
- `run <system_name> [--tags a,b] [--check]`
- `snapshot <system_name> <label>`

## Data Contracts
**Inputs**: CLI args, files under repo (schemas, blueprints, builds).
**Outputs**: files under `builds/<system>` and exit codes (0 success).

## Control Flow
Switch-based command dispatch; per-command functions implement logic.

## Error Handling & Exit Codes
Non-zero exits on validation/manifest/invocation errors.

## Logging & Observability
Standard stdout/stderr; Ansible logs go to file in builds/.

## Security Considerations
Avoid printing secrets; errors do not echo sensitive content.

## Edge Cases & Limitations
Concurrent invocations are not locked in this demo (file-lock may be added).

## Examples
```bash
./bin/unyca-builder build examples/config.json
```

## Change Log (high-level)
- 2025-08-31: initial CLI documentation.

# src/internal/ansible/run.go — Ansible Runner
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Execute blueprint entrypoint with pass-through metadata and logging.

## Responsibilities
- Locate entrypoint: `orchestrator.{yml|yaml|json}`; fallback `servers.*`.
- Set `UNYCA_BUILD_DIR` and use blueprint's `ansible.cfg`.
- Pass `@data.json` and `blueprint_meta` via `-e`.
- Capture stdout/stderr to `builds/<system>/logs/ansible-<ts>.log`.

**Not responsible for:**
- Manifest integrity verification; schema validation; snapshot logic.

## Public API / Invocation
```go
type RunOpts struct {
  BuildDir string; BlueprintDir string; BlueprintMeta map[string]any
  Tags []string; ExtraEnv map[string]string
}
func RunServersYml(o RunOpts) error
```

## Data Contracts
**Inputs**
- `BuildDir`, `BlueprintDir`, `BlueprintMeta`, optional `Tags`, `ExtraEnv`.
- ENV: `UNYCA_BUILD_DIR`, `ANSIBLE_CONFIG`.

**Outputs**
- Log file in `builds/<system>/logs/`.

## Control Flow
1) find entrypoint → 2) build args/env → 3) exec ansible-playbook → 4) write logs.

## Error Handling & Exit Codes
Returns Go `error` if entrypoint missing or Ansible exits non-zero.

## Logging & Observability
Structured (single file) log capture per run.

## Security Considerations
Runner does not handle secrets directly; ensure playbooks avoid leaking them.

## Edge Cases & Limitations
If the blueprint lacks an entrypoint, the function fails early.

## Examples
```go
_ = ansible.RunServersYml(ansible.RunOpts{ BuildDir: "builds/game-cp-01", ... })
```

## Change Log (high-level)
- 2025-08-31: entrypoint detection widened; docs added.

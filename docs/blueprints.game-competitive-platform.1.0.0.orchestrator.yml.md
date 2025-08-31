# blueprints/game-competitive-platform/1.0.0/orchestrator.yml — Blueprint Orchestrator Playbook
**Status:** Stable | **Owner:** Platform | **Since:** 2025-08-31 | **Last Updated:** 2025-08-31  
**Purpose:** Interpret `builds/<system>/data.json`, materialize inline SSH keys (0600), build dynamic inventory, and dispatch roles per group.

## Responsibilities
- Load `data.json` via `UNYCA_BUILD_DIR`.
- Decode `inline_ssh_key_b64` to temporary key files when present.
- Create runtime inventory using `add_host` with host/group/vars.
- Apply roles for groups: `databases`, `apis`, `game-nodes`.

**Not responsible for:**
- Schema validation (`schemas/config.schema.json`), version resolution.
- Snapshot rotation, manifest integrity checks (builder handles).

## Public API / Invocation
Called by the builder:
```bash
ansible-playbook orchestrator.yml -i localhost, -c local   -e @builds/<system>/data.json -e "blueprint_meta=<json>"
```

## Data Contracts
**Inputs**
- `ENV UNYCA_BUILD_DIR`: path to `builds/<system>`.
- `extra_vars`: `@data.json` with `data[].{id,ip,user,groups,ssh_key,inline_ssh_key_b64,connection,vars}`.
- `extra_vars`: `blueprint_meta` (opaque object).

**Outputs**
- Temporary private keys at `/tmp/unyca_<id>.key` if inline keys are used (lifetime = play duration).

## Control Flow
1. Read `data.json`.
2. Write temp keys (if inline).
3. `add_host` per item with connection variables and `vars`.
4. Execute roles for each group.

## Error Handling & Exit Codes
Fails when `UNYCA_BUILD_DIR`/`data.json` is missing; Ansible task failures abort the play.

## Logging & Observability
Standard Ansible output; builder captures to `builds/<system>/logs/ansible-<ts>.log`.

## Security Considerations
Temp keys are `0600`. No secrets persisted by the playbook; avoid printing secrets in tasks.

## Edge Cases & Limitations
When both `ssh_key` and `inline_ssh_key_b64` are present, path-based `ssh_key` takes precedence.

## Examples
```bash
ansible-playbook blueprints/.../orchestrator.yml -i localhost, -c local   -e @builds/game-cp-01/data.json -e "blueprint_meta={'region':'pt-porto','env':'prod'}"
```

## Change Log (high-level)
- 2025-08-31: renamed from servers.yml; clarified responsibilities.

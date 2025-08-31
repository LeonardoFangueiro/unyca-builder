# Unyca Builder — Blueprint-driven, JSON-first system orchestrator

**What it is (TL;DR):** a small Go CLI that drives **Ansible** blueprints using a strict **JSON data contract**.  
Create, upgrade, scale and migrate complex systems by dropping a `config.json` + a blueprint folder.  
Agentless, idempotent, portable across laptops, CI, and jump hosts.

---

## Core principles

- **Blueprint-agnostic:** (builder passes metadata; blueprint owns logic).
- **JSON-first input:** (`config.json` for hosts/groups/vars).
- **Immutability:** blueprints are versioned folders (e.g., `1.0.0`), and `LATEST` is a simple text pointer.
- **Safety nets:** Safety nets (validate/plan, snapshots, manifest integrity).
- **Portability:** (Go + Ansible, agentless).

---

## When to use

- **Provision** new environments (dev → prod) consistently.  
- **Upgrade** systems by switching the blueprint version.  
- **Scale** horizontally by adding hosts to groups in JSON.  
- **Migrate** heavy systems across networks/regions fast (minutes) using the same JSON + blueprint.  
- **Auto‑recovery**: re-run on the same JSON; idempotent roles and snapshots help you recover drift.

---

## Layout

```
unyca-builder/
├── bin/unyca-builder                      # built CLI (optional; use ./unyca runner)
├── docs/                                   # file-level documentation + index + checks
├── examples/config.json                     # sample input (JSON-first)
├── schemas/config.schema.json               # contract for config.json
├── blueprints/
│   └── game-competitive-platform/
│       ├── 1.0.0/
│       │   ├── VERSION
│       │   ├── MANIFEST.json               # integrity + compatibility
│       │   ├── ansible.cfg
│       │   ├── orchestrator.yml            # entrypoint playbook
│       │   └── servers/                    # roles by group
│       │       ├── databases/tasks/main.yml
│       │       ├── apis/tasks/main.yml
│       │       └── game-nodes/tasks/main.yml
│       └── LATEST                          # contains "1.0.0"
└── src/                                     # Go CLI
    ├── cmd/unyca-builder/
    └── internal/{ansible,manifest,schema,snapshots,version}/
```

---

## Data contract (JSON)

Your **single source of truth** (`config.json`):
```json
{
  "system_name": "game-cp-01",
  "system_type": "game-competitive-platform",
  "blueprint_meta": { "region": "pt-porto", "env": "prod" },
  "data": [
    {
      "id": "db1", "ip": "10.0.0.20", "user": "debian",
      "groups": ["databases"], "ssh_key": "/keys/db1",
      "vars": { "pg_version": "16" }
    },
    {
      "id": "api1", "ip": "10.0.0.30", "user": "ubuntu",
      "groups": ["apis"], "inline_ssh_key_b64": "BASE64-KEY...",
      "vars": { "go_max_procs": 4 }
    }
  ]
}
```
Validated by `schemas/config.schema.json`. The builder passes `blueprint_meta` and per-host `vars` to the blueprint **as-is**.

**SSH keys policy:** prefer **path** per host. Inline base64 is supported but only materialized as temp `0600` files during runs.

---

## Quick start

### Option A — Build a binary (recommended)
```bash
go mod tidy
go build -o bin/unyca-builder ./src/cmd/unyca-builder
```

### Option B — Use the runner (no binary)
```bash
./unyca validate examples/config.json
./unyca plan examples/config.json
./unyca build examples/config.json
./unyca run game-cp-01 --tags apis,databases
./unyca snapshot game-cp-01 before-upgrade
```

---

## Workflow

1) **validate** — fail fast on schema issues  
```bash
./unyca validate config.json
```

2) **plan** — resolve blueprint version, summarize hosts/groups, dry-run hint  
```bash
./unyca plan config.json
```

3) **build** — materialize `builds/<system>/` + auto-snapshot + rotation(100)  
```bash
./unyca build config.json          # keeps current blueprint_version.txt if exists
./unyca build --upgrade config.json # writes LATEST (or version from config)
```

4) **run** — execute blueprint entrypoint (`orchestrator.yml`)  
```bash
./unyca run <system_name> --tags apis,databases
./unyca run <system_name> --check             # best-effort check mode (blueprint-dependent)
```

5) **snapshot** — manual snapshot anytime  
```bash
./unyca snapshot <system_name> <label>
```

**Under the hood:**  
- Entrypoint detection: `orchestrator.{yml|yaml|json}` (fallback `servers.*`).  
- Inventory is **dynamic**: built in-memory by the orchestrator via `add_host`.  
- Logs: `builds/<system>/logs/ansible-<timestamp>.log`.  
- Integrity: `MANIFEST.json` verified before build/run; `min_engine` protects compatibility.

---

## Portability

- **OS**: Linux/macOS (Windows via WSL for control plane).  
- **Targets**: Linux (SSH), Windows (WinRM). No agents required.  
- **Runtimes**: laptops, CI systems, jump hosts, containers. Just need Go + Ansible.  
- **No vendor lock‑in**: blueprints are plain Ansible; configs are plain JSON.

---

## Auto‑recovery & migrations

- **Re-run** the same `config.json` to converge a drifted system (Ansible idempotency).  
- **Snapshots**: every build and manual snapshots copy `data.json` + `blueprint_version.txt` (retain 100).  
- **Recovery**: pick a snapshot, restore its files into `builds/<system>/`, and `run`.  
- **Migration**: duplicate `config.json`, adjust IPs/regions, `build` and `run` — same blueprint, new infra.  
- **Rollback**: pin `blueprint_version.txt` to an older version or maintain “before-upgrade” snapshots.

---

## Creating your own blueprints
Required files under `blueprints/<system_type>/<semver>/`: `VERSION`, `MANIFEST.json`, `ansible.cfg`, `orchestrator.yml`, `servers/<group>/tasks/main.yml`.
ssssssssssssssssssssssssssssssssssssss
**Generate MANIFEST (Go):**
```bash
./bin/unyca-builder manifest --bp blueprints/<system_type>/<semver> --min-engine 0.1.0 --write
# or:
make manifest TYPE=<system_type> VER=<semver>
```



Notas:
- Garante que tens o `Makefile` no root e o binário construído (`make verify` ou `go build -o bin/unyca-builder …`) antes dos comandos.
- Mantém `LATEST` alinhado com a versão criada.
::contentReference[oaicite:0]{index=0}



**Folder:** `blueprints/<system_type>/<semver>/`  
**Required files:**
- `VERSION` — `X.Y.Z`
- `LATEST` (at type root) — points to the default version
- `orchestrator.yml` — reads `UNYCA_BUILD_DIR/data.json`, builds inventory, dispatches roles
- `ansible.cfg` — per-blueprint settings
- `servers/<group>/tasks/main.yml` — roles per group
- `MANIFEST.json` — SHA256 of all files + `min_engine` (generated in CI)

**Rules:**
- Never edit old versions; create a new `X.Y.Z` folder and bump `LATEST`.
- Keep roles **idempotent** and tag tasks meaningfully.
- Treat `blueprint_meta` and `data[].vars` as opaque inputs owned by the blueprint.
- No secrets in the repo. Use Ansible Vault or secure channels to deliver secrets (path or inline for last resort).

**Minimal orchestrator (excerpt):**
```yaml
- hosts: localhost
  gather_facts: false
  vars_files:
    - "{{ lookup('env','UNYCA_BUILD_DIR') }}/data.json"
  tasks:
    - name: add hosts
      ansible.builtin.add_host:
        name: "{{ item.id }}"
        ansible_host: "{{ item.ip }}"
        ansible_user: "{{ item.user }}"
        ansible_private_key_file: >-
          {{ (item.ssh_key | default()) | ternary(item.ssh_key, '/tmp/unyca_' ~ item.id ~ '.key') }}
        groups: "{{ item.groups }}"
        vars: "{{ item.vars | default({}) }}"
      loop: "{{ data }}"
- hosts: databases
  roles: [ servers/databases ]
```

---

## Security

- **Secrets delivery:** configs arrive through secure channels; builder does not store secrets beyond runtime.
- **SSH keys:** prefer path per host; inline base64 is materialized as temp `0600` files during runs.
- **Logs:** avoid printing secrets; logs live under `builds/<system>/logs/`.
- **Integrity:** `MANIFEST.json` SHA256 and `min_engine`/`max_engine` checks before execution.
- **Immutability:** versioned blueprints; changes require new `X.Y.Z` directories.

---

## CI/CD integration (suggested gates)

```bash
./unyca validate config.json
./unyca plan config.json
./unyca build config.json
./unyca run <system> --tags smoke     # restricted tag set for quick checks
```

---

## FAQ

**Q: Where is the inventory file?**  
A: There isn’t one. The orchestrator builds a **dynamic inventory** with `add_host` from `data.json`.

**Q: Can I use JSON for playbooks?**  
A: Supported, but YAML is the de-facto standard. The runner accepts `orchestrator.json` if you prefer.

**Q: How many snapshots are kept?**  
A: The newest **100** per system (auto-pruned).

---

## License & contributions

- Contributions welcome via PRs. Keep docs in `docs/` in sync with code changes.
- Do not commit secrets or private keys.

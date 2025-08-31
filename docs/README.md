# Unyca Builder — Documentation

This `docs/` tree provides file-by-file, enterprise-style documentation. Conventions:
- Language: **EN** for documentation; PT in conversation.
- Status flags: Stable/Volatile.
- Each document mirrors its code/path name for easy discovery.

## Structure
- One Markdown per artifact: `docs/<relative_path>.md`.
- Index lives at `docs/INDEX.md`.
- Validation script: `docs/check-docs.sh` (fails if a code/playbook file lacks a doc).



## Manifest generation (Go CLI)

Use the builder itself to generate MANIFESTs (no Python deps):

```bash
./bin/unyca-builder manifest --bp blueprints/<type>/<ver> --min-engine 0.1.0 --write
# or via Makefile
make manifest TYPE=<type> VER=<ver>
```


## How to contribute
1. Update code and its corresponding `docs/...` file in the same PR.
2. Keep **Purpose**, **Responsibilities**, **Data Contracts** and **Security** sections updated.
3. Add an entry to **Change Log** explaining *why* the change was made.

# Changelog
All notable changes to this project will be documented in this file.

Format inspired by https://keepachangelog.com and follows https://semver.org.

## [Unreleased]

### Added
- Go CLI subcommand: `manifest` (generate `MANIFEST.json`).
- Makefile targets: `manifest`, `bump-blueprint`, `set-latest`, `plan`, `build`, `run-check`, `release`.
- PR template at `.github/pull_request_template.md`.

### Changed
- `orchestrator.yml`: `add_host` omits `ansible_private_key_file` when no key (path > inline > omit).
- `ansible.cfg`: builtin default callback + YAML result format; deprecation warnings off.
- Docs updated to reference **Go** manifest generator.

### Fixed
- CLI `run --check` passes `--check` to `ansible-playbook` and supports console tee via `UNYCA_TEE=1`.

### Security
- SSH key policy clarified (path preferred; inline base64 only as temp 0600 at runtime).

### Blueprints
- **game-competitive-platform**
  - **1.0.1** — Orchestrator key handling fix; deprecated callback removed; manifest via Go CLI.
  - **1.0.0** — Initial version.

---

## [0.1.0] - 2025-08-31
### Added
- Initial CLI (`validate|plan|build|run|snapshot`), JSON Schema, Ansible integration, snapshots, manifest verification.

[Unreleased]: https://github.com/LeonardoFangueiro/unyca-builder/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/LeonardoFangueiro/unyca-builder/releases/tag/v0.1.0

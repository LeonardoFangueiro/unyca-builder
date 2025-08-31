SHELL := /bin/bash

# Blueprint params (override on CLI)
TYPE ?= game-competitive-platform
VER  ?= 1.0.1
CUR  ?= 1.0.0
NEW  ?= 1.0.1

BP_DIR      := blueprints/$(TYPE)/$(VER)
BP_CUR_DIR  := blueprints/$(TYPE)/$(CUR)
BP_NEW_DIR  := blueprints/$(TYPE)/$(NEW)
LATEST_FILE := blueprints/$(TYPE)/LATEST

.PHONY: help manifest bump-blueprint set-latest plan build run-check verify release changelog-check

help:
	@echo "Targets:"
	@echo "  manifest TYPE=<t> VER=<v>               - Generate MANIFEST.json for blueprint <t>/<v> (Go CLI)"
	@echo "  bump-blueprint TYPE=<t> CUR=<v> NEW=<v2> - Copy CUR->NEW, write VERSION, MANIFEST, update LATEST"
	@echo "  set-latest TYPE=<t> VER=<v>             - Update LATEST pointer"
	@echo "  plan                                     - Build CLI and run 'plan' on examples/config.json"
	@echo "  build                                    - Build CLI and run 'build' on examples/config.json"
	@echo "  run-check                                - Run ansible in check mode (UNYCA_TEE=1)"
	@echo "  release VER=X.Y.Z                        - Cut a release: stamp CHANGELOG, create git tag"
	@echo "  changelog-check                          - Verify CHANGELOG has [Unreleased]"

manifest:
	@test -d "$(BP_DIR)" || { echo "Missing $(BP_DIR)"; exit 2; }
	go build -o bin/unyca-builder ./src/cmd/unyca-builder
	./bin/unyca-builder manifest --bp "$(BP_DIR)" --min-engine 0.1.0 --write
	@echo "Wrote $(BP_DIR)/MANIFEST.json"

bump-blueprint:
	@test -d "$(BP_CUR_DIR)" || { echo "Missing $(BP_CUR_DIR)"; exit 2; }
	@test ! -e "$(BP_NEW_DIR)" || { echo "Already exists: $(BP_NEW_DIR)"; exit 2; }
	cp -a "$(BP_CUR_DIR)" "$(BP_NEW_DIR)"
	echo "$(NEW)" > "$(BP_NEW_DIR)/VERSION"
	$(MAKE) manifest TYPE="$(TYPE)" VER="$(NEW)"
	$(MAKE) set-latest TYPE="$(TYPE)" VER="$(NEW)"
	@echo "Bumped blueprint $(TYPE): $(CUR) -> $(NEW)"

set-latest:
	@echo "$(VER)" > "$(LATEST_FILE)"
	@echo "Set LATEST ($(TYPE)) -> $(VER)"

# --- Builder helpers ---
verify:
	go mod tidy
	go build -o bin/unyca-builder ./src/cmd/unyca-builder

plan: verify
	./bin/unyca-builder plan examples/config.json

build: verify
	./bin/unyca-builder build examples/config.json

run-check: verify
	UNYCA_TEE=1 ./bin/unyca-builder run game-cp-01 --check

# --- Release helpers ---
# Usage: make release VER=0.1.1
release:
	@test -n "$(VER)" || { echo "VER required (e.g., make release VER=0.1.1)"; exit 2; }
	@cp CHANGELOG.md CHANGELOG.md.bak
	@DATE=$$(date +%Y-%m-%d); \
	awk -v ver="$(VER)" -v date="$$DATE" '\
	  $$0=="## [Unreleased]" { print; print ""; print "## [" ver "] - " date; next } \
	  { print }' CHANGELOG.md.bak > CHANGELOG.md
	@rm -f CHANGELOG.md.bak
	git add CHANGELOG.md
	git commit -m "chore(release): $(VER)"
	git tag v$(VER)
	@echo "Release v$(VER) staged. Push with: git push && git push --tags"

changelog-check:
	@grep -q "## \\[Unreleased\\]" CHANGELOG.md || { echo "Missing [Unreleased] section"; exit 2; }

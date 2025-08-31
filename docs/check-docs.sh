#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
miss=0
while IFS= read -r f; do
  rel="$f"
  doc="docs/${rel//\//.}.md"
  doc="${doc/.yml/.md}"; doc="${doc/.yaml/.md}"; doc="${doc/.json/.md}"; doc="${doc/.go/.md}"
  if [[ ! -f "$doc" ]]; then
    echo "MISSING DOC: $rel -> expected: $doc"
    miss=1
  fi
done < <(git ls-files | grep -E '^(src/.*\.go|blueprints/.*\.(yml|yaml|json)|schemas/.*\.json|examples/.*\.json|README\.md)$')
exit $miss

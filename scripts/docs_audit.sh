#!/usr/bin/env bash
set -euo pipefail

fail=0
bad() { echo "FAIL: $*" >&2; fail=1; }

echo "== docs_audit: required files =="
req=(
  "docs/README.md"
  "docs/core/ARCHITECTURE.md"
  "docs/core/STRUCTURE.md"
  "docs/core/THREAT_MODEL.md"
  "docs/core/ERROR_HANDLING.md"
  "docs/core/TESTING.md"
  "docs/adr/README.md"
  "docs/adr/TEMPLATE.md"
  "docs/internal/ai/README.md"
  "docs/internal/ai/AI_RULES.md"
  "docs/internal/ai/AI_PROMPT.md"
)
for f in "${req[@]}"; do
  [[ -f "$f" ]] || bad "missing $f"
done

echo
echo "== docs_audit: stale path refs =="
if rg -n --no-heading 'docs/AI_RULES\.md|docs/ai-rules/|internal/ai/ai-rules/' docs >/dev/null; then
  bad "found stale path references (run: rg -n 'docs/AI_RULES.md|docs/ai-rules/|internal/ai/ai-rules/' docs)"
fi

echo
echo "== docs_audit: ADR filename + date sanity =="
adr_files=()
while IFS= read -r f; do adr_files+=("$f"); done < <(find docs/adr -maxdepth 1 -type f -name "*.md" ! -name "README.md" ! -name "TEMPLATE.md" | sort)

# filename pattern 0001-*.md
for f in "${adr_files[@]}"; do
  base="$(basename "$f")"
  [[ "$base" =~ ^[0-9]{4}-.+\.md$ ]] || bad "ADR filename not normalized: $f (expected 0001-title.md)"
done

# date format: Tanggal: YYYY-MM-DD (optional time)
# and no placeholder "xx"
for f in "${adr_files[@]}"; do
  if ! rg -n --no-heading '^Tanggal:\s+[0-9]{4}-[0-9]{2}-[0-9]{2}(\s+[0-9]{2}:[0-9]{2}\s+WITA)?\s*$' "$f" >/dev/null; then
    bad "ADR date format not normalized in $f (expected: 'Tanggal: 2025-12-29' or 'Tanggal: 2025-12-29 09:42 WITA')"
  fi
  if rg -n --no-heading 'Tanggal:.*xx' "$f" >/dev/null; then
    bad "ADR has placeholder time in $f"
  fi
done

echo
echo "== docs_audit: markdown link existence (relative links) =="

python3 - <<'PY' || fail=1
import re, pathlib, sys

root = pathlib.Path("docs")
bad = []

link_re = re.compile(r'\[[^\]]+\]\(((?:\./|\../)[^)\s]+)\)')

for md in root.rglob("*.md"):
    text = md.read_text(encoding="utf-8", errors="ignore")
    for m in link_re.finditer(text):
        target = m.group(1)
        target = target.split("#", 1)[0]  # strip anchors
        if not target:
            continue
        p = (md.parent / target).resolve()
        if not p.exists():
            bad.append((str(md), target))

if bad:
    print("BROKEN LINKS:")
    for src, tgt in bad[:200]:
        print(f"- {src} -> {tgt}")
    sys.exit(2)

print("OK: no broken relative links")
PY

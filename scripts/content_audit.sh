#!/usr/bin/env bash
set -euo pipefail

fail=0
bad() { echo "FAIL: $*" >&2; fail=1; }

echo "== content_audit: head preview (first 5 non-empty lines) + meaningless check =="

# audit tracked Go files (exclude vendor)
mapfile -t go_files < <(git ls-files '*.go' | rg -v '^vendor/')

pkg_only=()

is_pkg_only_go() {
  local f="$1"
  python3 - "$f" <<'PY'
import sys, re, pathlib

p = pathlib.Path(sys.argv[1])
name = p.name

# doc.go is allowed to be "comment + package" only (package documentation file)
if name == "doc.go":
    sys.exit(0)

text = p.read_text(encoding="utf-8", errors="ignore")

# Remove block comments
text = re.sub(r'/\*.*?\*/', '', text, flags=re.S)

lines = []
for raw in text.splitlines():
    # Remove line comments
    raw = re.sub(r'//.*$', '', raw).strip()
    if raw:
        lines.append(raw)

# Meaningless: only "package xxx" remains
if len(lines) == 1 and re.fullmatch(r'package\s+[A-Za-z_]\w*', lines[0]):
    sys.exit(2)

sys.exit(0)
PY
}

for f in "${go_files[@]}"; do
  if is_pkg_only_go "$f"; then
    :
  else
    rc=$?
    if [[ "$rc" -eq 2 ]]; then
      pkg_only+=("$f")
    else
      bad "error parsing $f"
    fi
  fi
done

if [[ "${#pkg_only[@]}" -gt 0 ]]; then
  bad "meaningless Go files detected: package-only (.go with only 'package xxx')"
  echo
  echo "== package-only Go files =="
  for f in "${pkg_only[@]}"; do
    echo "----- $f -----"
    awk 'NF{print; n++; if(n==5) exit}' "$f"
    echo
  done
fi

if [[ "$fail" -ne 0 ]]; then
  echo "== RESULT: FAIL ==" >&2
  exit 1
fi

echo
echo "== RESULT: OK (no meaningless files detected by heuristics) =="

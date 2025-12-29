#!/usr/bin/env bash
set -euo pipefail

N="${N:-5}"
FAIL=0

say() { printf "%s\n" "$*"; }
bad() { say "FAIL: $*"; FAIL=1; }

# list tracked files only (biar gak kebawa sampah local)
mapfile -t FILES < <(git ls-files)

is_text() {
  # reject obvious binaries: contain NUL byte
  ! LC_ALL=C grep -Iq . "$1" 2>/dev/null
  return $?
}

head_nonempty() {
  # print first N non-empty lines (trim right)
  python3 - "$1" "$N" <<'PY'
import sys, pathlib
p = pathlib.Path(sys.argv[1])
n = int(sys.argv[2])
try:
    data = p.read_text(encoding="utf-8", errors="ignore").splitlines()
except Exception:
    sys.exit(0)
out=[]
for line in data:
    s=line.rstrip()
    if s.strip()=="":
        continue
    out.append(s)
    if len(out)>=n:
        break
print("\n".join(out))
PY
}

meaningless_go() {
  local f="$1"
  local stripped
  # 1. Strip komentar //
  # 2. Strip komentar block /* */ (menggunakan perl agar support multi-line)
  # 3. Strip baris kosong atau hanya whitespace
  stripped=$(perl -0777 -pe 's/\/\*.*?\*\///gs' "$f" | sed -E 's|//.*$||' | sed -E '/^[[:space:]]*$/d')
  
  local line_count
  line_count=$(printf "%s\n" "$stripped" | wc -l | tr -d ' ')

  # FAIL jika hasil strip hanya 1 baris dan isinya adalah 'package <name>'
  if [[ "$line_count" -eq 1 ]] && printf "%s\n" "$stripped" | grep -Eq '^[[:space:]]*package[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*[[:space:]]*$'; then
    return 0 # True, ini meaningless
  fi
  return 1 # False, ini berisi konten valid
}

meaningless_md() {
  python3 - "$1" <<'PY'
import re, sys, pathlib
p=pathlib.Path(sys.argv[1])
t=p.read_text(encoding="utf-8", errors="ignore")

# remove html comments
t=re.sub(r'<!--.*?-->', '', t, flags=re.S)
lines=[ln.strip() for ln in t.splitlines() if ln.strip()]

# meaningless if: empty OR only 1-2 headings and nothing else substantive
if not lines:
    sys.exit(0)

# tolerate title + short desc, but flag super minimal
only_headings = all(ln.startswith("#") for ln in lines)
if only_headings and len(lines) <= 2:
    sys.exit(0)

# placeholders
blob=" ".join(lines).lower()
bad_words=["tbd","todo","placeholder","<judul","isi di sini","tulis di sini","xx:xx"]
if any(w in blob for w in bad_words) and len(lines) <= 6:
    sys.exit(0)

sys.exit(1)
PY
}

say "== content_audit: head preview (first ${N} non-empty lines) + meaningless check =="
for f in "${FILES[@]}"; do
  # skip vendor-ish or generated-ish noise
  case "$f" in
    vendor/*|.git/*|go.sum) continue ;;
  esac

  [[ -f "$f" ]] || continue
  is_text "$f" || continue

  ext="${f##*.}"
  preview="$(head_nonempty "$f" || true)"

  # Always print preview for docs/internal areas; elsewhere only if flagged.
  always=0
  case "$f" in
    docs/*|internal/*|make/*|scripts/*|cmd/*|migrations/*|Makefile) always=1 ;;
  esac

  flagged=0
  if [[ "$ext" == "go" ]]; then
    if meaningless_go "$f"; then flagged=1; fi
  elif [[ "$ext" == "md" ]]; then
    if meaningless_md "$f"; then flagged=1; fi
  fi

  if [[ "$always" -eq 1 || "$flagged" -eq 1 ]]; then
    say ""
    say "--- FILE: $f ---"
    if [[ -n "$preview" ]]; then
      say "$preview"
    else
      say "(no non-empty lines)"
    fi
  fi

  if [[ "$flagged" -eq 1 ]]; then
    bad "$f looks meaningless (empty/placeholder-level)."
  fi
done

say ""
if [[ "$FAIL" -eq 1 ]]; then
  say "== RESULT: FAIL (some files look meaningless) =="
  exit 1
fi
say "== RESULT: OK (no meaningless files detected by heuristics) =="

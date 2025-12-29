#!/usr/bin/env bash
set -euo pipefail

fail=0
bad() { echo "FAIL: $*" >&2; fail=1; }

echo "== testtags_audit: component tag/name consistency =="

# component tag but filename not *_component_test.go
while IFS= read -r f; do
  case "$f" in
    *_component_test.go) ;;
    *) bad "component build tag but wrong filename: $f" ;;
  esac
done < <(rg -n '^//go:build component' -l internal | rg '_test\.go$' | sort)

# *_component_test.go but missing tag
while IFS= read -r f; do
  rg -q '^//go:build component' "$f" || bad "missing //go:build component in $f"
done < <(find internal -name '*_component_test.go' | sort)

echo
echo "== testtags_audit: integration tag/name consistency =="

# *_integration_test.go but missing tag
while IFS= read -r f; do
  rg -q '^//go:build integration' "$f" || bad "missing //go:build integration in $f"
done < <(find internal -name '*_integration_test.go' | sort)

[ "$fail" -eq 0 ] || exit 1
echo "OK: testtags_audit passed"

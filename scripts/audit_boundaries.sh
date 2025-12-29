#!/usr/bin/env bash
set -euo pipefail

MOD="example.com/your-api"

# include normal imports + test imports (biar ga bisa bypass via _test.go)
tmpl='{{.ImportPath}}|{{range $i,$e := .Imports}}{{if $i}},{{end}}{{$e}}{{end}}|{{range $i,$e := .TestImports}}{{if $i}},{{end}}{{$e}}{{end}}|{{range $i,$e := .XTestImports}}{{if $i}},{{end}}{{$e}}{{end}}'
fail=0

bad() {
  echo "FAIL: $1"
  fail=1
}

# matchers (sesuai struktur repo lu)
is_domain_pkg()        { [[ "$1" == "$MOD/internal/modules/"*"/domain"* ]]; }
is_ports_pkg()         { [[ "$1" == "$MOD/internal/modules/"*"/ports"*  ]]; }
is_usecase_pkg()       { [[ "$1" == "$MOD/internal/modules/"*"/usecase"* ]]; }
is_module_http_pkg()   { [[ "$1" == "$MOD/internal/modules/"*"/transport/http"* ]]; }
is_core_http_pkg()     { [[ "$1" == "$MOD/internal/transport/http"* ]]; }
is_platform_pkg()      { [[ "$1" == "$MOD/internal/platform/"* ]]; }

check_imports() {
  local pkg="$1" imps="$2"
  IFS=',' read -r -a arr <<< "${imps:-}"
  for imp in "${arr[@]}"; do
    [[ -z "${imp// }" ]] && continue

    # DOMAIN: stdlib only (no internal module import, no third-party module path)
    if is_domain_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/"* ]] || [[ "$imp" == *.* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi

    # PORTS: allow stdlib + domain + google/uuid
    if is_ports_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/internal/modules/"*"/domain"* ]]; then
        continue
      fi
      if [[ "$imp" == "github.com/google/uuid" ]]; then
        continue
      fi
      if [[ "$imp" == "$MOD/"* ]] || [[ "$imp" == *.* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi

    # USECASE: forbid platform/http/app/config
    if is_usecase_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/internal/platform/"* ]] ||
         [[ "$imp" == "$MOD/internal/transport/http/"* ]] ||
         [[ "$imp" == "$MOD/internal/app/"* ]] ||
         [[ "$imp" == "$MOD/internal/config"* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi

    # MODULE HTTP (internal/modules/*/transport/http): forbid platform
    if is_module_http_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/internal/platform/"* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi

    # CORE HTTP (internal/transport/http): forbid platform
    if is_core_http_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/internal/platform/"* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi

    # PLATFORM: forbid modules usecase/module http + internal/transport/http/app
    if is_platform_pkg "$pkg"; then
      if [[ "$imp" == "$MOD/internal/modules/"*"/usecase"* ]] ||
         [[ "$imp" == "$MOD/internal/modules/"*"/transport/http"* ]] ||
         [[ "$imp" == "$MOD/internal/transport/http/"* ]] ||
         [[ "$imp" == "$MOD/internal/app/"* ]]; then
        bad "$pkg imports forbidden: $imp"
      fi
    fi
  done
}

while IFS='|' read -r pkg imps timps ximps; do
  check_imports "$pkg" "${imps:-},${timps:-},${ximps:-}"
done < <(go list -f "$tmpl" ./...)

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "OK: boundary passed"

.PHONY: audit-lines audit-packages audit-makefiles audit

audit-lines:
	out="$$(find . -name '*.go' -not -path './vendor/*' -print0 \
	| xargs -0 wc -l \
	| awk '$$2!="total" && $$1>100 {printf "%4d  %s\n",$$1,$$2}' \
	| sort -nr)"; \
	if [ -n "$$out" ]; then \
		echo "FAIL: .go file >100 lines:"; \
		echo "$$out"; \
		exit 1; \
	fi

# Rule: 1 package per directory (non-recursive), ignore *_test.go
audit-packages:
	bad=0
	echo "Checking package consistency (per-dir, non-test)..."
	while IFS= read -r d; do
		files="$$(find "$$d" -maxdepth 1 -type f -name '*.go' ! -name '*_test.go' 2>/dev/null || true)"
		[ -z "$$files" ] && continue
		pkgs="$$(awk 'NR==1 && $$1=="package"{print $$2}' $$files | sort -u | tr '\n' ' ')"
		n="$$(echo "$$pkgs" | wc -w | tr -d ' ')"
		if [ "$$n" -gt 1 ]; then
			echo "$$d -> $$pkgs"
			bad=1
		fi
	done < <(find internal -type d | sort)
	[ "$$bad" -eq 0 ] || (echo "FAIL: multiple packages in one directory"; exit 1)

# Optional tapi kamu minta: Makefile / make/*.mk juga <= 100 baris
audit-makefiles:
	out="$$(for f in Makefile make/*.mk; do \
		[ -f "$$f" ] || continue; \
		n=$$(wc -l < "$$f" | tr -d ' '); \
		if [ "$$n" -gt 100 ]; then printf "%4d  %s\n" "$$n" "$$f"; fi; \
	done)"; \
	if [ -n "$$out" ]; then \
		echo "FAIL: Makefile/*.mk >100 lines:"; \
		echo "$$out"; \
		exit 1; \
	fi

audit: prereq check audit-lines audit-packages audit-makefiles
	@echo "OK: audit passed"

.PHONY: _env migrate-status migrate-up api api-log fmt fmt-check \
        test test-unit test-component test-integration vet check

_env:
	test -f "$(ENV_FILE)" || (echo "missing $(ENV_FILE) (copy from .env.example)"; exit 1)

migrate-status: _env
	set -a
	source $(ENV_FILE)
	set +a
	$(GO) run ./cmd/migrate status

migrate-up: _env
	set -a
	source $(ENV_FILE)
	set +a
	$(GO) run ./cmd/migrate up

api: _env
	set -a
	source $(ENV_FILE)
	set +a
	$(GO) run ./cmd/api

api-log: _env
	set -a
	source $(ENV_FILE)
	set +a
	$(GO) run ./cmd/api 2>&1 | tee /tmp/api.log

# ===== Formatting =====

fmt:
	$(GOFMT) -w .

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "FAIL: gofmt needed on:"; \
		echo "$$out"; \
		exit 1; \
	fi

# ===== Test Suites (audit-friendly) =====

test-unit:
	$(GO) test ./... -count=1

test-component:
	$(GO) test -tags=component \
		./internal/transport/http/... \
		./internal/modules/auth/transport/http/... \
		-count=1

test-integration:
	$(GO) test -tags=integration ./... -count=1

test: test-unit test-component
	@echo "OK: unit + component passed"

vet:
	$(GO) vet ./...

# CI-like (no auto-fix)
check: fmt-check test vet

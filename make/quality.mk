.PHONY: fmt fmt-check test-unit test-component test-integration test vet check

# Formatting
fmt:
	$(GOFMT) -w .

fmt-check:
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "FAIL: gofmt needed on:"; \
		echo "$$out"; \
		exit 1; \
	fi

# Tests
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

# Vet + CI-like
vet:
	$(GO) vet ./...

check: fmt-check test vet

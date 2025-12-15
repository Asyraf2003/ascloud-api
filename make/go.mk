.PHONY: _env migrate-status migrate-up api api-log fmt test vet check

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

fmt:
	$(GOFMT) -w .

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: fmt test vet

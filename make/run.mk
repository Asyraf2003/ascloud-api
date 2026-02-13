ENV_FILE ?= .env

.PHONY: run-api run-api-legacy env-check env-print

env-check:
	@test -f "$(ENV_FILE)" || (echo "missing $(ENV_FILE) (copy .env.example -> .env)"; exit 1)

run-api: env-check
	@bash -lc 'set -a; source "$(ENV_FILE)"; set +a; exec go run ./cmd/api'

run-api-legacy: env-check
	@bash -lc 'set -a; source "$(ENV_FILE)"; set +a; exec go run -tags=legacy_postgres ./cmd/api'

env-print: env-check
	@bash -lc 'set -a; source "$(ENV_FILE)"; set +a; env | grep -E "^(APP_ENV|SERVICE_NAME|HTTP_|AUTH_|AWS_|DYNAMODB_|DDB_)" | sort'

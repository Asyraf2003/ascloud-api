.PHONY: ddb-local

ddb-local: env-check
	@bash -lc 'set -a; source "$(ENV_FILE)"; set +a; ./scripts/ddb_local_create_tables.sh'

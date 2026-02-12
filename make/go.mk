.PHONY: legacy-migrate-status-postgres legacy-migrate-up-postgres migrate-status migrate-up

legacy-migrate-status-postgres: _env
	$(LOAD_ENV)
	$(GO) run ./cmd/migrate status

legacy-migrate-up-postgres: _env
	$(LOAD_ENV)
	$(GO) run ./cmd/migrate up

migrate-status:
	@echo "DEPRECATED: gunakan legacy-migrate-status-postgres"
	@$(MAKE) legacy-migrate-status-postgres

migrate-up:
	@echo "DEPRECATED: gunakan legacy-migrate-up-postgres"
	@$(MAKE) legacy-migrate-up-postgres

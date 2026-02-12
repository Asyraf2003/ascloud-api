.PHONY: \
  dev-up dev-down dev-ps dev-logs \
  legacy-dev-up-postgres legacy-dev-down-postgres legacy-dev-ps-postgres legacy-dev-logs-postgres

# ----- Legacy targets (Postgres/dev via docker compose) -----

legacy-dev-up-postgres: prereq-docker
	$(COMPOSE) -f $(COMPOSE_FILE) up -d
	$(COMPOSE) -f $(COMPOSE_FILE) ps

legacy-dev-down-postgres: prereq-docker
	$(COMPOSE) -f $(COMPOSE_FILE) down

legacy-dev-ps-postgres: prereq-docker
	$(COMPOSE) -f $(COMPOSE_FILE) ps

legacy-dev-logs-postgres: prereq-docker
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f --tail=200

# ----- Deprecated wrappers (keep old names, but delegate) -----

dev-up:
	@echo "DEPRECATED: gunakan legacy-dev-up-postgres"
	@$(MAKE) legacy-dev-up-postgres

dev-down:
	@echo "DEPRECATED: gunakan legacy-dev-down-postgres"
	@$(MAKE) legacy-dev-down-postgres

dev-ps:
	@echo "DEPRECATED: gunakan legacy-dev-ps-postgres"
	@$(MAKE) legacy-dev-ps-postgres

dev-logs:
	@echo "DEPRECATED: gunakan legacy-dev-logs-postgres"
	@$(MAKE) legacy-dev-logs-postgres

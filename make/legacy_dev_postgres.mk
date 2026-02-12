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

# ----- UX Wrappers (Official aliases to legacy) -----

dev-up: legacy-dev-up-postgres
dev-down: legacy-dev-down-postgres
dev-ps: legacy-dev-ps-postgres
dev-logs: legacy-dev-logs-postgres
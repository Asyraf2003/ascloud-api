.PHONY: db-psql legacy-db-psql-postgres

legacy-db-psql-postgres: prereq-docker
	$(COMPOSE) -f $(COMPOSE_FILE) exec -it postgres psql -U postgres -d app

db-psql:
	@echo "DEPRECATED: gunakan legacy-db-psql-postgres"
	@$(MAKE) legacy-db-psql-postgres

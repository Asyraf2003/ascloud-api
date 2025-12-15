.PHONY: db-psql
db-psql:
	$(COMPOSE) -f $(COMPOSE_FILE) exec -it postgres psql -U postgres -d app

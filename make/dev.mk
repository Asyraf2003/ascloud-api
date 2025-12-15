.PHONY: dev-up dev-down dev-ps dev-logs

dev-up:
	$(COMPOSE) -f $(COMPOSE_FILE) up -d
	$(COMPOSE) -f $(COMPOSE_FILE) ps

dev-down:
	$(COMPOSE) -f $(COMPOSE_FILE) down

dev-ps:
	$(COMPOSE) -f $(COMPOSE_FILE) ps

dev-logs:
	$(COMPOSE) -f $(COMPOSE_FILE) logs -f --tail=200

.PHONY: postgres-up
postgres-up: ## Start postgres
	@docker-compose up -d postgres

.PHONY: postgres-down
postgres-down: ## Shutdown postgres
	@docker stop simplebank_postgres

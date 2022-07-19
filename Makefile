.PHONY: postgres-up
postgres-up: ## Start postgres
	@docker-compose up -d postgres

.PHONY: postgres-down
postgres-down: ## Shutdown postgres
	@docker stop simplebank_postgres

.PHONY: createdb
createdb: ## create database
	@docker exec -it postgres createdb --username=root --owner=root simple_bank

.PHONY: dropdb
dropdb: ## drop database
	@docker exec -it postgres dropdb simple_bank

.PHONY: migrate-up
migrate-up: ## migrate up database
	@migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" \
	-path=db/migration -verbose up

.PHONY: migrate-down
migrate-down: ## migrate down database
	@migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" \
	-path=db/migration -verbose down

.PHONY: migrate-up1
migrate-up1: ## migrate up database
	@migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" \
	-path=db/migration -verbose up 1

.PHONY: migrate-down1
migrate-down1: ## migrate down database
	@migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" \
	-path=db/migration -verbose down 1

.PHONY: sqlc
sqlc: ## generate sqlc
	@sqlc generate

.PHONY: test
test: ## run test
	@go test -v -cover ./...

.PHONY: server
server: ## run server
	@go run -race main.go

.PHONY: fake
fake: ## generate fake
	@counterfeiter -o ./db/fake ./db/sqlc Store

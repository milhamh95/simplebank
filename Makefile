DB_URL=postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable

.PHONY: api-up
api-up: ## Start api
	@docker-compose up

.PHONY: api-stop
api-stop: ## Stop api
	@docker stop simplebank_api simplebank_postgres

.PHONY: api-down
api-down: ## Down / delete api
	@docker-compose down simplebank_api simplebank_postgres
	@docker rmi simplebank_api:latest

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
	@migrate -path db/migration -database "$(DB_URL)" \
	-path=db/migration -verbose up

.PHONY: migrate-down
migrate-down: ## migrate down database
	@migrate -path db/migration -database "$(DB_URL)" \
	-path=db/migration -verbose down

.PHONY: migrate-up1
migrate-up1: ## migrate up database
	@migrate -path db/migration -database "$(DB_URL)" \
	-path=db/migration -verbose up 1

.PHONY: migrate-down1
migrate-down1: ## migrate down database
	@migrate -path db/migration -database "$(DB_URL)" \
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

.PHONY: docker-image
docker-image: ## Build simplebank docker image
	@docker build -t simplebank:latest .

.PHONY: fake
fake: ## generate fake
	@counterfeiter -o ./db/fake ./db/sqlc Store

.PHONY: db_docs
db_docs: ## generate db docs
	@dbdocs build doc/db.dbml

.PHONY: db_schema
db_schema: ## generate db schema
	@dbml2sql --postgres -o doc/schema.sql doc/db.dbml

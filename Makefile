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

.PHONY: redis-up
redis-up: ## Start redis
	@docker-compose up -d redis

.PHONY: redis-down
redis-down: ## Shutdown redis
	@docker stop simplebank_redis


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

.PHONY: new-migration
new-migration: ## create new migration
	@migrate create -ext sql -dir db/migration -seq $(name)

.PHONY: sqlc
sqlc: ## generate sqlc
	@sqlc generate

.PHONY: test
test: ## run test
	@go test -v -cover -short ./...

.PHONY: server
server: ## run server
	@go run -race main.go

.PHONY: docker-image
docker-image: ## Build simplebank docker image
	@docker build -t simplebank:latest .

.PHONY: fake
fake: ## generate fake
	@counterfeiter -o ./db/fake ./db/sqlc Store
	@counterfeiter -o ./worker/fake  ./worker TaskDistributor

.PHONY: db_docs
db_docs: ## generate db docs
	@dbdocs build doc/db.dbml

.PHONY: db_schema
db_schema: ## generate db schema
	@dbml2sql --postgres -o doc/schema.sql doc/db.dbml

.PHONY: proto
proto: ## generate proto file
	@rm -f pb/*.go
	@rm -f doc/swagger/*.swagger.json
	@protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=simple_bank \
    proto/*.proto
	@statik -src=./doc/swagger -dest=./doc

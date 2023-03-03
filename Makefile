DB_URL=postgresql://root:secret@localhost:5432/shakespeare_dev?sslmode=disable

.PHONY: network
network:
	docker network create shakespeare-network

.PHONY: postgres
postgres:
	docker run --name postgres-shakespeare --network shakespeare-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

.PHONY: createdb
createdb:
	docker exec -it postgres-shakespeare createdb --username=root --owner=root shakespeare_dev

.PHONY: dropdb
dropdb:
	docker exec -it postgres-shakespeare dropdb shakespeare_dev

.PHONY: migrateup
migrateup:
	migrate -path postgres/migration -database "$(DB_URL)" -verbose up

.PHONY: migratedown
migratedown:
	migrate -path postgres/migration -database "$(DB_URL)" -verbose down

.PHONY: db_schema
db_schema:
	dbml2sql --postgres -o postgres/schema.sql postgres/db.dbml

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: server
server:
	go run cmd/shakespeare/main.go

.PHONY: mock
mock:
	mockgen -package mockdb -destination mock/store.go github.com/earlofurl/scenes-of-shakespeare/db/sqlc Store

.PHONY: redis
redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

.PHONY: seed
seed:
	go run postgres/seeder/seeder.go

.PHONY: start migrate build test test-storage mock cover

start:
	go run ./cmd/server

migrate:
	go run ./cmd/migrate

build:
	go build -o bin/server ./cmd/server
	go build -o bin/migrate ./cmd/migrate

# -race stays on: writes are serialized through a database lock, and a race only CI sees is a race nobody fixes.
test:
	go test ./... -race

# Same suite, but the storage tests need TEST_POSTGRES_DSN pointing at a reachable PostgreSQL whose database
# name ends in _test (they empty every table); without it they skip. A throwaway database:
#   docker run -d --name trading-board-test-pg -e POSTGRES_PASSWORD=testpass \
#     -e POSTGRES_DB=trading_board_test -p 55432:5432 postgres:17-alpine
#   export TEST_POSTGRES_DSN="host=localhost port=55432 user=postgres password=testpass dbname=trading_board_test sslmode=disable"
test-storage:
	go test ./... -count=1 -race

cover:
	go test ./... -count=1 -coverpkg=./internal/... -coverprofile=cover.out
	grep -v '/mocks/' cover.out > cover.nomocks.out
	go tool cover -func=cover.nomocks.out | tail -1

# Regenerates every mock; each interface file in internal/domain/interface/ carries its own directive.
mock:
	go generate ./...

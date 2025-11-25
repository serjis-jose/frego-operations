GO ?= go
SQLC ?= sqlc

.PHONY: generate sqlc run tidy build docker-build test clean

generate: sqlc

sqlc:
	$(SQLC) generate

run:
	$(GO) run ./cmd/server

build:
	$(GO) build -o bin/operations-server ./cmd/server

docker-build:
	docker build -t frego-operations:latest .

test:
	$(GO) test -v ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin/ internal/db/sqlc/ internal/api/operations.gen.go

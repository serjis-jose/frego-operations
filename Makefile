GO ?= go
SQLC ?= sqlc
OAPI ?= ~/go/bin/oapi-codegen

.PHONY: generate oapi sqlc run tidy sync-openapi build docker-build test clean

generate: oapi sqlc sync-openapi

oapi:
	$(OAPI) --config api/oapi-codegen.yaml api/operations_openapi.yaml

sqlc:
	$(SQLC) generate

sync-openapi:
	cp api/operations_openapi.yaml internal/server/openapi.yaml

run:
	$(GO) run ./cmd/server

build:
	$(GO) build -o bin/finance-server ./cmd/server

docker-build:
	docker build -t frego-operations:latest .

test:
	$(GO) test -v ./...

tidy:
	$(GO) mod tidy

clean:
	rm -rf bin/ internal/db/sqlc/ internal/api/finance.gen.go

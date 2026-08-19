MODULE       := github.com/fauzie/golang-sekeleton
PROTOC       := $(shell command -v protoc 2>/dev/null || echo tools/protoc/bin/protoc)
PROTO_DIR    := proto
THIRD_PARTY  := third_party/googleapis
DB_DRIVER    ?= mysql
DB_DSN_MYSQL := "mysql://root:root@tcp(localhost:3306)/golang_sekeleton"
DB_DSN_PG    := "postgres://postgres:postgres@localhost:5432/golang_sekeleton?sslmode=disable"

.PHONY: help
help:
	@echo "Common targets: run dev build test lint fmt proto migrate-up migrate-down docker-build compose-up compose-down"

## --- proto -----------------------------------------------------------------

.PHONY: install-proto-tools
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

.PHONY: proto
proto:
	$(PROTOC) \
		-I $(PROTO_DIR) -I $(THIRD_PARTY) \
		--go_out=$(PROTO_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(PROTO_DIR) --grpc-gateway_opt=paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=$(PROTO_DIR) \
		$(PROTO_DIR)/common.proto $(PROTO_DIR)/user.proto

## --- run / build -------------------------------------------------------------

.PHONY: run
run:
	go run ./cmd/server

.PHONY: dev
dev:
	air -c .air.toml

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

## --- quality -----------------------------------------------------------------

.PHONY: fmt
fmt:
	gofmt -w .
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test ./... -race

.PHONY: test-coverage
test-coverage:
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

## --- migrations ----------------------------------------------------------

.PHONY: install-migrate
install-migrate:
	go install -tags 'mysql postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: migrate-create
migrate-create:
	@read -p "migration name: " name; \
	migrate create -ext sql -dir migrations/$(DB_DRIVER) -seq $$name

.PHONY: migrate-up
migrate-up:
ifeq ($(DB_DRIVER),postgres)
	migrate -path migrations/postgres -database $(DB_DSN_PG) up
else
	migrate -path migrations/mysql -database $(DB_DSN_MYSQL) up
endif

.PHONY: migrate-down
migrate-down:
ifeq ($(DB_DRIVER),postgres)
	migrate -path migrations/postgres -database $(DB_DSN_PG) down 1
else
	migrate -path migrations/mysql -database $(DB_DSN_MYSQL) down 1
endif

## --- docker ----------------------------------------------------------------

.PHONY: docker-build
docker-build:
	docker build -f deployments/Dockerfile -t golang-sekeleton:latest .

.PHONY: compose-up
compose-up:
	docker compose up -d --build

.PHONY: compose-down
compose-down:
	docker compose down -v

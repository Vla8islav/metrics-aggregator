.PHONY: test build docker-build docker-run run db-start db-up

# allow override
DSN ?= postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable
GOOSE_MIGRATION_DIR=migrations
VERSION ?= v1.0.0
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_COMMIT := $(shell git rev-parse --short HEAD)

test:
	go test -race ./...

build:
	go build -ldflags "-X 'main.buildVersion=v1.0.0' \
                 -X 'main.buildDate=${BUILD_DATE}' \
                 -X 'main.buildCommit=${BUILD_COMMIT}'" \
                   -o bin/server ./cmd/server

run: build
	./bin/server

docker-build:
	docker buildx build --load -t metrics-aggregator:local .

docker-run: docker-build
	docker run --rm -p 8080:8080 --name metrics-aggregator metrics-aggregator:local

lint:
	go vet -vettool=$(which statictest) ./...

db-start:
	docker compose up -d

db-up: db-start
	goose -dir $(GOOSE_MIGRATION_DIR) postgres "$(DSN)" up

db-down: db-start
	goose -dir $(GOOSE_MIGRATION_DIR) postgres "$(DSN)" down

db-status: db-start
	goose -dir $(GOOSE_MIGRATION_DIR) postgres "$(DSN)" status

generate:
	go generate ./internal/mocks

generate-prot:
	protoc \
      --go_out=paths=source_relative:. \
      --go-grpc_out=paths=source_relative:. \
      --go_opt=default_api_level=API_OPAQUE \
      internal/proto/metrics.proto

cert-gen:
	go run ./cmd/generate_certs

static-lint:
	go run ./cmd/staticlint ./...
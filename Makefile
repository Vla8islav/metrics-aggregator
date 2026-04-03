.PHONY: test build docker-build docker-run run db-start db-up

# allow override
DSN ?= postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable

test:
	go test -race ./...

build:
	go build -o bin/server ./cmd/server

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
	goose postgres "$(DSN)"

db-down: db-start
	goose postgres "$(DSN)" down

db-status: db-start
	goose postgres "$(DSN)" status

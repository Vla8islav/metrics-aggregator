.PHONY: test build docker-build docker-run run

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

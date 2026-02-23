.PHONY: test build run

test:
	go test -race ./...

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

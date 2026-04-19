#!/bin/sh

go build -o cmd/agent/agent  cmd/agent/main.go
go build -o cmd/server/server  cmd/agent/main.go

          export SERVER_PORT=$(random unused-port)
          export SERVER_PORT=45101
          export ADDRESS="localhost:${SERVER_PORT}"
          export TEMP_FILE=$(random tempfile)
          ./metricstest -test.v -test.run=^TestIteration14$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn='postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable' \
            -key="${TEMP_FILE}" \
            -server-port=$SERVER_PORT \
            -source-path=./cmd/server

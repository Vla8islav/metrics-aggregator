FROM golang:1.25-bookworm AS builder
WORKDIR /app

# deps
COPY go.mod go.sum ./
RUN go mod download
# source
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
#build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /build/metrics-aggregator ./cmd/server
RUN ls /build/

# the app itself
FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /build/metrics-aggregator /usr/local/bin/metrics-aggregator

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/metrics-aggregator"]

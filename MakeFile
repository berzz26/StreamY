.PHONY: all build build-api build-worker run-api run-worker run-all test clean lint fmt vet tidy docker-up docker-down

# Binaries
API_BIN := bin/api
WORKER_BIN := bin/worker

# Go build flags
GOFLAGS := -ldflags="-s -w"

all: build

## Build both binaries
build: build-api build-worker

## Build API server binary
build-api:
	@mkdir -p bin
	go build $(GOFLAGS) -o $(API_BIN) ./cmd/api

## Build worker binary
build-worker:
	@mkdir -p bin
	go build $(GOFLAGS) -o $(WORKER_BIN) ./cmd/worker

## Run API server
run-api:
	go run ./cmd/api

## Run transcoding worker
run-worker:
	go run ./cmd/worker

## Run both API and worker (background)
run-all: run-api run-worker

## Run tests
test:
	go test -v ./...

## Run go vet
vet:
	go vet ./...

## Lint with golangci-lint
lint:
	golangci-lint run ./...

## Format code
fmt:
	gofmt -s -w .

## Tidy dependencies
tidy:
	go mod tidy

## Clean build artifacts
clean:
	rm -rf bin/
	rm -rf uploads/ processed/

## Download dependencies
deps:
	go mod download

## Show project info
info:
	@echo "StreamY - Video Streaming Platform"
	@echo "API:    go run ./cmd/api"
	@echo "Worker: go run ./cmd/worker"
	@echo "Port:   $${PORT:-8080}"

## Create required runtime directories
dirs:
	mkdir -p uploads processed storage/originals storage/processed

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' 2>/dev/null || sed -n '/^## /{s/^## /  /;p}' $(MAKEFILE_LIST)

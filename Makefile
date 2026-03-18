.PHONY: build test lint web dev docker-build clean help

BINARY  := openfiltr
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

## build: compile the server binary
build:
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARY) ./cmd/server

## test: run all Go tests
test:
	CGO_ENABLED=1 go test -race -count=1 ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## web: build the React frontend
web:
	cd web && npm ci && npm run build

## docker-build: build the Docker image
docker-build:
	docker build -t openfiltr:$(VERSION) -f deploy/docker/Dockerfile .

## clean: remove build artefacts
clean:
	rm -f $(BINARY) coverage.out
	rm -rf web/dist

## help: show available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'

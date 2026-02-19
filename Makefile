BINARY := drift
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X github.com/frankchan/drift/internal/cli.Version=$(VERSION) -X github.com/frankchan/drift/internal/cli.Commit=$(COMMIT) -X github.com/frankchan/drift/internal/cli.Date=$(DATE)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/drift

install:
	go install $(LDFLAGS) ./cmd/drift

test:
	go test ./...

test-integration:
	go test -tags=integration -count=1 ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

fmt:
	gofmt -s -w .

vet:
	go vet ./...

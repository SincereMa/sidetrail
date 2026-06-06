BINARY := cortex
PKG := ./...
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X github.com/SincereMa/cortex-sidemark/internal/version.Version=$(VERSION) \
           -X github.com/SincereMa/cortex-sidemark/internal/version.Commit=$(COMMIT)

.PHONY: all build test lint tidy run clean

all: build test

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test -race -count=1 $(PKG)

lint:
	golangci-lint run $(PKG)

tidy:
	go mod tidy

run: build
	./bin/$(BINARY) --version

clean:
	rm -rf bin/ dist/

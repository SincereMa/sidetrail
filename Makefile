BINARY := cortex
PKG := ./...
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -X github.com/SincereMa/cortex-sidemark/internal/version.Version=$(VERSION) \
           -X github.com/SincereMa/cortex-sidemark/internal/version.Commit=$(COMMIT)

.PHONY: all build test lint tidy run clean release-snapshot release-install release-install-windows

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

# Build local archives into ./dist/ without publishing. Use this to verify
# the .goreleaser.yml matrix and embed logic before tagging a release.
release-snapshot:
	goreleaser release --snapshot --clean --skip=publish

# Bootstraps a local cortex install from a snapshot build. Useful for
# smoke-testing the install script against artifacts in ./dist/.
release-install: release-snapshot
	CORTEX_VERSION=v0.0.0-dev ./scripts/install.sh --dir ./bin

# Syntax-check the install scripts without running them. CI runs both
# checks; locally, this is the cheap gate before pushing the PR.
release-install-windows:
	powershell -NoProfile -Command "$$null = [System.Management.Automation.Language.Parser]::ParseFile('scripts/install.ps1', [ref]$$null, [ref]$$null)"

clean:
	rm -rf bin/ dist/

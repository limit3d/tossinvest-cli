BINARY := bin/tossctl
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/junghoonkye/tossinvest-cli/internal/version.Version=$(VERSION) \
	-X github.com/junghoonkye/tossinvest-cli/internal/version.Commit=$(COMMIT) \
	-X github.com/junghoonkye/tossinvest-cli/internal/version.Date=$(DATE)

.PHONY: build build-mcp run test fmt tidy clean

build:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tossctl

build-mcp:
	mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/tossctl-mcp ./cmd/tossctl-mcp

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/tossctl

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

tidy:
	go mod tidy

clean:
	rm -rf bin coverage.out

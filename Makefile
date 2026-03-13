.PHONY: build test lint clean install run-dashboard

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0-dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X github.com/cqi/my_agentmux/cmd.Version=$(VERSION) \
           -X github.com/cqi/my_agentmux/cmd.GitCommit=$(GIT_COMMIT) \
           -X github.com/cqi/my_agentmux/cmd.BuildDate=$(BUILD_DATE)

BINARY := agentmux

## build: Build the agentmux binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## desktop: Build the desktop app (requires npm and Wails tags)
desktop: frontend-build
	CGO_ENABLED=1 CGO_LDFLAGS="-framework UniformTypeIdentifiers" go build -tags desktop,production -ldflags "$(LDFLAGS)" -o $(BINARY) .

## frontend-build: Build the Svelte frontend
frontend-build:
	cd frontend && npm run build

## test: Run all tests
test:
	go test ./... -v -count=1

## test-short: Run tests without verbose output
test-short:
	go test ./... -count=1

## lint: Run Go vet
lint:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/

## install: Install to $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## run-dashboard: Build and run the dashboard
run-dashboard: build
	./$(BINARY) dashboard

## fmt: Format Go code
fmt:
	gofmt -w .

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## help: Show this help
help:
	@echo "AgentMux — Multi-Agent Orchestrator"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /' | sed 's/:.*//' | column -t -s ':'

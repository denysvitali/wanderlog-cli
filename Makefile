BINARY_NAME := wanderlog
MODULE := github.com/denysvitali/wanderlog-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
COVERAGE_MIN ?= 35.0
LDFLAGS := -s -w \
	-X $(MODULE)/cmd.Version=$(VERSION) \
	-X $(MODULE)/cmd.Commit=$(COMMIT) \
	-X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

.PHONY: all build test test-race test-cover coverage-check integration integration-compile lint vet vulncheck fmt fmt-check toolchain-check tidy verify clean install run docker-build release-check help

all: fmt-check vet test build ## Run formatting, vet, tests, and build

build: ## Build the binary
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

install: ## Install the binary
	go install -trimpath -ldflags "$(LDFLAGS)" .

test: ## Run tests
	go test -shuffle=on ./...

test-race: ## Run tests with race detector
	go test -race ./...

test-cover: ## Run tests with coverage
	go test -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

coverage-check: test-cover ## Enforce the minimum total statement coverage
	./scripts/check_coverage.sh coverage.out $(COVERAGE_MIN)

integration-compile: ## Compile integration-tagged tests without contacting Wanderlog
	go test -tags=integration -run '^$$' ./...

integration: ## Run explicitly enabled production integration tests
	./test_integration.sh

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

vulncheck: ## Scan reachable code for known Go vulnerabilities
	govulncheck ./...

fmt: ## Format code
	gofmt -s -w .

fmt-check: ## Check formatting
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

toolchain-check: ## Check that go.mod and the container use the same Go release
	@version="$$(awk '$$1 == "go" { print $$2 }' go.mod)"; \
		grep -q "^FROM golang:$${version}-alpine AS builder$$" Dockerfile || \
		(echo "Dockerfile builder must use golang:$${version}-alpine"; exit 1)

tidy: ## Tidy and verify go.mod
	go mod tidy
	go mod verify

verify: ## Verify downloaded module checksums
	go mod verify

docker-build: ## Build the hardened container image
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t wanderlog-cli:$(VERSION) .

release-check: ## Validate the GoReleaser configuration
	goreleaser check

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME) coverage.out coverage.html

run: build ## Build and run
	./$(BINARY_NAME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

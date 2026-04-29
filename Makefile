BINARY_NAME := wanderlog
MODULE := github.com/denysvitali/wanderlog-cli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: all build generate extract-api test lint vet fmt clean install run help

all: lint test build ## Run lint, test, and build

build: ## Build the binary
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

generate: extract-api ## Generate extracted API call manifest

extract-api: ## Extract API call sites from the reference JS bundle
	python3 scripts/extract-api-calls.py

install: ## Install the binary
	go install -ldflags "$(LDFLAGS)" .

test: ## Run tests
	go test ./...

test-race: ## Run tests with race detector
	go test -race ./...

test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run golangci-lint
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofmt -s -w .

fmt-check: ## Check formatting
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

tidy: ## Tidy and verify go.mod
	go mod tidy
	go mod verify

clean: ## Remove build artifacts
	rm -f $(BINARY_NAME) coverage.out coverage.html

run: build ## Build and run
	./$(BINARY_NAME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

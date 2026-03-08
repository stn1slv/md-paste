.DEFAULT_GOAL := help

.PHONY: help setup test test-integration lint format build run upgrade-deps

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install dependencies and tools
	brew install gofumpt golangci-lint
	go mod tidy

test: ## Run unit tests
	go test ./...

test-integration: ## Run integration tests
	MD_PASTE_E2E=1 go test ./...

lint: ## Run linter
	golangci-lint run

format: ## Format code
	gofumpt -w -extra .

build: ## Build the application
	go build -o bin/md-paste ./cmd/md-paste

run: ## Run the application
	go run ./cmd/md-paste

upgrade-deps: ## Upgrade dependencies
	go get -u ./...
	go mod tidy

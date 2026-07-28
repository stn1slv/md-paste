.DEFAULT_GOAL := help

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)

.PHONY: help setup test test-integration lint format build run upgrade-deps icns bundle

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

icns: ## Generate the app icon (build/icon.icns) from build/icon-1024.png
	./scripts/make-icns.sh

bundle: ## Build the macOS menu bar .app bundle and zip (VERSION overridable)
	./scripts/build-app.sh $(VERSION)

upgrade-deps: ## Upgrade dependencies
	go get -u ./...
	go mod tidy

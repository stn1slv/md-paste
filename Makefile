.PHONY: setup test lint format build run upgrade-deps

setup:
	go mod tidy

test:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run

format:
	go run mvdan.cc/gofumpt@latest -w .

build:
	go build -o bin/md-paste ./cmd/md-paste

run:
	go run ./cmd/md-paste

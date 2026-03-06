.PHONY: setup test lint format build run upgrade-deps

setup:
	go mod tidy

test:
	go test ./...

lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6 run

format:
	go run mvdan.cc/gofumpt@v0.7.0 -w .

build:
	go build -o bin/md-paste ./cmd/md-paste

run:
	go run ./cmd/md-paste

upgrade-deps:
	go get -u ./...
	go mod tidy

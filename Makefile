.PHONY: build test clean lint fmt vet help

BINARY_NAME=evmql


build:
	@echo "Building evmql..."
	@go build -o bin/evmql cmd/evmql/main.go

test:
	@go test ./...

test-cover:
	@go test -cover -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

lint:
	@golangci-lint run

fmt:
	@go fmt ./...

vet:
	@go vet ./...

clean:
	@rm -f $(BINARY_NAME) coverage.out coverage.html

.DEFAULT_GOAL := help

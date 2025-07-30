# Transcriber Service Makefile

.PHONY: build-docker build-native test lint clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  build-docker  - Build Docker image"
	@echo "  build-native  - Build native binary"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"

# Build Docker image
build-docker:
	@echo "Building transcriber Docker image..."
	docker build -t speakr/transcriber:latest .

# Build native binary
build-native:
	@echo "Building transcriber native binary..."
	go mod tidy
	go build -o bin/transcriber ./cmd

# Run tests
test:
	@echo "Running transcriber tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running transcriber linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning transcriber build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker rmi speakr/transcriber:latest 2>/dev/null || true
# Embedding Service Makefile

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
	@echo "Building embedder Docker image..."
	docker build -t speakr/embedder:latest .

# Build native binary
build-native:
	@echo "Building embedder native binary..."
	go mod tidy
	go build -o bin/embedder ./cmd

# Run tests
test:
	@echo "Running embedder tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@echo "Running embedder linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning embedder build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker rmi speakr/embedder:latest 2>/dev/null || true
# Query Service Makefile

.PHONY: build-native build-docker test clean run

# Build native binary
build-native:
	@echo "Building query service native binary..."
	go build -o bin/query_svc ./cmd

# Build Docker image
build-docker:
	@echo "Building query service Docker image..."
	docker build -t speakr/query_svc:latest .

# Run tests
test:
	@echo "Running query service tests..."
	go test -v ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TEST=true go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	docker rmi speakr/query_svc:latest 2>/dev/null || true

# Run the service locally
run: build-native
	@echo "Running query service..."
	./bin/query_svc

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run
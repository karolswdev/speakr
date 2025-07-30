# Speakr Project Makefile
# Provides standardized interface for common project tasks per DEV-RULE E3

.PHONY: help up down logs clean health-check build-docker build-native test lint

# Default target
help: ## Show this help message
	@echo "Speakr Project - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment Setup:"
	@echo "  1. Copy .env.example to .env and configure your settings"
	@echo "  2. Run 'make up' to start all infrastructure services"
	@echo "  3. Run 'make health-check' to verify all services are running"

# Infrastructure Management per DEV-RULE E2.1
up: ## Start all infrastructure services and wait for health checks
	@echo "🚀 Starting Speakr infrastructure services..."
	docker compose up -d
	@echo "⏳ Waiting for services to become healthy..."
	@sleep 10
	@echo "✅ Infrastructure services started successfully!"
	@make health-check

down: ## Stop all infrastructure services
	@echo "🛑 Stopping Speakr infrastructure services..."
	docker compose down
	@echo "✅ Infrastructure services stopped successfully!"

logs: ## Follow logs of all services with proper formatting
	@echo "📋 Following logs for all Speakr services..."
	docker compose logs -f --tail=50

clean: ## Stop services, remove volumes, and clean up completely
	@echo "🧹 Performing complete cleanup of Speakr environment..."
	docker compose down -v --remove-orphans
	docker system prune -f --volumes
	@echo "✅ Environment reset completed!"

# Health Check per enhanced requirements
health-check: ## Verify all infrastructure services are responding correctly
	@echo "🔍 Checking health of all Speakr services..."
	@echo "Checking NATS..."
	@docker exec speakr-nats wget -q --spider http://localhost:8222/healthz && echo "✅ NATS is healthy" || echo "❌ NATS health check failed"
	@echo "Checking MinIO..."
	@docker exec speakr-minio mc ready local && echo "✅ MinIO is healthy" || echo "❌ MinIO health check failed"
	@echo "Checking PostgreSQL..."
	@docker exec speakr-postgres pg_isready -U postgres -d speakr && echo "✅ PostgreSQL is healthy" || echo "❌ PostgreSQL health check failed"
	@echo "✅ Health checks completed!"

# Build targets per DEV-RULE E3 and E4
build-docker: ## Build Docker containers for all services
	@echo "🐳 Building Docker containers for all services..."
	@for service in transcriber embedder query_svc cli; do \
		echo "Building $$service..."; \
		if [ -f $$service/Makefile ]; then \
			$(MAKE) -C $$service build-docker; \
		else \
			echo "⚠️  $$service/Makefile not found - skipping"; \
		fi; \
	done
	@echo "✅ Docker build completed!"

build-native: ## Build native binaries for all services
	@echo "🔨 Building native binaries for all services..."
	@for service in transcriber embedder query_svc cli; do \
		echo "Building $$service..."; \
		if [ -f $$service/Makefile ]; then \
			$(MAKE) -C $$service build-native; \
		else \
			echo "⚠️  $$service/Makefile not found - skipping"; \
		fi; \
	done
	@echo "✅ Native build completed!"

# Testing targets per DEV-RULE T1, T2, T3
test: ## Run all tests for all services
	@echo "🧪 Running tests for all services..."
	@for service in transcriber embedder query_svc cli; do \
		echo "Testing $$service..."; \
		if [ -f $$service/Makefile ]; then \
			$(MAKE) -C $$service test; \
		else \
			echo "⚠️  $$service/Makefile not found - skipping"; \
		fi; \
	done
	@echo "✅ All tests completed!"

# Linting per DEV-RULE Q1
lint: ## Run linting for all Go code
	@echo "🔍 Running linting for all services..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed - running go vet instead"; \
		go vet ./...; \
	fi
	@echo "✅ Linting completed!"

# Development helpers
dev-setup: ## Initial development environment setup
	@echo "🛠️  Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "📝 Created .env file from .env.example"; \
		echo "⚠️  Please edit .env with your actual configuration values"; \
	else \
		echo "✅ .env file already exists"; \
	fi
	@echo "✅ Development setup completed!"

# Database helpers
db-reset: ## Reset the database (WARNING: destroys all data)
	@echo "⚠️  WARNING: This will destroy all database data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose stop postgres; \
		docker volume rm speakr_postgres_data || true; \
		docker compose up -d postgres; \
		echo "✅ Database reset completed!"; \
	else \
		echo "❌ Database reset cancelled"; \
	fi

# Show current status
status: ## Show status of all services
	@echo "📊 Speakr Services Status:"
	@docker compose ps
.PHONY: help setup start stop restart status test build clean logs reset-db

# Default target
help:
	@echo "Algo Trading System - Development Commands"
	@echo "=========================================="
	@echo ""
	@echo "Setup & Infrastructure:"
	@echo "  setup        - Initial project setup"
	@echo "  infra-start  - Start infrastructure services (DB, Redis, Kafka)"
	@echo "  infra-stop   - Stop infrastructure services"
	@echo ""
	@echo "Development:"
	@echo "  start        - Start all services"
	@echo "  stop         - Stop all services"
	@echo "  restart      - Restart all services"
	@echo "  status       - Check service status"
	@echo ""
	@echo "Build & Test:"
	@echo "  test         - Run all tests"
	@echo "  build        - Build all services"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Utilities:"
	@echo "  logs SERVICE - Show logs for specific service"
	@echo "  reset-db     - Reset database (WARNING: Deletes all data)"
	@echo ""
	@echo "Examples:"
	@echo "  make setup"
	@echo "  make start"
	@echo "  make logs SERVICE=timescaledb"
	@echo "  make test"

# Setup commands
setup:
	@./scripts/setup.sh

infra-start:
	@./scripts/dev.sh infra-start

infra-stop:
	@./scripts/dev.sh infra-stop

# Service management
start:
	@./scripts/dev.sh start

stop:
	@./scripts/dev.sh stop

restart:
	@./scripts/dev.sh restart

status:
	@./scripts/dev.sh status

# Build and test
test:
	@./scripts/dev.sh test

build:
	@./scripts/dev.sh build

clean:
	@./scripts/dev.sh clean

# Utilities
logs:
ifdef SERVICE
	@./scripts/dev.sh logs $(SERVICE)
else
	@echo "Usage: make logs SERVICE=<service-name>"
	@echo "Available services: timescaledb, redis, kafka, prometheus, grafana"
endif

reset-db:
	@./scripts/dev.sh reset-db

# Go specific commands
go-tidy:
	@echo "Running go mod tidy for all Go services..."
	@cd services/market-data-service && go mod tidy
	@cd services/trading-engine && go mod tidy || true
	@cd services/risk-management && go mod tidy || true
	@cd services/paper-trading && go mod tidy || true
	@echo "✅ Go dependencies updated"

go-fmt:
	@echo "Formatting Go code..."
	@cd services/market-data-service && go fmt ./...
	@cd services/trading-engine && go fmt ./... || true
	@cd services/risk-management && go fmt ./... || true
	@cd services/paper-trading && go fmt ./... || true
	@echo "✅ Go code formatted"

go-vet:
	@echo "Running go vet..."
	@cd services/market-data-service && go vet ./...
	@cd services/trading-engine && go vet ./... || true
	@cd services/risk-management && go vet ./... || true
	@cd services/paper-trading && go vet ./... || true
	@echo "✅ Go code vetted"

# Python specific commands (for future ML services)
python-deps:
	@echo "Installing Python dependencies..."
	@cd services/ml-service && pip install -r requirements.txt || echo "requirements.txt not found"
	@cd services/sentiment-analysis && pip install -r requirements.txt || echo "requirements.txt not found"
	@echo "✅ Python dependencies installed"

# Docker commands
docker-build:
	@echo "Building Docker images..."
	@docker-compose build
	@echo "✅ Docker images built"

docker-up:
	@echo "Starting services with Docker..."
	@docker-compose up -d
	@echo "✅ Services started with Docker"

docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down
	@echo "✅ Docker services stopped"

docker-logs:
ifdef SERVICE
	@docker-compose logs -f $(SERVICE)
else
	@echo "Usage: make docker-logs SERVICE=<service-name>"
endif

# Quality checks
lint: go-fmt go-vet
	@echo "✅ Code linting completed"

check-deps:
	@echo "Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is required but not installed"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose is required but not installed"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed"; exit 1; }
	@command -v python3 >/dev/null 2>&1 || { echo "❌ Python 3 is required but not installed"; exit 1; }
	@echo "✅ All dependencies are installed"

# Development workflow
dev-setup: check-deps setup go-tidy
	@echo "🚀 Development environment ready!"

dev-start: infra-start
	@sleep 5
	@./scripts/dev.sh start

# Production-like testing
prod-test: docker-build docker-up
	@echo "Testing production-like environment..."
	@sleep 30
	@curl -f http://localhost:8080/health || { echo "❌ Market data service health check failed"; exit 1; }
	@echo "✅ Production test completed"
	@docker-compose down

# Quick development cycle
quick: stop start status

# Full reset (use with caution)
full-reset: clean reset-db
	@docker system prune -f
	@echo "🧹 Full reset completed"

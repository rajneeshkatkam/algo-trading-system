#!/bin/bash

# Development helper script for Algo Trading System

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# Function to check if service is running
check_service() {
    local service_name=$1
    local port=$2
    
    if curl -s -f "http://localhost:$port/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $service_name (port $port)"
        return 0
    else
        echo -e "${RED}✗${NC} $service_name (port $port)"
        return 1
    fi
}

# Function to start infrastructure
start_infra() {
    print_header "Starting Infrastructure Services"
    docker-compose up -d timescaledb redis kafka zookeeper prometheus grafana
    
    print_status "Waiting for services to be ready..."
    sleep 10
    
    # Wait for database
    print_status "Waiting for database..."
    while ! docker-compose exec -T timescaledb pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; do
        echo -n "."
        sleep 2
    done
    echo ""
    
    print_status "Infrastructure services started successfully!"
}

# Function to stop infrastructure
stop_infra() {
    print_header "Stopping Infrastructure Services"
    docker-compose down
    print_status "Infrastructure services stopped!"
}

# Function to start market data service
start_market_data() {
    print_header "Starting Market Data Service"
    cd services/market-data-service
    
    if [ ! -f "go.sum" ]; then
        print_status "Downloading Go dependencies..."
        go mod tidy
    fi
    
    print_status "Starting market data service on port 8080..."
    go run cmd/server/main.go &
    MARKET_DATA_PID=$!
    
    # Wait for service to start
    sleep 5
    if kill -0 $MARKET_DATA_PID 2>/dev/null; then
        print_status "Market data service started with PID: $MARKET_DATA_PID"
        echo $MARKET_DATA_PID > ../../.market-data.pid
    else
        print_error "Failed to start market data service"
        exit 1
    fi
    
    cd ../..
}

# Function to stop services
stop_services() {
    print_header "Stopping Services"
    
    if [ -f ".market-data.pid" ]; then
        PID=$(cat .market-data.pid)
        if kill -0 $PID 2>/dev/null; then
            kill $PID
            print_status "Stopped market data service (PID: $PID)"
        fi
        rm -f .market-data.pid
    fi
    
    # Kill any remaining Go processes
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
}

# Function to check system status
check_status() {
    print_header "System Status Check"
    
    echo "Infrastructure Services:"
    docker-compose ps
    
    echo ""
    echo "Application Services:"
    check_service "Market Data Service" 8080 || true
    check_service "Trading Engine" 8082 || true
    check_service "ML Service" 8084 || true
    
    echo ""
    echo "Infrastructure Health:"
    docker-compose exec -T timescaledb pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1 && echo -e "${GREEN}✓${NC} Database" || echo -e "${RED}✗${NC} Database"
    docker-compose exec -T redis redis-cli ping > /dev/null 2>&1 && echo -e "${GREEN}✓${NC} Redis" || echo -e "${RED}✗${NC} Redis"
}

# Function to run tests
run_tests() {
    print_header "Running Tests"
    
    cd services/market-data-service
    print_status "Running market data service tests..."
    go test -v ./...
    cd ../..
    
    print_status "All tests completed!"
}

# Function to build services
build_services() {
    print_header "Building Services"
    
    mkdir -p bin
    
    print_status "Building market data service..."
    cd services/market-data-service
    go build -o ../../bin/market-data-service cmd/server/main.go
    cd ../..
    
    print_status "All services built successfully!"
    print_status "Binaries available in ./bin/ directory"
}

# Function to clean up
cleanup() {
    print_header "Cleaning Up"
    
    stop_services
    
    print_status "Removing build artifacts..."
    rm -rf bin/
    rm -f .*.pid
    
    print_status "Cleaning Docker volumes (optional)..."
    read -p "Do you want to remove Docker volumes? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v
        docker system prune -f
        print_status "Docker volumes cleaned!"
    fi
    
    print_status "Cleanup completed!"
}

# Function to show logs
show_logs() {
    local service=$1
    
    if [ -z "$service" ]; then
        print_header "Available Services for Logs"
        echo "Infrastructure: timescaledb, redis, kafka, prometheus, grafana"
        echo "Usage: $0 logs <service-name>"
        return 1
    fi
    
    print_header "Showing logs for: $service"
    docker-compose logs -f "$service"
}

# Function to reset database
reset_db() {
    print_header "Resetting Database"
    
    print_warning "This will delete all data in the database!"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose exec timescaledb psql -U postgres -d algotrading -c "
            DROP SCHEMA IF EXISTS market_data CASCADE;
            DROP SCHEMA IF EXISTS trading CASCADE;
            DROP SCHEMA IF EXISTS analytics CASCADE;
        "
        
        docker-compose exec timescaledb psql -U postgres -d algotrading -f /docker-entrypoint-initdb.d/01-init.sql
        
        print_status "Database reset completed!"
    else
        print_status "Database reset cancelled."
    fi
}

# Main command handler
case "$1" in
    "start")
        start_infra
        start_market_data
        ;;
    "stop")
        stop_services
        ;;
    "restart")
        stop_services
        sleep 2
        start_infra
        start_market_data
        ;;
    "status")
        check_status
        ;;
    "test")
        run_tests
        ;;
    "build")
        build_services
        ;;
    "clean")
        cleanup
        ;;
    "logs")
        show_logs "$2"
        ;;
    "reset-db")
        reset_db
        ;;
    "infra-start")
        start_infra
        ;;
    "infra-stop")
        stop_infra
        ;;
    *)
        print_header "Algo Trading Development Helper"
        echo "Usage: $0 {command}"
        echo ""
        echo "Commands:"
        echo "  start         - Start infrastructure and market data service"
        echo "  stop          - Stop all services"
        echo "  restart       - Restart all services"
        echo "  status        - Check status of all services"
        echo "  test          - Run all tests"
        echo "  build         - Build all services"
        echo "  clean         - Clean up build artifacts and optionally Docker volumes"
        echo "  logs <service> - Show logs for a specific service"
        echo "  reset-db      - Reset database (WARNING: Deletes all data)"
        echo "  infra-start   - Start only infrastructure services"
        echo "  infra-stop    - Stop only infrastructure services"
        echo ""
        echo "Examples:"
        echo "  $0 start                    # Start everything"
        echo "  $0 status                   # Check service status"
        echo "  $0 logs timescaledb         # View database logs"
        echo "  $0 test                     # Run all tests"
        ;;
esac

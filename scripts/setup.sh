#!/bin/bash

# Algo Trading System Setup Script
# This script sets up the development environment

set -e

echo "🚀 Setting up Algo Trading System..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Python is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Please install Python 3.11+ first."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Create necessary directories
echo "📁 Creating project directories..."
mkdir -p logs
mkdir -p data/models
mkdir -p data/backups
mkdir -p config

echo "🐳 Starting infrastructure services..."
docker-compose up -d timescaledb redis kafka zookeeper

echo "⏳ Waiting for services to be ready..."
sleep 30

# Check if TimescaleDB is ready
echo "🔍 Checking database connection..."
max_attempts=30
attempt=1
while ! docker-compose exec -T timescaledb pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1; do
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Database failed to start within expected time"
        exit 1
    fi
    echo "Waiting for database... (attempt $attempt/$max_attempts)"
    sleep 2
    ((attempt++))
done

echo "✅ Database is ready"

# Check if Redis is ready
echo "🔍 Checking Redis connection..."
max_attempts=10
attempt=1
while ! docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; do
    if [ $attempt -eq $max_attempts ]; then
        echo "❌ Redis failed to start within expected time"
        exit 1
    fi
    echo "Waiting for Redis... (attempt $attempt/$max_attempts)"
    sleep 2
    ((attempt++))
done

echo "✅ Redis is ready"

# Initialize Go modules for services
echo "🔧 Initializing Go services..."

cd services/market-data-service
if [ ! -f "go.sum" ]; then
    go mod tidy
fi
cd ../..

# Create basic configuration files
echo "⚙️ Creating configuration files..."

cat > config/market-data.yaml << EOL
server:
  http_port: 8080
  grpc_port: 8081

database:
  host: localhost
  port: 5432
  name: algotrading
  user: postgres
  password: password123

redis:
  host: localhost
  port: 6379

kafka:
  brokers:
    - localhost:9092

api_providers:
  angel_one:
    enabled: false
    api_key: ""
    api_secret: ""
  
  mock:
    enabled: true
    symbols:
      - RELIANCE
      - TCS
      - HDFCBANK
      - INFY
      - HINDUNILVR

logging:
  level: info
  format: json

EOL

cat > config/trading-engine.yaml << EOL
server:
  http_port: 8082
  grpc_port: 8083

database:
  host: localhost
  port: 5432
  name: algotrading
  user: postgres
  password: password123

redis:
  host: localhost
  port: 6379

market_data_service:
  grpc_address: localhost:8081

risk_management:
  max_position_size: 100000
  max_daily_loss: 50000
  max_positions: 10

strategies:
  - name: "momentum_strategy"
    enabled: false
    parameters:
      rsi_period: 14
      rsi_oversold: 30
      rsi_overbought: 70

EOL

# Create environment file
cat > .env << EOL
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=algotrading
DB_USER=postgres
DB_PASSWORD=password123

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379

# Kafka Configuration
KAFKA_BROKERS=localhost:9092

# API Keys (Fill these in)
ANGEL_ONE_API_KEY=
ANGEL_ONE_API_SECRET=

# Environment
ENVIRONMENT=development
LOG_LEVEL=info

EOL

echo "📊 Setting up monitoring..."
docker-compose up -d prometheus grafana

echo "🎉 Setup completed successfully!"
echo ""
echo "📋 Next steps:"
echo "1. Start the market data service:"
echo "   cd services/market-data-service && go run cmd/server/main.go"
echo ""
echo "2. Access the services:"
echo "   - Database (pgAdmin): http://localhost:5050"
echo "   - Redis: localhost:6379"
echo "   - Prometheus: http://localhost:9090"
echo "   - Grafana: http://localhost:3000 (admin/admin123)"
echo ""
echo "3. Configure API credentials in config/market-data.yaml"
echo ""
echo "4. Test the API:"
echo "   curl http://localhost:8080/health"
echo ""
echo "📖 Check README.md for detailed documentation"

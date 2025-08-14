# Algo Trading System

A comprehensive automated trading system for the Indian stock market with real-time data processing, machine learning-based predictions, and risk management.

## 🏗️ Architecture

This system follows a microservices architecture with the following components:

- **Market Data Service** (Go) - Real-time data ingestion from Indian stock market APIs
- **Trading Engine** (Go) - Core trading logic and order execution
- **ML Service** (Python) - Machine learning models for price prediction
- **Sentiment Analysis** (Python) - News and social media sentiment analysis
- **Risk Management** (Go) - Position sizing and risk controls
- **Paper Trading** (Go) - Backtesting and paper trading simulation
- **Dashboard** (React/Next.js) - Web interface for monitoring and control

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Python 3.11+
- Docker & Docker Compose
- PostgreSQL/TimescaleDB

### Setup
```bash
# Clone and setup
git clone <your-repo>
cd algo-trading

# Start infrastructure
docker-compose up -d

# Initialize services
./scripts/setup.sh
```

## 📁 Project Structure

```
algo-trading/
├── services/                    # Microservices
│   ├── market-data-service/    # Real-time market data (Go)
│   ├── trading-engine/         # Core trading logic (Go)
│   ├── ml-service/             # ML models (Python)
│   ├── sentiment-analysis/     # Sentiment analysis (Python)
│   ├── risk-management/        # Risk controls (Go)
│   └── paper-trading/          # Paper trading (Go)
├── shared/                     # Shared libraries
│   ├── proto/                  # gRPC definitions
│   └── utils/                  # Common utilities
├── infrastructure/             # Infrastructure as code
│   ├── docker/                 # Docker configurations
│   └── kubernetes/             # K8s manifests
├── dashboard/                  # Web dashboard
├── scripts/                    # Automation scripts
├── docs/                       # Documentation
└── tests/                      # Integration tests
```

## 🛠️ Development

### Local Development
```bash
# Start individual services
cd services/market-data-service
go run cmd/server/main.go

# Start ML service
cd services/ml-service
python -m uvicorn main:app --reload
```

## 📊 Features

- ✅ Real-time Indian stock market data
- ✅ Technical analysis indicators
- ✅ Fundamental analysis
- ✅ Sentiment analysis from news/social media
- ✅ Machine learning predictions
- ✅ Risk management
- ✅ Paper trading
- ✅ Web dashboard
- ✅ Scalable microservices architecture

## 📈 Supported Markets

- NSE (National Stock Exchange)
- BSE (Bombay Stock Exchange)
- Equity, F&O, Currency segments

## 🔑 API Providers

- Zerodha Kite Connect
- Angel One SmartAPI
- Upstox API
- AliceBlue API

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

Please read CONTRIBUTING.md for contribution guidelines.

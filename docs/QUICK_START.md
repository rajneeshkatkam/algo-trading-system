# Quick Start Guide - Algo Trading System

## 🚀 Get Started in 5 Minutes

### Prerequisites Check
```bash
# Verify you have the required tools
docker --version          # Should show Docker version
docker-compose --version  # Should show Docker Compose version
go version                # Should show Go 1.21+
python3 --version         # Should show Python 3.11+
```

### Step 1: Initial Setup
```bash
cd /Users/rajneeshkatkam/Documents/algo-trading

# Make scripts executable (if not already done)
chmod +x scripts/*.sh

# Run the setup script
./scripts/setup.sh
```

### Step 2: Start the System
```bash
# Start infrastructure and services
make start

# Wait for services to be ready (about 30 seconds)
# Check status
make status
```

### Step 3: Test the API
```bash
# Health check
curl http://localhost:8080/health

# Get list of stocks
curl http://localhost:8080/api/v1/stocks

# Get sample OHLCV data for Reliance
curl "http://localhost:8080/api/v1/stocks/RELIANCE/ohlcv"
```

### Step 4: Access Web Interfaces
- **pgAdmin**: http://localhost:5050 (admin@algo.com / admin123)
- **Grafana**: http://localhost:3000 (admin / admin123)
- **Prometheus**: http://localhost:9090

## 🔧 Development Workflow

### Daily Development Commands
```bash
# Start everything
make start

# Check system status
make status

# View logs for specific service
make logs SERVICE=timescaledb

# Run tests
make test

# Stop everything
make stop
```

### Working with the Market Data Service
```bash
cd services/market-data-service

# Run the service directly
go run cmd/server/main.go

# Run tests
go test ./...

# Format code
go fmt ./...
```

## 📊 WebSocket Real-Time Data

### Connect to WebSocket
```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

// Subscribe to a stock
ws.send(JSON.stringify({
    type: 'subscribe',
    symbol: 'RELIANCE'
}));

// Listen for data
ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};
```

## 🧪 Testing the Technical Indicators

```bash
cd services/market-data-service

# Test specific indicators
go test -v ./pkg/indicators -run TestCalculateSMA
go test -v ./pkg/indicators -run TestCalculateRSI
go test -v ./pkg/indicators -run TestCalculateMACD

# Benchmark indicators
go test -bench=. ./pkg/indicators
```

## 📈 Sample Data

The system comes with sample data for 10 major Indian stocks:
- RELIANCE, TCS, HDFCBANK, INFY, HINDUNILVR
- ICICIBANK, SBIN, BHARTIARTL, ITC, LT

## 🛠️ Common Issues & Solutions

### Issue: Database Connection Failed
```bash
# Check if database is running
docker-compose ps timescaledb

# Restart database
docker-compose restart timescaledb

# Check logs
make logs SERVICE=timescaledb
```

### Issue: Port Already in Use
```bash
# Find what's using the port
lsof -i :8080

# Stop the process or change port in config
```

### Issue: Go Dependencies
```bash
cd services/market-data-service
go mod tidy
go mod download
```

## 🔄 Reset Everything
```bash
# Full reset (careful - deletes all data)
make clean
make reset-db

# Or reset just the database
make reset-db
```

## 📝 Next Steps

1. **Configure API Keys**: Edit `config/market-data.yaml` with real API credentials
2. **Test Real Data**: Switch from mock to real API provider
3. **Add Strategies**: Start building trading strategies
4. **Monitor Performance**: Use Grafana dashboards

## 📞 Getting Help

- Check logs: `make logs SERVICE=<service-name>`
- Run health checks: `make status`
- View documentation: `cat README.md`
- Check database: Connect via pgAdmin

---

🎉 **You're all set!** Your algo trading system is running and ready for development.



Steps:
1. To start with just building statistical metrics (e.g. P/E, alpha, etc. 10-15 metrics) around the stocks and predicting if it is good or bad stock.
2. Then do sentiment analysis of the data by scraping the internet/news data (using pretrained LLM/Classfier online) - To further strengthen the stock selection algorithm (which was generated statistical metric model). - Expectation from this: It should give reasons for its sentiments (good/bad).
3. Based on few iterations for few days and short comings on step-1 and 2, we will deep dive more into ML.
4. Once step-1 and step-2, we can build an ML regression model which predicts the targets i.e. price values along with the predict days/date.
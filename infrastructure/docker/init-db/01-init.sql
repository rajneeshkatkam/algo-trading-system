-- Initialize TimescaleDB for algo trading system
-- This script runs automatically when the container starts

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create schemas
CREATE SCHEMA IF NOT EXISTS market_data;
CREATE SCHEMA IF NOT EXISTS trading;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Market Data Tables
CREATE TABLE IF NOT EXISTS market_data.stocks (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL UNIQUE,
    company_name VARCHAR(255),
    sector VARCHAR(100),
    market_cap BIGINT,
    exchange VARCHAR(10) NOT NULL, -- NSE, BSE
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS market_data.ohlcv (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    open DECIMAL(12,2) NOT NULL,
    high DECIMAL(12,2) NOT NULL,
    low DECIMAL(12,2) NOT NULL,
    close DECIMAL(12,2) NOT NULL,
    volume BIGINT NOT NULL,
    timeframe VARCHAR(10) NOT NULL, -- 1m, 5m, 15m, 1h, 1d
    PRIMARY KEY (time, symbol, timeframe)
);

-- Convert to hypertable for time-series optimization
SELECT create_hypertable('market_data.ohlcv', 'time', if_not_exists => TRUE);

-- Create index on symbol and time for faster queries
CREATE INDEX IF NOT EXISTS idx_ohlcv_symbol_time ON market_data.ohlcv (symbol, time DESC);
CREATE INDEX IF NOT EXISTS idx_ohlcv_timeframe ON market_data.ohlcv (timeframe, time DESC);

-- Real-time ticks table
CREATE TABLE IF NOT EXISTS market_data.ticks (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    volume BIGINT DEFAULT 0,
    bid DECIMAL(12,2),
    ask DECIMAL(12,2),
    PRIMARY KEY (time, symbol)
);

SELECT create_hypertable('market_data.ticks', 'time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_ticks_symbol_time ON market_data.ticks (symbol, time DESC);

-- Trading Tables
CREATE TABLE IF NOT EXISTS trading.strategies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parameters JSONB,
    status VARCHAR(20) DEFAULT 'inactive', -- active, inactive, paused
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS trading.orders (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES trading.strategies(id),
    symbol VARCHAR(50) NOT NULL,
    order_type VARCHAR(20) NOT NULL, -- market, limit, stop_loss
    side VARCHAR(10) NOT NULL, -- buy, sell
    quantity INTEGER NOT NULL,
    price DECIMAL(12,2),
    status VARCHAR(20) DEFAULT 'pending', -- pending, filled, cancelled, rejected
    filled_quantity INTEGER DEFAULT 0,
    filled_price DECIMAL(12,2),
    order_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    filled_time TIMESTAMP WITH TIME ZONE,
    broker_order_id VARCHAR(100),
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_orders_symbol_time ON trading.orders (symbol, order_time DESC);
CREATE INDEX IF NOT EXISTS idx_orders_strategy ON trading.orders (strategy_id, order_time DESC);

-- Portfolio tracking
CREATE TABLE IF NOT EXISTS trading.positions (
    id SERIAL PRIMARY KEY,
    strategy_id INTEGER REFERENCES trading.strategies(id),
    symbol VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    average_price DECIMAL(12,2) NOT NULL,
    current_price DECIMAL(12,2),
    unrealized_pnl DECIMAL(12,2),
    realized_pnl DECIMAL(12,2) DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(strategy_id, symbol)
);

-- Analytics Tables
CREATE TABLE IF NOT EXISTS analytics.technical_indicators (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    timeframe VARCHAR(10) NOT NULL,
    indicator_name VARCHAR(50) NOT NULL,
    value DECIMAL(12,4),
    metadata JSONB,
    PRIMARY KEY (time, symbol, timeframe, indicator_name)
);

SELECT create_hypertable('analytics.technical_indicators', 'time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_indicators_symbol ON analytics.technical_indicators (symbol, indicator_name, time DESC);

-- Sentiment analysis
CREATE TABLE IF NOT EXISTS analytics.sentiment_scores (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    symbol VARCHAR(50),
    source VARCHAR(50) NOT NULL, -- news, twitter, reddit
    sentiment_score DECIMAL(5,4), -- -1 to 1
    confidence DECIMAL(5,4),
    content_hash VARCHAR(64),
    metadata JSONB,
    PRIMARY KEY (time, source, content_hash)
);

SELECT create_hypertable('analytics.sentiment_scores', 'time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_sentiment_symbol ON analytics.sentiment_scores (symbol, time DESC);

-- ML Predictions
CREATE TABLE IF NOT EXISTS analytics.predictions (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    model_name VARCHAR(100) NOT NULL,
    prediction_type VARCHAR(50) NOT NULL, -- price, direction, volatility
    predicted_value DECIMAL(12,4),
    confidence DECIMAL(5,4),
    target_time TIMESTAMP WITH TIME ZONE,
    actual_value DECIMAL(12,4),
    metadata JSONB,
    PRIMARY KEY (time, symbol, model_name, prediction_type)
);

SELECT create_hypertable('analytics.predictions', 'time', if_not_exists => TRUE);
CREATE INDEX IF NOT EXISTS idx_predictions_symbol ON analytics.predictions (symbol, model_name, time DESC);

-- Create some sample data
INSERT INTO market_data.stocks (symbol, company_name, sector, exchange) VALUES
('RELIANCE', 'Reliance Industries Limited', 'Energy', 'NSE'),
('TCS', 'Tata Consultancy Services', 'Information Technology', 'NSE'),
('HDFCBANK', 'HDFC Bank Limited', 'Financial Services', 'NSE'),
('INFY', 'Infosys Limited', 'Information Technology', 'NSE'),
('HINDUNILVR', 'Hindustan Unilever Limited', 'FMCG', 'NSE'),
('ICICIBANK', 'ICICI Bank Limited', 'Financial Services', 'NSE'),
('SBIN', 'State Bank of India', 'Financial Services', 'NSE'),
('BHARTIARTL', 'Bharti Airtel Limited', 'Telecommunication', 'NSE'),
('ITC', 'ITC Limited', 'FMCG', 'NSE'),
('LT', 'Larsen & Toubro Limited', 'Construction', 'NSE')
ON CONFLICT (symbol) DO NOTHING;

-- Create a default strategy
INSERT INTO trading.strategies (name, description, parameters, status) VALUES
('buy_and_hold', 'Simple buy and hold strategy', '{"risk_percentage": 2, "max_positions": 10}', 'inactive'),
('momentum_strategy', 'Momentum-based trading strategy', '{"rsi_period": 14, "rsi_oversold": 30, "rsi_overbought": 70}', 'inactive')
ON CONFLICT (name) DO NOTHING;

-- Set up retention policies for time-series data
-- Keep tick data for 30 days
SELECT add_retention_policy('market_data.ticks', INTERVAL '30 days', if_not_exists => TRUE);

-- Keep 1-minute OHLCV for 1 year, aggregate to higher timeframes
SELECT add_retention_policy('market_data.ohlcv', INTERVAL '2 years', if_not_exists => TRUE);

-- Keep technical indicators for 1 year
SELECT add_retention_policy('analytics.technical_indicators', INTERVAL '1 year', if_not_exists => TRUE);

-- Keep sentiment data for 6 months
SELECT add_retention_policy('analytics.sentiment_scores', INTERVAL '6 months', if_not_exists => TRUE);

COMMIT;

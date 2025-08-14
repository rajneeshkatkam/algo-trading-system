package storage

import (
    "database/sql"
    "fmt"
    "time"

    _ "github.com/lib/pq"
    "github.com/algo-trading/market-data-service/internal/models"
)

type Database struct {
    db *sql.DB
}

func NewDatabase(host, port, dbname, user, password string) (*Database, error) {
    psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname)
    
    db, err := sql.Open("postgres", psqlInfo)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    return &Database{db: db}, nil
}

func (d *Database) Close() error {
    return d.db.Close()
}

// Stock operations
func (d *Database) GetStocks() ([]models.Stock, error) {
    query := `
        SELECT id, symbol, company_name, sector, market_cap, exchange, created_at, updated_at
        FROM market_data.stocks
        ORDER BY symbol
    `
    
    rows, err := d.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query stocks: %w", err)
    }
    defer rows.Close()

    var stocks []models.Stock
    for rows.Next() {
        var stock models.Stock
        err := rows.Scan(
            &stock.ID, &stock.Symbol, &stock.CompanyName, &stock.Sector,
            &stock.MarketCap, &stock.Exchange, &stock.CreatedAt, &stock.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan stock row: %w", err)
        }
        stocks = append(stocks, stock)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating stocks: %w", err)
    }

    return stocks, nil
}

func (d *Database) GetStock(symbol string) (*models.Stock, error) {
    query := `
        SELECT id, symbol, company_name, sector, market_cap, exchange, created_at, updated_at
        FROM market_data.stocks
        WHERE symbol = $1
    `
    
    var stock models.Stock
    err := d.db.QueryRow(query, symbol).Scan(
        &stock.ID, &stock.Symbol, &stock.CompanyName, &stock.Sector,
        &stock.MarketCap, &stock.Exchange, &stock.CreatedAt, &stock.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("stock %s not found", symbol)
        }
        return nil, fmt.Errorf("failed to get stock: %w", err)
    }

    return &stock, nil
}

// OHLCV operations
func (d *Database) InsertOHLCV(ohlcv *models.OHLCV) error {
    query := `
        INSERT INTO market_data.ohlcv (time, symbol, open, high, low, close, volume, timeframe)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (time, symbol, timeframe) DO UPDATE SET
            open = EXCLUDED.open,
            high = EXCLUDED.high,
            low = EXCLUDED.low,
            close = EXCLUDED.close,
            volume = EXCLUDED.volume
    `
    
    _, err := d.db.Exec(query, ohlcv.Time, ohlcv.Symbol, ohlcv.Open, ohlcv.High,
        ohlcv.Low, ohlcv.Close, ohlcv.Volume, ohlcv.Timeframe)
    
    if err != nil {
        return fmt.Errorf("failed to insert OHLCV: %w", err)
    }

    return nil
}

func (d *Database) GetOHLCV(symbol, timeframe string, start, end time.Time, limit int) ([]models.OHLCV, error) {
    query := `
        SELECT time, symbol, open, high, low, close, volume, timeframe
        FROM market_data.ohlcv
        WHERE symbol = $1 AND timeframe = $2 AND time >= $3 AND time <= $4
        ORDER BY time DESC
        LIMIT $5
    `
    
    rows, err := d.db.Query(query, symbol, timeframe, start, end, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to query OHLCV: %w", err)
    }
    defer rows.Close()

    var ohlcvs []models.OHLCV
    for rows.Next() {
        var ohlcv models.OHLCV
        err := rows.Scan(
            &ohlcv.Time, &ohlcv.Symbol, &ohlcv.Open, &ohlcv.High,
            &ohlcv.Low, &ohlcv.Close, &ohlcv.Volume, &ohlcv.Timeframe,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan OHLCV row: %w", err)
        }
        ohlcvs = append(ohlcvs, ohlcv)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating OHLCV: %w", err)
    }

    return ohlcvs, nil
}

// Tick operations
func (d *Database) InsertTick(tick *models.Tick) error {
    query := `
        INSERT INTO market_data.ticks (time, symbol, price, volume, bid, ask)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (time, symbol) DO UPDATE SET
            price = EXCLUDED.price,
            volume = EXCLUDED.volume,
            bid = EXCLUDED.bid,
            ask = EXCLUDED.ask
    `
    
    _, err := d.db.Exec(query, tick.Time, tick.Symbol, tick.Price, tick.Volume, tick.Bid, tick.Ask)
    
    if err != nil {
        return fmt.Errorf("failed to insert tick: %w", err)
    }

    return nil
}

func (d *Database) GetTicks(symbol string, start, end time.Time, limit int) ([]models.Tick, error) {
    query := `
        SELECT time, symbol, price, volume, bid, ask
        FROM market_data.ticks
        WHERE symbol = $1 AND time >= $2 AND time <= $3
        ORDER BY time DESC
        LIMIT $4
    `
    
    rows, err := d.db.Query(query, symbol, start, end, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to query ticks: %w", err)
    }
    defer rows.Close()

    var ticks []models.Tick
    for rows.Next() {
        var tick models.Tick
        err := rows.Scan(&tick.Time, &tick.Symbol, &tick.Price, &tick.Volume, &tick.Bid, &tick.Ask)
        if err != nil {
            return nil, fmt.Errorf("failed to scan tick row: %w", err)
        }
        ticks = append(ticks, tick)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating ticks: %w", err)
    }

    return ticks, nil
}

// Technical Indicators operations
func (d *Database) InsertTechnicalIndicator(indicator *models.TechnicalIndicator) error {
    query := `
        INSERT INTO analytics.technical_indicators (time, symbol, timeframe, indicator_name, value, metadata)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (time, symbol, timeframe, indicator_name) DO UPDATE SET
            value = EXCLUDED.value,
            metadata = EXCLUDED.metadata
    `
    
    _, err := d.db.Exec(query, indicator.Time, indicator.Symbol, indicator.Timeframe,
        indicator.IndicatorName, indicator.Value, indicator.Metadata)
    
    if err != nil {
        return fmt.Errorf("failed to insert technical indicator: %w", err)
    }

    return nil
}

func (d *Database) GetTechnicalIndicators(symbol, timeframe, indicatorName string, start, end time.Time, limit int) ([]models.TechnicalIndicator, error) {
    query := `
        SELECT time, symbol, timeframe, indicator_name, value, metadata
        FROM analytics.technical_indicators
        WHERE symbol = $1 AND timeframe = $2 AND indicator_name = $3 
        AND time >= $4 AND time <= $5
        ORDER BY time DESC
        LIMIT $6
    `
    
    rows, err := d.db.Query(query, symbol, timeframe, indicatorName, start, end, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to query technical indicators: %w", err)
    }
    defer rows.Close()

    var indicators []models.TechnicalIndicator
    for rows.Next() {
        var indicator models.TechnicalIndicator
        err := rows.Scan(
            &indicator.Time, &indicator.Symbol, &indicator.Timeframe,
            &indicator.IndicatorName, &indicator.Value, &indicator.Metadata,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan technical indicator row: %w", err)
        }
        indicators = append(indicators, indicator)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating technical indicators: %w", err)
    }

    return indicators, nil
}

// Health check
func (d *Database) HealthCheck() error {
    return d.db.Ping()
}

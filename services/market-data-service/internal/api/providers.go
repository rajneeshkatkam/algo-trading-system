package api

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/algo-trading/market-data-service/internal/models"
)

// MarketDataProvider defines the interface that all market data providers must implement
type MarketDataProvider interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    GetQuote(ctx context.Context, symbol string) (*models.Tick, error)
    GetOHLCV(ctx context.Context, symbol, timeframe string, from, to time.Time) ([]models.OHLCV, error)
    SubscribeToTicks(ctx context.Context, symbols []string, callback func(*models.Tick)) error
    UnsubscribeFromTicks(ctx context.Context, symbols []string) error
    IsConnected() bool
    GetName() string
}

// APIManager manages multiple market data providers
type APIManager struct {
    providers map[string]MarketDataProvider
    active    MarketDataProvider
}

func NewAPIManager() *APIManager {
    return &APIManager{
        providers: make(map[string]MarketDataProvider),
    }
}

func (am *APIManager) RegisterProvider(name string, provider MarketDataProvider) {
    am.providers[name] = provider
    log.Printf("Registered market data provider: %s", name)
}

func (am *APIManager) SetActiveProvider(name string) error {
    provider, exists := am.providers[name]
    if !exists {
        return fmt.Errorf("provider %s not found", name)
    }
    
    am.active = provider
    log.Printf("Active provider set to: %s", name)
    return nil
}

func (am *APIManager) GetActiveProvider() MarketDataProvider {
    return am.active
}

func (am *APIManager) ConnectAll(ctx context.Context) error {
    for name, provider := range am.providers {
        if err := provider.Connect(ctx); err != nil {
            log.Printf("Failed to connect to provider %s: %v", name, err)
        } else {
            log.Printf("Connected to provider: %s", name)
        }
    }
    return nil
}

func (am *APIManager) DisconnectAll(ctx context.Context) error {
    for name, provider := range am.providers {
        if err := provider.Disconnect(ctx); err != nil {
            log.Printf("Failed to disconnect from provider %s: %v", name, err)
        } else {
            log.Printf("Disconnected from provider: %s", name)
        }
    }
    return nil
}

// GetQuote gets quote from active provider with fallback
func (am *APIManager) GetQuote(ctx context.Context, symbol string) (*models.Tick, error) {
    if am.active != nil && am.active.IsConnected() {
        return am.active.GetQuote(ctx, symbol)
    }
    
    // Try other connected providers as fallback
    for name, provider := range am.providers {
        if provider.IsConnected() {
            log.Printf("Using fallback provider %s for quote", name)
            return provider.GetQuote(ctx, symbol)
        }
    }
    
    return nil, fmt.Errorf("no connected providers available")
}

// GetOHLCV gets OHLCV data from active provider with fallback
func (am *APIManager) GetOHLCV(ctx context.Context, symbol, timeframe string, from, to time.Time) ([]models.OHLCV, error) {
    if am.active != nil && am.active.IsConnected() {
        return am.active.GetOHLCV(ctx, symbol, timeframe, from, to)
    }
    
    // Try other connected providers as fallback
    for name, provider := range am.providers {
        if provider.IsConnected() {
            log.Printf("Using fallback provider %s for OHLCV", name)
            return provider.GetOHLCV(ctx, symbol, timeframe, from, to)
        }
    }
    
    return nil, fmt.Errorf("no connected providers available")
}

// Mock Provider for testing and development
type MockProvider struct {
    name      string
    connected bool
}

func NewMockProvider(name string) *MockProvider {
    return &MockProvider{
        name:      name,
        connected: false,
    }
}

func (mp *MockProvider) Connect(ctx context.Context) error {
    mp.connected = true
    log.Printf("Mock provider %s connected", mp.name)
    return nil
}

func (mp *MockProvider) Disconnect(ctx context.Context) error {
    mp.connected = false
    log.Printf("Mock provider %s disconnected", mp.name)
    return nil
}

func (mp *MockProvider) GetQuote(ctx context.Context, symbol string) (*models.Tick, error) {
    if !mp.connected {
        return nil, fmt.Errorf("provider not connected")
    }
    
    // Generate mock data
    tick := &models.Tick{
        Time:   time.Now(),
        Symbol: symbol,
        Price:  1000.00 + float64(time.Now().Unix()%100), // Mock price variation
        Volume: 1000,
        Bid:    func() *float64 { v := 999.50; return &v }(),
        Ask:    func() *float64 { v := 1000.50; return &v }(),
    }
    
    return tick, nil
}

func (mp *MockProvider) GetOHLCV(ctx context.Context, symbol, timeframe string, from, to time.Time) ([]models.OHLCV, error) {
    if !mp.connected {
        return nil, fmt.Errorf("provider not connected")
    }
    
    // Generate mock OHLCV data
    var ohlcvs []models.OHLCV
    current := from
    
    for current.Before(to) {
        ohlcv := models.OHLCV{
            Time:      current,
            Symbol:    symbol,
            Open:      1000.00,
            High:      1010.00,
            Low:       990.00,
            Close:     1005.00,
            Volume:    50000,
            Timeframe: timeframe,
        }
        ohlcvs = append(ohlcvs, ohlcv)
        
        // Increment time based on timeframe
        switch timeframe {
        case "1m":
            current = current.Add(time.Minute)
        case "5m":
            current = current.Add(5 * time.Minute)
        case "15m":
            current = current.Add(15 * time.Minute)
        case "1h":
            current = current.Add(time.Hour)
        case "1d":
            current = current.Add(24 * time.Hour)
        default:
            current = current.Add(time.Minute)
        }
    }
    
    return ohlcvs, nil
}

func (mp *MockProvider) SubscribeToTicks(ctx context.Context, symbols []string, callback func(*models.Tick)) error {
    if !mp.connected {
        return fmt.Errorf("provider not connected")
    }
    
    // Start mock tick generation
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                for _, symbol := range symbols {
                    tick := &models.Tick{
                        Time:   time.Now(),
                        Symbol: symbol,
                        Price:  1000.00 + float64(time.Now().Unix()%100),
                        Volume: 100,
                    }
                    callback(tick)
                }
            }
        }
    }()
    
    return nil
}

func (mp *MockProvider) UnsubscribeFromTicks(ctx context.Context, symbols []string) error {
    log.Printf("Mock provider %s unsubscribed from ticks: %v", mp.name, symbols)
    return nil
}

func (mp *MockProvider) IsConnected() bool {
    return mp.connected
}

func (mp *MockProvider) GetName() string {
    return mp.name
}

// Angel One API Provider (placeholder for actual implementation)
type AngelOneProvider struct {
    apiKey    string
    apiSecret string
    connected bool
}

func NewAngelOneProvider(apiKey, apiSecret string) *AngelOneProvider {
    return &AngelOneProvider{
        apiKey:    apiKey,
        apiSecret: apiSecret,
        connected: false,
    }
}

func (aop *AngelOneProvider) Connect(ctx context.Context) error {
    // TODO: Implement actual Angel One API connection
    aop.connected = true
    log.Println("Angel One provider connected (mock)")
    return nil
}

func (aop *AngelOneProvider) Disconnect(ctx context.Context) error {
    aop.connected = false
    log.Println("Angel One provider disconnected")
    return nil
}

func (aop *AngelOneProvider) GetQuote(ctx context.Context, symbol string) (*models.Tick, error) {
    if !aop.connected {
        return nil, fmt.Errorf("Angel One provider not connected")
    }
    
    // TODO: Implement actual Angel One API call
    // For now, return mock data
    return &models.Tick{
        Time:   time.Now(),
        Symbol: symbol,
        Price:  1000.00,
        Volume: 1000,
    }, nil
}

func (aop *AngelOneProvider) GetOHLCV(ctx context.Context, symbol, timeframe string, from, to time.Time) ([]models.OHLCV, error) {
    if !aop.connected {
        return nil, fmt.Errorf("Angel One provider not connected")
    }
    
    // TODO: Implement actual Angel One API call
    return []models.OHLCV{}, nil
}

func (aop *AngelOneProvider) SubscribeToTicks(ctx context.Context, symbols []string, callback func(*models.Tick)) error {
    if !aop.connected {
        return fmt.Errorf("Angel One provider not connected")
    }
    
    // TODO: Implement actual Angel One WebSocket subscription
    return nil
}

func (aop *AngelOneProvider) UnsubscribeFromTicks(ctx context.Context, symbols []string) error {
    // TODO: Implement actual Angel One WebSocket unsubscription
    return nil
}

func (aop *AngelOneProvider) IsConnected() bool {
    return aop.connected
}

func (aop *AngelOneProvider) GetName() string {
    return "AngelOne"
}

package storage

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisClient struct {
    client *redis.Client
}

func NewRedisClient(host, port string) *RedisClient {
    rdb := redis.NewClient(&redis.Options{
        Addr:        fmt.Sprintf("%s:%s", host, port),
        Password:    "", // No password
        DB:          0,  // Default DB
        PoolSize:    10,
        MinIdleConns: 5,
    })

    return &RedisClient{client: rdb}
}

func (r *RedisClient) Close() error {
    return r.client.Close()
}

// Generic cache operations
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("failed to marshal value: %w", err)
    }

    return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return fmt.Errorf("key %s not found", key)
        }
        return fmt.Errorf("failed to get value: %w", err)
    }

    return json.Unmarshal([]byte(val), dest)
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
    return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
    return r.client.Exists(ctx, keys...).Result()
}

// Stock price caching
func (r *RedisClient) CacheCurrentPrice(ctx context.Context, symbol string, price float64) error {
    key := fmt.Sprintf("current_price:%s", symbol)
    return r.Set(ctx, key, price, 30*time.Second) // Cache for 30 seconds
}

func (r *RedisClient) GetCurrentPrice(ctx context.Context, symbol string) (float64, error) {
    key := fmt.Sprintf("current_price:%s", symbol)
    var price float64
    err := r.Get(ctx, key, &price)
    return price, err
}

// Market status caching
func (r *RedisClient) CacheMarketStatus(ctx context.Context, status string) error {
    key := "market_status"
    return r.Set(ctx, key, status, 1*time.Minute) // Cache for 1 minute
}

func (r *RedisClient) GetMarketStatus(ctx context.Context) (string, error) {
    key := "market_status"
    var status string
    err := r.Get(ctx, key, &status)
    return status, err
}

// Technical indicators caching
func (r *RedisClient) CacheTechnicalIndicator(ctx context.Context, symbol, timeframe, indicator string, value float64) error {
    key := fmt.Sprintf("indicator:%s:%s:%s", symbol, timeframe, indicator)
    return r.Set(ctx, key, value, 5*time.Minute) // Cache for 5 minutes
}

func (r *RedisClient) GetTechnicalIndicator(ctx context.Context, symbol, timeframe, indicator string) (float64, error) {
    key := fmt.Sprintf("indicator:%s:%s:%s", symbol, timeframe, indicator)
    var value float64
    err := r.Get(ctx, key, &value)
    return value, err
}

// Publish/Subscribe for real-time data
func (r *RedisClient) PublishTick(ctx context.Context, symbol string, data interface{}) error {
    channel := fmt.Sprintf("ticks:%s", symbol)
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal tick data: %w", err)
    }
    
    return r.client.Publish(ctx, channel, jsonData).Err()
}

func (r *RedisClient) SubscribeTicks(ctx context.Context, symbol string) *redis.PubSub {
    channel := fmt.Sprintf("ticks:%s", symbol)
    return r.client.Subscribe(ctx, channel)
}

func (r *RedisClient) PublishOHLCV(ctx context.Context, symbol string, data interface{}) error {
    channel := fmt.Sprintf("ohlcv:%s", symbol)
    jsonData, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("failed to marshal OHLCV data: %w", err)
    }
    
    return r.client.Publish(ctx, channel, jsonData).Err()
}

func (r *RedisClient) SubscribeOHLCV(ctx context.Context, symbol string) *redis.PubSub {
    channel := fmt.Sprintf("ohlcv:%s", symbol)
    return r.client.Subscribe(ctx, channel)
}

// Rate limiting
func (r *RedisClient) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    count, err := r.client.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }

    if count == 1 {
        r.client.Expire(ctx, key, window)
    }

    return count > int64(limit), nil
}

// Health check
func (r *RedisClient) HealthCheck(ctx context.Context) error {
    return r.client.Ping(ctx).Err()
}

// Session storage for WebSocket connections
func (r *RedisClient) StoreWebSocketSession(ctx context.Context, sessionID string, data map[string]interface{}) error {
    key := fmt.Sprintf("ws_session:%s", sessionID)
    return r.Set(ctx, key, data, 1*time.Hour) // Session expires in 1 hour
}

func (r *RedisClient) GetWebSocketSession(ctx context.Context, sessionID string) (map[string]interface{}, error) {
    key := fmt.Sprintf("ws_session:%s", sessionID)
    var data map[string]interface{}
    err := r.Get(ctx, key, &data)
    return data, err
}

func (r *RedisClient) DeleteWebSocketSession(ctx context.Context, sessionID string) error {
    key := fmt.Sprintf("ws_session:%s", sessionID)
    return r.Del(ctx, key)
}

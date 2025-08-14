package main

import (
    "context"
    "log"
    "net"
    "net/http"
    "os"
    "os/signal"
    "sync"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"

    "github.com/algo-trading/market-data-service/internal/api"
    "github.com/algo-trading/market-data-service/internal/storage"
    "github.com/algo-trading/market-data-service/internal/websocket"
)

type Config struct {
    HTTPPort    string
    GRPCPort    string
    DBHost      string
    DBPort      string
    DBName      string
    DBUser      string
    DBPassword  string
    RedisHost   string
    RedisPort   string
    KafkaBrokers string
}

func loadConfig() *Config {
    return &Config{
        HTTPPort:     getEnv("HTTP_PORT", "8080"),
        GRPCPort:     getEnv("GRPC_PORT", "8081"),
        DBHost:       getEnv("DB_HOST", "localhost"),
        DBPort:       getEnv("DB_PORT", "5432"),
        DBName:       getEnv("DB_NAME", "algotrading"),
        DBUser:       getEnv("DB_USER", "postgres"),
        DBPassword:   getEnv("DB_PASSWORD", "password123"),
        RedisHost:    getEnv("REDIS_HOST", "localhost"),
        RedisPort:    getEnv("REDIS_PORT", "6379"),
        KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func main() {
    log.Println("Starting Market Data Service...")
    
    config := loadConfig()
    
    // Initialize storage
    db, err := storage.NewDatabase(config.DBHost, config.DBPort, config.DBName, config.DBUser, config.DBPassword)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Initialize Redis
    redisClient := storage.NewRedisClient(config.RedisHost, config.RedisPort)
    defer redisClient.Close()
    
    // Initialize API clients
    apiManager := api.NewAPIManager()
    
    // Initialize WebSocket hub
    wsHub := websocket.NewHub()
    go wsHub.Run()
    
    // Create service
    service := &MarketDataService{
        db:         db,
        redis:      redisClient,
        apiManager: apiManager,
        wsHub:      wsHub,
    }
    
    // Start servers
    var wg sync.WaitGroup
    
    // HTTP Server
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := startHTTPServer(service, config.HTTPPort); err != nil {
            log.Printf("HTTP server error: %v", err)
        }
    }()
    
    // gRPC Server
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := startGRPCServer(service, config.GRPCPort); err != nil {
            log.Printf("gRPC server error: %v", err)
        }
    }()
    
    // Start data collection
    wg.Add(1)
    go func() {
        defer wg.Done()
        service.StartDataCollection()
    }()
    
    // Wait for interrupt signal
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    log.Println("Market Data Service is running...")
    <-c
    
    log.Println("Shutting down gracefully...")
    // Graceful shutdown logic here
    
    wg.Wait()
    log.Println("Market Data Service stopped")
}

type MarketDataService struct {
    db         *storage.Database
    redis      *storage.RedisClient
    apiManager *api.APIManager
    wsHub      *websocket.Hub
}

func (s *MarketDataService) StartDataCollection() {
    // TODO: Implement data collection logic
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Collect and process market data
            // This will be implemented in the next iteration
        }
    }
}

func startHTTPServer(service *MarketDataService, port string) error {
    router := gin.Default()
    
    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "healthy"})
    })
    
    // API routes
    v1 := router.Group("/api/v1")
    {
        v1.GET("/stocks", service.getStocks)
        v1.GET("/stocks/:symbol/ohlcv", service.getOHLCV)
        v1.GET("/stocks/:symbol/ticks", service.getTicks)
    }
    
    // WebSocket endpoint
    router.GET("/ws", service.handleWebSocket)
    
    log.Printf("HTTP server starting on port %s", port)
    return http.ListenAndServe(":"+port, router)
}

func startGRPCServer(service *MarketDataService, port string) error {
    lis, err := net.Listen("tcp", ":"+port)
    if err != nil {
        return err
    }
    
    s := grpc.NewServer()
    // TODO: Register gRPC services
    
    log.Printf("gRPC server starting on port %s", port)
    return s.Serve(lis)
}

func (s *MarketDataService) getStocks(c *gin.Context) {
    // TODO: Implement get stocks endpoint
    c.JSON(http.StatusOK, gin.H{"message": "Get stocks endpoint"})
}

func (s *MarketDataService) getOHLCV(c *gin.Context) {
    symbol := c.Param("symbol")
    // TODO: Implement get OHLCV endpoint
    c.JSON(http.StatusOK, gin.H{"symbol": symbol, "message": "Get OHLCV endpoint"})
}

func (s *MarketDataService) getTicks(c *gin.Context) {
    symbol := c.Param("symbol")
    // TODO: Implement get ticks endpoint
    c.JSON(http.StatusOK, gin.H{"symbol": symbol, "message": "Get ticks endpoint"})
}

func (s *MarketDataService) handleWebSocket(c *gin.Context) {
    websocket.HandleWebSocket(s.wsHub, c.Writer, c.Request)
}

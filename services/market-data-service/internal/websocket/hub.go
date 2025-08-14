package websocket

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    "github.com/algo-trading/market-data-service/internal/models"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow connections from any origin
    },
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
    // Registered clients
    clients map[*Client]bool

    // Inbound messages from the clients
    broadcast chan []byte

    // Register requests from the clients
    register chan *Client

    // Unregister requests from clients
    unregister chan *Client

    // Symbol subscriptions
    subscriptions map[string]map[*Client]bool

    mu sync.RWMutex
}

func NewHub() *Hub {
    return &Hub{
        broadcast:     make(chan []byte),
        register:      make(chan *Client),
        unregister:    make(chan *Client),
        clients:       make(map[*Client]bool),
        subscriptions: make(map[string]map[*Client]bool),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            log.Printf("Client connected: %s", client.id)

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
                
                // Remove from all subscriptions
                for symbol, clients := range h.subscriptions {
                    if _, exists := clients[client]; exists {
                        delete(clients, client)
                        if len(clients) == 0 {
                            delete(h.subscriptions, symbol)
                        }
                    }
                }
                
                log.Printf("Client disconnected: %s", client.id)
            }
            h.mu.Unlock()

        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *Hub) Subscribe(client *Client, symbol string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if h.subscriptions[symbol] == nil {
        h.subscriptions[symbol] = make(map[*Client]bool)
    }
    
    h.subscriptions[symbol][client] = true
    log.Printf("Client %s subscribed to %s", client.id, symbol)
}

func (h *Hub) Unsubscribe(client *Client, symbol string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    if clients, exists := h.subscriptions[symbol]; exists {
        delete(clients, client)
        if len(clients) == 0 {
            delete(h.subscriptions, symbol)
        }
        log.Printf("Client %s unsubscribed from %s", client.id, symbol)
    }
}

func (h *Hub) BroadcastToSymbol(symbol string, message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()
    
    if clients, exists := h.subscriptions[symbol]; exists {
        for client := range clients {
            select {
            case client.send <- message:
            default:
                close(client.send)
                delete(h.clients, client)
                delete(clients, client)
            }
        }
    }
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
    id   string
}

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512
)

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()

    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, message, err := c.conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                log.Printf("WebSocket error: %v", err)
            }
            break
        }

        c.handleMessage(message)
    }
}

func (c *Client) writePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            // Add queued chat messages to the current websocket message
            n := len(c.send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.send)
            }

            if err := w.Close(); err != nil {
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

func (c *Client) handleMessage(message []byte) {
    var msg struct {
        Type   string `json:"type"`
        Symbol string `json:"symbol,omitempty"`
    }

    if err := json.Unmarshal(message, &msg); err != nil {
        log.Printf("Error unmarshaling message: %v", err)
        return
    }

    switch msg.Type {
    case "subscribe":
        if msg.Symbol != "" {
            c.hub.Subscribe(c, msg.Symbol)
            c.sendMessage(models.WebSocketMessage{
                Type:      "subscribed",
                Symbol:    msg.Symbol,
                Timestamp: time.Now(),
            })
        }

    case "unsubscribe":
        if msg.Symbol != "" {
            c.hub.Unsubscribe(c, msg.Symbol)
            c.sendMessage(models.WebSocketMessage{
                Type:      "unsubscribed",
                Symbol:    msg.Symbol,
                Timestamp: time.Now(),
            })
        }

    case "ping":
        c.sendMessage(models.WebSocketMessage{
            Type:      "pong",
            Timestamp: time.Now(),
        })

    default:
        log.Printf("Unknown message type: %s", msg.Type)
    }
}

func (c *Client) sendMessage(msg models.WebSocketMessage) {
    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling message: %v", err)
        return
    }

    select {
    case c.send <- data:
    default:
        close(c.send)
        delete(c.hub.clients, c)
    }
}

func HandleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade error: %v", err)
        return
    }

    clientID := r.Header.Get("X-Client-Id")
    if clientID == "" {
        clientID = generateClientID()
    }

    client := &Client{
        hub:  hub,
        conn: conn,
        send: make(chan []byte, 256),
        id:   clientID,
    }

    client.hub.register <- client

    // Allow collection of memory referenced by the caller by doing all work in
    // new goroutines
    go client.writePump()
    go client.readPump()
}

func generateClientID() string {
    return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}

// Helper methods for the Hub to send different types of messages
func (h *Hub) SendTick(symbol string, tick *models.Tick) {
    msg := models.WebSocketMessage{
        Type:      "tick",
        Symbol:    symbol,
        Data:      tick,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling tick message: %v", err)
        return
    }

    h.BroadcastToSymbol(symbol, data)
}

func (h *Hub) SendOHLCV(symbol string, ohlcv *models.OHLCV) {
    msg := models.WebSocketMessage{
        Type:      "ohlcv",
        Symbol:    symbol,
        Data:      ohlcv,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling OHLCV message: %v", err)
        return
    }

    h.BroadcastToSymbol(symbol, data)
}

func (h *Hub) SendTechnicalIndicator(symbol string, indicator *models.TechnicalIndicator) {
    msg := models.WebSocketMessage{
        Type:      "indicator",
        Symbol:    symbol,
        Data:      indicator,
        Timestamp: time.Now(),
    }

    data, err := json.Marshal(msg)
    if err != nil {
        log.Printf("Error marshaling indicator message: %v", err)
        return
    }

    h.BroadcastToSymbol(symbol, data)
}

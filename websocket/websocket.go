// Package websocket provides WebSocket support for Ruum.
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/goruum/ruum/core"
)

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	ID        string      `json:"id,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID    string
	Conn  *websocket.Conn
	Hub   *Hub
	Send  chan Message
	Rooms map[string]bool
	Data  map[string]interface{}
	mu    sync.RWMutex
}

// Hub manages WebSocket clients
type Hub struct {
	clients    map[string]*Client
	rooms      map[string]map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
	handlers   map[string]MessageHandler
	middleware []MiddlewareFunc
}

// MessageHandler handles WebSocket messages
type MessageHandler func(*Client, Message) error

// MiddlewareFunc is a WebSocket middleware
type MiddlewareFunc func(MessageHandler) MessageHandler

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message, 256),
		handlers:   make(map[string]MessageHandler),
		middleware: make([]MiddlewareFunc, 0),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)

				// Remove from all rooms
				for room := range client.Rooms {
					h.leaveRoom(client, room)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// On registers a message handler
func (h *Hub) On(event string, handler MessageHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Apply middleware
	finalHandler := handler
	for i := len(h.middleware) - 1; i >= 0; i-- {
		finalHandler = h.middleware[i](finalHandler)
	}

	h.handlers[event] = finalHandler
}

// Use adds middleware
func (h *Hub) Use(middleware MiddlewareFunc) {
	h.middleware = append(h.middleware, middleware)
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(event string, data interface{}) {
	message := Message{
		Type:      "event",
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.broadcast <- message
}

// BroadcastToRoom sends a message to all clients in a room
func (h *Hub) BroadcastToRoom(room, event string, data interface{}) {
	message := Message{
		Type:      "event",
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		for _, client := range clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
			}
		}
	}
}

// GetClients returns all connected clients
func (h *Hub) GetClients() []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*Client, 0, len(h.clients))
	for _, client := range h.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetRoomClients returns all clients in a room
func (h *Hub) GetRoomClients(room string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		result := make([]*Client, 0, len(clients))
		for _, client := range clients {
			result = append(result, client)
		}
		return result
	}
	return []*Client{}
}

func (h *Hub) joinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[room] == nil {
		h.rooms[room] = make(map[string]*Client)
	}

	h.rooms[room][client.ID] = client
	client.Rooms[room] = true
}

func (h *Hub) leaveRoom(client *Client, room string) {
	if clients, ok := h.rooms[room]; ok {
		delete(clients, client.ID)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(client.Rooms, room)
}

// Client methods

// Send sends a message to the client
func (c *Client) SendMessage(event string, data interface{}) error {
	message := Message{
		Type:      "event",
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case c.Send <- message:
		return nil
	default:
		return fmt.Errorf("client send channel is full")
	}
}

// SendError sends an error message to the client
func (c *Client) SendError(errorMsg string) error {
	message := Message{
		Type:      "error",
		Error:     errorMsg,
		Timestamp: time.Now(),
	}

	select {
	case c.Send <- message:
		return nil
	default:
		return fmt.Errorf("client send channel is full")
	}
}

// JoinRoom adds the client to a room
func (c *Client) JoinRoom(room string) {
	c.Hub.joinRoom(c, room)
}

// LeaveRoom removes the client from a room
func (c *Client) LeaveRoom(room string) {
	c.Hub.leaveRoom(c, room)
}

// SetData stores data in the client context
func (c *Client) SetData(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data[key] = value
}

// GetData retrieves data from the client context
func (c *Client) GetData(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.Data[key]
	return val, ok
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			// Connection closed or error occurred
			break
		}

		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			_ = c.SendError("Invalid message format")
			continue
		}

		// Find and execute handler
		c.Hub.mu.RLock()
		handler, ok := c.Hub.handlers[message.Event]
		c.Hub.mu.RUnlock()

		if ok {
			if err := handler(c, message); err != nil {
				_ = c.SendError(err.Error())
			}
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_ = json.NewEncoder(w).Encode(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins by default
	},
}

// HandlerConfig configures the WebSocket handler
type HandlerConfig struct {
	Hub          *Hub
	CheckOrigin  func(r *http.Request) bool
	OnConnect    func(*Client) error
	OnDisconnect func(*Client)
	GenerateID   func() string
}

// Handler creates a WebSocket handler
func Handler(config HandlerConfig) core.HandlerFunc {
	if config.CheckOrigin != nil {
		upgrader.CheckOrigin = config.CheckOrigin
	}

	if config.GenerateID == nil {
		config.GenerateID = func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		}
	}

	return func(ctx core.Context) error {
		conn, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
		if err != nil {
			return err
		}

		client := &Client{
			ID:    config.GenerateID(),
			Conn:  conn,
			Hub:   config.Hub,
			Send:  make(chan Message, 256),
			Rooms: make(map[string]bool),
			Data:  make(map[string]interface{}),
		}

		// Call OnConnect hook
		if config.OnConnect != nil {
			if err := config.OnConnect(client); err != nil {
				_ = conn.Close()
				return err
			}
		}

		config.Hub.register <- client

		// Start read and write pumps
		go client.WritePump()
		go client.ReadPump()

		return nil
	}
}

// Middleware builders

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(authFunc func(*Client, Message) error) MiddlewareFunc {
	return func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			if err := authFunc(client, msg); err != nil {
				return err
			}
			return next(client, msg)
		}
	}
}

// LoggingMiddleware creates a logging middleware
func LoggingMiddleware(logger func(string, map[string]interface{})) MiddlewareFunc {
	return func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			logger("WebSocket message received", map[string]interface{}{
				"clientId": client.ID,
				"event":    msg.Event,
			})
			return next(client, msg)
		}
	}
}

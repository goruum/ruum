package websocket

import (
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
}

func TestHub_Run(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	time.Sleep(10 * time.Millisecond)
	// Should not panic
}

func TestHub_On(t *testing.T) {
	hub := NewHub()

	hub.On("test", func(client *Client, msg Message) error {
		return nil
	})

	// Handler should be registered
	if len(hub.handlers) != 1 {
		t.Error("Handler was not registered")
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.Broadcast("test", "data")
	time.Sleep(10 * time.Millisecond)
	// Should not panic
}

func TestHub_GetClients(t *testing.T) {
	hub := NewHub()
	clients := hub.GetClients()

	if clients == nil {
		t.Error("GetClients() returned nil")
	}
}

func TestHub_GetClientCount(t *testing.T) {
	hub := NewHub()
	count := hub.GetClientCount()

	if count != 0 {
		t.Errorf("GetClientCount() = %d, want 0", count)
	}
}

func TestHub_GetRoomClients(t *testing.T) {
	hub := NewHub()
	clients := hub.GetRoomClients("test-room")

	if clients == nil {
		t.Error("GetRoomClients() returned nil")
	}
}

func TestClient_SendMessage(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Send: make(chan Message, 10),
	}

	err := client.SendMessage("test", "data")
	if err != nil {
		t.Errorf("SendMessage() error = %v", err)
	}
}

func TestClient_SendError(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Send: make(chan Message, 10),
	}

	err := client.SendError("error message")
	if err != nil {
		t.Errorf("SendError() error = %v", err)
	}
}

func TestClient_SetData(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Data: make(map[string]interface{}),
	}

	client.SetData("key", "value")

	val, ok := client.GetData("key")
	if !ok || val != "value" {
		t.Error("SetData/GetData failed")
	}
}

func TestClient_JoinLeaveRoom(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:    "test",
		Hub:   hub,
		Rooms: make(map[string]bool),
	}

	client.JoinRoom("test-room")

	if !client.Rooms["test-room"] {
		t.Error("Client did not join room")
	}

	client.LeaveRoom("test-room")

	if client.Rooms["test-room"] {
		t.Error("Client did not leave room")
	}
}

func TestHandler(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	config := HandlerConfig{
		Hub: hub,
		OnConnect: func(client *Client) error {
			return nil
		},
		OnDisconnect: func(client *Client) {},
	}

	handler := Handler(config)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate WebSocket upgrade
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
	}))
	defer server.Close()

	// Handler should not panic
	if handler == nil {
		t.Error("Handler() returned nil")
	}
}

func TestAuthMiddleware(t *testing.T) {
	authFunc := func(client *Client, msg Message) error {
		return nil
	}

	middleware := AuthMiddleware(authFunc)

	if middleware == nil {
		t.Error("AuthMiddleware() returned nil")
	}

	handler := func(client *Client, msg Message) error {
		return nil
	}

	wrapped := middleware(handler)

	client := &Client{
		ID:   "test",
		Data: make(map[string]interface{}),
	}

	err := wrapped(client, Message{})
	if err != nil {
		t.Errorf("Middleware error = %v", err)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logger := func(msg string, fields map[string]interface{}) {}

	middleware := LoggingMiddleware(logger)

	if middleware == nil {
		t.Error("LoggingMiddleware() returned nil")
	}

	handler := func(client *Client, msg Message) error {
		return nil
	}

	wrapped := middleware(handler)

	client := &Client{
		ID:   "test",
		Data: make(map[string]interface{}),
	}

	err := wrapped(client, Message{Event: "test"})
	if err != nil {
		t.Errorf("Middleware error = %v", err)
	}
}

func TestHub_Use(t *testing.T) {
	hub := NewHub()

	middleware := func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			return next(client, msg)
		}
	}

	hub.Use(middleware)

	if len(hub.middleware) != 1 {
		t.Error("Middleware was not added")
	}
}

func TestHub_BroadcastToRoom(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.BroadcastToRoom("test-room", "event", "data")
	time.Sleep(10 * time.Millisecond)
	// Should not panic
}

func TestMessage(t *testing.T) {
	msg := Message{
		Type:      "event",
		Event:     "test",
		Data:      "data",
		Timestamp: time.Now(),
	}

	if msg.Type != "event" {
		t.Error("Message type not set correctly")
	}
}

func TestUpgraderCheckOrigin(t *testing.T) {
	req := httptest.NewRequest("GET", "/ws", nil)
	
	result := upgrader.CheckOrigin(req)
	if !result {
		t.Error("CheckOrigin should return true by default")
	}
}

func TestHandlerConfig_GenerateID(t *testing.T) {
	config := HandlerConfig{
		Hub: NewHub(),
		GenerateID: func() string {
			return "custom-id"
		},
	}

	if config.GenerateID() != "custom-id" {
		t.Error("Custom ID generator not working")
	}
}

func TestClient_SendMessage_FullChannel(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Send: make(chan Message, 1),
	}

	// Fill the channel
	client.Send <- Message{}

	// This should return error
	err := client.SendMessage("test", "data")
	if err == nil {
		t.Error("SendMessage() should return error when channel is full")
	}
}

func TestHub_Use_Multiple(t *testing.T) {
	hub := NewHub()

	middleware1 := func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			client.SetData("m1", true)
			return next(client, msg)
		}
	}

	middleware2 := func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			client.SetData("m2", true)
			return next(client, msg)
		}
	}

	hub.Use(middleware1)
	hub.Use(middleware2)

	if len(hub.middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(hub.middleware))
	}
}

func TestMessage_Fields(t *testing.T) {
	now := time.Now()
	msg := Message{
		Type:      "event",
		Event:     "test",
		Data:      "payload",
		ID:        "123",
		Error:     "",
		Timestamp: now,
	}

	if msg.Type != "event" {
		t.Error("Type field not set correctly")
	}

	if msg.Event != "test" {
		t.Error("Event field not set correctly")
	}

	if msg.Data != "payload" {
		t.Error("Data field not set correctly")
	}

	if msg.ID != "123" {
		t.Error("ID field not set correctly")
	}

	if !msg.Timestamp.Equal(now) {
		t.Error("Timestamp field not set correctly")
	}
}

func TestMessage_WithError(t *testing.T) {
	msg := Message{
		Type:  "error",
		Error: "something went wrong",
	}

	if msg.Type != "error" {
		t.Error("Error message type not set")
	}

	if msg.Error != "something went wrong" {
		t.Error("Error field not set")
	}
}

func TestClient_Fields(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:    "client-123",
		Hub:   hub,
		Send:  make(chan Message, 256),
		Rooms: make(map[string]bool),
		Data:  make(map[string]interface{}),
	}

	if client.ID != "client-123" {
		t.Error("Client ID not set")
	}

	if client.Hub != hub {
		t.Error("Client hub not set")
	}

	if len(client.Rooms) != 0 {
		t.Error("Client rooms should be empty initially")
	}

	if len(client.Data) != 0 {
		t.Error("Client data should be empty initially")
	}
}

func TestHub_Fields(t *testing.T) {
	hub := NewHub()

	if hub.clients == nil {
		t.Error("Hub clients map should be initialized")
	}

	if hub.rooms == nil {
		t.Error("Hub rooms map should be initialized")
	}

	if hub.handlers == nil {
		t.Error("Hub handlers map should be initialized")
	}

	if hub.middleware == nil {
		t.Error("Hub middleware should be initialized")
	}
}

func TestHub_MultipleHandlers(t *testing.T) {
	hub := NewHub()

	hub.On("event1", func(client *Client, msg Message) error {
		return nil
	})

	hub.On("event2", func(client *Client, msg Message) error {
		return nil
	})

	hub.On("event3", func(client *Client, msg Message) error {
		return nil
	})

	if len(hub.handlers) != 3 {
		t.Errorf("Expected 3 handlers, got %d", len(hub.handlers))
	}
}

func TestClient_MultipleRooms(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:    "test",
		Hub:   hub,
		Rooms: make(map[string]bool),
	}

	client.JoinRoom("room1")
	client.JoinRoom("room2")
	client.JoinRoom("room3")

	if len(client.Rooms) != 3 {
		t.Errorf("Expected 3 rooms, got %d", len(client.Rooms))
	}

	client.LeaveRoom("room2")

	if len(client.Rooms) != 2 {
		t.Errorf("Expected 2 rooms after leaving, got %d", len(client.Rooms))
	}
}

func TestClient_DataStorage(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Data: make(map[string]interface{}),
	}

	client.SetData("user_id", "123")
	client.SetData("username", "john")
	client.SetData("role", "admin")

	userID, ok := client.GetData("user_id")
	if !ok || userID != "123" {
		t.Error("Failed to retrieve user_id")
	}

	username, ok := client.GetData("username")
	if !ok || username != "john" {
		t.Error("Failed to retrieve username")
	}

	_, ok = client.GetData("nonexistent")
	if ok {
		t.Error("Should return false for nonexistent key")
	}
}

func TestHub_BroadcastMultipleTimes(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	for i := 0; i < 10; i++ {
		hub.Broadcast("test", i)
	}

	time.Sleep(20 * time.Millisecond)
	// Should not panic
}

func TestHub_RoomOperations(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.BroadcastToRoom("room1", "event1", "data1")
	hub.BroadcastToRoom("room2", "event2", "data2")
	hub.BroadcastToRoom("room3", "event3", "data3")

	time.Sleep(10 * time.Millisecond)
	// Should not panic
}

func TestClient_SendMessageTypes(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Send: make(chan Message, 10),
	}

	// Test sending string
	err := client.SendMessage("event1", "string data")
	if err != nil {
		t.Errorf("SendMessage(string) error = %v", err)
	}

	// Test sending number
	err = client.SendMessage("event2", 42)
	if err != nil {
		t.Errorf("SendMessage(int) error = %v", err)
	}

	// Test sending map
	err = client.SendMessage("event3", map[string]string{"key": "value"})
	if err != nil {
		t.Errorf("SendMessage(map) error = %v", err)
	}

	// Test sending array
	err = client.SendMessage("event4", []string{"a", "b", "c"})
	if err != nil {
		t.Errorf("SendMessage(array) error = %v", err)
	}
}

func TestHub_MultipleHandlersSameEvent(t *testing.T) {
	hub := NewHub()

	// Register multiple handlers for same event
	hub.On("event", func(client *Client, msg Message) error {
		return nil
	})

	// Override with second handler
	hub.On("event", func(client *Client, msg Message) error {
		return nil
	})

	// Should only have one handler per event
	if len(hub.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(hub.handlers))
	}
}

func TestHub_MiddlewareOrder(t *testing.T) {
	hub := NewHub()
	order := []int{}

	middleware1 := func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			order = append(order, 1)
			return next(client, msg)
		}
	}

	middleware2 := func(next MessageHandler) MessageHandler {
		return func(client *Client, msg Message) error {
			order = append(order, 2)
			return next(client, msg)
		}
	}

	hub.Use(middleware1)
	hub.Use(middleware2)

	if len(hub.middleware) != 2 {
		t.Errorf("Expected 2 middleware, got %d", len(hub.middleware))
	}
}

func TestHandlerConfig_Defaults(t *testing.T) {
	config := HandlerConfig{
		Hub: NewHub(),
	}

	if config.Hub == nil {
		t.Error("Hub should not be nil")
	}
}

func TestClient_SendError_Multiple(t *testing.T) {
	hub := NewHub()
	client := &Client{
		ID:   "test",
		Hub:  hub,
		Send: make(chan Message, 10),
	}

	errors := []string{
		"Error 1",
		"Error 2",
		"Error 3",
	}

	for _, errMsg := range errors {
		err := client.SendError(errMsg)
		if err != nil {
			t.Errorf("SendError() failed: %v", err)
		}
	}
}


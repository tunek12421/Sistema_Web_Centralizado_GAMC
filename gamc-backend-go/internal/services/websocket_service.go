// internal/services/websocket_service.go
package services

import (
	"net/http"
	"sync"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/pkg/logger"

	"github.com/gorilla/websocket"
)

// WebSocketService manages WebSocket connections
type WebSocketService struct {
	config      *config.Config
	upgrader    websocket.Upgrader
	connections map[string]*websocket.Conn
	mutex       sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(cfg *config.Config) *WebSocketService {
	return &WebSocketService{
		config: cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Configure as needed
			},
		},
		connections: make(map[string]*websocket.Conn),
	}
}

// AddConnection adds a new WebSocket connection
func (ws *WebSocketService) AddConnection(userID string, conn *websocket.Conn) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.connections[userID] = conn
}

// RemoveConnection removes a WebSocket connection
func (ws *WebSocketService) RemoveConnection(userID string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if conn, exists := ws.connections[userID]; exists {
		conn.Close()
		delete(ws.connections, userID)
	}
}

// SendMessage sends a message to a specific user
func (ws *WebSocketService) SendMessage(userID string, message interface{}) error {
	ws.mutex.RLock()
	conn, exists := ws.connections[userID]
	ws.mutex.RUnlock()

	if !exists {
		return nil // User not connected
	}

	return conn.WriteJSON(message)
}

// SendNotification sends a notification to a specific user
func (ws *WebSocketService) SendNotification(userID string, notification interface{}) error {
	return ws.SendMessage(userID, notification)
}

// BroadcastMessage sends a message to all connected users
func (ws *WebSocketService) BroadcastMessage(message interface{}) {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	for userID, conn := range ws.connections {
		if err := conn.WriteJSON(message); err != nil {
			logger.Error("Error sending message to user %s: %v", userID, err)
			// Remove broken connection
			go ws.RemoveConnection(userID)
		}
	}
}

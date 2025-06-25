// internal/config/websocket.go
package config

import (
	"fmt"
	"time"
)

// WebSocketConfig contiene la configuración específica de WebSocket
type WebSocketConfig struct {
	Port            string
	PingInterval    time.Duration
	PongTimeout     time.Duration
	MaxMessageSize  int64
	ReadBufferSize  int
	WriteBufferSize int
	AllowedOrigins  []string
}

// WebSocketEventType define los tipos de eventos de WebSocket
type WebSocketEventType string

const (
	// Eventos de conexión
	EventTypeConnect    WebSocketEventType = "connect"
	EventTypeDisconnect WebSocketEventType = "disconnect"
	EventTypePing       WebSocketEventType = "ping"
	EventTypePong       WebSocketEventType = "pong"

	// Eventos de mensajería
	EventTypeMessageNew       WebSocketEventType = "message:new"
	EventTypeMessageRead      WebSocketEventType = "message:read"
	EventTypeMessageUpdate    WebSocketEventType = "message:update"
	EventTypeMessageDelete    WebSocketEventType = "message:delete"
	EventTypeMessageTyping    WebSocketEventType = "message:typing"
	EventTypeMessageDelivered WebSocketEventType = "message:delivered"

	// Eventos de notificaciones
	EventTypeNotificationNew   WebSocketEventType = "notification:new"
	EventTypeNotificationRead  WebSocketEventType = "notification:read"
	EventTypeNotificationClear WebSocketEventType = "notification:clear"

	// Eventos de archivos
	EventTypeFileUploadStart    WebSocketEventType = "file:upload:start"
	EventTypeFileUploadProgress WebSocketEventType = "file:upload:progress"
	EventTypeFileUploadComplete WebSocketEventType = "file:upload:complete"
	EventTypeFileUploadError    WebSocketEventType = "file:upload:error"

	// Eventos de presencia
	EventTypeUserOnline       WebSocketEventType = "user:online"
	EventTypeUserOffline      WebSocketEventType = "user:offline"
	EventTypeUserStatusChange WebSocketEventType = "user:status:change"

	// Eventos del sistema
	EventTypeSystemMaintenance WebSocketEventType = "system:maintenance"
	EventTypeSystemBroadcast   WebSocketEventType = "system:broadcast"
	EventTypeSystemAlert       WebSocketEventType = "system:alert"

	// Eventos de sesión
	EventTypeSessionExpired WebSocketEventType = "session:expired"
	EventTypeSessionRevoked WebSocketEventType = "session:revoked"

	// Eventos de error
	EventTypeError           WebSocketEventType = "error"
	EventTypeUnauthorized    WebSocketEventType = "error:unauthorized"
	EventTypeRateLimitExceed WebSocketEventType = "error:rate_limit"
)

// WebSocketMessage estructura base para mensajes WebSocket
type WebSocketMessage struct {
	ID        string                 `json:"id"`
	Type      WebSocketEventType     `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"userId,omitempty"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// WebSocketRoom representa una sala/canal de WebSocket
type WebSocketRoom string

const (
	// Salas globales
	RoomGlobal WebSocketRoom = "global"
	RoomSystem WebSocketRoom = "system"

	// Prefijos para salas dinámicas
	RoomPrefixUser WebSocketRoom = "user:"
	RoomPrefixUnit WebSocketRoom = "unit:"
	RoomPrefixRole WebSocketRoom = "role:"
)

// GetUserRoom retorna la sala específica de un usuario
func GetUserRoom(userID string) string {
	return string(RoomPrefixUser) + userID
}

// GetUnitRoom retorna la sala específica de una unidad
func GetUnitRoom(unitID int) string {
	return string(RoomPrefixUnit) + fmt.Sprintf("%d", unitID)
}

// GetRoleRoom retorna la sala específica de un rol
func GetRoleRoom(role string) string {
	return string(RoomPrefixRole) + role
}

// WebSocketClientInfo información del cliente WebSocket
type WebSocketClientInfo struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	UnitID        int       `json:"unitId"`
	Role          string    `json:"role"`
	ConnectedAt   time.Time `json:"connectedAt"`
	LastPingAt    time.Time `json:"lastPingAt"`
	IPAddress     string    `json:"ipAddress"`
	UserAgent     string    `json:"userAgent"`
	Rooms         []string  `json:"rooms"`
	MessageCount  int64     `json:"messageCount"`
	BytesSent     int64     `json:"bytesSent"`
	BytesReceived int64     `json:"bytesReceived"`
}

// WebSocketStats estadísticas del servidor WebSocket
type WebSocketStats struct {
	TotalConnections   int            `json:"totalConnections"`
	ActiveConnections  int            `json:"activeConnections"`
	TotalMessages      int64          `json:"totalMessages"`
	TotalBytesSent     int64          `json:"totalBytesSent"`
	TotalBytesReceived int64          `json:"totalBytesReceived"`
	ConnectionsByRoom  map[string]int `json:"connectionsByRoom"`
	ConnectionsByUser  map[string]int `json:"connectionsByUser"`
	Uptime             time.Duration  `json:"uptime"`
	StartedAt          time.Time      `json:"startedAt"`
}

// WebSocketError estructura para errores de WebSocket
type WebSocketError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Códigos de error estándar
const (
	WSErrorInvalidMessage    = "INVALID_MESSAGE"
	WSErrorUnauthorized      = "UNAUTHORIZED"
	WSErrorRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	WSErrorInternalError     = "INTERNAL_ERROR"
	WSErrorInvalidRoom       = "INVALID_ROOM"
	WSErrorConnectionClosed  = "CONNECTION_CLOSED"
)

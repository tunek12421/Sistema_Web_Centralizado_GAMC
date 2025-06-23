package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gamc-backend-go/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// Initialize inicializa la conexión a Redis
func Initialize(redisURL string) (*redis.Client, error) {
	// Parsear URL de Redis
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Configuraciones adicionales
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.ConnMaxLifetime = time.Hour
	opts.PoolTimeout = 30 * time.Second
	opts.ConnMaxIdleTime = 5 * time.Minute

	// Crear cliente
	client := redis.NewClient(opts)

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return client, nil
}

// SessionData representa los datos de una sesión
type SessionData struct {
	UserID               string    `json:"userId"`
	Email                string    `json:"email"`
	Role                 string    `json:"role"`
	OrganizationalUnitID int       `json:"organizationalUnitId"`
	SessionID            string    `json:"sessionId"`
	CreatedAt            time.Time `json:"createdAt"`
	LastActivity         time.Time `json:"lastActivity"`
	IPAddress            string    `json:"ipAddress,omitempty"`
	UserAgent            string    `json:"userAgent,omitempty"`
}

// SessionManager maneja las operaciones de sesión
type SessionManager struct {
	client *redis.Client
}

// NewSessionManager crea un nuevo manejador de sesiones
func NewSessionManager(client *redis.Client) *SessionManager {
	return &SessionManager{client: client}
}

// SaveSession guarda una sesión en Redis (DB 0)
func (sm *SessionManager) SaveSession(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	return sm.client.SetEx(ctx, key, jsonData, ttl).Err()
}

// GetSession obtiene una sesión desde Redis
func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	result, err := sm.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Sesión no encontrada
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var data SessionData
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &data, nil
}

// DeleteSession elimina una sesión
func (sm *SessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sm.client.Del(ctx, key).Err()
}

// ExtendSession extiende el TTL de una sesión
func (sm *SessionManager) ExtendSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return sm.client.Expire(ctx, key, ttl).Err()
}

// GetUserSessions obtiene todas las sesiones de un usuario
func (sm *SessionManager) GetUserSessions(ctx context.Context, userID string) ([]string, error) {
	keys, err := sm.client.Keys(ctx, "session:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	var userSessions []string
	for _, key := range keys {
		sessionData, err := sm.GetSession(ctx, strings.TrimPrefix(key, "session:"))
		if err != nil {
			continue
		}
		if sessionData != nil && sessionData.UserID == userID {
			userSessions = append(userSessions, strings.TrimPrefix(key, "session:"))
		}
	}

	return userSessions, nil
}

// RefreshTokenManager maneja los refresh tokens
type RefreshTokenManager struct {
	client *redis.Client
}

// NewRefreshTokenManager crea un nuevo manejador de refresh tokens
func NewRefreshTokenManager(client *redis.Client) *RefreshTokenManager {
	return &RefreshTokenManager{client: client}
}

// SaveRefreshToken guarda un refresh token
func (rtm *RefreshTokenManager) SaveRefreshToken(ctx context.Context, userID, sessionID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, sessionID)
	return rtm.client.SetEx(ctx, key, token, ttl).Err()
}

// GetRefreshToken obtiene un refresh token
func (rtm *RefreshTokenManager) GetRefreshToken(ctx context.Context, userID, sessionID string) (string, error) {
	key := fmt.Sprintf("refresh:%s:%s", userID, sessionID)

	result, err := rtm.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	return result, nil
}

// DeleteRefreshToken elimina un refresh token
func (rtm *RefreshTokenManager) DeleteRefreshToken(ctx context.Context, userID, sessionID string) error {
	key := fmt.Sprintf("refresh:%s:%s", userID, sessionID)
	return rtm.client.Del(ctx, key).Err()
}

// DeleteAllUserRefreshTokens elimina todos los refresh tokens de un usuario
func (rtm *RefreshTokenManager) DeleteAllUserRefreshTokens(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("refresh:%s:*", userID)
	keys, err := rtm.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get refresh token keys: %w", err)
	}

	if len(keys) > 0 {
		return rtm.client.Del(ctx, keys...).Err()
	}

	return nil
}

// JWTBlacklistManager maneja la blacklist de JWT (DB 5)
type JWTBlacklistManager struct {
	client *redis.Client
}

// NewJWTBlacklistManager crea un nuevo manejador de blacklist
func NewJWTBlacklistManager(client *redis.Client) *JWTBlacklistManager {
	// Cambiar a DB 5 para blacklist
	opts := client.Options()
	opts.DB = 5

	blacklistClient := redis.NewClient(opts)

	return &JWTBlacklistManager{client: blacklistClient}
}

// BlacklistToken agrega un token a la blacklist
func (jbm *JWTBlacklistManager) BlacklistToken(ctx context.Context, jti string, exp int64) error {
	key := fmt.Sprintf("blacklist:%s", jti)

	// Calcular TTL basado en expiración del token
	now := time.Now().Unix()
	ttl := time.Duration(exp-now) * time.Second

	if ttl > 0 {
		return jbm.client.SetEx(ctx, key, "revoked", ttl).Err()
	}

	return nil // Token ya expirado
}

// IsTokenBlacklisted verifica si un token está en blacklist
func (jbm *JWTBlacklistManager) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("blacklist:%s", jti)

	exists, err := jbm.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}

	return exists == 1, nil
}

// CacheManager maneja el cache general (DB 1)
type CacheManager struct {
	client *redis.Client
}

// NewCacheManager crea un nuevo manejador de cache
func NewCacheManager(client *redis.Client) *CacheManager {
	// Cambiar a DB 1 para cache
	opts := client.Options()
	opts.DB = 1

	cacheClient := redis.NewClient(opts)

	return &CacheManager{client: cacheClient}
}

// Set guarda un valor en cache
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	return cm.client.SetEx(ctx, key, jsonData, ttl).Err()
}

// Get obtiene un valor desde cache
func (cm *CacheManager) Get(ctx context.Context, key string, dest interface{}) error {
	result, err := cm.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil // No encontrado
		}
		return fmt.Errorf("failed to get cache: %w", err)
	}

	return json.Unmarshal([]byte(result), dest)
}

// Delete elimina una clave del cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	return cm.client.Del(ctx, key).Err()
}

// RedisStats representa estadísticas de Redis
type RedisStats struct {
	Sessions          int64  `json:"sessions"`
	RefreshTokens     int64  `json:"refreshTokens"`
	BlacklistedTokens int64  `json:"blacklistedTokens"`
	MemoryUsed        string `json:"memoryUsed"`
	Connected         bool   `json:"connected"`
}

// GetStats obtiene estadísticas de Redis
func GetStats(ctx context.Context, client *redis.Client) (*RedisStats, error) {
	stats := &RedisStats{Connected: true}

	// Contar sesiones (DB 0)
	sessionKeys, err := client.Keys(ctx, "session:*").Result()
	if err != nil {
		logger.Error("Failed to get session keys: %v", err)
	} else {
		stats.Sessions = int64(len(sessionKeys))
	}

	// Contar refresh tokens (DB 0)
	refreshKeys, err := client.Keys(ctx, "refresh:*").Result()
	if err != nil {
		logger.Error("Failed to get refresh keys: %v", err)
	} else {
		stats.RefreshTokens = int64(len(refreshKeys))
	}

	// Contar tokens en blacklist (DB 5)
	blacklistClient := redis.NewClient(&redis.Options{
		Addr:     client.Options().Addr,
		Password: client.Options().Password,
		DB:       5,
	})
	defer blacklistClient.Close()

	blacklistKeys, err := blacklistClient.Keys(ctx, "blacklist:*").Result()
	if err != nil {
		logger.Error("Failed to get blacklist keys: %v", err)
	} else {
		stats.BlacklistedTokens = int64(len(blacklistKeys))
	}

	// Obtener uso de memoria
	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		logger.Error("Failed to get memory info: %v", err)
		stats.MemoryUsed = "N/A"
	} else {
		// Parsear info de memoria
		lines := strings.Split(info, "\r\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory_human:") {
				stats.MemoryUsed = strings.TrimSpace(strings.Split(line, ":")[1])
				break
			}
		}
	}

	return stats, nil
}

// internal/api/middleware/ratelimit.go
package middleware

import (
	"net/http"
	"sync"
	"time"

	"gamc-backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter estructura para rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

// NewRateLimiter crea un nuevo rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimitMiddleware middleware de rate limiting
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			response.Error(c, http.StatusTooManyRequests,
				"Demasiadas peticiones. Intente de nuevo más tarde.", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow verifica si una IP puede hacer una request
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Obtener requests de esta IP
	requests := rl.requests[clientIP]

	// Filtrar requests dentro de la ventana de tiempo
	var validRequests []time.Time
	for _, reqTime := range requests {
		if now.Sub(reqTime) < rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Verificar si excede el límite
	if len(validRequests) >= rl.limit {
		return false
	}

	// Agregar request actual
	validRequests = append(validRequests, now)
	rl.requests[clientIP] = validRequests

	return true
}

// UserRateLimitMiddleware rate limiting por usuario autenticado
func UserRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	userLimiter := NewRateLimiter(limit, window)

	return func(c *gin.Context) {
		// Obtener ID de usuario si está autenticado
		userID, exists := c.Get("userID")
		if !exists {
			// Si no está autenticado, usar IP
			clientIP := c.ClientIP()
			if !userLimiter.Allow(clientIP) {
				response.Error(c, http.StatusTooManyRequests,
					"Demasiadas peticiones. Intente de nuevo más tarde.", "")
				c.Abort()
				return
			}
		} else {
			// Rate limit por usuario
			if !userLimiter.Allow(userID.(string)) {
				response.Error(c, http.StatusTooManyRequests,
					"Demasiadas peticiones. Intente de nuevo más tarde.", "")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

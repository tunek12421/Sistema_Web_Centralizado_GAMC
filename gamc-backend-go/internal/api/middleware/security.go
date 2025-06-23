// internal/api/middleware/security.go
package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders middleware para agregar headers de seguridad
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevenir ataques XSS
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content Security Policy básico
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Prevenir información del servidor
		c.Header("Server", "GAMC-Auth-Service")

		// Strict Transport Security (solo para HTTPS)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// NoCache middleware para evitar caché en endpoints sensibles
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

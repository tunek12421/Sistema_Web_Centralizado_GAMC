package middleware

import (
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware middleware de logging personalizado
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Formato personalizado de logs
		return fmt.Sprintf("[%s] %s %s %d %s %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)
	})
}

// UserActivityLogger middleware para loggear actividad de usuarios
func UserActivityLogger(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Ejecutar el handler primero

		// Log después de la ejecución
		if user, exists := c.Get("user"); exists {
			if userProfile, ok := user.(*models.UserProfile); ok {
				logger.Info("📝 Actividad: %s - %s - %s %s - IP: %s", 
					userProfile.Email, 
					action, 
					c.Request.Method, 
					c.Request.URL.Path, 
					c.ClientIP(),
				)
			}
		}
	}
}

// RequestLogger middleware para logging detallado de requests
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Procesar request
		c.Next()

		// Calcular latencia
		latency := time.Since(startTime)

		// Obtener información del usuario si está disponible
		userEmail := "anonymous"
		if user, exists := c.Get("user"); exists {
			if userProfile, ok := user.(*models.UserProfile); ok {
				userEmail = userProfile.Email
			}
		}

		// Log estructurado
		logger.Info("HTTP Request: %s %s | Status: %d | Latency: %v | User: %s | IP: %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			userEmail,
			c.ClientIP(),
		)

		// Log de errores para status >= 400
		if c.Writer.Status() >= 400 {
			logger.Error("HTTP Error: %s %s | Status: %d | User: %s | IP: %s",
				c.Request.Method,
				c.Request.URL.Path,
				c.Writer.Status(),
				userEmail,
				c.ClientIP(),
			)
		}
	}
}

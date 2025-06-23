// internal/api/middleware/cors.go
package middleware

import (
	"gamc-backend-go/internal/config"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura CORS avanzado para el API del GAMC
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Para desarrollo, permitir todos los orígenes
	if cfg.Environment == "development" {
		config := cors.Config{
			AllowAllOrigins: true,
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
			AllowHeaders: []string{
				"Content-Type",
				"Authorization",
				"X-Requested-With",
				"Accept",
				"Origin",
				"Cache-Control",
				"X-CSRF-Token",
				"X-Forwarded-For",
				"X-Real-IP",
			},
			ExposeHeaders: []string{
				"X-Total-Count",
				"X-Page-Count",
				"X-Current-Page",
				"X-Per-Page",
			},
			AllowCredentials: false,        // Debe ser false cuando AllowAllOrigins es true
			MaxAge:           12 * 60 * 60, // 12 hours
		}
		return cors.New(config)
	}

	// Para producción, parsear múltiples orígenes desde CORS_ORIGIN
	var allowedOrigins []string
	if cfg.CORSOrigin != "" {
		// Dividir por comas y limpiar espacios
		origins := strings.Split(cfg.CORSOrigin, ",")
		for _, origin := range origins {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" && trimmed != "*" { // Evitar * en producción
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	// Si no hay orígenes específicos, usar configuración por defecto segura
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:5173", // Frontend desarrollo
			"http://localhost:3000", // Backend docs/admin
		}
	}

	config := cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Origin",
			"Cache-Control",
			"X-CSRF-Token",
			"X-Forwarded-For",
			"X-Real-IP",
		},
		ExposeHeaders: []string{
			"X-Total-Count",
			"X-Page-Count",
			"X-Current-Page",
			"X-Per-Page",
		},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}

	return cors.New(config)
}

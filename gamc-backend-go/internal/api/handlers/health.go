package handlers

import (
	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database"
	"gamc-backend-go/internal/redis"
	"gamc-backend-go/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler maneja las operaciones de health check
type HealthHandler struct {
	appCtx *config.AppContext
}

// NewHealthHandler crea una nueva instancia del handler de health
func NewHealthHandler(appCtx *config.AppContext) *HealthHandler {
	return &HealthHandler{appCtx: appCtx}
}

// HealthCheck maneja GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	// Obtener estadísticas básicas de la base de datos
	dbStats, err := database.GetBasicStats(h.appCtx.DB)
	if err != nil {
		dbStats = &database.DatabaseStats{Connected: false}
	}

	// Obtener estadísticas de Redis
	redisStats, err := redis.GetStats(c.Request.Context(), h.appCtx.Redis)
	if err != nil {
		redisStats = &redis.RedisStats{Connected: false}
	}

	// Determinar estado general
	status := "healthy"
	httpStatus := http.StatusOK
	if !dbStats.Connected || !redisStats.Connected {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	healthData := gin.H{
		"status":      status,
		"timestamp":   gin.H{"iso": response.GetTimestamp()},
		"environment": h.appCtx.Config.Environment,
		"version":     "1.0.0",
		"services": gin.H{
			"database": gin.H{
				"connected":      dbStats.Connected,
				"databaseName":   dbStats.DatabaseName,
				"tablesCount":    dbStats.TablesCount,
				"connectionTime": dbStats.ConnectionTime,
				"message":        "Usando base de datos existente sin migraciones",
			},
			"redis": gin.H{
				"connected":         redisStats.Connected,
				"sessions":          redisStats.Sessions,
				"refreshTokens":     redisStats.RefreshTokens,
				"blacklistedTokens": redisStats.BlacklistedTokens,
				"memoryUsed":        redisStats.MemoryUsed,
			},
		},
	}

	c.JSON(httpStatus, response.APIResponse{
		Success:   status == "healthy",
		Message:   "GAMC Auth Service - Health Check (Sin Migraciones Automáticas)",
		Data:      healthData,
		Timestamp: response.GetTimestamp(),
	})
}

// Info maneja GET /
func (h *HealthHandler) Info(c *gin.Context) {
	response.Success(c, "GAMC Sistema de Autenticación API", gin.H{
		"version":     "1.0.0",
		"environment": h.appCtx.Config.Environment,
		"timestamp":   response.GetTimestamp(),
		"mode":        "production-ready (sin migraciones automáticas)",
		"endpoints": gin.H{
			"health": "/health",
			"auth":   h.appCtx.Config.APIPrefix + "/auth",
		},
	})
}

// internal/api/routes/routes.go
package routes

import (
	"time"

	"gamc-backend-go/internal/api/handlers"
	"gamc-backend-go/internal/api/middleware"
	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/services"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configura todas las rutas de la aplicación
func SetupRoutes(appCtx *config.AppContext) *gin.Engine {
	// Crear router
	router := gin.New()

	// Middlewares globales
	router.Use(gin.Recovery())                           // Recuperación de panics
	router.Use(middleware.SecurityHeaders())             // Headers de seguridad
	router.Use(middleware.CORSMiddleware(appCtx.Config)) // CORS
	router.Use(middleware.RequestLogger())               // Logging personalizado

	// Rate limiting global
	router.Use(middleware.RateLimitMiddleware(
		appCtx.Config.RateLimitMaxRequests,
		appCtx.Config.RateLimitWindowMs,
	))

	// Crear handlers
	healthHandler := handlers.NewHealthHandler(appCtx)
	authHandler := handlers.NewAuthHandler(appCtx)

	// ========================================
	// RUTAS PÚBLICAS
	// ========================================

	// Información básica del servicio
	router.GET("/", healthHandler.Info)

	// Health check
	router.GET("/health", healthHandler.HealthCheck)

	// ========================================
	// RUTAS DE AUTENTICACIÓN (API v1)
	// ========================================

	apiV1 := router.Group(appCtx.Config.APIPrefix)
	{
		auth := apiV1.Group("/auth")
		{
			// Rate limiting específico para auth
			authRateLimit := middleware.UserRateLimitMiddleware(10, 15*time.Minute)

			// Rutas públicas de auth
			auth.POST("/login",
				authRateLimit,
				middleware.UserActivityLogger("LOGIN_ATTEMPT"),
				authHandler.Login)

			auth.POST("/register",
				authRateLimit,
				middleware.UserActivityLogger("REGISTER_ATTEMPT"),
				authHandler.Register)

			auth.POST("/refresh",
				authRateLimit,
				authHandler.RefreshToken)

			// ========================================
			// RUTAS PÚBLICAS DE RESET DE CONTRASEÑA
			// ========================================

			// Solicitar reset de contraseña (público)
			auth.POST("/forgot-password",
				authRateLimit,
				middleware.UserActivityLogger("PASSWORD_RESET_REQUEST"),
				authHandler.RequestPasswordReset)

			// Confirmar reset de contraseña (público)
			auth.POST("/reset-password",
				authRateLimit,
				middleware.NoCache(),
				middleware.UserActivityLogger("PASSWORD_RESET_CONFIRM"),
				authHandler.ConfirmPasswordReset)

			// Rutas protegidas de auth
			protected := auth.Group("/")
			protected.Use(middleware.AuthMiddleware(appCtx))
			{
				protected.POST("/logout",
					middleware.UserActivityLogger("LOGOUT"),
					authHandler.Logout)

				protected.GET("/profile",
					middleware.UserActivityLogger("GET_PROFILE"),
					authHandler.GetProfile)

				protected.PUT("/change-password",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("CHANGE_PASSWORD"),
					authHandler.ChangePassword)

				protected.GET("/verify",
					authHandler.VerifyToken)

				// ========================================
				// RUTAS PROTEGIDAS DE RESET (usuarios autenticados)
				// ========================================

				// Ver estado de reset del usuario actual
				protected.GET("/reset-status",
					middleware.UserActivityLogger("GET_RESET_STATUS"),
					authHandler.GetPasswordResetStatus)
			}

			// ========================================
			// RUTAS DE ADMINISTRACIÓN DE RESET (solo admins)
			// ========================================

			adminReset := auth.Group("/admin")
			adminReset.Use(middleware.AuthMiddleware(appCtx))
			adminReset.Use(middleware.RequireRole("admin"))
			{
				// Limpiar tokens expirados (solo admins)
				adminReset.POST("/cleanup-tokens",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("CLEANUP_RESET_TOKENS"),
					authHandler.CleanupExpiredTokens)
			}
		}

		// ========================================
		// RUTAS DE MENSAJERÍA (LIMPIO - SIN MIDDLEWARES RESTRICTIVOS)
		// ========================================

		// Crear servicios y handlers de mensajería
		messageService := services.NewMessageService(appCtx.DB)
		messageHandler := handlers.NewMessageHandler(messageService)

		messages := apiV1.Group("/messages")
		messages.Use(middleware.AuthMiddleware(appCtx))
		{
			// Obtener estadísticas de mensajes
			messages.GET("/stats", messageHandler.GetMessageStats)

			// Obtener tipos y estados de mensajes
			messages.GET("/types", messageHandler.GetMessageTypes)
			messages.GET("/statuses", messageHandler.GetMessageStatuses)

			// Listar mensajes - SIN middleware restrictivo
			messages.GET("/", messageHandler.GetMessages)

			// Crear nuevo mensaje - SIN middleware restrictivo
			messages.POST("/", messageHandler.CreateMessage)

			// Operaciones con mensajes específicos
			messageRoutes := messages.Group("/:id")
			{
				// Obtener mensaje por ID
				messageRoutes.GET("", messageHandler.GetMessageByID)

				// Marcar como leído
				messageRoutes.PUT("/read", messageHandler.MarkAsRead)

				// Actualizar estado - SIN middleware restrictivo
				messageRoutes.PUT("/status", messageHandler.UpdateMessageStatus)

				// Eliminar mensaje
				messageRoutes.DELETE("", messageHandler.DeleteMessage)
			}
		}

		// ========================================
		// RUTAS DE ARCHIVOS (futuras - Tarea 4.3)
		// ========================================

		files := apiV1.Group("/files")
		files.Use(middleware.AuthMiddleware(appCtx))
		{
			files.POST("/upload", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "File upload endpoint - Tarea 4.3 pendiente",
					"status":  "coming_soon",
				})
			})

			files.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "File download endpoint - Tarea 4.3 pendiente",
					"status":  "coming_soon",
				})
			})
		}

		// ========================================
		// RUTAS DE ADMINISTRACIÓN
		// ========================================

		admin := apiV1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(appCtx))
		admin.Use(middleware.RequireRole("admin"))
		{
			// Gestión de usuarios
			admin.GET("/users", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Admin users endpoint - Tarea 7.1 pendiente",
					"status":  "coming_soon",
				})
			})

			// Estadísticas del sistema
			admin.GET("/stats", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Admin stats endpoint - Tarea 6.x pendiente",
					"status":  "coming_soon",
				})
			})

			// Logs de auditoría
			admin.GET("/audit", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Admin audit endpoint - Tarea 7.3 pendiente",
					"status":  "coming_soon",
				})
			})
		}

		// ========================================
		// RUTAS DE NOTIFICACIONES (futuras - Tarea 4.4)
		// ========================================

		notifications := apiV1.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware(appCtx))
		{
			notifications.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "Notifications endpoint - Tarea 4.4 pendiente",
					"status":  "coming_soon",
				})
			})

			// WebSocket para notificaciones en tiempo real
			notifications.GET("/ws", func(c *gin.Context) {
				c.JSON(200, gin.H{
					"message": "WebSocket notifications - Tarea 4.4 pendiente",
					"status":  "coming_soon",
				})
			})
		}
	}

	// ========================================
	// CATCH-ALL PARA 404
	// ========================================

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success":   false,
			"message":   "Endpoint no encontrado",
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"timestamp": time.Now().Format(time.RFC3339),
			"available_endpoints": gin.H{
				"auth":          "/api/v1/auth/*",
				"messages":      "/api/v1/messages/*",
				"files":         "/api/v1/files/*",
				"admin":         "/api/v1/admin/*",
				"notifications": "/api/v1/notifications/*",
				"health":        "/health",
			},
		})
	})

	return router
}

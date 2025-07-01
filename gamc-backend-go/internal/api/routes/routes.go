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

	// IMPORTANTE: Deshabilitar redirecciones automáticas de trailing slash
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

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

			// ========================================
			// RUTAS PÚBLICAS DE AUTENTICACIÓN
			// ========================================

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
			// RUTAS PÚBLICAS DE PREGUNTAS DE SEGURIDAD
			// ========================================

			// Obtener catálogo de preguntas de seguridad (público para registro)
			auth.GET("/security-questions",
				authHandler.GetSecurityQuestions)

			// ========================================
			// RUTAS PÚBLICAS DE RESET DE CONTRASEÑA
			// ========================================

			// Solicitar reset de contraseña (público)
			auth.POST("/forgot-password",
				authRateLimit,
				middleware.UserActivityLogger("PASSWORD_RESET_REQUEST"),
				authHandler.RequestPasswordReset)

			// Obtener estado de proceso de reset por email (público)
			auth.GET("/reset-status",
				authRateLimit,
				middleware.UserActivityLogger("GET_RESET_STATUS"),
				authHandler.GetPasswordResetStatus)

			// Verificar pregunta de seguridad durante reset (público con token)
			auth.POST("/verify-security-question",
				authRateLimit,
				middleware.NoCache(),
				middleware.UserActivityLogger("VERIFY_SECURITY_QUESTION"),
				authHandler.VerifySecurityQuestion)

			// Confirmar reset de contraseña (público)
			auth.POST("/reset-password",
				authRateLimit,
				middleware.NoCache(),
				middleware.UserActivityLogger("PASSWORD_RESET_CONFIRM"),
				authHandler.ConfirmPasswordReset)

			// ========================================
			// RUTAS PROTEGIDAS DE AUTENTICACIÓN
			// ========================================

			protected := auth.Group("/")
			protected.Use(middleware.AuthMiddleware(appCtx))
			{
				// Operaciones básicas de autenticación
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
				// RUTAS PROTEGIDAS DE PREGUNTAS DE SEGURIDAD
				// ========================================

				// Obtener estado de preguntas de seguridad del usuario
				protected.GET("/security-status",
					middleware.UserActivityLogger("GET_SECURITY_STATUS"),
					authHandler.GetUserSecurityStatus)

				// Configurar preguntas de seguridad
				protected.POST("/security-questions",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("SETUP_SECURITY_QUESTIONS"),
					authHandler.SetupSecurityQuestions)

				// Actualizar una pregunta de seguridad específica
				protected.PUT("/security-questions/:questionId",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("UPDATE_SECURITY_QUESTION"),
					authHandler.UpdateSecurityQuestion)

				// Eliminar una pregunta de seguridad específica
				protected.DELETE("/security-questions/:questionId",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("REMOVE_SECURITY_QUESTION"),
					authHandler.RemoveSecurityQuestion)

				// ========================================
				// RUTAS PROTEGIDAS DE HISTORIAL DE RESET
				// ========================================

				// Ver historial de reset del usuario actual
				protected.GET("/reset-history",
					middleware.UserActivityLogger("GET_RESET_HISTORY"),
					authHandler.GetPasswordResetHistory)
			}

			// ========================================
			// RUTAS DE ADMINISTRACIÓN DE AUTH (solo admins)
			// ========================================

			adminAuth := auth.Group("/admin")
			adminAuth.Use(middleware.AuthMiddleware(appCtx))
			adminAuth.Use(middleware.RequireRole("admin"))
			{
				// Limpiar tokens expirados (solo admins)
				adminAuth.POST("/cleanup-tokens",
					authRateLimit,
					middleware.NoCache(),
					middleware.UserActivityLogger("CLEANUP_RESET_TOKENS"),
					authHandler.CleanupExpiredTokens)

				// Gestión de preguntas de seguridad (futuro)
				adminAuth.GET("/security-questions/stats", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Security questions stats - futuro",
						"status":  "coming_soon",
					})
				})

				// Auditoría de resets de contraseña (futuro)
				adminAuth.GET("/reset-audit", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Password reset audit - futuro",
						"status":  "coming_soon",
					})
				})
			}
		}

		// ========================================
		// RUTAS DE MENSAJERÍA (FUNCIONALES)
		// ========================================

		// Crear handler de mensajes
		messageService := services.NewMessageService(appCtx.DB)
		messageHandler := handlers.NewMessageHandler(messageService)

		messages := apiV1.Group("/messages")
		messages.Use(middleware.AuthMiddleware(appCtx))
		{
			// Rutas públicas de mensajes (solo autenticación requerida)
			messages.GET("/types", messageHandler.GetMessageTypes)
			messages.GET("/statuses", messageHandler.GetMessageStatuses)

			// Rutas básicas de mensajes
			// IMPORTANTE: Usar rutas sin trailing slash
			messages.GET("", messageHandler.GetMessages)
			messages.POST("", messageHandler.CreateMessage)
			messages.GET("/:id", messageHandler.GetMessageByID)
			messages.PUT("/:id/read", messageHandler.MarkAsRead)
			messages.PUT("/:id/status", messageHandler.UpdateMessageStatus)
			messages.DELETE("/:id", messageHandler.DeleteMessage)

			// Estadísticas (solo admin - se valida internamente)
			messages.GET("/stats", messageHandler.GetMessageStats)
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
		// RUTAS DE ADMINISTRACIÓN GENERAL
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

			// ========================================
			// ADMINISTRACIÓN DE SEGURIDAD
			// ========================================

			security := admin.Group("/security")
			{
				// Gestión de preguntas de seguridad del sistema
				security.GET("/questions", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Admin security questions management - futuro",
						"status":  "coming_soon",
					})
				})

				security.POST("/questions", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Create security question - futuro",
						"status":  "coming_soon",
					})
				})

				security.PUT("/questions/:id", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Update security question - futuro",
						"status":  "coming_soon",
					})
				})

				// Reportes de seguridad
				security.GET("/reports", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Security reports - futuro",
						"status":  "coming_soon",
					})
				})

				// Configuración de políticas de seguridad
				security.GET("/policies", func(c *gin.Context) {
					c.JSON(200, gin.H{
						"message": "Security policies management - futuro",
						"status":  "coming_soon",
					})
				})
			}
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
				"auth": gin.H{
					"public": []string{
						"POST /api/v1/auth/login",
						"POST /api/v1/auth/register",
						"POST /api/v1/auth/refresh",
						"GET  /api/v1/auth/security-questions",
						"POST /api/v1/auth/forgot-password",
						"GET  /api/v1/auth/reset-status/:token",
						"POST /api/v1/auth/verify-security-question",
						"POST /api/v1/auth/reset-password",
					},
					"protected": []string{
						"POST /api/v1/auth/logout",
						"GET  /api/v1/auth/profile",
						"PUT  /api/v1/auth/change-password",
						"GET  /api/v1/auth/verify",
						"GET  /api/v1/auth/security-status",
						"POST /api/v1/auth/security-questions",
						"PUT  /api/v1/auth/security-questions/:questionId",
						"DELETE /api/v1/auth/security-questions/:questionId",
						"GET  /api/v1/auth/reset-history",
					},
					"admin": []string{
						"POST /api/v1/auth/admin/cleanup-tokens",
						"GET  /api/v1/auth/admin/security-questions/stats",
						"GET  /api/v1/auth/admin/reset-audit",
					},
				},
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

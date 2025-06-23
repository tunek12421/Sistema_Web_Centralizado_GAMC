package routes

import (
	"time"

	"gamc-backend-go/internal/api/handlers"
	"gamc-backend-go/internal/api/middleware"
	"gamc-backend-go/internal/config"

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
			}
		}
	}

	// ========================================
	// RUTAS DE ADMINISTRACIÓN (futuras)
	// ========================================

	admin := apiV1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(appCtx))
	admin.Use(middleware.RequireAdmin())
	{
		// Rutas de admin se pueden agregar aquí en el futuro
		admin.GET("/users", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin users endpoint - Coming soon"})
		})

		admin.GET("/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Admin stats endpoint - Coming soon"})
		})
	}

	// ========================================
	// RUTAS DE MENSAJERÍA (futuras)
	// ========================================

	messages := apiV1.Group("/messages")
	messages.Use(middleware.AuthMiddleware(appCtx))
	messages.Use(middleware.RequireOutputRole()) // Requiere rol output o superior
	{
		// Endpoints de mensajes se pueden agregar aquí
		messages.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Messages endpoint - Coming soon"})
		})

		// Solo usuarios con rol input pueden crear mensajes
		messages.POST("/",
			middleware.RequireInputRole(),
			func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Create message endpoint - Coming soon"})
			})
	}

	// ========================================
	// RUTAS DE ARCHIVOS (futuras)
	// ========================================

	files := apiV1.Group("/files")
	files.Use(middleware.AuthMiddleware(appCtx))
	{
		files.POST("/upload", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "File upload endpoint - Coming soon"})
		})

		files.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "File download endpoint - Coming soon"})
		})
	}

	// ========================================
	// CATCH-ALL PARA 404
	// ========================================

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"success":   false,
			"message":   "Endpoint no encontrado",
			"path":      c.Request.URL.Path,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	return router
}

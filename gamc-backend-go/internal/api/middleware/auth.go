package middleware

import (
	"net/http"
	//	"strings"
	"time"

	"gamc-backend-go/internal/auth"
	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/redis"
	"gamc-backend-go/internal/services"
	"gamc-backend-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware middleware de autenticación JWT
func AuthMiddleware(appCtx *config.AppContext) gin.HandlerFunc {
	jwtService := auth.NewJWTService(appCtx.Config)
	sessionManager := redis.NewSessionManager(appCtx.Redis)
	blacklistManager := redis.NewJWTBlacklistManager(appCtx.Redis)
	authService := services.NewAuthService(appCtx)

	return func(c *gin.Context) {
		// Extraer token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Token de acceso requerido", "")
			c.Abort()
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "Formato de token inválido", "")
			c.Abort()
			return
		}

		// Verificar y parsear token
		claims, err := jwtService.VerifyAccessToken(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Token inválido o expirado", "")
			c.Abort()
			return
		}

		// Verificar blacklist usando JTI si está disponible
		if claims.JTI != "" {
			isBlacklisted, err := blacklistManager.IsTokenBlacklisted(c.Request.Context(), claims.JTI)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
				c.Abort()
				return
			}
			if isBlacklisted {
				response.Error(c, http.StatusUnauthorized, "Token revocado", "")
				c.Abort()
				return
			}
		}

		// Verificar que la sesión existe en Redis
		sessionData, err := sessionManager.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
			c.Abort()
			return
		}
		if sessionData == nil {
			response.Error(c, http.StatusUnauthorized, "Sesión expirada", "")
			c.Abort()
			return
		}

		// Verificar que los datos del token coincidan con la sesión
		if sessionData.UserID != claims.UserID || sessionData.Email != claims.Email {
			response.Error(c, http.StatusUnauthorized, "Datos de sesión inconsistentes", "")
			c.Abort()
			return
		}

		// Obtener perfil actualizado del usuario
		userProfile, err := authService.GetUserProfile(c.Request.Context(), claims.UserID)
		if err != nil || userProfile == nil || !userProfile.IsActive {
			response.Error(c, http.StatusUnauthorized, "Usuario no encontrado o inactivo", "")
			c.Abort()
			return
		}

		// Actualizar última actividad en la sesión
		sessionData.LastActivity = time.Now()
		sessionManager.SaveSession(c.Request.Context(), claims.SessionID, sessionData, 7*24*time.Hour)

		// Agregar datos al contexto
		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.SessionID)
		c.Set("user", userProfile)
		c.Set("claims", claims)

		c.Next()
	}
}

// OptionalAuthMiddleware middleware de autenticación opcional
func OptionalAuthMiddleware(appCtx *config.AppContext) gin.HandlerFunc {
	jwtService := auth.NewJWTService(appCtx.Config)
	sessionManager := redis.NewSessionManager(appCtx.Redis)
	authService := services.NewAuthService(appCtx)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token := auth.ExtractTokenFromHeader(authHeader)
		if token == "" {
			c.Next()
			return
		}

		// Intentar verificar token (sin fallar si es inválido)
		claims, err := jwtService.VerifyAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Verificar sesión
		sessionData, err := sessionManager.GetSession(c.Request.Context(), claims.SessionID)
		if err != nil || sessionData == nil {
			c.Next()
			return
		}

		// Obtener perfil de usuario
		userProfile, err := authService.GetUserProfile(c.Request.Context(), claims.UserID)
		if err != nil || userProfile == nil || !userProfile.IsActive {
			c.Next()
			return
		}

		// Agregar datos al contexto si todo está bien
		c.Set("userID", claims.UserID)
		c.Set("sessionID", claims.SessionID)
		c.Set("user", userProfile)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole middleware para verificar roles específicos
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Autenticación requerida", "")
			c.Abort()
			return
		}

		userProfile, ok := user.(*models.UserProfile)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
			c.Abort()
			return
		}

		// Verificar si el rol está permitido
		for _, role := range allowedRoles {
			if userProfile.Role == role {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "Permisos insuficientes", "")
		c.Abort()
	}
}

// RequireOrganizationalUnit middleware para verificar unidad organizacional
func RequireOrganizationalUnit(allowedUnitIDs ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Autenticación requerida", "")
			c.Abort()
			return
		}

		userProfile, ok := user.(*models.UserProfile)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
			c.Abort()
			return
		}

		if userProfile.OrganizationalUnit == nil {
			response.Error(c, http.StatusForbidden, "Usuario sin unidad organizacional", "")
			c.Abort()
			return
		}

		// Verificar si la unidad está permitida
		for _, unitID := range allowedUnitIDs {
			if userProfile.OrganizationalUnit.ID == unitID {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "Acceso restringido a su unidad organizacional", "")
		c.Abort()
	}
}

// RequireAdmin middleware para verificar que el usuario es admin
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RequireInputRole middleware para verificar que el usuario puede enviar mensajes
func RequireInputRole() gin.HandlerFunc {
	return RequireRole("admin", "input")
}

// RequireOutputRole middleware para verificar que el usuario puede ver mensajes
func RequireOutputRole() gin.HandlerFunc {
	return RequireRole("admin", "input", "output")
}

// RequireOwnership middleware para verificar propiedad del recurso
func RequireOwnership(getUserIDFromResource func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Autenticación requerida", "")
			c.Abort()
			return
		}

		userProfile, ok := user.(*models.UserProfile)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
			c.Abort()
			return
		}

		// Los admins pueden acceder a cualquier recurso
		if userProfile.Role == "admin" {
			c.Next()
			return
		}

		resourceUserID := getUserIDFromResource(c)
		if userProfile.ID.String() != resourceUserID {
			response.Error(c, http.StatusForbidden, "Solo puede acceder a sus propios recursos", "")
			c.Abort()
			return
		}

		c.Next()
	}
}

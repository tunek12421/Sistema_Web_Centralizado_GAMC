package handlers

import (
	"net/http"
	"time"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/services"
	"gamc-backend-go/pkg/response"
	"gamc-backend-go/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AuthHandler maneja las operaciones de autenticación
type AuthHandler struct {
	authService *services.AuthService
	config      *config.Config
}

// NewAuthHandler crea una nueva instancia del handler de autenticación
func NewAuthHandler(appCtx *config.AppContext) *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(appCtx),
		config:      appCtx.Config,
	}
}

// Login maneja POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos de entrada
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Obtener información del cliente
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Ejecutar login
	result, err := h.authService.Login(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Error de autenticación", err.Error())
		return
	}

	// Configurar cookie HttpOnly para refresh token
	c.SetCookie(
		"refreshToken",
		result.RefreshToken,
		int(7*24*time.Hour.Seconds()), // 7 días
		"/",
		"",
		h.config.Environment == "production", // Secure
		true,                                 // HttpOnly
	)

	// Respuesta exitosa (sin incluir refresh token en JSON)
	response.Success(c, "Login exitoso", gin.H{
		"user":        result.User,
		"accessToken": result.AccessToken,
		"expiresIn":   result.ExpiresIn,
	})
}

// Register maneja POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos de entrada
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Ejecutar registro
	userProfile, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Error en registro", err.Error())
		return
	}

	response.Success(c, "Usuario registrado exitosamente", userProfile)
}

// RefreshToken maneja POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Obtener refresh token desde cookie o body
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil || refreshToken == "" {
		// Intentar desde body como fallback
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := c.ShouldBindJSON(&body); err != nil || body.RefreshToken == "" {
			response.Error(c, http.StatusUnauthorized, "Refresh token requerido", "")
			return
		}
		refreshToken = body.RefreshToken
	}

	// Ejecutar refresh
	result, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "Error al renovar token", err.Error())
		return
	}

	// Actualizar cookie
	c.SetCookie(
		"refreshToken",
		result.RefreshToken,
		int(7*24*time.Hour.Seconds()),
		"/",
		"",
		h.config.Environment == "production",
		true,
	)

	response.Success(c, "Token renovado exitosamente", gin.H{
		"user":        result.User,
		"accessToken": result.AccessToken,
		"expiresIn":   result.ExpiresIn,
	})
}

// Logout maneja POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	sessionID, exists := c.Get("sessionID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Sesión no válida", "")
		return
	}

	// Obtener parámetro logoutAll desde body
	var body struct {
		LogoutAll bool `json:"logoutAll"`
	}
	c.ShouldBindJSON(&body)

	// Ejecutar logout
	if err := h.authService.Logout(c.Request.Context(), userID.(string), sessionID.(string), body.LogoutAll); err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al cerrar sesión", err.Error())
		return
	}

	// Limpiar cookie
	c.SetCookie("refreshToken", "", -1, "/", "", false, true)

	message := "Logout exitoso"
	if body.LogoutAll {
		message = "Logout de todas las sesiones exitoso"
	}

	response.Success(c, message, nil)
}

// GetProfile maneja GET /api/v1/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	// Obtener perfil actualizado
	profile, err := h.authService.GetUserProfile(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Perfil no encontrado", err.Error())
		return
	}

	response.Success(c, "Perfil obtenido exitosamente", profile)
}

// ChangePassword maneja PUT /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Ejecutar cambio de contraseña
	if err := h.authService.ChangePassword(c.Request.Context(), userID.(string), &req); err != nil {
		response.Error(c, http.StatusBadRequest, "Error al cambiar contraseña", err.Error())
		return
	}

	response.Success(c, "Contraseña cambiada exitosamente", nil)
}

// VerifyToken maneja GET /api/v1/auth/verify
func (h *AuthHandler) VerifyToken(c *gin.Context) {
	// Obtener datos del usuario desde middleware
	userID, userExists := c.Get("userID")
	user, profileExists := c.Get("user")

	if !userExists || !profileExists {
		response.Error(c, http.StatusUnauthorized, "Token inválido", "")
		return
	}

	response.Success(c, "Token válido", gin.H{
		"valid":  true,
		"userID": userID,
		"user":   user,
	})
}

// ========================================
// HANDLERS DE RESET DE CONTRASEÑA
// ========================================

// RequestPasswordReset maneja POST /api/v1/auth/forgot-password
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req models.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos de entrada
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Obtener información del cliente
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Ejecutar solicitud de reset
	err := h.authService.RequestPasswordReset(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		// Clasificar errores por tipo
		if err.Error() == "solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña" {
			response.Error(c, http.StatusForbidden, "Email no autorizado", "Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña")
			return
		}
		if err.Error() == "debe esperar 5 minutos entre solicitudes de reset" {
			response.Error(c, http.StatusTooManyRequests, "Demasiadas solicitudes", "Debe esperar 5 minutos entre solicitudes de reset")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Error interno", "Error al procesar solicitud de reset")
		return
	}

	// Respuesta exitosa (siempre igual por seguridad)
	response.Success(c, "Solicitud de reset procesada", gin.H{
		"message": "Si el email existe y es válido, recibirás un enlace de reset en tu correo institucional",
		"note":    "Solo emails @gamc.gov.bo pueden solicitar reset de contraseña",
	})
}

// ConfirmPasswordReset maneja POST /api/v1/auth/reset-password
func (h *AuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req models.PasswordResetConfirm
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos de entrada
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Obtener información del cliente
	ipAddress := c.ClientIP()

	// Ejecutar confirmación de reset
	err := h.authService.ConfirmPasswordReset(c.Request.Context(), &req, ipAddress)
	if err != nil {
		// Clasificar errores por tipo
		if err.Error() == "token de reset inválido o expirado" {
			response.Error(c, http.StatusBadRequest, "Token inválido", "El token de reset es inválido o ha expirado")
			return
		}
		if err.Error() == "token de reset expirado" {
			response.Error(c, http.StatusBadRequest, "Token expirado", "El token de reset ha expirado. Solicite un nuevo reset")
			return
		}
		if err.Error() == "token de reset ya utilizado" {
			response.Error(c, http.StatusBadRequest, "Token usado", "Este token ya fue utilizado. Solicite un nuevo reset si es necesario")
			return
		}
		response.Error(c, http.StatusBadRequest, "Error en reset", err.Error())
		return
	}

	response.Success(c, "Contraseña resetiada exitosamente", gin.H{
		"message": "Su contraseña ha sido cambiada exitosamente. Por seguridad, se han cerrado todas sus sesiones activas",
		"note":    "Inicie sesión con su nueva contraseña",
	})
}

// GetPasswordResetStatus maneja GET /api/v1/auth/reset-status (endpoint protegido para admins)
func (h *AuthHandler) GetPasswordResetStatus(c *gin.Context) {
	// Este endpoint está protegido por middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	// Obtener estado de tokens de reset
	tokens, err := h.authService.GetPasswordResetStatus(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al obtener estado", err.Error())
		return
	}

	response.Success(c, "Estado de reset obtenido", gin.H{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

// CleanupExpiredTokens maneja POST /api/v1/auth/cleanup-tokens (endpoint protegido para admins)
func (h *AuthHandler) CleanupExpiredTokens(c *gin.Context) {
	// Este endpoint está protegido por middleware de autenticación y requiere rol admin

	// Ejecutar limpieza
	cleanedCount, err := h.authService.CleanupExpiredResetTokens(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error en limpieza", err.Error())
		return
	}

	response.Success(c, "Limpieza completada", gin.H{
		"cleanedTokens": cleanedCount,
		"timestamp":     time.Now(),
	})
}

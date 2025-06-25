package handlers

import (
	"net/http"
	"strconv"
	"strings"
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
// HANDLERS DE PREGUNTAS DE SEGURIDAD
// ========================================

// GetSecurityQuestions maneja GET /api/v1/auth/security-questions
func (h *AuthHandler) GetSecurityQuestions(c *gin.Context) {
	questions, err := h.authService.GetSecurityQuestions(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al obtener preguntas", err.Error())
		return
	}

	// Convertir a respuestas
	var questionResponses []models.SecurityQuestionResponse
	for _, q := range questions {
		questionResponses = append(questionResponses, *q.ToResponse())
	}

	response.Success(c, "Preguntas de seguridad obtenidas", gin.H{
		"questions": questionResponses,
		"count":     len(questionResponses),
	})
}

// GetUserSecurityStatus maneja GET /api/v1/auth/security-status
func (h *AuthHandler) GetUserSecurityStatus(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	status, err := h.authService.GetUserSecurityStatus(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al obtener estado", err.Error())
		return
	}

	response.Success(c, "Estado de seguridad obtenido", status)
}

// SetupSecurityQuestions maneja POST /api/v1/auth/security-questions
func (h *AuthHandler) SetupSecurityQuestions(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	var req models.SecurityQuestionSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos de entrada
	if err := validator.Validate(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Ejecutar configuración
	if err := h.authService.SetupSecurityQuestions(c.Request.Context(), userID.(string), &req); err != nil {
		response.Error(c, http.StatusBadRequest, "Error al configurar preguntas", err.Error())
		return
	}

	response.Success(c, "Preguntas de seguridad configuradas exitosamente", gin.H{
		"message":      "Las preguntas de seguridad han sido configuradas",
		"questionsSet": len(req.Questions),
	})
}

// UpdateSecurityQuestion maneja PUT /api/v1/auth/security-questions/:questionId
func (h *AuthHandler) UpdateSecurityQuestion(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	// Obtener ID de pregunta desde parámetros
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de pregunta inválido", "")
		return
	}

	var body struct {
		NewAnswer string `json:"newAnswer" validate:"required,min=2,max=100"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	// Validar datos
	if err := validator.Validate(&body); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos de entrada inválidos", err.Error())
		return
	}

	req := &models.SecurityQuestionUpdateRequest{
		QuestionID: questionID,
		NewAnswer:  body.NewAnswer,
	}

	// Ejecutar actualización
	if err := h.authService.UpdateSecurityQuestion(c.Request.Context(), userID.(string), req); err != nil {
		response.Error(c, http.StatusBadRequest, "Error al actualizar pregunta", err.Error())
		return
	}

	response.Success(c, "Pregunta de seguridad actualizada exitosamente", gin.H{
		"questionId": questionID,
		"message":    "La respuesta ha sido actualizada",
	})
}

// RemoveSecurityQuestion maneja DELETE /api/v1/auth/security-questions/:questionId
func (h *AuthHandler) RemoveSecurityQuestion(c *gin.Context) {
	// Obtener usuario desde middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	// Obtener ID de pregunta desde parámetros
	questionIDStr := c.Param("questionId")
	questionID, err := strconv.Atoi(questionIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de pregunta inválido", "")
		return
	}

	// Ejecutar eliminación
	if err := h.authService.RemoveSecurityQuestion(c.Request.Context(), userID.(string), questionID); err != nil {
		response.Error(c, http.StatusBadRequest, "Error al eliminar pregunta", err.Error())
		return
	}

	response.Success(c, "Pregunta de seguridad eliminada exitosamente", gin.H{
		"questionId": questionID,
		"message":    "La pregunta ha sido eliminada",
	})
}

// ========================================
// HANDLERS DE RESET DE CONTRASEÑA CON PREGUNTAS DE SEGURIDAD
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
	result, err := h.authService.RequestPasswordReset(c.Request.Context(), &req, ipAddress, userAgent)
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

	// Respuesta exitosa con información de pregunta de seguridad si aplica
	response.Success(c, "Solicitud de reset procesada", result)
}

// GetPasswordResetStatus maneja GET /api/v1/auth/reset-status?email=
func (h *AuthHandler) GetPasswordResetStatus(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "Email requerido", "Debe proporcionar el parámetro email")
		return
	}

	// Obtener estado del proceso de reset por email
	status, err := h.authService.GetPasswordResetStatusByEmail(c.Request.Context(), email)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al obtener estado", err.Error())
		return
	}

	response.Success(c, "Estado de reset obtenido", status)
}

// VerifySecurityQuestion maneja POST /api/v1/auth/verify-security-question
func (h *AuthHandler) VerifySecurityQuestion(c *gin.Context) {
	var req models.PasswordResetVerifySecurityRequest
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

	// Ejecutar verificación
	result, err := h.authService.VerifySecurityQuestion(c.Request.Context(), &req, ipAddress)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al verificar pregunta", err.Error())
		return
	}

	// Respuesta basada en el resultado
	if result.Success {
		response.Success(c, "Pregunta verificada exitosamente", result)
	} else {
		response.Error(c, http.StatusBadRequest, "Verificación fallida", result.Message)
	}
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
		switch err.Error() {
		case "token de reset inválido o expirado":
			response.Error(c, http.StatusBadRequest, "Token inválido", "El token de reset es inválido o ha expirado")
		case "token de reset expirado":
			response.Error(c, http.StatusBadRequest, "Token expirado", "El token de reset ha expirado. Solicite un nuevo reset")
		case "token de reset ya utilizado":
			response.Error(c, http.StatusBadRequest, "Token usado", "Este token ya fue utilizado. Solicite un nuevo reset si es necesario")
		case "debe verificar la pregunta de seguridad primero":
			response.Error(c, http.StatusBadRequest, "Verificación requerida", "Debe verificar la pregunta de seguridad antes de cambiar la contraseña")
		case "demasiados intentos fallidos - token invalidado":
			response.Error(c, http.StatusBadRequest, "Token invalidado", "Demasiados intentos fallidos. Solicite un nuevo reset")
		default:
			if strings.Contains(err.Error(), "respuesta de seguridad incorrecta") {
				response.Error(c, http.StatusBadRequest, "Respuesta incorrecta", err.Error())
			} else {
				response.Error(c, http.StatusBadRequest, "Error en reset", err.Error())
			}
		}
		return
	}

	response.Success(c, "Contraseña resetiada exitosamente", gin.H{
		"message": "Su contraseña ha sido cambiada exitosamente. Por seguridad, se han cerrado todas sus sesiones activas",
		"note":    "Inicie sesión con su nueva contraseña",
	})
}

// ========================================
// HANDLERS ADMINISTRATIVOS
// ========================================

// GetPasswordResetHistory maneja GET /api/v1/auth/reset-history (endpoint protegido)
func (h *AuthHandler) GetPasswordResetHistory(c *gin.Context) {
	// Este endpoint está protegido por middleware de autenticación
	userID, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	// Obtener historial de tokens de reset
	tokens, err := h.authService.GetPasswordResetHistory(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al obtener historial", err.Error())
		return
	}

	response.Success(c, "Historial de reset obtenido", gin.H{
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

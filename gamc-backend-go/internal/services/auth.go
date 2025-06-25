package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gamc-backend-go/internal/auth"
	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/redis"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthService maneja la lógica de negocio de autenticación
type AuthService struct {
	db               *gorm.DB
	sessionManager   *redis.SessionManager
	refreshManager   *redis.RefreshTokenManager
	blacklistManager *redis.JWTBlacklistManager
	jwtService       *auth.JWTService
	passwordService  *auth.PasswordService
	config           *config.Config
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(appCtx *config.AppContext) *AuthService {
	return &AuthService{
		db:               appCtx.DB,
		sessionManager:   redis.NewSessionManager(appCtx.Redis),
		refreshManager:   redis.NewRefreshTokenManager(appCtx.Redis),
		blacklistManager: redis.NewJWTBlacklistManager(appCtx.Redis),
		jwtService:       auth.NewJWTService(appCtx.Config),
		passwordService:  auth.NewPasswordService(),
		config:           appCtx.Config,
	}
}

// LoginRequest representa una solicitud de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest representa una solicitud de registro
type RegisterRequest struct {
	Email                string                               `json:"email" validate:"required,email"`
	Password             string                               `json:"password" validate:"required,min=8"`
	FirstName            string                               `json:"firstName" validate:"required"`
	LastName             string                               `json:"lastName" validate:"required"`
	OrganizationalUnitID int                                  `json:"organizationalUnitId" validate:"required"`
	Role                 string                               `json:"role,omitempty" validate:"omitempty,oneof=admin input output"`
	SecurityQuestions    *models.SecurityQuestionSetupRequest `json:"securityQuestions,omitempty"`
}

// ChangePasswordRequest representa una solicitud de cambio de contraseña
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// AuthResponse representa la respuesta de autenticación
type AuthResponse struct {
	User         *models.UserProfile `json:"user"`
	AccessToken  string              `json:"accessToken"`
	RefreshToken string              `json:"refreshToken,omitempty"`
	ExpiresIn    int64               `json:"expiresIn"`
}

// Login autentica un usuario y genera tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest, ipAddress, userAgent string) (*AuthResponse, error) {
	// Buscar usuario por email
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("email = ? AND is_active = ?", req.Email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("credenciales inválidas")
		}
		return nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Verificar contraseña
	if err := s.passwordService.ComparePassword(req.Password, user.PasswordHash); err != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}

	// Generar ID de sesión
	sessionID := uuid.New().String()

	// Crear datos de sesión
	sessionData := &redis.SessionData{
		UserID:               user.ID.String(),
		Email:                user.Email,
		Role:                 user.Role,
		OrganizationalUnitID: *user.OrganizationalUnitID,
		SessionID:            sessionID,
		CreatedAt:            time.Now(),
		LastActivity:         time.Now(),
		IPAddress:            ipAddress,
		UserAgent:            userAgent,
	}

	// Guardar sesión en Redis
	sessionTTL := 7 * 24 * time.Hour // 7 días
	if err := s.sessionManager.SaveSession(ctx, sessionID, sessionData, sessionTTL); err != nil {
		return nil, fmt.Errorf("error al guardar sesión: %w", err)
	}

	// Generar tokens JWT
	accessToken, refreshToken, err := s.jwtService.GenerateTokens(
		user.ID.String(),
		user.Email,
		user.Role,
		*user.OrganizationalUnitID,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("error al generar tokens: %w", err)
	}

	// Guardar refresh token en Redis
	refreshTTL := 7 * 24 * time.Hour // 7 días
	if err := s.refreshManager.SaveRefreshToken(ctx, user.ID.String(), sessionID, refreshToken, refreshTTL); err != nil {
		return nil, fmt.Errorf("error al guardar refresh token: %w", err)
	}

	// Actualizar último login
	now := time.Now()
	user.LastLogin = &now
	s.db.WithContext(ctx).Save(&user)

	// Crear perfil de usuario para respuesta
	userProfile := user.ToProfile()

	return &AuthResponse{
		User:         userProfile,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.JWTExpiresIn.Seconds()),
	}, nil
}

// Register registra un nuevo usuario
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*models.UserProfile, error) {
	// Verificar si el usuario ya existe
	var existingUser models.User
	err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error
	if err == nil {
		return nil, fmt.Errorf("el usuario ya existe")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error al verificar usuario existente: %w", err)
	}

	// Verificar que existe la unidad organizacional
	var orgUnit models.OrganizationalUnit
	err = s.db.WithContext(ctx).Where("id = ? AND is_active = ?", req.OrganizationalUnitID, true).First(&orgUnit).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unidad organizacional no válida")
		}
		return nil, fmt.Errorf("error al verificar unidad organizacional: %w", err)
	}

	// Validar contraseña
	if isValid, validationErrors := s.passwordService.IsValidPassword(req.Password); !isValid {
		return nil, fmt.Errorf("contraseña inválida: %v", validationErrors)
	}

	// Validar preguntas de seguridad si se proporcionaron
	if req.SecurityQuestions != nil {
		if validationErrors := models.ValidateSecurityQuestionSetup(req.SecurityQuestions); len(validationErrors) > 0 {
			return nil, fmt.Errorf("preguntas de seguridad inválidas: %v", validationErrors)
		}

		// Verificar que las preguntas existen y están activas
		for _, q := range req.SecurityQuestions.Questions {
			var secQuestion models.SecurityQuestion
			err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", q.QuestionID, true).First(&secQuestion).Error
			if err != nil {
				return nil, fmt.Errorf("pregunta de seguridad ID %d no válida", q.QuestionID)
			}
		}
	}

	// Generar username único
	username, err := s.generateUniqueUsername(ctx, req.FirstName, req.LastName, req.Email, req.OrganizationalUnitID)
	if err != nil {
		return nil, fmt.Errorf("error al generar username: %w", err)
	}

	// Hashear contraseña
	passwordHash, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error al hashear contraseña: %w", err)
	}

	// Establecer role por defecto
	role := req.Role
	if role == "" {
		role = "output"
	}

	// Crear usuario en transacción
	var user models.User
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Crear usuario
		user = models.User{
			Username:             username,
			Email:                req.Email,
			PasswordHash:         passwordHash,
			FirstName:            req.FirstName,
			LastName:             req.LastName,
			Role:                 role,
			OrganizationalUnitID: &req.OrganizationalUnitID,
			IsActive:             true,
			PasswordChangedAt:    time.Now(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("error al crear usuario: %w", err)
		}

		// Configurar preguntas de seguridad si se proporcionaron
		if req.SecurityQuestions != nil {
			for _, q := range req.SecurityQuestions.Questions {
				// Hashear respuesta
				answerHash := s.hashSecurityAnswer(q.Answer)

				userSecQuestion := models.UserSecurityQuestion{
					UserID:             user.ID,
					SecurityQuestionID: q.QuestionID,
					AnswerHash:         answerHash,
					IsActive:           true,
				}

				if err := tx.Create(&userSecQuestion).Error; err != nil {
					return fmt.Errorf("error al crear pregunta de seguridad: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Cargar la unidad organizacional para la respuesta
	user.OrganizationalUnit = &orgUnit

	logger.Info("✅ Usuario registrado exitosamente: %s (%s) - %s", username, req.Email, orgUnit.Name)
	if req.SecurityQuestions != nil {
		logger.Info("✅ Preguntas de seguridad configuradas: %d", len(req.SecurityQuestions.Questions))
	}

	return user.ToProfile(), nil
}

// RefreshToken renueva un access token usando un refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Verificar refresh token
	claims, err := s.jwtService.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token inválido: %w", err)
	}

	// Verificar que el refresh token existe en Redis
	storedToken, err := s.refreshManager.GetRefreshToken(ctx, claims.UserID, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener refresh token: %w", err)
	}
	if storedToken != refreshToken {
		return nil, fmt.Errorf("refresh token inválido")
	}

	// Obtener datos de sesión
	sessionData, err := s.sessionManager.GetSession(ctx, claims.SessionID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener sesión: %w", err)
	}
	if sessionData == nil {
		return nil, fmt.Errorf("sesión expirada")
	}

	// Generar nuevos tokens
	newAccessToken, newRefreshToken, err := s.jwtService.GenerateTokens(
		sessionData.UserID,
		sessionData.Email,
		sessionData.Role,
		sessionData.OrganizationalUnitID,
		sessionData.SessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("error al generar nuevos tokens: %w", err)
	}

	// Actualizar refresh token en Redis
	refreshTTL := 7 * 24 * time.Hour
	if err := s.refreshManager.SaveRefreshToken(ctx, sessionData.UserID, sessionData.SessionID, newRefreshToken, refreshTTL); err != nil {
		return nil, fmt.Errorf("error al actualizar refresh token: %w", err)
	}

	// Obtener perfil actualizado del usuario
	userProfile, err := s.GetUserProfile(ctx, sessionData.UserID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener perfil de usuario: %w", err)
	}

	return &AuthResponse{
		User:         userProfile,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.config.JWTExpiresIn.Seconds()),
	}, nil
}

// Logout cierra la sesión del usuario
func (s *AuthService) Logout(ctx context.Context, userID, sessionID string, logoutAll bool) error {
	if logoutAll {
		// Logout de todas las sesiones del usuario
		sessions, err := s.sessionManager.GetUserSessions(ctx, userID)
		if err != nil {
			return fmt.Errorf("error al obtener sesiones de usuario: %w", err)
		}

		for _, sid := range sessions {
			s.sessionManager.DeleteSession(ctx, sid)
			s.refreshManager.DeleteRefreshToken(ctx, userID, sid)
		}
	} else {
		// Logout de sesión específica
		if err := s.sessionManager.DeleteSession(ctx, sessionID); err != nil {
			return fmt.Errorf("error al eliminar sesión: %w", err)
		}
		if err := s.refreshManager.DeleteRefreshToken(ctx, userID, sessionID); err != nil {
			return fmt.Errorf("error al eliminar refresh token: %w", err)
		}
	}

	return nil
}

// ChangePassword cambia la contraseña del usuario
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *ChangePasswordRequest) error {
	// Buscar usuario
	var user models.User
	err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("usuario no encontrado")
		}
		return fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Verificar contraseña actual
	if err := s.passwordService.ComparePassword(req.CurrentPassword, user.PasswordHash); err != nil {
		return fmt.Errorf("contraseña actual incorrecta")
	}

	// Validar nueva contraseña
	if isValid, validationErrors := s.passwordService.IsValidPassword(req.NewPassword); !isValid {
		return fmt.Errorf("nueva contraseña inválida: %v", validationErrors)
	}

	// Hashear nueva contraseña
	newPasswordHash, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("error al hashear nueva contraseña: %w", err)
	}

	// Actualizar contraseña
	user.PasswordHash = newPasswordHash
	user.PasswordChangedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return fmt.Errorf("error al actualizar contraseña: %w", err)
	}

	return nil
}

// GetUserProfile obtiene el perfil de un usuario
func (s *AuthService) GetUserProfile(ctx context.Context, userID string) (*models.UserProfile, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("OrganizationalUnit").
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("id = ? AND is_active = ?", userID, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al obtener perfil: %w", err)
	}

	return user.ToProfile(), nil
}

// ========================================
// FUNCIONES DE PREGUNTAS DE SEGURIDAD
// ========================================

// GetSecurityQuestions obtiene todas las preguntas de seguridad disponibles
func (s *AuthService) GetSecurityQuestions(ctx context.Context) ([]models.SecurityQuestion, error) {
	var questions []models.SecurityQuestion
	err := s.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category, sort_order, id").
		Find(&questions).Error

	if err != nil {
		return nil, fmt.Errorf("error al obtener preguntas de seguridad: %w", err)
	}

	return questions, nil
}

// GetUserSecurityStatus obtiene el estado de preguntas de seguridad del usuario
func (s *AuthService) GetUserSecurityStatus(ctx context.Context, userID string) (*models.UserSecurityStatusResponse, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("SecurityQuestions", "is_active = ?", true).
		Preload("SecurityQuestions.SecurityQuestion").
		Where("id = ? AND is_active = ?", userID, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("usuario no encontrado")
		}
		return nil, fmt.Errorf("error al obtener usuario: %w", err)
	}

	// Construir respuesta
	status := &models.UserSecurityStatusResponse{
		HasSecurityQuestions: user.HasSecurityQuestionsConfigured(),
		QuestionsCount:       user.GetSecurityQuestionsCount(),
		MaxQuestions:         models.MaxSecurityQuestionsPerUser,
		Questions:            make([]models.UserSecurityQuestionResponse, 0),
	}

	// Convertir preguntas a respuestas
	for _, sq := range user.GetActiveSecurityQuestions() {
		status.Questions = append(status.Questions, *sq.ToResponse())
	}

	// Obtener preguntas disponibles si se pueden agregar más
	if user.CanAddMoreSecurityQuestions() {
		availableQuestions, err := s.GetSecurityQuestions(ctx)
		if err == nil {
			// Filtrar preguntas ya configuradas
			usedQuestionIDs := make(map[int]bool)
			for _, sq := range user.SecurityQuestions {
				if sq.IsActive {
					usedQuestionIDs[sq.SecurityQuestionID] = true
				}
			}

			for _, q := range availableQuestions {
				if !usedQuestionIDs[q.ID] {
					status.AvailableQuestions = append(status.AvailableQuestions, *q.ToResponse())
				}
			}
		}
	}

	return status, nil
}

// SetupSecurityQuestions configura preguntas de seguridad para un usuario
func (s *AuthService) SetupSecurityQuestions(ctx context.Context, userID string, req *models.SecurityQuestionSetupRequest) error {
	// Validar request
	if validationErrors := models.ValidateSecurityQuestionSetup(req); len(validationErrors) > 0 {
		return fmt.Errorf("datos inválidos: %v", validationErrors)
	}

	// Verificar usuario
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("id = ? AND is_active = ?", userID, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("usuario no encontrado")
		}
		return fmt.Errorf("error al obtener usuario: %w", err)
	}

	// Verificar que no exceda el límite
	totalQuestions := user.GetSecurityQuestionsCount() + len(req.Questions)
	if totalQuestions > models.MaxSecurityQuestionsPerUser {
		return fmt.Errorf("no puede tener más de %d preguntas de seguridad", models.MaxSecurityQuestionsPerUser)
	}

	// Verificar que las preguntas no estén ya configuradas
	for _, reqQ := range req.Questions {
		if user.HasSecurityQuestion(reqQ.QuestionID) {
			return fmt.Errorf("la pregunta ID %d ya está configurada", reqQ.QuestionID)
		}
	}

	// Crear preguntas en transacción
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, reqQ := range req.Questions {
			// Verificar que la pregunta existe
			var secQuestion models.SecurityQuestion
			err := tx.Where("id = ? AND is_active = ?", reqQ.QuestionID, true).First(&secQuestion).Error
			if err != nil {
				return fmt.Errorf("pregunta de seguridad ID %d no válida", reqQ.QuestionID)
			}

			// Crear pregunta del usuario
			userSecQuestion := models.UserSecurityQuestion{
				UserID:             user.ID,
				SecurityQuestionID: reqQ.QuestionID,
				AnswerHash:         s.hashSecurityAnswer(reqQ.Answer),
				IsActive:           true,
			}

			if err := tx.Create(&userSecQuestion).Error; err != nil {
				return fmt.Errorf("error al crear pregunta de seguridad: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	logger.Info("Preguntas de seguridad configuradas para usuario %s: %d preguntas", user.Email, len(req.Questions))
	return nil
}

// UpdateSecurityQuestion actualiza una pregunta de seguridad específica
func (s *AuthService) UpdateSecurityQuestion(ctx context.Context, userID string, req *models.SecurityQuestionUpdateRequest) error {
	// Validar respuesta
	if validationErrors := models.ValidateSecurityAnswer(req.NewAnswer); len(validationErrors) > 0 {
		return fmt.Errorf("respuesta inválida: %v", validationErrors)
	}

	// Buscar pregunta del usuario
	var userSecQuestion models.UserSecurityQuestion
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND security_question_id = ? AND is_active = ?", userID, req.QuestionID, true).
		First(&userSecQuestion).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("pregunta de seguridad no encontrada")
		}
		return fmt.Errorf("error al buscar pregunta: %w", err)
	}

	// Actualizar respuesta
	userSecQuestion.AnswerHash = s.hashSecurityAnswer(req.NewAnswer)
	userSecQuestion.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&userSecQuestion).Error; err != nil {
		return fmt.Errorf("error al actualizar pregunta: %w", err)
	}

	logger.Info("Pregunta de seguridad actualizada para usuario %s, pregunta ID %d", userID, req.QuestionID)
	return nil
}

// RemoveSecurityQuestion elimina una pregunta de seguridad del usuario
func (s *AuthService) RemoveSecurityQuestion(ctx context.Context, userID string, questionID int) error {
	// Buscar pregunta del usuario
	var userSecQuestion models.UserSecurityQuestion
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND security_question_id = ? AND is_active = ?", userID, questionID, true).
		First(&userSecQuestion).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("pregunta de seguridad no encontrada")
		}
		return fmt.Errorf("error al buscar pregunta: %w", err)
	}

	// Marcar como inactiva
	userSecQuestion.MarkAsInactive()

	if err := s.db.WithContext(ctx).Save(&userSecQuestion).Error; err != nil {
		return fmt.Errorf("error al eliminar pregunta: %w", err)
	}

	logger.Info("Pregunta de seguridad eliminada para usuario %s, pregunta ID %d", userID, questionID)
	return nil
}

// ========================================
// FUNCIONES DE RESET DE CONTRASEÑA CON PREGUNTAS DE SEGURIDAD
// ========================================

// RequestPasswordReset solicita un reset de contraseña
func (s *AuthService) RequestPasswordReset(ctx context.Context, req *models.PasswordResetRequest, requestIP, userAgent string) (*models.PasswordResetInitResponse, error) {
	// Buscar usuario por email
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("SecurityQuestions", "is_active = ?", true).
		Preload("SecurityQuestions.SecurityQuestion").
		Where("email = ? AND is_active = ?", req.Email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Por seguridad, no revelar que el email no existe
			logger.Info("Intento de reset para email no existente: %s desde IP: %s", req.Email, requestIP)
			// Respuesta genérica
			return &models.PasswordResetInitResponse{
				Success:                  true,
				Message:                  "Si el email existe y es válido, continúe con el proceso",
				RequiresSecurityQuestion: false,
			}, nil
		}
		return nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Validar que es email institucional (doble validación)
	if !strings.HasSuffix(user.Email, "@gamc.gov.bo") {
		logger.Warn("Intento de reset para email no institucional: %s desde IP: %s", user.Email, requestIP)
		return nil, fmt.Errorf("solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña")
	}

	// Generar token único PERO NO LO EXPONEMOS AÚN
	token, err := s.generateResetToken()
	if err != nil {
		return nil, fmt.Errorf("error al generar token: %w", err)
	}

	// Verificar si tiene preguntas de seguridad
	hasSecurityQuestions := user.HasSecurityQuestionsConfigured()

	// Crear registro de reset token
	resetToken := &models.PasswordResetToken{
		UserID:                   user.ID,
		Token:                    token,
		ExpiresAt:                time.Now().Add(30 * time.Minute), // 30 minutos
		RequestIP:                requestIP,
		UserAgent:                &userAgent,
		IsActive:                 true,
		RequiresSecurityQuestion: hasSecurityQuestions,
		SecurityQuestionVerified: !hasSecurityQuestions, // Si no tiene preguntas, ya está verificado
	}

	// Guardar token (el trigger de DB validará automáticamente)
	if err := s.db.WithContext(ctx).Create(resetToken).Error; err != nil {
		if strings.Contains(err.Error(), "Solo usuarios con email @gamc.gov.bo") {
			return nil, fmt.Errorf("solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña")
		}
		if strings.Contains(err.Error(), "Debe esperar 5 minutos") {
			return nil, fmt.Errorf("debe esperar 5 minutos entre solicitudes de reset")
		}
		return nil, fmt.Errorf("error al crear token de reset: %w", err)
	}

	logger.Info("Proceso de reset iniciado para usuario %s desde IP %s (requiere pregunta: %v)", user.Email, requestIP, hasSecurityQuestions)

	// Preparar respuesta
	response := &models.PasswordResetInitResponse{
		Success:                  true,
		RequiresSecurityQuestion: hasSecurityQuestions,
	}

	if hasSecurityQuestions {
		response.Message = "Responda la pregunta de seguridad para continuar"
		// Incluir pregunta de seguridad pero NO el token
		if len(user.SecurityQuestions) > 0 {
			selectedQuestion := user.SecurityQuestions[0] // Por simplicidad, tomar la primera
			response.SecurityQuestion = &models.SecurityQuestionForResetResponse{
				QuestionID:   selectedQuestion.SecurityQuestionID,
				QuestionText: selectedQuestion.SecurityQuestion.QuestionText,
				Attempts:     0,
				MaxAttempts:  models.MaxSecurityQuestionAttempts,
			}
		}
	} else {
		response.Message = "Token de reset enviado a su email institucional"
		// TODO: Aquí enviar email real con el token cuando se implemente el servicio de email
		logger.Info("Token de reset para usuario sin preguntas de seguridad: %s", token)
	}

	return response, nil
}

// VerifySecurityQuestion verifica la respuesta a la pregunta de seguridad
func (s *AuthService) VerifySecurityQuestion(ctx context.Context, req *models.PasswordResetVerifySecurityRequest, requestIP string) (*models.PasswordResetVerifySecurityResponse, error) {
	// Buscar usuario y token por email y pregunta (sin exponer el token aún)
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("SecurityQuestions", "is_active = ?", true).
		Where("email = ? AND is_active = ?", req.Email, true).
		First(&user).Error

	if err != nil {
		return &models.PasswordResetVerifySecurityResponse{
			Success:           false,
			Message:           "Usuario no encontrado o inactivo",
			Verified:          false,
			CanProceedToReset: false,
			AttemptsRemaining: 0,
		}, nil
	}

	// Buscar el token más reciente y activo para este usuario
	var resetToken models.PasswordResetToken
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND requires_security_question = ?",
			user.ID, true, true).
		Order("created_at DESC").
		First(&resetToken).Error

	if err != nil {
		return &models.PasswordResetVerifySecurityResponse{
			Success:           false,
			Message:           "No hay solicitud de reset activa",
			Verified:          false,
			CanProceedToReset: false,
			AttemptsRemaining: 0,
		}, nil
	}

	// Validar token
	if !resetToken.IsValid() || resetToken.HasExceededSecurityAttempts() {
		return &models.PasswordResetVerifySecurityResponse{
			Success:           false,
			Message:           "Solicitud de reset inválida o expirada",
			Verified:          false,
			CanProceedToReset: false,
			AttemptsRemaining: 0,
		}, nil
	}

	// Buscar la pregunta del usuario
	var userSecQuestion models.UserSecurityQuestion
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND security_question_id = ? AND is_active = ?",
			user.ID, req.QuestionID, true).
		First(&userSecQuestion).Error

	if err != nil {
		resetToken.IncrementSecurityAttempts()
		s.db.WithContext(ctx).Save(&resetToken)

		return &models.PasswordResetVerifySecurityResponse{
			Success:           false,
			Message:           "Pregunta de seguridad no válida",
			Verified:          false,
			CanProceedToReset: false,
			AttemptsRemaining: resetToken.GetSecurityAttemptsRemaining(),
		}, nil
	}

	// Verificar respuesta
	providedAnswerHash := s.hashSecurityAnswer(req.Answer)
	if providedAnswerHash != userSecQuestion.AnswerHash {
		resetToken.IncrementSecurityAttempts()
		s.db.WithContext(ctx).Save(&resetToken)

		return &models.PasswordResetVerifySecurityResponse{
			Success:           false,
			Message:           "Respuesta incorrecta",
			Verified:          false,
			CanProceedToReset: false,
			AttemptsRemaining: resetToken.GetSecurityAttemptsRemaining(),
			ResetToken:        "", // NO devolver token si la respuesta es incorrecta
		}, nil
	}

	// Respuesta correcta - marcar como verificada y DEVOLVER EL TOKEN
	resetToken.MarkSecurityQuestionVerified()
	if err := s.db.WithContext(ctx).Save(&resetToken).Error; err != nil {
		return nil, fmt.Errorf("error al marcar pregunta como verificada: %w", err)
	}

	logger.Info("Pregunta de seguridad verificada para usuario %s desde IP %s - Token proporcionado", user.Email, requestIP)

	return &models.PasswordResetVerifySecurityResponse{
		Success:           true,
		Message:           "Pregunta de seguridad verificada correctamente",
		Verified:          true,
		CanProceedToReset: true,
		AttemptsRemaining: resetToken.GetSecurityAttemptsRemaining(),
		ResetToken:        resetToken.Token, // AQUÍ DEVOLVEMOS EL TOKEN
	}, nil
}

// GetPasswordResetStatus obtiene el estado de un usuario durante el proceso de reset
func (s *AuthService) GetPasswordResetStatusByEmail(ctx context.Context, email string) (*models.PasswordResetStatusResponse, error) {
	// Buscar usuario por email
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("SecurityQuestions", "is_active = ?", true).
		Preload("SecurityQuestions.SecurityQuestion").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.PasswordResetStatusResponse{
				TokenValid:               false,
				TokenExpired:             true,
				TokenUsed:                false,
				RequiresSecurityQuestion: false,
				SecurityQuestionVerified: false,
				CanProceedToReset:        false,
				AttemptsRemaining:        0,
			}, nil
		}
		return nil, fmt.Errorf("error al buscar usuario: %w", err)
	}

	// Buscar el token más reciente y activo
	var resetToken models.PasswordResetToken
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", user.ID, true).
		Order("created_at DESC").
		First(&resetToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.PasswordResetStatusResponse{
				TokenValid:               false,
				TokenExpired:             true,
				TokenUsed:                false,
				RequiresSecurityQuestion: false,
				SecurityQuestionVerified: false,
				CanProceedToReset:        false,
				AttemptsRemaining:        0,
			}, nil
		}
		return nil, fmt.Errorf("error al buscar token: %w", err)
	}

	status := resetToken.GetDetailedStatus()

	// Agregar información de pregunta de seguridad si es necesaria y no está verificada
	if resetToken.RequiresSecurityQuestion && !resetToken.SecurityQuestionVerified && len(user.SecurityQuestions) > 0 {
		// Seleccionar pregunta para mostrar
		selectedQuestion := user.SecurityQuestions[0]
		status.SecurityQuestion = &models.SecurityQuestionForResetResponse{
			QuestionID:   selectedQuestion.SecurityQuestionID,
			QuestionText: selectedQuestion.SecurityQuestion.QuestionText,
			Attempts:     resetToken.SecurityQuestionAttempts,
			MaxAttempts:  models.MaxSecurityQuestionAttempts,
		}
	}

	return status, nil
}

// ConfirmPasswordReset confirma y ejecuta el reset de contraseña
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, req *models.PasswordResetConfirm, requestIP string) error {
	// Buscar token válido
	var resetToken models.PasswordResetToken
	err := s.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND is_active = ?", req.Token, true).
		First(&resetToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("token de reset inválido o expirado")
		}
		return fmt.Errorf("error al buscar token: %w", err)
	}

	// Validar que puede proceder al reset
	if !resetToken.CanProceedToPasswordReset() {
		resetToken.IncrementAttempts()
		s.db.WithContext(ctx).Save(&resetToken)

		if resetToken.IsExpired() {
			return fmt.Errorf("token de reset expirado")
		}
		if resetToken.IsUsed() {
			return fmt.Errorf("token de reset ya utilizado")
		}
		if resetToken.RequiresSecurityQuestion && !resetToken.SecurityQuestionVerified {
			return fmt.Errorf("debe verificar la pregunta de seguridad primero")
		}
		if resetToken.HasExceededSecurityAttempts() {
			return fmt.Errorf("demasiados intentos fallidos - token invalidado")
		}
		return fmt.Errorf("token de reset inválido")
	}

	// Verificar pregunta de seguridad si se proporciona en el request
	if req.SecurityQuestionID != nil && req.SecurityAnswer != nil {
		verifyReq := &models.PasswordResetVerifySecurityRequest{
			Email:      resetToken.User.Email,
			QuestionID: *req.SecurityQuestionID,
			Answer:     *req.SecurityAnswer,
		}

		verifyResp, err := s.VerifySecurityQuestion(ctx, verifyReq, requestIP)
		if err != nil {
			return fmt.Errorf("error al verificar pregunta de seguridad: %w", err)
		}

		if !verifyResp.Verified {
			return fmt.Errorf("respuesta de seguridad incorrecta: %s", verifyResp.Message)
		}

		// Recargar token actualizado
		err = s.db.WithContext(ctx).
			Preload("User").
			Where("token = ?", req.Token).
			First(&resetToken).Error
		if err != nil {
			return fmt.Errorf("error al recargar token: %w", err)
		}
	}

	// Validar nueva contraseña
	if isValid, validationErrors := s.passwordService.IsValidPassword(req.NewPassword); !isValid {
		resetToken.IncrementAttempts()
		s.db.WithContext(ctx).Save(&resetToken)
		return fmt.Errorf("nueva contraseña inválida: %v", validationErrors)
	}

	// Hashear nueva contraseña
	newPasswordHash, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("error al hashear nueva contraseña: %w", err)
	}

	// Transacción para actualizar contraseña y marcar token como usado
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Actualizar contraseña del usuario
		if err := tx.Model(&resetToken.User).Updates(map[string]interface{}{
			"password_hash":       newPasswordHash,
			"password_changed_at": time.Now(),
		}).Error; err != nil {
			return fmt.Errorf("error al actualizar contraseña: %w", err)
		}

		// Marcar token como usado
		resetToken.MarkAsUsed()
		if err := tx.Save(&resetToken).Error; err != nil {
			return fmt.Errorf("error al marcar token como usado: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Invalidar todas las sesiones del usuario (logout forzado)
	if err := s.invalidateUserSessions(ctx, resetToken.User.ID.String()); err != nil {
		logger.Warn("Error al invalidar sesiones del usuario %s: %v", resetToken.User.Email, err)
		// No retornar error, el reset fue exitoso
	}

	logger.Info("Contraseña resetiada exitosamente para usuario %s desde IP %s", resetToken.User.Email, requestIP)

	return nil
}

// GetPasswordResetHistory obtiene el historial de tokens de reset de un usuario
func (s *AuthService) GetPasswordResetHistory(ctx context.Context, userID string) ([]models.PasswordResetToken, error) {
	var tokens []models.PasswordResetToken
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(5). // Solo los últimos 5 intentos
		Find(&tokens).Error

	if err != nil {
		return nil, fmt.Errorf("error al obtener tokens de reset: %w", err)
	}

	return tokens, nil
}

// CleanupExpiredResetTokens limpia tokens expirados (para cron job)
func (s *AuthService) CleanupExpiredResetTokens(ctx context.Context) (int, error) {
	// Usar la función SQL ya creada
	var cleanedCount int
	err := s.db.WithContext(ctx).Raw("SELECT cleanup_expired_reset_tokens()").Scan(&cleanedCount).Error
	if err != nil {
		return 0, fmt.Errorf("error al limpiar tokens expirados: %w", err)
	}

	logger.Info("Tokens de reset expirados limpiados: %d", cleanedCount)
	return cleanedCount, nil
}

// ========================================
// FUNCIONES AUXILIARES PRIVADAS
// ========================================

// generateResetToken genera un token único de 64 caracteres
func (s *AuthService) generateResetToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 caracteres hex
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error al generar bytes aleatorios: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// hashSecurityAnswer hashea una respuesta de seguridad con normalización
func (s *AuthService) hashSecurityAnswer(answer string) string {
	// Normalizar respuesta
	normalized := models.NormalizeSecurityAnswer(answer)

	// Generar hash SHA256
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// invalidateUserSessions invalida todas las sesiones de un usuario
func (s *AuthService) invalidateUserSessions(ctx context.Context, userID string) error {
	// Obtener todas las sesiones del usuario
	sessions, err := s.sessionManager.GetUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("error al obtener sesiones: %w", err)
	}

	// Eliminar cada sesión
	for _, sessionID := range sessions {
		if err := s.sessionManager.DeleteSession(ctx, sessionID); err != nil {
			logger.Warn("Error al eliminar sesión %s: %v", sessionID, err)
		}
		if err := s.refreshManager.DeleteRefreshToken(ctx, userID, sessionID); err != nil {
			logger.Warn("Error al eliminar refresh token %s: %v", sessionID, err)
		}
	}

	logger.Info("Sesiones invalidadas para usuario %s: %d sesiones", userID, len(sessions))
	return nil
}

// generateUniqueUsername genera un username único basado en el contexto GAMC
func (s *AuthService) generateUniqueUsername(ctx context.Context, firstName, lastName, email string, orgUnitID int) (string, error) {
	// Obtener código de la unidad organizacional
	var orgUnit models.OrganizationalUnit
	err := s.db.WithContext(ctx).Where("id = ?", orgUnitID).First(&orgUnit).Error
	if err != nil {
		return "", fmt.Errorf("error al obtener unidad organizacional: %w", err)
	}

	// Limpiar nombres (sin espacios, caracteres especiales, acentos)
	cleanFirstName := cleanString(firstName)
	cleanLastName := cleanString(lastName)
	unitCode := cleanString(orgUnit.Code)

	// Estrategia 1: unidad.nombre.apellido
	username := fmt.Sprintf("%s.%s.%s", unitCode, cleanFirstName, cleanLastName)

	// Verificar si existe
	var existingUser models.User
	err = s.db.WithContext(ctx).Where("username = ?", username).First(&existingUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return username, nil // Username disponible
	}

	// Estrategia 2: usar parte del email si es institucional
	if contains(email, "@gamc.gov.bo") {
		emailUsername := strings.Split(email, "@")[0]
		err = s.db.WithContext(ctx).Where("username = ?", emailUsername).First(&existingUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return emailUsername, nil
		}
	}

	// Estrategia 3: agregar números incrementales
	baseUsername := fmt.Sprintf("%s.%s.%s", unitCode, cleanFirstName, cleanLastName)
	for i := 1; i <= 99; i++ {
		testUsername := fmt.Sprintf("%s%d", baseUsername, i)
		err = s.db.WithContext(ctx).Where("username = ?", testUsername).First(&existingUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return testUsername, nil
		}
	}

	// Fallback: usar timestamp
	return fmt.Sprintf("%s_%d", baseUsername, time.Now().Unix()), nil
}

// Funciones auxiliares
func cleanString(s string) string {
	// Implementación simple para limpiar strings
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	// Remover caracteres especiales básicos
	reg := regexp.MustCompile(`[^a-z0-9]`)
	s = reg.ReplaceAllString(s, "")
	return s
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

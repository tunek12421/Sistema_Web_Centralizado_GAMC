// internal/database/models/password_reset.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PasswordResetToken representa un token de reset de contraseña
type PasswordResetToken struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uuid.UUID  `json:"userId" gorm:"type:uuid;not null;index"`
	Token     string     `json:"-" gorm:"uniqueIndex;size:255;not null"` // Nunca exponer en JSON
	ExpiresAt time.Time  `json:"expiresAt" gorm:"not null;index"`
	CreatedAt time.Time  `json:"createdAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	RequestIP string     `json:"requestIp" gorm:"size:45;not null"`
	UserAgent *string    `json:"userAgent,omitempty" gorm:"type:text"`
	IsActive  bool       `json:"isActive" gorm:"default:true;index"`

	// NUEVOS CAMPOS PARA PREGUNTAS DE SEGURIDAD
	RequiresSecurityQuestion bool `json:"requiresSecurityQuestion" gorm:"default:false"`
	SecurityQuestionVerified bool `json:"securityQuestionVerified" gorm:"default:false"`
	SecurityQuestionAttempts int  `json:"securityQuestionAttempts" gorm:"default:0"`

	// Metadatos para auditoría
	EmailSentAt   *time.Time `json:"emailSentAt,omitempty"`
	EmailOpenedAt *time.Time `json:"emailOpenedAt,omitempty"`
	AttemptsCount int        `json:"attemptsCount" gorm:"default:0"`

	// Relaciones
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// PasswordResetRequest representa una solicitud de reset
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordResetConfirm representa la confirmación del reset
type PasswordResetConfirm struct {
	Token              string  `json:"token" validate:"required,min=64,max=64"`
	NewPassword        string  `json:"newPassword" validate:"required,min=8"`
	SecurityQuestionID *int    `json:"securityQuestionId,omitempty"`
	SecurityAnswer     *string `json:"securityAnswer,omitempty"`
}

// NUEVAS ESTRUCTURAS PARA PREGUNTAS DE SEGURIDAD EN RESET

// PasswordResetInitResponse respuesta inicial del reset
type PasswordResetInitResponse struct {
	Success                  bool                              `json:"success"`
	Message                  string                            `json:"message"`
	RequiresSecurityQuestion bool                              `json:"requiresSecurityQuestion"`
	SecurityQuestion         *SecurityQuestionForResetResponse `json:"securityQuestion,omitempty"`
}

// PasswordResetStatusResponse estado del proceso de reset
type PasswordResetStatusResponse struct {
	TokenValid               bool                              `json:"tokenValid"`
	TokenExpired             bool                              `json:"tokenExpired"`
	TokenUsed                bool                              `json:"tokenUsed"`
	RequiresSecurityQuestion bool                              `json:"requiresSecurityQuestion"`
	SecurityQuestionVerified bool                              `json:"securityQuestionVerified"`
	SecurityQuestion         *SecurityQuestionForResetResponse `json:"securityQuestion,omitempty"`
	CanProceedToReset        bool                              `json:"canProceedToReset"`
	AttemptsRemaining        int                               `json:"attemptsRemaining"`
}

// PasswordResetVerifySecurityRequest para verificar pregunta de seguridad
type PasswordResetVerifySecurityRequest struct {
	Email      string `json:"email" validate:"required,email"`
	QuestionID int    `json:"questionId" validate:"required,min=1"`
	Answer     string `json:"answer" validate:"required,min=1,max=100"`
}

// PasswordResetVerifySecurityResponse respuesta de verificación
type PasswordResetVerifySecurityResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	Verified          bool   `json:"verified"`
	CanProceedToReset bool   `json:"canProceedToReset"`
	AttemptsRemaining int    `json:"attemptsRemaining"`
	ResetToken        string `json:"resetToken,omitempty"` // NUEVO: Solo se devuelve si la verificación es exitosa
}

// TableName especifica el nombre de la tabla
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// ===== MÉTODOS DE VALIDACIÓN =====

// IsValid verifica si el token es válido para uso
func (prt *PasswordResetToken) IsValid() bool {
	// Debe estar activo
	if !prt.IsActive {
		return false
	}

	// No debe haber expirado
	if time.Now().After(prt.ExpiresAt) {
		return false
	}

	// No debe haber sido usado
	if prt.UsedAt != nil {
		return false
	}

	return true
}

// IsExpired verifica si el token ha expirado
func (prt *PasswordResetToken) IsExpired() bool {
	return time.Now().After(prt.ExpiresAt)
}

// IsUsed verifica si el token ya fue utilizado
func (prt *PasswordResetToken) IsUsed() bool {
	return prt.UsedAt != nil
}

// CanProceedToPasswordReset verifica si puede proceder al reset final
func (prt *PasswordResetToken) CanProceedToPasswordReset() bool {
	if !prt.IsValid() {
		return false
	}

	// Si requiere pregunta de seguridad, debe estar verificada
	if prt.RequiresSecurityQuestion && !prt.SecurityQuestionVerified {
		return false
	}

	return true
}

// HasExceededSecurityAttempts verifica si excedió intentos de seguridad
func (prt *PasswordResetToken) HasExceededSecurityAttempts() bool {
	return prt.SecurityQuestionAttempts >= MaxSecurityQuestionAttempts
}

// GetSecurityAttemptsRemaining obtiene intentos restantes de pregunta de seguridad
func (prt *PasswordResetToken) GetSecurityAttemptsRemaining() int {
	remaining := MaxSecurityQuestionAttempts - prt.SecurityQuestionAttempts
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ===== MÉTODOS DE ACTUALIZACIÓN =====

// MarkAsUsed marca el token como utilizado
func (prt *PasswordResetToken) MarkAsUsed() {
	now := time.Now()
	prt.UsedAt = &now
	prt.IsActive = false
}

// MarkAsExpired marca el token como expirado
func (prt *PasswordResetToken) MarkAsExpired() {
	prt.IsActive = false
}

// IncrementAttempts incrementa el contador de intentos
func (prt *PasswordResetToken) IncrementAttempts() {
	prt.AttemptsCount++
}

// IncrementSecurityAttempts incrementa intentos de pregunta de seguridad
func (prt *PasswordResetToken) IncrementSecurityAttempts() {
	prt.SecurityQuestionAttempts++

	// Si excede los intentos máximos, invalidar token
	if prt.HasExceededSecurityAttempts() {
		prt.MarkAsExpired()
	}
}

// MarkSecurityQuestionVerified marca la pregunta de seguridad como verificada
func (prt *PasswordResetToken) MarkSecurityQuestionVerified() {
	prt.SecurityQuestionVerified = true
}

// MarkEmailSent marca cuando se envió el email
func (prt *PasswordResetToken) MarkEmailSent() {
	now := time.Now()
	prt.EmailSentAt = &now
}

// MarkEmailOpened marca cuando se abrió el email
func (prt *PasswordResetToken) MarkEmailOpened() {
	now := time.Now()
	prt.EmailOpenedAt = &now
}

// ===== MÉTODOS DE ESTADO =====

// GetStatus retorna el estado actual del token
func (prt *PasswordResetToken) GetStatus() string {
	if prt.IsUsed() {
		return "USED"
	}
	if prt.IsExpired() {
		return "EXPIRED"
	}
	if prt.HasExceededSecurityAttempts() {
		return "SECURITY_ATTEMPTS_EXCEEDED"
	}
	if prt.IsActive {
		return "ACTIVE"
	}
	return "INACTIVE"
}

// GetDetailedStatus retorna estado detallado incluyendo pregunta de seguridad
func (prt *PasswordResetToken) GetDetailedStatus() *PasswordResetStatusResponse {
	return &PasswordResetStatusResponse{
		TokenValid:               prt.IsValid(),
		TokenExpired:             prt.IsExpired(),
		TokenUsed:                prt.IsUsed(),
		RequiresSecurityQuestion: prt.RequiresSecurityQuestion,
		SecurityQuestionVerified: prt.SecurityQuestionVerified,
		CanProceedToReset:        prt.CanProceedToPasswordReset(),
		AttemptsRemaining:        prt.GetSecurityAttemptsRemaining(),
	}
}

// IsSecurityVerificationRequired verifica si necesita verificación de seguridad
func (prt *PasswordResetToken) IsSecurityVerificationRequired() bool {
	return prt.RequiresSecurityQuestion && !prt.SecurityQuestionVerified
}

// ===== MÉTODOS DE TIEMPO =====

// GetTimeRemaining obtiene tiempo restante antes de expiración
func (prt *PasswordResetToken) GetTimeRemaining() time.Duration {
	if prt.IsExpired() {
		return 0
	}
	return time.Until(prt.ExpiresAt)
}

// GetMinutesRemaining obtiene minutos restantes
func (prt *PasswordResetToken) GetMinutesRemaining() int {
	remaining := prt.GetTimeRemaining()
	return int(remaining.Minutes())
}

// ExtendExpiration extiende la expiración del token
func (prt *PasswordResetToken) ExtendExpiration(duration time.Duration) {
	prt.ExpiresAt = prt.ExpiresAt.Add(duration)
}

// ===== HOOKS GORM =====

// BeforeCreate hook que se ejecuta antes de crear el registro
func (prt *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	// Establecer expiración por defecto si no se especificó
	if prt.ExpiresAt.IsZero() {
		prt.ExpiresAt = time.Now().Add(30 * time.Minute)
	}

	// Asegurar que está activo por defecto
	prt.IsActive = true

	return nil
}

// BeforeUpdate hook que se ejecuta antes de actualizar
func (prt *PasswordResetToken) BeforeUpdate(tx *gorm.DB) error {
	// Si se marca como usado, también marcar como inactivo
	if prt.UsedAt != nil {
		prt.IsActive = false
	}

	return nil
}

// ===== CONSTANTES PARA CONFIGURACIÓN =====

const (
	// DefaultResetTokenExpiration tiempo por defecto de expiración
	DefaultResetTokenExpiration = 30 * time.Minute

	// MaxResetTokenExpiration tiempo máximo de expiración
	MaxResetTokenExpiration = 2 * time.Hour

	// MinResetTokenExpiration tiempo mínimo de expiración
	MinResetTokenExpiration = 5 * time.Minute

	// MaxPasswordResetAttempts intentos máximos de reset
	MaxPasswordResetAttempts = 5

	// ResetTokenCooldown tiempo de espera entre solicitudes
	ResetTokenCooldown = 5 * time.Minute
)

// ===== FUNCIONES DE UTILIDAD =====

// CalculateExpiration calcula tiempo de expiración basado en configuración
func CalculateExpiration(customDuration *time.Duration) time.Time {
	if customDuration != nil {
		return time.Now().Add(*customDuration)
	}
	return time.Now().Add(DefaultResetTokenExpiration)
}

// IsValidResetDuration verifica si una duración es válida para reset
func IsValidResetDuration(duration time.Duration) bool {
	return duration >= MinResetTokenExpiration && duration <= MaxResetTokenExpiration
}

// ===== MÉTODOS DE BÚSQUEDA Y FILTRADO =====

// FindValidTokenByUser busca token válido por usuario
func FindValidTokenByUser(db *gorm.DB, userID uuid.UUID) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := db.Where("user_id = ? AND is_active = ? AND expires_at > ?",
		userID, true, time.Now()).
		Order("created_at DESC").
		First(&token).Error
	return &token, err
}

// FindTokenByValue busca token por su valor
func FindTokenByValue(db *gorm.DB, tokenValue string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := db.Preload("User").
		Where("token = ?", tokenValue).
		First(&token).Error
	return &token, err
}

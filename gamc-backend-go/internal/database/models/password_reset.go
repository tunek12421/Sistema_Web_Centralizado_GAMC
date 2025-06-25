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
	Token       string `json:"token" validate:"required,min=64,max=64"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
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

// MarkEmailSent marca cuando se envió el email
func (prt *PasswordResetToken) MarkEmailSent() {
	now := time.Now()
	prt.EmailSentAt = &now
}

// GetStatus retorna el estado actual del token
func (prt *PasswordResetToken) GetStatus() string {
	if prt.IsUsed() {
		return "USED"
	}
	if prt.IsExpired() {
		return "EXPIRED"
	}
	if prt.IsActive {
		return "ACTIVE"
	}
	return "INACTIVE"
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

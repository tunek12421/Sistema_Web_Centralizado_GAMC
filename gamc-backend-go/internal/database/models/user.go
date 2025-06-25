// internal/database/models/user.go - ACTUALIZACIÓN COMPLETA
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User representa un usuario del sistema
type User struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username             string     `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email                string     `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash         string     `json:"-" gorm:"size:255;not null"`
	FirstName            string     `json:"firstName" gorm:"size:50;not null"`
	LastName             string     `json:"lastName" gorm:"size:50;not null"`
	Role                 string     `json:"role" gorm:"size:20;not null;check:role IN ('admin','input','output')"`
	OrganizationalUnitID *int       `json:"organizationalUnitId" gorm:"index"`
	IsActive             bool       `json:"isActive" gorm:"default:true"`
	LastLogin            *time.Time `json:"lastLogin"`
	PasswordChangedAt    time.Time  `json:"passwordChangedAt" gorm:"default:CURRENT_TIMESTAMP"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`

	// Relaciones
	OrganizationalUnit  *OrganizationalUnit    `json:"organizationalUnit,omitempty" gorm:"foreignKey:OrganizationalUnitID"`
	SecurityQuestions   []UserSecurityQuestion `json:"securityQuestions,omitempty" gorm:"foreignKey:UserID"`
	PasswordResetTokens []PasswordResetToken   `json:"-" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook de GORM para generar UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// UserProfile representa el perfil público del usuario
type UserProfile struct {
	ID                   uuid.UUID           `json:"id"`
	Email                string              `json:"email"`
	FirstName            string              `json:"firstName"`
	LastName             string              `json:"lastName"`
	Role                 string              `json:"role"`
	OrganizationalUnitID *int                `json:"organizationalUnitId"`
	OrganizationalUnit   *OrganizationalUnit `json:"organizationalUnit,omitempty"`
	IsActive             bool                `json:"isActive"`
	LastLogin            *time.Time          `json:"lastLogin"`
	CreatedAt            time.Time           `json:"createdAt"`

	// NUEVO: Estado de preguntas de seguridad
	HasSecurityQuestions   bool `json:"hasSecurityQuestions"`
	SecurityQuestionsCount int  `json:"securityQuestionsCount"`
}

// UserProfileExtended perfil extendido con información de seguridad
type UserProfileExtended struct {
	*UserProfile
	SecurityStatus *UserSecurityStatusResponse `json:"securityStatus,omitempty"`
}

// ToProfile convierte User a UserProfile (sin datos sensibles)
func (u *User) ToProfile() *UserProfile {
	profile := &UserProfile{
		ID:                   u.ID,
		Email:                u.Email,
		FirstName:            u.FirstName,
		LastName:             u.LastName,
		Role:                 u.Role,
		OrganizationalUnitID: u.OrganizationalUnitID,
		OrganizationalUnit:   u.OrganizationalUnit,
		IsActive:             u.IsActive,
		LastLogin:            u.LastLogin,
		CreatedAt:            u.CreatedAt,
	}

	// Contar preguntas de seguridad activas
	activeQuestions := 0
	for _, sq := range u.SecurityQuestions {
		if sq.IsActive {
			activeQuestions++
		}
	}

	profile.HasSecurityQuestions = activeQuestions > 0
	profile.SecurityQuestionsCount = activeQuestions

	return profile
}

// ToProfileExtended convierte User a UserProfileExtended con información completa de seguridad
func (u *User) ToProfileExtended() *UserProfileExtended {
	baseProfile := u.ToProfile()

	// Construir estado de seguridad
	securityStatus := &UserSecurityStatusResponse{
		HasSecurityQuestions: baseProfile.HasSecurityQuestions,
		QuestionsCount:       baseProfile.SecurityQuestionsCount,
		MaxQuestions:         MaxSecurityQuestionsPerUser,
		Questions:            make([]UserSecurityQuestionResponse, 0),
	}

	// Convertir preguntas de seguridad a respuestas
	for _, sq := range u.SecurityQuestions {
		if sq.IsActive {
			securityStatus.Questions = append(securityStatus.Questions, *sq.ToResponse())
		}
	}

	return &UserProfileExtended{
		UserProfile:    baseProfile,
		SecurityStatus: securityStatus,
	}
}

// ===== MÉTODOS DE VALIDACIÓN Y UTILIDAD =====

// IsActiveUser verifica si el usuario está activo
func (u *User) IsActiveUser() bool {
	return u.IsActive
}

// HasRole verifica si el usuario tiene un rol específico
func (u *User) HasRole(role string) bool {
	return u.Role == role
}

// IsAdmin verifica si el usuario es administrador
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// CanCreateMessages verifica si el usuario puede crear mensajes
func (u *User) CanCreateMessages() bool {
	return u.Role == "input" || u.Role == "admin"
}

// CanReceiveMessages verifica si el usuario puede recibir mensajes
func (u *User) CanReceiveMessages() bool {
	return u.Role == "output" || u.Role == "admin"
}

// GetFullName retorna el nombre completo del usuario
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// HasSecurityQuestionsConfigured verifica si el usuario tiene preguntas de seguridad
func (u *User) HasSecurityQuestionsConfigured() bool {
	for _, sq := range u.SecurityQuestions {
		if sq.IsActive {
			return true
		}
	}
	return false
}

// GetActiveSecurityQuestions retorna las preguntas de seguridad activas
func (u *User) GetActiveSecurityQuestions() []UserSecurityQuestion {
	var activeQuestions []UserSecurityQuestion
	for _, sq := range u.SecurityQuestions {
		if sq.IsActive {
			activeQuestions = append(activeQuestions, sq)
		}
	}
	return activeQuestions
}

// GetSecurityQuestionsCount retorna el número de preguntas de seguridad activas
func (u *User) GetSecurityQuestionsCount() int {
	count := 0
	for _, sq := range u.SecurityQuestions {
		if sq.IsActive {
			count++
		}
	}
	return count
}

// CanAddMoreSecurityQuestions verifica si puede agregar más preguntas
func (u *User) CanAddMoreSecurityQuestions() bool {
	return u.GetSecurityQuestionsCount() < MaxSecurityQuestionsPerUser
}

// HasSecurityQuestion verifica si el usuario tiene una pregunta específica
func (u *User) HasSecurityQuestion(questionID int) bool {
	for _, sq := range u.SecurityQuestions {
		if sq.SecurityQuestionID == questionID && sq.IsActive {
			return true
		}
	}
	return false
}

// GetSecurityQuestion obtiene una pregunta de seguridad específica
func (u *User) GetSecurityQuestion(questionID int) *UserSecurityQuestion {
	for _, sq := range u.SecurityQuestions {
		if sq.SecurityQuestionID == questionID && sq.IsActive {
			return &sq
		}
	}
	return nil
}

// UpdateLastLogin actualiza el timestamp del último login
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
}

// UpdatePasswordChangedAt actualiza el timestamp del cambio de contraseña
func (u *User) UpdatePasswordChangedAt() {
	u.PasswordChangedAt = time.Now()
}

// IsInstitutionalEmail verifica si el email es institucional
func (u *User) IsInstitutionalEmail() bool {
	return len(u.Email) > 12 && u.Email[len(u.Email)-12:] == "@gamc.gov.bo"
}

// ===== CONSTANTES PARA ROLES =====

const (
	RoleAdmin  = "admin"
	RoleInput  = "input"
	RoleOutput = "output"
)

// GetAvailableRoles retorna los roles disponibles
func GetAvailableRoles() []string {
	return []string{RoleAdmin, RoleInput, RoleOutput}
}

// IsValidRole verifica si un rol es válido
func IsValidRole(role string) bool {
	validRoles := GetAvailableRoles()
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// ===== ESTRUCTURAS PARA REQUESTS =====

// UserCreateRequest para crear un nuevo usuario
type UserCreateRequest struct {
	Username             string                        `json:"username" validate:"required,min=3,max=50"`
	Email                string                        `json:"email" validate:"required,email"`
	Password             string                        `json:"password" validate:"required,min=8"`
	FirstName            string                        `json:"firstName" validate:"required,min=2,max=50"`
	LastName             string                        `json:"lastName" validate:"required,min=2,max=50"`
	Role                 string                        `json:"role" validate:"required,oneof=admin input output"`
	OrganizationalUnitID int                           `json:"organizationalUnitId" validate:"required,min=1"`
	SecurityQuestions    *SecurityQuestionSetupRequest `json:"securityQuestions,omitempty"`
}

// UserUpdateRequest para actualizar un usuario
type UserUpdateRequest struct {
	FirstName            *string `json:"firstName,omitempty" validate:"omitempty,min=2,max=50"`
	LastName             *string `json:"lastName,omitempty" validate:"omitempty,min=2,max=50"`
	Role                 *string `json:"role,omitempty" validate:"omitempty,oneof=admin input output"`
	OrganizationalUnitID *int    `json:"organizationalUnitId,omitempty" validate:"omitempty,min=1"`
	IsActive             *bool   `json:"isActive,omitempty"`
}

// UserChangePasswordRequest para cambiar contraseña
type UserChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// ===== MÉTODOS DE CONVERSIÓN =====

// ToUserCreateRequest convierte User a UserCreateRequest (utilidad para testing)
func (u *User) ToUserCreateRequest() *UserCreateRequest {
	return &UserCreateRequest{
		Username:             u.Username,
		Email:                u.Email,
		FirstName:            u.FirstName,
		LastName:             u.LastName,
		Role:                 u.Role,
		OrganizationalUnitID: *u.OrganizationalUnitID,
	}
}

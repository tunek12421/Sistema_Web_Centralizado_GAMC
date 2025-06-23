// internal/database/models/user.go
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
	OrganizationalUnit *OrganizationalUnit `json:"organizationalUnit,omitempty" gorm:"foreignKey:OrganizationalUnitID"`
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
	ID                 uuid.UUID           `json:"id"`
	Email              string              `json:"email"`
	FirstName          string              `json:"firstName"`
	LastName           string              `json:"lastName"`
	Role               string              `json:"role"`
	OrganizationalUnit *OrganizationalUnit `json:"organizationalUnit,omitempty"`
	IsActive           bool                `json:"isActive"`
	LastLogin          *time.Time          `json:"lastLogin"`
	CreatedAt          time.Time           `json:"createdAt"`
}

// ToProfile convierte User a UserProfile (sin datos sensibles)
func (u *User) ToProfile() *UserProfile {
	return &UserProfile{
		ID:                 u.ID,
		Email:              u.Email,
		FirstName:          u.FirstName,
		LastName:           u.LastName,
		Role:               u.Role,
		OrganizationalUnit: u.OrganizationalUnit,
		IsActive:           u.IsActive,
		LastLogin:          u.LastLogin,
		CreatedAt:          u.CreatedAt,
	}
}

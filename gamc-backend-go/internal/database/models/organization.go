// internal/database/models/organization.go
package models

import "time"

// OrganizationalUnit representa una unidad organizacional del GAMC
type OrganizationalUnit struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description *string   `json:"description,omitempty"`
	ManagerName *string   `json:"managerName,omitempty" gorm:"size:100"`
	Email       *string   `json:"email,omitempty" gorm:"size:100"`
	Phone       *string   `json:"phone,omitempty" gorm:"size:20"`
	IsActive    bool      `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Relaciones
	Users []User `json:"users,omitempty" gorm:"foreignKey:OrganizationalUnitID"`
}

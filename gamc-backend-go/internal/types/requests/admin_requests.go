// internal/types/requests/admin_requests.go
package requests

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest estructura para crear usuario (admin)
type CreateUserRequest struct {
	Username             string     `json:"username" binding:"required,min=3,max=50,alphanum"`
	Email                string     `json:"email" binding:"required,email"`
	Password             string     `json:"password" binding:"required,min=8"`
	FirstName            string     `json:"firstName" binding:"required,min=2,max=50"`
	LastName             string     `json:"lastName" binding:"required,min=2,max=50"`
	Role                 string     `json:"role" binding:"required,oneof=admin input output"`
	OrganizationalUnitID int        `json:"organizationalUnitId" binding:"required,min=1"`
	Position             string     `json:"position,omitempty"`
	PhoneNumber          string     `json:"phoneNumber,omitempty"`
	IsActive             bool       `json:"isActive"`
	MustChangePassword   bool       `json:"mustChangePassword"`
	ExpiresAt            *time.Time `json:"expiresAt,omitempty"`
}

// UpdateUserRequest estructura para actualizar usuario
type UpdateUserRequest struct {
	Email                string     `json:"email,omitempty" binding:"omitempty,email"`
	FirstName            string     `json:"firstName,omitempty" binding:"omitempty,min=2,max=50"`
	LastName             string     `json:"lastName,omitempty" binding:"omitempty,min=2,max=50"`
	Role                 string     `json:"role,omitempty" binding:"omitempty,oneof=admin input output"`
	OrganizationalUnitID *int       `json:"organizationalUnitId,omitempty" binding:"omitempty,min=1"`
	Position             string     `json:"position,omitempty"`
	PhoneNumber          string     `json:"phoneNumber,omitempty"`
	IsActive             *bool      `json:"isActive,omitempty"`
	MustChangePassword   *bool      `json:"mustChangePassword,omitempty"`
	ExpiresAt            *time.Time `json:"expiresAt,omitempty"`
}

// UserFilterRequest filtros para listar usuarios
type UserFilterRequest struct {
	Username       string `form:"username,omitempty"`
	Email          string `form:"email,omitempty"`
	Role           string `form:"role,omitempty"`
	OrganizationID int    `form:"organizationId,omitempty"`
	IsActive       *bool  `form:"isActive,omitempty"`
	Search         string `form:"search,omitempty"`
	CreatedFrom    string `form:"createdFrom,omitempty"`
	CreatedTo      string `form:"createdTo,omitempty"`
	SortBy         string `form:"sortBy,omitempty" binding:"omitempty,oneof=username email created_at last_login"`
	SortOrder      string `form:"sortOrder,omitempty" binding:"omitempty,oneof=asc desc"`
	Page           int    `form:"page,omitempty" binding:"omitempty,min=1"`
	Limit          int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
}

// ResetUserPasswordRequest resetear contraseña de usuario
type ResetUserPasswordRequest struct {
	UserID      uuid.UUID `json:"userId" binding:"required"`
	NewPassword string    `json:"newPassword" binding:"required,min=8"`
	Notify      bool      `json:"notify"`
}

// SystemConfigRequest configuración del sistema
type SystemConfigRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description,omitempty"`
	IsSecret    bool   `json:"isSecret"`
}

// OrganizationalUnitRequest crear/actualizar unidad organizacional
type OrganizationalUnitRequest struct {
	Name        string     `json:"name" binding:"required,min=3,max=100"`
	Code        string     `json:"code" binding:"required,min=2,max=50"`
	Description string     `json:"description,omitempty"`
	ManagerID   *uuid.UUID `json:"managerId,omitempty"`
	ParentID    *int       `json:"parentId,omitempty"`
	IsActive    bool       `json:"isActive"`
}

// AuditLogFilterRequest filtros para logs de auditoría
type AuditLogFilterRequest struct {
	UserID     *uuid.UUID `form:"userId,omitempty"`
	Action     string     `form:"action,omitempty"`
	Resource   string     `form:"resource,omitempty"`
	ResourceID string     `form:"resourceId,omitempty"`
	Result     string     `form:"result,omitempty"`
	IPAddress  string     `form:"ipAddress,omitempty"`
	DateFrom   string     `form:"dateFrom,omitempty"`
	DateTo     string     `form:"dateTo,omitempty"`
	SortBy     string     `form:"sortBy,omitempty" binding:"omitempty,oneof=created_at action resource"`
	SortOrder  string     `form:"sortOrder,omitempty" binding:"omitempty,oneof=asc desc"`
	Page       int        `form:"page,omitempty" binding:"omitempty,min=1"`
	Limit      int        `form:"limit,omitempty" binding:"omitempty,min=1,max=1000"`
}

// SystemStatsRequest estadísticas del sistema
type SystemStatsRequest struct {
	Period  string   `form:"period,omitempty" binding:"omitempty,oneof=today week month quarter year"`
	GroupBy string   `form:"groupBy,omitempty" binding:"omitempty,oneof=hour day week month"`
	Metrics []string `form:"metrics,omitempty"`
}

// BackupRequest solicitud de backup
type BackupRequest struct {
	Type        string   `json:"type" binding:"required,oneof=full incremental differential"`
	Components  []string `json:"components" binding:"required,min=1"` // database, files, config
	Compress    bool     `json:"compress"`
	Encrypt     bool     `json:"encrypt"`
	Destination string   `json:"destination,omitempty"` // local, s3, ftp
}

// MaintenanceRequest solicitud de mantenimiento
type MaintenanceRequest struct {
	Type             string    `json:"type" binding:"required,oneof=scheduled emergency"`
	StartTime        time.Time `json:"startTime" binding:"required"`
	EndTime          time.Time `json:"endTime" binding:"required"`
	Description      string    `json:"description" binding:"required"`
	AffectedServices []string  `json:"affectedServices,omitempty"`
	NotifyUsers      bool      `json:"notifyUsers"`
}

// BulkUserActionRequest acciones masivas en usuarios
type BulkUserActionRequest struct {
	UserIDs []uuid.UUID            `json:"userIds" binding:"required,min=1"`
	Action  string                 `json:"action" binding:"required,oneof=activate deactivate delete reset_password"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// SecurityPolicyRequest política de seguridad
type SecurityPolicyRequest struct {
	MinPasswordLength      int      `json:"minPasswordLength" binding:"min=8,max=32"`
	RequireUppercase       bool     `json:"requireUppercase"`
	RequireLowercase       bool     `json:"requireLowercase"`
	RequireNumbers         bool     `json:"requireNumbers"`
	RequireSpecialChars    bool     `json:"requireSpecialChars"`
	PasswordExpirationDays int      `json:"passwordExpirationDays" binding:"min=0,max=365"`
	MaxLoginAttempts       int      `json:"maxLoginAttempts" binding:"min=3,max=10"`
	LockoutDurationMinutes int      `json:"lockoutDurationMinutes" binding:"min=5,max=1440"`
	SessionTimeoutMinutes  int      `json:"sessionTimeoutMinutes" binding:"min=5,max=480"`
	TwoFactorRequired      bool     `json:"twoFactorRequired"`
	IPWhitelist            []string `json:"ipWhitelist,omitempty"`
}

// NotificationTemplateRequest plantilla de notificación
type NotificationTemplateRequest struct {
	Code      string                 `json:"code" binding:"required,min=3,max=50"`
	Name      string                 `json:"name" binding:"required"`
	Subject   string                 `json:"subject,omitempty"`
	Content   string                 `json:"content" binding:"required"`
	Type      string                 `json:"type" binding:"required,oneof=email sms push in_app"`
	Variables []string               `json:"variables,omitempty"`
	IsActive  bool                   `json:"isActive"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SystemHealthCheckRequest verificación de salud
type SystemHealthCheckRequest struct {
	Components []string `json:"components,omitempty"` // database, redis, minio, api
	Deep       bool     `json:"deep"`
	Timeout    int      `json:"timeout,omitempty" binding:"omitempty,min=1,max=60"`
}

// ImportDataRequest importación de datos
type ImportDataRequest struct {
	Type    string        `json:"type" binding:"required,oneof=users organizations messages"`
	Format  string        `json:"format" binding:"required,oneof=csv json excel"`
	FileID  uuid.UUID     `json:"fileId" binding:"required"`
	Options ImportOptions `json:"options,omitempty"`
}

// ImportOptions opciones de importación
type ImportOptions struct {
	SkipExisting   bool                   `json:"skipExisting"`
	UpdateExisting bool                   `json:"updateExisting"`
	ValidateOnly   bool                   `json:"validateOnly"`
	FieldMappings  map[string]string      `json:"fieldMappings,omitempty"`
	DefaultValues  map[string]interface{} `json:"defaultValues,omitempty"`
}

// ExportDataRequest exportación de datos
type ExportDataRequest struct {
	Type    string        `json:"type" binding:"required,oneof=users organizations messages audit_logs"`
	Format  string        `json:"format" binding:"required,oneof=csv json excel pdf"`
	Filters interface{}   `json:"filters,omitempty"`
	Fields  []string      `json:"fields,omitempty"`
	Options ExportOptions `json:"options,omitempty"`
}

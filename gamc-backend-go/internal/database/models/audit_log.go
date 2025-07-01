// internal/database/models/audit_log.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditAction define las acciones que se pueden auditar
type AuditAction string

const (
	AuditActionCreate  AuditAction = "CREATE"
	AuditActionRead    AuditAction = "READ"
	AuditActionUpdate  AuditAction = "UPDATE"
	AuditActionDelete  AuditAction = "DELETE"
	AuditActionLogin   AuditAction = "LOGIN"
	AuditActionLogout  AuditAction = "LOGOUT"
	AuditActionExport  AuditAction = "EXPORT"
	AuditActionImport  AuditAction = "IMPORT"
	AuditActionApprove AuditAction = "APPROVE"
	AuditActionReject  AuditAction = "REJECT"
	AuditActionSend    AuditAction = "SEND"
	AuditActionReceive AuditAction = "RECEIVE"
	AuditActionArchive AuditAction = "ARCHIVE"
	AuditActionRestore AuditAction = "RESTORE"
)

// AuditResult define los resultados de una acción auditada
type AuditResult string

const (
	AuditResultSuccess AuditResult = "success"
	AuditResultFailure AuditResult = "failure"
	AuditResultPartial AuditResult = "partial"
)

// AuditLog representa un registro de auditoría
type AuditLog struct {
	ID         int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     *uuid.UUID             `json:"userId,omitempty" gorm:"type:uuid;index"`
	Action     AuditAction            `json:"action" gorm:"type:varchar(50);not null;index"`
	Resource   string                 `json:"resource" gorm:"size:100;not null;index"`
	ResourceID string                 `json:"resourceId,omitempty" gorm:"size:100;index"`
	OldValues  map[string]interface{} `json:"oldValues,omitempty" gorm:"type:jsonb"`
	NewValues  map[string]interface{} `json:"newValues,omitempty" gorm:"type:jsonb"`
	IPAddress  string                 `json:"ipAddress,omitempty" gorm:"type:inet"`
	UserAgent  string                 `json:"userAgent,omitempty" gorm:"type:text"`
	Result     AuditResult            `json:"result" gorm:"type:varchar(20);default:'success'"`
	ErrorMsg   string                 `json:"errorMsg,omitempty" gorm:"type:text"`
	Duration   int                    `json:"duration,omitempty"` // milliseconds
	SessionID  string                 `json:"sessionId,omitempty" gorm:"size:100;index"`
	CreatedAt  time.Time              `json:"createdAt"`

	// Relaciones
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName especifica el nombre de la tabla
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate hook para establecer valores por defecto
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.Result == "" {
		a.Result = AuditResultSuccess
	}
	return nil
}

// GetChangedFields retorna los campos que cambiaron
func (a *AuditLog) GetChangedFields() []string {
	if a.OldValues == nil || a.NewValues == nil {
		return []string{}
	}

	changedFields := []string{}
	for key, newValue := range a.NewValues {
		if oldValue, exists := a.OldValues[key]; !exists || oldValue != newValue {
			changedFields = append(changedFields, key)
		}
	}
	return changedFields
}

// HasSensitiveData verifica si el log contiene datos sensibles
func (a *AuditLog) HasSensitiveData() bool {
	sensitiveResources := []string{"users", "passwords", "security_questions", "tokens"}
	for _, resource := range sensitiveResources {
		if a.Resource == resource {
			return true
		}
	}
	return false
}

// AuditLogSummary estructura para resúmenes de auditoría
type AuditLogSummary struct {
	Date         time.Time `json:"date"`
	TotalActions int64     `json:"totalActions"`
	SuccessCount int64     `json:"successCount"`
	FailureCount int64     `json:"failureCount"`
	UniqueUsers  int64     `json:"uniqueUsers"`
	TopActions   []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	} `json:"topActions"`
	TopResources []struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	} `json:"topResources"`
}

// AuditLogFilter estructura para filtrar logs de auditoría
type AuditLogFilter struct {
	UserID     *uuid.UUID  `json:"userId,omitempty"`
	Action     AuditAction `json:"action,omitempty"`
	Resource   string      `json:"resource,omitempty"`
	ResourceID string      `json:"resourceId,omitempty"`
	Result     AuditResult `json:"result,omitempty"`
	IPAddress  string      `json:"ipAddress,omitempty"`
	DateFrom   *time.Time  `json:"dateFrom,omitempty"`
	DateTo     *time.Time  `json:"dateTo,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
}

// NewAuditLog crea un nuevo log de auditoría
func NewAuditLog(userID *uuid.UUID, action AuditAction, resource string, resourceID string) *AuditLog {
	return &AuditLog{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Result:     AuditResultSuccess,
		CreatedAt:  time.Now(),
	}
}

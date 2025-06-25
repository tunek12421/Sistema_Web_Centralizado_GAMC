// internal/database/models/report.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportType define los tipos de reportes
type ReportType string

const (
	ReportTypeDaily     ReportType = "daily"
	ReportTypeWeekly    ReportType = "weekly"
	ReportTypeMonthly   ReportType = "monthly"
	ReportTypeQuarterly ReportType = "quarterly"
	ReportTypeAnnual    ReportType = "annual"
	ReportTypeCustom    ReportType = "custom"
	ReportTypeOnDemand  ReportType = "on_demand"
)

// ReportStatus define los estados de un reporte
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusGenerating ReportStatus = "generating"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusScheduled  ReportStatus = "scheduled"
	ReportStatusCancelled  ReportStatus = "cancelled"
)

// ReportFormat define los formatos de salida
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatExcel ReportFormat = "excel"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatHTML  ReportFormat = "html"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatXML   ReportFormat = "xml"
)

// Report representa un reporte del sistema
type Report struct {
	ID                  uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name                string                 `json:"name" gorm:"size:255;not null"`
	Description         string                 `json:"description,omitempty" gorm:"type:text"`
	Type                ReportType             `json:"type" gorm:"type:varchar(20);not null;index"`
	Status              ReportStatus           `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Format              ReportFormat           `json:"format" gorm:"type:varchar(10);not null"`
	RequestedBy         uuid.UUID              `json:"requestedBy" gorm:"type:uuid;not null;index"`
	OrganizationID      int                    `json:"organizationId" gorm:"not null;index"`
	TemplateID          string                 `json:"templateId,omitempty" gorm:"size:100"`
	Parameters          map[string]interface{} `json:"parameters,omitempty" gorm:"type:jsonb"`
	DateRangeStart      *time.Time             `json:"dateRangeStart,omitempty"`
	DateRangeEnd        *time.Time             `json:"dateRangeEnd,omitempty"`
	ScheduledAt         *time.Time             `json:"scheduledAt,omitempty"`
	GenerationStarted   *time.Time             `json:"generationStarted,omitempty"`
	GenerationCompleted *time.Time             `json:"generationCompleted,omitempty"`
	FileID              *uuid.UUID             `json:"fileId,omitempty" gorm:"type:uuid"`
	FileSize            int64                  `json:"fileSize,omitempty"`
	RecordCount         int64                  `json:"recordCount,omitempty"`
	ErrorMessage        string                 `json:"errorMessage,omitempty" gorm:"type:text"`
	ProcessingTime      int                    `json:"processingTime,omitempty"` // seconds
	NotifyOnComplete    bool                   `json:"notifyOnComplete" gorm:"default:true"`
	IsPublic            bool                   `json:"isPublic" gorm:"default:false"`
	Tags                []string               `json:"tags,omitempty" gorm:"type:text[]"`
	ExpiresAt           *time.Time             `json:"expiresAt,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`

	// Relaciones
	Requester    *User               `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
	Organization *OrganizationalUnit `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	File         *FileMetadata       `json:"file,omitempty" gorm:"foreignKey:FileID"`
}

// TableName especifica el nombre de la tabla
func (Report) TableName() string {
	return "reports"
}

// BeforeCreate hook para establecer valores por defecto
func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = ReportStatusPending
	}
	return nil
}

// StartGeneration marca el inicio de la generación
func (r *Report) StartGeneration(db *gorm.DB) error {
	now := time.Now()
	r.GenerationStarted = &now
	r.Status = ReportStatusGenerating
	return db.Save(r).Error
}

// CompleteGeneration marca la finalización exitosa
func (r *Report) CompleteGeneration(db *gorm.DB, fileID uuid.UUID, fileSize int64, recordCount int64) error {
	now := time.Now()
	r.GenerationCompleted = &now
	r.Status = ReportStatusCompleted
	r.FileID = &fileID
	r.FileSize = fileSize
	r.RecordCount = recordCount

	if r.GenerationStarted != nil {
		r.ProcessingTime = int(now.Sub(*r.GenerationStarted).Seconds())
	}

	return db.Save(r).Error
}

// FailGeneration marca la generación como fallida
func (r *Report) FailGeneration(db *gorm.DB, errorMsg string) error {
	now := time.Now()
	r.GenerationCompleted = &now
	r.Status = ReportStatusFailed
	r.ErrorMessage = errorMsg

	if r.GenerationStarted != nil {
		r.ProcessingTime = int(now.Sub(*r.GenerationStarted).Seconds())
	}

	return db.Save(r).Error
}

// IsExpired verifica si el reporte ha expirado
func (r *Report) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// ReportTemplate define plantillas de reportes reutilizables
type ReportTemplate struct {
	ID               string                 `json:"id" gorm:"primaryKey;size:100"`
	Name             string                 `json:"name" gorm:"size:255;not null"`
	Description      string                 `json:"description" gorm:"type:text"`
	Category         string                 `json:"category" gorm:"size:50;index"`
	Type             ReportType             `json:"type" gorm:"type:varchar(20);not null"`
	DefaultFormat    ReportFormat           `json:"defaultFormat" gorm:"type:varchar(10);not null"`
	AvailableFormats []ReportFormat         `json:"availableFormats" gorm:"type:text[]"`
	Parameters       map[string]interface{} `json:"parameters" gorm:"type:jsonb"` // Definición de parámetros
	Query            string                 `json:"query,omitempty" gorm:"type:text"`
	IsActive         bool                   `json:"isActive" gorm:"default:true"`
	RequiredRole     string                 `json:"requiredRole,omitempty" gorm:"size:50"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
}

// TableName especifica el nombre de la tabla
func (ReportTemplate) TableName() string {
	return "report_templates"
}

// ReportSchedule define programación de reportes
type ReportSchedule struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TemplateID     string                 `json:"templateId" gorm:"size:100;not null;index"`
	UserID         uuid.UUID              `json:"userId" gorm:"type:uuid;not null;index"`
	OrganizationID int                    `json:"organizationId" gorm:"not null;index"`
	Name           string                 `json:"name" gorm:"size:255;not null"`
	CronExpression string                 `json:"cronExpression" gorm:"size:100;not null"`
	Format         ReportFormat           `json:"format" gorm:"type:varchar(10);not null"`
	Parameters     map[string]interface{} `json:"parameters,omitempty" gorm:"type:jsonb"`
	IsActive       bool                   `json:"isActive" gorm:"default:true"`
	LastRunAt      *time.Time             `json:"lastRunAt,omitempty"`
	NextRunAt      *time.Time             `json:"nextRunAt,omitempty"`
	Recipients     []string               `json:"recipients,omitempty" gorm:"type:text[]"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`

	// Relaciones
	Template     *ReportTemplate     `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	User         *User               `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Organization *OrganizationalUnit `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// TableName especifica el nombre de la tabla
func (ReportSchedule) TableName() string {
	return "report_schedules"
}

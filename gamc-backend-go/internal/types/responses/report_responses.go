// internal/types/responses/report_responses.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// ReportResponse respuesta detallada de un reporte
type ReportResponse struct {
	ID                  uuid.UUID              `json:"id"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	Type                string                 `json:"type"`
	Status              string                 `json:"status"`
	Format              string                 `json:"format"`
	Requester           UserSummary            `json:"requester"`
	Organization        OrganizationSummary    `json:"organization"`
	TemplateID          string                 `json:"templateId,omitempty"`
	TemplateName        string                 `json:"templateName,omitempty"`
	Parameters          map[string]interface{} `json:"parameters,omitempty"`
	DateRangeStart      *time.Time             `json:"dateRangeStart,omitempty"`
	DateRangeEnd        *time.Time             `json:"dateRangeEnd,omitempty"`
	ScheduledAt         *time.Time             `json:"scheduledAt,omitempty"`
	GenerationStarted   *time.Time             `json:"generationStarted,omitempty"`
	GenerationCompleted *time.Time             `json:"generationCompleted,omitempty"`
	ProcessingTime      int                    `json:"processingTime,omitempty"`
	FileID              *uuid.UUID             `json:"fileId,omitempty"`
	FileSize            int64                  `json:"fileSize,omitempty"`
	RecordCount         int64                  `json:"recordCount,omitempty"`
	DownloadUrl         string                 `json:"downloadUrl,omitempty"`
	PreviewUrl          string                 `json:"previewUrl,omitempty"`
	ErrorMessage        string                 `json:"errorMessage,omitempty"`
	Tags                []string               `json:"tags,omitempty"`
	ExpiresAt           *time.Time             `json:"expiresAt,omitempty"`
	CreatedAt           time.Time              `json:"createdAt"`
	UpdatedAt           time.Time              `json:"updatedAt"`
}

// ReportListResponse lista de reportes
type ReportListResponse struct {
	Reports    []ReportResponse `json:"reports"`
	Pagination PaginationInfo   `json:"pagination"`
}

// ReportGenerationResponse respuesta de generación
type ReportGenerationResponse struct {
	ReportID      uuid.UUID `json:"reportId"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	EstimatedTime int       `json:"estimatedTime,omitempty"` // segundos
	QueuePosition int       `json:"queuePosition,omitempty"`
}

// ReportTemplateResponse plantilla de reporte
type ReportTemplateResponse struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	Description      string                    `json:"description,omitempty"`
	Category         string                    `json:"category,omitempty"`
	Type             string                    `json:"type"`
	DefaultFormat    string                    `json:"defaultFormat"`
	AvailableFormats []string                  `json:"availableFormats"`
	Parameters       []ReportParameterResponse `json:"parameters"`
	RequiredRole     string                    `json:"requiredRole,omitempty"`
	IsActive         bool                      `json:"isActive"`
	UsageCount       int64                     `json:"usageCount"`
	LastUsedAt       *time.Time                `json:"lastUsedAt,omitempty"`
	CreatedAt        time.Time                 `json:"createdAt"`
	UpdatedAt        time.Time                 `json:"updatedAt"`
}

// ReportParameterResponse parámetro de reporte
type ReportParameterResponse struct {
	Name         string              `json:"name"`
	Label        string              `json:"label"`
	Type         string              `json:"type"` // string, number, date, boolean, select
	Required     bool                `json:"required"`
	DefaultValue interface{}         `json:"defaultValue,omitempty"`
	Options      []ParameterOption   `json:"options,omitempty"`
	Validation   ParameterValidation `json:"validation,omitempty"`
}

// ParameterOption opción de parámetro
type ParameterOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ParameterValidation validación de parámetro
type ParameterValidation struct {
	Min     interface{} `json:"min,omitempty"`
	Max     interface{} `json:"max,omitempty"`
	Pattern string      `json:"pattern,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ReportScheduleResponse programación de reporte
type ReportScheduleResponse struct {
	ID             uuid.UUID              `json:"id"`
	TemplateID     string                 `json:"templateId"`
	TemplateName   string                 `json:"templateName"`
	User           UserSummary            `json:"user"`
	Organization   OrganizationSummary    `json:"organization"`
	Name           string                 `json:"name"`
	CronExpression string                 `json:"cronExpression"`
	HumanSchedule  string                 `json:"humanSchedule"` // "Every Monday at 9:00 AM"
	Format         string                 `json:"format"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Recipients     []string               `json:"recipients,omitempty"`
	IsActive       bool                   `json:"isActive"`
	LastRunAt      *time.Time             `json:"lastRunAt,omitempty"`
	LastRunStatus  string                 `json:"lastRunStatus,omitempty"`
	NextRunAt      *time.Time             `json:"nextRunAt,omitempty"`
	RunCount       int64                  `json:"runCount"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// ReportStatsResponse estadísticas de reportes
type ReportStatsResponse struct {
	TotalReports          int64                `json:"totalReports"`
	GeneratedToday        int64                `json:"generatedToday"`
	GeneratedThisWeek     int64                `json:"generatedThisWeek"`
	GeneratedThisMonth    int64                `json:"generatedThisMonth"`
	PendingReports        int64                `json:"pendingReports"`
	FailedReports         int64                `json:"failedReports"`
	AverageProcessingTime float64              `json:"averageProcessingTime"` // segundos
	ReportsByType         []ReportTypeStats    `json:"reportsByType"`
	ReportsByFormat       []ReportFormatStats  `json:"reportsByFormat"`
	TopTemplates          []TemplateUsageStats `json:"topTemplates"`
	GenerationTrend       []GenerationTrend    `json:"generationTrend"`
}

// ReportTypeStats estadísticas por tipo
type ReportTypeStats struct {
	Type       string  `json:"type"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// ReportFormatStats estadísticas por formato
type ReportFormatStats struct {
	Format     string  `json:"format"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// TemplateUsageStats uso de plantillas
type TemplateUsageStats struct {
	TemplateID   string    `json:"templateId"`
	TemplateName string    `json:"templateName"`
	UsageCount   int64     `json:"usageCount"`
	LastUsedAt   time.Time `json:"lastUsedAt"`
}

// GenerationTrend tendencia de generación
type GenerationTrend struct {
	Date        time.Time `json:"date"`
	Generated   int64     `json:"generated"`
	Failed      int64     `json:"failed"`
	AverageTime float64   `json:"averageTime"`
}

// ReportQueueResponse cola de reportes
type ReportQueueResponse struct {
	QueuedReports []QueuedReport `json:"queuedReports"`
	TotalQueued   int64          `json:"totalQueued"`
	EstimatedWait int            `json:"estimatedWait"` // segundos
}

// QueuedReport reporte en cola
type QueuedReport struct {
	ReportID      uuid.UUID   `json:"reportId"`
	Name          string      `json:"name"`
	Requester     UserSummary `json:"requester"`
	Position      int         `json:"position"`
	EstimatedTime int         `json:"estimatedTime"`
	Priority      string      `json:"priority"`
	QueuedAt      time.Time   `json:"queuedAt"`
}

// ReportExportResponse exportación de reporte
type ReportExportResponse struct {
	ExportID    uuid.UUID `json:"exportId"`
	Format      string    `json:"format"`
	FileSize    int64     `json:"fileSize"`
	DownloadUrl string    `json:"downloadUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

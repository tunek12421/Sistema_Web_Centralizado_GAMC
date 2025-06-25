// internal/types/requests/report_requests.go
package requests

import (
	"time"

	"github.com/google/uuid"
)

// CreateReportRequest estructura para crear un reporte
type CreateReportRequest struct {
	Name             string                 `json:"name" binding:"required,min=3,max=255"`
	Description      string                 `json:"description,omitempty"`
	Type             string                 `json:"type" binding:"required,oneof=daily weekly monthly quarterly annual custom on_demand"`
	Format           string                 `json:"format" binding:"required,oneof=pdf excel csv html json xml"`
	TemplateID       string                 `json:"templateId,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	DateRangeStart   *time.Time             `json:"dateRangeStart,omitempty"`
	DateRangeEnd     *time.Time             `json:"dateRangeEnd,omitempty"`
	ScheduledAt      *time.Time             `json:"scheduledAt,omitempty"`
	NotifyOnComplete bool                   `json:"notifyOnComplete"`
	Recipients       []string               `json:"recipients,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
}

// ReportFilterRequest estructura para filtrar reportes
type ReportFilterRequest struct {
	Name           string    `form:"name,omitempty"`
	Type           string    `form:"type,omitempty"`
	Status         string    `form:"status,omitempty"`
	Format         string    `form:"format,omitempty"`
	RequestedBy    uuid.UUID `form:"requestedBy,omitempty"`
	OrganizationID int       `form:"organizationId,omitempty"`
	TemplateID     string    `form:"templateId,omitempty"`
	DateFrom       string    `form:"dateFrom,omitempty"`
	DateTo         string    `form:"dateTo,omitempty"`
	Tags           []string  `form:"tags,omitempty"`
	SortBy         string    `form:"sortBy,omitempty" binding:"omitempty,oneof=created_at scheduled_at name type status"`
	SortOrder      string    `form:"sortOrder,omitempty" binding:"omitempty,oneof=asc desc"`
	Page           int       `form:"page,omitempty" binding:"omitempty,min=1"`
	Limit          int       `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
}

// GenerateReportRequest solicitud para generar reporte inmediato
type GenerateReportRequest struct {
	TemplateID     string                 `json:"templateId" binding:"required"`
	Format         string                 `json:"format" binding:"required,oneof=pdf excel csv html json xml"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	DateRangeStart *time.Time             `json:"dateRangeStart,omitempty"`
	DateRangeEnd   *time.Time             `json:"dateRangeEnd,omitempty"`
}

// ScheduleReportRequest estructura para programar reportes
type ScheduleReportRequest struct {
	TemplateID     string                 `json:"templateId" binding:"required"`
	Name           string                 `json:"name" binding:"required,min=3,max=255"`
	CronExpression string                 `json:"cronExpression" binding:"required"`
	Format         string                 `json:"format" binding:"required,oneof=pdf excel csv html json xml"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Recipients     []string               `json:"recipients,omitempty"`
	IsActive       bool                   `json:"isActive"`
}

// UpdateScheduleRequest actualizar programación
type UpdateScheduleRequest struct {
	Name           string                 `json:"name,omitempty" binding:"omitempty,min=3,max=255"`
	CronExpression string                 `json:"cronExpression,omitempty"`
	Format         string                 `json:"format,omitempty" binding:"omitempty,oneof=pdf excel csv html json xml"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	Recipients     []string               `json:"recipients,omitempty"`
	IsActive       *bool                  `json:"isActive,omitempty"`
}

// ReportTemplateRequest crear/actualizar plantilla
type ReportTemplateRequest struct {
	ID               string                 `json:"id" binding:"required,min=3,max=100"`
	Name             string                 `json:"name" binding:"required,min=3,max=255"`
	Description      string                 `json:"description,omitempty"`
	Category         string                 `json:"category,omitempty"`
	Type             string                 `json:"type" binding:"required,oneof=daily weekly monthly quarterly annual custom on_demand"`
	DefaultFormat    string                 `json:"defaultFormat" binding:"required,oneof=pdf excel csv html json xml"`
	AvailableFormats []string               `json:"availableFormats" binding:"required,min=1"`
	Parameters       map[string]interface{} `json:"parameters"`
	Query            string                 `json:"query,omitempty"`
	RequiredRole     string                 `json:"requiredRole,omitempty"`
}

// ExportReportRequest exportar datos de reporte
type ExportReportRequest struct {
	ReportID uuid.UUID     `json:"reportId" binding:"required"`
	Format   string        `json:"format" binding:"required,oneof=pdf excel csv html json xml"`
	Options  ExportOptions `json:"options,omitempty"`
}

// ExportOptions opciones de exportación
type ExportOptions struct {
	IncludeHeaders  bool     `json:"includeHeaders"`
	IncludeFooters  bool     `json:"includeFooters"`
	IncludeSummary  bool     `json:"includeSummary"`
	IncludeCharts   bool     `json:"includeCharts"`
	DateFormat      string   `json:"dateFormat,omitempty"`
	NumberFormat    string   `json:"numberFormat,omitempty"`
	PageOrientation string   `json:"pageOrientation,omitempty" binding:"omitempty,oneof=portrait landscape"`
	PageSize        string   `json:"pageSize,omitempty" binding:"omitempty,oneof=A4 Letter Legal"`
	Columns         []string `json:"columns,omitempty"`
}

// ReportStatisticsRequest estadísticas de reportes
type ReportStatisticsRequest struct {
	OrganizationID int    `form:"organizationId,omitempty"`
	DateFrom       string `form:"dateFrom,omitempty"`
	DateTo         string `form:"dateTo,omitempty"`
	GroupBy        string `form:"groupBy,omitempty" binding:"omitempty,oneof=day week month type status template"`
}

// BulkReportActionRequest acciones masivas en reportes
type BulkReportActionRequest struct {
	ReportIDs []uuid.UUID `json:"reportIds" binding:"required,min=1"`
	Action    string      `json:"action" binding:"required,oneof=delete archive cancel regenerate"`
}

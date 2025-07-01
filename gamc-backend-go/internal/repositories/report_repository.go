// internal/repositories/report_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportRepository maneja las operaciones de base de datos para reportes
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository crea una nueva instancia del repositorio de reportes
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// CreateTemplate crea una nueva plantilla de reporte
func (r *ReportRepository) CreateTemplate(ctx context.Context, template *models.ReportTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// CreateReport crea un nuevo reporte generado
func (r *ReportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

// GetTemplateByID obtiene una plantilla de reporte por ID
func (r *ReportRepository) GetTemplateByID(ctx context.Context, id uuid.UUID) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := r.db.WithContext(ctx).
		Preload("CreatedBy").
		Where("id = ?", id).
		First(&template).Error

	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetTemplateByCode obtiene una plantilla por código
func (r *ReportRepository) GetTemplateByCode(ctx context.Context, code string) (*models.ReportTemplate, error) {
	var template models.ReportTemplate
	err := r.db.WithContext(ctx).
		Where("code = ? AND is_active = ?", code, true).
		First(&template).Error

	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetReportByID obtiene un reporte por ID
func (r *ReportRepository) GetReportByID(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	var report models.Report
	err := r.db.WithContext(ctx).
		Preload("Template").
		Preload("GeneratedBy").
		Preload("FileMetadata").
		Where("id = ?", id).
		First(&report).Error

	if err != nil {
		return nil, err
	}
	return &report, nil
}

// UpdateTemplate actualiza una plantilla de reporte
func (r *ReportRepository) UpdateTemplate(ctx context.Context, template *models.ReportTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// UpdateReport actualiza un reporte
func (r *ReportRepository) UpdateReport(ctx context.Context, report *models.Report) error {
	return r.db.WithContext(ctx).Save(report).Error
}

// GetActiveTemplates obtiene todas las plantillas activas
func (r *ReportRepository) GetActiveTemplates(ctx context.Context) ([]*models.ReportTemplate, error) {
	var templates []*models.ReportTemplate
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("category, name").
		Find(&templates).Error
	return templates, err
}

// GetTemplatesByCategory obtiene plantillas por categoría
func (r *ReportRepository) GetTemplatesByCategory(ctx context.Context, category string) ([]*models.ReportTemplate, error) {
	var templates []*models.ReportTemplate
	err := r.db.WithContext(ctx).
		Where("category = ? AND is_active = ?", category, true).
		Order("name").
		Find(&templates).Error
	return templates, err
}

// GetReportsByFilter obtiene reportes con filtros y paginación
func (r *ReportRepository) GetReportsByFilter(ctx context.Context, filter *ReportFilter) ([]*models.Report, int64, error) {
	var reports []*models.Report
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Report{})

	// Aplicar filtros
	query = r.applyReportFilters(query, filter)

	// Contar total antes de paginar
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenamiento
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortDesc {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("generated_at DESC")
	}

	// Aplicar paginación
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Precargar relaciones
	query = query.
		Preload("Template").
		Preload("GeneratedBy")

	// Ejecutar consulta
	if err := query.Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

// GetReportsByUser obtiene reportes generados por un usuario
func (r *ReportRepository) GetReportsByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Report, error) {
	var reports []*models.Report
	query := r.db.WithContext(ctx).
		Preload("Template").
		Where("generated_by = ?", userID).
		Order("generated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&reports).Error
	return reports, err
}

// GetScheduledReports obtiene reportes programados pendientes
func (r *ReportRepository) GetScheduledReports(ctx context.Context) ([]*models.Report, error) {
	var reports []*models.Report
	now := time.Now()

	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("status = ? AND scheduled_for <= ? AND scheduled_for IS NOT NULL",
			models.ReportStatusScheduled, now).
		Find(&reports).Error

	return reports, err
}

// UpdateReportStatus actualiza el estado de un reporte
func (r *ReportRepository) UpdateReportStatus(ctx context.Context, id uuid.UUID, status models.ReportStatus, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if status == models.ReportStatusFailed && errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	if status == models.ReportStatusCompleted {
		updates["completed_at"] = time.Now()
	}

	return r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// GetReportStats obtiene estadísticas de reportes
func (r *ReportRepository) GetReportStats(ctx context.Context, startDate, endDate time.Time) (*ReportStats, error) {
	stats := &ReportStats{
		ByStatus:   make(map[models.ReportStatus]int64),
		ByCategory: make(map[string]int64),
		ByFormat:   make(map[models.ReportFormat]int64),
	}

	// Total de reportes
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Where("generated_at BETWEEN ? AND ?", startDate, endDate).
		Count(&stats.Total)

	// Por estado
	var statusCounts []struct {
		Status models.ReportStatus
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("status, COUNT(*) as count").
		Where("generated_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Por categoría (a través de templates)
	var categoryCounts []struct {
		Category string
		Count    int64
	}
	r.db.WithContext(ctx).
		Table("reports r").
		Joins("JOIN report_templates t ON r.template_id = t.id").
		Select("t.category, COUNT(*) as count").
		Where("r.generated_at BETWEEN ? AND ?", startDate, endDate).
		Group("t.category").
		Scan(&categoryCounts)

	for _, cc := range categoryCounts {
		stats.ByCategory[cc.Category] = cc.Count
	}

	// Por formato
	var formatCounts []struct {
		Format models.ReportFormat
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.Report{}).
		Select("format, COUNT(*) as count").
		Where("generated_at BETWEEN ? AND ?", startDate, endDate).
		Group("format").
		Scan(&formatCounts)

	for _, fc := range formatCounts {
		stats.ByFormat[fc.Format] = fc.Count
	}

	// Tiempo promedio de generación
	var avgTime float64
	r.db.WithContext(ctx).
		Table("reports").
		Select("AVG(generation_duration_ms)").
		Where("generated_at BETWEEN ? AND ? AND generation_duration_ms > 0", startDate, endDate).
		Scan(&avgTime)
	stats.AvgGenerationTime = avgTime

	// Total de tamaño
	r.db.WithContext(ctx).
		Table("reports r").
		Joins("JOIN file_metadata f ON r.file_metadata_id = f.id").
		Select("COALESCE(SUM(f.file_size), 0)").
		Where("r.generated_at BETWEEN ? AND ?", startDate, endDate).
		Scan(&stats.TotalSize)

	return stats, nil
}

// GetMostGeneratedTemplates obtiene las plantillas más utilizadas
func (r *ReportRepository) GetMostGeneratedTemplates(ctx context.Context, limit int) ([]*TemplateUsageStats, error) {
	var stats []*TemplateUsageStats

	err := r.db.WithContext(ctx).
		Table("reports r").
		Joins("JOIN report_templates t ON r.template_id = t.id").
		Select("t.id, t.code, t.name, COUNT(*) as usage_count, AVG(r.generation_duration_ms) as avg_duration").
		Group("t.id, t.code, t.name").
		Order("usage_count DESC").
		Limit(limit).
		Scan(&stats).Error

	return stats, err
}

// DeleteOldReports elimina reportes antiguos
func (r *ReportRepository) DeleteOldReports(ctx context.Context, days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	result := r.db.WithContext(ctx).
		Where("generated_at < ? AND status = ?", cutoffDate, models.ReportStatusCompleted).
		Delete(&models.Report{})
	return result.RowsAffected, result.Error
}

// GetDashboardMetrics obtiene métricas predefinidas del dashboard
func (r *ReportRepository) GetDashboardMetrics(ctx context.Context, metricType string) ([]*models.DashboardMetric, error) {
	var metrics []*models.DashboardMetric
	err := r.db.WithContext(ctx).
		Where("metric_type = ? AND is_active = ?", metricType, true).
		Order("display_order, created_at").
		Find(&metrics).Error
	return metrics, err
}

// UpdateDashboardMetric actualiza una métrica del dashboard
func (r *ReportRepository) UpdateDashboardMetric(ctx context.Context, metric *models.DashboardMetric) error {
	metric.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(metric).Error
}

// applyReportFilters aplica los filtros a la consulta
func (r *ReportRepository) applyReportFilters(query *gorm.DB, filter *ReportFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por plantilla
	if filter.TemplateID != nil {
		query = query.Where("template_id = ?", *filter.TemplateID)
	}

	// Filtro por usuario generador
	if filter.GeneratedBy != nil {
		query = query.Where("generated_by = ?", *filter.GeneratedBy)
	}

	// Filtro por estado
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Filtro por formato
	if filter.Format != nil {
		query = query.Where("format = ?", *filter.Format)
	}

	// Filtro por rango de fechas
	if filter.GeneratedFrom != nil {
		query = query.Where("generated_at >= ?", *filter.GeneratedFrom)
	}
	if filter.GeneratedTo != nil {
		query = query.Where("generated_at <= ?", *filter.GeneratedTo)
	}

	// Búsqueda por nombre
	if filter.SearchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.SearchTerm)
		query = query.Where("name ILIKE ?", searchPattern)
	}

	return query
}

// ReportFilter estructura para filtrar reportes
type ReportFilter struct {
	TemplateID    *uuid.UUID
	GeneratedBy   *uuid.UUID
	Status        *models.ReportStatus
	Format        *models.ReportFormat
	GeneratedFrom *time.Time
	GeneratedTo   *time.Time
	SearchTerm    string
	SortBy        string
	SortDesc      bool
	Limit         int
	Offset        int
}

// ReportStats estadísticas de reportes
type ReportStats struct {
	Total             int64
	ByStatus          map[models.ReportStatus]int64
	ByCategory        map[string]int64
	ByFormat          map[models.ReportFormat]int64
	AvgGenerationTime float64
	TotalSize         int64
}

// TemplateUsageStats estadísticas de uso de plantillas
type TemplateUsageStats struct {
	ID          uuid.UUID
	Code        string
	Name        string
	UsageCount  int64
	AvgDuration float64
}

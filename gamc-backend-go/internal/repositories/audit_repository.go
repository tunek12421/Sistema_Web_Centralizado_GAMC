// internal/repositories/audit_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditRepository maneja las operaciones de base de datos para logs de auditoría
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository crea una nueva instancia del repositorio de auditoría
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create crea un nuevo log de auditoría
func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// CreateBatch crea múltiples logs de auditoría
func (r *AuditRepository) CreateBatch(ctx context.Context, logs []*models.AuditLog) error {
	return r.db.WithContext(ctx).CreateInBatches(logs, 100).Error
}

// GetByID obtiene un log de auditoría por ID
func (r *AuditRepository) GetByID(ctx context.Context, id int64) (*models.AuditLog, error) {
	var log models.AuditLog
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&log).Error

	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByFilter obtiene logs con filtros y paginación
func (r *AuditRepository) GetByFilter(ctx context.Context, filter *AuditFilter) ([]*models.AuditLog, int64, error) {
	var logs []*models.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	// Aplicar filtros
	query = r.applyAuditFilters(query, filter)

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
		query = query.Order("created_at DESC")
	}

	// Aplicar paginación
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Precargar relaciones
	query = query.Preload("User")

	// Ejecutar consulta
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetByUserID obtiene logs de auditoría de un usuario
func (r *AuditRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetByResource obtiene logs por recurso
func (r *AuditRepository) GetByResource(ctx context.Context, resource string, resourceID string) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).
		Preload("User").
		Where("resource = ?", resource)

	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}

	err := query.Order("created_at DESC").Find(&logs).Error
	return logs, err
}

// GetByAction obtiene logs por acción
func (r *AuditRepository) GetByAction(ctx context.Context, action models.AuditAction, startDate, endDate time.Time) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("action = ? AND created_at BETWEEN ? AND ?", action, startDate, endDate).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetFailedAttempts obtiene intentos fallidos
func (r *AuditRepository) GetFailedAttempts(ctx context.Context, action models.AuditAction, minutes int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	cutoffTime := time.Now().Add(-time.Duration(minutes) * time.Minute)

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("action = ? AND result = ? AND created_at >= ?",
			action, models.AuditResultFailure, cutoffTime).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// GetByIPAddress obtiene logs por dirección IP
func (r *AuditRepository) GetByIPAddress(ctx context.Context, ipAddress string, limit int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	query := r.db.WithContext(ctx).
		Preload("User").
		Where("ip_address = ?", ipAddress).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&logs).Error
	return logs, err
}

// GetSummary obtiene un resumen de auditoría
func (r *AuditRepository) GetSummary(ctx context.Context, startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	summary := &models.AuditLogSummary{
		Date: startDate,
	}

	// Total de acciones
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&summary.TotalActions)

	// Conteo de éxitos y fallos
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("created_at BETWEEN ? AND ? AND result = ?", startDate, endDate, models.AuditResultSuccess).
		Count(&summary.SuccessCount)

	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("created_at BETWEEN ? AND ? AND result = ?", startDate, endDate, models.AuditResultFailure).
		Count(&summary.FailureCount)

	// Usuarios únicos
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("created_at BETWEEN ? AND ? AND user_id IS NOT NULL", startDate, endDate).
		Distinct("user_id").
		Count(&summary.UniqueUsers)

	// Top acciones
	var topActions []struct {
		Action string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("action").
		Order("count DESC").
		Limit(5).
		Scan(&topActions)

	for _, ta := range topActions {
		summary.TopActions = append(summary.TopActions, struct {
			Action string `json:"action"`
			Count  int64  `json:"count"`
		}{Action: ta.Action, Count: ta.Count})
	}

	// Top recursos
	var topResources []struct {
		Resource string
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("resource").
		Order("count DESC").
		Limit(5).
		Scan(&topResources)

	for _, tr := range topResources {
		summary.TopResources = append(summary.TopResources, struct {
			Resource string `json:"resource"`
			Count    int64  `json:"count"`
		}{Resource: tr.Resource, Count: tr.Count})
	}

	return summary, nil
}

// DeleteOldLogs elimina logs antiguos
func (r *AuditRepository) DeleteOldLogs(ctx context.Context, days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	result := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{})
	return result.RowsAffected, result.Error
}

// GetSecurityEvents obtiene eventos de seguridad importantes
func (r *AuditRepository) GetSecurityEvents(ctx context.Context, hours int) ([]*models.AuditLog, error) {
	var logs []*models.AuditLog
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	securityActions := []models.AuditAction{
		models.AuditActionLogin,
		models.AuditActionLogout,
		models.AuditActionDelete,
		models.AuditActionCreate,
		models.AuditActionUpdate,
	}

	err := r.db.WithContext(ctx).
		Preload("User").
		Where("action IN ? AND created_at >= ? AND (result = ? OR resource IN ?)",
			securityActions,
			cutoffTime,
			models.AuditResultFailure,
			[]string{"users", "passwords", "security_questions"}).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, err
}

// GetUserActivity obtiene la actividad completa de un usuario
func (r *AuditRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*UserActivityReport, error) {
	report := &UserActivityReport{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		Actions:   make(map[string]int64),
		Resources: make(map[string]int64),
	}

	// Total de acciones
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Count(&report.TotalActions)

	// Acciones por tipo
	var actionCounts []struct {
		Action string
		Count  int64
	}
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Group("action").
		Scan(&actionCounts)

	for _, ac := range actionCounts {
		report.Actions[ac.Action] = ac.Count
	}

	// Recursos accedidos
	var resourceCounts []struct {
		Resource string
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Group("resource").
		Scan(&resourceCounts)

	for _, rc := range resourceCounts {
		report.Resources[rc.Resource] = rc.Count
	}

	// Último login
	var lastLogin models.AuditLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND action = ?", userID, models.AuditActionLogin).
		Order("created_at DESC").
		First(&lastLogin).Error; err == nil {
		report.LastLogin = &lastLogin.CreatedAt
	}

	// IPs utilizadas
	var ips []string
	r.db.WithContext(ctx).
		Model(&models.AuditLog{}).
		Where("user_id = ? AND ip_address IS NOT NULL", userID).
		Distinct("ip_address").
		Pluck("ip_address", &ips)
	report.IPAddresses = ips

	return report, nil
}

// applyAuditFilters aplica los filtros a la consulta
func (r *AuditRepository) applyAuditFilters(query *gorm.DB, filter *AuditFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por usuario
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	// Filtro por acción
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	// Filtro por recurso
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}

	// Filtro por ID de recurso
	if filter.ResourceID != "" {
		query = query.Where("resource_id = ?", filter.ResourceID)
	}

	// Filtro por resultado
	if filter.Result != "" {
		query = query.Where("result = ?", filter.Result)
	}

	// Filtro por IP
	if filter.IPAddress != "" {
		query = query.Where("ip_address = ?", filter.IPAddress)
	}

	// Filtro por sesión
	if filter.SessionID != "" {
		query = query.Where("session_id = ?", filter.SessionID)
	}

	// Filtro por rango de fechas
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}

	// Búsqueda en valores
	if filter.SearchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.SearchTerm)
		query = query.Where("CAST(old_values AS TEXT) ILIKE ? OR CAST(new_values AS TEXT) ILIKE ? OR error_msg ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	return query
}

// AuditFilter estructura para filtrar logs de auditoría
type AuditFilter struct {
	UserID     *uuid.UUID
	Action     string
	Resource   string
	ResourceID string
	Result     string
	IPAddress  string
	SessionID  string
	DateFrom   *time.Time
	DateTo     *time.Time
	SearchTerm string
	SortBy     string
	SortDesc   bool
	Limit      int
	Offset     int
}

// UserActivityReport reporte de actividad de usuario
type UserActivityReport struct {
	UserID       uuid.UUID
	StartDate    time.Time
	EndDate      time.Time
	TotalActions int64
	Actions      map[string]int64
	Resources    map[string]int64
	LastLogin    *time.Time
	IPAddresses  []string
}

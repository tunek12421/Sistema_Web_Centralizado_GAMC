// internal/repositories/notification_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationRepository maneja las operaciones de base de datos para notificaciones
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository crea una nueva instancia del repositorio de notificaciones
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create crea una nueva notificación
func (r *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// CreateBatch crea múltiples notificaciones
func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []*models.Notification) error {
	return r.db.WithContext(ctx).CreateInBatches(notifications, 100).Error
}

// GetByID obtiene una notificación por ID
func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("RelatedMessage").
		Where("id = ?", id).
		First(&notification).Error

	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetByUserID obtiene notificaciones de un usuario
func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) ([]*models.Notification, int64, error) {
	var notifications []*models.Notification
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Notification{}).Where("user_id = ?", userID)

	// Aplicar filtros
	query = r.applyNotificationFilters(query, filter)

	// Contar total antes de paginar
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenamiento
	query = query.Order("created_at DESC")

	// Aplicar paginación
	if filter != nil {
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	// Precargar relaciones
	query = query.Preload("RelatedMessage")

	// Ejecutar consulta
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

// Update actualiza una notificación
func (r *NotificationRepository) Update(ctx context.Context, notification *models.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete elimina una notificación
func (r *NotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Notification{}, id).Error
}

// MarkAsRead marca una notificación como leída
func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// MarkAllAsRead marca todas las notificaciones de un usuario como leídas
func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// GetUnreadCount obtiene el conteo de notificaciones no leídas
func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetUnreadByType obtiene notificaciones no leídas por tipo
func (r *NotificationRepository) GetUnreadByType(ctx context.Context, userID uuid.UUID, notifType models.NotificationType) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.WithContext(ctx).
		Preload("RelatedMessage").
		Where("user_id = ? AND type = ? AND is_read = ?", userID, notifType, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// GetByDateRange obtiene notificaciones en un rango de fechas
func (r *NotificationRepository) GetByDateRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// DeleteOldNotifications elimina notificaciones antiguas
func (r *NotificationRepository) DeleteOldNotifications(ctx context.Context, days int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	result := r.db.WithContext(ctx).
		Where("created_at < ? AND is_read = ?", cutoffDate, true).
		Delete(&models.Notification{})
	return result.RowsAffected, result.Error
}

// GetNotificationStats obtiene estadísticas de notificaciones
func (r *NotificationRepository) GetNotificationStats(ctx context.Context, userID uuid.UUID) (*NotificationStats, error) {
	stats := &NotificationStats{}

	// Total notificaciones
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.Total)

	// No leídas
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&stats.Unread)

	// Por tipo
	var typeCounts []struct {
		Type  models.NotificationType
		Count int64
	}
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeCounts)

	stats.ByType = make(map[models.NotificationType]int64)
	for _, tc := range typeCounts {
		stats.ByType[tc.Type] = tc.Count
	}

	// Por prioridad
	var priorityCounts []struct {
		Priority models.NotificationPriority
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Select("priority, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("priority").
		Scan(&priorityCounts)

	stats.ByPriority = make(map[models.NotificationPriority]int64)
	for _, pc := range priorityCounts {
		stats.ByPriority[pc.Priority] = pc.Count
	}

	return stats, nil
}

// CreateSystemNotification crea una notificación del sistema para múltiples usuarios
func (r *NotificationRepository) CreateSystemNotification(ctx context.Context, userIDs []uuid.UUID, title, content string, priority models.NotificationPriority) error {
	notifications := make([]*models.Notification, len(userIDs))
	now := time.Now()

	for i, userID := range userIDs {
		notifications[i] = &models.Notification{
			UserID:    userID,
			Type:      models.NotificationTypeSystem,
			Title:     title,
			Content:   content,
			Priority:  priority,
			CreatedAt: now,
		}
	}

	return r.CreateBatch(ctx, notifications)
}

// GetRecentNotifications obtiene las notificaciones más recientes
func (r *NotificationRepository) GetRecentNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]*models.Notification, error) {
	var notifications []*models.Notification
	err := r.db.WithContext(ctx).
		Preload("RelatedMessage").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// BatchMarkAsRead marca múltiples notificaciones como leídas
func (r *NotificationRepository) BatchMarkAsRead(ctx context.Context, ids []uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Notification{}).
		Where("id IN ?", ids).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// BatchDelete elimina múltiples notificaciones
func (r *NotificationRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&models.Notification{}).Error
}

// applyNotificationFilters aplica los filtros a la consulta
func (r *NotificationRepository) applyNotificationFilters(query *gorm.DB, filter *NotificationFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por estado de lectura
	if filter.IsRead != nil {
		query = query.Where("is_read = ?", *filter.IsRead)
	}

	// Filtro por tipo
	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}

	// Filtro por prioridad
	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}

	// Filtro por rango de fechas
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}

	// Búsqueda por texto
	if filter.SearchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.SearchTerm)
		query = query.Where("title ILIKE ? OR content ILIKE ?", searchPattern, searchPattern)
	}

	return query
}

// NotificationFilter estructura para filtrar notificaciones
type NotificationFilter struct {
	IsRead     *bool
	Type       *models.NotificationType
	Priority   *models.NotificationPriority
	DateFrom   *time.Time
	DateTo     *time.Time
	SearchTerm string
	Limit      int
	Offset     int
}

// NotificationStats estadísticas de notificaciones
type NotificationStats struct {
	Total      int64
	Unread     int64
	ByType     map[models.NotificationType]int64
	ByPriority map[models.NotificationPriority]int64
}

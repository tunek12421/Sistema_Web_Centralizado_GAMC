// internal/repositories/message_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageRepository maneja las operaciones de base de datos para mensajes
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository crea una nueva instancia del repositorio de mensajes
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create crea un nuevo mensaje
func (r *MessageRepository) Create(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Create(message).Error
}

// GetByID obtiene un mensaje por ID con sus relaciones
func (r *MessageRepository) GetByID(ctx context.Context, id int64) (*models.Message, error) {
	var message models.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Preload("Attachments").
		Preload("Attachments.Uploader").
		Where("id = ?", id).
		First(&message).Error

	if err != nil {
		return nil, err
	}
	return &message, nil
}

// Update actualiza un mensaje
func (r *MessageRepository) Update(ctx context.Context, message *models.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// Delete elimina un mensaje (soft delete)
func (r *MessageRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Message{}, id).Error
}

// Search busca mensajes usando GetMessagesRequest - MÉTODO CORREGIDO
func (r *MessageRepository) Search(ctx context.Context, filter *GetMessagesRequest) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Message{})

	// Crear filtro compatible
	messageFilter := &MessageFilter{
		UserID:        filter.UserID,
		MessageTypeID: filter.MessageType,
		StatusID:      filter.Status,
		IsUrgent:      filter.IsUrgent,
		DateFrom:      filter.DateFrom,
		DateTo:        filter.DateTo,
		SortBy:        filter.SortBy,
		Limit:         filter.Limit,
		Offset:        (filter.Page - 1) * filter.Limit,
	}

	// Filtro de texto de búsqueda
	if filter.SearchText != nil {
		messageFilter.SearchTerm = *filter.SearchText
	}

	// Filtro de unidad
	if filter.UnitID != nil {
		messageFilter.SenderUnitID = filter.UnitID
	}

	// Configurar orden
	if filter.SortOrder == "desc" {
		messageFilter.SortDesc = true
	}

	// Aplicar filtros
	query = r.applyFilters(query, messageFilter)

	// Contar total antes de paginar
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenamiento
	if messageFilter.SortBy != "" {
		order := messageFilter.SortBy
		if messageFilter.SortDesc {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Aplicar paginación
	if messageFilter.Limit > 0 {
		query = query.Limit(messageFilter.Limit)
	}
	if messageFilter.Offset > 0 {
		query = query.Offset(messageFilter.Offset)
	}

	// Precargar relaciones
	query = query.
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Preload("Attachments")

	// Ejecutar consulta
	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetMessagesRequest necesita ser definido en el repositorio
type GetMessagesRequest struct {
	UnitID      *int       `json:"unitId"`
	UserID      *uuid.UUID `json:"userId"`
	MessageType *int       `json:"messageType"`
	Status      *int       `json:"status"`
	IsUrgent    *bool      `json:"isUrgent"`
	DateFrom    *time.Time `json:"dateFrom"`
	DateTo      *time.Time `json:"dateTo"`
	SearchText  *string    `json:"searchText"`
	Page        int        `json:"page"`
	Limit       int        `json:"limit"`
	SortBy      string     `json:"sortBy"`
	SortOrder   string     `json:"sortOrder"`
}

// GetByFilter obtiene mensajes con filtros y paginación
func (r *MessageRepository) GetByFilter(ctx context.Context, filter *MessageFilter) ([]*models.Message, int64, error) {
	var messages []*models.Message
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Message{})

	// Aplicar filtros
	query = r.applyFilters(query, filter)

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
	query = query.
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Preload("Attachments")

	// Ejecutar consulta
	if err := query.Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetByUserID obtiene mensajes relacionados con un usuario
func (r *MessageRepository) GetByUserID(ctx context.Context, userID uuid.UUID, filter *MessageFilter) ([]*models.Message, int64, error) {
	filter.UserID = &userID
	return r.GetByFilter(ctx, filter)
}

// GetByUnitID obtiene mensajes relacionados con una unidad
func (r *MessageRepository) GetByUnitID(ctx context.Context, unitID int, sent bool) ([]*models.Message, error) {
	var messages []*models.Message
	query := r.db.WithContext(ctx)

	if sent {
		query = query.Where("sender_unit_id = ?", unitID)
	} else {
		query = query.Where("receiver_unit_id = ?", unitID)
	}

	err := query.
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Order("created_at DESC").
		Find(&messages).Error

	return messages, err
}

// GetUnreadCount obtiene el conteo de mensajes no leídos
func (r *MessageRepository) GetUnreadCount(ctx context.Context, unitID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_unit_id = ? AND read_at IS NULL", unitID).
		Count(&count).Error
	return count, err
}

// MarkAsRead marca un mensaje como leído
func (r *MessageRepository) MarkAsRead(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Update("read_at", now).Error
}

// MarkAsResponded marca un mensaje como respondido
func (r *MessageRepository) MarkAsResponded(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Update("responded_at", now).Error
}

// Archive archiva un mensaje
func (r *MessageRepository) Archive(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Update("archived_at", now).Error
}

// Unarchive desarchiva un mensaje
func (r *MessageRepository) Unarchive(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id = ?", id).
		Update("archived_at", nil).Error
}

// GetStatsByUnit obtiene estadísticas de mensajes por unidad
func (r *MessageRepository) GetStatsByUnit(ctx context.Context, unitID int, dateFrom, dateTo time.Time) (*MessageStats, error) {
	stats := &MessageStats{}

	// Total enviados
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("sender_unit_id = ? AND created_at BETWEEN ? AND ?", unitID, dateFrom, dateTo).
		Count(&stats.TotalSent)

	// Total recibidos
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_unit_id = ? AND created_at BETWEEN ? AND ?", unitID, dateFrom, dateTo).
		Count(&stats.TotalReceived)

	// No leídos
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("receiver_unit_id = ? AND read_at IS NULL AND created_at BETWEEN ? AND ?", unitID, dateFrom, dateTo).
		Count(&stats.Unread)

	// Urgentes
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_unit_id = ? OR receiver_unit_id = ?) AND is_urgent = ? AND created_at BETWEEN ? AND ?",
			unitID, unitID, true, dateFrom, dateTo).
		Count(&stats.Urgent)

	// Archivados
	r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("(sender_unit_id = ? OR receiver_unit_id = ?) AND archived_at IS NOT NULL AND created_at BETWEEN ? AND ?",
			unitID, unitID, dateFrom, dateTo).
		Count(&stats.Archived)

	return stats, nil
}

// SearchMessages busca mensajes por texto - MÉTODO ALTERNATIVO
func (r *MessageRepository) SearchMessages(ctx context.Context, searchTerm string, filter *MessageFilter) ([]*models.Message, int64, error) {
	filter.SearchTerm = searchTerm
	return r.GetByFilter(ctx, filter)
}

// GetMessageThread obtiene un hilo de conversación
func (r *MessageRepository) GetMessageThread(ctx context.Context, originalMessageID int64) ([]*models.Message, error) {
	var messages []*models.Message

	// Aquí implementarías la lógica para obtener respuestas a un mensaje
	// Por ahora, retornamos solo el mensaje original
	message, err := r.GetByID(ctx, originalMessageID)
	if err != nil {
		return nil, err
	}

	messages = append(messages, message)
	return messages, nil
}

// BatchUpdateStatus actualiza el estado de múltiples mensajes
func (r *MessageRepository) BatchUpdateStatus(ctx context.Context, messageIDs []int64, statusID int) error {
	return r.db.WithContext(ctx).
		Model(&models.Message{}).
		Where("id IN ?", messageIDs).
		Update("status_id", statusID).Error
}

// GetMessageTypes obtiene todos los tipos de mensaje activos
func (r *MessageRepository) GetMessageTypes(ctx context.Context) ([]*models.MessageType, error) {
	var types []*models.MessageType
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("priority_level DESC, name ASC").
		Find(&types).Error
	return types, err
}

// GetMessageStatuses obtiene todos los estados de mensaje
func (r *MessageRepository) GetMessageStatuses(ctx context.Context) ([]*models.MessageStatus, error) {
	var statuses []*models.MessageStatus
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, name ASC").
		Find(&statuses).Error
	return statuses, err
}

// applyFilters aplica los filtros a la consulta
func (r *MessageRepository) applyFilters(query *gorm.DB, filter *MessageFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por usuario (enviados o recibidos)
	if filter.UserID != nil {
		query = query.Where("sender_id = ? OR receiver_unit_id IN (SELECT organizational_unit_id FROM users WHERE id = ?)",
			*filter.UserID, *filter.UserID)
	}

	// Filtro por unidad emisora
	if filter.SenderUnitID != nil {
		query = query.Where("sender_unit_id = ?", *filter.SenderUnitID)
	}

	// Filtro por unidad receptora
	if filter.ReceiverUnitID != nil {
		query = query.Where("receiver_unit_id = ?", *filter.ReceiverUnitID)
	}

	// Filtro por tipo de mensaje
	if filter.MessageTypeID != nil {
		query = query.Where("message_type_id = ?", *filter.MessageTypeID)
	}

	// Filtro por estado
	if filter.StatusID != nil {
		query = query.Where("status_id = ?", *filter.StatusID)
	}

	// Filtro por nivel de prioridad
	if filter.PriorityLevel != nil {
		query = query.Where("priority_level = ?", *filter.PriorityLevel)
	}

	// Filtro por urgencia
	if filter.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *filter.IsUrgent)
	}

	// Filtro por archivados
	if filter.ShowArchived != nil && !*filter.ShowArchived {
		query = query.Where("archived_at IS NULL")
	}

	// Filtro por no leídos
	if filter.UnreadOnly != nil && *filter.UnreadOnly {
		query = query.Where("read_at IS NULL")
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
		query = query.Where("subject ILIKE ? OR content ILIKE ?", searchPattern, searchPattern)
	}

	return query
}

// MessageFilter estructura para filtrar mensajes
type MessageFilter struct {
	UserID         *uuid.UUID
	SenderUnitID   *int
	ReceiverUnitID *int
	MessageTypeID  *int
	StatusID       *int
	PriorityLevel  *int
	IsUrgent       *bool
	ShowArchived   *bool
	UnreadOnly     *bool
	DateFrom       *time.Time
	DateTo         *time.Time
	SearchTerm     string
	SortBy         string
	SortDesc       bool
	Limit          int
	Offset         int
}

// MessageStats estadísticas de mensajes
type MessageStats struct {
	TotalSent     int64
	TotalReceived int64
	Unread        int64
	Urgent        int64
	Archived      int64
}

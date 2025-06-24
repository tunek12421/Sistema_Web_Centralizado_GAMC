// internal/services/message_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageService maneja la lógica de negocio para mensajes
type MessageService struct {
	db *gorm.DB
}

// NewMessageService crea una nueva instancia del servicio
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{
		db: db,
	}
}

// CreateMessageRequest representa los datos para crear un mensaje
type CreateMessageRequest struct {
	Subject        string    `json:"subject" validate:"required,min=3,max=255"`
	Content        string    `json:"content" validate:"required,min=10"`
	ReceiverUnitID int       `json:"receiverUnitId" validate:"required,gt=0"`
	MessageTypeID  int       `json:"messageTypeId" validate:"required,gt=0"`
	PriorityLevel  int       `json:"priorityLevel" validate:"min=1,max=4"`
	IsUrgent       bool      `json:"isUrgent"`
	SenderID       uuid.UUID `json:"-"` // Se asigna desde el contexto del usuario
	SenderUnitID   int       `json:"-"` // Se asigna desde el usuario
}

// GetMessagesRequest representa los filtros para obtener mensajes
type GetMessagesRequest struct {
	UnitID      *int       `json:"unitId"`
	UserID      *uuid.UUID `json:"userId"`
	MessageType *int       `json:"messageType"`
	Status      *int       `json:"status"`
	IsUrgent    *bool      `json:"isUrgent"`
	DateFrom    *time.Time `json:"dateFrom"`
	DateTo      *time.Time `json:"dateTo"`
	SearchText  *string    `json:"searchText"`
	Page        int        `json:"page" validate:"min=1"`
	Limit       int        `json:"limit" validate:"min=1,max=100"`
	SortBy      string     `json:"sortBy" validate:"oneof=created_at subject priority_level"`
	SortOrder   string     `json:"sortOrder" validate:"oneof=asc desc"`
}

// MessageResponse representa la respuesta de un mensaje
type MessageResponse struct {
	ID             int64      `json:"id"`
	Subject        string     `json:"subject"`
	Content        string     `json:"content"`
	SenderID       uuid.UUID  `json:"senderId"`
	SenderUnitID   int        `json:"senderUnitId"`
	ReceiverUnitID int        `json:"receiverUnitId"`
	MessageTypeID  int        `json:"messageTypeId"`
	StatusID       int        `json:"statusId"`
	PriorityLevel  int        `json:"priorityLevel"`
	IsUrgent       bool       `json:"isUrgent"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
	RespondedAt    *time.Time `json:"respondedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	// Datos relacionales
	Sender       *models.User               `json:"sender,omitempty"`
	SenderUnit   *models.OrganizationalUnit `json:"senderUnit,omitempty"`
	ReceiverUnit *models.OrganizationalUnit `json:"receiverUnit,omitempty"`
	MessageType  *models.MessageType        `json:"messageType,omitempty"`
	Status       *models.MessageStatus      `json:"status,omitempty"`
	Attachments  []models.MessageAttachment `json:"attachments,omitempty"`
}

// CreateMessage crea un nuevo mensaje
func (s *MessageService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*MessageResponse, error) {
	logger.Info("📨 Creando nuevo mensaje: %s", req.Subject)

	// Validar que la unidad receptora existe y está activa
	var receiverUnit models.OrganizationalUnit
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", req.ReceiverUnitID, true).First(&receiverUnit).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unidad organizacional destino no encontrada o inactiva")
		}
		return nil, fmt.Errorf("error al validar unidad destino: %w", err)
	}

	// Validar que el tipo de mensaje existe y está activo
	var messageType models.MessageType
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", req.MessageTypeID, true).First(&messageType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("tipo de mensaje no encontrado o inactivo")
		}
		return nil, fmt.Errorf("error al validar tipo de mensaje: %w", err)
	}

	// Obtener el estado inicial (SENT)
	var sentStatus models.MessageStatus
	if err := s.db.WithContext(ctx).Where("code = ?", "SENT").First(&sentStatus).Error; err != nil {
		return nil, fmt.Errorf("error al obtener estado inicial: %w", err)
	}

	// Crear el mensaje
	message := models.Message{
		Subject:        req.Subject,
		Content:        req.Content,
		SenderID:       req.SenderID,
		SenderUnitID:   req.SenderUnitID,
		ReceiverUnitID: req.ReceiverUnitID,
		MessageTypeID:  req.MessageTypeID,
		StatusID:       sentStatus.ID,
		PriorityLevel:  req.PriorityLevel,
		IsUrgent:       req.IsUrgent,
	}

	// Si no se especifica prioridad, usar la del tipo de mensaje
	if message.PriorityLevel == 0 {
		message.PriorityLevel = messageType.PriorityLevel
	}

	// Crear el mensaje en la base de datos
	if err := s.db.WithContext(ctx).Create(&message).Error; err != nil {
		return nil, fmt.Errorf("error al crear mensaje: %w", err)
	}

	logger.Info("✅ Mensaje creado exitosamente - ID: %d", message.ID)

	// Retornar el mensaje completo con relaciones
	return s.GetMessageByID(ctx, message.ID)
}

// GetMessagesByUnit obtiene mensajes por unidad organizacional
func (s *MessageService) GetMessagesByUnit(ctx context.Context, unitID int, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Debug("📋 Obteniendo mensajes para unidad: %d", unitID)

	query := s.db.WithContext(ctx).Model(&models.Message{})

	// Filtrar por unidad (mensajes enviados o recibidos)
	query = query.Where("sender_unit_id = ? OR receiver_unit_id = ?", unitID, unitID)

	return s.applyFiltersAndGetMessages(ctx, query, req)
}

// GetMessagesByUser obtiene mensajes por usuario específico
func (s *MessageService) GetMessagesByUser(ctx context.Context, userID uuid.UUID, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Debug("📋 Obteniendo mensajes para usuario: %s", userID.String())

	// Primero obtener la unidad del usuario
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, 0, fmt.Errorf("usuario no encontrado: %w", err)
	}

	query := s.db.WithContext(ctx).Model(&models.Message{})

	// Filtrar por la unidad del usuario o mensajes que envió directamente
	userUnitID := 0
	if user.OrganizationalUnitID != nil {
		userUnitID = *user.OrganizationalUnitID
	}

	query = query.Where("sender_id = ? OR sender_unit_id = ? OR receiver_unit_id = ?",
		userID, userUnitID, userUnitID)

	return s.applyFiltersAndGetMessages(ctx, query, req)
}

// GetMessageByID obtiene un mensaje específico por ID
func (s *MessageService) GetMessageByID(ctx context.Context, messageID int64) (*MessageResponse, error) {
	logger.Debug("🔍 Obteniendo mensaje ID: %d", messageID)

	var message models.Message
	err := s.db.WithContext(ctx).
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Preload("Attachments").
		Where("id = ?", messageID).
		First(&message).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mensaje no encontrado")
		}
		return nil, fmt.Errorf("error al obtener mensaje: %w", err)
	}

	return s.convertToResponse(&message), nil
}

// MarkAsRead marca un mensaje como leído
func (s *MessageService) MarkAsRead(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Info("👁️ Marcando mensaje como leído - ID: %d, Usuario: %s", messageID, userID.String())

	// Verificar que el mensaje existe
	var message models.Message
	if err := s.db.WithContext(ctx).Where("id = ?", messageID).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mensaje no encontrado")
		}
		return fmt.Errorf("error al obtener mensaje: %w", err)
	}

	// Verificar que el usuario tiene permisos para leer el mensaje
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Solo puede marcar como leído si es de su unidad o si es admin
	userUnitID := 0
	if user.OrganizationalUnitID != nil {
		userUnitID = *user.OrganizationalUnitID
	}

	if user.Role != "admin" &&
		message.ReceiverUnitID != userUnitID &&
		message.SenderUnitID != userUnitID {
		return fmt.Errorf("no tiene permisos para acceder a este mensaje")
	}

	// Obtener el estado READ
	var readStatus models.MessageStatus
	if err := s.db.WithContext(ctx).Where("code = ?", "READ").First(&readStatus).Error; err != nil {
		return fmt.Errorf("error al obtener estado READ: %w", err)
	}

	// Actualizar el mensaje
	now := time.Now()
	updates := map[string]interface{}{
		"status_id": readStatus.ID,
		"read_at":   &now,
	}

	if err := s.db.WithContext(ctx).Model(&message).Updates(updates).Error; err != nil {
		return fmt.Errorf("error al marcar mensaje como leído: %w", err)
	}

	logger.Info("✅ Mensaje marcado como leído exitosamente")
	return nil
}

// UpdateMessageStatus actualiza el estado de un mensaje
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID int64, statusCode string, userID uuid.UUID) error {
	logger.Info("🔄 Actualizando estado de mensaje - ID: %d, Estado: %s", messageID, statusCode)

	// Verificar que el mensaje existe
	var message models.Message
	if err := s.db.WithContext(ctx).Where("id = ?", messageID).First(&message).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mensaje no encontrado")
		}
		return fmt.Errorf("error al obtener mensaje: %w", err)
	}

	// Verificar permisos del usuario
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Solo puede cambiar estado si es de la unidad receptora o admin
	userUnitID := 0
	if user.OrganizationalUnitID != nil {
		userUnitID = *user.OrganizationalUnitID
	}

	if user.Role != "admin" && message.ReceiverUnitID != userUnitID {
		return fmt.Errorf("no tiene permisos para cambiar el estado de este mensaje")
	}

	// Obtener el nuevo estado
	var newStatus models.MessageStatus
	if err := s.db.WithContext(ctx).Where("code = ?", statusCode).First(&newStatus).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("estado no válido: %s", statusCode)
		}
		return fmt.Errorf("error al obtener estado: %w", err)
	}

	// Preparar updates
	updates := map[string]interface{}{
		"status_id": newStatus.ID,
	}

	// Si es RESPONDED, marcar responded_at
	if statusCode == "RESPONDED" {
		now := time.Now()
		updates["responded_at"] = &now
	}

	// Actualizar el mensaje
	if err := s.db.WithContext(ctx).Model(&message).Updates(updates).Error; err != nil {
		return fmt.Errorf("error al actualizar estado: %w", err)
	}

	logger.Info("✅ Estado actualizado exitosamente")
	return nil
}

// SearchMessages busca mensajes por texto
func (s *MessageService) SearchMessages(ctx context.Context, searchText string, userID uuid.UUID, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Debug("🔍 Buscando mensajes con texto: %s", searchText)

	// Obtener la unidad del usuario
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, 0, fmt.Errorf("usuario no encontrado: %w", err)
	}

	query := s.db.WithContext(ctx).Model(&models.Message{})

	// Filtrar por unidad del usuario
	userUnitID := 0
	if user.OrganizationalUnitID != nil {
		userUnitID = *user.OrganizationalUnitID
	}

	if user.Role != "admin" {
		query = query.Where("sender_unit_id = ? OR receiver_unit_id = ?",
			userUnitID, userUnitID)
	}

	// Búsqueda por texto en subject y content
	searchPattern := "%" + searchText + "%"
	query = query.Where("subject ILIKE ? OR content ILIKE ?", searchPattern, searchPattern)

	return s.applyFiltersAndGetMessages(ctx, query, req)
}

// GetMessageStats obtiene estadísticas de mensajes
func (s *MessageService) GetMessageStats(ctx context.Context, unitID *int) (*MessageStats, error) {
	logger.Debug("📊 Obteniendo estadísticas de mensajes")

	stats := &MessageStats{}
	query := s.db.WithContext(ctx).Model(&models.Message{})

	// Filtrar por unidad si se especifica
	if unitID != nil {
		query = query.Where("sender_unit_id = ? OR receiver_unit_id = ?", *unitID, *unitID)
	}

	// Total de mensajes
	if err := query.Count(&stats.Total).Error; err != nil {
		return nil, fmt.Errorf("error al contar mensajes totales: %w", err)
	}

	// Mensajes no leídos
	if err := query.Joins("JOIN message_statuses ON messages.status_id = message_statuses.id").
		Where("message_statuses.code IN (?)", []string{"SENT", "IN_PROGRESS"}).
		Count(&stats.Unread).Error; err != nil {
		return nil, fmt.Errorf("error al contar mensajes no leídos: %w", err)
	}

	// Mensajes urgentes
	if err := query.Where("is_urgent = ?", true).Count(&stats.Urgent).Error; err != nil {
		return nil, fmt.Errorf("error al contar mensajes urgentes: %w", err)
	}

	// Mensajes del día
	today := time.Now().Truncate(24 * time.Hour)
	if err := query.Where("created_at >= ?", today).Count(&stats.Today).Error; err != nil {
		return nil, fmt.Errorf("error al contar mensajes del día: %w", err)
	}

	return stats, nil
}

// Funciones auxiliares

// applyFiltersAndGetMessages aplica filtros y retorna mensajes con paginación
func (s *MessageService) applyFiltersAndGetMessages(ctx context.Context, query *gorm.DB, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	// Aplicar filtros
	if req.MessageType != nil {
		query = query.Where("message_type_id = ?", *req.MessageType)
	}
	if req.Status != nil {
		query = query.Where("status_id = ?", *req.Status)
	}
	if req.IsUrgent != nil {
		query = query.Where("is_urgent = ?", *req.IsUrgent)
	}
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", *req.DateTo)
	}

	// Contar total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("error al contar mensajes: %w", err)
	}

	// Aplicar ordenamiento
	orderBy := "created_at DESC" // Por defecto
	if req.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", req.SortBy, req.SortOrder)
	}
	query = query.Order(orderBy)

	// Aplicar paginación
	offset := (req.Page - 1) * req.Limit
	query = query.Offset(offset).Limit(req.Limit)

	// Cargar relaciones y obtener mensajes
	var messages []models.Message
	err := query.
		Preload("Sender").
		Preload("SenderUnit").
		Preload("ReceiverUnit").
		Preload("MessageType").
		Preload("Status").
		Preload("Attachments").
		Find(&messages).Error

	if err != nil {
		return nil, 0, fmt.Errorf("error al obtener mensajes: %w", err)
	}

	// Convertir a responses
	responses := make([]MessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = *s.convertToResponse(&message)
	}

	return responses, total, nil
}

// convertToResponse convierte un modelo a response
func (s *MessageService) convertToResponse(message *models.Message) *MessageResponse {
	return &MessageResponse{
		ID:             message.ID,
		Subject:        message.Subject,
		Content:        message.Content,
		SenderID:       message.SenderID,
		SenderUnitID:   message.SenderUnitID,
		ReceiverUnitID: message.ReceiverUnitID,
		MessageTypeID:  message.MessageTypeID,
		StatusID:       message.StatusID,
		PriorityLevel:  message.PriorityLevel,
		IsUrgent:       message.IsUrgent,
		ReadAt:         message.ReadAt,
		RespondedAt:    message.RespondedAt,
		CreatedAt:      message.CreatedAt,
		UpdatedAt:      message.UpdatedAt,
		Sender:         message.Sender,
		SenderUnit:     message.SenderUnit,
		ReceiverUnit:   message.ReceiverUnit,
		MessageType:    message.MessageType,
		Status:         message.Status,
		Attachments:    message.Attachments,
	}
}

// MessageStats representa estadísticas de mensajes
type MessageStats struct {
	Total  int64 `json:"total"`
	Unread int64 `json:"unread"`
	Urgent int64 `json:"urgent"`
	Today  int64 `json:"today"`
}

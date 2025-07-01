// internal/services/message_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageService maneja la lógica de negocio para mensajes
type MessageService struct {
	messageRepo *repositories.MessageRepository
	userRepo    *repositories.UserRepository
	auditRepo   *repositories.AuditRepository
	notifyRepo  *repositories.NotificationRepository
	db          *gorm.DB
}

// NewMessageService crea una nueva instancia del servicio
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{
		messageRepo: repositories.NewMessageRepository(db),
		userRepo:    repositories.NewUserRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
		notifyRepo:  repositories.NewNotificationRepository(db),
		db:          db,
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
	messageTypes, err := s.messageRepo.GetMessageTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener tipos de mensaje: %w", err)
	}

	var messageType *models.MessageType
	for _, mt := range messageTypes {
		if mt.ID == req.MessageTypeID {
			messageType = mt
			break
		}
	}
	if messageType == nil {
		return nil, fmt.Errorf("tipo de mensaje no encontrado o inactivo")
	}

	// Obtener el estado inicial (SENT)
	statuses, err := s.messageRepo.GetMessageStatuses(ctx)
	if err != nil {
		return nil, fmt.Errorf("error al obtener estados: %w", err)
	}

	var sentStatus *models.MessageStatus
	for _, st := range statuses {
		if st.Code == "SENT" {
			sentStatus = st
			break
		}
	}
	if sentStatus == nil {
		return nil, fmt.Errorf("estado inicial no encontrado")
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

	// Crear el mensaje usando el repositorio
	if err := s.messageRepo.Create(ctx, &message); err != nil {
		return nil, fmt.Errorf("error al crear mensaje: %w", err)
	}

	// Registrar en auditoría
	auditLog := &models.AuditLog{
		UserID:     &req.SenderID,
		Action:     models.AuditActionCreate,
		Resource:   "messages",
		ResourceID: fmt.Sprintf("%d", message.ID),
		NewValues: map[string]interface{}{
			"subject":        message.Subject,
			"receiver_unit":  req.ReceiverUnitID,
			"message_type":   req.MessageTypeID,
			"priority_level": message.PriorityLevel,
			"is_urgent":      message.IsUrgent,
		},
		Result: models.AuditResultSuccess,
	}
	s.auditRepo.Create(ctx, auditLog)

	// Crear notificaciones para usuarios de la unidad receptora
	go s.createNotificationsForUnit(context.Background(), message.ID, req.ReceiverUnitID, req.Subject)

	logger.Info("✅ Mensaje creado exitosamente - ID: %d", message.ID)

	// Retornar el mensaje completo con relaciones
	return s.GetMessageByID(ctx, message.ID)
}

// GetMessagesByUnit obtiene mensajes por unidad organizacional
func (s *MessageService) GetMessagesByUnit(ctx context.Context, unitID int, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Debug("📋 Obteniendo mensajes para unidad: %d", unitID)

	filter := s.buildMessageFilter(req)

	// Filtrar por unidad
	filter.SenderUnitID = &unitID
	messages, total, err := s.messageRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Si no hay mensajes como emisor, buscar como receptor
	if len(messages) == 0 {
		filter.SenderUnitID = nil
		filter.ReceiverUnitID = &unitID
		messages, total, err = s.messageRepo.GetByFilter(ctx, filter)
		if err != nil {
			return nil, 0, err
		}
	}

	return s.convertMessagesToResponses(messages), total, nil
}

// GetMessagesByUser obtiene mensajes por usuario específico
func (s *MessageService) GetMessagesByUser(ctx context.Context, userID uuid.UUID, req *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Debug("📋 Obteniendo mensajes para usuario: %s", userID.String())

	filter := s.buildMessageFilter(req)
	filter.UserID = &userID

	messages, total, err := s.messageRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return s.convertMessagesToResponses(messages), total, nil
}

// GetMessageByID obtiene un mensaje específico por ID
func (s *MessageService) GetMessageByID(ctx context.Context, messageID int64) (*MessageResponse, error) {
	logger.Debug("🔍 Obteniendo mensaje ID: %d", messageID)

	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mensaje no encontrado")
		}
		return nil, fmt.Errorf("error al obtener mensaje: %w", err)
	}

	return s.convertToResponse(message), nil
}

// MarkAsRead marca un mensaje como leído
func (s *MessageService) MarkAsRead(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Info("👁️ Marcando mensaje como leído - ID: %d, Usuario: %s", messageID, userID.String())

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("mensaje no encontrado")
		}
		return fmt.Errorf("error al obtener mensaje: %w", err)
	}

	// Verificar permisos
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
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

	// Marcar como leído
	if err := s.messageRepo.MarkAsRead(ctx, messageID); err != nil {
		return fmt.Errorf("error al marcar como leído: %w", err)
	}

	// Registrar en auditoría
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     models.AuditActionRead,
		Resource:   "messages",
		ResourceID: fmt.Sprintf("%d", messageID),
		Result:     models.AuditResultSuccess,
	}
	s.auditRepo.Create(ctx, auditLog)

	logger.Info("✅ Mensaje marcado como leído")
	return nil
}

// UpdateMessageStatus actualiza el estado de un mensaje
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID int64, statusID int, userID uuid.UUID) error {
	logger.Info("🔄 Actualizando estado de mensaje - ID: %d, Nuevo estado: %d", messageID, statusID)

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Actualizar estado
	message.StatusID = statusID
	if err := s.messageRepo.Update(ctx, message); err != nil {
		return fmt.Errorf("error al actualizar estado: %w", err)
	}

	// Si el estado es "RESPONDED", marcar como respondido
	statuses, _ := s.messageRepo.GetMessageStatuses(ctx)
	for _, status := range statuses {
		if status.ID == statusID && status.Code == "RESPONDED" {
			s.messageRepo.MarkAsResponded(ctx, messageID)
			break
		}
	}

	// Registrar en auditoría
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     models.AuditActionUpdate,
		Resource:   "messages",
		ResourceID: fmt.Sprintf("%d", messageID),
		OldValues: map[string]interface{}{
			"status_id": message.StatusID,
		},
		NewValues: map[string]interface{}{
			"status_id": statusID,
		},
		Result: models.AuditResultSuccess,
	}
	s.auditRepo.Create(ctx, auditLog)

	logger.Info("✅ Estado actualizado exitosamente")
	return nil
}

// DeleteMessage elimina un mensaje
func (s *MessageService) DeleteMessage(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Warn("🗑️ Eliminando mensaje - ID: %d, Usuario: %s", messageID, userID.String())

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos (solo admin o el emisor pueden eliminar)
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	if user.Role != "admin" && message.SenderID != userID {
		return fmt.Errorf("no tiene permisos para eliminar este mensaje")
	}

	// Eliminar mensaje
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		return fmt.Errorf("error al eliminar mensaje: %w", err)
	}

	// Registrar en auditoría
	auditLog := &models.AuditLog{
		UserID:     &userID,
		Action:     models.AuditActionDelete,
		Resource:   "messages",
		ResourceID: fmt.Sprintf("%d", messageID),
		OldValues: map[string]interface{}{
			"subject":       message.Subject,
			"sender":        message.SenderID,
			"receiver_unit": message.ReceiverUnitID,
		},
		Result: models.AuditResultSuccess,
	}
	s.auditRepo.Create(ctx, auditLog)

	logger.Info("✅ Mensaje eliminado exitosamente")
	return nil
}

// GetMessageStats obtiene estadísticas de mensajes
func (s *MessageService) GetMessageStats(ctx context.Context, unitID int) (*MessageStats, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	stats, err := s.messageRepo.GetStatsByUnit(ctx, unitID, today, now)
	if err != nil {
		return nil, err
	}

	return &MessageStats{
		Total:  stats.TotalReceived + stats.TotalSent,
		Unread: stats.Unread,
		Urgent: stats.Urgent,
		Today:  stats.TotalReceived, // Mensajes recibidos hoy
	}, nil
}

// GetMessageTypes obtiene los tipos de mensaje disponibles
func (s *MessageService) GetMessageTypes(ctx context.Context) ([]*models.MessageType, error) {
	return s.messageRepo.GetMessageTypes(ctx)
}

// GetMessageStatuses obtiene los estados de mensaje disponibles
func (s *MessageService) GetMessageStatuses(ctx context.Context) ([]*models.MessageStatus, error) {
	return s.messageRepo.GetMessageStatuses(ctx)
}

// Funciones auxiliares

// buildMessageFilter construye el filtro del repositorio desde el request
func (s *MessageService) buildMessageFilter(req *GetMessagesRequest) *repositories.MessageFilter {
	filter := &repositories.MessageFilter{
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
		SortBy: req.SortBy,
	}

	if req.SortOrder == "desc" {
		filter.SortDesc = true
	}

	if req.MessageType != nil {
		filter.MessageTypeID = req.MessageType
	}

	if req.Status != nil {
		filter.StatusID = req.Status
	}

	if req.IsUrgent != nil {
		filter.IsUrgent = req.IsUrgent
	}

	if req.DateFrom != nil {
		filter.DateFrom = req.DateFrom
	}

	if req.DateTo != nil {
		filter.DateTo = req.DateTo
	}

	if req.SearchText != nil {
		filter.SearchTerm = *req.SearchText
	}

	return filter
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

// convertMessagesToResponses convierte múltiples modelos a responses
func (s *MessageService) convertMessagesToResponses(messages []*models.Message) []MessageResponse {
	responses := make([]MessageResponse, len(messages))
	for i, message := range messages {
		responses[i] = *s.convertToResponse(message)
	}
	return responses
}

// createNotificationsForUnit crea notificaciones para los usuarios de una unidad
func (s *MessageService) createNotificationsForUnit(ctx context.Context, messageID int64, unitID int, subject string) {
	// Obtener usuarios de la unidad
	users, err := s.userRepo.GetByOrganizationalUnit(ctx, unitID)
	if err != nil {
		logger.Error("Error al obtener usuarios de la unidad: %v", err)
		return
	}

	// Crear notificación para cada usuario activo
	for _, user := range users {
		if !user.IsActive {
			continue
		}

		notification := &models.Notification{
			UserID:           user.ID,
			Type:             models.NotificationTypeMessage,
			Title:            "Nuevo mensaje recibido",
			Content:          fmt.Sprintf("Has recibido un nuevo mensaje: %s", subject),
			Priority:         models.NotificationPriorityNormal,
			RelatedMessageID: &messageID,
		}

		if err := s.notifyRepo.Create(ctx, notification); err != nil {
			logger.Error("Error al crear notificación para usuario %s: %v", user.ID, err)
		}
	}
}

// MessageStats representa estadísticas de mensajes
type MessageStats struct {
	Total  int64 `json:"total"`
	Unread int64 `json:"unread"`
	Urgent int64 `json:"urgent"`
	Today  int64 `json:"today"`
}

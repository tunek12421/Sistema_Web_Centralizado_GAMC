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
	ID             int64                      `json:"id"`
	Subject        string                     `json:"subject"`
	Content        string                     `json:"content"`
	SenderID       uuid.UUID                  `json:"senderId"`
	SenderUnitID   int                        `json:"senderUnitId"`
	ReceiverUnitID int                        `json:"receiverUnitId"`
	MessageTypeID  int                        `json:"messageTypeId"`
	StatusID       int                        `json:"statusId"`
	PriorityLevel  int                        `json:"priorityLevel"`
	IsUrgent       bool                       `json:"isUrgent"`
	ReadAt         *time.Time                 `json:"readAt,omitempty"`
	RespondedAt    *time.Time                 `json:"respondedAt,omitempty"`
	CreatedAt      time.Time                  `json:"createdAt"`
	UpdatedAt      time.Time                  `json:"updatedAt"`
	Sender         *models.User               `json:"sender,omitempty"`
	SenderUnit     *models.OrganizationalUnit `json:"senderUnit,omitempty"`
	ReceiverUnit   *models.OrganizationalUnit `json:"receiverUnit,omitempty"`
	MessageType    *models.MessageType        `json:"messageType,omitempty"`
	Status         *models.MessageStatus      `json:"status,omitempty"`
	Attachments    []models.MessageAttachment `json:"attachments,omitempty"`
}

// CreateMessage crea un nuevo mensaje
func (s *MessageService) CreateMessage(ctx context.Context, req *CreateMessageRequest) (*MessageResponse, error) {
	logger.Info("📨 Creando mensaje: %s", req.Subject)

	// Validar unidad receptora
	var receiverUnit models.OrganizationalUnit
	if err := s.db.WithContext(ctx).First(&receiverUnit, req.ReceiverUnitID).Error; err != nil {
		return nil, fmt.Errorf("unidad receptora no encontrada")
	}

	// Validar tipo de mensaje
	var messageType models.MessageType
	if err := s.db.WithContext(ctx).First(&messageType, req.MessageTypeID).Error; err != nil {
		return nil, fmt.Errorf("tipo de mensaje no encontrado")
	}

	// Crear mensaje
	message := &models.Message{
		Subject:        req.Subject,
		Content:        req.Content,
		SenderID:       req.SenderID,
		SenderUnitID:   req.SenderUnitID,
		ReceiverUnitID: req.ReceiverUnitID,
		MessageTypeID:  req.MessageTypeID,
		StatusID:       1, // Estado inicial: "enviado"
		PriorityLevel:  req.PriorityLevel,
		IsUrgent:       req.IsUrgent,
	}

	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, fmt.Errorf("error al crear mensaje: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, req.SenderID, models.AuditActionCreate, "messages", fmt.Sprintf("%d", message.ID), nil, map[string]interface{}{
		"subject":        req.Subject,
		"receiver_unit":  req.ReceiverUnitID,
		"message_type":   req.MessageTypeID,
		"priority_level": req.PriorityLevel,
		"is_urgent":      req.IsUrgent,
	})

	// Crear notificaciones para la unidad receptora
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
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos (el usuario debe pertenecer a la unidad receptora)
	if err := s.verifyReadPermissions(ctx, message, userID); err != nil {
		return err
	}

	// Marcar como leído
	if err := s.messageRepo.MarkAsRead(ctx, messageID); err != nil {
		return fmt.Errorf("error al marcar como leído: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "messages", fmt.Sprintf("%d", messageID),
		map[string]interface{}{"read_at": nil},
		map[string]interface{}{"read_at": time.Now()})

	return nil
}

// MarkAsResponded marca un mensaje como respondido
func (s *MessageService) MarkAsResponded(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Info("💬 Marcando mensaje como respondido - ID: %d", messageID)

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos
	if err := s.verifyReadPermissions(ctx, message, userID); err != nil {
		return err
	}

	// Marcar como respondido
	if err := s.messageRepo.MarkAsResponded(ctx, messageID); err != nil {
		return fmt.Errorf("error al marcar como respondido: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "messages", fmt.Sprintf("%d", messageID),
		map[string]interface{}{"responded_at": nil},
		map[string]interface{}{"responded_at": time.Now()})

	return nil
}

// ArchiveMessage archiva un mensaje
func (s *MessageService) ArchiveMessage(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Info("📦 Archivando mensaje - ID: %d", messageID)

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos
	if err := s.verifyReadPermissions(ctx, message, userID); err != nil {
		return err
	}

	// Archivar mensaje
	if err := s.messageRepo.Archive(ctx, messageID); err != nil {
		return fmt.Errorf("error al archivar mensaje: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "messages", fmt.Sprintf("%d", messageID),
		map[string]interface{}{"archived_at": nil},
		map[string]interface{}{"archived_at": time.Now()})

	return nil
}

// UnarchiveMessage desarchivar un mensaje
func (s *MessageService) UnarchiveMessage(ctx context.Context, messageID int64, userID uuid.UUID) error {
	logger.Info("📤 Desarchivando mensaje - ID: %d", messageID)

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos
	if err := s.verifyReadPermissions(ctx, message, userID); err != nil {
		return err
	}

	// Desarchivar mensaje
	if err := s.messageRepo.Unarchive(ctx, messageID); err != nil {
		return fmt.Errorf("error al desarchivar mensaje: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "messages", fmt.Sprintf("%d", messageID),
		map[string]interface{}{"archived_at": time.Now()},
		map[string]interface{}{"archived_at": nil})

	return nil
}

// GetMessageStats obtiene estadísticas de mensajes
func (s *MessageService) GetMessageStats(ctx context.Context, unitID int, userID uuid.UUID) (*MessageStats, error) {
	logger.Debug("📊 Obteniendo estadísticas para unidad: %d", unitID)

	// CORREGIDO: Comentar la variable no utilizada temporalmente si existe
	// user, err := s.userRepo.GetByID(ctx, userID)
	// if err != nil {
	//     return nil, fmt.Errorf("usuario no encontrado: %w", err)
	// }

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	stats, err := s.messageRepo.GetStatsByUnit(ctx, unitID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("error al obtener estadísticas: %w", err)
	}

	return &MessageStats{
		Total:  stats.TotalSent + stats.TotalReceived,
		Unread: stats.Unread,
		Urgent: stats.Urgent,
		Today:  stats.TotalSent + stats.TotalReceived, // Estadísticas del día actual
	}, nil
}

// SearchMessages busca mensajes por texto
func (s *MessageService) SearchMessages(ctx context.Context, searchText string, userID uuid.UUID, filters *GetMessagesRequest) ([]MessageResponse, int64, error) {
	logger.Info("🔍 Buscando mensajes con texto: %s", searchText)

	// Obtener usuario para verificar permisos
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Configurar filtros base
	if filters == nil {
		filters = &GetMessagesRequest{
			Page:      1,
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}
	}

	// Agregar filtro de búsqueda
	filters.SearchText = &searchText

	// Si no es admin, filtrar por su unidad
	if user.Role != "admin" && user.OrganizationalUnitID != nil {
		filters.UnitID = user.OrganizationalUnitID
	}

	// CORREGIDO: Convertir a tipo compatible del repositorio
	repoFilter := &repositories.GetMessagesRequest{
		UnitID:      filters.UnitID,
		UserID:      filters.UserID,
		MessageType: filters.MessageType,
		Status:      filters.Status,
		IsUrgent:    filters.IsUrgent,
		DateFrom:    filters.DateFrom,
		DateTo:      filters.DateTo,
		SearchText:  filters.SearchText,
		Page:        filters.Page,
		Limit:       filters.Limit,
		SortBy:      filters.SortBy,
		SortOrder:   filters.SortOrder,
	}

	messages, total, err := s.messageRepo.Search(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("error al buscar mensajes: %w", err)
	}

	// Convertir a response
	responses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		responses[i] = *s.convertToResponse(msg)
	}

	return responses, total, nil
}

// GetMessageTypes obtiene los tipos de mensaje disponibles
func (s *MessageService) GetMessageTypes(ctx context.Context) ([]*models.MessageType, error) {
	return s.messageRepo.GetMessageTypes(ctx)
}

// GetMessageStatuses obtiene los estados de mensaje disponibles
func (s *MessageService) GetMessageStatuses(ctx context.Context) ([]*models.MessageStatus, error) {
	return s.messageRepo.GetMessageStatuses(ctx)
}

// UpdateMessageStatus actualiza el estado de un mensaje - MÉTODO FALTANTE AGREGADO
func (s *MessageService) UpdateMessageStatus(ctx context.Context, messageID int64, statusID int, userID uuid.UUID) error {
	logger.Info("🔄 Actualizando estado de mensaje - ID: %d, Nuevo estado: %d", messageID, statusID)

	// Verificar que el mensaje existe
	message, err := s.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("mensaje no encontrado: %w", err)
	}

	// Verificar permisos
	if err := s.verifyReadPermissions(ctx, message, userID); err != nil {
		return err
	}

	// Verificar que el estado es válido
	var status models.MessageStatus
	if err := s.db.WithContext(ctx).First(&status, statusID).Error; err != nil {
		return fmt.Errorf("estado de mensaje no válido")
	}

	// Actualizar estado
	oldStatus := message.StatusID
	message.StatusID = statusID

	if err := s.messageRepo.Update(ctx, message); err != nil {
		return fmt.Errorf("error al actualizar estado del mensaje: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "messages", fmt.Sprintf("%d", messageID),
		map[string]interface{}{"status_id": oldStatus},
		map[string]interface{}{"status_id": statusID})

	logger.Info("✅ Estado de mensaje actualizado exitosamente")
	return nil
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

// verifyReadPermissions verifica si un usuario tiene permisos para leer un mensaje
func (s *MessageService) verifyReadPermissions(ctx context.Context, message *models.Message, userID uuid.UUID) error {
	// Obtener usuario
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Los administradores pueden leer cualquier mensaje
	if user.Role == "admin" {
		return nil
	}

	// El usuario debe pertenecer a la unidad emisora o receptora
	if user.OrganizationalUnitID != nil {
		unitID := *user.OrganizationalUnitID
		if unitID == message.SenderUnitID || unitID == message.ReceiverUnitID {
			return nil
		}
	}

	return fmt.Errorf("no tiene permisos para acceder a este mensaje")
}

// auditLog registra una acción en el log de auditoría
func (s *MessageService) auditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, resource, resourceID string, oldValues, newValues map[string]interface{}) {
	log := &models.AuditLog{
		UserID:     &userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValues:  oldValues,
		NewValues:  newValues,
		Result:     models.AuditResultSuccess,
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Error al registrar en auditoría: %v", err)
	}
}

// MessageStats representa estadísticas de mensajes
type MessageStats struct {
	Total  int64 `json:"total"`
	Unread int64 `json:"unread"`
	Urgent int64 `json:"urgent"`
	Today  int64 `json:"today"`
}

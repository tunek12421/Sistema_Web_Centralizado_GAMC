// internal/services/notification_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService maneja la lógica de negocio para notificaciones
type NotificationService struct {
	notifyRepo *repositories.NotificationRepository
	userRepo   *repositories.UserRepository
	db         *gorm.DB
}

// NewNotificationService crea una nueva instancia del servicio
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{
		notifyRepo: repositories.NewNotificationRepository(db),
		userRepo:   repositories.NewUserRepository(db),
		db:         db,
	}
}

// CreateNotificationRequest solicitud para crear notificación
type CreateNotificationRequest struct {
	UserID           uuid.UUID                   `json:"userId" validate:"required"`
	Type             models.NotificationType     `json:"type" validate:"required"`
	Title            string                      `json:"title" validate:"required,min=3,max=255"`
	Content          string                      `json:"content" validate:"required"`
	Priority         models.NotificationPriority `json:"priority"`
	RelatedMessageID *int64                      `json:"relatedMessageId,omitempty"`
	ActionURL        string                      `json:"actionUrl,omitempty"`
	Metadata         map[string]interface{}      `json:"metadata,omitempty"`
}

// NotificationResponse respuesta de notificación
type NotificationResponse struct {
	ID               uuid.UUID              `json:"id"`
	UserID           uuid.UUID              `json:"userId"`
	Type             string                 `json:"type"`
	Title            string                 `json:"title"`
	Content          string                 `json:"content"`
	Priority         string                 `json:"priority"`
	IsRead           bool                   `json:"isRead"`
	ReadAt           *time.Time             `json:"readAt,omitempty"`
	RelatedMessageID *int64                 `json:"relatedMessageId,omitempty"`
	ActionURL        string                 `json:"actionUrl,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"createdAt"`
}

// CreateNotification crea una nueva notificación
func (s *NotificationService) CreateNotification(ctx context.Context, req *CreateNotificationRequest) (*NotificationResponse, error) {
	logger.Info("🔔 Creando notificación para usuario: %s", req.UserID)

	// Verificar que el usuario existe
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %w", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("usuario inactivo")
	}

	// Crear notificación
	notification := &models.Notification{
		UserID:           req.UserID,
		Type:             req.Type,
		Title:            req.Title,
		Content:          req.Content,
		Priority:         req.Priority,
		RelatedMessageID: req.RelatedMessageID,
		ActionURL:        req.ActionURL,
		Metadata:         req.Metadata,
	}

	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}

	if err := s.notifyRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("error al crear notificación: %w", err)
	}

	logger.Info("✅ Notificación creada: %s", notification.ID)

	// TODO: Aquí se podría enviar la notificación por WebSocket
	// go s.sendWebSocketNotification(notification)

	return s.convertToResponse(notification), nil
}

// CreateBulkNotifications crea notificaciones para múltiples usuarios
func (s *NotificationService) CreateBulkNotifications(ctx context.Context, userIDs []uuid.UUID, notifType models.NotificationType, title, content string, priority models.NotificationPriority) error {
	logger.Info("📢 Creando notificaciones masivas para %d usuarios", len(userIDs))

	if len(userIDs) == 0 {
		return fmt.Errorf("no se especificaron usuarios")
	}

	// Crear notificaciones
	notifications := make([]*models.Notification, 0, len(userIDs))
	for _, userID := range userIDs {
		notifications = append(notifications, &models.Notification{
			UserID:   userID,
			Type:     notifType,
			Title:    title,
			Content:  content,
			Priority: priority,
		})
	}

	if err := s.notifyRepo.CreateBatch(ctx, notifications); err != nil {
		return fmt.Errorf("error al crear notificaciones masivas: %w", err)
	}

	logger.Info("✅ Notificaciones masivas creadas: %d", len(notifications))
	return nil
}

// GetUserNotifications obtiene las notificaciones de un usuario
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, page, limit int) ([]*NotificationResponse, int64, error) {
	filter := &repositories.NotificationFilter{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	if unreadOnly {
		isRead := false
		filter.IsRead = &isRead
	}

	notifications, total, err := s.notifyRepo.GetByUserID(ctx, userID, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = s.convertToResponse(notification)
	}

	return responses, total, nil
}

// GetNotification obtiene una notificación específica
func (s *NotificationService) GetNotification(ctx context.Context, notificationID, userID uuid.UUID) (*NotificationResponse, error) {
	notification, err := s.notifyRepo.GetByID(ctx, notificationID)
	if err != nil {
		return nil, fmt.Errorf("notificación no encontrada: %w", err)
	}

	// Verificar que la notificación pertenece al usuario
	if notification.UserID != userID {
		return nil, fmt.Errorf("no tiene permisos para acceder a esta notificación")
	}

	return s.convertToResponse(notification), nil
}

// MarkAsRead marca una notificación como leída
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	// Verificar que la notificación existe y pertenece al usuario
	notification, err := s.notifyRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notificación no encontrada: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("no tiene permisos para modificar esta notificación")
	}

	if notification.IsRead {
		return nil // Ya está leída
	}

	if err := s.notifyRepo.MarkAsRead(ctx, notificationID); err != nil {
		return fmt.Errorf("error al marcar como leída: %w", err)
	}

	logger.Debug("✅ Notificación marcada como leída: %s", notificationID)
	return nil
}

// MarkAllAsRead marca todas las notificaciones de un usuario como leídas
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if err := s.notifyRepo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("error al marcar todas como leídas: %w", err)
	}

	logger.Info("✅ Todas las notificaciones marcadas como leídas para usuario: %s", userID)
	return nil
}

// DeleteNotification elimina una notificación
func (s *NotificationService) DeleteNotification(ctx context.Context, notificationID, userID uuid.UUID) error {
	// Verificar que la notificación existe y pertenece al usuario
	notification, err := s.notifyRepo.GetByID(ctx, notificationID)
	if err != nil {
		return fmt.Errorf("notificación no encontrada: %w", err)
	}

	if notification.UserID != userID {
		return fmt.Errorf("no tiene permisos para eliminar esta notificación")
	}

	if err := s.notifyRepo.Delete(ctx, notificationID); err != nil {
		return fmt.Errorf("error al eliminar notificación: %w", err)
	}

	logger.Debug("✅ Notificación eliminada: %s", notificationID)
	return nil
}

// DeleteMultipleNotifications elimina múltiples notificaciones
func (s *NotificationService) DeleteMultipleNotifications(ctx context.Context, notificationIDs []uuid.UUID, userID uuid.UUID) error {
	// Verificar que todas las notificaciones pertenecen al usuario
	for _, id := range notificationIDs {
		notification, err := s.notifyRepo.GetByID(ctx, id)
		if err != nil {
			continue // Ignorar si no existe
		}

		if notification.UserID != userID {
			return fmt.Errorf("no tiene permisos para eliminar la notificación: %s", id)
		}
	}

	if err := s.notifyRepo.BatchDelete(ctx, notificationIDs); err != nil {
		return fmt.Errorf("error al eliminar notificaciones: %w", err)
	}

	logger.Info("✅ Notificaciones eliminadas: %d", len(notificationIDs))
	return nil
}

// GetUnreadCount obtiene el conteo de notificaciones no leídas
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.notifyRepo.GetUnreadCount(ctx, userID)
}

// GetNotificationStats obtiene estadísticas de notificaciones
func (s *NotificationService) GetNotificationStats(ctx context.Context, userID uuid.UUID) (*repositories.NotificationStats, error) {
	return s.notifyRepo.GetNotificationStats(ctx, userID)
}

// CreateSystemNotification crea una notificación del sistema
func (s *NotificationService) CreateSystemNotification(ctx context.Context, title, content string, priority models.NotificationPriority) error {
	// Obtener todos los usuarios activos
	filter := &repositories.UserFilter{
		IsActive: &[]bool{true}[0],
		Limit:    10000, // Límite alto para obtener todos
	}

	users, _, err := s.userRepo.GetByFilter(ctx, filter)
	if err != nil {
		return fmt.Errorf("error al obtener usuarios: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no hay usuarios activos")
	}

	// Extraer IDs de usuarios
	userIDs := make([]uuid.UUID, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	// Crear notificaciones
	return s.notifyRepo.CreateSystemNotification(ctx, userIDs, title, content, priority)
}

// CreateNotificationForRole crea notificaciones para usuarios con un rol específico
func (s *NotificationService) CreateNotificationForRole(ctx context.Context, role string, notifType models.NotificationType, title, content string, priority models.NotificationPriority) error {
	// Obtener usuarios por rol
	users, err := s.userRepo.GetByRole(ctx, role)
	if err != nil {
		return fmt.Errorf("error al obtener usuarios por rol: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no hay usuarios con rol: %s", role)
	}

	// Extraer IDs de usuarios activos
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		if user.IsActive {
			userIDs = append(userIDs, user.ID)
		}
	}

	if len(userIDs) == 0 {
		return fmt.Errorf("no hay usuarios activos con rol: %s", role)
	}

	return s.CreateBulkNotifications(ctx, userIDs, notifType, title, content, priority)
}

// CreateNotificationForUnit crea notificaciones para usuarios de una unidad
func (s *NotificationService) CreateNotificationForUnit(ctx context.Context, unitID int, notifType models.NotificationType, title, content string, priority models.NotificationPriority) error {
	// Obtener usuarios de la unidad
	users, err := s.userRepo.GetByOrganizationalUnit(ctx, unitID)
	if err != nil {
		return fmt.Errorf("error al obtener usuarios de la unidad: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no hay usuarios en la unidad: %d", unitID)
	}

	// Extraer IDs de usuarios activos
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, user := range users {
		if user.IsActive {
			userIDs = append(userIDs, user.ID)
		}
	}

	if len(userIDs) == 0 {
		return fmt.Errorf("no hay usuarios activos en la unidad: %d", unitID)
	}

	return s.CreateBulkNotifications(ctx, userIDs, notifType, title, content, priority)
}

// CleanupOldNotifications limpia notificaciones antiguas
func (s *NotificationService) CleanupOldNotifications(ctx context.Context, days int) error {
	logger.Info("🧹 Limpiando notificaciones antiguas (> %d días)", days)

	deleted, err := s.notifyRepo.DeleteOldNotifications(ctx, days)
	if err != nil {
		return fmt.Errorf("error al limpiar notificaciones: %w", err)
	}

	logger.Info("✅ Notificaciones eliminadas: %d", deleted)
	return nil
}

// Funciones auxiliares

// convertToResponse convierte un modelo a response
func (s *NotificationService) convertToResponse(notification *models.Notification) *NotificationResponse {
	return &NotificationResponse{
		ID:               notification.ID,
		UserID:           notification.UserID,
		Type:             string(notification.Type),
		Title:            notification.Title,
		Content:          notification.Content,
		Priority:         string(notification.Priority),
		IsRead:           notification.IsRead,
		ReadAt:           notification.ReadAt,
		RelatedMessageID: notification.RelatedMessageID,
		ActionURL:        notification.ActionURL,
		Metadata:         notification.Metadata,
		CreatedAt:        notification.CreatedAt,
	}
}

// Tipos de notificación predefinidos

// CreateMessageNotification crea una notificación para un nuevo mensaje
func (s *NotificationService) CreateMessageNotification(ctx context.Context, userID uuid.UUID, messageID int64, messageSubject string) error {
	req := &CreateNotificationRequest{
		UserID:           userID,
		Type:             models.NotificationTypeMessage,
		Title:            "Nuevo mensaje recibido",
		Content:          fmt.Sprintf("Has recibido un nuevo mensaje: %s", messageSubject),
		Priority:         models.NotificationPriorityNormal,
		RelatedMessageID: &messageID,
		ActionURL:        fmt.Sprintf("/messages/%d", messageID),
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

// CreateLoginNotification crea una notificación de nuevo login
func (s *NotificationService) CreateLoginNotification(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error {
	req := &CreateNotificationRequest{
		UserID:   userID,
		Type:     models.NotificationTypeSecurity,
		Title:    "Nuevo inicio de sesión",
		Content:  fmt.Sprintf("Se ha iniciado sesión desde: %s", ipAddress),
		Priority: models.NotificationPriorityLow,
		Metadata: map[string]interface{}{
			"ip_address": ipAddress,
			"user_agent": userAgent,
			"timestamp":  time.Now(),
		},
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

// CreateAlertNotification crea una notificación de alerta
func (s *NotificationService) CreateAlertNotification(ctx context.Context, userID uuid.UUID, title, content string) error {
	req := &CreateNotificationRequest{
		UserID:   userID,
		Type:     models.NotificationTypeAlert,
		Title:    title,
		Content:  content,
		Priority: models.NotificationPriorityHigh,
	}

	_, err := s.CreateNotification(ctx, req)
	return err
}

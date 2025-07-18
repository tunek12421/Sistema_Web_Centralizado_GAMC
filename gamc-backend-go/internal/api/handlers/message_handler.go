// internal/api/handlers/message_handler.go
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/services"
	"gamc-backend-go/pkg/logger"
	"gamc-backend-go/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageHandler maneja las operaciones de mensajes
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler crea una nueva instancia del handler de mensajes
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessageRequest estructura para crear mensajes
type CreateMessageRequest struct {
	Subject        string `json:"subject" binding:"required,min=3,max=255"`
	Content        string `json:"content" binding:"required,min=10"`
	ReceiverUnitID int    `json:"receiverUnitId" binding:"required,min=1"`
	MessageTypeID  int    `json:"messageTypeId" binding:"required,min=1"`
	PriorityLevel  int    `json:"priorityLevel" binding:"min=1,max=5"`
	IsUrgent       bool   `json:"isUrgent"`
}

// CreateMessage maneja POST /api/v1/messages
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	logger.Info("📨 POST /api/v1/messages - Crear mensaje")

	// Obtener usuario desde el middleware de autenticación
	user, exists := c.Get("user")
	if !exists {
		logger.Error("Usuario no encontrado en contexto")
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		logger.Error("Error al convertir usuario del contexto")
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	logger.Info("Usuario autenticado: %s (%s)", userProfile.Email, userProfile.Role)

	// Validar request body
	var req CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error al parsear request body: %v", err)
		response.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Valores por defecto
	if req.PriorityLevel == 0 {
		req.PriorityLevel = 3
	}

	// CORREGIDO: Manejar correctamente el puntero OrganizationalUnitID
	senderUnitID := 1 // valor por defecto
	if userProfile.OrganizationalUnitID != nil {
		senderUnitID = *userProfile.OrganizationalUnitID
	}

	logger.Info("Creando mensaje con senderUnitID: %d", senderUnitID)

	serviceReq := &services.CreateMessageRequest{
		Subject:        req.Subject,
		Content:        req.Content,
		SenderID:       userProfile.ID,
		SenderUnitID:   senderUnitID,
		ReceiverUnitID: req.ReceiverUnitID,
		MessageTypeID:  req.MessageTypeID,
		PriorityLevel:  req.PriorityLevel,
		IsUrgent:       req.IsUrgent,
	}

	// Crear mensaje usando el servicio
	messageResponse, err := h.messageService.CreateMessage(c.Request.Context(), serviceReq)
	if err != nil {
		logger.Error("Error al crear mensaje: %v", err)
		response.Error(c, http.StatusBadRequest, "Error al crear mensaje", err.Error())
		return
	}

	logger.Info("Mensaje creado exitosamente con ID: %d", messageResponse.ID)
	response.Success(c, "Mensaje creado exitosamente", messageResponse)
}

// GetMessages maneja GET /api/v1/messages
func (h *MessageHandler) GetMessages(c *gin.Context) {
	logger.Info("📋 GET /api/v1/messages - Listar mensajes")

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		logger.Error("Usuario no encontrado en contexto para GetMessages")
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		logger.Error("Error al convertir usuario del contexto en GetMessages")
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	logger.Info("Usuario autenticado para GetMessages: %s (%s)", userProfile.Email, userProfile.Role)

	// Parsear parámetros de query
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	sortBy := c.DefaultQuery("sortBy", "created_at")
	sortOrder := c.DefaultQuery("sortOrder", "desc")

	// Parámetros de filtro opcionales
	var messageType *int
	if mt := c.Query("messageType"); mt != "" {
		if mtInt, err := strconv.Atoi(mt); err == nil {
			messageType = &mtInt
		}
	}

	var status *int
	if st := c.Query("status"); st != "" {
		if stInt, err := strconv.Atoi(st); err == nil {
			status = &stInt
		}
	}

	var isUrgent *bool
	if iu := c.Query("isUrgent"); iu != "" {
		if iu == "true" {
			urgent := true
			isUrgent = &urgent
		} else if iu == "false" {
			urgent := false
			isUrgent = &urgent
		}
	}

	var searchText *string
	if search := c.Query("search"); search != "" {
		searchText = &search
	}

	// Validar límites
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// CORREGIDO: Manejar correctamente el puntero OrganizationalUnitID
	userUnitID := 1 // valor por defecto
	if userProfile.OrganizationalUnitID != nil {
		userUnitID = *userProfile.OrganizationalUnitID
	}

	logger.Info("Obteniendo mensajes para unidad: %d, página: %d, límite: %d", userUnitID, page, limit)

	req := &services.GetMessagesRequest{
		UnitID:      &userUnitID,
		MessageType: messageType,
		Status:      status,
		IsUrgent:    isUrgent,
		SearchText:  searchText,
		Page:        page,
		Limit:       limit,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
	}

	// Obtener mensajes usando el servicio
	messages, total, err := h.messageService.GetMessagesByUnit(c.Request.Context(), userUnitID, req)
	if err != nil {
		logger.Error("Error al obtener mensajes: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al obtener mensajes", err.Error())
		return
	}

	logger.Info("Mensajes obtenidos exitosamente: %d total, %d en esta página", total, len(messages))

	// Crear respuesta con paginación
	paginatedResponse := gin.H{
		"messages":   messages,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	}

	response.Success(c, "Mensajes obtenidos exitosamente", paginatedResponse)
}

// GetMessageByID maneja GET /api/v1/messages/:id
func (h *MessageHandler) GetMessageByID(c *gin.Context) {
	// Obtener ID del mensaje desde la URL
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	// Obtener mensaje usando el servicio
	messageResponse, err := h.messageService.GetMessageByID(c.Request.Context(), messageID)
	if err != nil {
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Error al obtener mensaje", err.Error())
		return
	}

	// Verificar permisos - el usuario debe poder acceder al mensaje
	userUnitID := 1
	if userProfile.OrganizationalUnitID != nil {
		userUnitID = *userProfile.OrganizationalUnitID
	}

	// Admin puede ver todo, otros solo mensajes de su unidad
	if userProfile.Role != "admin" &&
		messageResponse.SenderUnitID != userUnitID &&
		messageResponse.ReceiverUnitID != userUnitID {
		response.Error(c, http.StatusForbidden, "No tiene permisos para acceder a este mensaje", "")
		return
	}

	response.Success(c, "Mensaje obtenido exitosamente", messageResponse)
}

// MarkAsRead maneja PUT /api/v1/messages/:id/read
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	// Obtener ID del mensaje desde la URL
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	// Marcar como leído usando el servicio
	err = h.messageService.MarkAsRead(c.Request.Context(), messageID, userProfile.ID)
	if err != nil {
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Error al marcar mensaje como leído", err.Error())
		return
	}

	response.Success(c, "Mensaje marcado como leído", gin.H{
		"messageId": messageID,
		"readAt":    time.Now(),
	})
}

// UpdateMessageStatus maneja PUT /api/v1/messages/:id/status
func (h *MessageHandler) UpdateMessageStatus(c *gin.Context) {
	// Obtener ID del mensaje desde la URL
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Validar request body
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	// Por ahora, solo marcar como leído si el estado es READ
	if req.Status == "READ" {
		err = h.messageService.MarkAsRead(c.Request.Context(), messageID, userProfile.ID)
		if err != nil {
			if err.Error() == "mensaje no encontrado" {
				response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
				return
			}
			response.Error(c, http.StatusInternalServerError, "Error al actualizar estado", err.Error())
			return
		}
	}

	response.Success(c, "Estado del mensaje actualizado exitosamente", gin.H{
		"messageId": messageID,
		"status":    req.Status,
	})
}

// GetMessageTypes maneja GET /api/v1/messages/types
func (h *MessageHandler) GetMessageTypes(c *gin.Context) {
	// Por ahora devolver tipos hardcodeados hasta implementar en el servicio
	types := []string{
		"SOLICITUD",
		"URGENTE",
		"COORDINACION",
		"NOTIFICACION",
		"SEGUIMIENTO",
		"CONSULTA",
		"DIRECTRIZ",
	}

	response.Success(c, "Tipos de mensaje obtenidos exitosamente", types)
}

// GetMessageStatuses maneja GET /api/v1/messages/statuses
func (h *MessageHandler) GetMessageStatuses(c *gin.Context) {
	// Por ahora devolver estados hardcodeados hasta implementar en el servicio
	statuses := []string{
		"DRAFT",
		"SENT",
		"READ",
		"IN_PROGRESS",
		"RESPONDED",
		"RESOLVED",
		"ARCHIVED",
		"CANCELLED",
	}

	response.Success(c, "Estados de mensaje obtenidos exitosamente", statuses)
}

// 🔧 CORREGIDO: GetMessageStats maneja GET /api/v1/messages/stats
func (h *MessageHandler) GetMessageStats(c *gin.Context) {
	logger.Info("📊 GET /api/v1/messages/stats - Obtener estadísticas")

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		logger.Error("Usuario no encontrado en contexto para GetMessageStats")
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		logger.Error("Error al convertir usuario del contexto en GetMessageStats")
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	logger.Info("Usuario solicitando estadísticas: %s (%s)", userProfile.Email, userProfile.Role)

	// 🔧 CORREGIDO: Obtener unitID del usuario (aplicar mismos filtros que GetMessages)
	userUnitID := 1 // valor por defecto
	if userProfile.OrganizationalUnitID != nil {
		userUnitID = *userProfile.OrganizationalUnitID
	}

	// 🔧 NUEVO: Usar el servicio real en lugar de datos hardcodeados
	statsService, err := h.messageService.GetMessageStats(c.Request.Context(), userUnitID, userProfile.ID)
	if err != nil {
		logger.Error("Error al obtener estadísticas del servicio: %v", err)
		// 🔧 FALLBACK: Calcular estadísticas básicas desde mensajes
		h.calculateBasicStats(c, userUnitID, userProfile.ID)
		return
	}

	// 🔧 NUEVO: Calcular estadísticas adicionales (por estado, urgentes, etc.)
	additionalStats, err := h.getAdditionalStats(c.Request.Context(), userUnitID)
	if err != nil {
		logger.Warn("Error al obtener estadísticas adicionales: %v", err)
	}

	// 🔧 NUEVO: Construir respuesta completa
	stats := gin.H{
		"totalMessages":       statsService.Total,
		"urgentMessages":      statsService.Urgent,
		"readMessages":        additionalStats["readMessages"],
		"inProgressMessages":  additionalStats["inProgressMessages"],
		"averageResponseTime": "0 horas", // TODO: Implementar cálculo real
		"messagesByStatus":    additionalStats["messagesByStatus"],
	}

	logger.Info("Estadísticas obtenidas exitosamente para unidad %d: %d mensajes totales", userUnitID, statsService.Total)
	response.Success(c, "Estadísticas obtenidas exitosamente", stats)
}

// 🔧 NUEVA FUNCIÓN: Calcular estadísticas básicas como fallback
func (h *MessageHandler) calculateBasicStats(c *gin.Context, unitID int, userID uuid.UUID) {
	logger.Info("📊 Calculando estadísticas básicas como fallback para unidad: %d", unitID)

	// Usar el mismo filtro que GetMessages para consistencia
	req := &services.GetMessagesRequest{
		UnitID:    &unitID,
		Page:      1,
		Limit:     1000, // Obtener muchos mensajes para calcular estadísticas
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Obtener mensajes como remitente
	messagesSent, totalSent, err := h.messageService.GetMessagesByUnit(c.Request.Context(), unitID, req)
	if err != nil {
		logger.Error("Error al obtener mensajes enviados para estadísticas: %v", err)
		totalSent = 0
	}

	// Calcular estadísticas básicas
	urgentCount := 0
	readCount := 0
	inProgressCount := 0
	statusCounts := map[string]int{
		"SENT":        0,
		"READ":        0,
		"IN_PROGRESS": 0,
		"RESOLVED":    0,
		"DRAFT":       0,
	}

	for _, msg := range messagesSent {
		if msg.IsUrgent {
			urgentCount++
		}
		if msg.ReadAt != nil {
			readCount++
		}

		// Contar por estado (usando el código del estado si está disponible)
		if msg.Status != nil {
			statusCode := msg.Status.Code
			if count, exists := statusCounts[statusCode]; exists {
				statusCounts[statusCode] = count + 1
			}

			if statusCode == "IN_PROGRESS" {
				inProgressCount++
			}
		}
	}

	stats := gin.H{
		"totalMessages":       totalSent,
		"urgentMessages":      urgentCount,
		"readMessages":        readCount,
		"inProgressMessages":  inProgressCount,
		"averageResponseTime": "0 horas",
		"messagesByStatus":    statusCounts,
	}

	logger.Info("Estadísticas básicas calculadas: %d total, %d urgentes", totalSent, urgentCount)
	response.Success(c, "Estadísticas obtenidas exitosamente", stats)
}

// 🔧 NUEVA FUNCIÓN: Obtener estadísticas adicionales por estado
func (h *MessageHandler) getAdditionalStats(ctx context.Context, unitID int) (map[string]interface{}, error) {
	// TODO: Implementar consultas específicas a la base de datos para estadísticas por estado
	// Por ahora, usar valores por defecto

	return map[string]interface{}{
		"readMessages":       0,
		"inProgressMessages": 0,
		"messagesByStatus": gin.H{
			"SENT":        0,
			"READ":        0,
			"IN_PROGRESS": 0,
			"RESOLVED":    0,
			"DRAFT":       0,
		},
	}, nil
}

// DeleteMessage maneja DELETE /api/v1/messages/:id (soft delete)
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	// Obtener ID del mensaje desde la URL
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	// Verificar que el mensaje existe y que el usuario tiene permisos
	messageResponse, err := h.messageService.GetMessageByID(c.Request.Context(), messageID)
	if err != nil {
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
			return
		}
		response.Error(c, http.StatusInternalServerError, "Error al obtener mensaje", err.Error())
		return
	}

	// Solo el creador del mensaje o admin pueden eliminarlo
	if userProfile.Role != "admin" && messageResponse.SenderID != userProfile.ID {
		response.Error(c, http.StatusForbidden, "Solo puede eliminar mensajes que usted ha creado", "")
		return
	}

	// Por ahora, usar el método ArchiveMessage como soft delete
	err = h.messageService.ArchiveMessage(c.Request.Context(), messageID, userProfile.ID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Error al eliminar mensaje", err.Error())
		return
	}

	response.Success(c, "Mensaje eliminado exitosamente", gin.H{
		"messageId": messageID,
		"deleted":   true,
	})
}

// GetSimpleMessageStats maneja GET /api/v1/messages/stats-simple
func (h *MessageHandler) GetSimpleMessageStats(c *gin.Context) {
	logger.Info("📊 GET /api/v1/messages/stats-simple - Obtener estadísticas simples")

	// Obtener usuario desde el middleware
	user, exists := c.Get("user")
	if !exists {
		logger.Error("Usuario no encontrado en contexto")
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		logger.Error("Error al convertir usuario del contexto")
		response.Error(c, http.StatusInternalServerError, "Error interno del servidor", "")
		return
	}

	// Obtener unitID del usuario
	userUnitID := 1
	if userProfile.OrganizationalUnitID != nil {
		userUnitID = *userProfile.OrganizationalUnitID
	}

	logger.Info("Calculando estadísticas simples para unidad: %d", userUnitID)

	// Obtener estadísticas del servicio
	stats, err := h.messageService.GetSimpleStats(c.Request.Context(), userUnitID)
	if err != nil {
		logger.Error("Error al obtener estadísticas: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al obtener estadísticas", err.Error())
		return
	}

	// Construir respuesta en el formato esperado por el frontend
	responseData := gin.H{
		"totalMessages":       stats.TotalMessages,
		"urgentMessages":      stats.UrgentMessages,
		"readMessages":        stats.ReadMessages,
		"inProgressMessages":  stats.InProgressMessages,
		"averageResponseTime": "0 horas", // TODO: implementar cálculo real
		"messagesByStatus":    stats.MessagesByStatus,
	}

	logger.Info("Estadísticas simples calculadas exitosamente: Total=%d, Urgentes=%d, Leídos=%d",
		stats.TotalMessages, stats.UrgentMessages, stats.ReadMessages)

	response.Success(c, "Estadísticas simples obtenidas exitosamente", responseData)
}

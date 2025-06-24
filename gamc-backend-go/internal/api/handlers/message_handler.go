// internal/api/handlers/message_handler.go
package handlers

import (
	"net/http"
	"strconv"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/services"
	"gamc-backend-go/pkg/logger"
	"gamc-backend-go/pkg/response"
	"gamc-backend-go/pkg/validator"

	"github.com/gin-gonic/gin"
)

// MessageHandler maneja las peticiones HTTP relacionadas con mensajes
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler crea una nueva instancia del handler
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessage crea un nuevo mensaje
// @Summary Crear mensaje
// @Description Crea un nuevo mensaje entre unidades organizacionales
// @Tags messages
// @Accept json
// @Produce json
// @Param message body services.CreateMessageRequest true "Datos del mensaje"
// @Success 201 {object} response.APIResponse{data=services.MessageResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	logger.Info("📨 POST /api/v1/messages - Crear mensaje")

	var req services.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Error al parsear request: %v", err)
		response.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar datos
	if err := validator.Validate(&req); err != nil {
		logger.Error("Error de validación: %v", err)
		response.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Verificar que el usuario tenga rol de input o admin
	if userProfile.Role != "input" && userProfile.Role != "admin" {
		response.Error(c, http.StatusForbidden, "No tiene permisos para crear mensajes", "")
		return
	}

	// Asignar datos del usuario al request
	req.SenderID = userProfile.ID
	if userProfile.OrganizationalUnitID != nil {
		req.SenderUnitID = *userProfile.OrganizationalUnitID
	}

	// Crear mensaje
	message, err := h.messageService.CreateMessage(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Error al crear mensaje: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al crear mensaje", err.Error())
		return
	}

	logger.Info("✅ Mensaje creado exitosamente - ID: %d", message.ID)
	response.Success(c, "Mensaje creado exitosamente", message)
}

// GetMessages obtiene lista de mensajes con filtros
// @Summary Listar mensajes
// @Description Obtiene lista de mensajes con filtros y paginación
// @Tags messages
// @Accept json
// @Produce json
// @Param unitId query int false "ID de unidad organizacional"
// @Param messageType query int false "Tipo de mensaje"
// @Param status query int false "Estado del mensaje"
// @Param isUrgent query bool false "Solo mensajes urgentes"
// @Param dateFrom query string false "Fecha desde (YYYY-MM-DD)"
// @Param dateTo query string false "Fecha hasta (YYYY-MM-DD)"
// @Param searchText query string false "Texto a buscar"
// @Param page query int false "Página" default(1)
// @Param limit query int false "Elementos por página" default(20)
// @Param sortBy query string false "Ordenar por" Enums(created_at, subject, priority_level)
// @Param sortOrder query string false "Orden" Enums(asc, desc)
// @Success 200 {object} response.APIResponse{data=GetMessagesResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	logger.Info("📋 GET /api/v1/messages - Listar mensajes")

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Construir request de filtros
	req := &services.GetMessagesRequest{
		Page:      1,
		Limit:     20,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	// Parsear parámetros de query
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	if unitIDStr := c.Query("unitId"); unitIDStr != "" {
		if unitID, err := strconv.Atoi(unitIDStr); err == nil {
			req.UnitID = &unitID
		}
	}

	if messageTypeStr := c.Query("messageType"); messageTypeStr != "" {
		if messageType, err := strconv.Atoi(messageTypeStr); err == nil {
			req.MessageType = &messageType
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		if status, err := strconv.Atoi(statusStr); err == nil {
			req.Status = &status
		}
	}

	if isUrgentStr := c.Query("isUrgent"); isUrgentStr != "" {
		if isUrgent, err := strconv.ParseBool(isUrgentStr); err == nil {
			req.IsUrgent = &isUrgent
		}
	}

	if dateFromStr := c.Query("dateFrom"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			req.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("dateTo"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			// Agregar 23:59:59 para incluir todo el día
			endOfDay := dateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			req.DateTo = &endOfDay
		}
	}

	if searchText := c.Query("searchText"); searchText != "" {
		req.SearchText = &searchText
	}

	if sortBy := c.Query("sortBy"); sortBy != "" {
		if sortBy == "created_at" || sortBy == "subject" || sortBy == "priority_level" {
			req.SortBy = sortBy
		}
	}

	if sortOrder := c.Query("sortOrder"); sortOrder != "" {
		if sortOrder == "asc" || sortOrder == "desc" {
			req.SortOrder = sortOrder
		}
	}

	// Obtener mensajes según el rol del usuario
	var messages []services.MessageResponse
	var total int64
	var err error

	if req.SearchText != nil && *req.SearchText != "" {
		// Si hay texto de búsqueda, usar SearchMessages
		messages, total, err = h.messageService.SearchMessages(c.Request.Context(), *req.SearchText, userProfile.ID, req)
	} else if userProfile.Role == "admin" || req.UnitID != nil {
		// Si es admin o se especifica una unidad, obtener por unidad
		unitID := 0
		if userProfile.OrganizationalUnitID != nil {
			unitID = *userProfile.OrganizationalUnitID
		}
		if req.UnitID != nil && userProfile.Role == "admin" {
			unitID = *req.UnitID
		}
		messages, total, err = h.messageService.GetMessagesByUnit(c.Request.Context(), unitID, req)
	} else {
		// Por defecto, obtener mensajes del usuario
		messages, total, err = h.messageService.GetMessagesByUser(c.Request.Context(), userProfile.ID, req)
	}

	if err != nil {
		logger.Error("Error al obtener mensajes: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al obtener mensajes", err.Error())
		return
	}

	// Calcular paginación
	totalPages := (total + int64(req.Limit) - 1) / int64(req.Limit)

	responseData := GetMessagesResponse{
		Messages:    messages,
		Total:       total,
		Page:        req.Page,
		Limit:       req.Limit,
		TotalPages:  int(totalPages),
		HasNext:     req.Page < int(totalPages),
		HasPrevious: req.Page > 1,
	}

	response.Success(c, "Mensajes obtenidos exitosamente", responseData)
}

// GetMessageByID obtiene un mensaje específico por ID
// @Summary Obtener mensaje por ID
// @Description Obtiene un mensaje específico por su ID
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "ID del mensaje"
// @Success 200 {object} response.APIResponse{data=services.MessageResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessageByID(c *gin.Context) {
	logger.Info("🔍 GET /api/v1/messages/:id - Obtener mensaje por ID")

	// Parsear ID del mensaje
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Obtener mensaje
	message, err := h.messageService.GetMessageByID(c.Request.Context(), messageID)
	if err != nil {
		logger.Error("Error al obtener mensaje: %v", err)
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
		} else {
			response.Error(c, http.StatusInternalServerError, "Error al obtener mensaje", err.Error())
		}
		return
	}

	// Verificar permisos (solo puede ver mensajes de su unidad o si es admin)
	userUnitID := 0
	if userProfile.OrganizationalUnitID != nil {
		userUnitID = *userProfile.OrganizationalUnitID
	}

	if userProfile.Role != "admin" &&
		message.SenderUnitID != userUnitID &&
		message.ReceiverUnitID != userUnitID {
		response.Error(c, http.StatusForbidden, "No tiene permisos para acceder a este mensaje", "")
		return
	}

	response.Success(c, "Mensaje obtenido exitosamente", message)
}

// MarkAsRead marca un mensaje como leído
// @Summary Marcar mensaje como leído
// @Description Marca un mensaje como leído por el usuario actual
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "ID del mensaje"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/{id}/read [put]
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	logger.Info("👁️ PUT /api/v1/messages/:id/read - Marcar como leído")

	// Parsear ID del mensaje
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Marcar como leído
	err = h.messageService.MarkAsRead(c.Request.Context(), messageID, userProfile.ID)
	if err != nil {
		logger.Error("Error al marcar mensaje como leído: %v", err)
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
		} else if err.Error() == "no tiene permisos para acceder a este mensaje" {
			response.Error(c, http.StatusForbidden, "No tiene permisos para acceder a este mensaje", "")
		} else {
			response.Error(c, http.StatusInternalServerError, "Error al marcar mensaje como leído", err.Error())
		}
		return
	}

	response.Success(c, "Mensaje marcado como leído exitosamente", nil)
}

// UpdateMessageStatus actualiza el estado de un mensaje
// @Summary Actualizar estado de mensaje
// @Description Actualiza el estado de un mensaje (IN_PROGRESS, RESPONDED, RESOLVED, etc.)
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "ID del mensaje"
// @Param status body UpdateStatusRequest true "Nuevo estado"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/{id}/status [put]
func (h *MessageHandler) UpdateMessageStatus(c *gin.Context) {
	logger.Info("🔄 PUT /api/v1/messages/:id/status - Actualizar estado")

	// Parsear ID del mensaje
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Datos inválidos", err.Error())
		return
	}

	// Validar estado
	validStatuses := []string{"IN_PROGRESS", "RESPONDED", "RESOLVED", "ARCHIVED", "CANCELLED"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		response.Error(c, http.StatusBadRequest, "Estado inválido", "Estados válidos: IN_PROGRESS, RESPONDED, RESOLVED, ARCHIVED, CANCELLED")
		return
	}

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Actualizar estado
	err = h.messageService.UpdateMessageStatus(c.Request.Context(), messageID, req.Status, userProfile.ID)
	if err != nil {
		logger.Error("Error al actualizar estado: %v", err)
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
		} else if err.Error() == "no tiene permisos para cambiar el estado de este mensaje" {
			response.Error(c, http.StatusForbidden, "No tiene permisos para cambiar el estado de este mensaje", "")
		} else {
			response.Error(c, http.StatusInternalServerError, "Error al actualizar estado", err.Error())
		}
		return
	}

	response.Success(c, "Estado actualizado exitosamente", nil)
}

// DeleteMessage elimina un mensaje (soft delete)
// @Summary Eliminar mensaje
// @Description Elimina un mensaje (solo el creador o admin)
// @Tags messages
// @Accept json
// @Produce json
// @Param id path int true "ID del mensaje"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	logger.Info("🗑️ DELETE /api/v1/messages/:id - Eliminar mensaje")

	// Parsear ID del mensaje
	messageIDStr := c.Param("id")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "ID de mensaje inválido", "")
		return
	}

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Primero obtener el mensaje para verificar permisos
	message, err := h.messageService.GetMessageByID(c.Request.Context(), messageID)
	if err != nil {
		if err.Error() == "mensaje no encontrado" {
			response.Error(c, http.StatusNotFound, "Mensaje no encontrado", "")
		} else {
			response.Error(c, http.StatusInternalServerError, "Error al obtener mensaje", err.Error())
		}
		return
	}

	// Solo el creador del mensaje o admin pueden eliminarlo
	if userProfile.Role != "admin" && message.SenderID != userProfile.ID {
		response.Error(c, http.StatusForbidden, "Solo puede eliminar mensajes que usted ha creado", "")
		return
	}

	// Cambiar estado a CANCELLED (soft delete)
	err = h.messageService.UpdateMessageStatus(c.Request.Context(), messageID, "CANCELLED", userProfile.ID)
	if err != nil {
		logger.Error("Error al eliminar mensaje: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al eliminar mensaje", err.Error())
		return
	}

	response.Success(c, "Mensaje eliminado exitosamente", nil)
}

// GetMessageStats obtiene estadísticas de mensajes
// @Summary Obtener estadísticas de mensajes
// @Description Obtiene estadísticas de mensajes para la unidad del usuario
// @Tags messages
// @Accept json
// @Produce json
// @Param unitId query int false "ID de unidad (solo admin)"
// @Success 200 {object} response.APIResponse{data=services.MessageStats}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/stats [get]
func (h *MessageHandler) GetMessageStats(c *gin.Context) {
	logger.Info("📊 GET /api/v1/messages/stats - Obtener estadísticas")

	// Obtener usuario del contexto
	user, exists := c.Get("user")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Usuario no autenticado", "")
		return
	}

	userProfile, ok := user.(*models.UserProfile)
	if !ok {
		response.Error(c, http.StatusInternalServerError, "Error interno", "invalid user context")
		return
	}

	// Determinar unidad para estadísticas
	var unitID *int
	if unitIDStr := c.Query("unitId"); unitIDStr != "" && userProfile.Role == "admin" {
		if parsedUnitID, err := strconv.Atoi(unitIDStr); err == nil {
			unitID = &parsedUnitID
		}
	} else {
		// Para usuarios no admin, solo estadísticas de su unidad
		if userProfile.OrganizationalUnitID != nil {
			unitID = userProfile.OrganizationalUnitID
		}
	}

	// Obtener estadísticas
	stats, err := h.messageService.GetMessageStats(c.Request.Context(), unitID)
	if err != nil {
		logger.Error("Error al obtener estadísticas: %v", err)
		response.Error(c, http.StatusInternalServerError, "Error al obtener estadísticas", err.Error())
		return
	}

	response.Success(c, "Estadísticas obtenidas exitosamente", stats)
}

// GetMessageTypes obtiene todos los tipos de mensajes
// @Summary Obtener tipos de mensajes
// @Description Obtiene lista de todos los tipos de mensajes disponibles
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse{data=[]models.MessageType}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/types [get]
func (h *MessageHandler) GetMessageTypes(c *gin.Context) {
	logger.Info("📝 GET /api/v1/messages/types - Obtener tipos de mensajes")

	// Esta funcionalidad sería agregada al MessageService
	response.Success(c, "Funcionalidad en desarrollo", []string{
		"SOLICITUD", "URGENTE", "COORDINACION", "NOTIFICACION",
		"SEGUIMIENTO", "CONSULTA", "DIRECTRIZ",
	})
}

// GetMessageStatuses obtiene todos los estados de mensajes
// @Summary Obtener estados de mensajes
// @Description Obtiene lista de todos los estados de mensajes disponibles
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse{data=[]models.MessageStatus}
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Security BearerAuth
// @Router /api/v1/messages/statuses [get]
func (h *MessageHandler) GetMessageStatuses(c *gin.Context) {
	logger.Info("📊 GET /api/v1/messages/statuses - Obtener estados de mensajes")

	// Esta funcionalidad sería agregada al MessageService
	response.Success(c, "Funcionalidad en desarrollo", []string{
		"DRAFT", "SENT", "READ", "IN_PROGRESS",
		"RESPONDED", "RESOLVED", "ARCHIVED", "CANCELLED",
	})
}

// Estructuras de request/response

// UpdateStatusRequest representa el request para actualizar estado
type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// GetMessagesResponse representa la respuesta paginada de mensajes
type GetMessagesResponse struct {
	Messages    []services.MessageResponse `json:"messages"`
	Total       int64                      `json:"total"`
	Page        int                        `json:"page"`
	Limit       int                        `json:"limit"`
	TotalPages  int                        `json:"totalPages"`
	HasNext     bool                       `json:"hasNext"`
	HasPrevious bool                       `json:"hasPrevious"`
}

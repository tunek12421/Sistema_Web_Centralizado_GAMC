// internal/types/requests/message_requests.go
package requests

import (
	"mime/multipart"
)

// CreateMessageRequest estructura para crear un mensaje
type CreateMessageRequest struct {
	Subject        string                  `json:"subject" binding:"required,min=3,max=255"`
	Content        string                  `json:"content" binding:"required,min=10"`
	ReceiverUnitID int                     `json:"receiverUnitId" binding:"required,min=1"`
	MessageTypeID  int                     `json:"messageTypeId" binding:"required,min=1"`
	PriorityLevel  int                     `json:"priorityLevel" binding:"min=1,max=5"`
	IsUrgent       bool                    `json:"isUrgent"`
	Attachments    []*multipart.FileHeader `form:"attachments" swaggerignore:"true"`
}

// UpdateMessageRequest estructura para actualizar un mensaje
type UpdateMessageRequest struct {
	Subject       string `json:"subject,omitempty" binding:"omitempty,min=3,max=255"`
	Content       string `json:"content,omitempty" binding:"omitempty,min=10"`
	PriorityLevel int    `json:"priorityLevel,omitempty" binding:"omitempty,min=1,max=5"`
	IsUrgent      bool   `json:"isUrgent,omitempty"`
}

// UpdateMessageStatusRequest estructura para actualizar el estado
type UpdateMessageStatusRequest struct {
	StatusID int    `json:"statusId" binding:"required,min=1"`
	Notes    string `json:"notes,omitempty"`
}

// MessageFilterRequest estructura para filtrar mensajes
type MessageFilterRequest struct {
	SenderUnitID   int    `form:"senderUnitId,omitempty"`
	ReceiverUnitID int    `form:"receiverUnitId,omitempty"`
	MessageTypeID  int    `form:"messageTypeId,omitempty"`
	StatusID       int    `form:"statusId,omitempty"`
	PriorityLevel  int    `form:"priorityLevel,omitempty"`
	IsUrgent       *bool  `form:"isUrgent,omitempty"`
	IsRead         *bool  `form:"isRead,omitempty"`
	IsArchived     *bool  `form:"isArchived,omitempty"`
	DateFrom       string `form:"dateFrom,omitempty"`
	DateTo         string `form:"dateTo,omitempty"`
	Search         string `form:"search,omitempty"`
	SortBy         string `form:"sortBy,omitempty" binding:"omitempty,oneof=created_at updated_at subject priority"`
	SortOrder      string `form:"sortOrder,omitempty" binding:"omitempty,oneof=asc desc"`
	Page           int    `form:"page,omitempty" binding:"omitempty,min=1"`
	Limit          int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
}

// MessageResponseRequest estructura para responder a un mensaje
type MessageResponseRequest struct {
	Content       string                  `json:"content" binding:"required,min=10"`
	MessageTypeID int                     `json:"messageTypeId" binding:"required,min=1"`
	Attachments   []*multipart.FileHeader `form:"attachments" swaggerignore:"true"`
}

// MarkAsReadRequest estructura para marcar como leído
type MarkAsReadRequest struct {
	MessageIDs []int64 `json:"messageIds" binding:"required,min=1"`
}

// ArchiveMessagesRequest estructura para archivar mensajes
type ArchiveMessagesRequest struct {
	MessageIDs []int64 `json:"messageIds" binding:"required,min=1"`
	Archive    bool    `json:"archive"`
}

// ForwardMessageRequest estructura para reenviar un mensaje
type ForwardMessageRequest struct {
	ReceiverUnitIDs    []int  `json:"receiverUnitIds" binding:"required,min=1"`
	Notes              string `json:"notes,omitempty"`
	IncludeAttachments bool   `json:"includeAttachments"`
}

// MessageSearchRequest búsqueda avanzada de mensajes
type MessageSearchRequest struct {
	Query          string    `json:"query" binding:"required,min=2"`
	SearchIn       []string  `json:"searchIn,omitempty"` // subject, content, attachments
	Units          []int     `json:"units,omitempty"`
	MessageTypes   []int     `json:"messageTypes,omitempty"`
	Statuses       []int     `json:"statuses,omitempty"`
	DateRange      DateRange `json:"dateRange,omitempty"`
	HasAttachments *bool     `json:"hasAttachments,omitempty"`
	Page           int       `json:"page,omitempty" binding:"omitempty,min=1"`
	Limit          int       `json:"limit,omitempty" binding:"omitempty,min=1,max=100"`
}

// DateRange estructura para rangos de fecha
type DateRange struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}

// BulkMessageActionRequest estructura para acciones masivas
type BulkMessageActionRequest struct {
	MessageIDs []int64 `json:"messageIds" binding:"required,min=1"`
	Action     string  `json:"action" binding:"required,oneof=read unread archive unarchive delete"`
}

// MessageStatsRequest estructura para solicitar estadísticas
type MessageStatsRequest struct {
	UnitID   int    `form:"unitId,omitempty"`
	DateFrom string `form:"dateFrom,omitempty"`
	DateTo   string `form:"dateTo,omitempty"`
	GroupBy  string `form:"groupBy,omitempty" binding:"omitempty,oneof=day week month unit type status"`
}

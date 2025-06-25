// internal/types/responses/message_responses.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// MessageResponse respuesta detallada de un mensaje
type MessageResponse struct {
	ID              int64                `json:"id"`
	Subject         string               `json:"subject"`
	Content         string               `json:"content"`
	Sender          UserSummary          `json:"sender"`
	SenderUnit      OrganizationSummary  `json:"senderUnit"`
	ReceiverUnit    OrganizationSummary  `json:"receiverUnit"`
	MessageType     MessageTypeSummary   `json:"messageType"`
	Status          MessageStatusSummary `json:"status"`
	PriorityLevel   int                  `json:"priorityLevel"`
	IsUrgent        bool                 `json:"isUrgent"`
	ReadAt          *time.Time           `json:"readAt,omitempty"`
	RespondedAt     *time.Time           `json:"respondedAt,omitempty"`
	ArchivedAt      *time.Time           `json:"archivedAt,omitempty"`
	Attachments     []AttachmentSummary  `json:"attachments"`
	AttachmentCount int                  `json:"attachmentCount"`
	ResponseCount   int                  `json:"responseCount"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

// MessageListResponse respuesta para lista de mensajes
type MessageListResponse struct {
	Messages   []MessageResponse `json:"messages"`
	Pagination PaginationInfo    `json:"pagination"`
	Filters    MessageFilters    `json:"filters"`
}

// MessageStatsResponse estadísticas de mensajes
type MessageStatsResponse struct {
	TotalMessages       int64                `json:"totalMessages"`
	UnreadMessages      int64                `json:"unreadMessages"`
	UrgentMessages      int64                `json:"urgentMessages"`
	PendingMessages     int64                `json:"pendingMessages"`
	ArchivedMessages    int64                `json:"archivedMessages"`
	AverageResponseTime float64              `json:"averageResponseTime"` // minutos
	MessagesByType      []MessageTypeStats   `json:"messagesByType"`
	MessagesByStatus    []MessageStatusStats `json:"messagesByStatus"`
	MessagesByUnit      []MessageUnitStats   `json:"messagesByUnit"`
	TrendData           []MessageTrend       `json:"trendData"`
}

// UserSummary resumen de usuario
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"fullName"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar,omitempty"`
}

// OrganizationSummary resumen de unidad organizacional
type OrganizationSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// MessageTypeSummary resumen de tipo de mensaje
type MessageTypeSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Code  string `json:"code"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

// MessageStatusSummary resumen de estado de mensaje
type MessageStatusSummary struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Code  string `json:"code"`
	Color string `json:"color,omitempty"`
}

// AttachmentSummary resumen de archivo adjunto
type AttachmentSummary struct {
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"originalName"`
	FileSize     int64     `json:"fileSize"`
	MimeType     string    `json:"mimeType"`
	Url          string    `json:"url"`
	IsImage      bool      `json:"isImage"`
	ThumbnailUrl string    `json:"thumbnailUrl,omitempty"`
}

// PaginationInfo información de paginación
type PaginationInfo struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	PageSize    int   `json:"pageSize"`
	TotalItems  int64 `json:"totalItems"`
	HasPrevious bool  `json:"hasPrevious"`
	HasNext     bool  `json:"hasNext"`
}

// MessageFilters filtros aplicados
type MessageFilters struct {
	SenderUnit    *int    `json:"senderUnit,omitempty"`
	ReceiverUnit  *int    `json:"receiverUnit,omitempty"`
	MessageType   *int    `json:"messageType,omitempty"`
	Status        *int    `json:"status,omitempty"`
	PriorityLevel *int    `json:"priorityLevel,omitempty"`
	IsUrgent      *bool   `json:"isUrgent,omitempty"`
	DateFrom      *string `json:"dateFrom,omitempty"`
	DateTo        *string `json:"dateTo,omitempty"`
	Search        *string `json:"search,omitempty"`
}

// MessageTypeStats estadísticas por tipo
type MessageTypeStats struct {
	TypeID     int     `json:"typeId"`
	TypeName   string  `json:"typeName"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// MessageStatusStats estadísticas por estado
type MessageStatusStats struct {
	StatusID   int     `json:"statusId"`
	StatusName string  `json:"statusName"`
	Count      int64   `json:"count"`
	Percentage float64 `json:"percentage"`
}

// MessageUnitStats estadísticas por unidad
type MessageUnitStats struct {
	UnitID        int    `json:"unitId"`
	UnitName      string `json:"unitName"`
	SentCount     int64  `json:"sentCount"`
	ReceivedCount int64  `json:"receivedCount"`
	TotalCount    int64  `json:"totalCount"`
}

// MessageTrend tendencia de mensajes
type MessageTrend struct {
	Date          time.Time `json:"date"`
	Count         int64     `json:"count"`
	SentCount     int64     `json:"sentCount"`
	ReceivedCount int64     `json:"receivedCount"`
}

// MessageThreadResponse hilo de conversación
type MessageThreadResponse struct {
	OriginalMessage MessageResponse   `json:"originalMessage"`
	Responses       []MessageResponse `json:"responses"`
	TotalResponses  int               `json:"totalResponses"`
}

// MessageSearchResponse respuesta de búsqueda
type MessageSearchResponse struct {
	Results     []MessageSearchResult `json:"results"`
	TotalHits   int64                 `json:"totalHits"`
	SearchTime  float64               `json:"searchTime"` // milisegundos
	Suggestions []string              `json:"suggestions,omitempty"`
}

// MessageSearchResult resultado individual de búsqueda
type MessageSearchResult struct {
	Message    MessageResponse     `json:"message"`
	Score      float64             `json:"score"`
	Highlights map[string][]string `json:"highlights,omitempty"`
}

// MessageActivityResponse actividad reciente
type MessageActivityResponse struct {
	Activities []MessageActivity `json:"activities"`
	LastUpdate time.Time         `json:"lastUpdate"`
}

// MessageActivity actividad de mensaje
type MessageActivity struct {
	ID        int64       `json:"id"`
	MessageID int64       `json:"messageId"`
	Action    string      `json:"action"`
	User      UserSummary `json:"user"`
	Timestamp time.Time   `json:"timestamp"`
	Details   string      `json:"details,omitempty"`
}

// BulkActionResponse respuesta de acciones masivas
type BulkActionResponse struct {
	Successful  int                 `json:"successful"`
	Failed      int                 `json:"failed"`
	Total       int                 `json:"total"`
	FailedItems []BulkActionFailure `json:"failedItems,omitempty"`
}

// BulkActionFailure falla en acción masiva
type BulkActionFailure struct {
	MessageID int64  `json:"messageId"`
	Reason    string `json:"reason"`
}

// MessageExportResponse respuesta de exportación
type MessageExportResponse struct {
	FileID      uuid.UUID `json:"fileId"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	RecordCount int64     `json:"recordCount"`
	Format      string    `json:"format"`
	DownloadUrl string    `json:"downloadUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

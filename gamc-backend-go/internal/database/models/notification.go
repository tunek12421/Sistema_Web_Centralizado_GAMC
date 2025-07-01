// internal/database/models/notification.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationType define los tipos de notificación
type NotificationType string

const (
	NotificationTypeMessage      NotificationType = "message"
	NotificationTypeSystem       NotificationType = "system"
	NotificationTypeAlert        NotificationType = "alert"
	NotificationTypeReminder     NotificationType = "reminder"
	NotificationTypeAnnouncement NotificationType = "announcement"
)

const (
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationTypeSecurity   NotificationType     = "security"
)

// NotificationStatus define los estados de una notificación
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusDismissed NotificationStatus = "dismissed"
	NotificationStatusFailed    NotificationStatus = "failed"
)

// NotificationPriority define los niveles de prioridad
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityMedium NotificationPriority = "medium"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// Notification representa una notificación del sistema
type Notification struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID           uuid.UUID              `json:"userId" gorm:"type:uuid;not null;index"`
	Type             NotificationType       `json:"type" gorm:"type:varchar(50);not null;index"`
	Status           NotificationStatus     `json:"status" gorm:"type:varchar(20);not null;default:'pending';index"`
	Priority         NotificationPriority   `json:"priority" gorm:"type:varchar(20);not null;default:'medium'"`
	Title            string                 `json:"title" gorm:"size:255;not null"`
	Content          string                 `json:"content" gorm:"type:text;not null"`
	RelatedEntity    string                 `json:"relatedEntity,omitempty" gorm:"size:100"` // messages, users, etc.
	RelatedEntityID  string                 `json:"relatedEntityId,omitempty" gorm:"size:100"`
	ActionURL        string                 `json:"actionUrl,omitempty" gorm:"size:500"`
	IconType         string                 `json:"iconType,omitempty" gorm:"size:50"`
	ReadAt           *time.Time             `json:"readAt,omitempty"`
	DismissedAt      *time.Time             `json:"dismissedAt,omitempty"`
	ScheduledFor     *time.Time             `json:"scheduledFor,omitempty"`
	SentAt           *time.Time             `json:"sentAt,omitempty"`
	ExpiresAt        *time.Time             `json:"expiresAt,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt        time.Time              `json:"createdAt"`
	UpdatedAt        time.Time              `json:"updatedAt"`
	IsRead           bool                   `json:"isRead" gorm:"default:false;index"`
	RelatedMessageID *int64                 `json:"relatedMessageId,omitempty" gorm:"index"`

	// Relaciones
	User           *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
	RelatedMessage *Message `json:"relatedMessage,omitempty" gorm:"foreignKey:RelatedMessageID"`
}

// TableName especifica el nombre de la tabla
func (Notification) TableName() string {
	return "notifications"
}

// BeforeCreate hook para establecer valores por defecto
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Status == "" {
		n.Status = NotificationStatusPending
	}
	if n.Priority == "" {
		n.Priority = NotificationPriorityMedium
	}
	return nil
}

// MarkAsRead marca la notificación como leída
func (n *Notification) MarkAsRead(db *gorm.DB) error {
	now := time.Now()
	n.ReadAt = &now
	n.Status = NotificationStatusRead
	return db.Save(n).Error
}

// MarkAsSent marca la notificación como enviada
func (n *Notification) MarkAsSent(db *gorm.DB) error {
	now := time.Now()
	n.SentAt = &now
	n.Status = NotificationStatusSent
	return db.Save(n).Error
}

// Dismiss descarta la notificación
func (n *Notification) Dismiss(db *gorm.DB) error {
	now := time.Now()
	n.DismissedAt = &now
	n.Status = NotificationStatusDismissed
	return db.Save(n).Error
}

// IsExpired verifica si la notificación ha expirado
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// ShouldSendNow verifica si la notificación debe enviarse ahora
func (n *Notification) ShouldSendNow() bool {
	if n.Status != NotificationStatusPending {
		return false
	}
	if n.ScheduledFor == nil {
		return true
	}
	return time.Now().After(*n.ScheduledFor)
}

// NotificationChannel define los canales de notificación
type NotificationChannel struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	NotificationID uuid.UUID  `json:"notificationId" gorm:"type:uuid;not null;index"`
	ChannelType    string     `json:"channelType" gorm:"size:50;not null"` // email, websocket, sms
	Status         string     `json:"status" gorm:"size:20;not null"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	DeliveredAt    *time.Time `json:"deliveredAt,omitempty"`
	FailedAt       *time.Time `json:"failedAt,omitempty"`
	FailureReason  string     `json:"failureReason,omitempty" gorm:"type:text"`
	CreatedAt      time.Time  `json:"createdAt"`

	// Relaciones
	Notification *Notification `json:"notification,omitempty" gorm:"foreignKey:NotificationID"`
}

// TableName especifica el nombre de la tabla
func (NotificationChannel) TableName() string {
	return "notification_channels"
}

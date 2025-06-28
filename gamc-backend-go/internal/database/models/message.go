// internal/database/models/message.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType representa un tipo de mensaje
type MessageType struct {
	ID            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code          string    `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Name          string    `json:"name" gorm:"size:100;not null"`
	Description   *string   `json:"description,omitempty"`
	PriorityLevel int       `json:"priorityLevel" gorm:"default:3"`
	Color         string    `json:"color" gorm:"size:7;default:'#007bff'"`
	IsActive      bool      `json:"isActive" gorm:"default:true"`
	CreatedAt     time.Time `json:"createdAt"`
}

// MessageStatus representa un estado de mensaje
type MessageStatus struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code        string    `json:"code" gorm:"uniqueIndex;size:50;not null"`
	Name        string    `json:"name" gorm:"size:50;not null"`
	Description *string   `json:"description,omitempty"`
	Color       string    `json:"color" gorm:"size:7;default:'#6c757d'"`
	IsFinal     bool      `json:"isFinal" gorm:"default:false"`
	SortOrder   int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Message representa un mensaje del sistema
type Message struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Subject        string     `json:"subject" gorm:"size:255;not null"`
	Content        string     `json:"content" gorm:"type:text;not null"`
	SenderID       uuid.UUID  `json:"senderId" gorm:"type:uuid;not null;index"`
	SenderUnitID   int        `json:"senderUnitId" gorm:"not null;index"`
	ReceiverUnitID int        `json:"receiverUnitId" gorm:"not null;index"`
	MessageTypeID  int        `json:"messageTypeId" gorm:"not null;index"`
	StatusID       int        `json:"statusId" gorm:"not null;index"`
	PriorityLevel  int        `json:"priorityLevel" gorm:"default:3"`
	IsUrgent       bool       `json:"isUrgent" gorm:"default:false"`
	ReadAt         *time.Time `json:"readAt,omitempty"`
	RespondedAt    *time.Time `json:"respondedAt,omitempty"`
	ArchivedAt     *time.Time `json:"archivedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	// Relaciones
	Sender       *User               `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
	SenderUnit   *OrganizationalUnit `json:"senderUnit,omitempty" gorm:"foreignKey:SenderUnitID"`
	ReceiverUnit *OrganizationalUnit `json:"receiverUnit,omitempty" gorm:"foreignKey:ReceiverUnitID"`
	MessageType  *MessageType        `json:"messageType,omitempty" gorm:"foreignKey:MessageTypeID"`
	Status       *MessageStatus      `json:"status,omitempty" gorm:"foreignKey:StatusID"`
	Attachments  []MessageAttachment `json:"attachments,omitempty" gorm:"foreignKey:MessageID"`
}

// MessageAttachment representa un archivo adjunto
type MessageAttachment struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MessageID    int64     `json:"messageId" gorm:"not null;index"`
	OriginalName string    `json:"originalName" gorm:"size:255;not null"`
	FileName     string    `json:"fileName" gorm:"size:255;not null"`
	FilePath     string    `json:"filePath" gorm:"size:500;not null"`
	FileSize     int64     `json:"fileSize" gorm:"not null"`
	MimeType     *string   `json:"mimeType,omitempty" gorm:"size:100"`
	UploadedBy   uuid.UUID `json:"uploadedBy" gorm:"type:uuid;not null;index"`
	CreatedAt    time.Time `json:"createdAt"`

	// Relaciones
	Message  *Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	Uploader *User    `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy"`
}

// BeforeCreate hook para MessageAttachment
func (ma *MessageAttachment) BeforeCreate(tx *gorm.DB) error {
	if ma.ID == uuid.Nil {
		ma.ID = uuid.New()
	}
	return nil
}

// AuditLog representa un log de auditoría
type AuditLog struct {
	ID         int64                  `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID     *uuid.UUID             `json:"userId,omitempty" gorm:"type:uuid;index"`
	Action     string                 `json:"action" gorm:"size:50;not null"`
	Resource   string                 `json:"resource" gorm:"size:100;not null"`
	ResourceID *string                `json:"resourceId,omitempty" gorm:"size:100"`
	OldValues  map[string]interface{} `json:"oldValues,omitempty" gorm:"type:jsonb"`
	NewValues  map[string]interface{} `json:"newValues,omitempty" gorm:"type:jsonb"`
	IPAddress  *string                `json:"ipAddress,omitempty" gorm:"type:inet"`
	UserAgent  *string                `json:"userAgent,omitempty"`
	Result     string                 `json:"result" gorm:"size:20;default:'success'"`
	CreatedAt  time.Time              `json:"createdAt"`

	// Relaciones
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

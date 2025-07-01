// internal/database/models/message.go
package models

import (
	"time"

	"github.com/google/uuid"
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

// TableName especifica el nombre de la tabla para MessageType
func (MessageType) TableName() string {
	return "message_types"
}

// TableName especifica el nombre de la tabla para MessageStatus
func (MessageStatus) TableName() string {
	return "message_statuses"
}

// Message representa un mensaje en el sistema
type Message struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Subject        string     `json:"subject" gorm:"not null;index"`
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

// TableName especifica el nombre de la tabla
func (Message) TableName() string {
	return "messages"
}

// IsRead verifica si el mensaje ha sido leído
func (m *Message) IsRead() bool {
	return m.ReadAt != nil
}

// IsResponded verifica si el mensaje ha sido respondido
func (m *Message) IsResponded() bool {
	return m.RespondedAt != nil
}

// IsArchived verifica si el mensaje está archivado
func (m *Message) IsArchived() bool {
	return m.ArchivedAt != nil
}

// MarkAsRead marca el mensaje como leído
func (m *Message) MarkAsRead() {
	now := time.Now()
	m.ReadAt = &now
}

// MarkAsResponded marca el mensaje como respondido
func (m *Message) MarkAsResponded() {
	now := time.Now()
	m.RespondedAt = &now
}

// Archive archiva el mensaje
func (m *Message) Archive() {
	now := time.Now()
	m.ArchivedAt = &now
}

// Unarchive desarchiva el mensaje
func (m *Message) Unarchive() {
	m.ArchivedAt = nil
}

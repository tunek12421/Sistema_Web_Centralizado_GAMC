// internal/database/models/attachment.go
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageAttachment representa un archivo adjunto de un mensaje
// Mapea a la tabla 'message_attachments' en PostgreSQL
type MessageAttachment struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	MessageID    int64     `json:"messageId" gorm:"not null;index"`
	OriginalName string    `json:"originalName" gorm:"size:255;not null"`
	FileName     string    `json:"fileName" gorm:"size:255;not null"` // Nombre en MinIO
	FilePath     string    `json:"filePath" gorm:"size:500;not null"`
	FileSize     int64     `json:"fileSize" gorm:"not null"`
	MimeType     string    `json:"mimeType" gorm:"size:100"`
	UploadedBy   uuid.UUID `json:"uploadedBy" gorm:"type:uuid;not null"`
	CreatedAt    time.Time `json:"createdAt"`

	// Relaciones
	Message  *Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
	Uploader *User    `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy"`
}

// TableName especifica el nombre de la tabla
func (MessageAttachment) TableName() string {
	return "message_attachments"
}

// BeforeCreate hook para establecer valores por defecto
func (m *MessageAttachment) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// GetPublicURL genera la URL pública del archivo
func (m *MessageAttachment) GetPublicURL(minioEndpoint string) string {
	return fmt.Sprintf("%s/%s", minioEndpoint, m.FilePath)
}

// IsImage verifica si el archivo es una imagen
func (m *MessageAttachment) IsImage() bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/svg+xml", "image/bmp",
	}
	for _, imgType := range imageTypes {
		if m.MimeType == imgType {
			return true
		}
	}
	return false
}

// IsDocument verifica si el archivo es un documento
func (m *MessageAttachment) IsDocument() bool {
	docTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"application/rtf",
	}
	for _, docType := range docTypes {
		if m.MimeType == docType {
			return true
		}
	}
	return false
}

// GetHumanFileSize retorna el tamaño del archivo en formato legible
func (m *MessageAttachment) GetHumanFileSize() string {
	size := float64(m.FileSize)
	units := []string{"B", "KB", "MB", "GB"}
	unitIndex := 0

	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}

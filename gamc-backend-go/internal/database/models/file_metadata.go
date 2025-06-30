// internal/database/models/file_metadata.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileStatus define los estados de un archivo
type FileStatus string

const (
	FileStatusUploading   FileStatus = "uploading"
	FileStatusActive      FileStatus = "active"
	FileStatusArchived    FileStatus = "archived"
	FileStatusDeleted     FileStatus = "deleted"
	FileStatusQuarantined FileStatus = "quarantined" // Para archivos sospechosos
)

// FileCategory define las categorías de archivos
type FileCategory string

const (
	FileCategoryDocument     FileCategory = "document"
	FileCategoryImage        FileCategory = "image"
	FileCategoryVideo        FileCategory = "video"
	FileCategoryAudio        FileCategory = "audio"
	FileCategorySpreadsheet  FileCategory = "spreadsheet"
	FileCategoryPresentation FileCategory = "presentation"
	FileCategoryArchive      FileCategory = "archive"
	FileCategoryOther        FileCategory = "other"
)

// FileMetadata representa los metadatos de un archivo en el sistema
type FileMetadata struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BucketName      string                 `json:"bucketName" gorm:"size:100;not null;index"`
	ObjectName      string                 `json:"objectName" gorm:"size:255;not null"`
	OriginalName    string                 `json:"originalName" gorm:"size:255;not null"`
	FileSize        int64                  `json:"fileSize" gorm:"not null"`
	MimeType        string                 `json:"mimeType" gorm:"size:100"`
	Category        FileCategory           `json:"category" gorm:"type:varchar(50);not null;index"`
	Status          FileStatus             `json:"status" gorm:"type:varchar(20);not null;default:'active';index"`
	UploadedBy      uuid.UUID              `json:"uploadedBy" gorm:"type:uuid;not null;index"`
	OrganizationID  int                    `json:"organizationId" gorm:"not null;index"`
	Checksum        string                 `json:"checksum" gorm:"size:64"` // SHA256
	ContentType     string                 `json:"contentType" gorm:"size:100"`
	Tags            []string               `json:"tags,omitempty" gorm:"type:text[]"`
	Description     string                 `json:"description,omitempty" gorm:"type:text"`
	IsPublic        bool                   `json:"isPublic" gorm:"default:false"`
	ExpiresAt       *time.Time             `json:"expiresAt,omitempty"`
	LastAccessedAt  *time.Time             `json:"lastAccessedAt,omitempty"`
	AccessCount     int64                  `json:"accessCount" gorm:"default:0"`
	VirusScanStatus string                 `json:"virusScanStatus,omitempty" gorm:"size:50"`
	VirusScanDate   *time.Time             `json:"virusScanDate,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
	DeletedAt       gorm.DeletedAt         `json:"deletedAt,omitempty" gorm:"index"`

	// Relaciones
	Uploader     *User               `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy"`
	Organization *OrganizationalUnit `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// TableName especifica el nombre de la tabla
func (FileMetadata) TableName() string {
	return "file_metadata"
}

// BeforeCreate hook para establecer valores por defecto
func (f *FileMetadata) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.Status == "" {
		f.Status = FileStatusActive
	}
	if f.Category == "" {
		f.Category = f.DetermineCategory()
	}
	return nil
}

// DetermineCategory determina la categoría basada en el tipo MIME
func (f *FileMetadata) DetermineCategory() FileCategory {
	mimeCategories := map[string]FileCategory{
		"image/":             FileCategoryImage,
		"video/":             FileCategoryVideo,
		"audio/":             FileCategoryAudio,
		"application/pdf":    FileCategoryDocument,
		"application/msword": FileCategoryDocument,
		"application/vnd.openxmlformats-officedocument.wordprocessingml": FileCategoryDocument,
		"application/vnd.ms-excel":                                       FileCategorySpreadsheet,
		"application/vnd.openxmlformats-officedocument.spreadsheetml":    FileCategorySpreadsheet,
		"application/vnd.ms-powerpoint":                                  FileCategoryPresentation,
		"application/vnd.openxmlformats-officedocument.presentationml":   FileCategoryPresentation,
		"application/zip":   FileCategoryArchive,
		"application/x-rar": FileCategoryArchive,
		"application/x-7z":  FileCategoryArchive,
	}

	for prefix, category := range mimeCategories {
		if len(f.MimeType) >= len(prefix) && f.MimeType[:len(prefix)] == prefix {
			return category
		}
	}

	// Verificar tipos específicos
	for mimeType, category := range mimeCategories {
		if f.MimeType == mimeType {
			return category
		}
	}

	return FileCategoryOther
}

// IncrementAccessCount incrementa el contador de accesos
func (f *FileMetadata) IncrementAccessCount(db *gorm.DB) error {
	now := time.Now()
	f.LastAccessedAt = &now
	f.AccessCount++
	return db.Model(f).Updates(map[string]interface{}{
		"access_count":     f.AccessCount,
		"last_accessed_at": f.LastAccessedAt,
	}).Error
}

// IsExpired verifica si el archivo ha expirado
func (f *FileMetadata) IsExpired() bool {
	if f.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*f.ExpiresAt)
}

// GetFullPath retorna la ruta completa del archivo en MinIO
func (f *FileMetadata) GetFullPath() string {
	return f.BucketName + "/" + f.ObjectName
}

// FileAccessLog registro de accesos a archivos
type FileAccessLog struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	FileID     uuid.UUID `json:"fileId" gorm:"type:uuid;not null;index"`
	UserID     uuid.UUID `json:"userId" gorm:"type:uuid;not null;index"`
	Action     string    `json:"action" gorm:"size:50;not null"` // download, view, share
	IPAddress  string    `json:"ipAddress" gorm:"type:inet"`
	UserAgent  string    `json:"userAgent" gorm:"type:text"`
	AccessedAt time.Time `json:"accessedAt"`

	// Relaciones
	File *FileMetadata `json:"file,omitempty" gorm:"foreignKey:FileID"`
	User *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName especifica el nombre de la tabla
func (FileAccessLog) TableName() string {
	return "file_access_logs"
}

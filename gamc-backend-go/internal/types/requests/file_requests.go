// internal/types/requests/file_requests.go
package requests

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// FileUploadRequest estructura para subir archivos
type FileUploadRequest struct {
	File           *multipart.FileHeader `form:"file" binding:"required"`
	Description    string                `form:"description,omitempty"`
	Tags           []string              `form:"tags,omitempty"`
	IsPublic       bool                  `form:"isPublic"`
	ExpiresAt      *time.Time            `form:"expiresAt,omitempty"`
	OrganizationID int                   `form:"organizationId,omitempty"`
}

// MultiFileUploadRequest estructura para subir múltiples archivos
type MultiFileUploadRequest struct {
	Files          []*multipart.FileHeader `form:"files" binding:"required,min=1,max=10"`
	Description    string                  `form:"description,omitempty"`
	Tags           []string                `form:"tags,omitempty"`
	IsPublic       bool                    `form:"isPublic"`
	OrganizationID int                     `form:"organizationId,omitempty"`
}

// FileUpdateRequest estructura para actualizar metadatos de archivo
type FileUpdateRequest struct {
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	IsPublic    *bool      `json:"isPublic,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// FileFilterRequest estructura para filtrar archivos
type FileFilterRequest struct {
	BucketName     string    `form:"bucketName,omitempty"`
	Category       string    `form:"category,omitempty"`
	Status         string    `form:"status,omitempty"`
	OrganizationID int       `form:"organizationId,omitempty"`
	UploadedBy     uuid.UUID `form:"uploadedBy,omitempty"`
	IsPublic       *bool     `form:"isPublic,omitempty"`
	Tags           []string  `form:"tags,omitempty"`
	Search         string    `form:"search,omitempty"`
	MimeTypes      []string  `form:"mimeTypes,omitempty"`
	MinSize        int64     `form:"minSize,omitempty"`
	MaxSize        int64     `form:"maxSize,omitempty"`
	DateFrom       string    `form:"dateFrom,omitempty"`
	DateTo         string    `form:"dateTo,omitempty"`
	SortBy         string    `form:"sortBy,omitempty" binding:"omitempty,oneof=name size created_at updated_at access_count"`
	SortOrder      string    `form:"sortOrder,omitempty" binding:"omitempty,oneof=asc desc"`
	Page           int       `form:"page,omitempty" binding:"omitempty,min=1"`
	Limit          int       `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
}

// FileShareRequest estructura para compartir archivos
type FileShareRequest struct {
	FileIDs        []uuid.UUID `json:"fileIds" binding:"required,min=1"`
	RecipientUnits []int       `json:"recipientUnits,omitempty"`
	RecipientUsers []uuid.UUID `json:"recipientUsers,omitempty"`
	ExpiresAt      *time.Time  `json:"expiresAt,omitempty"`
	AllowDownload  bool        `json:"allowDownload"`
	Message        string      `json:"message,omitempty"`
}

// FileMoveRequest estructura para mover archivos
type FileMoveRequest struct {
	FileIDs      []uuid.UUID `json:"fileIds" binding:"required,min=1"`
	TargetBucket string      `json:"targetBucket" binding:"required"`
	TargetPath   string      `json:"targetPath,omitempty"`
}

// FileBulkActionRequest estructura para acciones masivas
type FileBulkActionRequest struct {
	FileIDs []uuid.UUID `json:"fileIds" binding:"required,min=1"`
	Action  string      `json:"action" binding:"required,oneof=archive restore delete quarantine"`
	Reason  string      `json:"reason,omitempty"`
}

// FileDownloadRequest estructura para solicitar descarga
type FileDownloadRequest struct {
	FileID        uuid.UUID `json:"fileId" binding:"required"`
	Format        string    `json:"format,omitempty"` // Para conversión de formato
	Quality       string    `json:"quality,omitempty" binding:"omitempty,oneof=low medium high original"`
	WatermarkText string    `json:"watermarkText,omitempty"`
}

// FilePreviewRequest estructura para solicitar preview
type FilePreviewRequest struct {
	FileID  uuid.UUID `json:"fileId" binding:"required"`
	Width   int       `json:"width,omitempty" binding:"omitempty,min=50,max=2000"`
	Height  int       `json:"height,omitempty" binding:"omitempty,min=50,max=2000"`
	Quality int       `json:"quality,omitempty" binding:"omitempty,min=10,max=100"`
	Format  string    `json:"format,omitempty" binding:"omitempty,oneof=jpeg png webp"`
}

// FileQuotaRequest estructura para gestión de cuotas
type FileQuotaRequest struct {
	OrganizationID int   `json:"organizationId" binding:"required"`
	QuotaBytes     int64 `json:"quotaBytes" binding:"required,min=0"`
	AlertThreshold int   `json:"alertThreshold,omitempty" binding:"omitempty,min=50,max=95"`
}

// FileStatisticsRequest estructura para solicitar estadísticas
type FileStatisticsRequest struct {
	OrganizationID int       `form:"organizationId,omitempty"`
	UserID         uuid.UUID `form:"userId,omitempty"`
	DateFrom       string    `form:"dateFrom,omitempty"`
	DateTo         string    `form:"dateTo,omitempty"`
	GroupBy        string    `form:"groupBy,omitempty" binding:"omitempty,oneof=day week month category mime_type organization"`
}

// FileCleanupRequest estructura para limpieza de archivos
type FileCleanupRequest struct {
	OlderThanDays int      `json:"olderThanDays,omitempty" binding:"omitempty,min=1"`
	Status        []string `json:"status,omitempty"`
	Categories    []string `json:"categories,omitempty"`
	MaxSizeBytes  int64    `json:"maxSizeBytes,omitempty"`
	DryRun        bool     `json:"dryRun"`
}

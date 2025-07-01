// internal/types/responses/file_responses.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// FileResponse respuesta detallada de un archivo
type FileResponse struct {
	ID              uuid.UUID              `json:"id"`
	BucketName      string                 `json:"bucketName"`
	ObjectName      string                 `json:"objectName"`
	OriginalName    string                 `json:"originalName"`
	FileSize        int64                  `json:"fileSize"`
	HumanSize       string                 `json:"humanSize"`
	MimeType        string                 `json:"mimeType"`
	Category        string                 `json:"category"`
	Status          string                 `json:"status"`
	Uploader        UserSummary            `json:"uploader"`
	Organization    OrganizationSummary    `json:"organization"`
	Checksum        string                 `json:"checksum,omitempty"`
	Tags            []string               `json:"tags"`
	Description     string                 `json:"description,omitempty"`
	IsPublic        bool                   `json:"isPublic"`
	ExpiresAt       *time.Time             `json:"expiresAt,omitempty"`
	LastAccessedAt  *time.Time             `json:"lastAccessedAt,omitempty"`
	AccessCount     int64                  `json:"accessCount"`
	VirusScanStatus string                 `json:"virusScanStatus,omitempty"`
	VirusScanDate   *time.Time             `json:"virusScanDate,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	ThumbnailUrl    string                 `json:"thumbnailUrl,omitempty"`
	PreviewUrl      string                 `json:"previewUrl,omitempty"`
	DownloadUrl     string                 `json:"downloadUrl"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`
}

// FileListResponse respuesta para lista de archivos
type FileListResponse struct {
	Files      []FileResponse `json:"files"`
	Pagination PaginationInfo `json:"pagination"`
}

// FileUploadResponse respuesta de carga de archivo
type FileUploadResponse struct {
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"originalName"`
	FileSize     int64     `json:"fileSize"`
	MimeType     string    `json:"mimeType"`
	UploadUrl    string    `json:"uploadUrl"`
	Success      bool      `json:"success"`
	Message      string    `json:"message,omitempty"`
}

// MultiFileUploadResponse respuesta de carga múltiple
type MultiFileUploadResponse struct {
	Successful  int                  `json:"successful"`
	Failed      int                  `json:"failed"`
	Total       int                  `json:"total"`
	Files       []FileUploadResponse `json:"files"`
	FailedFiles []FileUploadFailure  `json:"failedFiles,omitempty"`
}

// FileUploadFailure falla en carga de archivo
type FileUploadFailure struct {
	FileName string `json:"fileName"`
	Reason   string `json:"reason"`
	Error    string `json:"error,omitempty"`
}

// FileStatsResponse estadísticas de archivos
type FileStatsResponse struct {
	TotalFiles      int64                   `json:"totalFiles"`
	TotalSize       int64                   `json:"totalSize"`
	HumanTotalSize  string                  `json:"humanTotalSize"`
	UsedSpace       int64                   `json:"usedSpace"`
	AvailableSpace  int64                   `json:"availableSpace"`
	QuotaUsage      float64                 `json:"quotaUsage"` // porcentaje
	FilesByCategory []FileCategoryStats     `json:"filesByCategory"`
	FilesByType     []FileMimeTypeStats     `json:"filesByType"`
	FilesByOrg      []FileOrganizationStats `json:"filesByOrg"`
	StorageTrend    []StorageTrend          `json:"storageTrend"`
	TopFiles        []FileResponse          `json:"topFiles"`
}

// FileCategoryStats estadísticas por categoría
type FileCategoryStats struct {
	Category   string  `json:"category"`
	Count      int64   `json:"count"`
	TotalSize  int64   `json:"totalSize"`
	Percentage float64 `json:"percentage"`
}

// FileMimeTypeStats estadísticas por tipo MIME
type FileMimeTypeStats struct {
	MimeType   string  `json:"mimeType"`
	Count      int64   `json:"count"`
	TotalSize  int64   `json:"totalSize"`
	Percentage float64 `json:"percentage"`
}

// FileOrganizationStats estadísticas por organización
type FileOrganizationStats struct {
	OrganizationID   int     `json:"organizationId"`
	OrganizationName string  `json:"organizationName"`
	FileCount        int64   `json:"fileCount"`
	TotalSize        int64   `json:"totalSize"`
	QuotaUsed        float64 `json:"quotaUsed"`
}

// StorageTrend tendencia de almacenamiento
type StorageTrend struct {
	Date      time.Time `json:"date"`
	UsedSpace int64     `json:"usedSpace"`
	FileCount int64     `json:"fileCount"`
	Uploads   int64     `json:"uploads"`
	Deletions int64     `json:"deletions"`
}

// FileShareResponse respuesta de compartir archivo
type FileShareResponse struct {
	ShareID    uuid.UUID     `json:"shareId"`
	ShareUrl   string        `json:"shareUrl"`
	ShareCode  string        `json:"shareCode,omitempty"`
	Recipients []UserSummary `json:"recipients"`
	ExpiresAt  *time.Time    `json:"expiresAt,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
}

// FileVersionResponse respuesta de versiones de archivo
type FileVersionResponse struct {
	CurrentVersion FileVersion   `json:"currentVersion"`
	Versions       []FileVersion `json:"versions"`
	TotalVersions  int           `json:"totalVersions"`
}

// FileVersion versión de archivo
type FileVersion struct {
	VersionID  string      `json:"versionId"`
	FileSize   int64       `json:"fileSize"`
	Checksum   string      `json:"checksum"`
	UploadedBy UserSummary `json:"uploadedBy"`
	UploadedAt time.Time   `json:"uploadedAt"`
	Comment    string      `json:"comment,omitempty"`
	IsCurrent  bool        `json:"isCurrent"`
}

// FilePreviewResponse respuesta de vista previa
type FilePreviewResponse struct {
	PreviewUrl  string    `json:"previewUrl"`
	PreviewType string    `json:"previewType"` // image, pdf, video, text, etc.
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	Duration    int       `json:"duration,omitempty"`  // segundos para video/audio
	PageCount   int       `json:"pageCount,omitempty"` // para PDFs
	ExpiresAt   time.Time `json:"expiresAt"`
}

// FileQuotaResponse respuesta de cuota
type FileQuotaResponse struct {
	OrganizationID int     `json:"organizationId"`
	QuotaBytes     int64   `json:"quotaBytes"`
	UsedBytes      int64   `json:"usedBytes"`
	AvailableBytes int64   `json:"availableBytes"`
	UsagePercent   float64 `json:"usagePercent"`
	AlertThreshold int     `json:"alertThreshold"`
	IsOverQuota    bool    `json:"isOverQuota"`
	IsNearQuota    bool    `json:"isNearQuota"`
}

// FileAccessLogResponse registro de accesos
type FileAccessLogResponse struct {
	Logs      []FileAccessEntry `json:"logs"`
	TotalLogs int64             `json:"totalLogs"`
	DateRange DateRange         `json:"dateRange"`
}

// FileAccessEntry entrada de acceso
type FileAccessEntry struct {
	User       UserSummary `json:"user"`
	Action     string      `json:"action"`
	IPAddress  string      `json:"ipAddress"`
	UserAgent  string      `json:"userAgent,omitempty"`
	AccessedAt time.Time   `json:"accessedAt"`
}

// DateRange rango de fechas
type DateRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// FileCleanupResponse respuesta de limpieza
type FileCleanupResponse struct {
	FilesDeleted   int64               `json:"filesDeleted"`
	SpaceRecovered int64               `json:"spaceRecovered"`
	HumanSpace     string              `json:"humanSpace"`
	DryRun         bool                `json:"dryRun"`
	Details        []FileCleanupDetail `json:"details,omitempty"`
}

// FileCleanupDetail detalle de limpieza
type FileCleanupDetail struct {
	FileID      uuid.UUID `json:"fileId"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	Reason      string    `json:"reason"`
	WouldDelete bool      `json:"wouldDelete"`
}

// internal/config/minio.go
package config

import (
	"fmt"
	"strings"
	"time"
)

// MinIOConfig contiene la configuración específica de MinIO
type MinIOConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	UseSSL          bool
	Region          string
	Buckets         MinIOBuckets
	MaxFileSize     int64
	AllowedTypes    []string
	PresignedExpiry int // segundos
}

// FileCategory representa las categorías de archivos
type FileCategory string

const (
	FileCategoryAttachment FileCategory = "attachment"
	FileCategoryDocument   FileCategory = "document"
	FileCategoryImage      FileCategory = "image"
	FileCategoryReport     FileCategory = "report"
	FileCategoryTemp       FileCategory = "temp"
	FileCategoryBackup     FileCategory = "backup"
)

// AllowedMIMETypes define los tipos MIME permitidos por categoría
var AllowedMIMETypes = map[FileCategory][]string{
	FileCategoryAttachment: {
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
		"application/zip",
		"application/x-rar-compressed",
	},
	FileCategoryDocument: {
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.oasis.opendocument.text",
	},
	FileCategoryImage: {
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/svg+xml",
		"image/webp",
	},
	FileCategoryReport: {
		"application/pdf",
		"text/csv",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	},
}

// ExtensionToMIME mapea extensiones a tipos MIME
var ExtensionToMIME = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".txt":  "text/plain",
	".zip":  "application/zip",
	".rar":  "application/x-rar-compressed",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".csv":  "text/csv",
	".odt":  "application/vnd.oasis.opendocument.text",
}

// GetBucketForCategory retorna el bucket correspondiente a una categoría
func (c *MinIOConfig) GetBucketForCategory(category FileCategory) (string, error) {
	switch category {
	case FileCategoryAttachment:
		return c.Buckets.Attachments, nil
	case FileCategoryDocument:
		return c.Buckets.Documents, nil
	case FileCategoryImage:
		return c.Buckets.Images, nil
	case FileCategoryReport:
		return c.Buckets.Reports, nil
	case FileCategoryTemp:
		return c.Buckets.Temp, nil
	case FileCategoryBackup:
		return c.Buckets.Backups, nil
	default:
		return "", fmt.Errorf("categoría de archivo no válida: %s", category)
	}
}

// DetermineFileCategory determina la categoría basada en el tipo MIME
func DetermineFileCategory(mimeType string) FileCategory {
	// Verificar imágenes primero (más específico)
	for _, mime := range AllowedMIMETypes[FileCategoryImage] {
		if mime == mimeType {
			return FileCategoryImage
		}
	}

	// Verificar reportes (CSV y Excel específicamente para reportes)
	if mimeType == "text/csv" || mimeType == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		return FileCategoryReport
	}

	// Verificar documentos
	for _, mime := range AllowedMIMETypes[FileCategoryDocument] {
		if mime == mimeType {
			return FileCategoryDocument
		}
	}

	// Por defecto, es un attachment
	return FileCategoryAttachment
}

// GetMIMEType obtiene el tipo MIME desde la extensión
func GetMIMEType(filename string) string {
	ext := strings.ToLower(getFileExtension(filename))
	if mimeType, ok := ExtensionToMIME[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// IsAllowedMIMEType verifica si un tipo MIME está permitido
func IsAllowedMIMEType(mimeType string, category FileCategory) bool {
	allowedTypes, exists := AllowedMIMETypes[category]
	if !exists {
		return false
	}

	for _, allowed := range allowedTypes {
		if allowed == mimeType {
			return true
		}
	}
	return false
}

// GenerateObjectKey genera una clave única para el objeto en MinIO
func GenerateObjectKey(category FileCategory, unitID int, filename string) string {
	timestamp := fmt.Sprintf("%d", timeNow().Unix())
	sanitizedFilename := sanitizeFilename(filename)

	// Estructura: categoria/unidad_id/año/mes/timestamp_filename
	return fmt.Sprintf("%s/unit_%d/%s/%s_%s",
		category,
		unitID,
		timeNow().Format("2006/01"),
		timestamp,
		sanitizedFilename,
	)
}

// Funciones auxiliares
func getFileExtension(filename string) string {
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 {
		return ""
	}
	return filename[lastDot:]
}

func sanitizeFilename(filename string) string {
	// Reemplazar caracteres no seguros
	replacer := strings.NewReplacer(
		" ", "_",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
		"#", "",
		"%", "",
		"&", "",
		"?", "",
		"=", "",
		"+", "",
		"*", "",
	)
	return replacer.Replace(filename)
}

// Para testing - permite mockear time.Now()
var timeNow = func() time.Time {
	return time.Now()
}

// pkg/minio/helpers.go
package minio

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// FileHelper funciones auxiliares para manejo de archivos
type FileHelper struct{}

// NewFileHelper crea una nueva instancia de FileHelper
func NewFileHelper() *FileHelper {
	return &FileHelper{}
}

// GenerateUniqueFilename genera un nombre de archivo único
func (fh *FileHelper) GenerateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	nameWithoutExt := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().UnixNano()

	return fmt.Sprintf("%s_%d%s", sanitizeFilename(nameWithoutExt), timestamp, ext)
}

// GenerateHashedFilename genera un nombre de archivo usando hash
func (fh *FileHelper) GenerateHashedFilename(originalName string, content []byte) string {
	ext := filepath.Ext(originalName)
	hash := md5.Sum(content)
	hashStr := hex.EncodeToString(hash[:])

	return fmt.Sprintf("%s%s", hashStr, ext)
}

// DetectContentType detecta el tipo de contenido de un archivo
func (fh *FileHelper) DetectContentType(filename string, content []byte) string {
	// Primero intentar por extensión
	ext := strings.ToLower(filepath.Ext(filename))
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}

	// Si no, detectar por contenido
	if len(content) > 512 {
		return http.DetectContentType(content[:512])
	}
	return http.DetectContentType(content)
}

// FormatFileSize formatea el tamaño del archivo en formato legible
func (fh *FileHelper) FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// CalculateChunkSize calcula el tamaño óptimo de chunk para subidas
func (fh *FileHelper) CalculateChunkSize(fileSize int64) int64 {
	const (
		minChunkSize = 5 * 1024 * 1024   // 5MB
		maxChunkSize = 100 * 1024 * 1024 // 100MB
		maxParts     = 10000             // Límite de MinIO
	)

	// Calcular tamaño de chunk basado en el tamaño del archivo
	chunkSize := fileSize / maxParts

	if chunkSize < minChunkSize {
		return minChunkSize
	}
	if chunkSize > maxChunkSize {
		return maxChunkSize
	}

	return chunkSize
}

// IsImageFile verifica si un archivo es una imagen por su extensión
func (fh *FileHelper) IsImageFile(filename string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".ico"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// IsDocumentFile verifica si un archivo es un documento
func (fh *FileHelper) IsDocumentFile(filename string) bool {
	docExts := []string{".pdf", ".doc", ".docx", ".odt", ".rtf", ".txt"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, docExt := range docExts {
		if ext == docExt {
			return true
		}
	}
	return false
}

// IsCompressedFile verifica si un archivo está comprimido
func (fh *FileHelper) IsCompressedFile(filename string) bool {
	compressedExts := []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2"}
	ext := strings.ToLower(filepath.Ext(filename))

	for _, compExt := range compressedExts {
		if ext == compExt {
			return true
		}
	}
	return false
}

// sanitizeFilename limpia un nombre de archivo de caracteres problemáticos
func sanitizeFilename(filename string) string {
	// Reemplazar caracteres problemáticos
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"#", "_",
		"%", "_",
		"&", "_",
		"{", "_",
		"}", "_",
		"$", "_",
		"!", "_",
		"@", "_",
		"+", "_",
		"`", "_",
		"=", "_",
	)

	cleaned := replacer.Replace(filename)

	// Eliminar múltiples guiones bajos consecutivos
	for strings.Contains(cleaned, "__") {
		cleaned = strings.ReplaceAll(cleaned, "__", "_")
	}

	// Eliminar guiones bajos al inicio y final
	cleaned = strings.Trim(cleaned, "_")

	// Si el nombre queda vacío, usar un nombre por defecto
	if cleaned == "" {
		cleaned = "file"
	}

	return cleaned
}

// PathHelper funciones auxiliares para rutas
type PathHelper struct{}

// NewPathHelper crea una nueva instancia de PathHelper
func NewPathHelper() *PathHelper {
	return &PathHelper{}
}

// GenerateObjectPath genera una ruta estructurada para objetos
func (ph *PathHelper) GenerateObjectPath(category, unitCode string, date time.Time) string {
	year := date.Format("2006")
	month := date.Format("01")
	day := date.Format("02")

	return fmt.Sprintf("%s/%s/%s/%s/%s", category, unitCode, year, month, day)
}

// GenerateBackupPath genera una ruta para backups
func (ph *PathHelper) GenerateBackupPath(backupType string, date time.Time) string {
	return fmt.Sprintf("backups/%s/%s/%s",
		backupType,
		date.Format("2006/01"),
		date.Format("backup_20060102_150405"),
	)
}

// GenerateTempPath genera una ruta temporal
func (ph *PathHelper) GenerateTempPath(sessionID string) string {
	return fmt.Sprintf("temp/%s/%d", sessionID, time.Now().UnixNano())
}

// ExtractMetadataFromPath extrae información de una ruta de objeto
func (ph *PathHelper) ExtractMetadataFromPath(objectPath string) map[string]string {
	parts := strings.Split(objectPath, "/")
	metadata := make(map[string]string)

	if len(parts) >= 1 {
		metadata["category"] = parts[0]
	}
	if len(parts) >= 2 {
		metadata["unit"] = parts[1]
	}
	if len(parts) >= 3 {
		metadata["year"] = parts[2]
	}
	if len(parts) >= 4 {
		metadata["month"] = parts[3]
	}
	if len(parts) >= 5 {
		metadata["day"] = parts[4]
	}

	return metadata
}

// ValidationHelper funciones de validación
type ValidationHelper struct {
	maxFileSize  int64
	allowedTypes []string
}

// NewValidationHelper crea una nueva instancia de ValidationHelper
func NewValidationHelper(maxFileSize int64, allowedTypes []string) *ValidationHelper {
	return &ValidationHelper{
		maxFileSize:  maxFileSize,
		allowedTypes: allowedTypes,
	}
}

// ValidateFileSize valida el tamaño del archivo
func (vh *ValidationHelper) ValidateFileSize(size int64) error {
	if size <= 0 {
		return fmt.Errorf("el tamaño del archivo debe ser mayor a 0")
	}
	if size > vh.maxFileSize {
		return fmt.Errorf("el archivo excede el tamaño máximo permitido de %s",
			NewFileHelper().FormatFileSize(vh.maxFileSize))
	}
	return nil
}

// ValidateFileExtension valida la extensión del archivo
func (vh *ValidationHelper) ValidateFileExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("el archivo no tiene extensión")
	}

	for _, allowed := range vh.allowedTypes {
		if ext == allowed {
			return nil
		}
	}

	return fmt.Errorf("tipo de archivo no permitido: %s", ext)
}

// CalculateETag calcula el ETag de un contenido
func CalculateETag(content []byte) string {
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// CalculateETagFromReader calcula el ETag desde un reader
func CalculateETagFromReader(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GeneratePresignedURLPath genera la ruta para URLs pre-firmadas
func GeneratePresignedURLPath(bucketName, objectKey string) string {
	return fmt.Sprintf("/%s/%s", bucketName, objectKey)
}

// ParsePresignedURL extrae información de una URL pre-firmada
func ParsePresignedURL(url string) (bucket, object string, err error) {
	// Implementación básica - en producción sería más robusta
	parts := strings.Split(url, "/")
	if len(parts) < 5 {
		return "", "", fmt.Errorf("URL inválida")
	}

	// Asumir formato: http(s)://endpoint/bucket/object
	bucket = parts[3]
	object = strings.Join(parts[4:], "/")

	return bucket, object, nil
}

// internal/integrations/minio/upload.go
package minio

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

// UploadOptions opciones para la subida de archivos
type UploadOptions struct {
	BucketName   string
	ObjectKey    string
	ContentType  string
	Metadata     map[string]string
	CacheControl string
	Progress     chan UploadProgress
}

// UploadProgress información del progreso de subida
type UploadProgress struct {
	TotalBytes      int64
	UploadedBytes   int64
	PercentComplete float64
}

// UploadResult resultado de una operación de subida
type UploadResult struct {
	ObjectKey    string    `json:"objectKey"`
	BucketName   string    `json:"bucketName"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"contentType"`
	ETag         string    `json:"etag"`
	UploadedAt   time.Time `json:"uploadedAt"`
	PresignedURL string    `json:"presignedUrl,omitempty"`
}

// UploadFile sube un archivo desde multipart.FileHeader
func (c *Client) UploadFile(ctx context.Context, file *multipart.FileHeader, category config.FileCategory, unitID int) (*UploadResult, error) {
	// Validar tamaño del archivo
	if file.Size > c.config.MinIOMaxFileSize {
		return nil, fmt.Errorf("el archivo excede el tamaño máximo permitido de %d MB", c.config.MinIOMaxFileSize/(1024*1024))
	}

	// Validar tipo de archivo
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !c.config.IsFileTypeAllowed(ext) {
		return nil, fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	// Determinar el bucket según la categoría
	bucketName, err := c.GetBucketForCategory(category)
	if err != nil {
		return nil, err
	}

	// Generar clave del objeto
	objectKey := c.GenerateObjectKey(category, unitID, file.Filename)

	// Abrir el archivo
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo: %w", err)
	}
	defer src.Close()

	// Determinar content type
	contentType := config.GetMIMEType(file.Filename)

	// Preparar opciones de subida
	opts := UploadOptions{
		BucketName:  bucketName,
		ObjectKey:   objectKey,
		ContentType: contentType,
		Metadata: map[string]string{
			"original-name": file.Filename,
			"uploaded-by":   fmt.Sprintf("unit-%d", unitID),
			"category":      string(category),
		},
	}

	// Subir archivo
	return c.UploadFromReader(ctx, src, file.Size, opts)
}

// UploadFromReader sube un archivo desde un io.Reader
func (c *Client) UploadFromReader(ctx context.Context, reader io.Reader, size int64, opts UploadOptions) (*UploadResult, error) {
	// Validar opciones
	if opts.BucketName == "" || opts.ObjectKey == "" {
		return nil, fmt.Errorf("bucket name y object key son requeridos")
	}

	// Preparar opciones de MinIO
	putOpts := minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	}

	if opts.CacheControl != "" {
		putOpts.CacheControl = opts.CacheControl
	}

	// Crear un reader con progreso si se proporciona canal
	var uploadReader io.Reader = reader
	if opts.Progress != nil {
		uploadReader = &progressReader{
			reader:   reader,
			total:    size,
			progress: opts.Progress,
		}
	}

	// Subir objeto
	info, err := c.client.PutObject(ctx, opts.BucketName, opts.ObjectKey, uploadReader, size, putOpts)
	if err != nil {
		return nil, fmt.Errorf("error al subir archivo: %w", err)
	}

	logger.Info("✅ Archivo subido: %s/%s (%.2f MB)", opts.BucketName, opts.ObjectKey, float64(size)/(1024*1024))

	// Generar URL pre-firmada para acceso temporal
	presignedURL, err := c.GetPresignedURL(ctx, opts.BucketName, opts.ObjectKey, 24*time.Hour)
	if err != nil {
		logger.Error("Error al generar URL pre-firmada: %v", err)
	}

	return &UploadResult{
		ObjectKey:    opts.ObjectKey,
		BucketName:   opts.BucketName,
		Size:         size,
		ContentType:  opts.ContentType,
		ETag:         info.ETag,
		UploadedAt:   time.Now(),
		PresignedURL: presignedURL,
	}, nil
}

// UploadMultipleFiles sube múltiples archivos
func (c *Client) UploadMultipleFiles(ctx context.Context, files []*multipart.FileHeader, category config.FileCategory, unitID int) ([]*UploadResult, error) {
	var results []*UploadResult
	var errors []string

	for _, file := range files {
		result, err := c.UploadFile(ctx, file, category, unitID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", file.Filename, err))
			continue
		}
		results = append(results, result)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("errores en la subida de archivos: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// CreateMultipartUpload inicia una subida multipart para archivos grandes
func (c *Client) CreateMultipartUpload(ctx context.Context, bucketName, objectKey string, opts UploadOptions) (string, error) {
	// Implementación para subidas multipart de archivos grandes
	// Esto es útil para archivos > 100MB
	logger.Info("Iniciando subida multipart para %s/%s", bucketName, objectKey)

	// Por ahora retornamos un ID simulado
	// En producción, esto iniciaría una subida multipart real
	return uuid.New().String(), nil
}

// GetBucketForCategory obtiene el bucket correspondiente a una categoría
func (c *Client) GetBucketForCategory(category config.FileCategory) (string, error) {
	minioConfig := &config.MinIOConfig{
		Buckets: c.buckets,
	}
	return minioConfig.GetBucketForCategory(category)
}

// GenerateObjectKey genera una clave única para el objeto
func (c *Client) GenerateObjectKey(category config.FileCategory, unitID int, filename string) string {
	return config.GenerateObjectKey(category, unitID, filename)
}

// progressReader implementa io.Reader con reporte de progreso
type progressReader struct {
	reader   io.Reader
	total    int64
	uploaded int64
	progress chan UploadProgress
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.uploaded += int64(n)
		if pr.progress != nil {
			progress := UploadProgress{
				TotalBytes:      pr.total,
				UploadedBytes:   pr.uploaded,
				PercentComplete: float64(pr.uploaded) / float64(pr.total) * 100,
			}

			// Enviar progreso de forma no bloqueante
			select {
			case pr.progress <- progress:
			default:
			}
		}
	}
	return n, err
}

// ChunkedUpload realiza una subida por chunks para archivos muy grandes
func (c *Client) ChunkedUpload(ctx context.Context, reader io.Reader, size int64, opts UploadOptions, chunkSize int64) (*UploadResult, error) {
	if chunkSize == 0 {
		chunkSize = 5 * 1024 * 1024 // 5MB por defecto
	}

	// Para archivos pequeños, usar subida normal
	if size < chunkSize*2 {
		return c.UploadFromReader(ctx, reader, size, opts)
	}

	// TODO: Implementar subida por chunks para archivos grandes
	// Por ahora, delegamos a la subida normal
	return c.UploadFromReader(ctx, reader, size, opts)
}

// ValidateFileBeforeUpload valida un archivo antes de subirlo
func (c *Client) ValidateFileBeforeUpload(file *multipart.FileHeader, category config.FileCategory) error {
	// Validar tamaño
	if file.Size > c.config.MinIOMaxFileSize {
		return fmt.Errorf("archivo demasiado grande: %d bytes (máximo: %d bytes)",
			file.Size, c.config.MinIOMaxFileSize)
	}

	// Validar extensión
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !c.config.IsFileTypeAllowed(ext) {
		return fmt.Errorf("tipo de archivo no permitido: %s", ext)
	}

	// Validar tipo MIME
	contentType := config.GetMIMEType(file.Filename)
	if !config.IsAllowedMIMEType(contentType, category) {
		return fmt.Errorf("tipo MIME no permitido para la categoría %s: %s", category, contentType)
	}

	return nil
}

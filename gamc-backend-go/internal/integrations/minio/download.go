// internal/integrations/minio/download.go
package minio

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gamc-backend-go/pkg/logger"

	"github.com/minio/minio-go/v7"
)

// DownloadOptions opciones para la descarga de archivos
type DownloadOptions struct {
	BucketName string
	ObjectKey  string
	FilePath   string // Ruta local donde guardar el archivo (opcional)
	Progress   chan DownloadProgress
}

// DownloadProgress información del progreso de descarga
type DownloadProgress struct {
	TotalBytes      int64
	DownloadedBytes int64
	PercentComplete float64
}

// DownloadResult resultado de una operación de descarga
type DownloadResult struct {
	ObjectKey    string    `json:"objectKey"`
	BucketName   string    `json:"bucketName"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"contentType"`
	DownloadedAt time.Time `json:"downloadedAt"`
	FilePath     string    `json:"filePath,omitempty"`
}

// DownloadFile descarga un archivo a una ruta local
func (c *Client) DownloadFile(ctx context.Context, bucketName, objectKey, destPath string) (*DownloadResult, error) {
	// Verificar que el objeto existe
	info, err := c.StatObject(ctx, bucketName, objectKey)
	if err != nil {
		return nil, fmt.Errorf("objeto no encontrado: %w", err)
	}

	// Crear directorio si no existe
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio: %w", err)
	}

	// Descargar objeto
	err = c.client.FGetObject(ctx, bucketName, objectKey, destPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("error al descargar archivo: %w", err)
	}

	logger.Info("✅ Archivo descargado: %s/%s -> %s", bucketName, objectKey, destPath)

	return &DownloadResult{
		ObjectKey:    objectKey,
		BucketName:   bucketName,
		Size:         info.Size,
		ContentType:  info.ContentType,
		DownloadedAt: time.Now(),
		FilePath:     destPath,
	}, nil
}

// DownloadToReader descarga un archivo y retorna un io.ReadCloser
func (c *Client) DownloadToReader(ctx context.Context, bucketName, objectKey string) (io.ReadCloser, *ObjectInfo, error) {
	// Verificar que el objeto existe
	info, err := c.StatObject(ctx, bucketName, objectKey)
	if err != nil {
		return nil, nil, fmt.Errorf("objeto no encontrado: %w", err)
	}

	// Obtener objeto
	object, err := c.client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("error al obtener objeto: %w", err)
	}

	return object, info, nil
}

// DownloadToMemory descarga un archivo completo a memoria
func (c *Client) DownloadToMemory(ctx context.Context, bucketName, objectKey string) ([]byte, *ObjectInfo, error) {
	reader, info, err := c.DownloadToReader(ctx, bucketName, objectKey)
	if err != nil {
		return nil, nil, err
	}
	defer reader.Close()

	// Leer todo el contenido
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("error al leer archivo: %w", err)
	}

	return data, info, nil
}

// DownloadWithProgress descarga un archivo con reporte de progreso
func (c *Client) DownloadWithProgress(ctx context.Context, opts DownloadOptions) (*DownloadResult, error) {
	// Obtener información del objeto
	info, err := c.StatObject(ctx, opts.BucketName, opts.ObjectKey)
	if err != nil {
		return nil, fmt.Errorf("objeto no encontrado: %w", err)
	}

	// Si no se especifica ruta, usar nombre del objeto
	if opts.FilePath == "" {
		opts.FilePath = filepath.Base(opts.ObjectKey)
	}

	// Crear directorio si no existe
	dir := filepath.Dir(opts.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio: %w", err)
	}

	// Abrir archivo de destino
	file, err := os.Create(opts.FilePath)
	if err != nil {
		return nil, fmt.Errorf("error al crear archivo: %w", err)
	}
	defer file.Close()

	// Obtener objeto con reader
	reader, _, err := c.DownloadToReader(ctx, opts.BucketName, opts.ObjectKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Crear writer con progreso si se proporciona canal
	var writer io.Writer = file
	if opts.Progress != nil {
		writer = &progressWriter{
			writer:   file,
			total:    info.Size,
			progress: opts.Progress,
		}
	}

	// Copiar contenido
	written, err := io.Copy(writer, reader)
	if err != nil {
		os.Remove(opts.FilePath) // Limpiar archivo parcial
		return nil, fmt.Errorf("error al escribir archivo: %w", err)
	}

	if written != info.Size {
		os.Remove(opts.FilePath) // Limpiar archivo parcial
		return nil, fmt.Errorf("descarga incompleta: esperados %d bytes, descargados %d", info.Size, written)
	}

	logger.Info("✅ Archivo descargado con éxito: %s (%.2f MB)", opts.FilePath, float64(written)/(1024*1024))

	return &DownloadResult{
		ObjectKey:    opts.ObjectKey,
		BucketName:   opts.BucketName,
		Size:         written,
		ContentType:  info.ContentType,
		DownloadedAt: time.Now(),
		FilePath:     opts.FilePath,
	}, nil
}

// DownloadDirectory descarga todos los archivos de un directorio (prefijo)
func (c *Client) DownloadDirectory(ctx context.Context, bucketName, prefix, destDir string) ([]*DownloadResult, error) {
	// Listar objetos con el prefijo
	objects, err := c.ListObjects(ctx, bucketName, prefix, true)
	if err != nil {
		return nil, err
	}

	if len(objects) == 0 {
		return nil, fmt.Errorf("no se encontraron archivos con el prefijo: %s", prefix)
	}

	var results []*DownloadResult
	var errors []string

	for _, obj := range objects {
		// Calcular ruta de destino manteniendo la estructura
		relativePath := obj.Name[len(prefix):]
		if relativePath[0] == '/' {
			relativePath = relativePath[1:]
		}
		destPath := filepath.Join(destDir, relativePath)

		// Descargar archivo
		result, err := c.DownloadFile(ctx, bucketName, obj.Name, destPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", obj.Name, err))
			continue
		}
		results = append(results, result)
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("errores en la descarga: %s", joinErrors(errors))
	}

	return results, nil
}

// StreamFile transmite un archivo sin descargarlo completamente
func (c *Client) StreamFile(ctx context.Context, bucketName, objectKey string, writer io.Writer) (*ObjectInfo, error) {
	reader, info, err := c.DownloadToReader(ctx, bucketName, objectKey)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Copiar contenido al writer
	_, err = io.Copy(writer, reader)
	if err != nil {
		return nil, fmt.Errorf("error al transmitir archivo: %w", err)
	}

	return info, nil
}

// GetDirectDownloadURL genera una URL de descarga directa (si el bucket es público)
func (c *Client) GetDirectDownloadURL(bucketName, objectKey string) string {
	// Construir URL directa
	protocol := "http"
	if c.config.MinIOUseSSL {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", protocol, c.config.MinIOEndpoint, bucketName, objectKey)
}

// progressWriter implementa io.Writer con reporte de progreso
type progressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	progress   chan DownloadProgress
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.downloaded += int64(n)
		if pw.progress != nil {
			progress := DownloadProgress{
				TotalBytes:      pw.total,
				DownloadedBytes: pw.downloaded,
				PercentComplete: float64(pw.downloaded) / float64(pw.total) * 100,
			}

			// Enviar progreso de forma no bloqueante
			select {
			case pw.progress <- progress:
			default:
			}
		}
	}
	return n, err
}

// joinErrors une múltiples errores en un string
func joinErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	if len(errors) == 1 {
		return errors[0]
	}
	return fmt.Sprintf("%s (y %d errores más)", errors[0], len(errors)-1)
}

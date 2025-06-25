// internal/integrations/minio/client.go
package minio

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wrapper para MinIO client
type Client struct {
	client  *minio.Client
	config  *config.Config
	buckets config.MinIOBuckets
}

// NewClient crea una nueva instancia del cliente MinIO
func NewClient(cfg *config.Config) (*Client, error) {
	// Inicializar cliente MinIO
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
		Region: cfg.MinIORegion,
	})
	if err != nil {
		return nil, fmt.Errorf("error al crear cliente MinIO: %w", err)
	}

	// Crear wrapper
	client := &Client{
		client:  minioClient,
		config:  cfg,
		buckets: config.GetMinIOBuckets(),
	}

	// Verificar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.CheckConnection(ctx); err != nil {
		return nil, fmt.Errorf("error al verificar conexión MinIO: %w", err)
	}

	// Asegurar que los buckets existan
	if err := client.EnsureBuckets(ctx); err != nil {
		return nil, fmt.Errorf("error al verificar buckets: %w", err)
	}

	logger.Info("✅ Cliente MinIO inicializado correctamente")
	return client, nil
}

// CheckConnection verifica la conexión con MinIO
func (c *Client) CheckConnection(ctx context.Context) error {
	// Intentar listar buckets como prueba de conexión
	_, err := c.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("no se pudo conectar a MinIO: %w", err)
	}
	return nil
}

// EnsureBuckets verifica que todos los buckets necesarios existan
func (c *Client) EnsureBuckets(ctx context.Context) error {
	buckets := []struct {
		name   string
		policy string
	}{
		{c.buckets.Attachments, "private"},
		{c.buckets.Documents, "download"},
		{c.buckets.Images, "download"},
		{c.buckets.Backups, "private"},
		{c.buckets.Temp, "private"},
		{c.buckets.Reports, "download"},
	}

	for _, bucket := range buckets {
		exists, err := c.client.BucketExists(ctx, bucket.name)
		if err != nil {
			return fmt.Errorf("error al verificar bucket %s: %w", bucket.name, err)
		}

		if !exists {
			logger.Warn("Bucket %s no existe, creando...", bucket.name)
			if err := c.client.MakeBucket(ctx, bucket.name, minio.MakeBucketOptions{
				Region: c.config.MinIORegion,
			}); err != nil {
				return fmt.Errorf("error al crear bucket %s: %w", bucket.name, err)
			}
			logger.Info("✅ Bucket %s creado", bucket.name)
		}

		// Configurar política si es necesario
		if bucket.policy == "download" {
			if err := c.SetBucketPolicy(ctx, bucket.name, bucket.policy); err != nil {
				logger.Error("Error al configurar política para bucket %s: %v", bucket.name, err)
			}
		}
	}

	return nil
}

// SetBucketPolicy configura la política de acceso de un bucket
func (c *Client) SetBucketPolicy(ctx context.Context, bucketName, policy string) error {
	var policyJSON string

	switch policy {
	case "download":
		// Política para permitir descargas públicas
		policyJSON = fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)
	case "private":
		// Política privada (sin acceso público)
		policyJSON = ""
	default:
		return fmt.Errorf("política no válida: %s", policy)
	}

	if policyJSON != "" {
		return c.client.SetBucketPolicy(ctx, bucketName, policyJSON)
	}
	return nil
}

// GetClient retorna el cliente MinIO interno
func (c *Client) GetClient() *minio.Client {
	return c.client
}

// GetBuckets retorna la configuración de buckets
func (c *Client) GetBuckets() config.MinIOBuckets {
	return c.buckets
}

// GetPresignedURL genera una URL pre-firmada para acceso temporal
func (c *Client) GetPresignedURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = c.config.MinIOPresignedExpiry
	}

	url, err := c.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("error al generar URL pre-firmada: %w", err)
	}

	return url.String(), nil
}

// GetPresignedUploadURL genera una URL pre-firmada para subida
func (c *Client) GetPresignedUploadURL(ctx context.Context, bucketName, objectName string, expiry time.Duration) (string, error) {
	if expiry == 0 {
		expiry = c.config.MinIOPresignedExpiry
	}

	url, err := c.client.PresignedPutObject(ctx, bucketName, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("error al generar URL de subida pre-firmada: %w", err)
	}

	return url.String(), nil
}

// StatObject obtiene información sobre un objeto
func (c *Client) StatObject(ctx context.Context, bucketName, objectName string) (*ObjectInfo, error) {
	info, err := c.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("error al obtener información del objeto: %w", err)
	}

	return &ObjectInfo{
		Name:         info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		ETag:         info.ETag,
		LastModified: info.LastModified,
		UserMetadata: info.UserMetadata,
	}, nil
}

// RemoveObject elimina un objeto
func (c *Client) RemoveObject(ctx context.Context, bucketName, objectName string) error {
	err := c.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("error al eliminar objeto: %w", err)
	}
	return nil
}

// ListObjects lista objetos en un bucket con prefijo opcional
func (c *Client) ListObjects(ctx context.Context, bucketName, prefix string, recursive bool) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error al listar objetos: %w", object.Err)
		}
		objects = append(objects, ObjectInfo{
			Name:         object.Key,
			Size:         object.Size,
			ContentType:  object.ContentType,
			ETag:         object.ETag,
			LastModified: object.LastModified,
		})
	}

	return objects, nil
}

// CopyObject copia un objeto de un lugar a otro
func (c *Client) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcObject,
	}

	dstOpts := minio.CopyDestOptions{
		Bucket: dstBucket,
		Object: dstObject,
	}

	_, err := c.client.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("error al copiar objeto: %w", err)
	}

	return nil
}

// GetBucketStats obtiene estadísticas de un bucket
func (c *Client) GetBucketStats(ctx context.Context, bucketName string) (*BucketStats, error) {
	stats := &BucketStats{
		Name:        bucketName,
		ObjectCount: 0,
		TotalSize:   0,
	}

	objectCh := c.client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error al obtener estadísticas: %w", object.Err)
		}
		stats.ObjectCount++
		stats.TotalSize += object.Size
	}

	return stats, nil
}

// ObjectInfo información sobre un objeto en MinIO
type ObjectInfo struct {
	Name         string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	UserMetadata map[string]string
}

// BucketStats estadísticas de un bucket
type BucketStats struct {
	Name        string
	ObjectCount int64
	TotalSize   int64
}

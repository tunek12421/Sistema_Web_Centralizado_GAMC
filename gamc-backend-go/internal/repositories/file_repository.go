// internal/repositories/file_repository.go
package repositories

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileRepository maneja las operaciones de base de datos para archivos
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository crea una nueva instancia del repositorio de archivos
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// CreateAttachment crea un nuevo registro de archivo adjunto
func (r *FileRepository) CreateAttachment(ctx context.Context, attachment *models.MessageAttachment) error {
	return r.db.WithContext(ctx).Create(attachment).Error
}

// CreateMetadata crea un nuevo registro de metadatos de archivo
func (r *FileRepository) CreateMetadata(ctx context.Context, metadata *models.FileMetadata) error {
	return r.db.WithContext(ctx).Create(metadata).Error
}

// GetAttachmentByID obtiene un archivo adjunto por ID
func (r *FileRepository) GetAttachmentByID(ctx context.Context, id uuid.UUID) (*models.MessageAttachment, error) {
	var attachment models.MessageAttachment
	err := r.db.WithContext(ctx).
		Preload("Message").
		Preload("Uploader").
		Where("id = ?", id).
		First(&attachment).Error

	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// GetMetadataByID obtiene metadatos de archivo por ID
func (r *FileRepository) GetMetadataByID(ctx context.Context, id uuid.UUID) (*models.FileMetadata, error) {
	var metadata models.FileMetadata
	err := r.db.WithContext(ctx).
		Preload("UploadedBy").
		Where("id = ?", id).
		First(&metadata).Error

	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// GetAttachmentsByMessageID obtiene todos los archivos adjuntos de un mensaje
func (r *FileRepository) GetAttachmentsByMessageID(ctx context.Context, messageID int64) ([]*models.MessageAttachment, error) {
	var attachments []*models.MessageAttachment
	err := r.db.WithContext(ctx).
		Preload("Uploader").
		Where("message_id = ?", messageID).
		Order("created_at ASC").
		Find(&attachments).Error

	return attachments, err
}

// GetMetadataByObjectKey obtiene metadatos por clave de objeto MinIO
func (r *FileRepository) GetMetadataByObjectKey(ctx context.Context, objectKey string) (*models.FileMetadata, error) {
	var metadata models.FileMetadata
	err := r.db.WithContext(ctx).
		Where("object_key = ?", objectKey).
		First(&metadata).Error

	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

// UpdateAttachment actualiza un archivo adjunto
func (r *FileRepository) UpdateAttachment(ctx context.Context, attachment *models.MessageAttachment) error {
	return r.db.WithContext(ctx).Save(attachment).Error
}

// UpdateMetadata actualiza metadatos de archivo
func (r *FileRepository) UpdateMetadata(ctx context.Context, metadata *models.FileMetadata) error {
	return r.db.WithContext(ctx).Save(metadata).Error
}

// DeleteAttachment elimina un archivo adjunto
func (r *FileRepository) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.MessageAttachment{}, id).Error
}

// DeleteMetadata elimina metadatos de archivo
func (r *FileRepository) DeleteMetadata(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.FileMetadata{}, id).Error
}

// GetFilesByFilter obtiene archivos con filtros y paginación
func (r *FileRepository) GetFilesByFilter(ctx context.Context, filter *FileFilter) ([]*models.FileMetadata, int64, error) {
	var files []*models.FileMetadata
	var total int64

	query := r.db.WithContext(ctx).Model(&models.FileMetadata{})

	// Aplicar filtros
	query = r.applyFileFilters(query, filter)

	// Contar total antes de paginar
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Aplicar ordenamiento
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortDesc {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	// Aplicar paginación
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Precargar relaciones
	query = query.Preload("UploadedBy")

	// Ejecutar consulta
	if err := query.Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

// GetFilesByCategory obtiene archivos por categoría
func (r *FileRepository) GetFilesByCategory(ctx context.Context, category string) ([]*models.FileMetadata, error) {
	var files []*models.FileMetadata
	err := r.db.WithContext(ctx).
		Where("category = ?", category).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// GetFilesByUploader obtiene archivos subidos por un usuario
func (r *FileRepository) GetFilesByUploader(ctx context.Context, uploaderID uuid.UUID, limit int) ([]*models.FileMetadata, error) {
	var files []*models.FileMetadata
	query := r.db.WithContext(ctx).
		Where("uploaded_by = ?", uploaderID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&files).Error
	return files, err
}

// GetStorageStats obtiene estadísticas de almacenamiento
func (r *FileRepository) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	stats := &StorageStats{}

	// Total de archivos
	r.db.WithContext(ctx).Model(&models.FileMetadata{}).Count(&stats.TotalFiles)

	// Tamaño total
	r.db.WithContext(ctx).
		Model(&models.FileMetadata{}).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&stats.TotalSize)

	// Por categoría
	var categoryStats []struct {
		Category string
		Count    int64
		Size     int64
	}
	r.db.WithContext(ctx).
		Model(&models.FileMetadata{}).
		Select("category, COUNT(*) as count, COALESCE(SUM(file_size), 0) as size").
		Group("category").
		Scan(&categoryStats)

	stats.ByCategory = make(map[string]CategoryStats)
	for _, cs := range categoryStats {
		stats.ByCategory[cs.Category] = CategoryStats{
			Count: cs.Count,
			Size:  cs.Size,
		}
	}

	// Por tipo MIME
	var mimeStats []struct {
		MimeType string
		Count    int64
	}
	r.db.WithContext(ctx).
		Model(&models.FileMetadata{}).
		Select("mime_type, COUNT(*) as count").
		Group("mime_type").
		Order("count DESC").
		Limit(10).
		Scan(&mimeStats)

	stats.TopMimeTypes = make(map[string]int64)
	for _, ms := range mimeStats {
		stats.TopMimeTypes[ms.MimeType] = ms.Count
	}

	return stats, nil
}

// GetExpiredTempFiles obtiene archivos temporales expirados
func (r *FileRepository) GetExpiredTempFiles(ctx context.Context, expirationHours int) ([]*models.FileMetadata, error) {
	var files []*models.FileMetadata
	cutoffTime := time.Now().Add(-time.Duration(expirationHours) * time.Hour)

	err := r.db.WithContext(ctx).
		Where("category = ? AND created_at < ?", "temp", cutoffTime).
		Find(&files).Error

	return files, err
}

// BatchDeleteMetadata elimina múltiples registros de metadatos
func (r *FileRepository) BatchDeleteMetadata(ctx context.Context, ids []uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&models.FileMetadata{}).Error
}

// UpdateAccessCount incrementa el contador de accesos
func (r *FileRepository) UpdateAccessCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.FileMetadata{}).
		Where("id = ?", id).
		UpdateColumn("access_count", gorm.Expr("access_count + ?", 1)).
		UpdateColumn("last_accessed_at", time.Now()).Error
}

// GetMostAccessedFiles obtiene los archivos más accedidos
func (r *FileRepository) GetMostAccessedFiles(ctx context.Context, limit int) ([]*models.FileMetadata, error) {
	var files []*models.FileMetadata
	err := r.db.WithContext(ctx).
		Where("access_count > ?", 0).
		Order("access_count DESC").
		Limit(limit).
		Find(&files).Error
	return files, err
}

// SearchFiles busca archivos por nombre
func (r *FileRepository) SearchFiles(ctx context.Context, searchTerm string, filter *FileFilter) ([]*models.FileMetadata, int64, error) {
	if filter == nil {
		filter = &FileFilter{}
	}
	filter.SearchTerm = searchTerm
	return r.GetFilesByFilter(ctx, filter)
}

// applyFileFilters aplica los filtros a la consulta de archivos
func (r *FileRepository) applyFileFilters(query *gorm.DB, filter *FileFilter) *gorm.DB {
	if filter == nil {
		return query
	}

	// Filtro por categoría
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	// Filtro por tipo MIME
	if filter.MimeType != "" {
		query = query.Where("mime_type = ?", filter.MimeType)
	}

	// Filtro por uploader
	if filter.UploaderID != nil {
		query = query.Where("uploaded_by = ?", *filter.UploaderID)
	}

	// Filtro por unidad organizacional
	if filter.UnitID != nil {
		query = query.Where("unit_id = ?", *filter.UnitID)
	}

	// Filtro por tamaño mínimo
	if filter.MinSize > 0 {
		query = query.Where("file_size >= ?", filter.MinSize)
	}

	// Filtro por tamaño máximo
	if filter.MaxSize > 0 {
		query = query.Where("file_size <= ?", filter.MaxSize)
	}

	// Filtro por fecha de creación
	if filter.CreatedFrom != nil {
		query = query.Where("created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("created_at <= ?", *filter.CreatedTo)
	}

	// Búsqueda por texto
	if filter.SearchTerm != "" {
		searchPattern := fmt.Sprintf("%%%s%%", filter.SearchTerm)
		query = query.Where("original_name ILIKE ?", searchPattern)
	}

	return query
}

// FileFilter estructura para filtrar archivos
type FileFilter struct {
	Category    string
	MimeType    string
	UploaderID  *uuid.UUID
	UnitID      *int
	MinSize     int64
	MaxSize     int64
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	SearchTerm  string
	SortBy      string
	SortDesc    bool
	Limit       int
	Offset      int
}

// StorageStats estadísticas de almacenamiento
type StorageStats struct {
	TotalFiles   int64
	TotalSize    int64
	ByCategory   map[string]CategoryStats
	TopMimeTypes map[string]int64
}

// CategoryStats estadísticas por categoría
type CategoryStats struct {
	Count int64
	Size  int64
}

// internal/services/file_service.go
package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/integrations/minio"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileService maneja la lógica de negocio para archivos
type FileService struct {
	fileRepo    *repositories.FileRepository
	auditRepo   *repositories.AuditRepository
	minioClient *minio.Client
	config      *config.Config
	db          *gorm.DB
}

// NewFileService crea una nueva instancia del servicio
func NewFileService(db *gorm.DB, cfg *config.Config) (*FileService, error) {
	// Inicializar cliente MinIO
	minioClient, err := minio.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error al inicializar cliente MinIO: %w", err)
	}

	return &FileService{
		fileRepo:    repositories.NewFileRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
		minioClient: minioClient,
		config:      cfg,
		db:          db,
	}, nil
}

// UploadFileRequest representa la solicitud de carga de archivo
type UploadFileRequest struct {
	File      *multipart.FileHeader
	Category  config.FileCategory
	UnitID    int
	UserID    uuid.UUID
	MessageID *int64 // Opcional, si es adjunto de mensaje
	Tags      []string
	Metadata  map[string]string
}

// FileResponse representa la respuesta de un archivo
type FileResponse struct {
	ID           uuid.UUID         `json:"id"`
	OriginalName string            `json:"originalName"`
	FileName     string            `json:"fileName"`
	FileSize     int64             `json:"fileSize"`
	MimeType     string            `json:"mimeType"`
	Category     string            `json:"category"`
	DownloadURL  string            `json:"downloadUrl"`
	ThumbnailURL string            `json:"thumbnailUrl,omitempty"`
	UploadedBy   uuid.UUID         `json:"uploadedBy"`
	UploadedAt   time.Time         `json:"uploadedAt"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// UploadFile maneja la carga de un archivo
func (s *FileService) UploadFile(ctx context.Context, req *UploadFileRequest) (*FileResponse, error) {
	logger.Info("📤 Subiendo archivo: %s", req.File.Filename)

	// Validar archivo
	if err := s.minioClient.ValidateFileBeforeUpload(req.File, req.Category); err != nil {
		return nil, fmt.Errorf("validación de archivo fallida: %w", err)
	}

	// Subir archivo a MinIO
	uploadResult, err := s.minioClient.UploadFile(ctx, req.File, req.Category, req.UnitID)
	if err != nil {
		return nil, fmt.Errorf("error al subir archivo: %w", err)
	}

	// Crear registro de metadatos
	metadata := &models.FileMetadata{
		OriginalName: req.File.Filename,
		StoredName:   filepath.Base(uploadResult.ObjectKey),
		FilePath:     uploadResult.ObjectKey,
		BucketName:   uploadResult.BucketName,
		FileSize:     uploadResult.Size,
		MimeType:     uploadResult.ContentType,
		Category:     string(req.Category),
		UploadedBy:   req.UserID,
		UnitID:       req.UnitID,
		Tags:         req.Tags,
		Metadata:     req.Metadata,
		Checksum:     uploadResult.ETag,
		IsPublic:     s.isPublicCategory(req.Category),
	}

	if err := s.fileRepo.CreateMetadata(ctx, metadata); err != nil {
		// Intentar eliminar archivo de MinIO si falla el registro
		s.minioClient.RemoveObject(ctx, uploadResult.BucketName, uploadResult.ObjectKey)
		return nil, fmt.Errorf("error al guardar metadatos: %w", err)
	}

	// Si es un adjunto de mensaje, crear registro de attachment
	if req.MessageID != nil {
		attachment := &models.MessageAttachment{
			MessageID:    *req.MessageID,
			OriginalName: req.File.Filename,
			FileName:     metadata.StoredName,
			FilePath:     metadata.FilePath,
			FileSize:     metadata.FileSize,
			MimeType:     &metadata.MimeType,
			UploadedBy:   req.UserID,
		}

		if err := s.fileRepo.CreateAttachment(ctx, attachment); err != nil {
			logger.Error("Error al crear registro de attachment: %v", err)
		}
	}

	// Registrar en auditoría
	s.auditLog(ctx, req.UserID, models.AuditActionCreate, "files", metadata.ID.String(), nil, map[string]interface{}{
		"filename": req.File.Filename,
		"size":     uploadResult.Size,
		"category": req.Category,
	})

	// Generar URLs
	downloadURL := uploadResult.PresignedURL
	if downloadURL == "" {
		downloadURL, _ = s.minioClient.GetPresignedURL(ctx, uploadResult.BucketName, uploadResult.ObjectKey, 24*time.Hour)
	}

	response := &FileResponse{
		ID:           metadata.ID,
		OriginalName: metadata.OriginalName,
		FileName:     metadata.StoredName,
		FileSize:     metadata.FileSize,
		MimeType:     metadata.MimeType,
		Category:     metadata.Category,
		DownloadURL:  downloadURL,
		UploadedBy:   metadata.UploadedBy,
		UploadedAt:   metadata.CreatedAt,
		Metadata:     metadata.Metadata,
	}

	// Generar thumbnail si es imagen
	if s.isImageFile(metadata.MimeType) {
		response.ThumbnailURL = s.generateThumbnailURL(uploadResult.BucketName, uploadResult.ObjectKey)
	}

	logger.Info("✅ Archivo subido exitosamente: %s", metadata.ID)
	return response, nil
}

// UploadMultipleFiles maneja la carga de múltiples archivos
func (s *FileService) UploadMultipleFiles(ctx context.Context, files []*multipart.FileHeader, category config.FileCategory, unitID int, userID uuid.UUID, messageID *int64) ([]*FileResponse, error) {
	var responses []*FileResponse
	var errors []error

	for _, file := range files {
		req := &UploadFileRequest{
			File:      file,
			Category:  category,
			UnitID:    unitID,
			UserID:    userID,
			MessageID: messageID,
		}

		response, err := s.UploadFile(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", file.Filename, err))
			continue
		}
		responses = append(responses, response)
	}

	if len(errors) > 0 {
		return responses, fmt.Errorf("algunos archivos fallaron: %v", errors)
	}

	return responses, nil
}

// GetFile obtiene un archivo por ID
func (s *FileService) GetFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (*FileResponse, error) {
	logger.Debug("📄 Obteniendo archivo: %s", fileID)

	// Obtener metadatos
	metadata, err := s.fileRepo.GetMetadataByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("archivo no encontrado: %w", err)
	}

	// Verificar permisos
	if !s.hasFileAccess(ctx, metadata, userID) {
		return nil, fmt.Errorf("no tiene permisos para acceder a este archivo")
	}

	// Incrementar contador de accesos
	s.fileRepo.UpdateAccessCount(ctx, fileID)

	// Generar URL de descarga
	downloadURL, err := s.minioClient.GetPresignedURL(ctx, metadata.BucketName, metadata.FilePath, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("error al generar URL de descarga: %w", err)
	}

	response := &FileResponse{
		ID:           metadata.ID,
		OriginalName: metadata.OriginalName,
		FileName:     metadata.StoredName,
		FileSize:     metadata.FileSize,
		MimeType:     metadata.MimeType,
		Category:     metadata.Category,
		DownloadURL:  downloadURL,
		UploadedBy:   metadata.UploadedBy,
		UploadedAt:   metadata.CreatedAt,
		Metadata:     metadata.Metadata,
	}

	// Generar thumbnail si es imagen
	if s.isImageFile(metadata.MimeType) {
		response.ThumbnailURL = s.generateThumbnailURL(metadata.BucketName, metadata.FilePath)
	}

	// Registrar acceso en auditoría
	s.auditLog(ctx, userID, models.AuditActionRead, "files", fileID.String(), nil, nil)

	return response, nil
}

// DownloadFile descarga un archivo
func (s *FileService) DownloadFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) (io.ReadCloser, *models.FileMetadata, error) {
	logger.Info("⬇️ Descargando archivo: %s", fileID)

	// Obtener metadatos
	metadata, err := s.fileRepo.GetMetadataByID(ctx, fileID)
	if err != nil {
		return nil, nil, fmt.Errorf("archivo no encontrado: %w", err)
	}

	// Verificar permisos
	if !s.hasFileAccess(ctx, metadata, userID) {
		return nil, nil, fmt.Errorf("no tiene permisos para descargar este archivo")
	}

	// Descargar de MinIO
	reader, _, err := s.minioClient.DownloadToReader(ctx, metadata.BucketName, metadata.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error al descargar archivo: %w", err)
	}

	// Incrementar contador y actualizar último acceso
	s.fileRepo.UpdateAccessCount(ctx, fileID)

	// Registrar descarga en auditoría
	s.auditLog(ctx, userID, models.AuditActionRead, "files", fileID.String(), nil, map[string]interface{}{
		"action":   "download",
		"filename": metadata.OriginalName,
	})

	return reader, metadata, nil
}

// DeleteFile elimina un archivo
func (s *FileService) DeleteFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	logger.Warn("🗑️ Eliminando archivo: %s", fileID)

	// Obtener metadatos
	metadata, err := s.fileRepo.GetMetadataByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("archivo no encontrado: %w", err)
	}

	// Verificar permisos (solo el que subió o admin puede eliminar)
	if !s.canDeleteFile(ctx, metadata, userID) {
		return fmt.Errorf("no tiene permisos para eliminar este archivo")
	}

	// Eliminar de MinIO
	if err := s.minioClient.RemoveObject(ctx, metadata.BucketName, metadata.FilePath); err != nil {
		logger.Error("Error al eliminar archivo de MinIO: %v", err)
	}

	// Eliminar metadatos
	if err := s.fileRepo.DeleteMetadata(ctx, fileID); err != nil {
		return fmt.Errorf("error al eliminar metadatos: %w", err)
	}

	// Eliminar attachments si existen
	if attachments, _ := s.fileRepo.GetAttachmentsByMessageID(ctx, 0); len(attachments) > 0 {
		for _, att := range attachments {
			if att.FilePath == metadata.FilePath {
				s.fileRepo.DeleteAttachment(ctx, att.ID)
			}
		}
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionDelete, "files", fileID.String(), map[string]interface{}{
		"filename": metadata.OriginalName,
		"size":     metadata.FileSize,
		"category": metadata.Category,
	}, nil)

	logger.Info("✅ Archivo eliminado exitosamente")
	return nil
}

// GetFilesByFilter obtiene archivos con filtros
func (s *FileService) GetFilesByFilter(ctx context.Context, filter *repositories.FileFilter) ([]*FileResponse, int64, error) {
	files, total, err := s.fileRepo.GetFilesByFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*FileResponse, len(files))
	for i, file := range files {
		downloadURL, _ := s.minioClient.GetPresignedURL(ctx, file.BucketName, file.FilePath, 24*time.Hour)

		responses[i] = &FileResponse{
			ID:           file.ID,
			OriginalName: file.OriginalName,
			FileName:     file.StoredName,
			FileSize:     file.FileSize,
			MimeType:     file.MimeType,
			Category:     file.Category,
			DownloadURL:  downloadURL,
			UploadedBy:   file.UploadedBy,
			UploadedAt:   file.CreatedAt,
			Metadata:     file.Metadata,
		}

		if s.isImageFile(file.MimeType) {
			responses[i].ThumbnailURL = s.generateThumbnailURL(file.BucketName, file.FilePath)
		}
	}

	return responses, total, nil
}

// GetStorageStats obtiene estadísticas de almacenamiento
func (s *FileService) GetStorageStats(ctx context.Context) (*repositories.StorageStats, error) {
	return s.fileRepo.GetStorageStats(ctx)
}

// CleanupTempFiles limpia archivos temporales expirados
func (s *FileService) CleanupTempFiles(ctx context.Context, expirationHours int) error {
	logger.Info("🧹 Limpiando archivos temporales expirados")

	// Obtener archivos temporales expirados
	expiredFiles, err := s.fileRepo.GetExpiredTempFiles(ctx, expirationHours)
	if err != nil {
		return fmt.Errorf("error al obtener archivos expirados: %w", err)
	}

	deleted := 0
	for _, file := range expiredFiles {
		// Eliminar de MinIO
		if err := s.minioClient.RemoveObject(ctx, file.BucketName, file.FilePath); err != nil {
			logger.Error("Error al eliminar archivo temporal de MinIO: %v", err)
			continue
		}

		// Eliminar metadatos
		if err := s.fileRepo.DeleteMetadata(ctx, file.ID); err != nil {
			logger.Error("Error al eliminar metadatos: %v", err)
			continue
		}

		deleted++
	}

	logger.Info("✅ Archivos temporales eliminados: %d de %d", deleted, len(expiredFiles))
	return nil
}

// Funciones auxiliares

// hasFileAccess verifica si un usuario tiene acceso a un archivo
func (s *FileService) hasFileAccess(ctx context.Context, file *models.FileMetadata, userID uuid.UUID) bool {
	// Si es público, todos tienen acceso
	if file.IsPublic {
		return true
	}

	// Si es el que lo subió, tiene acceso
	if file.UploadedBy == userID {
		return true
	}

	// Verificar si es de la misma unidad
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	// Admin tiene acceso a todo
	if user.Role == "admin" {
		return true
	}

	// Si es de la misma unidad
	if user.OrganizationalUnitID != nil && *user.OrganizationalUnitID == file.UnitID {
		return true
	}

	return false
}

// canDeleteFile verifica si un usuario puede eliminar un archivo
func (s *FileService) canDeleteFile(ctx context.Context, file *models.FileMetadata, userID uuid.UUID) bool {
	// Si es el que lo subió, puede eliminar
	if file.UploadedBy == userID {
		return true
	}

	// Admin puede eliminar cualquier archivo
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	return user.Role == "admin"
}

// isPublicCategory determina si una categoría es pública
func (s *FileService) isPublicCategory(category config.FileCategory) bool {
	publicCategories := []config.FileCategory{
		config.FileCategoryImage,
		config.FileCategoryDocument,
		config.FileCategoryReport,
	}

	for _, pc := range publicCategories {
		if pc == category {
			return true
		}
	}
	return false
}

// isImageFile verifica si es un archivo de imagen
func (s *FileService) isImageFile(mimeType string) bool {
	imageMimeTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/svg+xml", "image/bmp",
	}

	for _, imt := range imageMimeTypes {
		if imt == mimeType {
			return true
		}
	}
	return false
}

// generateThumbnailURL genera URL para thumbnail
func (s *FileService) generateThumbnailURL(bucket, objectKey string) string {
	// Por ahora, retornamos la misma URL
	// En el futuro, podríamos generar thumbnails reales
	url, _ := s.minioClient.GetPresignedURL(context.Background(), bucket, objectKey, 24*time.Hour)
	return url
}

// auditLog registra una acción en el log de auditoría
func (s *FileService) auditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, resource, resourceID string, oldValues, newValues map[string]interface{}) {
	log := &models.AuditLog{
		UserID:     &userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValues:  oldValues,
		NewValues:  newValues,
		Result:     models.AuditResultSuccess,
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Error al registrar en auditoría: %v", err)
	}
}

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
		Category:     s.convertConfigCategoryToModelCategory(req.Category), // CORREGIDO: Conversión apropiada
		UploadedBy:   req.UserID,
		UnitID:       req.UnitID,
		Tags:         req.Tags,
		Checksum:     uploadResult.ETag,
		IsPublic:     s.isPublicCategory(req.Category),
	}

	// CORREGIDO: Usar función auxiliar para convertir metadata
	metadata.SetMetadataFromString(req.Metadata)

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
			MimeType:     metadata.MimeType, // CORREGIDO: Sin & ya que es string, no *string
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
		Category:     string(metadata.Category), // CORREGIDO: Conversión a string para response
		DownloadURL:  downloadURL,
		UploadedBy:   metadata.UploadedBy,
		UploadedAt:   metadata.CreatedAt,
		Metadata:     metadata.ConvertMetadataToString(), // CORREGIDO: Usar función de conversión
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
		Category:     string(metadata.Category), // CORREGIDO: Conversión a string
		DownloadURL:  downloadURL,
		UploadedBy:   metadata.UploadedBy,
		UploadedAt:   metadata.CreatedAt,
		Metadata:     metadata.ConvertMetadataToString(), // CORREGIDO: Conversión de map
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
	s.auditLog(ctx, userID, models.AuditActionRead, "files", fileID.String(), nil, nil)

	return reader, metadata, nil
}

// UpdateFile actualiza metadatos de un archivo
func (s *FileService) UpdateFile(ctx context.Context, fileID uuid.UUID, req *UpdateFileRequest, userID uuid.UUID) (*FileResponse, error) {
	// Obtener archivo actual
	metadata, err := s.fileRepo.GetMetadataByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("archivo no encontrado: %w", err)
	}

	// Verificar permisos
	if !s.hasFileWriteAccess(ctx, metadata, userID) {
		return nil, fmt.Errorf("no tiene permisos para modificar este archivo")
	}

	// Actualizar campos
	oldValues := map[string]interface{}{
		"description": metadata.Description,
		"tags":        metadata.Tags,
		"is_public":   metadata.IsPublic,
	}

	if req.Description != nil {
		metadata.Description = *req.Description
	}
	if req.Tags != nil {
		metadata.Tags = req.Tags
	}
	if req.IsPublic != nil {
		metadata.IsPublic = *req.IsPublic
	}
	if req.ExpiresAt != nil {
		metadata.ExpiresAt = req.ExpiresAt
	}

	// Guardar cambios
	if err := s.fileRepo.UpdateMetadata(ctx, metadata); err != nil {
		return nil, fmt.Errorf("error al actualizar archivo: %w", err)
	}

	// Registrar en auditoría
	newValues := map[string]interface{}{
		"description": metadata.Description,
		"tags":        metadata.Tags,
		"is_public":   metadata.IsPublic,
	}
	s.auditLog(ctx, userID, models.AuditActionUpdate, "files", fileID.String(), oldValues, newValues)

	// Retornar respuesta actualizada
	return s.GetFile(ctx, fileID, userID)
}

// DeleteFile elimina un archivo
func (s *FileService) DeleteFile(ctx context.Context, fileID uuid.UUID, userID uuid.UUID) error {
	logger.Info("🗑️ Eliminando archivo: %s", fileID)

	// Obtener metadatos
	metadata, err := s.fileRepo.GetMetadataByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("archivo no encontrado: %w", err)
	}

	// Verificar permisos
	if !s.hasFileWriteAccess(ctx, metadata, userID) {
		return fmt.Errorf("no tiene permisos para eliminar este archivo")
	}

	// Eliminar de MinIO
	if err := s.minioClient.RemoveObject(ctx, metadata.BucketName, metadata.FilePath); err != nil {
		logger.Error("Error al eliminar de MinIO: %v", err)
		// Continuar con eliminación de BD aunque falle MinIO
	}

	// Eliminar de base de datos
	if err := s.fileRepo.DeleteMetadata(ctx, fileID); err != nil {
		return fmt.Errorf("error al eliminar registro: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionDelete, "files", fileID.String(),
		map[string]interface{}{"filename": metadata.OriginalName}, nil)

	logger.Info("✅ Archivo eliminado: %s", fileID)
	return nil
}

// GetFiles obtiene archivos con filtros
func (s *FileService) GetFiles(ctx context.Context, filter *repositories.FileFilter) ([]*FileResponse, int64, error) {
	files, total, err := s.fileRepo.GetFilesByFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*FileResponse, len(files))
	for i, file := range files {
		responses[i] = &FileResponse{
			ID:           file.ID,
			OriginalName: file.OriginalName,
			FileName:     file.StoredName,
			FileSize:     file.FileSize,
			MimeType:     file.MimeType,
			Category:     string(file.Category), // CORREGIDO: Conversión a string
			UploadedBy:   file.UploadedBy,
			UploadedAt:   file.CreatedAt,
			Metadata:     file.ConvertMetadataToString(), // CORREGIDO: Conversión de map
		}
	}

	return responses, total, nil
}

// Funciones auxiliares

// convertConfigCategoryToModelCategory convierte config.FileCategory a models.FileCategory
func (s *FileService) convertConfigCategoryToModelCategory(configCategory config.FileCategory) models.FileCategory {
	switch configCategory {
	case config.FileCategoryAttachment:
		return models.FileCategoryAttachment
	case config.FileCategoryDocument:
		return models.FileCategoryDocument
	case config.FileCategoryImage:
		return models.FileCategoryImage
	case config.FileCategoryReport:
		return models.FileCategoryReport
	case config.FileCategoryTemp:
		return models.FileCategoryTemp
	case config.FileCategoryBackup:
		return models.FileCategoryBackup
	default:
		return models.FileCategoryOther
	}
}

// hasFileAccess verifica si un usuario tiene acceso de lectura a un archivo
func (s *FileService) hasFileAccess(ctx context.Context, file *models.FileMetadata, userID uuid.UUID) bool {
	// Si es público, todos tienen acceso
	if file.IsPublic {
		return true
	}

	// Si es quien lo subió, tiene acceso
	if file.UploadedBy == userID {
		return true
	}

	// Verificar si es admin o de la misma unidad
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	if user.Role == "admin" {
		return true
	}

	// Si es de la misma unidad
	if user.OrganizationalUnitID != nil && *user.OrganizationalUnitID == file.GetUnitID() {
		return true
	}

	return false
}

// hasFileWriteAccess verifica si un usuario tiene acceso de escritura a un archivo
func (s *FileService) hasFileWriteAccess(ctx context.Context, file *models.FileMetadata, userID uuid.UUID) bool {
	// Si es quien lo subió, tiene acceso
	if file.UploadedBy == userID {
		return true
	}

	// Verificar si es admin
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	return user.Role == "admin"
}

// isPublicCategory verifica si una categoría debe ser pública por defecto
func (s *FileService) isPublicCategory(category config.FileCategory) bool {
	publicCategories := []config.FileCategory{
		config.FileCategoryImage,
	}

	for _, pubCat := range publicCategories {
		if category == pubCat {
			return true
		}
	}
	return false
}

// isImageFile verifica si un archivo es una imagen
func (s *FileService) isImageFile(mimeType string) bool {
	imageTypes := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/svg+xml", "image/bmp",
	}
	for _, imgType := range imageTypes {
		if mimeType == imgType {
			return true
		}
	}
	return false
}

// generateThumbnailURL genera URL de thumbnail para imágenes
func (s *FileService) generateThumbnailURL(bucketName, objectKey string) string {
	// Implementación simplificada
	// En producción, esto generaría thumbnails reales
	return fmt.Sprintf("/api/v1/files/thumbnail/%s/%s", bucketName, objectKey)
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

// UpdateFileRequest estructura para actualizar archivo
type UpdateFileRequest struct {
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	IsPublic    *bool      `json:"isPublic,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

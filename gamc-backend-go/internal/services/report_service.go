// internal/services/report_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/integrations/minio"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportService maneja la lógica de negocio para reportes
type ReportService struct {
	reportRepo  *repositories.ReportRepository
	fileRepo    *repositories.FileRepository
	auditRepo   *repositories.AuditRepository
	minioClient *minio.Client
	config      *config.Config
	db          *gorm.DB
}

// NewReportService crea una nueva instancia del servicio
func NewReportService(db *gorm.DB, cfg *config.Config) (*ReportService, error) {
	// Inicializar cliente MinIO
	minioClient, err := minio.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error al inicializar cliente MinIO: %w", err)
	}

	return &ReportService{
		reportRepo:  repositories.NewReportRepository(db),
		fileRepo:    repositories.NewFileRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
		minioClient: minioClient,
		config:      cfg,
		db:          db,
	}, nil
}

// GenerateReportRequest solicitud para generar reporte
type GenerateReportRequest struct {
	TemplateID   string                 `json:"templateId" validate:"required"`
	Name         string                 `json:"name" validate:"required,min=3,max=255"`
	Description  string                 `json:"description,omitempty"`
	Format       models.ReportFormat    `json:"format" validate:"required"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	ScheduledFor *time.Time             `json:"scheduledFor,omitempty"`
	UserID       uuid.UUID              `json:"-"`
	UnitID       int                    `json:"-"`
}

// ReportResponse respuesta de reporte
type ReportResponse struct {
	ID                  uuid.UUID         `json:"id"`
	Name                string            `json:"name"`
	Description         string            `json:"description,omitempty"`
	Type                string            `json:"type"`
	Status              string            `json:"status"`
	Format              string            `json:"format"`
	FileID              *uuid.UUID        `json:"fileId,omitempty"`
	FileSize            int64             `json:"fileSize,omitempty"`
	RecordCount         int64             `json:"recordCount,omitempty"`
	ProcessingTime      int               `json:"processingTime,omitempty"`
	GeneratedAt         time.Time         `json:"generatedAt"`
	GenerationStarted   *time.Time        `json:"generationStarted,omitempty"`
	GenerationCompleted *time.Time        `json:"generationCompleted,omitempty"`
	ScheduledFor        *time.Time        `json:"scheduledFor,omitempty"`
	ExpiresAt           *time.Time        `json:"expiresAt,omitempty"`
	ErrorMessage        string            `json:"errorMessage,omitempty"`
	DownloadURL         string            `json:"downloadUrl,omitempty"`
	RequestedBy         uuid.UUID         `json:"requestedBy"`
	Template            *TemplateResponse `json:"template,omitempty"`
}

// TemplateResponse respuesta de plantilla
type TemplateResponse struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description,omitempty"`
	Category         string   `json:"category,omitempty"`
	Type             string   `json:"type"`
	DefaultFormat    string   `json:"defaultFormat"`
	AvailableFormats []string `json:"availableFormats"`
	RequiredRole     string   `json:"requiredRole,omitempty"`
}

// GenerateReport genera un nuevo reporte
func (s *ReportService) GenerateReport(ctx context.Context, req *GenerateReportRequest) (*ReportResponse, error) {
	logger.Info("📊 Generando reporte: %s", req.Name)

	// Obtener plantilla
	template, err := s.reportRepo.GetTemplateByID(ctx, uuid.MustParse(req.TemplateID))
	if err != nil {
		return nil, fmt.Errorf("plantilla no encontrada: %w", err)
	}

	if !template.IsActive {
		return nil, fmt.Errorf("plantilla inactiva")
	}

	// Verificar formato disponible
	formatAvailable := false
	for _, f := range template.AvailableFormats {
		if f == req.Format {
			formatAvailable = true
			break
		}
	}
	if !formatAvailable {
		return nil, fmt.Errorf("formato no disponible para esta plantilla")
	}

	// Crear registro de reporte
	report := &models.Report{
		Name:                req.Name,
		Description:         req.Description,
		Type:                template.Type,
		Status:              models.ReportStatusPending,
		Format:              req.Format,
		RequestedBy:         req.UserID,
		OrganizationID:      req.UnitID,
		TemplateID:          &template.ID,
		Parameters:          req.Parameters,
		ScheduledFor:        req.ScheduledFor,
		GenerationTimeoutMs: 300000, // 5 minutos por defecto
	}

	// Si está programado, marcarlo como tal
	if req.ScheduledFor != nil && req.ScheduledFor.After(time.Now()) {
		report.Status = models.ReportStatusScheduled
	}

	// Establecer expiración (30 días por defecto)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	report.ExpiresAt = &expiresAt

	if err := s.reportRepo.CreateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("error al crear reporte: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, req.UserID, models.AuditActionCreate, "reports", report.ID.String(), nil, map[string]interface{}{
		"name":     req.Name,
		"template": req.TemplateID,
		"format":   req.Format,
	})

	// Si no está programado, iniciar generación inmediata
	if report.Status == models.ReportStatusPending {
		go s.processReport(context.Background(), report.ID)
	}

	logger.Info("✅ Reporte creado: %s", report.ID)
	return s.convertReportToResponse(report), nil
}

// processReport procesa la generación de un reporte
func (s *ReportService) processReport(ctx context.Context, reportID uuid.UUID) {
	logger.Info("⚙️ Procesando reporte: %s", reportID)

	// Obtener reporte
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		logger.Error("Error al obtener reporte: %v", err)
		return
	}

	// Marcar inicio de generación
	if err := report.StartGeneration(s.db); err != nil {
		logger.Error("Error al marcar inicio: %v", err)
		return
	}

	// Generar contenido del reporte
	content, recordCount, err := s.generateReportContent(ctx, report)
	if err != nil {
		logger.Error("Error al generar contenido: %v", err)
		report.FailGeneration(s.db, err.Error())
		return
	}

	// Guardar archivo en MinIO
	fileMetadata, err := s.saveReportFile(ctx, report, content)
	if err != nil {
		logger.Error("Error al guardar archivo: %v", err)
		report.FailGeneration(s.db, fmt.Sprintf("Error al guardar archivo: %v", err))
		return
	}

	// Marcar como completado
	if err := report.CompleteGeneration(s.db, fileMetadata.ID, fileMetadata.FileSize, recordCount); err != nil {
		logger.Error("Error al completar generación: %v", err)
		return
	}

	logger.Info("✅ Reporte generado exitosamente: %s", reportID)
}

// generateReportContent genera el contenido del reporte
func (s *ReportService) generateReportContent(ctx context.Context, report *models.Report) ([]byte, int64, error) {
	// Esta es una implementación simplificada
	// En producción, aquí se ejecutarían las consultas y se generaría el contenido real

	switch report.Type {
	case models.ReportTypeDaily:
		return s.generateDailyReport(ctx, report)
	case models.ReportTypeWeekly:
		return s.generateWeeklyReport(ctx, report)
	case models.ReportTypeMonthly:
		return s.generateMonthlyReport(ctx, report)
	case models.ReportTypeCustom:
		return s.generateCustomReport(ctx, report)
	default:
		return nil, 0, fmt.Errorf("tipo de reporte no soportado: %s", report.Type)
	}
}

// saveReportFile guarda el archivo del reporte en MinIO
func (s *ReportService) saveReportFile(ctx context.Context, report *models.Report, content []byte) (*models.FileMetadata, error) {
	// Generar nombre de archivo
	timestamp := time.Now().Format("20060102_150405")
	extension := s.getFileExtension(report.Format)
	filename := fmt.Sprintf("%s_%s%s", report.Name, timestamp, extension)

	// Crear metadatos del archivo
	metadata := &models.FileMetadata{
		OriginalName: filename,
		StoredName:   filename,
		FilePath:     fmt.Sprintf("reports/%d/%s/%s", report.OrganizationID, time.Now().Format("2006/01"), filename),
		BucketName:   config.GetMinIOBuckets().Reports,
		FileSize:     int64(len(content)),
		MimeType:     s.getMimeType(report.Format),
		Category:     "report",
		UploadedBy:   report.RequestedBy,
		UnitID:       report.OrganizationID,
		Metadata: map[string]string{
			"report_id":   report.ID.String(),
			"report_type": string(report.Type),
			"format":      string(report.Format),
		},
	}

	// TODO: Implementar la subida real a MinIO
	// Por ahora, solo creamos el registro
	if err := s.fileRepo.CreateMetadata(ctx, metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// GetReport obtiene un reporte por ID
func (s *ReportService) GetReport(ctx context.Context, reportID, userID uuid.UUID) (*ReportResponse, error) {
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("reporte no encontrado: %w", err)
	}

	// Verificar permisos
	if !s.hasReportAccess(ctx, report, userID) {
		return nil, fmt.Errorf("no tiene permisos para acceder a este reporte")
	}

	response := s.convertReportToResponse(report)

	// Si está completado, generar URL de descarga
	if report.Status == models.ReportStatusCompleted && report.FileID != nil {
		if file, err := s.fileRepo.GetMetadataByID(ctx, *report.FileID); err == nil {
			url, _ := s.minioClient.GetPresignedURL(ctx, file.BucketName, file.FilePath, 24*time.Hour)
			response.DownloadURL = url
		}
	}

	return response, nil
}

// GetReports obtiene reportes con filtros
func (s *ReportService) GetReports(ctx context.Context, filter *repositories.ReportFilter) ([]*ReportResponse, int64, error) {
	reports, total, err := s.reportRepo.GetReportsByFilter(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = s.convertReportToResponse(report)
	}

	return responses, total, nil
}

// GetTemplates obtiene las plantillas disponibles
func (s *ReportService) GetTemplates(ctx context.Context, userRole string) ([]*TemplateResponse, error) {
	templates, err := s.reportRepo.GetActiveTemplates(ctx)
	if err != nil {
		return nil, err
	}

	// Filtrar por rol si es necesario
	var responses []*TemplateResponse
	for _, template := range templates {
		// Si la plantilla requiere un rol específico, verificar
		if template.RequiredRole != "" && template.RequiredRole != userRole && userRole != "admin" {
			continue
		}

		responses = append(responses, s.convertTemplateToResponse(template))
	}

	return responses, nil
}

// CancelReport cancela un reporte programado o en proceso
func (s *ReportService) CancelReport(ctx context.Context, reportID, userID uuid.UUID) error {
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("reporte no encontrado: %w", err)
	}

	// Verificar permisos
	if report.RequestedBy != userID {
		// Verificar si es admin
		var user models.User
		if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil || user.Role != "admin" {
			return fmt.Errorf("no tiene permisos para cancelar este reporte")
		}
	}

	// Solo se pueden cancelar reportes pendientes o programados
	if report.Status != models.ReportStatusPending && report.Status != models.ReportStatusScheduled {
		return fmt.Errorf("el reporte no se puede cancelar en su estado actual")
	}

	// Actualizar estado
	if err := s.reportRepo.UpdateReportStatus(ctx, reportID, models.ReportStatusCancelled, "Cancelado por usuario"); err != nil {
		return fmt.Errorf("error al cancelar reporte: %w", err)
	}

	// Registrar en auditoría
	s.auditLog(ctx, userID, models.AuditActionUpdate, "reports", reportID.String(),
		map[string]interface{}{"status": report.Status},
		map[string]interface{}{"status": models.ReportStatusCancelled})

	logger.Info("✅ Reporte cancelado: %s", reportID)
	return nil
}

// GetReportStats obtiene estadísticas de reportes
func (s *ReportService) GetReportStats(ctx context.Context, startDate, endDate time.Time) (*repositories.ReportStats, error) {
	return s.reportRepo.GetReportStats(ctx, startDate, endDate)
}

// GetDashboardMetrics obtiene métricas del dashboard
func (s *ReportService) GetDashboardMetrics(ctx context.Context, metricType string) ([]*models.DashboardMetric, error) {
	return s.reportRepo.GetDashboardMetrics(ctx, metricType)
}

// UpdateDashboardMetric actualiza una métrica del dashboard
func (s *ReportService) UpdateDashboardMetric(ctx context.Context, metric *models.DashboardMetric) error {
	return s.reportRepo.UpdateDashboardMetric(ctx, metric)
}

// CleanupOldReports limpia reportes antiguos
func (s *ReportService) CleanupOldReports(ctx context.Context, days int) error {
	logger.Info("🧹 Limpiando reportes antiguos (> %d días)", days)

	// Obtener reportes a eliminar
	filter := &repositories.ReportFilter{
		GeneratedTo: &[]time.Time{time.Now().AddDate(0, 0, -days)}[0],
		Status:      &models.ReportStatusCompleted,
	}

	reports, _, err := s.reportRepo.GetReportsByFilter(ctx, filter)
	if err != nil {
		return err
	}

	deleted := 0
	for _, report := range reports {
		// Eliminar archivo si existe
		if report.FileMetadata != nil {
			if err := s.minioClient.RemoveObject(ctx, report.FileMetadata.BucketName, report.FileMetadata.FilePath); err != nil {
				logger.Error("Error al eliminar archivo de reporte: %v", err)
			}
			s.fileRepo.DeleteMetadata(ctx, report.FileMetadata.ID)
		}

		deleted++
	}

	// Eliminar registros de la base de datos
	deletedCount, err := s.reportRepo.DeleteOldReports(ctx, days)
	if err != nil {
		return fmt.Errorf("error al eliminar reportes: %w", err)
	}

	logger.Info("✅ Reportes eliminados: %d", deletedCount)
	return nil
}

// Funciones auxiliares

// hasReportAccess verifica si un usuario tiene acceso a un reporte
func (s *ReportService) hasReportAccess(ctx context.Context, report *models.Report, userID uuid.UUID) bool {
	// Si es quien lo solicitó, tiene acceso
	if report.RequestedBy == userID {
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
	if user.OrganizationalUnitID != nil && *user.OrganizationalUnitID == report.OrganizationID {
		return true
	}

	return false
}

// convertReportToResponse convierte un modelo a response
func (s *ReportService) convertReportToResponse(report *models.Report) *ReportResponse {
	response := &ReportResponse{
		ID:                  report.ID,
		Name:                report.Name,
		Description:         report.Description,
		Type:                string(report.Type),
		Status:              string(report.Status),
		Format:              string(report.Format),
		FileID:              report.FileID,
		FileSize:            report.FileSize,
		RecordCount:         report.RecordCount,
		ProcessingTime:      report.ProcessingTime,
		GeneratedAt:         report.GeneratedAt,
		GenerationStarted:   report.GenerationStarted,
		GenerationCompleted: report.GenerationCompleted,
		ScheduledFor:        report.ScheduledFor,
		ExpiresAt:           report.ExpiresAt,
		ErrorMessage:        report.ErrorMessage,
		RequestedBy:         report.RequestedBy,
	}

	if report.Template != nil {
		response.Template = s.convertTemplateToResponse(report.Template)
	}

	return response
}

// convertTemplateToResponse convierte una plantilla a response
func (s *ReportService) convertTemplateToResponse(template *models.ReportTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:               template.ID,
		Name:             template.Name,
		Description:      template.Description,
		Category:         template.Category,
		Type:             string(template.Type),
		DefaultFormat:    string(template.DefaultFormat),
		AvailableFormats: s.convertFormats(template.AvailableFormats),
		RequiredRole:     template.RequiredRole,
	}
}

// convertFormats convierte formatos
func (s *ReportService) convertFormats(formats []models.ReportFormat) []string {
	result := make([]string, len(formats))
	for i, f := range formats {
		result[i] = string(f)
	}
	return result
}

// getFileExtension obtiene la extensión según el formato
func (s *ReportService) getFileExtension(format models.ReportFormat) string {
	extensions := map[models.ReportFormat]string{
		models.ReportFormatPDF:   ".pdf",
		models.ReportFormatExcel: ".xlsx",
		models.ReportFormatCSV:   ".csv",
		models.ReportFormatHTML:  ".html",
		models.ReportFormatJSON:  ".json",
		models.ReportFormatXML:   ".xml",
	}
	return extensions[format]
}

// getMimeType obtiene el tipo MIME según el formato
func (s *ReportService) getMimeType(format models.ReportFormat) string {
	mimeTypes := map[models.ReportFormat]string{
		models.ReportFormatPDF:   "application/pdf",
		models.ReportFormatExcel: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		models.ReportFormatCSV:   "text/csv",
		models.ReportFormatHTML:  "text/html",
		models.ReportFormatJSON:  "application/json",
		models.ReportFormatXML:   "application/xml",
	}
	return mimeTypes[format]
}

// auditLog registra una acción en el log de auditoría
func (s *ReportService) auditLog(ctx context.Context, userID uuid.UUID, action models.AuditAction, resource, resourceID string, oldValues, newValues map[string]interface{}) {
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

// Métodos de generación de reportes específicos (simplificados)

func (s *ReportService) generateDailyReport(ctx context.Context, report *models.Report) ([]byte, int64, error) {
	// Implementación simplificada
	content := fmt.Sprintf("Reporte Diario: %s\nGenerado: %s\n", report.Name, time.Now().Format("2006-01-02 15:04:05"))
	return []byte(content), 1, nil
}

func (s *ReportService) generateWeeklyReport(ctx context.Context, report *models.Report) ([]byte, int64, error) {
	// Implementación simplificada
	content := fmt.Sprintf("Reporte Semanal: %s\nGenerado: %s\n", report.Name, time.Now().Format("2006-01-02 15:04:05"))
	return []byte(content), 1, nil
}

func (s *ReportService) generateMonthlyReport(ctx context.Context, report *models.Report) ([]byte, int64, error) {
	// Implementación simplificada
	content := fmt.Sprintf("Reporte Mensual: %s\nGenerado: %s\n", report.Name, time.Now().Format("2006-01-02 15:04:05"))
	return []byte(content), 1, nil
}

func (s *ReportService) generateCustomReport(ctx context.Context, report *models.Report) ([]byte, int64, error) {
	// Implementación simplificada
	content := fmt.Sprintf("Reporte Personalizado: %s\nGenerado: %s\n", report.Name, time.Now().Format("2006-01-02 15:04:05"))
	return []byte(content), 1, nil
}

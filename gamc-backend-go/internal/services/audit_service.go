// internal/services/audit_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService maneja la lógica de negocio para auditoría
type AuditService struct {
	auditRepo *repositories.AuditRepository
	userRepo  *repositories.UserRepository
	db        *gorm.DB
}

// NewAuditService crea una nueva instancia del servicio
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		auditRepo: repositories.NewAuditRepository(db),
		userRepo:  repositories.NewUserRepository(db),
		db:        db,
	}
}

// LogRequest información de la petición HTTP
type LogRequest struct {
	UserID     *uuid.UUID
	Action     models.AuditAction
	Resource   string
	ResourceID string
	OldValues  map[string]interface{}
	NewValues  map[string]interface{}
	IPAddress  string
	UserAgent  string
	SessionID  string
	Duration   int // milliseconds
	Result     models.AuditResult
	ErrorMsg   string
}

// AuditLogFilter filtros para consultar logs
type AuditLogFilter struct {
	UserID    *uuid.UUID
	Action    string
	Resource  string
	DateFrom  *time.Time
	DateTo    *time.Time
	Result    string
	IPAddress string
	Page      int
	Limit     int
}

// AuditLogResponse respuesta de log de auditoría
type AuditLogResponse struct {
	ID         int64                  `json:"id"`
	User       *UserSummary           `json:"user,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resourceId,omitempty"`
	OldValues  map[string]interface{} `json:"oldValues,omitempty"`
	NewValues  map[string]interface{} `json:"newValues,omitempty"`
	IPAddress  string                 `json:"ipAddress,omitempty"`
	UserAgent  string                 `json:"userAgent,omitempty"`
	Result     string                 `json:"result"`
	ErrorMsg   string                 `json:"errorMsg,omitempty"`
	Duration   int                    `json:"duration,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// UserSummary resumen de usuario para logs
type UserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	FullName string    `json:"fullName"`
	Email    string    `json:"email"`
}

// Log registra una acción en el log de auditoría
func (s *AuditService) Log(ctx context.Context, req *LogRequest) error {
	log := &models.AuditLog{
		UserID:     req.UserID,
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		OldValues:  req.OldValues,
		NewValues:  req.NewValues,
		IPAddress:  req.IPAddress,
		UserAgent:  req.UserAgent,
		SessionID:  req.SessionID,
		Duration:   req.Duration,
		Result:     req.Result,
		ErrorMsg:   req.ErrorMsg,
	}

	if err := s.auditRepo.Create(ctx, log); err != nil {
		logger.Error("Error al crear log de auditoría: %v", err)
		return err
	}

	// Si es una acción crítica, podríamos enviar notificaciones
	if s.isCriticalAction(req.Action, req.Resource) && req.Result == models.AuditResultFailure {
		go s.notifyCriticalEvent(ctx, log)
	}

	return nil
}

// LogFromGinContext registra una acción desde el contexto de Gin
func (s *AuditService) LogFromGinContext(c *gin.Context, action models.AuditAction, resource, resourceID string, oldValues, newValues map[string]interface{}) {
	// Obtener usuario del contexto
	var userID *uuid.UUID
	if user, exists := c.Get("user"); exists {
		if userProfile, ok := user.(*models.UserProfile); ok {
			userID = &userProfile.ID
		}
	}

	// Obtener información de la petición
	startTime, _ := c.Get("startTime")
	duration := 0
	if st, ok := startTime.(time.Time); ok {
		duration = int(time.Since(st).Milliseconds())
	}

	sessionID := ""
	if sid, exists := c.Get("sessionID"); exists {
		sessionID, _ = sid.(string)
	}

	req := &LogRequest{
		UserID:     userID,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		OldValues:  oldValues,
		NewValues:  newValues,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		SessionID:  sessionID,
		Duration:   duration,
		Result:     models.AuditResultSuccess,
	}

	s.Log(context.Background(), req)
}

// GetLogs obtiene logs de auditoría con filtros
func (s *AuditService) GetLogs(ctx context.Context, filter *AuditLogFilter) ([]*AuditLogResponse, int64, error) {
	// Convertir a filtro del repositorio
	repoFilter := &repositories.AuditFilter{
		UserID:    filter.UserID,
		Action:    filter.Action,
		Resource:  filter.Resource,
		DateFrom:  filter.DateFrom,
		DateTo:    filter.DateTo,
		Result:    filter.Result,
		IPAddress: filter.IPAddress,
		Limit:     filter.Limit,
		Offset:    (filter.Page - 1) * filter.Limit,
		SortDesc:  true,
	}

	logs, total, err := s.auditRepo.GetByFilter(ctx, repoFilter)
	if err != nil {
		return nil, 0, err
	}

	// Convertir a respuestas
	responses := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.convertToResponse(log)
	}

	return responses, total, nil
}

// GetUserActivity obtiene la actividad de un usuario
func (s *AuditService) GetUserActivity(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) (*repositories.UserActivityReport, error) {
	return s.auditRepo.GetUserActivity(ctx, userID, startDate, endDate)
}

// GetSecurityEvents obtiene eventos de seguridad recientes
func (s *AuditService) GetSecurityEvents(ctx context.Context, hours int) ([]*AuditLogResponse, error) {
	logs, err := s.auditRepo.GetSecurityEvents(ctx, hours)
	if err != nil {
		return nil, err
	}

	responses := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.convertToResponse(log)
	}

	return responses, nil
}

// GetFailedLoginAttempts obtiene intentos fallidos de login
func (s *AuditService) GetFailedLoginAttempts(ctx context.Context, minutes int) ([]*AuditLogResponse, error) {
	logs, err := s.auditRepo.GetFailedAttempts(ctx, models.AuditActionLogin, minutes)
	if err != nil {
		return nil, err
	}

	responses := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.convertToResponse(log)
	}

	return responses, nil
}

// GetSummary obtiene un resumen de auditoría
func (s *AuditService) GetSummary(ctx context.Context, startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	return s.auditRepo.GetSummary(ctx, startDate, endDate)
}

// GetResourceHistory obtiene el historial de cambios de un recurso
func (s *AuditService) GetResourceHistory(ctx context.Context, resource, resourceID string) ([]*AuditLogResponse, error) {
	logs, err := s.auditRepo.GetByResource(ctx, resource, resourceID)
	if err != nil {
		return nil, err
	}

	responses := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.convertToResponse(log)
	}

	return responses, nil
}

// ExportAuditLogs exporta logs de auditoría
func (s *AuditService) ExportAuditLogs(ctx context.Context, filter *AuditLogFilter, format string) ([]byte, error) {
	// Obtener todos los logs sin paginación
	repoFilter := &repositories.AuditFilter{
		UserID:    filter.UserID,
		Action:    filter.Action,
		Resource:  filter.Resource,
		DateFrom:  filter.DateFrom,
		DateTo:    filter.DateTo,
		Result:    filter.Result,
		IPAddress: filter.IPAddress,
		Limit:     10000, // Límite máximo para exportación
		SortDesc:  true,
	}

	logs, _, err := s.auditRepo.GetByFilter(ctx, repoFilter)
	if err != nil {
		return nil, err
	}

	switch format {
	case "csv":
		return s.exportToCSV(logs)
	case "json":
		return s.exportToJSON(logs)
	default:
		return nil, fmt.Errorf("formato de exportación no soportado: %s", format)
	}
}

// CleanupOldLogs limpia logs antiguos
func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) error {
	logger.Info("🧹 Limpiando logs de auditoría antiguos (> %d días)", retentionDays)

	deleted, err := s.auditRepo.DeleteOldLogs(ctx, retentionDays)
	if err != nil {
		return fmt.Errorf("error al limpiar logs: %w", err)
	}

	logger.Info("✅ Logs eliminados: %d", deleted)

	// Registrar la limpieza en sí misma
	s.Log(ctx, &LogRequest{
		Action:     models.AuditActionDelete,
		Resource:   "audit_logs",
		ResourceID: "bulk",
		NewValues: map[string]interface{}{
			"deleted_count":  deleted,
			"retention_days": retentionDays,
		},
		Result: models.AuditResultSuccess,
	})

	return nil
}

// CheckSuspiciousActivity verifica actividad sospechosa
func (s *AuditService) CheckSuspiciousActivity(ctx context.Context, userID uuid.UUID) (bool, string, error) {
	// Verificar múltiples intentos fallidos de login
	failedLogins, err := s.auditRepo.GetFailedAttempts(ctx, models.AuditActionLogin, 30)
	if err != nil {
		return false, "", err
	}

	userFailedCount := 0
	for _, log := range failedLogins {
		if log.UserID != nil && *log.UserID == userID {
			userFailedCount++
		}
	}

	if userFailedCount >= 5 {
		return true, "múltiples intentos fallidos de login", nil
	}

	// Verificar cambios de IP frecuentes
	last24h := time.Now().Add(-24 * time.Hour)
	logs, err := s.auditRepo.GetByUserID(ctx, userID, 100)
	if err != nil {
		return false, "", err
	}

	ipAddresses := make(map[string]bool)
	for _, log := range logs {
		if log.CreatedAt.After(last24h) && log.IPAddress != "" {
			ipAddresses[log.IPAddress] = true
		}
	}

	if len(ipAddresses) > 5 {
		return true, "múltiples direcciones IP en 24 horas", nil
	}

	// Verificar acceso a recursos sensibles
	sensitiveCount := 0
	for _, log := range logs {
		if log.CreatedAt.After(last24h) && s.isSensitiveResource(log.Resource) {
			sensitiveCount++
		}
	}

	if sensitiveCount > 20 {
		return true, "acceso excesivo a recursos sensibles", nil
	}

	return false, "", nil
}

// Funciones auxiliares

// convertToResponse convierte un modelo a response
func (s *AuditService) convertToResponse(log *models.AuditLog) *AuditLogResponse {
	response := &AuditLogResponse{
		ID:         log.ID,
		Action:     string(log.Action),
		Resource:   log.Resource,
		ResourceID: log.ResourceID,
		OldValues:  log.OldValues,
		NewValues:  log.NewValues,
		IPAddress:  log.IPAddress,
		UserAgent:  log.UserAgent,
		Result:     string(log.Result),
		ErrorMsg:   log.ErrorMsg,
		Duration:   log.Duration,
		CreatedAt:  log.CreatedAt,
	}

	// Agregar información del usuario si existe
	if log.User != nil {
		response.User = &UserSummary{
			ID:       log.User.ID,
			Username: log.User.Username,
			FullName: fmt.Sprintf("%s %s", log.User.FirstName, log.User.LastName),
			Email:    log.User.Email,
		}
	}

	return response
}

// isCriticalAction determina si una acción es crítica
func (s *AuditService) isCriticalAction(action models.AuditAction, resource string) bool {
	criticalActions := []models.AuditAction{
		models.AuditActionDelete,
		models.AuditActionLogin,
		models.AuditActionLogout,
	}

	for _, ca := range criticalActions {
		if ca == action {
			return true
		}
	}

	// Recursos críticos
	criticalResources := []string{"users", "passwords", "security_questions", "audit_logs"}
	for _, cr := range criticalResources {
		if cr == resource {
			return true
		}
	}

	return false
}

// isSensitiveResource determina si un recurso es sensible
func (s *AuditService) isSensitiveResource(resource string) bool {
	sensitiveResources := []string{
		"users", "passwords", "security_questions",
		"audit_logs", "reports", "financial_data",
	}

	for _, sr := range sensitiveResources {
		if sr == resource {
			return true
		}
	}

	return false
}

// notifyCriticalEvent notifica eventos críticos
func (s *AuditService) notifyCriticalEvent(ctx context.Context, log *models.AuditLog) {
	// Por ahora solo registramos
	logger.Warn("⚠️ Evento crítico detectado: %s en %s", log.Action, log.Resource)

	// En el futuro, aquí se podrían enviar:
	// - Emails a administradores
	// - Notificaciones push
	// - Alertas a sistemas de monitoreo
}

// exportToCSV exporta logs a formato CSV
func (s *AuditService) exportToCSV(logs []*models.AuditLog) ([]byte, error) {
	// Implementación simplificada
	// En producción, usar una librería CSV apropiada
	csv := "ID,Usuario,Acción,Recurso,ID Recurso,IP,Resultado,Fecha\n"

	for _, log := range logs {
		userName := ""
		if log.User != nil {
			userName = log.User.Email
		}

		csv += fmt.Sprintf("%d,%s,%s,%s,%s,%s,%s,%s\n",
			log.ID,
			userName,
			log.Action,
			log.Resource,
			log.ResourceID,
			log.IPAddress,
			log.Result,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	return []byte(csv), nil
}

// exportToJSON exporta logs a formato JSON
func (s *AuditService) exportToJSON(logs []*models.AuditLog) ([]byte, error) {
	responses := make([]*AuditLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = s.convertToResponse(log)
	}

	// Usar json.Marshal en producción con manejo de errores apropiado
	return []byte("[]"), nil // Placeholder
}

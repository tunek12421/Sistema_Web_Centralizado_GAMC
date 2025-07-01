// internal/services/dashboard_service.go
package services

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/internal/database/models"
	"gamc-backend-go/internal/repositories"
	"gamc-backend-go/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DashboardService maneja la lógica de negocio para el dashboard
type DashboardService struct {
	messageRepo *repositories.MessageRepository
	userRepo    *repositories.UserRepository
	fileRepo    *repositories.FileRepository
	reportRepo  *repositories.ReportRepository
	auditRepo   *repositories.AuditRepository
	db          *gorm.DB
}

// NewDashboardService crea una nueva instancia del servicio
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{
		messageRepo: repositories.NewMessageRepository(db),
		userRepo:    repositories.NewUserRepository(db),
		fileRepo:    repositories.NewFileRepository(db),
		reportRepo:  repositories.NewReportRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
		db:          db,
	}
}

// DashboardSummary resumen general del dashboard
type DashboardSummary struct {
	Messages       MessageSummary `json:"messages"`
	Users          UserSummary    `json:"users"`
	Storage        StorageSummary `json:"storage"`
	Reports        ReportSummary  `json:"reports"`
	RecentActivity []ActivityItem `json:"recentActivity"`
	Alerts         []AlertItem    `json:"alerts"`
	Metrics        []MetricItem   `json:"metrics"`
}

// MessageSummary resumen de mensajes
type MessageSummary struct {
	Total        int64            `json:"total"`
	Unread       int64            `json:"unread"`
	Urgent       int64            `json:"urgent"`
	Today        int64            `json:"today"`
	ThisWeek     int64            `json:"thisWeek"`
	ThisMonth    int64            `json:"thisMonth"`
	ByType       map[string]int64 `json:"byType"`
	ByStatus     map[string]int64 `json:"byStatus"`
	ResponseTime float64          `json:"averageResponseTime"` // horas
	TrendData    []TrendDataPoint `json:"trendData"`
}

// UserSummary resumen de usuarios
type UserSummary struct {
	Total              int64            `json:"total"`
	Active             int64            `json:"active"`
	Online             int64            `json:"online"`
	NewThisMonth       int64            `json:"newThisMonth"`
	ByRole             map[string]int64 `json:"byRole"`
	ByUnit             map[string]int64 `json:"byUnit"`
	ActiveSessions     int64            `json:"activeSessions"`
	AverageSessionTime float64          `json:"averageSessionTime"` // minutos
}

// StorageSummary resumen de almacenamiento
type StorageSummary struct {
	TotalSize       int64            `json:"totalSize"`
	UsedSize        int64            `json:"usedSize"`
	AvailableSize   int64            `json:"availableSize"`
	UsagePercentage float64          `json:"usagePercentage"`
	FileCount       int64            `json:"fileCount"`
	ByCategory      map[string]int64 `json:"byCategory"`
	LargestFiles    []FileInfo       `json:"largestFiles"`
}

// ReportSummary resumen de reportes
type ReportSummary struct {
	TotalGenerated    int64            `json:"totalGenerated"`
	InProgress        int64            `json:"inProgress"`
	Scheduled         int64            `json:"scheduled"`
	Failed            int64            `json:"failed"`
	AverageGenTime    float64          `json:"averageGenerationTime"` // segundos
	MostUsedTemplates []TemplateUsage  `json:"mostUsedTemplates"`
	GenerationTrend   []TrendDataPoint `json:"generationTrend"`
}

// ActivityItem elemento de actividad reciente
type ActivityItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Action      string    `json:"action"`
	User        string    `json:"user"`
	UserID      uuid.UUID `json:"userId"`
	Resource    string    `json:"resource"`
	ResourceID  string    `json:"resourceId"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	IPAddress   string    `json:"ipAddress,omitempty"`
}

// AlertItem elemento de alerta
type AlertItem struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Severity     string    `json:"severity"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
}

// MetricItem elemento de métrica
type MetricItem struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	DisplayName   string                 `json:"displayName"`
	Value         float64                `json:"value"`
	PreviousValue float64                `json:"previousValue"`
	Change        float64                `json:"change"`
	ChangePercent float64                `json:"changePercent"`
	Unit          string                 `json:"unit"`
	Status        string                 `json:"status"`
	Icon          string                 `json:"icon"`
	Color         string                 `json:"color"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TrendDataPoint punto de datos de tendencia
type TrendDataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Label string    `json:"label"`
}

// FileInfo información de archivo
type FileInfo struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Type     string    `json:"type"`
	Uploaded time.Time `json:"uploaded"`
}

// TemplateUsage uso de plantilla
type TemplateUsage struct {
	TemplateID   string  `json:"templateId"`
	TemplateName string  `json:"templateName"`
	UsageCount   int64   `json:"usageCount"`
	Percentage   float64 `json:"percentage"`
}

// GetDashboardSummary obtiene el resumen del dashboard
func (s *DashboardService) GetDashboardSummary(ctx context.Context, userID uuid.UUID, unitID int, role string) (*DashboardSummary, error) {
	logger.Info("📊 Obteniendo resumen del dashboard para usuario: %s", userID)

	summary := &DashboardSummary{}

	// Obtener resumen de mensajes
	messageSummary, err := s.getMessageSummary(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener resumen de mensajes: %v", err)
	} else {
		summary.Messages = *messageSummary
	}

	// Obtener resumen de usuarios
	userSummary, err := s.getUserSummary(ctx, role)
	if err != nil {
		logger.Error("Error al obtener resumen de usuarios: %v", err)
	} else {
		summary.Users = *userSummary
	}

	// Obtener resumen de almacenamiento
	storageSummary, err := s.getStorageSummary(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener resumen de almacenamiento: %v", err)
	} else {
		summary.Storage = *storageSummary
	}

	// Obtener resumen de reportes
	reportSummary, err := s.getReportSummary(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener resumen de reportes: %v", err)
	} else {
		summary.Reports = *reportSummary
	}

	// Obtener actividad reciente
	recentActivity, err := s.getRecentActivity(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener actividad reciente: %v", err)
	} else {
		summary.RecentActivity = recentActivity
	}

	// Obtener alertas
	alerts, err := s.getAlerts(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener alertas: %v", err)
	} else {
		summary.Alerts = alerts
	}

	// Obtener métricas principales
	metrics, err := s.getMainMetrics(ctx, unitID, role)
	if err != nil {
		logger.Error("Error al obtener métricas: %v", err)
	} else {
		summary.Metrics = metrics
	}

	return summary, nil
}

// getMessageSummary obtiene el resumen de mensajes
func (s *DashboardService) getMessageSummary(ctx context.Context, unitID int, role string) (*MessageSummary, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := today.AddDate(0, 0, -int(today.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	summary := &MessageSummary{
		ByType:   make(map[string]int64),
		ByStatus: make(map[string]int64),
	}

	// Estadísticas básicas
	stats, err := s.messageRepo.GetStatsByUnit(ctx, unitID, monthStart, now)
	if err != nil {
		return nil, err
	}

	summary.Total = stats.TotalSent + stats.TotalReceived
	summary.Unread = stats.Unread
	summary.Urgent = stats.Urgent

	// Mensajes de hoy
	todayStats, _ := s.messageRepo.GetStatsByUnit(ctx, unitID, today, now)
	summary.Today = todayStats.TotalReceived

	// Mensajes de esta semana
	weekStats, _ := s.messageRepo.GetStatsByUnit(ctx, unitID, weekStart, now)
	summary.ThisWeek = weekStats.TotalReceived

	// Mensajes de este mes
	summary.ThisMonth = stats.TotalReceived

	// Por tipo y estado
	if role == "admin" {
		// Admin ve estadísticas globales
		s.getGlobalMessageStats(ctx, summary)
	}

	// Datos de tendencia (últimos 7 días)
	summary.TrendData = s.getMessageTrend(ctx, unitID, 7)

	return summary, nil
}

// getUserSummary obtiene el resumen de usuarios
func (s *DashboardService) getUserSummary(ctx context.Context, role string) (*UserSummary, error) {
	if role != "admin" {
		// Solo admin puede ver estadísticas de usuarios
		return &UserSummary{}, nil
	}

	stats, err := s.userRepo.GetUserStats(ctx)
	if err != nil {
		return nil, err
	}

	summary := &UserSummary{
		Total:        stats.Total,
		Active:       stats.Active,
		NewThisMonth: stats.RecentRegistrations,
		ByRole:       stats.ByRole,
		ByUnit:       make(map[string]int64),
	}

	// TODO: Implementar conteo de usuarios online y sesiones activas
	summary.Online = 0
	summary.ActiveSessions = 0
	summary.AverageSessionTime = 0

	return summary, nil
}

// getStorageSummary obtiene el resumen de almacenamiento
func (s *DashboardService) getStorageSummary(ctx context.Context, unitID int, role string) (*StorageSummary, error) {
	stats, err := s.fileRepo.GetStorageStats(ctx)
	if err != nil {
		return nil, err
	}

	summary := &StorageSummary{
		TotalSize:    104857600000, // 100GB por defecto
		UsedSize:     stats.TotalSize,
		FileCount:    stats.TotalFiles,
		ByCategory:   make(map[string]int64),
		LargestFiles: []FileInfo{},
	}

	summary.AvailableSize = summary.TotalSize - summary.UsedSize
	summary.UsagePercentage = float64(summary.UsedSize) / float64(summary.TotalSize) * 100

	// Por categoría
	for cat, catStats := range stats.ByCategory {
		summary.ByCategory[cat] = catStats.Size
	}

	// Archivos más grandes
	if role == "admin" {
		// TODO: Implementar obtención de archivos más grandes
	}

	return summary, nil
}

// getReportSummary obtiene el resumen de reportes
func (s *DashboardService) getReportSummary(ctx context.Context, unitID int, role string) (*ReportSummary, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	stats, err := s.reportRepo.GetReportStats(ctx, monthStart, now)
	if err != nil {
		return nil, err
	}

	summary := &ReportSummary{
		TotalGenerated:    stats.Total,
		InProgress:        stats.ByStatus[models.ReportStatusGenerating],
		Scheduled:         stats.ByStatus[models.ReportStatusScheduled],
		Failed:            stats.ByStatus[models.ReportStatusFailed],
		AverageGenTime:    stats.AvgGenerationTime,
		MostUsedTemplates: []TemplateUsage{},
	}

	// Plantillas más usadas
	templates, _ := s.reportRepo.GetMostGeneratedTemplates(ctx, 5)
	for _, t := range templates {
		summary.MostUsedTemplates = append(summary.MostUsedTemplates, TemplateUsage{
			TemplateID:   t.ID.String(),
			TemplateName: t.Name,
			UsageCount:   t.UsageCount,
			Percentage:   float64(t.UsageCount) / float64(stats.Total) * 100,
		})
	}

	return summary, nil
}

// getRecentActivity obtiene la actividad reciente
func (s *DashboardService) getRecentActivity(ctx context.Context, unitID int, role string) ([]ActivityItem, error) {
	filter := &repositories.AuditFilter{
		Limit:    20,
		SortDesc: true,
	}

	// Si no es admin, filtrar por recursos relacionados con la unidad
	if role != "admin" {
		filter.Resource = "messages"
	}

	logs, _, err := s.auditRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	activities := make([]ActivityItem, 0, len(logs))
	for _, log := range logs {
		activity := ActivityItem{
			ID:         fmt.Sprintf("activity_%d", log.ID),
			Type:       "audit",
			Action:     string(log.Action),
			Resource:   log.Resource,
			ResourceID: log.ResourceID,
			Timestamp:  log.CreatedAt,
			IPAddress:  log.IPAddress,
		}

		if log.User != nil {
			activity.User = fmt.Sprintf("%s %s", log.User.FirstName, log.User.LastName)
			activity.UserID = log.User.ID
		}

		// Generar descripción
		activity.Description = s.generateActivityDescription(log)

		activities = append(activities, activity)
	}

	return activities, nil
}

// getAlerts obtiene las alertas activas
func (s *DashboardService) getAlerts(ctx context.Context, unitID int, role string) ([]AlertItem, error) {
	alerts := []AlertItem{}

	// Verificar mensajes urgentes no leídos
	filter := &repositories.MessageFilter{
		ReceiverUnitID: &unitID,
		IsUrgent:       &[]bool{true}[0],
		UnreadOnly:     &[]bool{true}[0],
		Limit:          5,
	}

	messages, _, err := s.messageRepo.GetByFilter(ctx, filter)
	if err == nil && len(messages) > 0 {
		for _, msg := range messages {
			alerts = append(alerts, AlertItem{
				ID:          fmt.Sprintf("alert_msg_%d", msg.ID),
				Type:        "message",
				Severity:    "high",
				Title:       "Mensaje urgente sin leer",
				Description: fmt.Sprintf("Mensaje de %s: %s", msg.SenderUnit.Name, msg.Subject),
				Timestamp:   msg.CreatedAt,
			})
		}
	}

	// Verificar espacio en disco si es admin
	if role == "admin" {
		storageStats, _ := s.fileRepo.GetStorageStats(ctx)
		if storageStats != nil {
			usagePercent := float64(storageStats.TotalSize) / float64(104857600000) * 100
			if usagePercent > 80 {
				alerts = append(alerts, AlertItem{
					ID:          "alert_storage",
					Type:        "system",
					Severity:    "warning",
					Title:       "Espacio de almacenamiento bajo",
					Description: fmt.Sprintf("El almacenamiento está al %.1f%% de capacidad", usagePercent),
					Timestamp:   time.Now(),
				})
			}
		}
	}

	// Verificar intentos fallidos de login
	if role == "admin" {
		failedLogins, _ := s.auditRepo.GetFailedAttempts(ctx, models.AuditActionLogin, 60)
		if len(failedLogins) > 10 {
			alerts = append(alerts, AlertItem{
				ID:          "alert_security",
				Type:        "security",
				Severity:    "critical",
				Title:       "Múltiples intentos fallidos de login",
				Description: fmt.Sprintf("Se detectaron %d intentos fallidos en la última hora", len(failedLogins)),
				Timestamp:   time.Now(),
			})
		}
	}

	return alerts, nil
}

// getMainMetrics obtiene las métricas principales
func (s *DashboardService) getMainMetrics(ctx context.Context, unitID int, role string) ([]MetricItem, error) {
	metrics := []MetricItem{}

	// Tasa de respuesta de mensajes
	responseRate := s.calculateMessageResponseRate(ctx, unitID)
	metrics = append(metrics, MetricItem{
		ID:            "metric_response_rate",
		Name:          "response_rate",
		DisplayName:   "Tasa de Respuesta",
		Value:         responseRate,
		PreviousValue: 85.0, // Valor simulado
		Change:        responseRate - 85.0,
		ChangePercent: ((responseRate - 85.0) / 85.0) * 100,
		Unit:          "%",
		Status:        s.getMetricStatus(responseRate, 85.0),
		Icon:          "message-circle",
		Color:         "#10b981",
	})

	// Tiempo promedio de respuesta
	avgResponseTime := s.calculateAverageResponseTime(ctx, unitID)
	metrics = append(metrics, MetricItem{
		ID:            "metric_avg_response",
		Name:          "avg_response_time",
		DisplayName:   "Tiempo Promedio de Respuesta",
		Value:         avgResponseTime,
		PreviousValue: 4.5,
		Change:        avgResponseTime - 4.5,
		ChangePercent: ((avgResponseTime - 4.5) / 4.5) * 100,
		Unit:          "horas",
		Status:        s.getMetricStatus(4.5, avgResponseTime), // Invertido porque menos es mejor
		Icon:          "clock",
		Color:         "#3b82f6",
	})

	// Eficiencia del sistema
	efficiency := s.calculateSystemEfficiency(ctx)
	metrics = append(metrics, MetricItem{
		ID:            "metric_efficiency",
		Name:          "system_efficiency",
		DisplayName:   "Eficiencia del Sistema",
		Value:         efficiency,
		PreviousValue: 92.0,
		Change:        efficiency - 92.0,
		ChangePercent: ((efficiency - 92.0) / 92.0) * 100,
		Unit:          "%",
		Status:        s.getMetricStatus(efficiency, 92.0),
		Icon:          "trending-up",
		Color:         "#8b5cf6",
	})

	// Satisfacción estimada
	satisfaction := s.calculateUserSatisfaction(ctx)
	metrics = append(metrics, MetricItem{
		ID:            "metric_satisfaction",
		Name:          "user_satisfaction",
		DisplayName:   "Satisfacción Estimada",
		Value:         satisfaction,
		PreviousValue: 88.0,
		Change:        satisfaction - 88.0,
		ChangePercent: ((satisfaction - 88.0) / 88.0) * 100,
		Unit:          "%",
		Status:        s.getMetricStatus(satisfaction, 88.0),
		Icon:          "smile",
		Color:         "#f59e0b",
	})

	return metrics, nil
}

// GetWidgetData obtiene datos específicos para un widget
func (s *DashboardService) GetWidgetData(ctx context.Context, widgetType string, params map[string]interface{}) (interface{}, error) {
	switch widgetType {
	case "message_chart":
		return s.getMessageChartData(ctx, params)
	case "user_activity":
		return s.getUserActivityData(ctx, params)
	case "storage_breakdown":
		return s.getStorageBreakdownData(ctx, params)
	case "report_timeline":
		return s.getReportTimelineData(ctx, params)
	default:
		return nil, fmt.Errorf("tipo de widget no soportado: %s", widgetType)
	}
}

// Funciones auxiliares

// getMessageTrend obtiene la tendencia de mensajes
func (s *DashboardService) getMessageTrend(ctx context.Context, unitID int, days int) []TrendDataPoint {
	trend := []TrendDataPoint{}
	now := time.Now()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		stats, _ := s.messageRepo.GetStatsByUnit(ctx, unitID, startOfDay, endOfDay)

		trend = append(trend, TrendDataPoint{
			Date:  startOfDay,
			Value: float64(stats.TotalReceived),
			Label: startOfDay.Format("02/01"),
		})
	}

	return trend
}

// generateActivityDescription genera una descripción legible de la actividad
func (s *DashboardService) generateActivityDescription(log *models.AuditLog) string {
	actionDescriptions := map[models.AuditAction]string{
		models.AuditActionCreate:  "creó",
		models.AuditActionRead:    "visualizó",
		models.AuditActionUpdate:  "actualizó",
		models.AuditActionDelete:  "eliminó",
		models.AuditActionLogin:   "inició sesión",
		models.AuditActionLogout:  "cerró sesión",
		models.AuditActionExport:  "exportó",
		models.AuditActionImport:  "importó",
		models.AuditActionApprove: "aprobó",
		models.AuditActionReject:  "rechazó",
		models.AuditActionSend:    "envió",
		models.AuditActionReceive: "recibió",
		models.AuditActionArchive: "archivó",
		models.AuditActionRestore: "restauró",
	}

	resourceDescriptions := map[string]string{
		"messages":             "mensaje",
		"users":                "usuario",
		"files":                "archivo",
		"reports":              "reporte",
		"organizational_units": "unidad organizacional",
	}

	action := actionDescriptions[log.Action]
	resource := resourceDescriptions[log.Resource]
	if resource == "" {
		resource = log.Resource
	}

	userName := "Usuario desconocido"
	if log.User != nil {
		userName = fmt.Sprintf("%s %s", log.User.FirstName, log.User.LastName)
	}

	return fmt.Sprintf("%s %s un %s", userName, action, resource)
}

// calculateMessageResponseRate calcula la tasa de respuesta
func (s *DashboardService) calculateMessageResponseRate(ctx context.Context, unitID int) float64 {
	// Implementación simplificada
	// En producción, calcularía basándose en mensajes respondidos vs recibidos
	return 87.5
}

// calculateAverageResponseTime calcula el tiempo promedio de respuesta
func (s *DashboardService) calculateAverageResponseTime(ctx context.Context, unitID int) float64 {
	// Implementación simplificada
	// En producción, calcularía basándose en timestamps de creación y respuesta
	return 3.8
}

// calculateSystemEfficiency calcula la eficiencia del sistema
func (s *DashboardService) calculateSystemEfficiency(ctx context.Context) float64 {
	// Implementación simplificada
	// En producción, consideraría múltiples factores como uptime, errores, etc.
	return 94.2
}

// calculateUserSatisfaction calcula la satisfacción estimada
func (s *DashboardService) calculateUserSatisfaction(ctx context.Context) float64 {
	// Implementación simplificada
	// En producción, consideraría tiempos de respuesta, tasas de error, etc.
	return 89.5
}

// getMetricStatus determina el estado de una métrica
func (s *DashboardService) getMetricStatus(current, previous float64) string {
	change := current - previous
	percentChange := (change / previous) * 100

	if percentChange > 5 {
		return "up"
	} else if percentChange < -5 {
		return "down"
	}
	return "stable"
}

// getGlobalMessageStats obtiene estadísticas globales de mensajes
func (s *DashboardService) getGlobalMessageStats(ctx context.Context, summary *MessageSummary) {
	// Obtener tipos de mensaje
	types, _ := s.messageRepo.GetMessageTypes(ctx)
	for _, t := range types {
		// TODO: Contar mensajes por tipo
		summary.ByType[t.Name] = 0
	}

	// Obtener estados
	statuses, _ := s.messageRepo.GetMessageStatuses(ctx)
	for _, status := range statuses {
		// TODO: Contar mensajes por estado
		summary.ByStatus[status.Name] = 0
	}
}

// Métodos para obtener datos de widgets específicos

func (s *DashboardService) getMessageChartData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementación para obtener datos del gráfico de mensajes
	return map[string]interface{}{
		"labels": []string{"Lun", "Mar", "Mié", "Jue", "Vie", "Sáb", "Dom"},
		"datasets": []map[string]interface{}{
			{
				"label": "Enviados",
				"data":  []int{12, 19, 3, 5, 2, 3, 8},
				"color": "#3b82f6",
			},
			{
				"label": "Recibidos",
				"data":  []int{8, 11, 7, 9, 5, 2, 4},
				"color": "#10b981",
			},
		},
	}, nil
}

func (s *DashboardService) getUserActivityData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementación para obtener datos de actividad de usuarios
	return map[string]interface{}{
		"activeNow":      12,
		"peakToday":      45,
		"averageToday":   28,
		"hourlyActivity": []int{5, 8, 15, 22, 35, 42, 38, 45, 40, 35, 28, 20, 15, 12},
	}, nil
}

func (s *DashboardService) getStorageBreakdownData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementación para obtener desglose de almacenamiento
	return map[string]interface{}{
		"categories": []map[string]interface{}{
			{"name": "Documentos", "value": 35, "size": 3670016000},
			{"name": "Imágenes", "value": 25, "size": 2621440000},
			{"name": "Reportes", "value": 20, "size": 2097152000},
			{"name": "Adjuntos", "value": 15, "size": 1572864000},
			{"name": "Otros", "value": 5, "size": 524288000},
		},
		"totalSize": 10485760000,
	}, nil
}

func (s *DashboardService) getReportTimelineData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Implementación para obtener línea de tiempo de reportes
	return map[string]interface{}{
		"timeline": []map[string]interface{}{
			{
				"date":   time.Now().Add(-5 * time.Hour),
				"type":   "generated",
				"report": "Reporte Mensual de Actividades",
				"status": "completed",
			},
			{
				"date":   time.Now().Add(-3 * time.Hour),
				"type":   "scheduled",
				"report": "Reporte Semanal de Mensajes",
				"status": "pending",
			},
			{
				"date":   time.Now().Add(-1 * time.Hour),
				"type":   "failed",
				"report": "Reporte de Auditoría",
				"status": "failed",
				"error":  "Timeout al generar datos",
			},
		},
	}, nil
}

// internal/types/responses/dashboard_responses.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// DashboardResponse respuesta principal del dashboard
type DashboardResponse struct {
	Metrics    []MetricResponse   `json:"metrics"`
	Charts     []ChartResponse    `json:"charts"`
	Activities []ActivityResponse `json:"activities"`
	Alerts     []AlertResponse    `json:"alerts"`
	Summary    DashboardSummary   `json:"summary"`
	LastUpdate time.Time          `json:"lastUpdate"`
}

// MetricResponse métrica individual
type MetricResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	DisplayName    string    `json:"displayName"`
	Description    string    `json:"description,omitempty"`
	Type           string    `json:"type"`
	Category       string    `json:"category"`
	Value          float64   `json:"value"`
	FormattedValue string    `json:"formattedValue"`
	PreviousValue  float64   `json:"previousValue,omitempty"`
	Change         float64   `json:"change,omitempty"`
	ChangePercent  float64   `json:"changePercent,omitempty"`
	Trend          string    `json:"trend"` // up, down, stable
	Unit           string    `json:"unit,omitempty"`
	Icon           string    `json:"icon,omitempty"`
	Color          string    `json:"color,omitempty"`
	Status         string    `json:"status"` // normal, warning, alert, success
	IsAlert        bool      `json:"isAlert"`
	TargetValue    float64   `json:"targetValue,omitempty"`
	Progress       float64   `json:"progress,omitempty"`
	CalculatedAt   time.Time `json:"calculatedAt"`
}

// ChartResponse gráfico del dashboard
type ChartResponse struct {
	ID        uuid.UUID    `json:"id"`
	Type      string       `json:"type"` // line, bar, pie, donut, area, scatter
	Title     string       `json:"title"`
	Subtitle  string       `json:"subtitle,omitempty"`
	Data      ChartData    `json:"data"`
	Options   ChartOptions `json:"options"`
	Period    string       `json:"period"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// ChartData datos del gráfico
type ChartData struct {
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

// ChartDataset conjunto de datos
type ChartDataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor,omitempty"`
	BorderColor     string    `json:"borderColor,omitempty"`
	Fill            bool      `json:"fill,omitempty"`
}

// ChartOptions opciones del gráfico
type ChartOptions struct {
	Responsive     bool                   `json:"responsive"`
	MaintainAspect bool                   `json:"maintainAspectRatio"`
	Legend         LegendOptions          `json:"legend"`
	Scales         map[string]interface{} `json:"scales,omitempty"`
}

// LegendOptions opciones de leyenda
type LegendOptions struct {
	Display  bool   `json:"display"`
	Position string `json:"position"`
}

// ActivityResponse actividad reciente
type ActivityResponse struct {
	ID          uuid.UUID   `json:"id"`
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	User        UserSummary `json:"user"`
	Icon        string      `json:"icon,omitempty"`
	Color       string      `json:"color,omitempty"`
	Link        string      `json:"link,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// AlertResponse alerta del dashboard
type AlertResponse struct {
	ID          uuid.UUID     `json:"id"`
	Type        string        `json:"type"`     // info, warning, danger, success
	Severity    string        `json:"severity"` // low, medium, high, critical
	Title       string        `json:"title"`
	Message     string        `json:"message"`
	Source      string        `json:"source"`
	MetricID    *uuid.UUID    `json:"metricId,omitempty"`
	IsRead      bool          `json:"isRead"`
	IsDismissed bool          `json:"isDismissed"`
	Actions     []AlertAction `json:"actions,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
}

// AlertAction acción de alerta
type AlertAction struct {
	Label  string `json:"label"`
	Action string `json:"action"`
	Style  string `json:"style"` // primary, secondary, danger
}

// DashboardSummary resumen del dashboard
type DashboardSummary struct {
	Period              string    `json:"period"`
	TotalMessages       int64     `json:"totalMessages"`
	ActiveUsers         int64     `json:"activeUsers"`
	SystemHealth        string    `json:"systemHealth"`
	StorageUsed         int64     `json:"storageUsed"`
	StorageAvailable    int64     `json:"storageAvailable"`
	AverageResponseTime float64   `json:"averageResponseTime"`
	LastBackup          time.Time `json:"lastBackup"`
}

// MetricHistoryResponse historial de métricas
type MetricHistoryResponse struct {
	MetricID   uuid.UUID         `json:"metricId"`
	MetricName string            `json:"metricName"`
	Period     string            `json:"period"`
	History    []MetricDataPoint `json:"history"`
	Statistics MetricStatistics  `json:"statistics"`
}

// MetricDataPoint punto de datos
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Min       float64   `json:"min,omitempty"`
	Max       float64   `json:"max,omitempty"`
	Avg       float64   `json:"avg,omitempty"`
}

// MetricStatistics estadísticas de métrica
type MetricStatistics struct {
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Average      float64 `json:"average"`
	StdDeviation float64 `json:"stdDeviation"`
	Trend        float64 `json:"trend"`
}

// WidgetResponse widget del dashboard
type WidgetResponse struct {
	ID              uuid.UUID              `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Position        int                    `json:"position"`
	Width           int                    `json:"width"`
	Height          int                    `json:"height"`
	Configuration   map[string]interface{} `json:"configuration"`
	Data            interface{}            `json:"data"`
	RefreshInterval int                    `json:"refreshInterval,omitempty"`
	LastRefresh     time.Time              `json:"lastRefresh"`
}

// SystemOverviewResponse vista general del sistema
type SystemOverviewResponse struct {
	Status      string             `json:"status"` // healthy, warning, critical
	Uptime      int64              `json:"uptime"` // segundos
	Version     string             `json:"version"`
	Components  []ComponentStatus  `json:"components"`
	Resources   ResourceUsage      `json:"resources"`
	Performance PerformanceMetrics `json:"performance"`
}

// ComponentStatus estado de componente
type ComponentStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
	LastCheck time.Time `json:"lastCheck"`
}

// ResourceUsage uso de recursos
type ResourceUsage struct {
	CPU     float64      `json:"cpu"`
	Memory  float64      `json:"memory"`
	Disk    float64      `json:"disk"`
	Network NetworkUsage `json:"network"`
}

// NetworkUsage uso de red
type NetworkUsage struct {
	BytesIn  int64 `json:"bytesIn"`
	BytesOut int64 `json:"bytesOut"`
	Requests int64 `json:"requests"`
}

// PerformanceMetrics métricas de rendimiento
type PerformanceMetrics struct {
	ResponseTime float64 `json:"responseTime"` // ms
	Throughput   int64   `json:"throughput"`   // requests/sec
	ErrorRate    float64 `json:"errorRate"`    // percentage
	Latency      float64 `json:"latency"`      // ms
}

// CustomDashboardResponse dashboard personalizado
type CustomDashboardResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Layout      []WidgetLayout   `json:"layout"`
	Widgets     []WidgetResponse `json:"widgets"`
	IsDefault   bool             `json:"isDefault"`
	IsShared    bool             `json:"isShared"`
	Owner       UserSummary      `json:"owner"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

// WidgetLayout disposición de widget
type WidgetLayout struct {
	WidgetID uuid.UUID `json:"widgetId"`
	X        int       `json:"x"`
	Y        int       `json:"y"`
	Width    int       `json:"width"`
	Height   int       `json:"height"`
}

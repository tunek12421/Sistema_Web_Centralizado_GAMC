// internal/database/models/dashboard_metric.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetricType define los tipos de métricas
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeAverage   MetricType = "average"
	MetricTypeSum       MetricType = "sum"
	MetricTypeRate      MetricType = "rate"
)

// MetricPeriod define los períodos de agregación
type MetricPeriod string

const (
	MetricPeriodRealtime MetricPeriod = "realtime"
	MetricPeriodHourly   MetricPeriod = "hourly"
	MetricPeriodDaily    MetricPeriod = "daily"
	MetricPeriodWeekly   MetricPeriod = "weekly"
	MetricPeriodMonthly  MetricPeriod = "monthly"
	MetricPeriodYearly   MetricPeriod = "yearly"
)

// DashboardMetric representa una métrica del dashboard
type DashboardMetric struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name            string                 `json:"name" gorm:"size:100;not null;uniqueIndex:idx_metric_org_period"`
	DisplayName     string                 `json:"displayName" gorm:"size:255;not null"`
	Description     string                 `json:"description,omitempty" gorm:"type:text"`
	Type            MetricType             `json:"type" gorm:"type:varchar(20);not null"`
	Category        string                 `json:"category" gorm:"size:50;index"`
	Unit            string                 `json:"unit,omitempty" gorm:"size:20"`
	Value           float64                `json:"value" gorm:"not null"`
	PreviousValue   float64                `json:"previousValue,omitempty"`
	TargetValue     float64                `json:"targetValue,omitempty"`
	MinValue        float64                `json:"minValue,omitempty"`
	MaxValue        float64                `json:"maxValue,omitempty"`
	Period          MetricPeriod           `json:"period" gorm:"type:varchar(20);not null;uniqueIndex:idx_metric_org_period"`
	OrganizationID  *int                   `json:"organizationId,omitempty" gorm:"uniqueIndex:idx_metric_org_period"`
	IsGlobal        bool                   `json:"isGlobal" gorm:"default:false"`
	Icon            string                 `json:"icon,omitempty" gorm:"size:50"`
	Color           string                 `json:"color,omitempty" gorm:"size:20"`
	SortOrder       int                    `json:"sortOrder" gorm:"default:0"`
	IsVisible       bool                   `json:"isVisible" gorm:"default:true"`
	AlertEnabled    bool                   `json:"alertEnabled" gorm:"default:false"`
	AlertThreshold  float64                `json:"alertThreshold,omitempty"`
	AlertComparator string                 `json:"alertComparator,omitempty" gorm:"size:10"` // >, <, >=, <=, ==, !=
	Metadata        map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CalculatedAt    time.Time              `json:"calculatedAt"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`

	// Relaciones
	Organization *OrganizationalUnit `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
}

// TableName especifica el nombre de la tabla
func (DashboardMetric) TableName() string {
	return "dashboard_metrics"
}

// BeforeCreate hook para establecer valores por defecto
func (d *DashboardMetric) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	d.CalculatedAt = time.Now()
	return nil
}

// CalculateChange calcula el cambio porcentual
func (d *DashboardMetric) CalculateChange() float64 {
	if d.PreviousValue == 0 {
		return 0
	}
	return ((d.Value - d.PreviousValue) / d.PreviousValue) * 100
}

// IsAlert verifica si la métrica debe generar una alerta
func (d *DashboardMetric) IsAlert() bool {
	if !d.AlertEnabled || d.AlertComparator == "" {
		return false
	}

	switch d.AlertComparator {
	case ">":
		return d.Value > d.AlertThreshold
	case "<":
		return d.Value < d.AlertThreshold
	case ">=":
		return d.Value >= d.AlertThreshold
	case "<=":
		return d.Value <= d.AlertThreshold
	case "==":
		return d.Value == d.AlertThreshold
	case "!=":
		return d.Value != d.AlertThreshold
	default:
		return false
	}
}

// GetStatus retorna el estado de la métrica
func (d *DashboardMetric) GetStatus() string {
	if d.IsAlert() {
		return "alert"
	}

	if d.TargetValue > 0 {
		progress := (d.Value / d.TargetValue) * 100
		if progress >= 100 {
			return "success"
		} else if progress >= 75 {
			return "warning"
		} else {
			return "info"
		}
	}

	change := d.CalculateChange()
	if change > 10 {
		return "success"
	} else if change < -10 {
		return "danger"
	}

	return "normal"
}

// DashboardMetricHistory historial de métricas
type DashboardMetricHistory struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	MetricID     uuid.UUID `json:"metricId" gorm:"type:uuid;not null;index"`
	Value        float64   `json:"value" gorm:"not null"`
	Period       string    `json:"period" gorm:"size:20;not null"`
	PeriodStart  time.Time `json:"periodStart" gorm:"not null;index"`
	PeriodEnd    time.Time `json:"periodEnd" gorm:"not null"`
	SampleCount  int64     `json:"sampleCount,omitempty"`
	MinValue     float64   `json:"minValue,omitempty"`
	MaxValue     float64   `json:"maxValue,omitempty"`
	AvgValue     float64   `json:"avgValue,omitempty"`
	StdDeviation float64   `json:"stdDeviation,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`

	// Relaciones
	Metric *DashboardMetric `json:"metric,omitempty" gorm:"foreignKey:MetricID"`
}

// TableName especifica el nombre de la tabla
func (DashboardMetricHistory) TableName() string {
	return "dashboard_metric_history"
}

// DashboardWidget configuración de widgets del dashboard
type DashboardWidget struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID              `json:"userId" gorm:"type:uuid;not null;index"`
	Name            string                 `json:"name" gorm:"size:100;not null"`
	Type            string                 `json:"type" gorm:"size:50;not null"` // chart, metric, table, map
	Position        int                    `json:"position" gorm:"not null"`
	Width           int                    `json:"width" gorm:"default:6"`  // Grid de 12 columnas
	Height          int                    `json:"height" gorm:"default:4"` // Unidades de altura
	MetricIDs       []uuid.UUID            `json:"metricIds,omitempty" gorm:"type:uuid[]"`
	ChartType       string                 `json:"chartType,omitempty" gorm:"size:50"` // line, bar, pie, etc
	Configuration   map[string]interface{} `json:"configuration,omitempty" gorm:"type:jsonb"`
	RefreshInterval int                    `json:"refreshInterval,omitempty"` // segundos
	IsVisible       bool                   `json:"isVisible" gorm:"default:true"`
	CreatedAt       time.Time              `json:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt"`

	// Relaciones
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName especifica el nombre de la tabla
func (DashboardWidget) TableName() string {
	return "dashboard_widgets"
}

package models

// UserSummary representa un resumen de usuarios
type UserSummary struct {
	Total              int64            `json:"total"`
	Active             int64            `json:"active"`
	NewThisMonth       int64            `json:"newThisMonth"`
	Online             int64            `json:"online"`
	ActiveSessions     int64            `json:"activeSessions"`
	AverageSessionTime float64          `json:"averageSessionTime"`
	ByRole             map[string]int64 `json:"byRole"`
	ByUnit             map[int]int64    `json:"byUnit"`
}

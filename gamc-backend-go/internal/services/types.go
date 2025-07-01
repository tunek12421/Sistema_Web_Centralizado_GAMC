// internal/services/types.go
package services

import (
	"fmt"

	"github.com/google/uuid"
)

// AuditUserSummary resumen de usuario para logs de auditoría
type AuditUserSummary struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	FullName string    `json:"fullName"`
	Email    string    `json:"email"`
}

// Funciones auxiliares para conversión de metadata

// ConvertMetadata convierte map[string]string a map[string]interface{}
func ConvertMetadata(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// ConvertMetadataToString convierte map[string]interface{} a map[string]string
func ConvertMetadataToString(m map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// ConvertMetadataInterface convierte map[string]string a map[string]interface{} para reportes
func ConvertMetadataInterface(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

// pkg/validator/message_validator.go
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidateCreateMessage valida los datos para crear un mensaje
func ValidateCreateMessage(subject, content string, receiverUnitID, messageTypeID, priorityLevel int) error {
	// Validar subject
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("el asunto es requerido")
	}
	if len(subject) < 3 {
		return fmt.Errorf("el asunto debe tener al menos 3 caracteres")
	}
	if len(subject) > 255 {
		return fmt.Errorf("el asunto no puede exceder 255 caracteres")
	}

	// Validar content
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("el contenido es requerido")
	}
	if len(content) < 10 {
		return fmt.Errorf("el contenido debe tener al menos 10 caracteres")
	}

	// Validar receiverUnitID
	if receiverUnitID <= 0 {
		return fmt.Errorf("la unidad destinataria es requerida")
	}

	// Validar messageTypeID
	if messageTypeID <= 0 {
		return fmt.Errorf("el tipo de mensaje es requerido")
	}

	// Validar priorityLevel
	if priorityLevel < 1 || priorityLevel > 4 {
		return fmt.Errorf("el nivel de prioridad debe estar entre 1 y 4")
	}

	return nil
}

// ValidateMessageStatus valida que el estado sea válido
func ValidateMessageStatus(status string) error {
	validStatuses := []string{
		"DRAFT", "SENT", "READ", "IN_PROGRESS",
		"RESPONDED", "RESOLVED", "ARCHIVED", "CANCELLED",
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return fmt.Errorf("estado inválido: %s", status)
}

// ValidateMessageFilters valida los filtros de búsqueda
func ValidateMessageFilters(page, limit int, sortBy, sortOrder string) error {
	// Validar paginación
	if page < 1 {
		return fmt.Errorf("la página debe ser mayor a 0")
	}
	if limit < 1 || limit > 100 {
		return fmt.Errorf("el límite debe estar entre 1 y 100")
	}

	// Validar sortBy
	if sortBy != "" {
		validSortFields := []string{"created_at", "subject", "priority_level"}
		isValid := false
		for _, field := range validSortFields {
			if sortBy == field {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("campo de ordenamiento inválido: %s", sortBy)
		}
	}

	// Validar sortOrder
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return fmt.Errorf("orden inválido, debe ser 'asc' o 'desc'")
	}

	return nil
}

// RegisterMessageValidators registra validadores personalizados para mensajes
func RegisterMessageValidators(validate *validator.Validate) {
	// Validador para estados de mensaje
	validate.RegisterValidation("message_status", func(fl validator.FieldLevel) bool {
		status := fl.Field().String()
		return ValidateMessageStatus(status) == nil
	})

	// Validador para prioridad de mensaje
	validate.RegisterValidation("message_priority", func(fl validator.FieldLevel) bool {
		priority := int(fl.Field().Int())
		return priority >= 1 && priority <= 4
	})

	// Validador para contenido de mensaje (sin HTML básico)
	validate.RegisterValidation("message_content", func(fl validator.FieldLevel) bool {
		content := fl.Field().String()
		// Validación básica - no permitir tags script
		return !strings.Contains(strings.ToLower(content), "<script")
	})
}

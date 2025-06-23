// pkg/validator/validator.go
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Registrar validaciones personalizadas
	validate.RegisterValidation("gamc_email", validateGAMCEmail)
}

// Validate valida una estructura usando tags
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrors)
		}
		return err
	}
	return nil
}

// formatValidationErrors formatea los errores de validación
func formatValidationErrors(errors validator.ValidationErrors) error {
	var messages []string

	for _, err := range errors {
		message := getErrorMessage(err)
		messages = append(messages, message)
	}

	return fmt.Errorf(strings.Join(messages, "; "))
}

// getErrorMessage obtiene un mensaje de error legible
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s es requerido", field)
	case "email":
		return fmt.Sprintf("%s debe ser un email válido", field)
	case "min":
		return fmt.Sprintf("%s debe tener al menos %s caracteres", field, err.Param())
	case "max":
		return fmt.Sprintf("%s no puede tener más de %s caracteres", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s debe ser uno de: %s", field, err.Param())
	case "gamc_email":
		return fmt.Sprintf("%s debe ser un email institucional (@gamc.gov.bo)", field)
	default:
		return fmt.Sprintf("%s no es válido", field)
	}
}

// validateGAMCEmail validación personalizada para emails del GAMC
func validateGAMCEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return strings.HasSuffix(email, "@gamc.gov.bo")
}

// IsValidEmail verifica si un email es válido
func IsValidEmail(email string) bool {
	return validate.Var(email, "email") == nil
}

// IsValidPassword verifica si una contraseña cumple requisitos
func IsValidPassword(password string) (bool, []string) {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "Debe tener al menos 8 caracteres")
	}

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("@$!%*?&", char):
			hasSpecial = true
		}
	}

	if !hasLower {
		errors = append(errors, "Debe contener al menos una letra minúscula")
	}
	if !hasUpper {
		errors = append(errors, "Debe contener al menos una letra mayúscula")
	}
	if !hasDigit {
		errors = append(errors, "Debe contener al menos un número")
	}
	if !hasSpecial {
		errors = append(errors, "Debe contener al menos un carácter especial (@$!%*?&)")
	}

	return len(errors) == 0, errors
}

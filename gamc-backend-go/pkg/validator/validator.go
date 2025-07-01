// pkg/validator/validator.go
package validator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Registrar validaciones personalizadas
	validate.RegisterValidation("gamc_email", validateGAMCEmail)
	validate.RegisterValidation("security_answer", validateSecurityAnswer)
	validate.RegisterValidation("no_repeated_chars", validateNoRepeatedChars)
	validate.RegisterValidation("token_hex", validateTokenHex)
	validate.RegisterValidation("strong_password", validateStrongPassword)
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
	case "len":
		return fmt.Sprintf("%s debe tener exactamente %s caracteres", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s debe ser uno de: %s", field, err.Param())
	case "gamc_email":
		return fmt.Sprintf("%s debe ser un email institucional (@gamc.gov.bo)", field)
	case "security_answer":
		return fmt.Sprintf("%s debe ser una respuesta válida (2-100 caracteres, sin solo espacios)", field)
	case "no_repeated_chars":
		return fmt.Sprintf("%s no puede ser solo caracteres repetidos", field)
	case "token_hex":
		return fmt.Sprintf("%s debe ser un token hexadecimal válido de 64 caracteres", field)
	case "strong_password":
		return fmt.Sprintf("%s debe cumplir con los requisitos de seguridad", field)
	default:
		return fmt.Sprintf("%s no es válido", field)
	}
}

// ========================================
// VALIDACIONES PERSONALIZADAS
// ========================================

// validateGAMCEmail validación personalizada para emails del GAMC
func validateGAMCEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return strings.HasSuffix(email, "@gamc.gov.bo")
}

// validateSecurityAnswer validación para respuestas de seguridad
func validateSecurityAnswer(fl validator.FieldLevel) bool {
	answer := strings.TrimSpace(fl.Field().String())

	// Verificar longitud
	if len(answer) < 2 || len(answer) > 100 {
		return false
	}

	// Verificar que no sea solo espacios
	if len(answer) == 0 {
		return false
	}

	// Verificar que no sea solo caracteres repetidos
	return !isOnlyRepeatedChars(answer)
}

// validateNoRepeatedChars validación para evitar caracteres repetidos
func validateNoRepeatedChars(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return !isOnlyRepeatedChars(value)
}

// validateTokenHex validación para tokens hexadecimales
func validateTokenHex(fl validator.FieldLevel) bool {
	token := fl.Field().String()

	// Debe tener exactamente 64 caracteres
	if len(token) != 64 {
		return false
	}

	// Debe ser hexadecimal válido
	matched, _ := regexp.MatchString("^[a-fA-F0-9]{64}$", token)
	return matched
}

// validateStrongPassword validación avanzada de contraseñas
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	valid, _ := IsValidPassword(password)
	return valid
}

// ========================================
// FUNCIONES DE VALIDACIÓN DIRECTA
// ========================================

// IsValidEmail verifica si un email es válido
func IsValidEmail(email string) bool {
	return validate.Var(email, "email") == nil
}

// IsValidGAMCEmail verifica si un email es del GAMC
func IsValidGAMCEmail(email string) bool {
	return validate.Var(email, "email,gamc_email") == nil
}

// IsValidPassword verifica si una contraseña cumple requisitos
func IsValidPassword(password string) (bool, []string) {
	var errors []string

	// Longitud mínima
	if len(password) < 8 {
		errors = append(errors, "Debe tener al menos 8 caracteres")
	}

	// Longitud máxima
	if len(password) > 128 {
		errors = append(errors, "No puede tener más de 128 caracteres")
	}

	// Verificar tipos de caracteres
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsDigit(char):
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

	// Verificar patrones débiles
	if isCommonPassword(password) {
		errors = append(errors, "No puede ser una contraseña común")
	}

	if hasSequentialChars(password) {
		errors = append(errors, "No puede contener secuencias obvias (123, abc, etc.)")
	}

	return len(errors) == 0, errors
}

// IsValidSecurityAnswer verifica si una respuesta de seguridad es válida
func IsValidSecurityAnswer(answer string) (bool, []string) {
	var errors []string

	answer = strings.TrimSpace(answer)

	// Longitud
	if len(answer) < 2 {
		errors = append(errors, "Debe tener al menos 2 caracteres")
	}
	if len(answer) > 100 {
		errors = append(errors, "No puede tener más de 100 caracteres")
	}

	// No puede estar vacía
	if len(answer) == 0 {
		errors = append(errors, "No puede estar vacía")
	}

	// No puede ser solo caracteres repetidos
	if isOnlyRepeatedChars(answer) {
		errors = append(errors, "No puede ser solo caracteres repetidos")
	}

	// No puede ser solo números
	if isOnlyNumbers(answer) {
		errors = append(errors, "No puede ser solo números")
	}

	// No puede ser solo espacios
	if len(strings.TrimSpace(answer)) == 0 {
		errors = append(errors, "No puede ser solo espacios")
	}

	return len(errors) == 0, errors
}

// IsValidResetToken verifica si un token de reset es válido
func IsValidResetToken(token string) bool {
	return validate.Var(token, "required,len=64,token_hex") == nil
}

// IsValidUserRole verifica si un rol de usuario es válido
func IsValidUserRole(role string) bool {
	return validate.Var(role, "required,oneof=admin input output") == nil
}

// ========================================
// VALIDACIONES ESPECÍFICAS PARA PREGUNTAS DE SEGURIDAD
// ========================================

// ValidateSecurityQuestionSetup valida la configuración completa de preguntas
func ValidateSecurityQuestionSetup(questions []SecurityQuestionRequest) []string {
	var errors []string

	// Verificar cantidad
	if len(questions) == 0 {
		errors = append(errors, "Debe configurar al menos 1 pregunta de seguridad")
	}
	if len(questions) > 3 {
		errors = append(errors, "No puede configurar más de 3 preguntas de seguridad")
	}

	// Verificar unicidad de preguntas
	questionIDs := make(map[int]bool)
	for i, q := range questions {
		// Verificar pregunta duplicada
		if questionIDs[q.QuestionID] {
			errors = append(errors, "No puede seleccionar la misma pregunta múltiples veces")
			break
		}
		questionIDs[q.QuestionID] = true

		// Validar ID de pregunta
		if q.QuestionID <= 0 {
			errors = append(errors, fmt.Sprintf("ID de pregunta %d inválido", i+1))
		}

		// Validar respuesta
		if valid, respErrors := IsValidSecurityAnswer(q.Answer); !valid {
			for _, respErr := range respErrors {
				errors = append(errors, fmt.Sprintf("Pregunta %d: %s", i+1, respErr))
			}
		}
	}

	return errors
}

// ========================================
// FUNCIONES AUXILIARES
// ========================================

// isOnlyRepeatedChars verifica si una cadena es solo caracteres repetidos
func isOnlyRepeatedChars(s string) bool {
	if len(s) <= 1 {
		return false
	}

	first := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != first {
			return false
		}
	}
	return true
}

// isOnlyNumbers verifica si una cadena es solo números
func isOnlyNumbers(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

// isCommonPassword verifica si es una contraseña común
func isCommonPassword(password string) bool {
	commonPasswords := []string{
		"password", "123456", "123456789", "12345678", "12345",
		"password123", "admin", "administrator", "root", "user",
		"gamc", "bolivia", "lapaz", "gamc123", "admin123",
		"qwerty", "abc123", "welcome", "login", "pass",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}
	return false
}

// hasSequentialChars verifica si tiene secuencias obvias
func hasSequentialChars(password string) bool {
	sequences := []string{
		"123", "234", "345", "456", "567", "678", "789",
		"abc", "bcd", "cde", "def", "efg", "fgh", "ghi",
		"hij", "ijk", "jkl", "klm", "lmn", "mno", "nop",
		"opq", "pqr", "qrs", "rst", "stu", "tuv", "uvw",
		"vwx", "wxy", "xyz",
	}

	lowerPassword := strings.ToLower(password)
	for _, seq := range sequences {
		if strings.Contains(lowerPassword, seq) {
			return true
		}
	}
	return false
}

// ========================================
// ESTRUCTURAS AUXILIARES PARA VALIDACIÓN
// ========================================

// SecurityQuestionRequest estructura auxiliar para validación
type SecurityQuestionRequest struct {
	QuestionID int    `json:"questionId" validate:"required,min=1"`
	Answer     string `json:"answer" validate:"required,security_answer"`
}

// PasswordResetRequest estructura auxiliar para validación
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email,gamc_email"`
}

// PasswordResetConfirm estructura auxiliar para validación
type PasswordResetConfirm struct {
	Token       string `json:"token" validate:"required,token_hex"`
	NewPassword string `json:"newPassword" validate:"required,strong_password"`
}

// ========================================
// FUNCIONES DE VALIDACIÓN RÁPIDA
// ========================================

// QuickValidateEmail validación rápida de email
func QuickValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email es requerido")
	}
	if !IsValidEmail(email) {
		return fmt.Errorf("email debe ser válido")
	}
	return nil
}

// QuickValidateGAMCEmail validación rápida de email institucional
func QuickValidateGAMCEmail(email string) error {
	if err := QuickValidateEmail(email); err != nil {
		return err
	}
	if !IsValidGAMCEmail(email) {
		return fmt.Errorf("email debe ser institucional (@gamc.gov.bo)")
	}
	return nil
}

// QuickValidatePassword validación rápida de contraseña
func QuickValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("contraseña es requerida")
	}
	if valid, errors := IsValidPassword(password); !valid {
		return fmt.Errorf("contraseña inválida: %s", strings.Join(errors, ", "))
	}
	return nil
}

// QuickValidateSecurityAnswer validación rápida de respuesta
func QuickValidateSecurityAnswer(answer string) error {
	if answer == "" {
		return fmt.Errorf("respuesta es requerida")
	}
	if valid, errors := IsValidSecurityAnswer(answer); !valid {
		return fmt.Errorf("respuesta inválida: %s", strings.Join(errors, ", "))
	}
	return nil
}

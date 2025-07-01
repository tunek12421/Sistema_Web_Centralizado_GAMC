// internal/auth/password.go
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost es el costo por defecto para bcrypt
	DefaultCost = 12
)

// PasswordService maneja el hashing y verificación de contraseñas
type PasswordService struct {
	cost int
}

// NewPasswordService crea una nueva instancia del servicio de contraseñas
func NewPasswordService() *PasswordService {
	return &PasswordService{cost: DefaultCost}
}

// HashPassword hashea una contraseña usando bcrypt
func (p *PasswordService) HashPassword(password string) (string, error) {
	if len(password) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePassword compara una contraseña en texto plano con su hash
func (p *PasswordService) ComparePassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// IsValidPassword verifica si una contraseña cumple con los requisitos
func (p *PasswordService) IsValidPassword(password string) (bool, []string) {
	var errors []string

	// Longitud mínima
	if len(password) < 8 {
		errors = append(errors, "La contraseña debe tener al menos 8 caracteres")
	}

	// Verificar al menos una minúscula
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		errors = append(errors, "Debe contener al menos una letra minúscula")
	}

	// Verificar al menos una mayúscula
	hasUpper := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		errors = append(errors, "Debe contener al menos una letra mayúscula")
	}

	// Verificar al menos un número
	hasDigit := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		errors = append(errors, "Debe contener al menos un número")
	}

	// Verificar al menos un carácter especial
	specialChars := "@$!%*?&"
	hasSpecial := false
	for _, char := range password {
		for _, special := range specialChars {
			if char == special {
				hasSpecial = true
				break
			}
		}
		if hasSpecial {
			break
		}
	}
	if !hasSpecial {
		errors = append(errors, "Debe contener al menos un carácter especial (@$!%*?&)")
	}

	return len(errors) == 0, errors
}

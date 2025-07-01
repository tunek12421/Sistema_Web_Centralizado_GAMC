// pkg/utils/utils.go
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateRandomString genera una cadena aleatoria de longitud específica
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

// CleanString limpia una cadena removiendo acentos y caracteres especiales
func CleanString(s string) string {
	// Normalizar y remover acentos
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)
	
	// Convertir a minúsculas
	s = strings.ToLower(s)
	
	// Remover espacios y caracteres especiales
	reg := regexp.MustCompile(`[^a-z0-9]`)
	s = reg.ReplaceAllString(s, "")
	
	return s
}

// IsValidUUID verifica si una cadena es un UUID válido
func IsValidUUID(uuid string) bool {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidPattern, uuid)
	return matched
}

// Contains verifica si un slice contiene un elemento
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// RemoveFromSlice remueve un elemento de un slice
func RemoveFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

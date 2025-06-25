// internal/database/models/security_question.go
package models

// ===== IMPORTS NECESARIOS =====
import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SecurityQuestion representa una pregunta de seguridad predefinida
type SecurityQuestion struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	QuestionText string    `json:"questionText" gorm:"type:text;not null"`
	Category     string    `json:"category" gorm:"size:50;default:'general'"`
	IsActive     bool      `json:"isActive" gorm:"default:true;index"`
	SortOrder    int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt    time.Time `json:"createdAt"`
}

// UserSecurityQuestion representa la respuesta de seguridad de un usuario
type UserSecurityQuestion struct {
	ID                 int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID             uuid.UUID `json:"userId" gorm:"type:uuid;not null;index"`
	SecurityQuestionID int       `json:"securityQuestionId" gorm:"not null"`
	AnswerHash         string    `json:"-" gorm:"size:255;not null"` // Nunca exponer en JSON
	IsActive           bool      `json:"isActive" gorm:"default:true;index"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`

	// Relaciones
	User             *User             `json:"user,omitempty" gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	SecurityQuestion *SecurityQuestion `json:"securityQuestion,omitempty" gorm:"foreignKey:SecurityQuestionID"`
}

// ===== ESTRUCTURAS PARA REQUESTS =====

// SecurityQuestionSetupRequest para configurar preguntas de seguridad
type SecurityQuestionSetupRequest struct {
	Questions []UserSecurityQuestionRequest `json:"questions" validate:"required,min=1,max=3"`
}

// UserSecurityQuestionRequest para una pregunta individual
type UserSecurityQuestionRequest struct {
	QuestionID int    `json:"questionId" validate:"required,min=1"`
	Answer     string `json:"answer" validate:"required,min=2,max=100"`
}

// SecurityQuestionUpdateRequest para actualizar una pregunta específica
type SecurityQuestionUpdateRequest struct {
	QuestionID int    `json:"questionId" validate:"required,min=1"`
	NewAnswer  string `json:"newAnswer" validate:"required,min=2,max=100"`
}

// SecurityQuestionVerifyRequest para verificar respuesta durante reset
type SecurityQuestionVerifyRequest struct {
	Token      string `json:"token" validate:"required,len=64"`
	QuestionID int    `json:"questionId" validate:"required,min=1"`
	Answer     string `json:"answer" validate:"required,min=1,max=100"`
}

// ===== ESTRUCTURAS PARA RESPONSES =====

// SecurityQuestionResponse para respuestas de API
type SecurityQuestionResponse struct {
	ID           int    `json:"id"`
	QuestionText string `json:"questionText"`
	Category     string `json:"category"`
}

// UserSecurityQuestionResponse para mostrar preguntas del usuario
type UserSecurityQuestionResponse struct {
	ID         int                       `json:"id"`
	QuestionID int                       `json:"questionId"`
	Question   *SecurityQuestionResponse `json:"question"`
	IsActive   bool                      `json:"isActive"`
	CreatedAt  time.Time                 `json:"createdAt"`
	UpdatedAt  time.Time                 `json:"updatedAt"`
}

// UserSecurityStatusResponse para mostrar estado de seguridad del usuario
type UserSecurityStatusResponse struct {
	HasSecurityQuestions bool                           `json:"hasSecurityQuestions"`
	QuestionsCount       int                            `json:"questionsCount"`
	MaxQuestions         int                            `json:"maxQuestions"`
	Questions            []UserSecurityQuestionResponse `json:"questions"`
	AvailableQuestions   []SecurityQuestionResponse     `json:"availableQuestions,omitempty"`
}

// SecurityQuestionForResetResponse para mostrar pregunta durante reset
type SecurityQuestionForResetResponse struct {
	QuestionID   int    `json:"questionId"`
	QuestionText string `json:"questionText"`
	Attempts     int    `json:"attempts"`
	MaxAttempts  int    `json:"maxAttempts"`
}

// ===== MÉTODOS DE LA TABLA =====

// TableName especifica el nombre de la tabla
func (SecurityQuestion) TableName() string {
	return "security_questions"
}

// TableName especifica el nombre de la tabla
func (UserSecurityQuestion) TableName() string {
	return "user_security_questions"
}

// ===== MÉTODOS DE VALIDACIÓN Y UTILIDAD =====

// IsValid verifica si la pregunta de seguridad está activa
func (sq *SecurityQuestion) IsValid() bool {
	return sq.IsActive
}

// ToResponse convierte a estructura de respuesta
func (sq *SecurityQuestion) ToResponse() *SecurityQuestionResponse {
	return &SecurityQuestionResponse{
		ID:           sq.ID,
		QuestionText: sq.QuestionText,
		Category:     sq.Category,
	}
}

// IsValid verifica si la respuesta de seguridad del usuario está activa
func (usq *UserSecurityQuestion) IsValid() bool {
	return usq.IsActive
}

// ToResponse convierte a estructura de respuesta
func (usq *UserSecurityQuestion) ToResponse() *UserSecurityQuestionResponse {
	response := &UserSecurityQuestionResponse{
		ID:         usq.ID,
		QuestionID: usq.SecurityQuestionID,
		IsActive:   usq.IsActive,
		CreatedAt:  usq.CreatedAt,
		UpdatedAt:  usq.UpdatedAt,
	}

	if usq.SecurityQuestion != nil {
		response.Question = usq.SecurityQuestion.ToResponse()
	}

	return response
}

// MarkAsInactive marca la pregunta como inactiva
func (usq *UserSecurityQuestion) MarkAsInactive() {
	usq.IsActive = false
	usq.UpdatedAt = time.Now()
}

// ===== CONSTANTES PARA VALIDACIÓN =====

const (
	// MaxSecurityQuestionsPerUser cantidad máxima de preguntas por usuario
	MaxSecurityQuestionsPerUser = 3

	// MinSecurityQuestionAnswer longitud mínima de respuesta
	MinSecurityQuestionAnswer = 2

	// MaxSecurityQuestionAnswer longitud máxima de respuesta
	MaxSecurityQuestionAnswer = 100

	// MaxSecurityQuestionAttempts intentos máximos para verificación
	MaxSecurityQuestionAttempts = 3
)

// ===== FUNCIONES DE VALIDACIÓN =====

// ValidateSecurityQuestionSetup valida la configuración de preguntas
func ValidateSecurityQuestionSetup(req *SecurityQuestionSetupRequest) []string {
	var errors []string

	if len(req.Questions) == 0 {
		errors = append(errors, "Debe configurar al menos 1 pregunta de seguridad")
	}

	if len(req.Questions) > MaxSecurityQuestionsPerUser {
		errors = append(errors, "No puede configurar más de 3 preguntas de seguridad")
	}

	// Verificar que no haya preguntas duplicadas
	questionIDs := make(map[int]bool)
	for i, q := range req.Questions {
		if questionIDs[q.QuestionID] {
			errors = append(errors, "No puede seleccionar la misma pregunta múltiples veces")
			break
		}
		questionIDs[q.QuestionID] = true

		if len(q.Answer) < MinSecurityQuestionAnswer {
			errors = append(errors, fmt.Sprintf("La respuesta de la pregunta %d es muy corta (mínimo %d caracteres)", i+1, MinSecurityQuestionAnswer))
		}

		if len(q.Answer) > MaxSecurityQuestionAnswer {
			errors = append(errors, fmt.Sprintf("La respuesta de la pregunta %d es muy larga (máximo %d caracteres)", i+1, MaxSecurityQuestionAnswer))
		}

		if len(strings.TrimSpace(q.Answer)) == 0 {
			errors = append(errors, fmt.Sprintf("La respuesta de la pregunta %d no puede estar vacía", i+1))
		}
	}

	return errors
}

// ValidateSecurityAnswer valida una respuesta individual
func ValidateSecurityAnswer(answer string) []string {
	var errors []string

	answer = strings.TrimSpace(answer)

	if len(answer) == 0 {
		errors = append(errors, "La respuesta no puede estar vacía")
	}

	if len(answer) < MinSecurityQuestionAnswer {
		errors = append(errors, fmt.Sprintf("La respuesta es muy corta (mínimo %d caracteres)", MinSecurityQuestionAnswer))
	}

	if len(answer) > MaxSecurityQuestionAnswer {
		errors = append(errors, fmt.Sprintf("La respuesta es muy larga (máximo %d caracteres)", MaxSecurityQuestionAnswer))
	}

	// Validar que no sea solo caracteres repetidos
	if len(answer) > 0 && isRepeatedCharacters(answer) {
		errors = append(errors, "La respuesta no puede ser solo caracteres repetidos")
	}

	return errors
}

// isRepeatedCharacters verifica si la cadena es solo caracteres repetidos
func isRepeatedCharacters(s string) bool {
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

// ===== FUNCIONES DE UTILIDAD =====

// NormalizeSecurityAnswer normaliza una respuesta de seguridad
func NormalizeSecurityAnswer(answer string) string {
	// Convertir a minúsculas, eliminar espacios extra
	normalized := strings.ToLower(strings.TrimSpace(answer))

	// Reemplazar múltiples espacios por uno solo
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	// Remover acentos básicos (opcional - implementar si es necesario)
	// normalized = removeAccents(normalized)

	return normalized
}

// GetSecurityQuestionCategories retorna las categorías disponibles
func GetSecurityQuestionCategories() []string {
	return []string{
		"personal",
		"education",
		"professional",
		"preferences",
		"general",
	}
}

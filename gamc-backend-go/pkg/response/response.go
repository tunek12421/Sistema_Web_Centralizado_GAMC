// pkg/response/response.go
package response

import (
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse estructura estándar de respuesta
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// Success respuesta exitosa
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(200, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: GetTimestamp(),
	})
}

// Error respuesta de error
func Error(c *gin.Context, statusCode int, message, errorDetail string) {
	response := APIResponse{
		Success:   false,
		Message:   message,
		Timestamp: GetTimestamp(),
	}
	
	if errorDetail != "" {
		response.Error = errorDetail
	}
	
	c.JSON(statusCode, response)
}

// Created respuesta de recurso creado
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(201, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: GetTimestamp(),
	})
}

// NoContent respuesta sin contenido
func NoContent(c *gin.Context) {
	c.JSON(204, gin.H{})
}

// GetTimestamp obtiene timestamp en formato ISO
func GetTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

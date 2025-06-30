package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	// Servidor
	Port        string
	Environment string
	APIPrefix   string

	// Base de datos
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret           string
	JWTRefreshSecret    string
	JWTExpiresIn        time.Duration
	JWTRefreshExpiresIn time.Duration
	JWTIssuer           string
	JWTAudience         string

	// CORS
	CORSOrigin string

	// Rate Limiting
	RateLimitWindowMs    time.Duration
	RateLimitMaxRequests int

	// Timezone
	Timezone string

	// MinIO Configuration
	MinIOEndpoint        string
	MinIOAccessKey       string
	MinIOSecretKey       string
	MinIOUseSSL          bool
	MinIORegion          string
	MinIOMaxFileSize     int64  // bytes
	MinIOAllowedTypes    string // comma separated
	MinIOPresignedExpiry time.Duration

	// WebSocket Configuration
	WebSocketPort            string
	WebSocketPingInterval    time.Duration
	WebSocketPongTimeout     time.Duration
	WebSocketMaxMessageSize  int64
	WebSocketReadBufferSize  int
	WebSocketWriteBufferSize int

	// Email Configuration (para notificaciones)
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	SMTPUseTLS   bool
}

// AppContext contiene las dependencias de la aplicación
type AppContext struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Config *Config
}

// Load carga la configuración desde variables de entorno
func Load() *Config {
	return &Config{
		// Servidor
		Port:        getEnv("PORT", "3000"),
		Environment: getEnv("NODE_ENV", "development"),
		APIPrefix:   getEnv("API_PREFIX", "/api/v1"),

		// Base de datos
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://gamc_user:gamc_password_2024@localhost:5432/gamc_system"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://:gamc_redis_password_2024@localhost:6379/0"),

		// JWT
		JWTSecret:           getEnv("JWT_SECRET", "gamc_jwt_secret_super_secure_2024_key_never_share"),
		JWTRefreshSecret:    getEnv("JWT_REFRESH_SECRET", "gamc_jwt_refresh_secret_super_secure_2024_key"),
		JWTExpiresIn:        parseDuration(getEnv("JWT_EXPIRES_IN", "15m")),
		JWTRefreshExpiresIn: parseDuration(getEnv("JWT_REFRESH_EXPIRES_IN", "7d")),
		JWTIssuer:           getEnv("JWT_ISSUER", "gamc-auth"),
		JWTAudience:         getEnv("JWT_AUDIENCE", "gamc-system"),

		// CORS
		CORSOrigin: getEnv("CORS_ORIGIN", "http://localhost:5173"),

		// Rate Limiting
		RateLimitWindowMs:    parseDuration(getEnv("RATE_LIMIT_WINDOW_MS", "15m")),
		RateLimitMaxRequests: parseInt(getEnv("RATE_LIMIT_MAX_REQUESTS", "100")),

		// Timezone
		Timezone: getEnv("TZ", "America/La_Paz"),

		// MinIO
		MinIOEndpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:       getEnv("MINIO_ACCESS_KEY", "gamc_backend"),
		MinIOSecretKey:       getEnv("MINIO_SECRET_KEY", "gamc_backend_secret_2024"),
		MinIOUseSSL:          getEnvBool("MINIO_USE_SSL", false),
		MinIORegion:          getEnv("MINIO_REGION", "us-east-1"),
		MinIOMaxFileSize:     parseBytes(getEnv("MINIO_MAX_FILE_SIZE", "100MB")),
		MinIOAllowedTypes:    getEnv("MINIO_ALLOWED_TYPES", ".pdf,.doc,.docx,.xls,.xlsx,.txt,.zip,.rar,.jpg,.jpeg,.png,.gif"),
		MinIOPresignedExpiry: parseDuration(getEnv("MINIO_PRESIGNED_EXPIRY", "24h")),

		// WebSocket
		WebSocketPort:            getEnv("WEBSOCKET_PORT", "3001"),
		WebSocketPingInterval:    parseDuration(getEnv("WEBSOCKET_PING_INTERVAL", "30s")),
		WebSocketPongTimeout:     parseDuration(getEnv("WEBSOCKET_PONG_TIMEOUT", "60s")),
		WebSocketMaxMessageSize:  parseBytes(getEnv("WEBSOCKET_MAX_MESSAGE_SIZE", "1MB")),
		WebSocketReadBufferSize:  parseInt(getEnv("WEBSOCKET_READ_BUFFER_SIZE", "1024")),
		WebSocketWriteBufferSize: parseInt(getEnv("WEBSOCKET_WRITE_BUFFER_SIZE", "1024")),

		// Email
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     parseInt(getEnv("SMTP_PORT", "587")),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@gamc.gov.bo"),
		SMTPUseTLS:   getEnvBool("SMTP_USE_TLS", true),
	}
}

// MinIOBuckets define los buckets de MinIO
type MinIOBuckets struct {
	Attachments string
	Documents   string
	Images      string
	Backups     string
	Temp        string
	Reports     string
}

// GetMinIOBuckets retorna la configuración de buckets
func GetMinIOBuckets() MinIOBuckets {
	return MinIOBuckets{
		Attachments: "gamc-attachments",
		Documents:   "gamc-documents",
		Images:      "gamc-images",
		Backups:     "gamc-backups",
		Temp:        "gamc-temp",
		Reports:     "gamc-reports",
	}
}

// getEnv obtiene una variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool obtiene una variable de entorno booleana
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// parseDuration parsea una duración desde string
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		// Fallback para formatos como "7d"
		switch s {
		case "7d":
			return 7 * 24 * time.Hour
		case "30d":
			return 30 * 24 * time.Hour
		case "15m":
			return 15 * time.Minute
		case "1h":
			return time.Hour
		case "24h":
			return 24 * time.Hour
		default:
			return 15 * time.Minute // default
		}
	}
	return d
}

// parseInt parsea un entero desde string
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	// Fallback values
	switch s {
	case "100":
		return 100
	case "200":
		return 200
	case "50":
		return 50
	case "587":
		return 587
	case "1024":
		return 1024
	default:
		return 100
	}
}

// parseBytes parsea tamaño en bytes desde string
func parseBytes(s string) int64 {
	s = strings.ToUpper(s)
	multiplier := int64(1)

	if strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}

	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val * multiplier
	}

	// Default 100MB
	return 100 * 1024 * 1024
}

// GetAllowedFileExtensions retorna las extensiones permitidas como slice
func (c *Config) GetAllowedFileExtensions() []string {
	return strings.Split(c.MinIOAllowedTypes, ",")
}

// IsFileTypeAllowed verifica si un tipo de archivo está permitido
func (c *Config) IsFileTypeAllowed(extension string) bool {
	allowed := c.GetAllowedFileExtensions()
	for _, ext := range allowed {
		if strings.EqualFold(ext, extension) {
			return true
		}
	}
	return false
}

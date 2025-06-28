package config

import (
	"os"
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
	}
}

// getEnv obtiene una variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
		default:
			return 15 * time.Minute // default
		}
	}
	return d
}

// parseInt parsea un entero desde string
func parseInt(s string) int {
	switch s {
	case "100":
		return 100
	case "200":
		return 200
	case "50":
		return 50
	default:
		return 100
	}
}

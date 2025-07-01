package database

import (
	"context"
	"fmt"
	"time"

	"gamc-backend-go/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Initialize inicializa la conexión a la base de datos SIN MIGRACIONES
func Initialize(databaseURL string) (*gorm.DB, error) {
	logger.Info("🚀 Conectando a PostgreSQL...")

	// Configurar logger de GORM - MODO SILENCIOSO para evitar spam
	config := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		NowFunc: func() time.Time {
			// Usar timezone de Bolivia
			loc, _ := time.LoadLocation("America/La_Paz")
			return time.Now().In(loc)
		},
		// Deshabilitar TODAS las funciones de migración
		DisableForeignKeyConstraintWhenMigrating: true,
		// No crear automáticamente relaciones
		CreateBatchSize: 1000,
	}

	// Conectar a PostgreSQL
	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configurar pool de conexiones
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configuración del pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verificar conexión con timeout corto
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("✅ Conexión a PostgreSQL establecida")
	logger.Info("📊 Usando base de datos existente SIN migraciones automáticas")

	// Solo verificar conectividad básica
	if err := basicConnectivityTest(db); err != nil {
		logger.Warn("⚠️ Problema de conectividad básica: %v", err)
	} else {
		logger.Info("✅ Test de conectividad básica exitoso")
	}

	return db, nil
}

// basicConnectivityTest hace una prueba básica de conectividad
func basicConnectivityTest(db *gorm.DB) error {
	// Test simple sin tocar estructura de tablas
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("basic query failed: %w", err)
	}

	// Verificar que tenemos acceso a la base de datos correcta
	var dbName string
	if err := db.Raw("SELECT current_database()").Scan(&dbName).Error; err != nil {
		return fmt.Errorf("database name query failed: %w", err)
	}

	logger.Info("📍 Conectado a base de datos: %s", dbName)
	return nil
}

// DatabaseStats representa estadísticas básicas de la base de datos
type DatabaseStats struct {
	DatabaseName   string `json:"databaseName"`
	TablesCount    int64  `json:"tablesCount"`
	Connected      bool   `json:"connected"`
	ConnectionTime string `json:"connectionTime"`
}

// GetBasicStats obtiene estadísticas básicas sin tocar tablas específicas
func GetBasicStats(db *gorm.DB) (*DatabaseStats, error) {
	stats := &DatabaseStats{
		Connected:      true,
		ConnectionTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Obtener nombre de la base de datos
	if err := db.Raw("SELECT current_database()").Scan(&stats.DatabaseName).Error; err != nil {
		logger.Warn("No se pudo obtener nombre de DB: %v", err)
		stats.DatabaseName = "unknown"
	}

	// Contar tablas disponibles
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&stats.TablesCount).Error; err != nil {
		logger.Warn("No se pudo contar tablas: %v", err)
		stats.TablesCount = 0
	}

	return stats, nil
}

// FUNCIONES DE MIGRACIÓN COMPLETAMENTE DESHABILITADAS

// AutoMigrate - DESHABILITADO COMPLETAMENTE
func AutoMigrate(db *gorm.DB) error {
	logger.Info("🚫 AutoMigrate DESHABILITADO - Usando base de datos existente")
	return nil
}

// ManualMigrate - DESHABILITADO COMPLETAMENTE
func ManualMigrate(db *gorm.DB) error {
	logger.Info("🚫 ManualMigrate DESHABILITADO - Usando base de datos existente")
	return nil
}

// TestDatabaseConnection hace una prueba completa de la conexión
func TestDatabaseConnection(db *gorm.DB) error {
	// Test 1: Consulta básica
	var one int
	if err := db.Raw("SELECT 1").Scan(&one).Error; err != nil {
		return fmt.Errorf("test query failed: %w", err)
	}

	// Test 2: Verificar timezone
	var timezone string
	if err := db.Raw("SHOW timezone").Scan(&timezone).Error; err != nil {
		logger.Warn("No se pudo obtener timezone: %v", err)
	} else {
		logger.Info("🌍 Timezone de la base de datos: %s", timezone)
	}

	// Test 3: Verificar versión de PostgreSQL
	var version string
	if err := db.Raw("SELECT version()").Scan(&version).Error; err != nil {
		logger.Warn("No se pudo obtener versión de PostgreSQL: %v", err)
	} else {
		logger.Info("🐘 PostgreSQL version: %s", version[:50]) // Solo primeros 50 caracteres
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gamc-backend-go/internal/api/routes"
	"gamc-backend-go/internal/config"
	"gamc-backend-go/internal/database"
	"gamc-backend-go/internal/redis"
	"gamc-backend-go/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Inicializar logger
	logger.Init()
	logger.Info("🚀 Iniciando GAMC Backend Auth Service...")

	// Cargar configuración
	cfg := config.Load()

	// Configurar modo Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar base de datos
	db, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("❌ Error conectando a PostgreSQL: %v", err)
	}
	logger.Info("✅ Conexión a PostgreSQL establecida")

	// Inicializar Redis
	redisClient, err := redis.Initialize(cfg.RedisURL)
	if err != nil {
		logger.Fatal("❌ Error conectando a Redis: %v", err)
	}
	logger.Info("✅ Conexión a Redis establecida")

	// Crear contexto de aplicación
	appCtx := &config.AppContext{
		DB:     db,
		Redis:  redisClient,
		Config: cfg,
	}

	// Configurar rutas
	router := routes.SetupRoutes(appCtx)

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
		IdleTimeout:  time.Second * 60,
	}

	// Canal para manejar shutdown graceful
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor en goroutine
	go func() {
		logger.Info("🎉 ========================================")
		logger.Info("🎉 GAMC Sistema de Autenticación")
		logger.Info("🎉 ========================================")
		logger.Info("🚀 Servidor ejecutándose en puerto %s", cfg.Port)
		logger.Info("🌍 Entorno: %s", cfg.Environment)
		logger.Info("📡 Health check: http://localhost:%s/health", cfg.Port)
		logger.Info("🔐 Auth API: http://localhost:%s/api/v1/auth", cfg.Port)
		logger.Info("🎉 ========================================")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("❌ Error al iniciar servidor: %v", err)
		}
	}()

	// Esperar señal de shutdown
	<-quit
	logger.Info("📴 Recibida señal de shutdown. Cerrando servidor...")

	// Graceful shutdown con timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Cerrar servidor HTTP
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("❌ Error en shutdown del servidor: %v", err)
	}

	// Cerrar conexiones de base de datos
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.Close()
		logger.Info("✅ Conexión a PostgreSQL cerrada")
	}

	// Cerrar conexión de Redis
	if err := redisClient.Close(); err != nil {
		logger.Error("❌ Error cerrando Redis: %v", err)
	} else {
		logger.Info("✅ Conexión a Redis cerrada")
	}

	logger.Info("✅ Servidor cerrado exitosamente")
}

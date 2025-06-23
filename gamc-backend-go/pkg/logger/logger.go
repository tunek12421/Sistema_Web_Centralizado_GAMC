// pkg/logger/logger.go
package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init inicializa el logger
func Init() {
	log = logrus.New()

	// Configurar formato
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Configurar nivel según entorno
	env := os.Getenv("NODE_ENV")
	if env == "development" {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.SetOutput(os.Stdout)
}

// Info log de información
func Info(format string, args ...interface{}) {
	if log != nil {
		log.Infof(format, args...)
	}
}

// Error log de error
func Error(format string, args ...interface{}) {
	if log != nil {
		log.Errorf(format, args...)
	}
}

// Debug log de debug
func Debug(format string, args ...interface{}) {
	if log != nil {
		log.Debugf(format, args...)
	}
}

// Warn log de advertencia
func Warn(format string, args ...interface{}) {
	if log != nil {
		log.Warnf(format, args...)
	}
}

// Fatal log fatal (termina la aplicación)
func Fatal(format string, args ...interface{}) {
	if log != nil {
		log.Fatalf(format, args...)
	}
}

// GetLevel obtiene el nivel actual del logger
func GetLevel() string {
	if log != nil {
		return log.GetLevel().String()
	}
	return "info"
}

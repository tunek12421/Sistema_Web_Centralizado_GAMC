# =====================================
# Makefile
.PHONY: build run test clean docker-build docker-run migrate

# Variables
BINARY_NAME=gamc-backend-go
MAIN_PATH=cmd/server/main.go

# Build de la aplicación
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Ejecutar en desarrollo
run:
	go run $(MAIN_PATH)

# Ejecutar con hot reload
dev:
	air -c .air.toml

# Ejecutar tests
test:
	go test -v ./...

# Test con coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Limpiar archivos generados
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Build Docker
docker-build:
	docker build -t $(BINARY_NAME) -f docker/Dockerfile .

# Ejecutar con Docker
docker-run:
	docker-compose up -d

# Parar Docker
docker-stop:
	docker-compose down

# Logs de Docker
docker-logs:
	docker-compose logs -f gamc-backend-go

# Ejecutar migraciones (cuando estén implementadas)
migrate:
	@echo "Migraciones ejecutadas automáticamente al iniciar"

# Generar documentación
docs:
	@echo "Generando documentación..."
	@echo "Endpoints disponibles en README.md"

# Instalar dependencias de desarrollo
install-dev:
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint del código
lint:
	golangci-lint run

# Formatear código
fmt:
	go fmt ./...

# Verificar módulos
mod-tidy:
	go mod tidy
	go mod verify

# Build para producción
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BINARY_NAME) $(MAIN_PATH)

# Ayuda
help:
	@echo "Comandos disponibles:"
	@echo "  build         - Compilar la aplicación"
	@echo "  run           - Ejecutar en desarrollo"
	@echo "  dev           - Ejecutar con hot reload (requiere air)"
	@echo "  test          - Ejecutar tests"
	@echo "  test-coverage - Tests con reporte de coverage"
	@echo "  clean         - Limpiar archivos generados"
	@echo "  docker-build  - Build de imagen Docker"
	@echo "  docker-run    - Ejecutar con Docker Compose"
	@echo "  docker-stop   - Parar contenedores"
	@echo "  docker-logs   - Ver logs del contenedor"
	@echo "  lint          - Lint del código"
	@echo "  fmt           - Formatear código"
	@echo "  mod-tidy      - Limpiar módulos de Go"
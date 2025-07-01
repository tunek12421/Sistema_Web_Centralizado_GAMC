#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Scripts de Gestión del Sistema Unificado
# ========================================

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Función para mostrar banner
show_banner() {
    echo -e "${BLUE}"
    echo "=========================================="
    echo "  GAMC Sistema Web Centralizado"
    echo "  Gestión de Servicios Unificados"
    echo "=========================================="
    echo -e "${NC}"
}

# Función para mostrar estado de servicios
show_status() {
    echo -e "${YELLOW}Verificando estado de servicios...${NC}"
    docker-compose ps
    echo ""
    echo -e "${YELLOW}Servicios por perfil:${NC}"
    echo "• Básicos: postgres, redis, minio, gamc-auth-backend, gamc-auth-frontend"
    echo "• Admin: pgadmin, redis-commander"
    echo "• Setup: minio-client"
    echo "• Gateway: minio-gateway"
    echo "• Monitor: healthcheck-monitor"
}

# Función para iniciar servicios básicos
start_basic() {
    echo -e "${GREEN}Iniciando servicios básicos...${NC}"
    docker-compose up -d postgres redis minio gamc-auth-backend gamc-auth-frontend
    echo -e "${GREEN}Servicios básicos iniciados${NC}"
    show_urls
}

# Función para iniciar todos los servicios
start_all() {
    echo -e "${GREEN}Iniciando todos los servicios...${NC}"
    docker-compose --profile admin --profile setup --profile gateway --profile monitor up -d
    echo -e "${GREEN}Todos los servicios iniciados${NC}"
    show_urls
}

# Función para iniciar solo con herramientas de admin
start_with_admin() {
    echo -e "${GREEN}Iniciando servicios con herramientas de administración...${NC}"
    docker-compose --profile admin up -d
    echo -e "${GREEN}Servicios con admin iniciados${NC}"
    show_urls
}

# Función para parar servicios
stop_services() {
    echo -e "${YELLOW}Deteniendo servicios...${NC}"
    docker-compose --profile admin --profile setup --profile gateway --profile monitor down
    echo -e "${GREEN}Servicios detenidos${NC}"
}

# Función para reiniciar servicios
restart_services() {
    echo -e "${YELLOW}Reiniciando servicios...${NC}"
    stop_services
    sleep 3
    start_basic
}

# Función para mostrar logs
show_logs() {
    if [ -z "$1" ]; then
        echo -e "${YELLOW}Mostrando logs de todos los servicios...${NC}"
        docker-compose logs -f --tail=50
    else
        echo -e "${YELLOW}Mostrando logs de $1...${NC}"
        docker-compose logs -f --tail=50 "$1"
    fi
}

# Función para limpiar el sistema
cleanup() {
    echo -e "${RED}¿Estás seguro de que quieres limpiar todos los datos? (y/N)${NC}"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY]|[sS][iI])$ ]]; then
        echo -e "${YELLOW}Limpiando sistema...${NC}"
        docker-compose --profile admin --profile setup --profile gateway --profile monitor down -v
        docker system prune -f
        docker volume prune -f
        echo -e "${GREEN}Sistema limpiado${NC}"
    else
        echo -e "${GREEN}Operación cancelada${NC}"
    fi
}

# Función para hacer backup
backup_data() {
    BACKUP_DIR="./backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    echo -e "${YELLOW}Realizando backup...${NC}"
    
    # Backup PostgreSQL
    docker-compose exec postgres pg_dump -U gamc_user gamc_system > "$BACKUP_DIR/postgres_backup.sql"
    
    # Backup Redis
    docker-compose exec redis redis-cli --rdb - > "$BACKUP_DIR/redis_backup.rdb"
    
    # Backup MinIO (configuración)
    docker cp gamc_minio:/data "$BACKUP_DIR/minio_data"
    
    echo -e "${GREEN}Backup completado en: $BACKUP_DIR${NC}"
}

# Función para actualizar servicios
update_services() {
    echo -e "${YELLOW}Actualizando imágenes...${NC}"
    docker-compose pull
    echo -e "${YELLOW}Reconstruyendo servicios personalizados...${NC}"
    docker-compose build --no-cache gamc-auth-backend gamc-auth-frontend
    echo -e "${GREEN}Servicios actualizados${NC}"
}

# Función para mostrar URLs de acceso
show_urls() {
    echo ""
    echo -e "${BLUE}=========================================="
    echo "  URLs de Acceso"
    echo "==========================================${NC}"
    echo -e "${GREEN}• Frontend Auth:${NC} http://localhost:5173"
    echo -e "${GREEN}• Backend Auth API:${NC} http://localhost:3000/api/v1"
    echo -e "${GREEN}• MinIO Console:${NC} http://localhost:9001"
    echo -e "${GREEN}• MinIO API:${NC} http://localhost:9000"
    echo -e "${GREEN}• PgAdmin:${NC} http://localhost:8080"
    echo -e "${GREEN}• Redis Commander:${NC} http://localhost:8081"
    echo -e "${GREEN}• MinIO Gateway:${NC} http://localhost:8090"
    echo ""
    echo -e "${YELLOW}Credenciales por defecto:${NC}"
    echo "• PgAdmin: admin@gamc.gov.bo / admin123"
    echo "• Redis Commander: admin / admin123"
    echo "• MinIO: gamc_admin / gamc_minio_password_2024"
}

# Función para verificar configuración de red
check_network() {
    echo -e "${YELLOW}Verificando configuración de red...${NC}"
    if docker network ls | grep -q gamc_network; then
        echo -e "${GREEN}✓ Red gamc_network existe${NC}"
        docker network inspect gamc_network
    else
        echo -e "${RED}✗ Red gamc_network no existe${NC}"
        echo -e "${YELLOW}Creando red...${NC}"
        docker network create gamc_network
    fi
}

# Función para verificar volúmenes
check_volumes() {
    echo -e "${YELLOW}Verificando volúmenes...${NC}"
    for volume in gamc_postgres_data gamc_redis_data gamc_minio_data gamc_pgadmin_data; do
        if docker volume ls | grep -q $volume; then
            echo -e "${GREEN}✓ Volumen $volume existe${NC}"
        else
            echo -e "${RED}✗ Volumen $volume no existe${NC}"
        fi
    done
}

# Función para configuración inicial
initial_setup() {
    echo -e "${YELLOW}Configuración inicial del sistema...${NC}"
    
    # Crear directorios necesarios
    mkdir -p database/backups
    mkdir -p redis/logs
    mkdir -p minio/config
    mkdir -p minio/policies
    
    # Verificar archivos de configuración
    if [ ! -f "backend-auth/.env" ]; then
        echo -e "${RED}✗ Archivo backend-auth/.env no encontrado${NC}"
        echo "Creando archivo de ejemplo..."
        cat > backend-auth/.env << 'EOF'
NODE_ENV=development
PORT=3000
API_PREFIX=/api/v1
DATABASE_URL=postgresql://gamc_user:gamc_password_2024@postgres:5432/gamc_system
REDIS_URL=redis://:gamc_redis_password_2024@redis:6379/0
JWT_SECRET=gamc_jwt_secret_super_secure_2024_key_never_share
JWT_REFRESH_SECRET=gamc_jwt_refresh_secret_super_secure_2024_key
JWT_EXPIRES_IN=15m
JWT_REFRESH_EXPIRES_IN=7d
CORS_ORIGIN=http://localhost:5173
TZ=America/La_Paz
EOF
    else
        echo -e "${GREEN}✓ Archivo backend-auth/.env existe${NC}"
    fi
    
    if [ ! -f "frontend-auth/.env" ]; then
        echo -e "${RED}✗ Archivo frontend-auth/.env no encontrado${NC}"
        echo "Creando archivo de ejemplo..."
        cat > frontend-auth/.env << 'EOF'
VITE_API_URL=http://localhost:3000/api/v1
EOF
    else
        echo -e "${GREEN}✓ Archivo frontend-auth/.env existe${NC}"
    fi
    
    check_network
    check_volumes
    
    echo -e "${GREEN}Configuración inicial completada${NC}"
}

# Función para mostrar ayuda
show_help() {
    echo -e "${BLUE}Comandos disponibles:${NC}"
    echo ""
    echo -e "${GREEN}start${NC}          - Iniciar servicios básicos"
    echo -e "${GREEN}start-all${NC}      - Iniciar todos los servicios (incluyendo admin)"
    echo -e "${GREEN}start-admin${NC}    - Iniciar servicios con herramientas de admin"
    echo -e "${GREEN}stop${NC}           - Detener todos los servicios"
    echo -e "${GREEN}restart${NC}        - Reiniciar servicios"
    echo -e "${GREEN}status${NC}         - Mostrar estado de servicios"
    echo -e "${GREEN}logs [servicio]${NC} - Mostrar logs (todos o de un servicio específico)"
    echo -e "${GREEN}urls${NC}           - Mostrar URLs de acceso"
    echo -e "${GREEN}backup${NC}         - Realizar backup de datos"
    echo -e "${GREEN}update${NC}         - Actualizar servicios"
    echo -e "${GREEN}cleanup${NC}        - Limpiar todo el sistema (CUIDADO!)"
    echo -e "${GREEN}setup${NC}          - Configuración inicial"
    echo -e "${GREEN}network${NC}        - Verificar configuración de red"
    echo -e "${GREEN}volumes${NC}        - Verificar volúmenes"
    echo -e "${GREEN}help${NC}           - Mostrar esta ayuda"
    echo ""
    echo -e "${YELLOW}Ejemplos:${NC}"
    echo "  ./gamc.sh start          # Iniciar servicios básicos"
    echo "  ./gamc.sh logs backend   # Ver logs del backend"
    echo "  ./gamc.sh backup         # Hacer backup"
}

# Script principal
main() {
    show_banner
    
    case "${1:-help}" in
        "start")
            start_basic
            ;;
        "start-all")
            start_all
            ;;
        "start-admin")
            start_with_admin
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs "$2"
            ;;
        "urls")
            show_urls
            ;;
        "backup")
            backup_data
            ;;
        "update")
            update_services
            ;;
        "cleanup")
            cleanup
            ;;
        "setup")
            initial_setup
            ;;
        "network")
            check_network
            ;;
        "volumes")
            check_volumes
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Ejecutar función principal con todos los argumentos
main "$@"

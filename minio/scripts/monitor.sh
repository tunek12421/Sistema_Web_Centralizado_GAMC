#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Script de Monitoreo de MinIO
# ========================================

set -e

# Configuración
MINIO_ALIAS=${MINIO_ALIAS:-"gamc-local"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"http://localhost:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"gamc_admin"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"gamc_minio_password_2024"}
REFRESH_INTERVAL=${REFRESH_INTERVAL:-10}

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Función para limpiar pantalla
clear_screen() {
    clear
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║              GAMC MinIO - Monitor en Tiempo Real                ║${NC}"
    echo -e "${BLUE}║              Actualización cada $REFRESH_INTERVAL segundos                     ║${NC}"
    echo -e "${BLUE}║              Presiona Ctrl+C para salir                         ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Función para configurar cliente MinIO
setup_client() {
    if ! mc alias set $MINIO_ALIAS $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY >/dev/null 2>&1; then
        echo -e "${RED}❌ Error al configurar cliente MinIO${NC}"
        exit 1
    fi
}

# Función para obtener información del servidor
get_server_info() {
    echo -e "${CYAN}🖥️  INFORMACIÓN DEL SERVIDOR${NC}"
    
    local server_info=$(mc admin info $MINIO_ALIAS 2>/dev/null || echo "Error al obtener información")
    
    if [ "$server_info" != "Error al obtener información" ]; then
        local version=$(echo "$server_info" | grep -i "MinIO version" | head -1 | awk '{print $3}' || echo "Unknown")
        local uptime=$(echo "$server_info" | grep -i "uptime" | head -1 | awk '{print $2}' || echo "Unknown")
        local online_drives=$(echo "$server_info" | grep -i "online" | wc -l || echo "0")
        
        echo -e "   Versión: ${GREEN}$version${NC}"
        echo -e "   Uptime: ${GREEN}$uptime${NC}"
        echo -e "   Drives online: ${GREEN}$online_drives${NC}"
        echo -e "   Timestamp: ${GREEN}$(date '+%Y-%m-%d %H:%M:%S')${NC}"
    else
        echo -e "   ${RED}❌ No se puede obtener información del servidor${NC}"
    fi
    echo ""
}

# Función para obtener información de almacenamiento
get_storage_info() {
    echo -e "${YELLOW}💾 ALMACENAMIENTO${NC}"
    
    # Información del sistema de archivos donde está montado MinIO
    local storage_info=$(df -h /data 2>/dev/null || df -h . 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "$storage_info" | tail -1 | while read -r filesystem size used avail percent mount; do
            echo -e "   Tamaño total: ${GREEN}$size${NC}"
            echo -e "   Usado: ${GREEN}$used${NC}"
            echo -e "   Disponible: ${GREEN}$avail${NC}"
            echo -e "   Porcentaje usado: ${GREEN}$percent${NC}"
        done
    else
        echo -e "   ${YELLOW}⚠️  Información de almacenamiento no disponible${NC}"
    fi
    echo ""
}

# Función para obtener información de buckets
get_buckets_info() {
    echo -e "${MAGENTA}🗂️  BUCKETS Y CONTENIDO${NC}"
    
    local buckets=$(mc ls $MINIO_ALIAS 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$buckets" ]; then
        echo "$buckets" | while read -r line; do
            local bucket_date=$(echo "$line" | awk '{print $1, $2}')
            local bucket_name=$(echo "$line" | awk '{print $5}')
            
            if [ -n "$bucket_name" ]; then
                # Obtener información detallada del bucket
                local file_count=$(mc ls $MINIO_ALIAS/$bucket_name --recursive 2>/dev/null | wc -l || echo "0")
                local bucket_size=$(mc du $MINIO_ALIAS/$bucket_name 2>/dev/null | tail -1 | awk '{print $1, $2}' || echo "0 B")
                
                echo -e "   📦 ${GREEN}$bucket_name${NC}"
                echo -e "      Archivos: ${GREEN}$file_count${NC} | Tamaño: ${GREEN}$bucket_size${NC}"
                echo -e "      Creado: ${GREEN}$bucket_date${NC}"
                echo ""
            fi
        done
    else
        echo -e "   ${YELLOW}⚠️  No se encontraron buckets o error de conexión${NC}"
        echo ""
    fi
}

# Función para obtener actividad reciente
get_recent_activity() {
    echo -e "${CYAN}📊 ACTIVIDAD RECIENTE${NC}"
    
    # Buscar archivos modificados recientemente en todos los buckets
    local recent_files=0
    local recent_uploads=0
    
    mc ls $MINIO_ALIAS 2>/dev/null | awk '{print $5}' | while read -r bucket; do
        if [ -n "$bucket" ]; then
            # Archivos de las últimas 24 horas
            local bucket_recent=$(mc ls $MINIO_ALIAS/$bucket --recursive 2>/dev/null | \
                awk -v cutoff="$(date -d '24 hours ago' '+%Y-%m-%d %H:%M')" \
                '$1 " " $2 > cutoff {count++} END {print count+0}' || echo "0")
            recent_files=$((recent_files + bucket_recent))
        fi
    done
    
    echo -e "   Archivos subidos (24h): ${GREEN}$recent_files${NC}"
    echo -e "   Operaciones activas: ${GREEN}En tiempo real${NC}"
    echo ""
}

# Función para obtener estadísticas de conexión
get_connection_stats() {
    echo -e "${BLUE}🔗 ESTADÍSTICAS DE CONEXIÓN${NC}"
    
    # Verificar conectividad
    if mc admin info $MINIO_ALIAS >/dev/null 2>&1; then
        echo -e "   Estado de conexión: ${GREEN}✅ Conectado${NC}"
        
        # Tiempo de respuesta
        local start_time=$(date +%s%N)
        mc ls $MINIO_ALIAS >/dev/null 2>&1
        local end_time=$(date +%s%N)
        local response_time=$(( (end_time - start_time) / 1000000 ))
        
        echo -e "   Tiempo de respuesta: ${GREEN}${response_time}ms${NC}"
        echo -e "   Endpoint: ${GREEN}$MINIO_ENDPOINT${NC}"
    else
        echo -e "   Estado de conexión: ${RED}❌ Desconectado${NC}"
        echo -e "   Error: ${RED}No se puede conectar al servidor${NC}"
    fi
    echo ""
}

# Función para obtener información de usuarios
get_users_info() {
    echo -e "${MAGENTA}👥 USUARIOS DEL SISTEMA${NC}"
    
    local users=$(mc admin user list $MINIO_ALIAS 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$users" ]; then
        echo "$users" | while read -r user status; do
            if [ -n "$user" ]; then
                local status_color="${GREEN}"
                if [ "$status" = "disabled" ]; then
                    status_color="${RED}"
                fi
                echo -e "   👤 ${GREEN}$user${NC} - Estado: ${status_color}$status${NC}"
            fi
        done
    else
        echo -e "   ${YELLOW}⚠️  No se puede obtener lista de usuarios${NC}"
    fi
    echo ""
}

# Función para obtener logs recientes
get_recent_logs() {
    echo -e "${CYAN}📝 LOGS RECIENTES${NC}"
    
    # Obtener logs de MinIO (si están disponibles)
    local log_file="/var/log/minio/minio.log"
    
    if [ -f "$log_file" ]; then
        echo -e "   Últimas 5 entradas del log:"
        tail -5 "$log_file" 2>/dev/null | while read -r line; do
            echo -e "   ${GREEN}$line${NC}"
        done
    else
        echo -e "   ${YELLOW}⚠️  Logs no disponibles en esta ubicación${NC}"
        echo -e "   Verificar logs con: docker logs gamc_minio"
    fi
    echo ""
}

# Función para obtener alertas y problemas
get_alerts() {
    echo -e "${RED}⚠️  ALERTAS Y PROBLEMAS${NC}"
    
    local alerts_count=0
    
    # Verificar espacio en disco
    local disk_usage=$(df -h /data 2>/dev/null | tail -1 | awk '{print $5}' | tr -d '%' || echo "0")
    if [ "$disk_usage" -gt 80 ]; then
        echo -e "   ${RED}🚨 Espacio en disco bajo: ${disk_usage}%${NC}"
        alerts_count=$((alerts_count + 1))
    fi
    
    # Verificar conectividad
    if ! mc admin info $MINIO_ALIAS >/dev/null 2>&1; then
        echo -e "   ${RED}🚨 Servidor MinIO no accesible${NC}"
        alerts_count=$((alerts_count + 1))
    fi
    
    # Verificar buckets críticos
    local critical_buckets=("gamc-attachments" "gamc-documents" "gamc-backups")
    for bucket in "${critical_buckets[@]}"; do
        if ! mc ls $MINIO_ALIAS/$bucket >/dev/null 2>&1; then
            echo -e "   ${RED}🚨 Bucket crítico no accesible: $bucket${NC}"
            alerts_count=$((alerts_count + 1))
        fi
    done
    
    if [ $alerts_count -eq 0 ]; then
        echo -e "   ${GREEN}✅ No hay alertas activas${NC}"
    else
        echo -e "   ${RED}Total de alertas: $alerts_count${NC}"
    fi
    echo ""
}

# Función para mostrar comandos útiles
show_useful_commands() {
    echo -e "${BLUE}🛠️  COMANDOS ÚTILES${NC}"
    echo -e "   Listar buckets: ${YELLOW}mc ls $MINIO_ALIAS${NC}"
    echo -e "   Info del servidor: ${YELLOW}mc admin info $MINIO_ALIAS${NC}"
    echo -e "   Backup bucket: ${YELLOW}./backup.sh incremental bucket-name${NC}"
    echo -e "   Ver logs: ${YELLOW}docker logs gamc_minio${NC}"
    echo -e "   Consola web: ${YELLOW}http://localhost:9001${NC}"
    echo ""
}

# Función principal de monitoreo
monitor_minio() {
    # Configurar cliente
    setup_client
    
    # Loop principal
    while true; do
        clear_screen
        get_server_info
        get_storage_info
        get_buckets_info
        get_recent_activity
        get_connection_stats
        get_users_info
        get_recent_logs
        get_alerts
        show_useful_commands
        
        echo -e "${BLUE}Próxima actualización en $REFRESH_INTERVAL segundos...${NC}"
        sleep $REFRESH_INTERVAL
    done
}

# Función para mostrar ayuda
show_help() {
    echo "Uso: $0 [opciones]"
    echo ""
    echo "Opciones:"
    echo "  -h, --help        Mostrar esta ayuda"
    echo "  -i, --interval N  Intervalo de actualización en segundos (default: 10)"
    echo ""
    echo "Variables de entorno:"
    echo "  MINIO_ENDPOINT    Endpoint de MinIO (default: http://localhost:9000)"
    echo "  MINIO_ACCESS_KEY  Access key (default: gamc_admin)"
    echo "  MINIO_SECRET_KEY  Secret key"
    echo "  REFRESH_INTERVAL  Intervalo de actualización (default: 10)"
    echo ""
    echo "Ejemplos:"
    echo "  $0                    # Monitor con configuración por defecto"
    echo "  $0 -i 5               # Actualizar cada 5 segundos"
    echo "  MINIO_ENDPOINT=http://prod:9000 $0  # Conectar a servidor de producción"
}

# Manejo de argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -i|--interval)
            REFRESH_INTERVAL="$2"
            shift 2
            ;;
        *)
            echo "Opción desconocida: $1"
            show_help
            exit 1
            ;;
    esac
done

# Validar intervalo
if ! [[ "$REFRESH_INTERVAL" =~ ^[0-9]+$ ]] || [ "$REFRESH_INTERVAL" -lt 1 ]; then
    echo "Error: El intervalo debe ser un número positivo"
    exit 1
fi

# Verificar que mc esté disponible
if ! command -v mc &> /dev/null; then
    echo -e "${RED}Error: MinIO client (mc) no está instalado${NC}"
    echo -e "${YELLOW}Instalando mc...${NC}"
    
    # Intentar instalar mc
    if curl -sL https://dl.min.io/client/mc/release/linux-amd64/mc -o /tmp/mc; then
        chmod +x /tmp/mc
        sudo mv /tmp/mc /usr/local/bin/ 2>/dev/null || mv /tmp/mc ./mc
        echo -e "${GREEN}mc instalado exitosamente${NC}"
    else
        echo -e "${RED}Error al instalar mc. Instálalo manualmente.${NC}"
        exit 1
    fi
fi

# Manejo de señales para salida limpia
trap 'echo -e "\n${GREEN}Monitoreo detenido${NC}"; exit 0' INT TERM

# Iniciar monitoreo
echo -e "${GREEN}Iniciando monitor de MinIO...${NC}"
monitor_minio
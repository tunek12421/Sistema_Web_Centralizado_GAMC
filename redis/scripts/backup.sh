#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Script de Backup de Redis
# ========================================

set -e

# Configuración
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-gamc_redis_password_2024}
CONTAINER_NAME=${CONTAINER_NAME:-gamc_redis}
BACKUP_DIR=${BACKUP_DIR:-./backups}
RETENTION_DAYS=${RETENTION_DAYS:-7}

# Crear directorio de backups si no existe
mkdir -p "$BACKUP_DIR"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Timestamp para archivos
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GAMC Redis - Backup de Datos${NC}"
echo -e "${BLUE}========================================${NC}"

# Función para hacer backup RDB
backup_rdb() {
    echo -e "${YELLOW}1. Creando backup RDB...${NC}"
    
    # Forzar un BGSAVE
    if docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD bgsave > /dev/null 2>&1; then
        echo -e "${GREEN}✅ BGSAVE iniciado${NC}"
        
        # Esperar a que termine el BGSAVE
        while true; do
            local bgsave_status=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD lastsave 2>/dev/null)
            sleep 2
            local current_status=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD lastsave 2>/dev/null)
            
            if [ "$current_status" != "$bgsave_status" ]; then
                break
            fi
            echo -e "${YELLOW}   Esperando que termine BGSAVE...${NC}"
        done
        
        # Copiar el archivo RDB
        if docker cp $CONTAINER_NAME:/data/gamc_dump.rdb "$BACKUP_DIR/gamc_dump_${TIMESTAMP}.rdb"; then
            echo -e "${GREEN}✅ Backup RDB creado: gamc_dump_${TIMESTAMP}.rdb${NC}"
        else
            echo -e "${RED}❌ Error al copiar backup RDB${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ Error al ejecutar BGSAVE${NC}"
        return 1
    fi
}

# Función para hacer backup AOF
backup_aof() {
    echo -e "${YELLOW}2. Creando backup AOF...${NC}"
    
    # Verificar si AOF está habilitado
    local aof_enabled=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD config get appendonly 2>/dev/null | tail -1)
    
    if [ "$aof_enabled" = "yes" ]; then
        # Forzar reescritura del AOF
        if docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD bgrewriteaof > /dev/null 2>&1; then
            echo -e "${GREEN}✅ BGREWRITEAOF iniciado${NC}"
            
            # Esperar a que termine
            sleep 5
            
            # Copiar el archivo AOF
            if docker cp $CONTAINER_NAME:/data/gamc_appendonly.aof "$BACKUP_DIR/gamc_appendonly_${TIMESTAMP}.aof" 2>/dev/null; then
                echo -e "${GREEN}✅ Backup AOF creado: gamc_appendonly_${TIMESTAMP}.aof${NC}"
            else
                echo -e "${YELLOW}⚠️  No se pudo copiar AOF (puede no existir)${NC}"
            fi
        else
            echo -e "${RED}❌ Error al ejecutar BGREWRITEAOF${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  AOF no está habilitado${NC}"
    fi
}

# Función para hacer backup por bases de datos
backup_databases() {
    echo -e "${YELLOW}3. Creando backups por base de datos...${NC}"
    
    local db_names=("sesiones" "cache" "notificaciones" "metricas" "query_cache" "jwt_blacklist")
    
    for i in {0..5}; do
        local db_name=${db_names[$i]}
        local backup_file="$BACKUP_DIR/gamc_db${i}_${db_name}_${TIMESTAMP}.txt"
        
        # Obtener todas las llaves de la DB
        local keys=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i keys "*" 2>/dev/null)
        
        if [ -n "$keys" ]; then
            echo -e "   DB$i ($db_name): Exportando llaves..."
            
            # Crear archivo de backup
            echo "# GAMC Redis Backup - DB$i ($db_name) - $TIMESTAMP" > "$backup_file"
            echo "# Generado: $(date)" >> "$backup_file"
            echo "" >> "$backup_file"
            
            # Exportar cada llave
            echo "$keys" | while read -r key; do
                if [ -n "$key" ]; then
                    local key_type=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i type "$key" 2>/dev/null)
                    local ttl=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i ttl "$key" 2>/dev/null)
                    
                    echo "# Key: $key (Type: $key_type, TTL: $ttl)" >> "$backup_file"
                    
                    case $key_type in
                        "string")
                            docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i get "$key" 2>/dev/null >> "$backup_file"
                            ;;
                        "hash")
                            docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i hgetall "$key" 2>/dev/null >> "$backup_file"
                            ;;
                        "list")
                            docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i lrange "$key" 0 -1 2>/dev/null >> "$backup_file"
                            ;;
                        "set")
                            docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i smembers "$key" 2>/dev/null >> "$backup_file"
                            ;;
                        "zset")
                            docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i zrange "$key" 0 -1 withscores 2>/dev/null >> "$backup_file"
                            ;;
                    esac
                    echo "" >> "$backup_file"
                fi
            done
            
            echo -e "${GREEN}   ✅ DB$i backup: $(basename "$backup_file")${NC}"
        else
            echo -e "${YELLOW}   ⚠️  DB$i ($db_name): Sin datos${NC}"
        fi
    done
}

# Función para crear backup completo comprimido
create_full_backup() {
    echo -e "${YELLOW}4. Creando backup completo comprimido...${NC}"
    
    local full_backup_file="$BACKUP_DIR/gamc_redis_full_${TIMESTAMP}.tar.gz"
    
    # Comprimir todos los backups del timestamp actual
    if tar -czf "$full_backup_file" -C "$BACKUP_DIR" $(ls "$BACKUP_DIR" | grep "$TIMESTAMP") 2>/dev/null; then
        echo -e "${GREEN}✅ Backup completo creado: $(basename "$full_backup_file")${NC}"
        
        # Obtener tamaño del archivo
        local file_size=$(du -h "$full_backup_file" | cut -f1)
        echo -e "   Tamaño: ${GREEN}$file_size${NC}"
    else
        echo -e "${RED}❌ Error al crear backup completo${NC}"
    fi
}

# Función para limpiar backups antiguos
cleanup_old_backups() {
    echo -e "${YELLOW}5. Limpiando backups antiguos (>$RETENTION_DAYS días)...${NC}"
    
    local deleted_count=0
    
    # Buscar y eliminar archivos antiguos
    find "$BACKUP_DIR" -name "gamc_*" -type f -mtime +$RETENTION_DAYS | while read -r file; do
        if rm "$file" 2>/dev/null; then
            echo -e "   Eliminado: $(basename "$file")"
            deleted_count=$((deleted_count + 1))
        fi
    done
    
    if [ $deleted_count -eq 0 ]; then
        echo -e "${GREEN}✅ No hay backups antiguos para eliminar${NC}"
    else
        echo -e "${GREEN}✅ $deleted_count backups antiguos eliminados${NC}"
    fi
}

# Función para mostrar estadísticas del backup
show_backup_stats() {
    echo -e "${YELLOW}6. Estadísticas del backup...${NC}"
    
    # Contar archivos de backup
    local total_files=$(ls -1 "$BACKUP_DIR"/gamc_* 2>/dev/null | wc -l)
    local total_size=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1)
    
    echo -e "   Total de archivos de backup: ${GREEN}$total_files${NC}"
    echo -e "   Tamaño total del directorio: ${GREEN}$total_size${NC}"
    
    # Mostrar archivos del backup actual
    echo -e "   Archivos creados en este backup:"
    ls -lh "$BACKUP_DIR" | grep "$TIMESTAMP" | while read -r line; do
        echo -e "   ${GREEN}$line${NC}"
    done
}

# Función principal
main() {
    local exit_code=0
    
    echo -e "Iniciando backup de Redis en: ${GREEN}$(date)${NC}"
    echo -e "Directorio de backup: ${GREEN}$BACKUP_DIR${NC}"
    echo ""
    
    # Verificar conexión a Redis
    if ! docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping > /dev/null 2>&1; then
        echo -e "${RED}❌ No se puede conectar a Redis${NC}"
        exit 1
    fi
    
    # Ejecutar backups
    if ! backup_rdb; then
        exit_code=1
    fi
    
    backup_aof
    backup_databases
    create_full_backup
    cleanup_old_backups
    show_backup_stats
    
    echo ""
    echo -e "${BLUE}========================================${NC}"
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Backup completado exitosamente${NC}"
    else
        echo -e "${RED}❌ Backup completado con errores${NC}"
    fi
    echo -e "Finalizado: ${GREEN}$(date)${NC}"
    echo -e "${BLUE}========================================${NC}"
    
    exit $exit_code
}

# Verificar argumentos
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    echo "Uso: $0 [opciones]"
    echo ""
    echo "Variables de entorno:"
    echo "  REDIS_HOST        - Host de Redis (default: localhost)"
    echo "  REDIS_PORT        - Puerto de Redis (default: 6379)"
    echo "  REDIS_PASSWORD    - Contraseña de Redis"
    echo "  CONTAINER_NAME    - Nombre del contenedor (default: gamc_redis)"
    echo "  BACKUP_DIR        - Directorio de backups (default: ./backups)"
    echo "  RETENTION_DAYS    - Días de retención (default: 7)"
    echo ""
    echo "Ejemplos:"
    echo "  $0                           # Backup normal"
    echo "  BACKUP_DIR=/tmp/redis $0     # Backup en directorio personalizado"
    echo "  RETENTION_DAYS=30 $0         # Retener backups por 30 días"
    exit 0
fi

# Ejecutar backup
main "$@"
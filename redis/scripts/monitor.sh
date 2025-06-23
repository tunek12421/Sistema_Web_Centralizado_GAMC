#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Script de Monitoreo en Tiempo Real de Redis
# ========================================

set -e

# Configuración
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-gamc_redis_password_2024}
CONTAINER_NAME=${CONTAINER_NAME:-gamc_redis}
REFRESH_INTERVAL=${REFRESH_INTERVAL:-5}

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
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║           GAMC Redis - Monitor en Tiempo Real                 ║${NC}"
    echo -e "${BLUE}║           Actualización cada $REFRESH_INTERVAL segundos                        ║${NC}"
    echo -e "${BLUE}║           Presiona Ctrl+C para salir                          ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Función para obtener información general
get_general_info() {
    local info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info server 2>/dev/null)
    local version=$(echo "$info" | grep "redis_version:" | cut -d: -f2 | tr -d '\r')
    local uptime=$(echo "$info" | grep "uptime_in_seconds:" | cut -d: -f2 | tr -d '\r')
    local uptime_days=$((uptime / 86400))
    local uptime_hours=$(((uptime % 86400) / 3600))
    local uptime_minutes=$(((uptime % 3600) / 60))
    
    echo -e "${CYAN}📊 INFORMACIÓN GENERAL${NC}"
    echo -e "   Versión: ${GREEN}$version${NC}"
    echo -e "   Uptime: ${GREEN}${uptime_days}d ${uptime_hours}h ${uptime_minutes}m${NC}"
    echo -e "   Timestamp: ${GREEN}$(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# Función para obtener información de memoria
get_memory_info() {
    local memory_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info memory 2>/dev/null)
    local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
    local used_memory_peak=$(echo "$memory_info" | grep "used_memory_peak_human:" | cut -d: -f2 | tr -d '\r')
    local max_memory=$(echo "$memory_info" | grep "maxmemory_human:" | cut -d: -f2 | tr -d '\r')
    local fragmentation=$(echo "$memory_info" | grep "mem_fragmentation_ratio:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "${YELLOW}🧠 MEMORIA${NC}"
    echo -e "   Usada: ${GREEN}$used_memory${NC} | Pico: ${GREEN}$used_memory_peak${NC} | Máxima: ${GREEN}$max_memory${NC}"
    echo -e "   Fragmentación: ${GREEN}$fragmentation${NC}"
    echo ""
}

# Función para obtener información de clientes
get_clients_info() {
    local clients_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info clients 2>/dev/null)
    local connected_clients=$(echo "$clients_info" | grep "connected_clients:" | cut -d: -f2 | tr -d '\r')
    local blocked_clients=$(echo "$clients_info" | grep "blocked_clients:" | cut -d: -f2 | tr -d '\r')
    local max_clients=$(echo "$clients_info" | grep "maxclients:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "${MAGENTA}👥 CLIENTES${NC}"
    echo -e "   Conectados: ${GREEN}$connected_clients${NC} | Bloqueados: ${GREEN}$blocked_clients${NC} | Máximo: ${GREEN}$max_clients${NC}"
    echo ""
}

# Función para obtener estadísticas de comandos
get_stats_info() {
    local stats_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info stats 2>/dev/null)
    local total_commands=$(echo "$stats_info" | grep "total_commands_processed:" | cut -d: -f2 | tr -d '\r')
    local commands_per_sec=$(echo "$stats_info" | grep "instantaneous_ops_per_sec:" | cut -d: -f2 | tr -d '\r')
    local hits=$(echo "$stats_info" | grep "keyspace_hits:" | cut -d: -f2 | tr -d '\r')
    local misses=$(echo "$stats_info" | grep "keyspace_misses:" | cut -d: -f2 | tr -d '\r')
    
    # Calcular hit ratio
    local hit_ratio=0
    if [ "$hits" -gt 0 ] && [ "$misses" -gt 0 ]; then
        hit_ratio=$((hits * 100 / (hits + misses)))
    fi
    
    echo -e "${BLUE}📈 ESTADÍSTICAS${NC}"
    echo -e "   Comandos totales: ${GREEN}$total_commands${NC} | Ops/seg: ${GREEN}$commands_per_sec${NC}"
    echo -e "   Cache hits: ${GREEN}$hits${NC} | Misses: ${GREEN}$misses${NC} | Hit ratio: ${GREEN}${hit_ratio}%${NC}"
    echo ""
}

# Función para obtener información de bases de datos
get_keyspace_info() {
    echo -e "${CYAN}🗂️  BASES DE DATOS${NC}"
    
    local db_names=("Sesiones" "Cache" "Notificaciones" "Métricas" "Query Cache" "JWT Blacklist")
    
    for i in {0..5}; do
        local keys_count=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $i dbsize 2>/dev/null)
        local db_name=${db_names[$i]}
        
        if [ "$keys_count" -gt 0 ]; then
            echo -e "   DB$i (${db_name}): ${GREEN}$keys_count llaves${NC}"
        else
            echo -e "   DB$i (${db_name}): ${YELLOW}0 llaves${NC}"
        fi
    done
    echo ""
}

# Función para obtener comandos más usados
get_top_commands() {
    echo -e "${GREEN}🔝 TOP COMANDOS${NC}"
    
    local commands=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info commandstats 2>/dev/null | grep "cmdstat_" | head -10)
    
    if [ -n "$commands" ]; then
        echo "$commands" | while read -r line; do
            local cmd=$(echo "$line" | cut -d: -f1 | sed 's/cmdstat_//')
            local stats=$(echo "$line" | cut -d: -f2)
            local calls=$(echo "$stats" | sed 's/.*calls=\([0-9]*\).*/\1/')
            echo -e "   ${cmd}: ${GREEN}$calls${NC} llamadas"
        done
    else
        echo -e "   ${YELLOW}No hay estadísticas de comandos disponibles${NC}"
    fi
    echo ""
}

# Función para obtener slow queries
get_slow_queries() {
    echo -e "${RED}🐌 CONSULTAS LENTAS${NC}"
    
    local slowlog_len=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD slowlog len 2>/dev/null)
    
    if [ "$slowlog_len" -gt 0 ]; then
        echo -e "   Total en slow log: ${RED}$slowlog_len${NC}"
        echo -e "   Últimas 3 consultas:"
        
        local slowlog=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD slowlog get 3 2>/dev/null)
        echo "$slowlog" | grep -E "^\d+\)|[0-9]+\)" | head -3 | while read -r line; do
            echo -e "   ${YELLOW}$line${NC}"
        done
    else
        echo -e "   ${GREEN}No hay consultas lentas recientes${NC}"
    fi
    echo ""
}

# Función para obtener información de persistencia
get_persistence_info() {
    local persistence_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info persistence 2>/dev/null)
    local rdb_last_save=$(echo "$persistence_info" | grep "rdb_last_save_time:" | cut -d: -f2 | tr -d '\r')
    local aof_enabled=$(echo "$persistence_info" | grep "aof_enabled:" | cut -d: -f2 | tr -d '\r')
    local rdb_changes=$(echo "$persistence_info" | grep "rdb_changes_since_last_save:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "${MAGENTA}💾 PERSISTENCIA${NC}"
    
    if [ "$aof_enabled" = "1" ]; then
        echo -e "   AOF: ${GREEN}Habilitado${NC}"
    else
        echo -e "   AOF: ${YELLOW}Deshabilitado${NC}"
    fi
    
    if [ -n "$rdb_last_save" ] && [ "$rdb_last_save" != "0" ]; then
        local last_save_date=$(date -d "@$rdb_last_save" '+%Y-%m-%d %H:%M:%S' 2>/dev/null || echo "Fecha inválida")
        echo -e "   Último RDB: ${GREEN}$last_save_date${NC}"
    else
        echo -e "   Último RDB: ${YELLOW}Nunca${NC}"
    fi
    
    echo -e "   Cambios desde último save: ${GREEN}$rdb_changes${NC}"
    echo ""
}

# Función para mostrar actividad en tiempo real
show_activity() {
    echo -e "${CYAN}🔄 ACTIVIDAD EN TIEMPO REAL (últimos 5 comandos)${NC}"
    
    # Usar timeout para limitar el monitor
    timeout 1s docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD monitor 2>/dev/null | head -5 | while read -r line; do
        echo -e "   ${GREEN}$line${NC}"
    done 2>/dev/null || echo -e "   ${YELLOW}Monitor no disponible en este momento${NC}"
    echo ""
}

# Función principal de monitoreo
monitor_redis() {
    # Verificar conexión
    if ! docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping > /dev/null 2>&1; then
        echo -e "${RED}❌ No se puede conectar a Redis${NC}"
        exit 1
    fi
    
    # Loop principal
    while true; do
        clear_screen
        get_general_info
        get_memory_info
        get_clients_info
        get_stats_info
        get_keyspace_info
        get_top_commands
        get_slow_queries
        get_persistence_info
        show_activity
        
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
    echo "  -i, --interval N  Intervalo de actualización en segundos (default: 5)"
    echo ""
    echo "Variables de entorno:"
    echo "  REDIS_HOST        Host de Redis (default: localhost)"
    echo "  REDIS_PORT        Puerto de Redis (default: 6379)"
    echo "  REDIS_PASSWORD    Contraseña de Redis"
    echo "  CONTAINER_NAME    Nombre del contenedor (default: gamc_redis)"
    echo "  REFRESH_INTERVAL  Intervalo de actualización (default: 5)"
    echo ""
    echo "Ejemplos:"
    echo "  $0                    # Monitor con configuración por defecto"
    echo "  $0 -i 10              # Actualizar cada 10 segundos"
    echo "  REDIS_HOST=prod $0    # Conectar a servidor de producción"
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

# Manejo de señales para salida limpia
trap 'echo -e "\n${GREEN}Monitoreo detenido${NC}"; exit 0' INT TERM

# Iniciar monitoreo
echo -e "${GREEN}Iniciando monitor de Redis...${NC}"
monitor_redis
#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Script de Verificación de Salud de Redis
# ========================================

set -e

# Configuración
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-gamc_redis_password_2024}
CONTAINER_NAME=${CONTAINER_NAME:-gamc_redis}

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GAMC Redis - Verificación de Salud${NC}"
echo -e "${BLUE}========================================${NC}"

# Función para verificar conexión
check_connection() {
    echo -e "${YELLOW}1. Verificando conexión a Redis...${NC}"
    if docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD ping > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Redis está respondiendo${NC}"
        return 0
    else
        echo -e "${RED}❌ Redis no está respondiendo${NC}"
        return 1
    fi
}

# Función para verificar memoria
check_memory() {
    echo -e "${YELLOW}2. Verificando uso de memoria...${NC}"
    
    local memory_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info memory 2>/dev/null)
    local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
    local max_memory=$(echo "$memory_info" | grep "maxmemory_human:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "   Memoria usada: ${GREEN}$used_memory${NC}"
    echo -e "   Memoria máxima: ${GREEN}$max_memory${NC}"
    
    # Verificar porcentaje de uso
    local used_bytes=$(echo "$memory_info" | grep "used_memory:" | cut -d: -f2 | tr -d '\r')
    local max_bytes=$(echo "$memory_info" | grep "maxmemory:" | cut -d: -f2 | tr -d '\r')
    
    if [ "$max_bytes" -gt 0 ]; then
        local usage_percent=$((used_bytes * 100 / max_bytes))
        if [ $usage_percent -gt 80 ]; then
            echo -e "${RED}⚠️  Uso de memoria alto: ${usage_percent}%${NC}"
        else
            echo -e "${GREEN}✅ Uso de memoria normal: ${usage_percent}%${NC}"
        fi
    fi
}

# Función para verificar persistencia
check_persistence() {
    echo -e "${YELLOW}3. Verificando persistencia...${NC}"
    
    local persistence_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info persistence 2>/dev/null)
    local rdb_last_save=$(echo "$persistence_info" | grep "rdb_last_save_time:" | cut -d: -f2 | tr -d '\r')
    local aof_enabled=$(echo "$persistence_info" | grep "aof_enabled:" | cut -d: -f2 | tr -d '\r')
    
    if [ "$aof_enabled" = "1" ]; then
        echo -e "${GREEN}✅ AOF habilitado${NC}"
    else
        echo -e "${YELLOW}⚠️  AOF deshabilitado${NC}"
    fi
    
    if [ -n "$rdb_last_save" ] && [ "$rdb_last_save" != "0" ]; then
        local last_save_date=$(date -d "@$rdb_last_save" 2>/dev/null || echo "Fecha inválida")
        echo -e "${GREEN}✅ Último RDB guardado: $last_save_date${NC}"
    else
        echo -e "${YELLOW}⚠️  No hay snapshots RDB recientes${NC}"
    fi
}

# Función para verificar clientes conectados
check_clients() {
    echo -e "${YELLOW}4. Verificando clientes conectados...${NC}"
    
    local clients_info=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info clients 2>/dev/null)
    local connected_clients=$(echo "$clients_info" | grep "connected_clients:" | cut -d: -f2 | tr -d '\r')
    local max_clients=$(echo "$clients_info" | grep "maxclients:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "   Clientes conectados: ${GREEN}$connected_clients${NC}"
    echo -e "   Máximo de clientes: ${GREEN}$max_clients${NC}"
    
    if [ "$connected_clients" -gt $((max_clients * 80 / 100)) ]; then
        echo -e "${RED}⚠️  Muchos clientes conectados${NC}"
    else
        echo -e "${GREEN}✅ Número de clientes normal${NC}"
    fi
}

# Función para verificar bases de datos
check_databases() {
    echo -e "${YELLOW}5. Verificando bases de datos...${NC}"
    
    for db in 0 1 2 3 4 5; do
        local keys_count=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD -n $db dbsize 2>/dev/null)
        local db_name=""
        
        case $db in
            0) db_name="Sesiones" ;;
            1) db_name="Cache" ;;
            2) db_name="Notificaciones" ;;
            3) db_name="Métricas" ;;
            4) db_name="Query Cache" ;;
            5) db_name="JWT Blacklist" ;;
        esac
        
        echo -e "   DB$db ($db_name): ${GREEN}$keys_count llaves${NC}"
    done
}

# Función para verificar slow log
check_slowlog() {
    echo -e "${YELLOW}6. Verificando slow queries...${NC}"
    
    local slowlog_len=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD slowlog len 2>/dev/null)
    
    if [ "$slowlog_len" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  $slowlog_len consultas lentas en el log${NC}"
        echo -e "   Últimas 3 consultas lentas:"
        docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD slowlog get 3 2>/dev/null | head -15
    else
        echo -e "${GREEN}✅ No hay consultas lentas recientes${NC}"
    fi
}

# Función para mostrar estadísticas generales
show_stats() {
    echo -e "${YELLOW}7. Estadísticas generales...${NC}"
    
    local stats=$(docker exec $CONTAINER_NAME redis-cli -h $REDIS_HOST -p $REDIS_PORT -a $REDIS_PASSWORD info stats 2>/dev/null)
    local total_commands=$(echo "$stats" | grep "total_commands_processed:" | cut -d: -f2 | tr -d '\r')
    local hits=$(echo "$stats" | grep "keyspace_hits:" | cut -d: -f2 | tr -d '\r')
    local misses=$(echo "$stats" | grep "keyspace_misses:" | cut -d: -f2 | tr -d '\r')
    
    echo -e "   Comandos procesados: ${GREEN}$total_commands${NC}"
    echo -e "   Cache hits: ${GREEN}$hits${NC}"
    echo -e "   Cache misses: ${GREEN}$misses${NC}"
    
    if [ "$hits" -gt 0 ] && [ "$misses" -gt 0 ]; then
        local hit_ratio=$((hits * 100 / (hits + misses)))
        if [ $hit_ratio -gt 80 ]; then
            echo -e "   Hit ratio: ${GREEN}${hit_ratio}%${NC}"
        else
            echo -e "   Hit ratio: ${YELLOW}${hit_ratio}%${NC}"
        fi
    fi
}

# Ejecutar verificaciones
main() {
    local exit_code=0
    
    if ! check_connection; then
        exit_code=1
    fi
    
    if [ $exit_code -eq 0 ]; then
        check_memory
        check_persistence
        check_clients
        check_databases
        check_slowlog
        show_stats
    fi
    
    echo -e "${BLUE}========================================${NC}"
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ Verificación completada exitosamente${NC}"
    else
        echo -e "${RED}❌ Se encontraron problemas${NC}"
    fi
    echo -e "${BLUE}========================================${NC}"
    
    exit $exit_code
}

# Ejecutar script principal
main "$@"
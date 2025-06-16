#!/bin/bash

# ========================================
# GAMC Sistema Web Centralizado
# Script de Backup de MinIO
# ========================================

set -e

# Configuración
MINIO_ALIAS=${MINIO_ALIAS:-"gamc-local"}
MINIO_ENDPOINT=${MINIO_ENDPOINT:-"http://localhost:9000"}
MINIO_ACCESS_KEY=${MINIO_ACCESS_KEY:-"gamc_admin"}
MINIO_SECRET_KEY=${MINIO_SECRET_KEY:-"gamc_minio_password_2024"}
BACKUP_DIR=${BACKUP_DIR:-"./backups"}
RETENTION_DAYS=${RETENTION_DAYS:-30}
EXTERNAL_S3_ENDPOINT=${EXTERNAL_S3_ENDPOINT:-""}
EXTERNAL_S3_BUCKET=${EXTERNAL_S3_BUCKET:-"gamc-backup-external"}

# Crear directorio de backups si no existe
mkdir -p "$BACKUP_DIR"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Timestamp para archivos
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
DATE_DIR=$(date +"%Y/%m/%d")

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GAMC MinIO - Backup de Datos${NC}"
echo -e "${BLUE}========================================${NC}"

# Función para configurar cliente MinIO
setup_minio_client() {
    echo -e "${YELLOW}1. Configurando cliente MinIO...${NC}"
    
    if ! command -v mc &> /dev/null; then
        echo -e "${RED}❌ MinIO client (mc) no está instalado${NC}"
        echo -e "${YELLOW}Instalando mc...${NC}"
        curl -O https://dl.min.io/client/mc/release/linux-amd64/mc
        chmod +x mc
        sudo mv mc /usr/local/bin/
    fi
    
    # Configurar alias para MinIO local
    if mc alias set $MINIO_ALIAS $MINIO_ENDPOINT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY; then
        echo -e "${GREEN}✅ Cliente MinIO configurado${NC}"
    else
        echo -e "${RED}❌ Error al configurar cliente MinIO${NC}"
        exit 1
    fi
    
    # Verificar conexión
    if mc admin info $MINIO_ALIAS > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Conexión a MinIO verificada${NC}"
    else
        echo -e "${RED}❌ No se puede conectar a MinIO${NC}"
        exit 1
    fi
}

# Función para listar buckets
list_buckets() {
    echo -e "${YELLOW}2. Listando buckets disponibles...${NC}"
    
    local buckets=$(mc ls $MINIO_ALIAS | awk '{print $5}' | grep -v "^$")
    
    if [ -z "$buckets" ]; then
        echo -e "${RED}❌ No se encontraron buckets${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Buckets encontrados:${NC}"
    echo "$buckets" | while read -r bucket; do
        local size=$(mc du $MINIO_ALIAS/$bucket 2>/dev/null | tail -1 | awk '{print $1, $2}' || echo "0 B")
        local count=$(mc ls $MINIO_ALIAS/$bucket --recursive 2>/dev/null | wc -l || echo "0")
        echo -e "   ${GREEN}📦 $bucket${NC} - Tamaño: $size, Archivos: $count"
    done
    
    echo "$buckets"
}

# Función para hacer backup de un bucket específico
backup_bucket() {
    local bucket_name="$1"
    local backup_type="$2"  # full, incremental
    
    echo -e "${YELLOW}3. Haciendo backup del bucket: $bucket_name${NC}"
    
    local bucket_backup_dir="$BACKUP_DIR/$bucket_name/$DATE_DIR"
    mkdir -p "$bucket_backup_dir"
    
    case $backup_type in
        "full")
            echo -e "   Tipo: ${CYAN}Backup completo${NC}"
            if mc mirror $MINIO_ALIAS/$bucket_name "$bucket_backup_dir/full_$TIMESTAMP"; then
                echo -e "   ${GREEN}✅ Backup completo de $bucket_name realizado${NC}"
                
                # Crear archivo de metadatos
                cat > "$bucket_backup_dir/full_$TIMESTAMP.meta" << EOF
{
  "bucket": "$bucket_name",
  "type": "full",
  "timestamp": "$TIMESTAMP",
  "date": "$(date -Iseconds)",
  "source": "$MINIO_ALIAS/$bucket_name",
  "destination": "$bucket_backup_dir/full_$TIMESTAMP",
  "files_count": $(mc ls $MINIO_ALIAS/$bucket_name --recursive 2>/dev/null | wc -l),
  "backup_size": "$(du -sh "$bucket_backup_dir/full_$TIMESTAMP" 2>/dev/null | cut -f1 || echo 'unknown')"
}
EOF
                return 0
            else
                echo -e "   ${RED}❌ Error en backup de $bucket_name${NC}"
                return 1
            fi
            ;;
            
        "incremental")
            echo -e "   Tipo: ${CYAN}Backup incremental${NC}"
            
            # Buscar último backup completo
            local last_full=$(find "$BACKUP_DIR/$bucket_name" -name "full_*" -type d 2>/dev/null | sort | tail -1)
            
            if [ -z "$last_full" ]; then
                echo -e "   ${YELLOW}⚠️  No hay backup completo previo, realizando backup completo${NC}"
                backup_bucket "$bucket_name" "full"
                return $?
            fi
            
            local incremental_dir="$bucket_backup_dir/incremental_$TIMESTAMP"
            
            # Hacer backup incremental usando rsync-like behavior
            if mc mirror $MINIO_ALIAS/$bucket_name "$incremental_dir" --newer-than="24h"; then
                echo -e "   ${GREEN}✅ Backup incremental de $bucket_name realizado${NC}"
                
                # Crear archivo de metadatos
                cat > "$incremental_dir.meta" << EOF
{
  "bucket": "$bucket_name",
  "type": "incremental",
  "timestamp": "$TIMESTAMP",
  "date": "$(date -Iseconds)",
  "source": "$MINIO_ALIAS/$bucket_name",
  "destination": "$incremental_dir",
  "base_backup": "$last_full",
  "files_count": $(find "$incremental_dir" -type f 2>/dev/null | wc -l),
  "backup_size": "$(du -sh "$incremental_dir" 2>/dev/null | cut -f1 || echo 'unknown')"
}
EOF
                return 0
            else
                echo -e "   ${RED}❌ Error en backup incremental de $bucket_name${NC}"
                return 1
            fi
            ;;
    esac
}

# Función para comprimir backups antiguos
compress_old_backups() {
    echo -e "${YELLOW}4. Comprimiendo backups antiguos...${NC}"
    
    # Buscar directorios de backup de más de 7 días sin comprimir
    find "$BACKUP_DIR" -type d -name "*_*" -mtime +7 ! -name "*.tar.gz" | while read -r backup_dir; do
        if [ -d "$backup_dir" ]; then
            echo -e "   Comprimiendo: $(basename "$backup_dir")"
            
            local compressed_file="${backup_dir}.tar.gz"
            if tar -czf "$compressed_file" -C "$(dirname "$backup_dir")" "$(basename "$backup_dir")"; then
                rm -rf "$backup_dir"
                echo -e "   ${GREEN}✅ Comprimido: $(basename "$compressed_file")${NC}"
            else
                echo -e "   ${RED}❌ Error al comprimir: $(basename "$backup_dir")${NC}"
            fi
        fi
    done
}

# Función para sincronizar con S3 externo
sync_to_external_s3() {
    echo -e "${YELLOW}5. Sincronizando con S3 externo...${NC}"
    
    if [ -z "$EXTERNAL_S3_ENDPOINT" ]; then
        echo -e "${YELLOW}⚠️  S3 externo no configurado${NC}"
        return 0
    fi
    
    echo -e "   Configurando S3 externo..."
    if mc alias set external-s3 "$EXTERNAL_S3_ENDPOINT" "$EXTERNAL_S3_ACCESS_KEY" "$EXTERNAL_S3_SECRET_KEY"; then
        echo -e "   ${GREEN}✅ S3 externo configurado${NC}"
        
        # Sincronizar backups importantes
        if mc mirror "$BACKUP_DIR" "external-s3/$EXTERNAL_S3_BUCKET/gamc-backups/$TIMESTAMP"; then
            echo -e "   ${GREEN}✅ Sincronización con S3 externo completada${NC}"
        else
            echo -e "   ${RED}❌ Error en sincronización con S3 externo${NC}"
        fi
    else
        echo -e "   ${RED}❌ Error al configurar S3 externo${NC}"
    fi
}

# Función para limpiar backups antiguos
cleanup_old_backups() {
    echo -e "${YELLOW}6. Limpiando backups antiguos (>$RETENTION_DAYS días)...${NC}"
    
    local deleted_count=0
    
    # Eliminar backups antiguos
    find "$BACKUP_DIR" -type f \( -name "*.tar.gz" -o -name "*.meta" \) -mtime +$RETENTION_DAYS | while read -r file; do
        if rm "$file" 2>/dev/null; then
            echo -e "   Eliminado: $(basename "$file")"
            deleted_count=$((deleted_count + 1))
        fi
    done
    
    # Eliminar directorios vacíos
    find "$BACKUP_DIR" -type d -empty -delete 2>/dev/null || true
    
    if [ $deleted_count -eq 0 ]; then
        echo -e "${GREEN}✅ No hay backups antiguos para eliminar${NC}"
    else
        echo -e "${GREEN}✅ $deleted_count archivos antiguos eliminados${NC}"
    fi
}

# Función para mostrar estadísticas del backup
show_backup_stats() {
    echo -e "${YELLOW}7. Estadísticas del backup...${NC}"
    
    # Tamaño total de backups
    local total_size=$(du -sh "$BACKUP_DIR" 2>/dev/null | cut -f1 || echo "0")
    local total_files=$(find "$BACKUP_DIR" -type f 2>/dev/null | wc -l)
    
    echo -e "   Tamaño total de backups: ${GREEN}$total_size${NC}"
    echo -e "   Total de archivos: ${GREEN}$total_files${NC}"
    
    # Estadísticas por bucket
    echo -e "   Backups por bucket:"
    find "$BACKUP_DIR" -maxdepth 1 -type d ! -path "$BACKUP_DIR" | while read -r bucket_dir; do
        local bucket_name=$(basename "$bucket_dir")
        local bucket_size=$(du -sh "$bucket_dir" 2>/dev/null | cut -f1 || echo "0")
        local bucket_files=$(find "$bucket_dir" -type f 2>/dev/null | wc -l)
        echo -e "     ${CYAN}$bucket_name${NC}: $bucket_size ($bucket_files archivos)"
    done
    
    # Último backup de cada bucket
    echo -e "   Último backup por bucket:"
    find "$BACKUP_DIR" -name "*.meta" -exec ls -t {} \; 2>/dev/null | head -10 | while read -r meta_file; do
        local bucket_name=$(dirname "$meta_file" | xargs dirname | xargs basename)
        local backup_date=$(grep '"date"' "$meta_file" 2>/dev/null | cut -d'"' -f4 || echo "unknown")
        echo -e "     ${CYAN}$bucket_name${NC}: $backup_date"
    done
}

# Función principal
main() {
    local backup_type=${1:-"incremental"}
    local specific_bucket=${2:-""}
    local exit_code=0
    
    echo -e "Iniciando backup de MinIO: ${GREEN}$(date)${NC}"
    echo -e "Tipo de backup: ${GREEN}$backup_type${NC}"
    echo -e "Directorio de backup: ${GREEN}$BACKUP_DIR${NC}"
    echo ""
    
    # Configurar cliente
    setup_minio_client
    
    # Listar buckets disponibles
    local buckets=$(list_buckets)
    
    # Realizar backup
    if [ -n "$specific_bucket" ]; then
        echo -e "${CYAN}Backup específico del bucket: $specific_bucket${NC}"
        if ! backup_bucket "$specific_bucket" "$backup_type"; then
            exit_code=1
        fi
    else
        echo -e "${CYAN}Backup de todos los buckets...${NC}"
        echo "$buckets" | while read -r bucket; do
            if [ -n "$bucket" ]; then
                if ! backup_bucket "$bucket" "$backup_type"; then
                    exit_code=1
                fi
            fi
        done
    fi
    
    # Comprimir backups antiguos
    compress_old_backups
    
    # Sincronizar con S3 externo si está configurado
    sync_to_external_s3
    
    # Limpiar backups antiguos
    cleanup_old_backups
    
    # Mostrar estadísticas
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

# Función para mostrar ayuda
show_help() {
    echo "Uso: $0 [tipo] [bucket]"
    echo ""
    echo "Tipos de backup:"
    echo "  full         - Backup completo de todos los archivos"
    echo "  incremental  - Backup incremental (solo cambios recientes)"
    echo ""
    echo "Parámetros:"
    echo "  bucket       - Nombre específico del bucket (opcional)"
    echo ""
    echo "Variables de entorno:"
    echo "  MINIO_ENDPOINT       - Endpoint de MinIO (default: http://localhost:9000)"
    echo "  MINIO_ACCESS_KEY     - Access key (default: gamc_admin)"
    echo "  MINIO_SECRET_KEY     - Secret key"
    echo "  BACKUP_DIR           - Directorio de backups (default: ./backups)"
    echo "  RETENTION_DAYS       - Días de retención (default: 30)"
    echo "  EXTERNAL_S3_ENDPOINT - S3 externo para sincronización"
    echo ""
    echo "Ejemplos:"
    echo "  $0                           # Backup incremental de todos los buckets"
    echo "  $0 full                      # Backup completo de todos los buckets"
    echo "  $0 incremental gamc-attachments  # Backup incremental de un bucket específico"
    echo "  RETENTION_DAYS=60 $0 full    # Backup completo con retención de 60 días"
}

# Verificar argumentos
case "$1" in
    "-h"|"--help")
        show_help
        exit 0
        ;;
    "full"|"incremental"|"")
        main "$@"
        ;;
    *)
        echo "Tipo de backup inválido: $1"
        show_help
        exit 1
        ;;
esac
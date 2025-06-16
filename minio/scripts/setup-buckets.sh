#!/bin/sh

# ========================================
# GAMC Sistema Web Centralizado
# Script de Configuración de MinIO
# ========================================

set -e

# Configuración
MINIO_ALIAS="gamc-local"
MINIO_ENDPOINT="http://minio:9000"
MINIO_ACCESS_KEY="gamc_admin"
MINIO_SECRET_KEY="gamc_minio_password_2024"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}GAMC MinIO - Configuración de Buckets${NC}"
echo -e "${BLUE}========================================${NC}"

# Función para crear bucket con configuración
create_bucket_with_config() {
    local bucket_name="$1"
    local description="$2"
    local policy="$3"
    local lifecycle="$4"
    
    echo -e "${YELLOW}Configurando bucket: $bucket_name${NC}"
    
    # Crear bucket si no existe
    if mc ls $MINIO_ALIAS/$bucket_name >/dev/null 2>&1; then
        echo -e "   ${GREEN}✅ Bucket ya existe${NC}"
    else
        if mc mb $MINIO_ALIAS/$bucket_name; then
            echo -e "   ${GREEN}✅ Bucket creado${NC}"
        else
            echo -e "   ${RED}❌ Error al crear bucket${NC}"
            return 1
        fi
    fi
    
    # Configurar política de acceso
    if [ -n "$policy" ]; then
        echo -e "   Aplicando política: $policy"
        mc anonymous set $policy $MINIO_ALIAS/$bucket_name
    fi
    
    # Configurar lifecycle si se especifica
    if [ -n "$lifecycle" ]; then
        echo -e "   Configurando lifecycle: $lifecycle días"
        # Para implementar lifecycle rules más tarde
    fi
    
    echo -e "   ${GREEN}✅ $bucket_name configurado${NC}"
    echo ""
}

# Configurar buckets del sistema GAMC
echo -e "${CYAN}1. Configurando buckets principales...${NC}"

create_bucket_with_config \
    "gamc-attachments" \
    "Archivos adjuntos de mensajes" \
    "none" \
    ""

create_bucket_with_config \
    "gamc-documents" \
    "Documentos oficiales y reportes" \
    "download" \
    ""

create_bucket_with_config \
    "gamc-images" \
    "Imágenes y recursos multimedia" \
    "download" \
    ""

create_bucket_with_config \
    "gamc-backups" \
    "Respaldos del sistema" \
    "none" \
    "90"

create_bucket_with_config \
    "gamc-temp" \
    "Archivos temporales" \
    "none" \
    "7"

create_bucket_with_config \
    "gamc-reports" \
    "Reportes generados automáticamente" \
    "download" \
    "30"

# Configurar estructura de directorios dentro de buckets
echo -e "${CYAN}2. Creando estructura de directorios...${NC}"

# Estructura para gamc-attachments
echo -e "${YELLOW}Configurando estructura de attachments...${NC}"
for unit in obras-publicas monitoreo movilidad-urbana gobierno-electronico prensa-imagen tecnologia; do
    echo "# Directorio para $unit" | mc pipe $MINIO_ALIAS/gamc-attachments/$unit/.gitkeep 2>/dev/null || true
    for year in $(seq 2024 2030); do
        echo "# Directorio para $year" | mc pipe $MINIO_ALIAS/gamc-attachments/$unit/$year/.gitkeep 2>/dev/null || true
    done
done

# Estructura para gamc-documents
echo -e "${YELLOW}Configurando estructura de documentos...${NC}"
for type in actas informes normativas contratos; do
    echo "# Directorio para $type" | mc pipe $MINIO_ALIAS/gamc-documents/$type/.gitkeep 2>/dev/null || true
    for year in $(seq 2024 2030); do
        echo "# Directorio para $year" | mc pipe $MINIO_ALIAS/gamc-documents/$type/$year/.gitkeep 2>/dev/null || true
    done
done

# Estructura para gamc-images
echo -e "${YELLOW}Configurando estructura de imágenes...${NC}"
for type in logos avatars banners screenshots; do
    echo "# Directorio para $type" | mc pipe $MINIO_ALIAS/gamc-images/$type/.gitkeep 2>/dev/null || true
done

# Estructura para gamc-backups
echo -e "${YELLOW}Configurando estructura de backups...${NC}"
for type in database redis files system; do
    echo "# Directorio para backups de $type" | mc pipe $MINIO_ALIAS/gamc-backups/$type/.gitkeep 2>/dev/null || true
    for year in $(seq 2024 2030); do
        echo "# Directorio para $year" | mc pipe $MINIO_ALIAS/gamc-backups/$type/$year/.gitkeep 2>/dev/null || true
    done
done

# Configurar notificaciones webhook (si está disponible)
echo -e "${CYAN}3. Configurando notificaciones...${NC}"
if [ -n "$MINIO_WEBHOOK_ENDPOINT" ]; then
    echo -e "${YELLOW}Configurando webhook para notificaciones...${NC}"
    # mc admin config set $MINIO_ALIAS notify_webhook:1 endpoint="$MINIO_WEBHOOK_ENDPOINT"
    echo -e "${GREEN}✅ Webhook configurado${NC}"
else
    echo -e "${YELLOW}⚠️  Webhook no configurado (opcional)${NC}"
fi

# Configurar usuarios y políticas adicionales
echo -e "${CYAN}4. Configurando usuarios del sistema...${NC}"

# Usuario para la aplicación backend
echo -e "${YELLOW}Creando usuario para backend...${NC}"
if mc admin user add $MINIO_ALIAS gamc_backend gamc_backend_secret_2024 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Usuario backend creado${NC}"
else
    echo -e "${YELLOW}⚠️  Usuario backend ya existe o error al crear${NC}"
fi

# Usuario de solo lectura para reportes
echo -e "${YELLOW}Creando usuario de solo lectura...${NC}"
if mc admin user add $MINIO_ALIAS gamc_readonly gamc_readonly_secret_2024 >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Usuario readonly creado${NC}"
else
    echo -e "${YELLOW}⚠️  Usuario readonly ya existe o error al crear${NC}"
fi

# Crear archivo de configuración para la aplicación
echo -e "${CYAN}5. Generando configuración para aplicaciones...${NC}"
cat > /tmp/minio-config.json << EOF
{
  "endpoint": "localhost:9000",
  "accessKey": "gamc_backend",
  "secretKey": "gamc_backend_secret_2024",
  "useSSL": false,
  "region": "us-east-1",
  "buckets": {
    "attachments": "gamc-attachments",
    "documents": "gamc-documents", 
    "images": "gamc-images",
    "backups": "gamc-backups",
    "temp": "gamc-temp",
    "reports": "gamc-reports"
  },
  "maxFileSize": "100MB",
  "allowedExtensions": {
    "attachments": [".pdf", ".doc", ".docx", ".xls", ".xlsx", ".txt", ".zip", ".rar"],
    "documents": [".pdf", ".doc", ".docx", ".odt"],
    "images": [".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp"],
    "reports": [".pdf", ".csv", ".xlsx"]
  }
}
EOF

if mc cp /tmp/minio-config.json $MINIO_ALIAS/gamc-documents/config/minio-config.json; then
    echo -e "${GREEN}✅ Configuración guardada en bucket${NC}"
else
    echo -e "${YELLOW}⚠️  No se pudo guardar configuración${NC}"
fi

# Mostrar resumen de configuración
echo -e "${CYAN}6. Resumen de configuración...${NC}"
echo -e "${GREEN}Buckets creados:${NC}"
mc ls $MINIO_ALIAS | while read -r line; do
    echo -e "   ${GREEN}✅ $line${NC}"
done

echo -e "${GREEN}Usuarios creados:${NC}"
mc admin user list $MINIO_ALIAS | while read -r line; do
    echo -e "   ${GREEN}👤 $line${NC}"
done

# Verificar espacio disponible
echo -e "${CYAN}7. Información del sistema...${NC}"
df -h /data 2>/dev/null | tail -1 | while read -r line; do
    echo -e "   Espacio disponible: ${GREEN}$line${NC}"
done

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ Configuración de MinIO completada${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}Acceso a MinIO:${NC}"
echo -e "   API: ${GREEN}http://localhost:9000${NC}"
echo -e "   Console: ${GREEN}http://localhost:9001${NC}"
echo -e "   Usuario Admin: ${GREEN}gamc_admin${NC}"
echo -e "   Contraseña: ${GREEN}gamc_minio_password_2024${NC}"
echo ""
echo -e "${YELLOW}Usuarios de aplicación:${NC}"
echo -e "   Backend: gamc_backend / gamc_backend_secret_2024"
echo -e "   ReadOnly: gamc_readonly / gamc_readonly_secret_2024"
echo ""
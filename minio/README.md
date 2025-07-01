# MinIO - Sistema Web Centralizado GAMC

## Descripción
Sistema de almacenamiento de objetos S3-compatible para el Sistema Web Centralizado del GAMC, diseñado para gestionar archivos adjuntos, documentos oficiales, imágenes y respaldos de manera segura y escalable.

## Características
- ✅ MinIO Server con interfaz web de administración
- ✅ 6 buckets organizados por tipo de contenido
- ✅ Políticas de acceso granulares
- ✅ Gateway Nginx con proxy y cache
- ✅ Scripts de backup y monitoreo automatizados
- ✅ Configuración de usuarios y permisos
- ✅ Compatible con SDK de AWS S3

## Inicio Rápido

### 1. Levantar MinIO Server
```bash
# Desde el directorio minio/
docker-compose -f docker-compose.minio.yml up -d minio
```

### 2. Configurar Buckets y Usuarios
```bash
# Ejecutar configuración inicial
docker-compose -f docker-compose.minio.yml --profile setup up minio-client

# Ver logs de configuración
docker logs gamc_minio_client
```

### 3. Verificar Instalación
```bash
# Verificar estado de MinIO
docker exec gamc_minio curl -f http://localhost:9000/minio/health/live

# Listar buckets creados
docker exec gamc_minio_client mc ls gamc-local
```

### 4. Acceder a Interfaces Web
```bash
# Opcional: Levantar gateway web
docker-compose -f docker-compose.minio.yml --profile gateway up -d

# Acceder a interfaces:
# - Consola MinIO: http://localhost:9001
# - Gateway GAMC: http://localhost:8090
# - API directa: http://localhost:9000
```

## Estructura de Buckets

El sistema está organizado en 6 buckets especializados:

| Bucket | Propósito | Política | Retención |
|--------|-----------|----------|-----------|
| **gamc-attachments** | Archivos adjuntos de mensajes | Privado | 5 años |
| **gamc-documents** | Documentos oficiales y normativas | Público (lectura) | Permanente |
| **gamc-images** | Logos, avatars, banners | Público (lectura) | Permanente |
| **gamc-backups** | Respaldos del sistema | Privado | 90 días |
| **gamc-temp** | Archivos temporales | Privado | 7 días |
| **gamc-reports** | Reportes generados | Público (lectura) | 30 días |

### Estructura de Directorios

```
gamc-attachments/
├── obras-publicas/2024/
├── monitoreo/2024/
├── movilidad-urbana/2024/
├── gobierno-electronico/2024/
├── prensa-imagen/2024/
└── tecnologia/2024/

gamc-documents/
├── actas/2024/
├── informes/2024/
├── normativas/2024/
└── contratos/2024/

gamc-images/
├── logos/
├── avatars/
├── banners/
└── screenshots/

gamc-backups/
├── database/2024/
├── redis/2024/
├── files/2024/
└── system/2024/
```

## Conexión desde Aplicaciones

### Variables de Entorno
```bash
# Configuración para aplicaciones
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=gamc_backend
MINIO_SECRET_KEY=gamc_backend_secret_2024
MINIO_USE_SSL=false
MINIO_REGION=us-east-1
```

### Ejemplos de Uso

#### Node.js
```javascript
const Minio = require('minio');

const minioClient = new Minio.Client({
  endPoint: 'localhost',
  port: 9000,
  useSSL: false,
  accessKey: 'gamc_backend',
  secretKey: 'gamc_backend_secret_2024'
});

// Subir archivo
await minioClient.putObject('gamc-attachments', 'mensaje-123/documento.pdf', fileStream);

// Obtener URL de descarga
const url = await minioClient.presignedGetObject('gamc-attachments', 'mensaje-123/documento.pdf', 24*60*60);
```

#### Python
```python
from minio import Minio

client = Minio(
    'localhost:9000',
    access_key='gamc_backend',
    secret_key='gamc_backend_secret_2024',
    secure=False
)

# Subir archivo
client.fput_object('gamc-documents', 'informe-2024.pdf', '/path/to/file.pdf')

# Obtener archivo
client.fget_object('gamc-documents', 'informe-2024.pdf', '/path/to/download.pdf')
```

#### cURL (API S3)
```bash
# Configurar AWS CLI con credenciales de MinIO
aws configure set aws_access_key_id gamc_backend
aws configure set aws_secret_access_key gamc_backend_secret_2024
aws configure set default.region us-east-1
aws configure set default.s3.signature_version s3v4

# Usar con endpoint personalizado
aws --endpoint-url http://localhost:9000 s3 ls s3://gamc-attachments/
aws --endpoint-url http://localhost:9000 s3 cp file.pdf s3://gamc-documents/
```

## Usuarios y Permisos

### Usuarios del Sistema
| Usuario | Contraseña | Permisos | Uso |
|---------|------------|----------|-----|
| **gamc_admin** | gamc_minio_password_2024 | Administrador completo | Gestión del sistema |
| **gamc_backend** | gamc_backend_secret_2024 | Lectura/Escritura en buckets de aplicación | Backend del sistema |
| **gamc_readonly** | gamc_readonly_secret_2024 | Solo lectura | Consultas y reportes |

### Configuración de Políticas
```bash
# Política personalizada para backend
mc admin policy add gamc-local gamc-backend-policy /policies/backend-policy.json

# Asignar política a usuario
mc admin policy set gamc-local gamc-backend-policy user=gamc_backend
```

## Scripts de Administración

### 1. Monitoreo en Tiempo Real
```bash
# Monitor completo del sistema
./scripts/monitor.sh

# Monitor con intervalo personalizado
./scripts/monitor.sh -i 5
```

### 2. Backup de Datos
```bash
# Backup incremental (por defecto)
./scripts/backup.sh

# Backup completo
./scripts/backup.sh full

# Backup de bucket específico
./scripts/backup.sh incremental gamc-attachments
```

### 3. Configuración de Buckets
```bash
# Reconfigurar buckets
./scripts/setup-buckets.sh

# Verificar configuración
mc ls gamc-local
mc policy list gamc-local
```

## Configuración del Gateway Nginx

### Endpoints Disponibles
- **`/`** - Página de inicio del sistema GAMC
- **`/minio/`** - Proxy hacia API de MinIO
- **`/console/`** - Proxy hacia consola web de MinIO
- **`/public/`** - Acceso directo a buckets públicos
- **`/health`** - Health check del sistema

### Configuración de CORS
```nginx
# Headers para desarrollo
add_header 'Access-Control-Allow-Origin' '*' always;
add_header 'Access-Control-Allow-Methods' 'GET, POST, PUT, DELETE, OPTIONS' always;
add_header 'Access-Control-Allow-Headers' 'Authorization, Range, Content-Type' always;
```

## Casos de Uso Específicos del GAMC

### 1. Gestión de Archivos Adjuntos
```javascript
// Subir archivo adjunto con metadatos
const metadata = {
  'X-Amz-Meta-Unit': 'obras-publicas',
  'X-Amz-Meta-Message-Id': '123',
  'X-Amz-Meta-User': 'juan.mendoza'
};

await minioClient.putObject('gamc-attachments', 
  `obras-publicas/${new Date().getFullYear()}/mensaje-123-documento.pdf`, 
  fileStream, 
  metadata
);
```

### 2. Almacenamiento de Documentos Oficiales
```javascript
// Subir documento con URL pública
await minioClient.putObject('gamc-documents', 'normativas/2024/resolucion-001.pdf', fileStream);

// Obtener URL pública (solo lectura)
const publicUrl = `http://localhost:8090/public/gamc-documents/normativas/2024/resolucion-001.pdf`;
```

### 3. Gestión de Backups Automáticos
```bash
# Script de backup para cron
0 2 * * * /path/to/minio/scripts/backup.sh incremental >> /var/log/minio-backup.log 2>&1
0 2 * * 0 /path/to/minio/scripts/backup.sh full >> /var/log/minio-backup.log 2>&1
```

### 4. Limpieza Automática de Archivos Temporales
```javascript
// Configurar lifecycle para bucket temporal
const lifecycleConfig = {
  Rule: [{
    ID: 'temp-cleanup',
    Status: 'Enabled',
    Expiration: { Days: 7 }
  }]
};

await minioClient.setBucketLifecycle('gamc-temp', lifecycleConfig);
```

## Monitoreo y Métricas

### Métricas Clave a Monitorear
- **Espacio de almacenamiento usado/disponible**
- **Número de objetos por bucket**
- **Actividad de subida/descarga**
- **Tiempo de respuesta de API**
- **Errores de conexión**

### Dashboard de Monitoring
```bash
# Ver estadísticas en tiempo real
./scripts/monitor.sh

# Información del servidor
docker exec gamc_minio_client mc admin info gamc-local

# Uso por bucket
docker exec gamc_minio_client mc du gamc-local --recursive
```

## Backup y Restauración

### Backup Automático
```bash
# Configurar backup diario
crontab -e
0 2 * * * /path/to/minio/scripts/backup.sh incremental
0 2 * * 0 /path/to/minio/scripts/backup.sh full
```

### Restaurar desde Backup
```bash
# Restaurar bucket completo
mc mirror /path/to/backup/gamc-attachments/2024/12/15/full_20241215_020000 gamc-local/gamc-attachments

# Restaurar archivo específico
mc cp /path/to/backup/file.pdf gamc-local/gamc-documents/restored-file.pdf
```

### Sincronización con S3 Externo
```bash
# Configurar S3 externo en .env
EXTERNAL_S3_ENDPOINT=https://s3.amazonaws.com
EXTERNAL_S3_ACCESS_KEY=your_aws_access_key
EXTERNAL_S3_SECRET_KEY=your_aws_secret_key
EXTERNAL_S3_BUCKET=gamc-backup-external

# El backup automático sincronizará con S3 externo
```

## Troubleshooting

### Problema: MinIO no inicia
```bash
# Verificar logs
docker logs gamc_minio

# Verificar permisos de directorio de datos
ls -la data/

# Verificar configuración
docker exec gamc_minio cat /etc/passwd | grep minio
```

### Problema: No se pueden crear buckets
```bash
# Verificar credenciales
docker exec gamc_minio_client mc alias list

# Verificar conectividad
docker exec gamc_minio_client mc admin info gamc-local

# Recrear configuración
docker-compose -f docker-compose.minio.yml --profile setup up --force-recreate minio-client
```

### Problema: Errores de permisos
```bash
# Verificar políticas de usuario
docker exec gamc_minio_client mc admin policy list gamc-local

# Verificar permisos de bucket
docker exec gamc_minio_client mc policy get gamc-local/gamc-attachments
```

## Comandos Útiles

### Gestión de Buckets
```bash
# Listar todos los buckets
mc ls gamc-local

# Información detallada de bucket
mc du gamc-local/gamc-attachments

# Buscar archivos
mc find gamc-local/gamc-documents --name "*.pdf"

# Sincronizar directorios
mc mirror /local/path gamc-local/gamc-backups/manual
```

### Gestión de Usuarios
```bash
# Listar usuarios
mc admin user list gamc-local

# Crear usuario
mc admin user add gamc-local nuevo_usuario contraseña_segura

# Asignar política
mc admin policy set gamc-local readwrite user=nuevo_usuario
```

### Información del Sistema
```bash
# Estado del servidor
mc admin info gamc-local

# Configuración actual
mc admin config get gamc-local

# Logs del servidor
mc admin logs gamc-local
```

## Configuración para Producción

### 1. Seguridad
```bash
# Cambiar contraseñas por defecto
# Habilitar HTTPS/TLS
# Configurar firewall para puertos 9000-9001
# Usar políticas de acceso restrictivas
```

### 2. Alta Disponibilidad
```bash
# Configurar MinIO en modo distribuido
# Usar múltiples nodos
# Configurar load balancer
```

### 3. Monitoreo Avanzado
```bash
# Integrar con Prometheus/Grafana
# Configurar alertas automáticas
# Logs centralizados con ELK Stack
```

## Soporte
Para problemas técnicos relacionados con MinIO y almacenamiento, contactar al equipo de Tecnología del GAMC.
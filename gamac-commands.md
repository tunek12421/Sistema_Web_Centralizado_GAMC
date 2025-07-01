# GAMC Sistema Web Centralizado - Comandos de Docker

## Comandos Principales

### 🚀 Inicio Completo del Sistema
```bash
# Levantar todos los servicios principales
docker-compose -f docker-compose-unificado.yml up -d

# Verificar estado de todos los contenedores
docker-compose -f docker-compose-unificado.yml ps

# Ver logs en tiempo real
docker-compose -f docker-compose-unificado.yml logs -f
```

### 🔧 Servicios Individuales

#### Backend Go
```bash
# Solo el backend con sus dependencias
docker-compose -f docker-compose-unificado.yml up -d postgres redis gamc-backend-go

# Reconstruir backend después de cambios
docker-compose -f docker-compose-unificado.yml build gamc-backend-go
docker-compose -f docker-compose-unificado.yml up -d gamc-backend-go

# Ver logs del backend
docker-compose -f docker-compose-unificado.yml logs -f gamc-backend-go
```

#### Frontend
```bash
# Solo el frontend con backend
docker-compose -f docker-compose-unificado.yml up -d gamc-backend-go gamc-auth-frontend

# Reconstruir frontend
docker-compose -f docker-compose-unificado.yml build gamc-auth-frontend
docker-compose -f docker-compose-unificado.yml up -d gamc-auth-frontend

# Ver logs del frontend
docker-compose -f docker-compose-unificado.yml logs -f gamc-auth-frontend
```

### 📊 Servicios de Administración
```bash
# Levantar interfaces de administración
docker-compose -f docker-compose-unificado.yml --profile admin up -d

# Configurar MinIO (solo primera vez)
docker-compose -f docker-compose-unificado.yml --profile setup up minio-client

# Gateway Nginx (opcional)
docker-compose -f docker-compose-unificado.yml --profile gateway up -d
```

### 🛠️ Desarrollo

#### Hot Reload
```bash
# Para desarrollo con hot reload activado
# Asegúrate de que los volúmenes estén descomentados en el docker-compose

# Backend Go (requiere air o similar para hot reload)
docker-compose -f docker-compose-unificado.yml up -d postgres redis
# Ejecutar backend localmente para desarrollo:
cd gamc-backend-go && go run cmd/server/main.go

# Frontend (Vite hot reload automático)
docker-compose -f docker-compose-unificado.yml up -d gamc-auth-frontend
```

#### Reconstruir todo
```bash
# Reconstruir todas las imágenes
docker-compose -f docker-compose-unificado.yml build --no-cache

# Reconstruir y levantar
docker-compose -f docker-compose-unificado.yml up -d --build
```

### 🗄️ Gestión de Base de Datos

#### Conectar a PostgreSQL
```bash
# Desde el contenedor
docker exec -it gamc_postgres psql -U gamc_user -d gamc_system

# Desde host (si tienes psql instalado)
psql -h localhost -p 5432 -U gamc_user -d gamc_system
```

#### Backup y Restore
```bash
# Backup
docker exec gamc_postgres pg_dump -U gamc_user gamc_system > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore
cat backup_file.sql | docker exec -i gamc_postgres psql -U gamc_user -d gamc_system
```

### 🔴 Redis

#### Conectar a Redis
```bash
# CLI de Redis
docker exec -it gamc_redis redis-cli -a gamc_redis_password_2024

# Ver todas las claves
docker exec -it gamc_redis redis-cli -a gamc_redis_password_2024 KEYS "*"

# Limpiar cache
docker exec -it gamc_redis redis-cli -a gamc_redis_password_2024 FLUSHALL
```

### 📦 MinIO

#### Comandos de MinIO
```bash
# Listar buckets
docker exec gamc_minio_client mc ls gamc-local

# Subir archivo
docker exec gamc_minio_client mc cp /path/to/file gamc-local/gamc-images/

# Ver políticas
docker exec gamc_minio_client mc admin policy list gamc-local
```

## URLs de Acceso

### Aplicaciones Principales
- **Frontend Auth**: http://localhost:5173
- **Backend API**: http://localhost:3000
- **API Health**: http://localhost:3000/health

### Servicios de Infraestructura
- **MinIO Console**: http://localhost:9001 (admin: gamc_admin / gamc_minio_password_2024)
- **MinIO API**: http://localhost:9000

### Administración (con --profile admin)
- **PgAdmin**: http://localhost:8080 (admin@gamc.gov.bo / admin123)
- **Redis Commander**: http://localhost:8081 (admin / admin123)

### Gateway (con --profile gateway)
- **Nginx Gateway**: http://localhost:8090

## Solución de Problemas

### Verificar Estado de Servicios
```bash
# Estado de contenedores
docker-compose -f docker-compose-unificado.yml ps

# Health checks
docker inspect gamc_postgres | grep -A 10 '"Health"'
docker inspect gamc_redis | grep -A 10 '"Health"'
docker inspect gamc_minio | grep -A 10 '"Health"'

# Uso de recursos
docker stats gamc_postgres gamc_redis gamc_minio gamc_backend_go gamc_auth_frontend
```

### Limpiar Sistema
```bash
# Parar todos los servicios
docker-compose -f docker-compose-unificado.yml down

# Parar y eliminar volúmenes (¡CUIDADO! Borra datos)
docker-compose -f docker-compose-unificado.yml down -v

# Limpiar imágenes no utilizadas
docker image prune -f

# Limpiar todo el sistema Docker (¡EXTREMO CUIDADO!)
docker system prune -a -f
```

### Variables de Entorno para Desarrollo
```bash
# Crear archivo .env en el directorio raíz (opcional)
echo "COMPOSE_PROJECT_NAME=gamc" > .env
echo "POSTGRES_PASSWORD=gamc_password_2024" >> .env
echo "REDIS_PASSWORD=gamc_redis_password_2024" >> .env
echo "MINIO_PASSWORD=gamc_minio_password_2024" >> .env
```

## Perfiles de Docker Compose

El sistema incluye varios perfiles para diferentes entornos:

- **default**: Servicios principales (postgres, redis, minio, backend-go, frontend)
- **admin**: Herramientas de administración (pgadmin, redis-commander)
- **setup**: Configuración inicial (minio-client)
- **gateway**: Gateway nginx para desarrollo

### Usar múltiples perfiles
```bash
# Servicios principales + administración
docker-compose -f docker-compose-unificado.yml --profile admin up -d

# Todo el stack completo
docker-compose -f docker-compose-unificado.yml --profile admin --profile gateway --profile setup up -d
```
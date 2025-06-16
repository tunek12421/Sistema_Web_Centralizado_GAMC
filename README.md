# GAMC Sistema Web Centralizado - Configuración Unificada

## 📋 Descripción

Sistema web centralizado del Gobierno Autónomo Municipal de Cochabamba (GAMC) con todos los servicios unificados en un solo stack de Docker Compose.

## 🏗️ Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────────┐
│                    GAMC Sistema Web                     │
├─────────────────────────────────────────────────────────┤
│  Frontend (React + Vite)     │  Backend (Node.js + TS)  │
│  Puerto: 5173                │  Puerto: 3000            │
├─────────────────────────────────────────────────────────┤
│              Servicios de Infraestructura              │
├─────────────────────────────────────────────────────────┤
│  PostgreSQL    │    Redis      │      MinIO            │
│  Puerto: 5432  │  Puerto: 6379 │  Puertos: 9000/9001   │
├─────────────────────────────────────────────────────────┤
│              Herramientas de Administración            │
├─────────────────────────────────────────────────────────┤
│  PgAdmin       │ Redis Cmd     │   MinIO Gateway       │
│  Puerto: 8080  │ Puerto: 8081  │   Puerto: 8090        │
└─────────────────────────────────────────────────────────┘
```

## 🚀 Inicio Rápido

### 1. Preparación del Entorno

```bash
# Clonar o acceder al directorio del proyecto
cd Sistema_Web_Centralizado_GAMC

# Hacer ejecutable el script de gestión
chmod +x gamc.sh

# Configuración inicial
./gamc.sh setup
```

### 2. Iniciar Servicios

#### Opción A: Servicios Básicos (Recomendado para desarrollo)
```bash
./gamc.sh start
```

#### Opción B: Todos los Servicios (Incluye herramientas de admin)
```bash
./gamc.sh start-all
```

#### Opción C: Solo con herramientas de administración
```bash
./gamc.sh start-admin
```

### 3. Verificar Estado
```bash
./gamc.sh status
```

### 4. Ver URLs de Acceso
```bash
./gamc.sh urls
```

## 📚 Comandos Disponibles

| Comando | Descripción |
|---------|-------------|
| `./gamc.sh start` | Iniciar servicios básicos |
| `./gamc.sh start-all` | Iniciar todos los servicios |
| `./gamc.sh start-admin` | Iniciar con herramientas de admin |
| `./gamc.sh stop` | Detener todos los servicios |
| `./gamc.sh restart` | Reiniciar servicios |
| `./gamc.sh status` | Mostrar estado de servicios |
| `./gamc.sh logs [servicio]` | Mostrar logs |
| `./gamc.sh urls` | Mostrar URLs de acceso |
| `./gamc.sh backup` | Realizar backup de datos |
| `./gamc.sh update` | Actualizar servicios |
| `./gamc.sh cleanup` | Limpiar sistema completo |
| `./gamc.sh setup` | Configuración inicial |

## 🌐 URLs de Acceso

### Aplicaciones Principales
- **Frontend Auth**: http://localhost:5173
- **Backend Auth API**: http://localhost:3000/api/v1
- **API Docs**: http://localhost:3000/api/v1/docs

### Servicios de Infraestructura
- **MinIO Console**: http://localhost:9001
- **MinIO API**: http://localhost:9000

### Herramientas de Administración
- **PgAdmin**: http://localhost:8080
- **Redis Commander**: http://localhost:8081
- **MinIO Gateway**: http://localhost:8090

## 🔐 Credenciales por Defecto

### Base de Datos PostgreSQL
- **Usuario**: `gamc_user`
- **Contraseña**: `gamc_password_2024`
- **Base de datos**: `gamc_system`

### Redis
- **Contraseña**: `gamc_redis_password_2024`

### MinIO
- **Usuario**: `gamc_admin`
- **Contraseña**: `gamc_minio_password_2024`

### PgAdmin
- **Email**: `admin@gamc.gov.bo`
- **Contraseña**: `admin123`

### Redis Commander
- **Usuario**: `admin`
- **Contraseña**: `admin123`

## 📁 Estructura del Proyecto

```
Sistema_Web_Centralizado_GAMC/
├── docker-compose.yml              # Configuración unificada
├── gamc.sh                        # Script de gestión
├── .env                           # Variables globales
├── backend-auth/                  # Backend de autenticación
│   ├── Dockerfile
│   ├── package.json
│   ├── .env
│   └── src/
├── frontend-auth/                 # Frontend de autenticación
│   ├── Dockerfile
│   ├── package.json
│   ├── .env
│   └── src/
├── database/                      # Configuración PostgreSQL
│   ├── init/
│   ├── postgresql.conf
│   └── backups/
├── redis/                         # Configuración Redis
│   ├── redis.conf
│   └── logs/
└── minio/                         # Configuración MinIO
    ├── config/
    ├── nginx/
    └── scripts/
```

## 🔧 Configuración por Perfiles

El sistema utiliza perfiles de Docker Compose para organizar los servicios:

### Perfil `default` (Servicios Básicos)
- ✅ PostgreSQL
- ✅ Redis  
- ✅ MinIO
- ✅ Backend Auth
- ✅ Frontend Auth

### Perfil `admin` (Herramientas de Administración)
- ✅ PgAdmin
- ✅ Redis Commander

### Perfil `setup` (Configuración Inicial)
- ✅ MinIO Client (configuración automática)

### Perfil `gateway` (Gateway MinIO)
- ✅ Nginx Gateway para MinIO

### Perfil `monitor` (Monitoreo)
- ✅ Health Check Monitor

## 🛠️ Desarrollo

### Hot Reload
Los servicios están configurados para hot reload automático:
- **Frontend**: Vite con hot module replacement
- **Backend**: tsx watch para recarga automática

### Logs en Tiempo Real
```bash
# Ver logs de todos los servicios
./gamc.sh logs

# Ver logs de un servicio específico
./gamc.sh logs gamc-auth-backend
./gamc.sh logs gamc-auth-frontend
./gamc.sh logs postgres
```

### Depuración
```bash
# Verificar estado de la red
./gamc.sh network

# Verificar volúmenes
./gamc.sh volumes

# Ejecutar comandos dentro de contenedores
docker-compose exec postgres psql -U gamc_user gamc_system
docker-compose exec redis redis-cli
```

## 📊 Backup y Restauración

### Realizar Backup
```bash
./gamc.sh backup
```
Esto creará un backup completo en `./backups/YYYYMMDD_HHMMSS/`

### Backup Manual
```bash
# PostgreSQL
docker-compose exec postgres pg_dump -U gamc_user gamc_system > backup.sql

# Redis
docker-compose exec redis redis-cli BGSAVE

# MinIO
docker cp gamc_minio:/data ./minio_backup
```

## 🔄 Actualización de Servicios

```bash
# Actualizar todas las imágenes
./gamc.sh update

# Reconstruir servicios específicos
docker-compose build --no-cache gamc-auth-backend
docker-compose build --no-cache gamc-auth-frontend
```

## 🐛 Solución de Problemas

### Problema: Servicios no se conectan
```bash
# Verificar red
docker network ls | grep gamc
./gamc.sh network

# Recrear red si es necesario
docker network rm gamc_network
docker network create gamc_network
```

### Problema: Volúmenes no persisten datos
```bash
# Verificar volúmenes
./gamc.sh volumes

# Listar volúmenes
docker volume ls | grep gamc
```

### Problema: Puerto ya en uso
```bash
# Verificar puertos en uso
netstat -tulpn | grep :5173
netstat -tulpn | grep :3000

# Cambiar puertos en docker-compose.yml si es necesario
```

### Problema: Backend no se conecta a la base de datos
```bash
# Verificar logs del backend
./gamc.sh logs gamc-auth-backend

# Verificar conexión a PostgreSQL
docker-compose exec postgres pg_isready -U gamc_user

# Verificar variables de entorno
docker-compose exec gamc-auth-backend env | grep DATABASE
```

## 🔒 Seguridad

### Cambiar Contraseñas para Producción
1. Editar `.env` con contraseñas seguras
2. Actualizar archivos `.env` en `backend-auth/` y `frontend-auth/`
3. Regenerar secretos JWT
4. Reiniciar servicios

### Configuración de Red Segura
```bash
# Para producción, usar red personalizada
docker network create --driver bridge --subnet=172.20.0.0/16 gamc_network_prod
```

## 📈 Monitoreo

### Health Checks Automáticos
Todos los servicios incluyen health checks configurados:
- PostgreSQL: `pg_isready`
- Redis: `redis-cli ping`
- MinIO: endpoint `/minio/health/live`
- Backend: endpoint `/health`

### Monitoreo Manual
```bash
# Ver estado de salud
docker-compose ps

# Estadísticas de uso
docker stats

# Logs de sistema
docker-compose logs --tail=100 healthcheck-monitor
```

## 🤝 Contribución

1. Fork el proyecto
2. Crear rama para feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Crear Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver `LICENSE` para más detalles.

## 📞 Soporte

Para soporte técnico:
- **Email**: soporte.ti@gamc.gov.bo
- **Documentación**: [Wiki del proyecto]
- **Issues**: [GitHub Issues]

---

**Desarrollado por el Equipo de Tecnología del GAMC**
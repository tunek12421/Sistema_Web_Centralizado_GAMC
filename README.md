# Sistema Web Centralizado GAMC 🏛️

## Descripción General

El Sistema Web Centralizado del Gobierno Autónomo Municipal de Cochabamba (GAMC) es una plataforma integral que unifica todos los servicios municipales en un solo stack tecnológico. Incluye un sistema de mensajería interna para comunicación controlada entre unidades organizacionales, gestión de autenticación, y servicios de infraestructura completos.

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

### 3. Verificar Estado y URLs
```bash
./gamc.sh status
./gamc.sh urls
```

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

## 📚 Comandos de Gestión del Sistema

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
│       ├── controllers/
│       ├── services/
│       ├── repositories/
│       └── models/
├── frontend-auth/                 # Frontend de autenticación
│   ├── Dockerfile
│   ├── package.json
│   ├── .env
│   └── src/
│       ├── components/
│       ├── pages/
│       └── services/
├── database/                      # Configuración PostgreSQL
│   ├── init/
│   │   ├── 01-init.sql
│   │   └── 02-seed.sql
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

## 🔧 Perfiles de Docker Compose

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

### Perfil `monitor` (Monitoreo)
- ✅ Health Check Monitor

## 🛠️ Desarrollo y Contribución

### Instalación para Desarrollo
```bash
# Clonar repositorio
git clone [repository-url]
cd Sistema_Web_Centralizado_GAMC

# Configuración inicial
./gamc.sh setup

# Iniciar en modo desarrollo
./gamc.sh start

# El sistema estará disponible en:
# Frontend: http://localhost:5173
# Backend: http://localhost:3000/api/v1
```

### Hot Reload Automático
```bash
# Frontend: Vite con hot module replacement
# Backend: tsx watch para recarga automática

# Ver logs en tiempo real
./gamc.sh logs gamc-auth-backend
./gamc.sh logs gamc-auth-frontend
```

### Comandos de Depuración
```bash
# Verificar red de contenedores
docker network ls | grep gamc

# Verificar volúmenes
docker volume ls | grep gamc

# Ejecutar comandos dentro de contenedores
docker-compose exec postgres psql -U gamc_user gamc_system
docker-compose exec redis redis-cli

# Ver estadísticas de recursos
docker stats
```

### Proceso de Contribución
1. Fork el proyecto
2. Crear rama para feature (`git checkout -b feature/AmazingFeature`)
3. Commit cambios (`git commit -m 'Add AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Crear Pull Request

## 📊 Backup y Restauración

### Backup Automático
```bash
./gamc.sh backup
# Crear backup completo en ./backups/YYYYMMDD_HHMMSS/
```

### Backup Manual por Servicio
```bash
# PostgreSQL
docker-compose exec postgres pg_dump -U gamc_user gamc_system > backup.sql

# Redis
docker-compose exec redis redis-cli BGSAVE

# MinIO
docker cp gamc_minio:/data ./minio_backup
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

## 🔒 Configuración de Seguridad

### Variables de Entorno para Producción
```bash
# Editar .env con contraseñas seguras
DATABASE_PASSWORD=your_secure_password
REDIS_PASSWORD=your_redis_password
MINIO_ROOT_PASSWORD=your_minio_password
JWT_SECRET=your_jwt_secret_key
```

### Configuración de Red Segura
```bash
# Para producción, usar red personalizada
docker network create --driver bridge --subnet=172.20.0.0/16 gamc_network_prod
```

### Health Checks Configurados
- PostgreSQL: `pg_isready`
- Redis: `redis-cli ping`
- MinIO: endpoint `/minio/health/live`
- Backend: endpoint `/health`

## 📈 Monitoreo del Sistema

### Health Checks Automáticos
```bash
# Ver estado de salud
docker-compose ps

# Estadísticas de uso
docker stats

# Logs de sistema
docker-compose logs --tail=100 healthcheck-monitor
```

### Métricas Disponibles
- Total de mensajes por unidad
- Tiempo promedio de respuesta
- Mensajes urgentes vs regulares
- Distribución por tipos de mensaje
- Usuarios activos por unidad
- Estado de servicios de infraestructura

---

# 📨 Sistema de Mensajería Interna

La funcionalidad principal del sistema incluye un módulo de mensajería interna para comunicación controlada entre unidades organizacionales.

## 🔐 Control de Acceso y Permisos

### Roles de Usuario

El sistema implementa un control de acceso basado en roles que determina quién puede enviar mensajes:

#### 👥 Rol "INPUT" (Emisores)
- **Permisos**: Crear y enviar mensajes
- **Función**: Personal autorizado para generar comunicaciones oficiales
- **Restricción**: Solo estos usuarios pueden iniciar comunicaciones

#### 👁️ Rol "OUTPUT" (Receptores)
- **Permisos**: Recibir y leer mensajes
- **Función**: Personal que recibe y procesa comunicaciones
- **Restricción**: No pueden crear nuevos mensajes

### ⚠️ Restricción Principal
**Solo usuarios con rol "INPUT" pueden enviar mensajes a otras unidades**

## 📊 Matriz de Comunicación: Quién Envía a Quién

### Usuarios INPUT Autorizados (Emisores)

| Usuario Emisor | Unidad de Origen | Email | Puede Enviar A |
|---|---|---|---|
| `obras.input` | **Obras Públicas** | obras.input@gamc.gov.bo | Todas las unidades |
| `monitoreo.input` | **Monitoreo** | monitoreo.input@gamc.gov.bo | Todas las unidades |
| `movilidad.input` | **Movilidad Urbana** | movilidad.input@gamc.gov.bo | Todas las unidades |
| `gobierno.input` | **Gobierno Electrónico** | gobierno.input@gamc.gov.bo | Todas las unidades |
| `prensa.input` | **Prensa e Imagen** | prensa.input@gamc.gov.bo | Todas las unidades |
| `tecnologia.input` | **Tecnología** | tecnologia.input@gamc.gov.bo | Todas las unidades |

### Usuarios OUTPUT (Solo Receptores)

| Usuario Receptor | Unidad | Email | Solo Recibe |
|---|---|---|---|
| `obras.output` | **Obras Públicas** | obras.output@gamc.gov.bo | ✅ Solo lectura |
| `monitoreo.output` | **Monitoreo** | monitoreo.output@gamc.gov.bo | ✅ Solo lectura |
| `movilidad.output` | **Movilidad Urbana** | movilidad.output@gamc.gov.bo | ✅ Solo lectura |
| `gobierno.output` | **Gobierno Electrónico** | gobierno.output@gamc.gov.bo | ✅ Solo lectura |
| `prensa.output` | **Prensa e Imagen** | prensa.output@gamc.gov.bo | ✅ Solo lectura |
| `tecnologia.output` | **Tecnología** | tecnologia.output@gamc.gov.bo | ✅ Solo lectura |

### 🔄 Flujos de Comunicación Típicos

#### 1. **Obras Públicas** → Otras Unidades
```
obras.input@gamc.gov.bo PUEDE ENVIAR A:
├── Movilidad Urbana (coordinación vial) ✅
├── Finanzas (presupuestos) ✅
├── Tecnología (sistemas) ✅
├── Monitoreo (seguimiento) ✅
├── Prensa (comunicación) ✅
└── Gobierno Electrónico (digitales) ✅
```

#### 2. **Monitoreo** → Otras Unidades
```
monitoreo.input@gamc.gov.bo PUEDE ENVIAR A:
├── Tecnología (accesos/sistemas) ✅
├── Obras Públicas (seguimiento) ✅
├── Finanzas (reportes) ✅
├── Prensa (informes) ✅
├── Movilidad Urbana (evaluación) ✅
└── Gobierno Electrónico (datos) ✅
```

#### 3. **Tecnología** → Otras Unidades
```
tecnologia.input@gamc.gov.bo PUEDE ENVIAR A:
├── Todas las unidades (mantenimientos) ✅
├── Gobierno Electrónico (sistemas) ✅
├── Monitoreo (reportes técnicos) ✅
├── Obras Públicas (soporte) ✅
├── Prensa (actualizaciones web) ✅
└── Movilidad Urbana (sistemas) ✅
```

#### 4. **Prensa e Imagen** → Otras Unidades
```
prensa.input@gamc.gov.bo PUEDE ENVIAR A:
├── Gobierno Electrónico (web/urgente) ✅
├── Todas las unidades (comunicados) ✅
├── Monitoreo (información) ✅
├── Obras Públicas (difusión) ✅
├── Tecnología (contenido) ✅
└── Movilidad Urbana (campañas) ✅
```

### 🚫 Restricciones Importantes

1. **❌ Usuarios OUTPUT NO pueden enviar**: Solo pueden leer y gestionar mensajes recibidos
2. **❌ Sin restricciones por unidad**: Cualquier usuario INPUT puede enviar a cualquier unidad
3. **❌ No hay envío masivo**: Un mensaje = una unidad destinataria
4. **❌ No hay autoenvío**: No se pueden enviar mensajes a la propia unidad

## 🏢 Unidades Organizacionales

El sistema está estructurado en las siguientes unidades:

| Código | Unidad | Descripción |
|--------|--------|-------------|
| `OBRAS_PUBLICAS` | Obras Públicas | Gestión de infraestructura y construcción |
| `FINANZAS` | Finanzas | Administración financiera y presupuestaria |
| `MOVILIDAD_URBANA` | Movilidad Urbana | Gestión del tráfico y transporte |
| `MONITOREO` | Monitoreo | Seguimiento y evaluación de proyectos |
| `GOBIERNO_ELECTRONICO` | Gobierno Electrónico | Tecnologías de la información |
| `PRENSA_IMAGEN` | Prensa e Imagen | Comunicación institucional |
| `TECNOLOGIA` | Tecnología | Soporte técnico y sistemas |

## 📋 Tipos de Mensajes

### Categorías Disponibles
- **Coordinación**: Mensajes para coordinar actividades entre unidades
- **Solicitudes**: Peticiones formales de información o servicios
- **Informes**: Comunicación de resultados o estados
- **Urgentes**: Comunicaciones que requieren atención inmediata

### Niveles de Prioridad
1. **Baja** - Información general
2. **Media** - Asuntos regulares
3. **Alta** - Asuntos importantes
4. **Crítica** - Emergencias

## 🔄 Flujo de Mensajería

### Proceso de Envío
1. **Creación**: Usuario INPUT crea mensaje
2. **Destinación**: Selecciona unidad receptora (`receiverUnitId`)
3. **Clasificación**: Asigna tipo y prioridad
4. **Envío**: Sistema procesa y distribuye
5. **Notificación**: Se generan notificaciones automáticas para la unidad receptora

### Estados del Mensaje
- `DRAFT` - Borrador
- `SENT` - Enviado
- `DELIVERED` - Entregado
- `READ` - Leído
- `ARCHIVED` - Archivado

## 💡 Ejemplos Prácticos de Comunicación

### Ejemplo 1: Obras Públicas → Movilidad Urbana
**Escenario**: Coordinación para trabajos viales
```json
{
  "subject": "Coordinación para obras en Av. Ballivián",
  "content": "Estimados colegas de Movilidad Urbana,\n\nLes informamos que el próximo lunes 19 de junio iniciaremos trabajos de repavimentación en la Av. Ballivián entre calles 25 de Mayo y Heroínas.\n\nSolicitamos coordinar el desvío de tráfico y señalización correspondiente.",
  "receiverUnitId": 4, // MOVILIDAD_URBANA
  "messageTypeId": 1, // COORDINACION
  "priorityLevel": 2,
  "isUrgent": false
}
```
**✅ Usuario autorizado**: `obras.input@gamc.gov.bo`

### Ejemplo 2: Monitoreo → Tecnología
**Escenario**: Solicitud de acceso a sistemas
```json
{
  "subject": "Solicitud de acceso a sistema de seguimiento",
  "content": "Estimado equipo de Tecnología,\n\nNecesitamos acceso al nuevo módulo de seguimiento de proyectos para 5 usuarios adicionales de nuestra unidad.\n\nAdjunto la lista de usuarios que requieren acceso.",
  "receiverUnitId": 7, // TECNOLOGIA
  "messageTypeId": 2, // SOLICITUD
  "priorityLevel": 3,
  "isUrgent": false
}
```
**✅ Usuario autorizado**: `monitoreo.input@gamc.gov.bo`

### Ejemplo 3: Prensa → Gobierno Electrónico
**Escenario**: Actualización urgente de contenido web
```json
{
  "subject": "URGENTE: Actualización página web municipal",
  "content": "Estimados colegas,\n\nNecesitamos actualizar URGENTEMENTE la información en la página web sobre el nuevo horario de atención al público.\n\nLa información debe estar publicada antes de las 14:00 de hoy.",
  "receiverUnitId": 5, // GOBIERNO_ELECTRONICO
  "messageTypeId": 4, // URGENTE
  "priorityLevel": 1,
  "isUrgent": true
}
```
**✅ Usuario autorizado**: `prensa.input@gamc.gov.bo`

### Ejemplo 4: Tecnología → Administración
**Escenario**: Notificación de mantenimiento
```json
{
  "subject": "Mantenimiento programado del sistema",
  "content": "Estimados usuarios,\n\nLes informamos que el sábado 24 de junio de 02:00 a 06:00 AM se realizará mantenimiento programado del sistema.\n\nDurante este periodo el sistema no estará disponible.",
  "receiverUnitId": 1, // ADMINISTRACION
  "messageTypeId": 3, // NOTIFICACION
  "priorityLevel": 4,
  "isUrgent": false
}
```
**✅ Usuario autorizado**: `tecnologia.input@gamc.gov.bo`

### ❌ Ejemplo de Intento No Autorizado
```json
// ❌ ESTO FALLARÁ - Usuario OUTPUT intentando enviar
{
  "from": "obras.output@gamc.gov.bo", // ❌ ROL OUTPUT
  "subject": "Este mensaje será rechazado",
  "receiverUnitId": 2,
  "messageTypeId": 1
}
// Resultado: HTTP 403 Forbidden
```

## 🛡️ Seguridad y Validaciones

### Control de Usuarios
- Autenticación mediante tokens JWT
- Validación de permisos por rol
- Verificación de pertenencia a unidad organizacional

### Validaciones de Mensajes
- Usuario emisor debe tener rol "INPUT"
- Unidad receptora debe existir y estar activa
- Contenido debe cumplir políticas de la institución
- Rate limiting para prevenir spam

### Auditoría
- Registro completo de todas las comunicaciones
- Trazabilidad de envíos y recepciones
- Logs de acceso y operaciones

## 🚀 API Endpoints

### Crear Mensaje
```http
POST /api/v1/messages
Authorization: Bearer {token}
Content-Type: application/json

{
  "subject": "Asunto del mensaje",
  "content": "Contenido del mensaje",
  "receiverUnitId": 2,
  "messageTypeId": 1,
  "priorityLevel": 2,
  "isUrgent": false
}
```

### Listar Mensajes
```http
GET /api/v1/messages?page=1&limit=10&sortBy=created_at&sortOrder=desc
Authorization: Bearer {token}
```

### Obtener Mensaje Específico
```http
GET /api/v1/messages/{messageId}
Authorization: Bearer {token}
```

### Marcar como Leído
```http
PATCH /api/v1/messages/{messageId}/read
Authorization: Bearer {token}
```

## 🔧 Configuración

### Variables de Entorno
```env
# Base de datos
DB_HOST=localhost
DB_PORT=5432
DB_NAME=gamc_messaging
DB_USER=postgres
DB_PASSWORD=password

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRES_IN=24h

# Rate Limiting
RATE_LIMIT_WINDOW=300000  # 5 minutos
RATE_LIMIT_MAX_REQUESTS=5
```

### Dependencias Principales
- **Backend**: Go con Gin framework
- **Base de Datos**: PostgreSQL
- **Autenticación**: JWT tokens
- **Frontend**: React con TypeScript

## 📊 Estadísticas del Sistema

### Métricas Disponibles
- Total de mensajes por unidad
- Tiempo promedio de respuesta
- Mensajes urgentes vs regulares
- Distribución por tipos de mensaje

### Reportes
- Comunicaciones por período
- Eficiencia por unidad
- Mensajes sin leer
- Tendencias de comunicación

## 🚨 Limitaciones y Restricciones Específicas

### 🔒 Restricciones de Envío por Usuario

| Tipo de Usuario | Puede Enviar | Puede Recibir | Restricciones |
|---|---|---|---|
| **INPUT** | ✅ Sí | ✅ Sí | Solo a unidades diferentes a la suya |
| **OUTPUT** | ❌ No | ✅ Sí | Solo puede leer y gestionar |
| **ADMIN** | ✅ Sí | ✅ Sí | Sin restricciones |

### 📋 Matriz Detallada de Permisos

#### Usuarios que SÍ pueden enviar:
- ✅ `obras.input@gamc.gov.bo` (Obras Públicas)
- ✅ `monitoreo.input@gamc.gov.bo` (Monitoreo) 
- ✅ `movilidad.input@gamc.gov.bo` (Movilidad Urbana)
- ✅ `gobierno.input@gamc.gov.bo` (Gobierno Electrónico)
- ✅ `prensa.input@gamc.gov.bo` (Prensa e Imagen)
- ✅ `tecnologia.input@gamc.gov.bo` (Tecnología)

#### Usuarios que NO pueden enviar:
- ❌ `obras.output@gamc.gov.bo` (Solo recepción)
- ❌ `monitoreo.output@gamc.gov.bo` (Solo recepción)
- ❌ `movilidad.output@gamc.gov.bo` (Solo recepción)
- ❌ `gobierno.output@gamc.gov.bo` (Solo recepción)
- ❌ `prensa.output@gamc.gov.bo` (Solo recepción)
- ❌ `tecnologia.output@gamc.gov.bo` (Solo recepción)

### 🚫 Limitaciones Técnicas

1. **Envío único**: Un mensaje = una unidad destinataria (no broadcast)
2. **Sin autoenvío**: No se puede enviar a la propia unidad
3. **Archivos adjuntos**: Actualmente no soportado (en desarrollo)
4. **Sin respuesta directa**: No hay sistema de threading automático
5. **Rate limiting**: Máximo 5 mensajes por usuario cada 5 minutos
6. **Validación de dominio**: Solo emails @gamc.gov.bo autorizados

### ⚡ Validaciones del Sistema

#### En tiempo de envío:
```javascript
// Validaciones automáticas
if (user.role !== 'input' && user.role !== 'admin') {
  throw new Error('Usuario no autorizado para enviar mensajes');
}

if (receiverUnitId === user.organizationalUnitId) {
  throw new Error('No se puede enviar mensajes a la propia unidad');
}

if (messageContent.length > 2000) {
  throw new Error('Contenido excede límite de caracteres');
}
```

#### Códigos de respuesta HTTP:
- `403 Forbidden`: Usuario OUTPUT intentando enviar
- `401 Unauthorized`: Token inválido o faltante  
- `400 Bad Request`: Datos de mensaje inválidos
- `429 Too Many Requests`: Rate limit excedido
- `201 Created`: Mensaje enviado exitosamente

## 🛠️ Desarrollo y Contribución

### Instalación Local
```bash
# Backend
cd gamc-backend-go
go mod download
go run main.go

# Frontend
cd frontend-auth
npm install
npm start
```

### Testing
```bash
# Ejecutar tests
go test ./...
npm test
```

## 📞 Soporte

Para soporte técnico o consultas sobre el sistema:
- **Email**: soporte.ti@gamc.gov.bo
- **Documentación**: [Link a documentación técnica]
- **Issues**: Utilizar el sistema de issues de GitHub

## 📞 Soporte y Contacto

### Soporte Técnico
- **Email**: soporte.ti@gamc.gov.bo
- **Documentación**: [Wiki del proyecto]
- **Issues**: Utilizar el sistema de issues de GitHub

### Equipo de Desarrollo
- **Desarrollado por**: Equipo de Tecnología del GAMC
- **Mantenido por**: Unidad de Gobierno Electrónico

### Recursos Adicionales
- **API Documentation**: http://localhost:3000/api/v1/docs
- **Database Admin**: http://localhost:8080 (PgAdmin)
- **Cache Management**: http://localhost:8081 (Redis Commander)
- **File Storage**: http://localhost:9001 (MinIO Console)

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver `LICENSE` para más detalles.

```
Copyright (c) 2024 Gobierno Autónomo Municipal de Cochabamba (GAMC)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.
```

## 🔄 Actualizaciones del Sistema

### Historial de Versiones
- **v1.0.0**: Sistema base de autenticación y mensajería
- **v1.1.0**: Integración Docker Compose completa
- **v1.2.0**: Herramientas de administración y monitoreo
- **v1.3.0**: Sistema de backup automático

### Roadmap
- [ ] Sistema de archivos adjuntos
- [ ] Notificaciones push en tiempo real
- [ ] Dashboard de analytics avanzado
- [ ] Integración con Active Directory
- [ ] Aplicación móvil (Android/iOS)
- [ ] API para integraciones externas

## 🚀 Producción

### Consideraciones para Deploy
1. **Cambiar todas las contraseñas por defecto**
2. **Configurar SSL/TLS para HTTPS**
3. **Implementar firewall y VPN**
4. **Configurar backup automático diario**
5. **Establecer monitoreo 24/7**
6. **Documentar procedimientos de recuperación**

### Variables Críticas de Producción
```env
# Seguridad
JWT_SECRET=your-super-secure-jwt-secret-key
DATABASE_PASSWORD=your-secure-database-password
REDIS_PASSWORD=your-secure-redis-password

# Configuración
NODE_ENV=production
CORS_ORIGIN=https://gamc.gov.bo
RATE_LIMIT_WINDOW=300000
RATE_LIMIT_MAX_REQUESTS=100

# SSL/TLS
SSL_CERT_PATH=/path/to/certificate.pem
SSL_KEY_PATH=/path/to/private-key.pem
```

---

**Desarrollado con ❤️ por el Equipo de Tecnología del GAMC**

**Sistema Web Centralizado - Versión 1.3.0**

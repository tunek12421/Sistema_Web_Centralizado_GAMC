# Sistema de Mensajería GAMC 📨

## Descripción General

El Sistema de Mensajería del Gobierno Autónomo Municipal de Cochabamba (GAMC) es una plataforma de comunicación interna que permite el intercambio controlado de mensajes entre diferentes unidades organizacionales del municipio.

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

---

**Nota**: Este sistema está diseñado para uso interno del GAMC y sigue las políticas de comunicación institucional establecidas por la administración municipal.
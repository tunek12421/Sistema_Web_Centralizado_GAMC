# Redis - Sistema Web Centralizado GAMC

## Descripción
Configuración optimizada de Redis para el Sistema Web Centralizado del GAMC, incluyendo cache de aplicación, gestión de sesiones, notificaciones en tiempo real y almacenamiento temporal de datos.

## Características
- ✅ Redis 7 Alpine con configuración optimizada
- ✅ Persistencia RDB + AOF habilitada
- ✅ Múltiples bases de datos organizadas por propósito
- ✅ Redis Commander para administración web
- ✅ Redis Sentinel para alta disponibilidad (producción)
- ✅ Scripts de monitoreo y backup automatizados
- ✅ Configuración de seguridad robusta

## Inicio Rápido

### 1. Levantar Redis
```bash
# Desde el directorio redis/
docker-compose -f docker-compose.redis.yml up -d redis
```

### 2. Verificar Estado
```bash
# Ver logs
docker-compose -f docker-compose.redis.yml logs -f redis

# Verificar conexión
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 ping
```

### 3. Acceder a Redis Commander (Opcional)
```bash
# Levantar interfaz web de administración
docker-compose -f docker-compose.redis.yml --profile admin up -d

# Acceder en: http://localhost:8081
# Usuario: admin | Contraseña: admin123
```

## Estructura de Bases de Datos

Redis está configurado con 6 bases de datos especializadas:

| DB | Propósito | Descripción |
|----|-----------|-------------|
| 0 | **Sesiones** | Sesiones de usuario, tokens JWT activos |
| 1 | **Cache** | Cache de aplicación, datos temporales |
| 2 | **Notificaciones** | Cola de notificaciones en tiempo real |
| 3 | **Métricas** | Estadísticas temporales, contadores |
| 4 | **Query Cache** | Cache de consultas de base de datos |
| 5 | **JWT Blacklist** | Tokens JWT revocados/invalidados |

## Conexión desde Aplicaciones

### Variables de Entorno
```bash
# Configuración básica
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=gamc_redis_password_2024

# URLs específicas por base de datos
REDIS_URL_SESSIONS=redis://:gamc_redis_password_2024@localhost:6379/0
REDIS_URL_CACHE=redis://:gamc_redis_password_2024@localhost:6379/1
REDIS_URL_NOTIFICATIONS=redis://:gamc_redis_password_2024@localhost:6379/2
```

### Ejemplos de Uso

#### Node.js
```javascript
const redis = require('redis');

// Cliente para sesiones
const sessionClient = redis.createClient({
  url: 'redis://:gamc_redis_password_2024@localhost:6379/0'
});

// Cliente para cache
const cacheClient = redis.createClient({
  url: 'redis://:gamc_redis_password_2024@localhost:6379/1'
});

await sessionClient.connect();
await cacheClient.connect();
```

#### Python
```python
import redis

# Cliente para sesiones
session_redis = redis.Redis(
    host='localhost',
    port=6379,
    password='gamc_redis_password_2024',
    db=0,
    decode_responses=True
)

# Cliente para cache
cache_redis = redis.Redis(
    host='localhost',
    port=6379,
    password='gamc_redis_password_2024',
    db=1,
    decode_responses=True
)
```

## Scripts de Administración

### 1. Verificación de Salud
```bash
# Ejecutar verificación completa
./scripts/health-check.sh

# Verificar memoria, conexiones, persistencia, etc.
```

### 2. Backup de Datos
```bash
# Backup completo
./scripts/backup.sh

# Backup con configuración personalizada
BACKUP_DIR=/path/to/backups RETENTION_DAYS=30 ./scripts/backup.sh
```

### 3. Monitoreo en Tiempo Real
```bash
# Monitor con actualización cada 5 segundos
./scripts/monitor.sh

# Monitor con intervalo personalizado
./scripts/monitor.sh -i 10
```

## Configuración de Producción

### 1. Alta Disponibilidad con Sentinel
```bash
# Levantar con Sentinel para failover automático
docker-compose -f docker-compose.redis.yml --profile production up -d
```

### 2. Configuración de Seguridad
```bash
# Cambiar contraseñas por defecto
sed -i 's/gamc_redis_password_2024/nueva_contraseña_segura/g' .env redis.conf

# Restringir comandos peligrosos (ya configurado)
# FLUSHDB, FLUSHALL, EVAL, DEBUG están deshabilitados
```

### 3. Optimización de Memoria
```bash
# Ajustar límite de memoria según servidor
# En redis.conf: maxmemory 1gb

# Monitorear uso de memoria
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 info memory
```

## Casos de Uso Específicos del GAMC

### 1. Gestión de Sesiones
```javascript
// Guardar sesión de usuario
await sessionClient.setEx(`session:${userId}`, 3600, JSON.stringify(sessionData));

// Verificar sesión activa
const session = await sessionClient.get(`session:${userId}`);
```

### 2. Cache de Mensajes
```javascript
// Cache de lista de mensajes por unidad
await cacheClient.setEx(`messages:unit:${unitId}`, 300, JSON.stringify(messages));

// Cache de estadísticas del dashboard
await cacheClient.setEx('dashboard:stats', 60, JSON.stringify(stats));
```

### 3. Notificaciones en Tiempo Real
```javascript
// Publicar notificación
await notificationClient.publish('notifications:new_message', JSON.stringify({
  unitId: 'obras_publicas',
  messageId: 123,
  type: 'urgent'
}));

// Suscribirse a notificaciones
await notificationClient.subscribe('notifications:new_message');
```

### 4. Blacklist de JWT
```javascript
// Invalidar token JWT
await jwtClient.setEx(`blacklist:${tokenJti}`, tokenExp - Date.now(), 'revoked');

// Verificar si token está en blacklist
const isBlacklisted = await jwtClient.exists(`blacklist:${tokenJti}`);
```

## Monitoreo y Métricas

### Métricas Clave a Monitorear
- **Memoria usada vs máxima**
- **Número de clientes conectados**
- **Hit ratio del cache**
- **Consultas lentas (slowlog)**
- **Persistencia (último save)**

### Alertas Recomendadas
```bash
# Uso de memoria > 80%
# Clientes conectados > 80% del máximo
# Hit ratio < 70%
# Consultas lentas > 10 en slowlog
# Último save > 1 hora
```

## Backup y Restauración

### Backup Automático
```bash
# Configurar cron para backup diario
0 2 * * * /path/to/redis/scripts/backup.sh

# Backup se guarda en redis/backups/ con retención de 7 días
```

### Restaurar desde Backup
```bash
# Parar Redis
docker-compose -f docker-compose.redis.yml down

# Restaurar archivo RDB
docker cp backup_file.rdb gamc_redis:/data/gamc_dump.rdb

# Restaurar archivo AOF (si existe)
docker cp backup_file.aof gamc_redis:/data/gamc_appendonly.aof

# Reiniciar Redis
docker-compose -f docker-compose.redis.yml up -d redis
```

## Troubleshooting

### Problema: Redis no inicia
```bash
# Verificar logs
docker logs gamc_redis

# Verificar configuración
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 config get "*"

# Verificar permisos de archivos
ls -la data/
```

### Problema: Memoria insuficiente
```bash
# Ver uso actual
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 info memory

# Limpiar cache si es necesario
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 -n 1 flushdb

# Ajustar maxmemory en redis.conf
```

### Problema: Conexiones rechazadas
```bash
# Verificar número de clientes
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 info clients

# Verificar configuración de red
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 config get bind
```

## Comandos Útiles

### Información del Sistema
```bash
# Información general
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 info

# Ver todas las llaves en una DB
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 -n 0 keys "*"

# Tamaño de cada DB
for db in {0..5}; do
  echo "DB$db: $(docker exec gamc_redis redis-cli -a gamc_redis_password_2024 -n $db dbsize) llaves"
done
```

### Limpieza y Mantenimiento
```bash
# Limpiar DB específica (CUIDADO en producción)
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 -n 1 flushdb

# Forzar save manual
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 bgsave

# Ver estadísticas de comandos
docker exec gamc_redis redis-cli -a gamc_redis_password_2024 info commandstats
```

## Configuración para Desarrollo vs Producción

### Desarrollo
- Persistencia básica (RDB cada 15 min)
- Memoria limitada (512MB)
- Logs de nivel notice
- Sin autenticación adicional

### Producción
- Persistencia completa (RDB + AOF)
- Memoria optimizada (según servidor)
- Logs detallados
- Sentinel para alta disponibilidad
- Monitoreo automático
- Backups diarios

## Soporte
Para problemas técnicos relacionados con Redis, contactar al equipo de Tecnología del GAMC.
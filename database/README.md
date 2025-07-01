# Base de Datos - Sistema Web Centralizado GAMC

## Descripción
Base de datos PostgreSQL optimizada para el Sistema Web Centralizado del GAMC con esquema completo, datos iniciales y configuración de producción.

## Características
- ✅ PostgreSQL 15 con configuración optimizada
- ✅ Esquema completo con relaciones
- ✅ Data Warehouse con esquema estrella
- ✅ Datos iniciales (seed data)
- ✅ Funciones y triggers automáticos
- ✅ Índices optimizados para consultas
- ✅ PgAdmin para administración
- ✅ Configuración de timezone Bolivia

## Inicio Rápido

### 1. Levantar la Base de Datos
```bash
# Desde el directorio database/
docker-compose -f docker-compose.database.yml up -d
```

### 2. Verificar Estado
```bash
# Ver logs
docker-compose -f docker-compose.database.yml logs -f postgres

# Verificar salud del contenedor
docker-compose -f docker-compose.database.yml ps
```

### 3. Acceder a PgAdmin (Opcional)
```bash
# Levantar PgAdmin
docker-compose -f docker-compose.database.yml --profile admin up -d

# Acceder en: http://localhost:8080
# Email: admin@gamc.gov.bo
# Password: admin123
```

## Estructura de la Base de Datos

### Tablas Principales
- `organizational_units` - Unidades organizacionales del GAMC
- `users` - Usuarios del sistema con roles
- `messages` - Mensajes entre unidades
- `message_attachments` - Archivos adjuntos
- `audit_logs` - Logs de auditoría

### Data Warehouse
- `fact_messages` - Tabla de hechos para análisis
- `dim_date` - Dimensión fecha
- `dim_time` - Dimensión tiempo

### Roles de Usuario
- **admin**: Acceso completo al sistema
- **input**: Puede crear y enviar mensajes
- **output**: Puede ver y responder mensajes

## Usuarios por Defecto

| Usuario | Contraseña | Rol | Unidad |
|---------|------------|-----|--------|
| admin | admin123 | admin | Administración |
| obras.input | admin123 | input | Obras Públicas |
| obras.output | admin123 | output | Obras Públicas |
| monitoreo.input | admin123 | input | Monitoreo |
| movilidad.input | admin123 | input | Movilidad Urbana |

> **Importante**: Cambiar las contraseñas por defecto en producción

## Conexión desde Aplicaciones

```bash
# URL de conexión
DATABASE_URL=postgresql://gamc_user:gamc_password_2024@localhost:5432/gamc_system
```

## Funciones Disponibles

### `get_dashboard_stats()`
Obtiene estadísticas para el dashboard principal
```sql
SELECT * FROM get_dashboard_stats();
```

### `get_messages_by_unit(unit_id, limit)`
Obtiene mensajes por unidad organizacional
```sql
SELECT * FROM get_messages_by_unit(1, 50);
```

### `mark_message_as_read(message_id, user_id)`
Marca un mensaje como leído
```sql
SELECT mark_message_as_read(1, 'user-uuid');
```

## Backup y Restauración

### Crear Backup
```bash
# Backup completo
docker exec gamc_postgres pg_dump -U gamc_user gamc_system > backup_$(date +%Y%m%d_%H%M%S).sql

# Solo esquema
docker exec gamc_postgres pg_dump -U gamc_user -s gamc_system > schema_backup.sql

# Solo datos
docker exec gamc_postgres pg_dump -U gamc_user -a gamc_system > data_backup.sql
```

### Restaurar Backup
```bash
# Restaurar desde backup
docker exec -i gamc_postgres psql -U gamc_user gamc_system < backup_file.sql
```

## Monitoreo

### Ver Conexiones Activas
```sql
SELECT pid, usename, application_name, client_addr, state, query_start 
FROM pg_stat_activity 
WHERE state = 'active';
```

### Estadísticas de Tablas
```sql
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del, n_live_tup
FROM pg_stat_user_tables 
ORDER BY n_live_tup DESC;
```

### Consultas Lentas
```sql
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;
```

## Mantenimiento

### Actualizar Estadísticas
```sql
ANALYZE;
```

### Reindexar Tablas
```sql
REINDEX DATABASE gamc_system;
```

### Limpiar Logs Antiguos
```bash
# Ejecutar cada semana
docker exec gamc_postgres find /var/lib/postgresql/data/pg_log -name "*.log" -mtime +7 -delete
```

## Troubleshooting

### Problema: No se puede conectar
```bash
# Verificar que el contenedor esté corriendo
docker ps | grep gamc_postgres

# Ver logs detallados
docker logs gamc_postgres

# Verificar configuración de red
docker network ls | grep gamc
```

### Problema: Rendimiento lento
```sql
-- Verificar consultas lentas
SELECT query, calls, total_time, mean_time 
FROM pg_stat_statements 
WHERE mean_time > 1000
ORDER BY mean_time DESC;
```

### Problema: Espacio en disco
```bash
# Ver tamaño de la base de datos
docker exec gamc_postgres psql -U gamc_user gamc_system -c "SELECT pg_size_pretty(pg_database_size('gamc_system'));"

# Ver tamaño por tabla
docker exec gamc_postgres psql -U gamc_user gamc_system -c "SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename::text)) as size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(tablename::text) DESC;"
```

## Configuración de Producción

### Variables de Entorno Recomendadas
```bash
POSTGRES_PASSWORD=<contraseña_segura>
POSTGRES_SHARED_BUFFERS=512MB
POSTGRES_EFFECTIVE_CACHE_SIZE=2GB
POSTGRES_MAX_CONNECTIONS=200
```

### Configuración de Seguridad
1. Cambiar contraseñas por defecto
2. Configurar SSL/TLS
3. Restringir acceso por IP
4. Habilitar auditoría completa
5. Configurar backups automáticos

## Soporte
Para problemas técnicos, contactar al equipo de Tecnología del GAMC.
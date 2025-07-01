-- ========================================
-- GAMC Sistema Web Centralizado
-- Datos Iniciales (Seed Data)
-- ========================================

-- ========================================
-- UNIDADES ORGANIZACIONALES
-- ========================================

INSERT INTO organizational_units (code, name, description, manager_name, email, phone) VALUES
('OBRAS_PUBLICAS', 'Obras Públicas', 'Responsable de la infraestructura y construcción municipal', 'Juan Carlos Mendoza', 'obras.publicas@gamc.gov.bo', '591-4-4234567'),
('MONITOREO', 'Monitoreo', 'Supervisión y seguimiento de proyectos municipales', 'María Elena Vargas', 'monitoreo@gamc.gov.bo', '591-4-4234568'),
('MOVILIDAD_URBANA', 'Movilidad Urbana', 'Gestión del tráfico y transporte urbano', 'Roberto Silva Paz', 'movilidad@gamc.gov.bo', '591-4-4234569'),
('GOBIERNO_ELECTRONICO', 'Gobierno Electrónico', 'Digitalización y tecnología municipal', 'Ana Patricia Rojas', 'gobierno.electronico@gamc.gov.bo', '591-4-4234570'),
('PRENSA_IMAGEN', 'Prensa e Imagen', 'Comunicación institucional y relaciones públicas', 'Carlos Fernando Luna', 'prensa@gamc.gov.bo', '591-4-4234571'),
('TECNOLOGIA', 'Tecnología', 'Infraestructura tecnológica y sistemas', 'Patricia Gonzales', 'tecnologia@gamc.gov.bo', '591-4-4234572'),
('ADMINISTRACION', 'Administración', 'Gestión administrativa general', 'Miguel Angel Torres', 'admin@gamc.gov.bo', '591-4-4234573');

-- ========================================
-- TIPOS DE MENSAJES
-- ========================================

INSERT INTO message_types (code, name, description, priority_level, color) VALUES
('SOLICITUD', 'Solicitud', 'Petición de información o recursos', 3, '#007bff'),
('INFORME', 'Informe', 'Reporte de actividades o resultados', 3, '#28a745'),
('URGENTE', 'Urgente', 'Comunicación que requiere atención inmediata', 1, '#dc3545'),
('COORDINACION', 'Coordinación', 'Mensajes para coordinar actividades entre unidades', 2, '#fd7e14'),
('NOTIFICACION', 'Notificación', 'Avisos informativos generales', 4, '#6c757d'),
('SEGUIMIENTO', 'Seguimiento', 'Control y seguimiento de tareas o proyectos', 2, '#20c997'),
('CONSULTA', 'Consulta', 'Consultas técnicas o procedimentales', 3, '#6f42c1'),
('DIRECTRIZ', 'Directriz', 'Instrucciones o lineamientos oficiales', 1, '#e83e8c');

-- ========================================
-- ESTADOS DE MENSAJES
-- ========================================

INSERT INTO message_statuses (code, name, description, color, is_final, sort_order) VALUES
('DRAFT', 'Borrador', 'Mensaje en proceso de creación', '#6c757d', false, 1),
('SENT', 'Enviado', 'Mensaje enviado, pendiente de lectura', '#007bff', false, 2),
('READ', 'Leído', 'Mensaje leído por el destinatario', '#17a2b8', false, 3),
('IN_PROGRESS', 'En Proceso', 'Mensaje en proceso de atención', '#ffc107', false, 4),
('RESPONDED', 'Respondido', 'Mensaje con respuesta enviada', '#28a745', false, 5),
('RESOLVED', 'Resuelto', 'Mensaje completamente resuelto', '#28a745', true, 6),
('ARCHIVED', 'Archivado', 'Mensaje archivado', '#6c757d', true, 7),
('CANCELLED', 'Cancelado', 'Mensaje cancelado', '#dc3545', true, 8);

-- ========================================
-- USUARIOS ADMINISTRADORES
-- ========================================

-- Contraseña por defecto: "admin123" (hasheada con bcrypt)
-- Hash generado: $2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG

INSERT INTO users (username, email, password_hash, first_name, last_name, role, organizational_unit_id) VALUES
-- Administrador principal
('admin', 'admin@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Administrador', 'Sistema', 'admin', 
 (SELECT id FROM organizational_units WHERE code = 'ADMINISTRACION')),

-- Usuarios INPUT (pueden enviar mensajes)
('obras.input', 'obras.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Juan Carlos', 'Mendoza', 'input',
 (SELECT id FROM organizational_units WHERE code = 'OBRAS_PUBLICAS')),

('monitoreo.input', 'monitoreo.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'María Elena', 'Vargas', 'input',
 (SELECT id FROM organizational_units WHERE code = 'MONITOREO')),

('movilidad.input', 'movilidad.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Roberto', 'Silva Paz', 'input',
 (SELECT id FROM organizational_units WHERE code = 'MOVILIDAD_URBANA')),

('gobierno.input', 'gobierno.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Ana Patricia', 'Rojas', 'input',
 (SELECT id FROM organizational_units WHERE code = 'GOBIERNO_ELECTRONICO')),

('prensa.input', 'prensa.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Carlos Fernando', 'Luna', 'input',
 (SELECT id FROM organizational_units WHERE code = 'PRENSA_IMAGEN')),

('tecnologia.input', 'tecnologia.input@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Patricia', 'Gonzales', 'input',
 (SELECT id FROM organizational_units WHERE code = 'TECNOLOGIA')),

-- Usuarios OUTPUT (pueden ver y responder mensajes)
('obras.output', 'obras.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Luis Alberto', 'Mamani', 'output',
 (SELECT id FROM organizational_units WHERE code = 'OBRAS_PUBLICAS')),

('monitoreo.output', 'monitoreo.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Sandra Patricia', 'Quispe', 'output',
 (SELECT id FROM organizational_units WHERE code = 'MONITOREO')),

('movilidad.output', 'movilidad.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Fernando Jose', 'Torrez', 'output',
 (SELECT id FROM organizational_units WHERE code = 'MOVILIDAD_URBANA')),

('gobierno.output', 'gobierno.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Carla Beatriz', 'Mendoza', 'output',
 (SELECT id FROM organizational_units WHERE code = 'GOBIERNO_ELECTRONICO')),

('prensa.output', 'prensa.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Jorge Antonio', 'Vargas', 'output',
 (SELECT id FROM organizational_units WHERE code = 'PRENSA_IMAGEN')),

('tecnologia.output', 'tecnologia.output@gamc.gov.bo', '$2b$10$rOz8VQ2kJqwjVcJWMHOj6O6J5gSLOvdUzrr6hE8bYXhCFqRZzFTQG', 'Mónica Isabel', 'Fernandez', 'output',
 (SELECT id FROM organizational_units WHERE code = 'TECNOLOGIA'));

-- ========================================
-- MENSAJES DE EJEMPLO
-- ========================================

-- Mensajes entre diferentes unidades para demostrar funcionalidad
INSERT INTO messages (subject, content, sender_id, sender_unit_id, receiver_unit_id, message_type_id, status_id, priority_level, is_urgent) VALUES

-- De Obras Públicas a Movilidad Urbana
('Coordinación para obras en Av. Ballivián', 
 'Estimados colegas de Movilidad Urbana,\n\nLes informamos que el próximo lunes 19 de junio iniciaremos trabajos de repavimentación en la Av. Ballivián entre calles 25 de Mayo y Heroínas.\n\nSolicitamos coordinar el desvío de tráfico y señalización correspondiente.\n\nSaludos cordiales.',
 (SELECT id FROM users WHERE username = 'obras.input'),
 (SELECT id FROM organizational_units WHERE code = 'OBRAS_PUBLICAS'),
 (SELECT id FROM organizational_units WHERE code = 'MOVILIDAD_URBANA'),
 (SELECT id FROM message_types WHERE code = 'COORDINACION'),
 (SELECT id FROM message_statuses WHERE code = 'SENT'),
 2, false),

-- De Monitoreo a Tecnología
('Solicitud de acceso a sistema de seguimiento',
 'Estimado equipo de Tecnología,\n\nNecesitamos acceso al nuevo módulo de seguimiento de proyectos para 5 usuarios adicionales de nuestra unidad.\n\nAdjunto la lista de usuarios que requieren acceso.\n\nGracias por su apoyo.',
 (SELECT id FROM users WHERE username = 'monitoreo.input'),
 (SELECT id FROM organizational_units WHERE code = 'MONITOREO'),
 (SELECT id FROM organizational_units WHERE code = 'TECNOLOGIA'),
 (SELECT id FROM message_types WHERE code = 'SOLICITUD'),
 (SELECT id FROM message_statuses WHERE code = 'SENT'),
 3, false),

-- De Prensa a Gobierno Electrónico
('URGENTE: Actualización página web municipal',
 'Estimados colegas,\n\nNecesitamos actualizar URGENTEMENTE la información en la página web sobre el nuevo horario de atención al público.\n\nLa información debe estar publicada antes de las 14:00 de hoy.\n\nPor favor confirmen recepción.',
 (SELECT id FROM users WHERE username = 'prensa.input'),
 (SELECT id FROM organizational_units WHERE code = 'PRENSA_IMAGEN'),
 (SELECT id FROM organizational_units WHERE code = 'GOBIERNO_ELECTRONICO'),
 (SELECT id FROM message_types WHERE code = 'URGENTE'),
 (SELECT id FROM message_statuses WHERE code = 'SENT'),
 1, true),

-- De Tecnología a todos (notificación)
('Mantenimiento programado del sistema',
 'Estimados usuarios,\n\nLes informamos que el sábado 24 de junio de 02:00 a 06:00 AM se realizará mantenimiento programado del sistema.\n\nDurante este periodo el sistema no estará disponible.\n\nDisculpen las molestias.',
 (SELECT id FROM users WHERE username = 'tecnologia.input'),
 (SELECT id FROM organizational_units WHERE code = 'TECNOLOGIA'),
 (SELECT id FROM organizational_units WHERE code = 'ADMINISTRACION'),
 (SELECT id FROM message_types WHERE code = 'NOTIFICACION'),
 (SELECT id FROM message_statuses WHERE code = 'SENT'),
 4, false),

-- De Gobierno Electrónico a Monitoreo
('Informe de implementación sistema web',
 'Estimados colegas,\n\nAdjunto el informe de avance de la implementación del nuevo sistema web centralizado.\n\nHasta la fecha se ha completado el 75% del desarrollo backend y 60% del frontend.\n\nQuedamos atentos a sus observaciones.',
 (SELECT id FROM users WHERE username = 'gobierno.input'),
 (SELECT id FROM organizational_units WHERE code = 'GOBIERNO_ELECTRONICO'),
 (SELECT id FROM organizational_units WHERE code = 'MONITOREO'),
 (SELECT id FROM message_types WHERE code = 'INFORME'),
 (SELECT id FROM message_statuses WHERE code = 'READ'),
 3, false);

-- ========================================
-- POPULAR TABLA DE HECHOS
-- ========================================

-- Insertar datos en fact_messages basados en los mensajes creados
INSERT INTO fact_messages (
    message_id, sender_unit_id, receiver_unit_id, message_type_id,
    date_id, time_id, priority_level, has_attachments, attachment_count
)
SELECT 
    m.id,
    m.sender_unit_id,
    m.receiver_unit_id,
    m.message_type_id,
    TO_CHAR(m.created_at, 'YYYYMMDD')::INTEGER as date_id,
    EXTRACT(HOUR FROM m.created_at)::INTEGER * 100 + 
    EXTRACT(MINUTE FROM m.created_at)::INTEGER as time_id,
    m.priority_level,
    false as has_attachments,
    0 as attachment_count
FROM messages m;

-- ========================================
-- LOGS DE AUDITORÍA DE EJEMPLO
-- ========================================

INSERT INTO audit_logs (user_id, action, resource, resource_id, new_values, ip_address, result) VALUES
((SELECT id FROM users WHERE username = 'admin'), 'CREATE', 'organizational_units', '1', '{"name":"Obras Públicas","code":"OBRAS_PUBLICAS"}', '192.168.1.100', 'success'),
((SELECT id FROM users WHERE username = 'admin'), 'CREATE', 'message_types', '1', '{"name":"Solicitud","code":"SOLICITUD"}', '192.168.1.100', 'success'),
((SELECT id FROM users WHERE username = 'admin'), 'CREATE', 'users', (SELECT id::text FROM users WHERE username = 'obras.input'), '{"username":"obras.input","role":"input"}', '192.168.1.100', 'success'),
((SELECT id FROM users WHERE username = 'obras.input'), 'CREATE', 'messages', '1', '{"subject":"Coordinación para obras en Av. Ballivián"}', '192.168.1.101', 'success'),
((SELECT id FROM users WHERE username = 'monitoreo.input'), 'CREATE', 'messages', '2', '{"subject":"Solicitud de acceso a sistema"}', '192.168.1.102', 'success');

-- ========================================
-- DATOS ADICIONALES PARA PRUEBAS
-- ========================================

-- Actualizar algunos mensajes para simular flujo completo
UPDATE messages SET 
    status_id = (SELECT id FROM message_statuses WHERE code = 'READ'),
    read_at = CURRENT_TIMESTAMP - INTERVAL '2 hours'
WHERE id = 1;

UPDATE messages SET 
    status_id = (SELECT id FROM message_statuses WHERE code = 'IN_PROGRESS'),
    read_at = CURRENT_TIMESTAMP - INTERVAL '1 hour'
WHERE id = 2;

UPDATE messages SET 
    status_id = (SELECT id FROM message_statuses WHERE code = 'RESPONDED'),
    read_at = CURRENT_TIMESTAMP - INTERVAL '30 minutes',
    responded_at = CURRENT_TIMESTAMP - INTERVAL '10 minutes'
WHERE id = 3;

-- ========================================
-- FUNCIONES ÚTILES PARA LA APLICACIÓN
-- ========================================

-- Función para obtener estadísticas del dashboard
CREATE OR REPLACE FUNCTION get_dashboard_stats()
RETURNS TABLE(
    total_messages BIGINT,
    messages_today BIGINT,
    urgent_messages BIGINT,
    pending_messages BIGINT,
    avg_response_time NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*) FROM messages) as total_messages,
        (SELECT COUNT(*) FROM messages WHERE DATE(created_at) = CURRENT_DATE) as messages_today,
        (SELECT COUNT(*) FROM messages WHERE is_urgent = true AND status_id NOT IN 
         (SELECT id FROM message_statuses WHERE is_final = true)) as urgent_messages,
        (SELECT COUNT(*) FROM messages WHERE status_id IN 
         (SELECT id FROM message_statuses WHERE code IN ('SENT', 'READ', 'IN_PROGRESS'))) as pending_messages,
        (SELECT AVG(EXTRACT(EPOCH FROM (responded_at - created_at))/60) 
         FROM messages WHERE responded_at IS NOT NULL) as avg_response_time;
END;
$$ LANGUAGE plpgsql;

-- Función para obtener mensajes por unidad organizacional
CREATE OR REPLACE FUNCTION get_messages_by_unit(unit_id INTEGER, limit_count INTEGER DEFAULT 50)
RETURNS TABLE(
    id BIGINT,
    subject VARCHAR,
    sender_name TEXT,
    sender_unit VARCHAR,
    receiver_unit VARCHAR,
    message_type VARCHAR,
    status VARCHAR,
    priority_level INTEGER,
    is_urgent BOOLEAN,
    created_at TIMESTAMP,
    attachment_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.id,
        m.subject,
        u.first_name || ' ' || u.last_name as sender_name,
        ou_sender.name as sender_unit,
        ou_receiver.name as receiver_unit,
        mt.name as message_type,
        ms.name as status,
        m.priority_level,
        m.is_urgent,
        m.created_at,
        (SELECT COUNT(*) FROM message_attachments WHERE message_id = m.id) as attachment_count
    FROM messages m
    JOIN users u ON m.sender_id = u.id
    JOIN organizational_units ou_sender ON m.sender_unit_id = ou_sender.id
    JOIN organizational_units ou_receiver ON m.receiver_unit_id = ou_receiver.id
    JOIN message_types mt ON m.message_type_id = mt.id
    JOIN message_statuses ms ON m.status_id = ms.id
    WHERE m.sender_unit_id = unit_id OR m.receiver_unit_id = unit_id
    ORDER BY m.created_at DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Función para marcar mensaje como leído
CREATE OR REPLACE FUNCTION mark_message_as_read(message_id BIGINT, user_id UUID)
RETURNS BOOLEAN AS $$
DECLARE
    message_exists BOOLEAN;
    user_has_access BOOLEAN;
BEGIN
    -- Verificar si el mensaje existe
    SELECT EXISTS(SELECT 1 FROM messages WHERE id = message_id) INTO message_exists;
    
    IF NOT message_exists THEN
        RETURN FALSE;
    END IF;
    
    -- Verificar si el usuario tiene acceso al mensaje
    SELECT EXISTS(
        SELECT 1 FROM messages m 
        JOIN users u ON u.id = user_id
        WHERE m.id = message_id 
        AND (m.receiver_unit_id = u.organizational_unit_id OR u.role = 'admin')
    ) INTO user_has_access;
    
    IF NOT user_has_access THEN
        RETURN FALSE;
    END IF;
    
    -- Actualizar el mensaje
    UPDATE messages SET 
        status_id = (SELECT id FROM message_statuses WHERE code = 'READ'),
        read_at = CURRENT_TIMESTAMP,
        updated_at = CURRENT_TIMESTAMP
    WHERE id = message_id 
    AND status_id = (SELECT id FROM message_statuses WHERE code = 'SENT');
    
    -- Registrar en auditoría
    INSERT INTO audit_logs (user_id, action, resource, resource_id, result)
    VALUES (user_id, 'READ', 'messages', message_id::TEXT, 'success');
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- CREAR ÍNDICES ADICIONALES PARA OPTIMIZACIÓN
-- ========================================

-- Índice compuesto para consultas frecuentes
CREATE INDEX idx_messages_unit_status ON messages(receiver_unit_id, status_id, created_at DESC);
CREATE INDEX idx_messages_sender_date ON messages(sender_unit_id, created_at DESC);
CREATE INDEX idx_messages_urgent_status ON messages(is_urgent, status_id) WHERE is_urgent = true;

-- Índice para búsquedas de texto en asunto y contenido
CREATE INDEX idx_messages_subject_text ON messages USING gin(to_tsvector('spanish', subject));
CREATE INDEX idx_messages_content_text ON messages USING gin(to_tsvector('spanish', content));

-- ========================================
-- CONFIGURACIÓN FINAL
-- ========================================

-- Configurar timezone por defecto
SET timezone = 'America/La_Paz';

-- Actualizar estadísticas de la base de datos
ANALYZE;

-- Mensaje de confirmación
DO $$
BEGIN
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'GAMC Sistema Web Centralizado';
    RAISE NOTICE 'Base de datos inicializada correctamente';
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Unidades Organizacionales: %', (SELECT COUNT(*) FROM organizational_units);
    RAISE NOTICE 'Usuarios creados: %', (SELECT COUNT(*) FROM users);
    RAISE NOTICE 'Tipos de mensajes: %', (SELECT COUNT(*) FROM message_types);
    RAISE NOTICE 'Estados de mensajes: %', (SELECT COUNT(*) FROM message_statuses);
    RAISE NOTICE 'Mensajes de ejemplo: %', (SELECT COUNT(*) FROM messages);
    RAISE NOTICE '==============================================';
    RAISE NOTICE 'Credenciales por defecto:';
    RAISE NOTICE 'Usuario: admin | Contraseña: admin123';
    RAISE NOTICE 'Usuario: obras.input | Contraseña: admin123';
    RAISE NOTICE 'Usuario: monitoreo.output | Contraseña: admin123';
    RAISE NOTICE '==============================================';
END $$;
-- ========================================
-- GAMC Sistema Web Centralizado
-- Inicialización de Base de Datos
-- ========================================

-- Crear extensiones necesarias
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ========================================
-- TABLAS INDEPENDIENTES (SIN DEPENDENCIAS)
-- ========================================

-- Tabla de Unidades Organizacionales
CREATE TABLE organizational_units (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    manager_name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Tipos de Mensajes
CREATE TABLE message_types (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    priority_level INTEGER DEFAULT 3, -- 1=Urgente, 2=Alto, 3=Normal, 4=Bajo
    color VARCHAR(7) DEFAULT '#007bff', -- Color hex para UI
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Estados de Mensajes
CREATE TABLE message_statuses (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#6c757d',
    is_final BOOLEAN DEFAULT false, -- Si es un estado final
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- TABLAS CON DEPENDENCIAS DE PRIMER NIVEL
-- ========================================

-- Tabla de Usuarios
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'input', 'output')),
    organizational_unit_id INTEGER REFERENCES organizational_units(id),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    password_changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Sesiones (para JWT refresh tokens)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(255) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- TABLA PARA RESET DE CONTRASEÑAS
-- ========================================

-- Tabla de tokens de reset de contraseña
CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP WITH TIME ZONE NULL,
    request_ip VARCHAR(45) NOT NULL,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT true,
    
    -- Metadatos adicionales para auditoría
    email_sent_at TIMESTAMP WITH TIME ZONE,
    email_opened_at TIMESTAMP WITH TIME ZONE,
    attempts_count INTEGER DEFAULT 0,
    
    -- Constraints
    CONSTRAINT chk_expires_future CHECK (expires_at > created_at),
    CONSTRAINT chk_used_after_created CHECK (used_at IS NULL OR used_at >= created_at),
    CONSTRAINT chk_attempts_positive CHECK (attempts_count >= 0)
);

-- ========================================
-- TABLAS PRINCIPALES DE MENSAJERÍA
-- ========================================

-- Tabla de Mensajes
CREATE TABLE messages (
    id BIGSERIAL PRIMARY KEY,
    subject VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    sender_id UUID NOT NULL REFERENCES users(id),
    sender_unit_id INTEGER NOT NULL REFERENCES organizational_units(id),
    receiver_unit_id INTEGER NOT NULL REFERENCES organizational_units(id),
    message_type_id INTEGER NOT NULL REFERENCES message_types(id),
    status_id INTEGER NOT NULL REFERENCES message_statuses(id),
    priority_level INTEGER DEFAULT 3,
    is_urgent BOOLEAN DEFAULT false,
    read_at TIMESTAMP,
    responded_at TIMESTAMP,
    archived_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de Respuestas a Mensajes
CREATE TABLE message_responses (
    id BIGSERIAL PRIMARY KEY,
    original_message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    response_message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(original_message_id, response_message_id)
);

-- Tabla de Archivos Adjuntos
CREATE TABLE message_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    original_name VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL, -- Nombre en MinIO
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- TABLAS DE AUDITORÍA Y LOG
-- ========================================

-- Tabla de Logs de Auditoría
CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    result VARCHAR(20) DEFAULT 'success', -- success, failure
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- TABLAS PARA DATA WAREHOUSE (ESQUEMA ESTRELLA)
-- ========================================

-- Dimensión Tiempo
CREATE TABLE dim_date (
    date_id INTEGER PRIMARY KEY,
    full_date DATE NOT NULL,
    year INTEGER NOT NULL,
    quarter INTEGER NOT NULL,
    month INTEGER NOT NULL,
    month_name VARCHAR(20) NOT NULL,
    week INTEGER NOT NULL,
    day_of_month INTEGER NOT NULL,
    day_of_week INTEGER NOT NULL,
    day_name VARCHAR(20) NOT NULL,
    is_weekend BOOLEAN NOT NULL,
    is_holiday BOOLEAN DEFAULT false
);

-- Dimensión Tiempo (Horas)
CREATE TABLE dim_time (
    time_id INTEGER PRIMARY KEY,
    hour INTEGER NOT NULL,
    minute INTEGER NOT NULL,
    time_period VARCHAR(10) NOT NULL, -- AM/PM
    hour_24 INTEGER NOT NULL,
    minute_of_day INTEGER NOT NULL
);

-- Tabla de Hechos - Mensajes
CREATE TABLE fact_messages (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES messages(id),
    sender_unit_id INTEGER NOT NULL REFERENCES organizational_units(id),
    receiver_unit_id INTEGER NOT NULL REFERENCES organizational_units(id),
    message_type_id INTEGER NOT NULL REFERENCES message_types(id),
    date_id INTEGER NOT NULL REFERENCES dim_date(date_id),
    time_id INTEGER NOT NULL REFERENCES dim_time(time_id),
    priority_level INTEGER NOT NULL,
    response_time_minutes INTEGER, -- Tiempo de respuesta en minutos
    has_attachments BOOLEAN DEFAULT false,
    attachment_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- ÍNDICES PARA OPTIMIZACIÓN
-- ========================================

-- Índices para usuarios
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_org_unit ON users(organizational_unit_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(is_active);

-- Índices para password reset tokens
CREATE UNIQUE INDEX idx_password_reset_tokens_token ON password_reset_tokens (token) WHERE is_active = true;
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id, is_active);
CREATE INDEX idx_password_reset_tokens_expires ON password_reset_tokens (expires_at) WHERE is_active = true;
CREATE INDEX idx_password_reset_tokens_ip_created ON password_reset_tokens (request_ip, created_at);

-- Índices para mensajes
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_sender_unit ON messages(sender_unit_id);
CREATE INDEX idx_messages_receiver_unit ON messages(receiver_unit_id);
CREATE INDEX idx_messages_status ON messages(status_id);
CREATE INDEX idx_messages_type ON messages(message_type_id);
CREATE INDEX idx_messages_created ON messages(created_at);
CREATE INDEX idx_messages_priority ON messages(priority_level);
CREATE INDEX idx_messages_urgent ON messages(is_urgent);

-- Índices para archivos adjuntos
CREATE INDEX idx_attachments_message ON message_attachments(message_id);
CREATE INDEX idx_attachments_uploader ON message_attachments(uploaded_by);

-- Índices para auditoría
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_resource ON audit_logs(resource);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

-- Índices para data warehouse
CREATE INDEX idx_fact_messages_date ON fact_messages(date_id);
CREATE INDEX idx_fact_messages_time ON fact_messages(time_id);
CREATE INDEX idx_fact_messages_sender_unit ON fact_messages(sender_unit_id);
CREATE INDEX idx_fact_messages_receiver_unit ON fact_messages(receiver_unit_id);

-- ========================================
-- FUNCIONES DE UTILIDAD PARA PASSWORD RESET
-- ========================================

-- Función para generar token único
CREATE OR REPLACE FUNCTION generate_reset_token()
RETURNS VARCHAR(255) AS $$
BEGIN
    RETURN encode(gen_random_bytes(32), 'hex');
END;
$$ LANGUAGE plpgsql;

-- Función para limpiar tokens expirados automáticamente
CREATE OR REPLACE FUNCTION cleanup_expired_reset_tokens()
RETURNS INTEGER AS $$
DECLARE
    cleaned_count INTEGER;
BEGIN
    -- Marcar como inactivos los tokens expirados
    UPDATE password_reset_tokens 
    SET is_active = false
    WHERE is_active = true 
      AND expires_at < NOW();
    
    GET DIAGNOSTICS cleaned_count = ROW_COUNT;
    
    -- Log para auditoría
    INSERT INTO audit_logs (user_id, action, resource, resource_id, old_values, result, ip_address)
    VALUES (
        NULL,
        'CLEANUP',
        'password_reset_tokens',
        'expired_tokens',
        jsonb_build_object('cleaned_count', cleaned_count, 'timestamp', NOW()),
        'success',
        'system'
    );
    
    RETURN cleaned_count;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- TRIGGERS PARA ACTUALIZACIÓN AUTOMÁTICA
-- ========================================

-- Función para actualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger para validar email institucional antes de crear token
CREATE OR REPLACE FUNCTION validate_reset_request()
RETURNS TRIGGER AS $$
DECLARE
    user_email VARCHAR(100);
    user_active BOOLEAN;
BEGIN
    -- Obtener datos del usuario
    SELECT email, is_active INTO user_email, user_active
    FROM users 
    WHERE id = NEW.user_id;
    
    -- Validar que el usuario existe y está activo
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Usuario no encontrado'
            USING ERRCODE = 'foreign_key_violation';
    END IF;
    
    IF NOT user_active THEN
        RAISE EXCEPTION 'Usuario inactivo no puede solicitar reset de contraseña'
            USING ERRCODE = 'check_violation';
    END IF;
    
    -- VALIDACIÓN PRINCIPAL: Solo emails institucionales
    IF user_email NOT LIKE '%@gamc.gov.bo' THEN
        RAISE EXCEPTION 'Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña'
            USING ERRCODE = 'check_violation',
                  HINT = 'Contacte al administrador para cambio de contraseña';
    END IF;
    
    -- Validar que no hay tokens activos recientes (prevenir spam)
    IF EXISTS (
        SELECT 1 FROM password_reset_tokens 
        WHERE user_id = NEW.user_id 
          AND is_active = true 
          AND created_at > (NOW() - INTERVAL '5 minutes')
    ) THEN
        RAISE EXCEPTION 'Debe esperar 5 minutos entre solicitudes de reset'
            USING ERRCODE = 'check_violation';
    END IF;
    
    -- Invalidar tokens anteriores del mismo usuario
    UPDATE password_reset_tokens 
    SET is_active = false 
    WHERE user_id = NEW.user_id AND is_active = true;
    
    -- Generar token único si no se proporcionó
    IF NEW.token IS NULL OR NEW.token = '' THEN
        NEW.token := generate_reset_token();
    END IF;
    
    -- Establecer expiración por defecto (30 minutos)
    IF NEW.expires_at IS NULL THEN
        NEW.expires_at := NOW() + INTERVAL '30 minutes';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers para updated_at
CREATE TRIGGER update_organizational_units_updated_at BEFORE UPDATE ON organizational_units
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger para validar password reset
CREATE TRIGGER trigger_validate_reset_request
    BEFORE INSERT ON password_reset_tokens
    FOR EACH ROW
    EXECUTE FUNCTION validate_reset_request();

-- ========================================
-- FUNCIÓN PARA POPULAR DIMENSIONES DE TIEMPO
-- ========================================

-- Función para popular dim_date
CREATE OR REPLACE FUNCTION populate_dim_date(start_date DATE, end_date DATE)
RETURNS VOID AS $$
DECLARE
    curr_date DATE := start_date;
BEGIN
    WHILE curr_date <= end_date LOOP
        INSERT INTO dim_date (
            date_id, full_date, year, quarter, month, month_name,
            week, day_of_month, day_of_week, day_name, is_weekend
        ) VALUES (
            TO_CHAR(curr_date, 'YYYYMMDD')::INTEGER,
            curr_date,
            EXTRACT(YEAR FROM curr_date)::INTEGER,
            EXTRACT(QUARTER FROM curr_date)::INTEGER,
            EXTRACT(MONTH FROM curr_date)::INTEGER,
            TO_CHAR(curr_date, 'Month'),
            EXTRACT(WEEK FROM curr_date)::INTEGER,
            EXTRACT(DAY FROM curr_date)::INTEGER,
            EXTRACT(DOW FROM curr_date)::INTEGER,
            TO_CHAR(curr_date, 'Day'),
            EXTRACT(DOW FROM curr_date) IN (0, 6)
        ) ON CONFLICT (date_id) DO NOTHING;
        
        curr_date := curr_date + INTERVAL '1 day';
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Función para popular dim_time
CREATE OR REPLACE FUNCTION populate_dim_time()
RETURNS VOID AS $$
DECLARE
    h INTEGER;
    m INTEGER;
BEGIN
    FOR h IN 0..23 LOOP
        FOR m IN 0..59 LOOP
            INSERT INTO dim_time (
                time_id, hour, minute, time_period, hour_24, minute_of_day
            ) VALUES (
                h * 100 + m,
                CASE WHEN h = 0 THEN 12 WHEN h > 12 THEN h - 12 ELSE h END,
                m,
                CASE WHEN h < 12 THEN 'AM' ELSE 'PM' END,
                h,
                h * 60 + m
            ) ON CONFLICT (time_id) DO NOTHING;
        END LOOP;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Popular las dimensiones de tiempo (2020-2030)
SELECT populate_dim_date('2020-01-01'::DATE, '2030-12-31'::DATE);
SELECT populate_dim_time();

-- ========================================
-- VISTAS ÚTILES
-- ========================================

-- Vista de mensajes con información completa
CREATE VIEW v_messages_full AS
SELECT 
    m.id,
    m.subject,
    m.content,
    m.priority_level,
    m.is_urgent,
    m.created_at,
    m.updated_at,
    m.read_at,
    m.responded_at,
    u.first_name || ' ' || u.last_name as sender_name,
    u.email as sender_email,
    ou_sender.name as sender_unit_name,
    ou_receiver.name as receiver_unit_name,
    mt.name as message_type_name,
    mt.color as message_type_color,
    ms.name as status_name,
    ms.color as status_color,
    (SELECT COUNT(*) FROM message_attachments WHERE message_id = m.id) as attachment_count
FROM messages m
JOIN users u ON m.sender_id = u.id
JOIN organizational_units ou_sender ON m.sender_unit_id = ou_sender.id
JOIN organizational_units ou_receiver ON m.receiver_unit_id = ou_receiver.id
JOIN message_types mt ON m.message_type_id = mt.id
JOIN message_statuses ms ON m.status_id = ms.id;

-- Vista de estadísticas por unidad organizacional
CREATE VIEW v_unit_stats AS
SELECT 
    ou.id,
    ou.name,
    (SELECT COUNT(*) FROM messages WHERE sender_unit_id = ou.id) as messages_sent,
    (SELECT COUNT(*) FROM messages WHERE receiver_unit_id = ou.id) as messages_received,
    (SELECT COUNT(*) FROM users WHERE organizational_unit_id = ou.id AND is_active = true) as active_users
FROM organizational_units ou
WHERE ou.is_active = true;

-- Vista para auditoría de resets de contraseña
CREATE VIEW v_password_reset_audit AS
SELECT 
    prt.id,
    prt.user_id,
    u.email,
    u.first_name,
    u.last_name,
    ou.name as organizational_unit,
    prt.created_at,
    prt.expires_at,
    prt.used_at,
    prt.request_ip,
    prt.attempts_count,
    CASE 
        WHEN prt.used_at IS NOT NULL THEN 'USED'
        WHEN prt.expires_at < NOW() THEN 'EXPIRED'
        WHEN prt.is_active THEN 'ACTIVE'
        ELSE 'INACTIVE'
    END as status,
    EXTRACT(EPOCH FROM (prt.expires_at - prt.created_at))/60 as duration_minutes
FROM password_reset_tokens prt
JOIN users u ON prt.user_id = u.id
JOIN organizational_units ou ON u.organizational_unit_id = ou.id
ORDER BY prt.created_at DESC;

-- ========================================
-- COMENTARIOS PARA DOCUMENTACIÓN
-- ========================================

COMMENT ON TABLE password_reset_tokens IS 'Tokens de reset de contraseña para usuarios institucionales (@gamc.gov.bo)';
COMMENT ON COLUMN password_reset_tokens.token IS 'Token único de 64 caracteres hex para reset';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'Expiración del token (30 minutos por defecto)';
COMMENT ON COLUMN password_reset_tokens.used_at IS 'Timestamp cuando se usó el token (previene reutilización)';
COMMENT ON COLUMN password_reset_tokens.request_ip IS 'IP desde donde se solicitó el reset';
COMMENT ON COLUMN password_reset_tokens.attempts_count IS 'Número de intentos de uso del token';
COMMENT ON COLUMN password_reset_tokens.email_sent_at IS 'Cuándo se envió el email de reset';
COMMENT ON COLUMN password_reset_tokens.email_opened_at IS 'Cuándo el usuario abrió el email (tracking)';

COMMENT ON DATABASE postgres IS 'GAMC Sistema Web Centralizado - Base de Datos Principal';
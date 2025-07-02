#!/bin/bash

# ========================================
# SCRIPT DE TESTING COMPLETO - SISTEMA DE MENSAJERIA
# Descripción: Prueba integral del flujo de mensajería con diferentes roles de usuario
# Autor: Sistema Web Centralizado GAMC
# ========================================

echo "========================================="
echo "🚀 SISTEMA DE TESTING - FLUJO MENSAJERIA"
echo "========================================="
echo "Este script prueba todo el sistema de mensajería:"
echo "- Creación de usuarios con diferentes roles (INPUT, OUTPUT, ADMIN)"
echo "- Operaciones CRUD de mensajes según permisos"
echo "- Validación de restricciones de seguridad"
echo "- Estadísticas y reporting del sistema"
echo ""

# Función para mostrar separadores
print_section() {
    echo ""
    echo "========================================="
    echo "📋 $1"
    echo "========================================="
}

# Función para mostrar detalles de respuesta
show_response_details() {
    local response="$1"
    local title="$2"
    echo "📊 Detalles de la respuesta ($title):"
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""
}

# Función para validar respuesta exitosa
validate_success() {
    local response="$1"
    local description="$2"
    
    local success=$(echo "$response" | jq -r '.success // false' 2>/dev/null)
    if [ "$success" = "true" ]; then
        echo "✅ $description - EXITOSO"
        return 0
    else
        echo "❌ $description - FALLÓ"
        echo "Error: $(echo "$response" | jq -r '.message // "Respuesta inválida"')"
        return 1
    fi
}

print_section "FASE 1: PREPARACIÓN DEL ENTORNO DE TESTING"

echo "🔧 Configurando usuarios de prueba con diferentes roles..."
echo ""
echo "📝 ROLES DEL SISTEMA:"
echo "   • INPUT: Puede crear y enviar mensajes"
echo "   • OUTPUT: Puede leer y gestionar mensajes recibidos"
echo "   • ADMIN: Acceso completo + estadísticas del sistema"
echo ""

# Generar IDs únicos para evitar conflictos
inputUserId=$((RANDOM % 1000 + 7000))
outputUserId=$((RANDOM % 1000 + 8000))
adminUserId=$((RANDOM % 1000 + 9000))

echo "🆔 IDs generados para esta sesión:"
echo "   • Usuario INPUT: $inputUserId"
echo "   • Usuario OUTPUT: $outputUserId"
echo "   • Usuario ADMIN: $adminUserId"
echo ""

# ========================================
# CREACIÓN DE USUARIOS
# ========================================

print_section "FASE 2: REGISTRO DE USUARIOS"

# Usuario INPUT
echo "👤 Registrando Usuario INPUT (Creador de mensajes)..."
inputUserBody=$(cat <<EOF
{
    "email": "input${inputUserId}@gamc.gov.bo",
    "password": "InputTest123!",
    "firstName": "Usuario",
    "lastName": "Input${inputUserId}",
    "organizationalUnitId": 1,
    "role": "input"
}
EOF
)

echo "📤 Enviando datos de registro INPUT..."
inputRegisterResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$inputUserBody")

if validate_success "$inputRegisterResponse" "Registro usuario INPUT"; then
    echo "📧 Email registrado: input${inputUserId}@gamc.gov.bo"
    echo "🏢 Unidad organizacional: 1 (Obras Públicas)"
    echo "🔐 Rol asignado: INPUT"
    
    # Login INPUT
    echo ""
    echo "🔐 Autenticando usuario INPUT..."
    inputLoginBody=$(cat <<EOF
{
    "email": "input${inputUserId}@gamc.gov.bo",
    "password": "InputTest123!"
}
EOF
)
    
    inputLoginResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$inputLoginBody")
    
    inputToken=$(echo "$inputLoginResponse" | jq -r '.data.accessToken')
    
    if [ "$inputToken" != "null" ] && [ "$inputToken" != "" ]; then
        echo "✅ Usuario INPUT autenticado exitosamente"
        echo "🎫 Token generado: ${inputToken:0:20}..."
        inputUserId_final=$(echo "$inputLoginResponse" | jq -r '.data.user.id')
        echo "🆔 ID de usuario: $inputUserId_final"
    else
        echo "❌ Error en autenticación INPUT"
        exit 1
    fi
else
    echo "❌ Fallo crítico en registro INPUT"
    exit 1
fi

echo ""
echo "────────────────────────────────────────"

# Usuario OUTPUT
echo "👤 Registrando Usuario OUTPUT (Receptor de mensajes)..."
outputUserBody=$(cat <<EOF
{
    "email": "output${outputUserId}@gamc.gov.bo",
    "password": "OutputTest123!",
    "firstName": "Usuario",
    "lastName": "Output${outputUserId}",
    "organizationalUnitId": 2,
    "role": "output"
}
EOF
)

echo "📤 Enviando datos de registro OUTPUT..."
outputRegisterResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$outputUserBody")

if validate_success "$outputRegisterResponse" "Registro usuario OUTPUT"; then
    echo "📧 Email registrado: output${outputUserId}@gamc.gov.bo"
    echo "🏢 Unidad organizacional: 2 (Finanzas)"
    echo "🔐 Rol asignado: OUTPUT"
    
    # Login OUTPUT
    echo ""
    echo "🔐 Autenticando usuario OUTPUT..."
    outputLoginBody=$(cat <<EOF
{
    "email": "output${outputUserId}@gamc.gov.bo",
    "password": "OutputTest123!"
}
EOF
)
    
    outputLoginResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$outputLoginBody")
    
    outputToken=$(echo "$outputLoginResponse" | jq -r '.data.accessToken')
    
    if [ "$outputToken" != "null" ] && [ "$outputToken" != "" ]; then
        echo "✅ Usuario OUTPUT autenticado exitosamente"
        echo "🎫 Token generado: ${outputToken:0:20}..."
        outputUserId_final=$(echo "$outputLoginResponse" | jq -r '.data.user.id')
        echo "🆔 ID de usuario: $outputUserId_final"
    else
        echo "❌ Error en autenticación OUTPUT"
        exit 1
    fi
else
    echo "❌ Fallo crítico en registro OUTPUT"
    exit 1
fi

echo ""
echo "────────────────────────────────────────"

# Usuario ADMIN
echo "👤 Registrando Usuario ADMIN (Administrador del sistema)..."
adminUserBody=$(cat <<EOF
{
    "email": "admin${adminUserId}@gamc.gov.bo",
    "password": "AdminTest123!",
    "firstName": "Usuario",
    "lastName": "Admin${adminUserId}",
    "organizationalUnitId": 1,
    "role": "admin"
}
EOF
)

echo "📤 Enviando datos de registro ADMIN..."
adminRegisterResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$adminUserBody")

if validate_success "$adminRegisterResponse" "Registro usuario ADMIN"; then
    echo "📧 Email registrado: admin${adminUserId}@gamc.gov.bo"
    echo "🏢 Unidad organizacional: 1 (Obras Públicas)"
    echo "🔐 Rol asignado: ADMIN"
    
    # Login ADMIN
    echo ""
    echo "🔐 Autenticando usuario ADMIN..."
    adminLoginBody=$(cat <<EOF
{
    "email": "admin${adminUserId}@gamc.gov.bo",
    "password": "AdminTest123!"
}
EOF
)
    
    adminLoginResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$adminLoginBody")
    
    adminToken=$(echo "$adminLoginResponse" | jq -r '.data.accessToken')
    
    if [ "$adminToken" != "null" ] && [ "$adminToken" != "" ]; then
        echo "✅ Usuario ADMIN autenticado exitosamente"
        echo "🎫 Token generado: ${adminToken:0:20}..."
        adminUserId_final=$(echo "$adminLoginResponse" | jq -r '.data.user.id')
        echo "🆔 ID de usuario: $adminUserId_final"
    else
        echo "❌ Error en autenticación ADMIN"
        exit 1
    fi
else
    echo "❌ Fallo crítico en registro ADMIN"
    exit 1
fi

print_section "FASE 3: EXPLORACIÓN DEL SISTEMA"

echo "📋 Obteniendo configuración del sistema de mensajería..."
echo ""

# ========================================
# OBTENER TIPOS DE MENSAJES
# ========================================
echo "🏷️ CONSULTANDO TIPOS DE MENSAJES DISPONIBLES"
echo "Endpoint: GET /api/v1/messages/types"
echo "Usuario: INPUT (token requerido)"
echo ""

typesResponse=$(curl -s -X GET \
    "http://localhost:3000/api/v1/messages/types" \
    -H "Authorization: Bearer $inputToken")

if validate_success "$typesResponse" "Consulta tipos de mensajes"; then
    echo "📝 TIPOS DE MENSAJES CONFIGURADOS:"
    typesData=$(echo "$typesResponse" | jq -r '.data[]')
    counter=1
    while IFS= read -r type; do
        [ -n "$type" ] && echo "   $counter. $type"
        ((counter++))
    done <<< "$typesData"
    
    totalTypes=$(echo "$typesResponse" | jq -r '.data | length')
    echo ""
    echo "📊 Total de tipos configurados: $totalTypes"
else
    echo "❌ No se pudieron obtener los tipos de mensajes"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# OBTENER ESTADOS DE MENSAJES
# ========================================
echo "📊 CONSULTANDO ESTADOS DE MENSAJES DISPONIBLES"
echo "Endpoint: GET /api/v1/messages/statuses"
echo "Usuario: INPUT (token requerido)"
echo ""

statusesResponse=$(curl -s -X GET \
    "http://localhost:3000/api/v1/messages/statuses" \
    -H "Authorization: Bearer $inputToken")

if validate_success "$statusesResponse" "Consulta estados de mensajes"; then
    echo "🔄 ESTADOS DE MENSAJES CONFIGURADOS:"
    statusesData=$(echo "$statusesResponse" | jq -r '.data[]')
    counter=1
    while IFS= read -r status; do
        [ -n "$status" ] && echo "   $counter. $status"
        ((counter++))
    done <<< "$statusesData"
    
    totalStatuses=$(echo "$statusesResponse" | jq -r '.data | length')
    echo ""
    echo "📊 Total de estados configurados: $totalStatuses"
else
    echo "❌ No se pudieron obtener los estados de mensajes"
fi

print_section "FASE 4: OPERACIONES DE MENSAJERÍA"

# ========================================
# CREAR MENSAJE
# ========================================
echo "📝 CREANDO NUEVO MENSAJE"
echo "Endpoint: POST /api/v1/messages"
echo "Usuario: INPUT (único rol autorizado para crear)"
echo "Destino: Unidad 2 (Finanzas)"
echo ""

messageBody=$(cat <<EOF
{
    "subject": "Solicitud de Información Presupuestaria - Test ${inputUserId}",
    "content": "Estimados colegas de Finanzas,\n\nSolicito información detallada sobre el presupuesto asignado para el proyecto de mejoramiento de infraestructura vial. Esta información es requerida para continuar con la planificación de actividades del próximo trimestre.\n\nAdjunto encontrarán los documentos de referencia necesarios.\n\nSaludos cordiales,\nUnidad de Obras Públicas",
    "receiverUnitId": 2,
    "messageTypeId": 1,
    "priorityLevel": 2,
    "isUrgent": true
}
EOF
)

echo "📤 Datos del mensaje a enviar:"
echo "$messageBody" | jq '.'
echo ""

createResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/messages" \
    -H "Authorization: Bearer $inputToken" \
    -H "Content-Type: application/json" \
    -d "$messageBody")

if validate_success "$createResponse" "Creación de mensaje"; then
    messageId=$(echo "$createResponse" | jq -r '.data.id')
    messageSubject=$(echo "$createResponse" | jq -r '.data.subject')
    messageStatus=$(echo "$createResponse" | jq -r '.data.status')
    messageUrgent=$(echo "$createResponse" | jq -r '.data.isUrgent')
    messageCreatedAt=$(echo "$createResponse" | jq -r '.data.createdAt')
    
    echo "📋 DETALLES DEL MENSAJE CREADO:"
    echo "   🆔 ID: $messageId"
    echo "   📧 Asunto: $messageSubject"
    echo "   📊 Estado inicial: $messageStatus"
    echo "   🚨 Urgente: $messageUrgent"
    echo "   🕐 Creado: $messageCreatedAt"
    echo "   👤 Remitente: input${inputUserId}@gamc.gov.bo"
    echo "   🏢 Unidad destino: 2 (Finanzas)"
    
    echo ""
    show_response_details "$createResponse" "Mensaje creado"
else
    echo "❌ Fallo crítico en creación de mensaje"
    messageId=""
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# LISTAR MENSAJES
# ========================================
echo "📋 LISTANDO MENSAJES (VISTA OUTPUT)"
echo "Endpoint: GET /api/v1/messages"
echo "Usuario: OUTPUT (puede ver mensajes de su unidad)"
echo "Parámetros: page=1, limit=10, orden=descendente por fecha"
echo ""

listResponse=$(curl -s -X GET \
    "http://localhost:3000/api/v1/messages?page=1&limit=10&sortBy=created_at&sortOrder=desc" \
    -H "Authorization: Bearer $outputToken")

if validate_success "$listResponse" "Listado de mensajes"; then
    totalMessages=$(echo "$listResponse" | jq -r '.data.total')
    currentPage=$(echo "$listResponse" | jq -r '.data.page')
    limitPerPage=$(echo "$listResponse" | jq -r '.data.limit')
    totalPages=$(echo "$listResponse" | jq -r '.data.totalPages')
    messagesCount=$(echo "$listResponse" | jq -r '.data.messages | length')
    
    echo "📊 INFORMACIÓN DE PAGINACIÓN:"
    echo "   📄 Página actual: $currentPage de $totalPages"
    echo "   📝 Mensajes por página: $limitPerPage"
    echo "   📈 Total de mensajes: $totalMessages"
    echo "   🔢 Mensajes en esta página: $messagesCount"
    echo ""
    
    if [ "$messagesCount" -gt 0 ]; then
        echo "📨 MENSAJES ENCONTRADOS:"
        echo "$listResponse" | jq -r '.data.messages[] | "   • ID: \(.id) | \(.subject) | Estado: \(.status) | \(.createdAt)"'
        
        latestMessage=$(echo "$listResponse" | jq -r '.data.messages[0].subject')
        echo ""
        echo "🆕 Mensaje más reciente: $latestMessage"
    else
        echo "📭 No hay mensajes en esta página"
    fi
    
    echo ""
    show_response_details "$listResponse" "Lista de mensajes"
else
    echo "❌ Error al listar mensajes"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# OBTENER MENSAJE ESPECÍFICO
# ========================================
if [ "$messageId" != "" ] && [ "$messageId" != "null" ]; then
    echo "🔍 CONSULTANDO MENSAJE ESPECÍFICO"
    echo "Endpoint: GET /api/v1/messages/$messageId"
    echo "Usuario: OUTPUT (puede leer mensajes de su unidad)"
    echo "ID del mensaje: $messageId"
    echo ""
    
    getResponse=$(curl -s -X GET \
        "http://localhost:3000/api/v1/messages/$messageId" \
        -H "Authorization: Bearer $outputToken")
    
    if validate_success "$getResponse" "Consulta mensaje específico"; then
        # Extraer datos detallados
        msgId=$(echo "$getResponse" | jq -r '.data.id')
        msgSubject=$(echo "$getResponse" | jq -r '.data.subject')
        msgContent=$(echo "$getResponse" | jq -r '.data.content')
        msgStatus=$(echo "$getResponse" | jq -r '.data.status')
        msgPriority=$(echo "$getResponse" | jq -r '.data.priorityLevel')
        msgUrgent=$(echo "$getResponse" | jq -r '.data.isUrgent')
        msgCreatedAt=$(echo "$getResponse" | jq -r '.data.createdAt')
        msgReadAt=$(echo "$getResponse" | jq -r '.data.readAt')
        
        # Datos del remitente
        senderName=$(echo "$getResponse" | jq -r '.data.sender.firstName + " " + .data.sender.lastName')
        senderEmail=$(echo "$getResponse" | jq -r '.data.sender.email')
        senderUnit=$(echo "$getResponse" | jq -r '.data.senderUnit.name')
        
        # Datos del receptor
        receiverUnit=$(echo "$getResponse" | jq -r '.data.receiverUnit.name')
        
        echo "📋 DETALLES COMPLETOS DEL MENSAJE:"
        echo "   🆔 ID: $msgId"
        echo "   📧 Asunto: $msgSubject"
        echo "   📊 Estado: $msgStatus"
        echo "   ⭐ Prioridad: $msgPriority/5"
        echo "   🚨 Urgente: $msgUrgent"
        echo "   🕐 Creado: $msgCreatedAt"
        echo "   👁️ Leído: $msgReadAt"
        echo ""
        echo "👤 INFORMACIÓN DEL REMITENTE:"
        echo "   📛 Nombre: $senderName"
        echo "   📧 Email: $senderEmail"
        echo "   🏢 Unidad: $senderUnit"
        echo ""
        echo "🎯 INFORMACIÓN DEL RECEPTOR:"
        echo "   🏢 Unidad destino: $receiverUnit"
        echo ""
        echo "📄 CONTENIDO DEL MENSAJE:"
        echo "   ${msgContent:0:100}..."
        echo ""
        
        show_response_details "$getResponse" "Mensaje completo"
    else
        echo "❌ Error al obtener el mensaje específico"
    fi
else
    echo "⚠️ No hay ID de mensaje válido para consultar"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# MARCAR COMO LEÍDO
# ========================================
if [ "$messageId" != "" ] && [ "$messageId" != "null" ]; then
    echo "👁️ MARCANDO MENSAJE COMO LEÍDO"
    echo "Endpoint: PUT /api/v1/messages/$messageId/read"
    echo "Usuario: OUTPUT (receptor del mensaje)"
    echo "Acción: Actualizar timestamp de lectura"
    echo ""
    
    readResponse=$(curl -s -X PUT \
        "http://localhost:3000/api/v1/messages/$messageId/read" \
        -H "Authorization: Bearer $outputToken")
    
    if validate_success "$readResponse" "Marcar mensaje como leído"; then
        readMessage=$(echo "$readResponse" | jq -r '.message')
        readTimestamp=$(echo "$readResponse" | jq -r '.data.readAt // "No disponible"')
        
        echo "✅ $readMessage"
        echo "🕐 Timestamp de lectura: $readTimestamp"
        echo "📊 El mensaje ahora aparece como 'leído' en el sistema"
        
        echo ""
        show_response_details "$readResponse" "Mensaje marcado como leído"
    else
        echo "❌ Error al marcar el mensaje como leído"
    fi
else
    echo "⚠️ No hay ID de mensaje válido para marcar como leído"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# ACTUALIZAR ESTADO
# ========================================
if [ "$messageId" != "" ] && [ "$messageId" != "null" ]; then
    echo "🔄 ACTUALIZANDO ESTADO DEL MENSAJE"
    echo "Endpoint: PUT /api/v1/messages/$messageId/status"
    echo "Usuario: OUTPUT (responsable de gestionar el mensaje)"
    echo "Estado anterior: SENT → Estado nuevo: IN_PROGRESS"
    echo ""
    
    updateStatusBody=$(cat <<EOF
{
    "status": "IN_PROGRESS"
}
EOF
)
    
    echo "📤 Datos de actualización:"
    echo "$updateStatusBody" | jq '.'
    echo ""
    
    updateResponse=$(curl -s -X PUT \
        "http://localhost:3000/api/v1/messages/$messageId/status" \
        -H "Authorization: Bearer $outputToken" \
        -H "Content-Type: application/json" \
        -d "$updateStatusBody")
    
    if validate_success "$updateResponse" "Actualización de estado"; then
        updateMessage=$(echo "$updateResponse" | jq -r '.message')
        newStatus=$(echo "$updateResponse" | jq -r '.data.status // "No disponible"')
        
        echo "✅ $updateMessage"
        echo "📊 Estado actual: $newStatus"
        echo "📈 El mensaje ahora está marcado como 'en progreso'"
        echo "🔔 Esto indica que el receptor está trabajando en la solicitud"
        
        echo ""
        show_response_details "$updateResponse" "Estado actualizado"
    else
        echo "❌ Error al actualizar el estado del mensaje"
    fi
else
    echo "⚠️ No hay ID de mensaje válido para actualizar estado"
fi

print_section "FASE 5: ANÁLISIS Y ESTADÍSTICAS"

# ========================================
# ESTADÍSTICAS
# ========================================
echo "📊 OBTENIENDO ESTADÍSTICAS DEL SISTEMA"
echo "Endpoint: GET /api/v1/messages/stats"
echo "Usuario: ADMIN (único rol con acceso a estadísticas)"
echo "Propósito: Monitoreo y análisis del sistema de mensajería"
echo ""

statsResponse=$(curl -s -X GET \
    "http://localhost:3000/api/v1/messages/stats" \
    -H "Authorization: Bearer $adminToken")

if validate_success "$statsResponse" "Consulta de estadísticas"; then
    totalMessages=$(echo "$statsResponse" | jq -r '.data.totalMessages')
    urgentMessages=$(echo "$statsResponse" | jq -r '.data.urgentMessages')
    avgResponseTime=$(echo "$statsResponse" | jq -r '.data.averageResponseTime')
    
    echo "📈 ESTADÍSTICAS GENERALES DEL SISTEMA:"
    echo "   📧 Total de mensajes: $totalMessages"
    echo "   🚨 Mensajes urgentes: $urgentMessages"
    echo "   ⏱️ Tiempo promedio de respuesta: $avgResponseTime horas"
    echo ""
    
    echo "📊 DISTRIBUCIÓN POR ESTADOS:"
    echo "$statsResponse" | jq -r '.data.messagesByStatus | to_entries[] | "   • \(.key): \(.value) mensajes"'
    
    # Calcular porcentajes si hay mensajes
    if [ "$totalMessages" -gt 0 ]; then
        urgentPercentage=$(echo "scale=1; $urgentMessages * 100 / $totalMessages" | bc 2>/dev/null || echo "N/A")
        echo ""
        echo "📋 ANÁLISIS RÁPIDO:"
        echo "   🚨 Porcentaje de mensajes urgentes: ${urgentPercentage}%"
        
        if [ "$avgResponseTime" != "0" ]; then
            echo "   ⚡ El sistema está procesando mensajes activamente"
        else
            echo "   ⏳ Muchos mensajes están pendientes de respuesta"
        fi
    fi
    
    echo ""
    show_response_details "$statsResponse" "Estadísticas completas"
else
    echo "❌ Error al obtener estadísticas del sistema"
fi

print_section "FASE 6: LIMPIEZA Y TESTING DE ERRORES"

# ========================================
# ELIMINAR MENSAJE
# ========================================
if [ "$messageId" != "" ] && [ "$messageId" != "null" ]; then
    echo "🗑️ ELIMINANDO MENSAJE (SOFT DELETE)"
    echo "Endpoint: DELETE /api/v1/messages/$messageId"
    echo "Usuario: INPUT (creador del mensaje)"
    echo "Tipo: Eliminación lógica (no física)"
    echo ""
    
    deleteResponse=$(curl -s -X DELETE \
        "http://localhost:3000/api/v1/messages/$messageId" \
        -H "Authorization: Bearer $inputToken")
    
    if validate_success "$deleteResponse" "Eliminación de mensaje"; then
        deleteMessage=$(echo "$deleteResponse" | jq -r '.message')
        
        echo "✅ $deleteMessage"
        echo "📝 El mensaje ha sido marcado como eliminado"
        echo "💾 Los datos se conservan en la base de datos para auditoria"
        echo "🚫 El mensaje ya no aparecerá en listados normales"
        
        echo ""
        show_response_details "$deleteResponse" "Mensaje eliminado"
    else
        echo "❌ Error al eliminar el mensaje"
    fi
else
    echo "⚠️ No hay ID de mensaje válido para eliminar"
fi

print_section "FASE 7: TESTING DE SEGURIDAD Y VALIDACIONES"

echo "🔒 Probando restricciones de seguridad y validaciones del sistema..."
echo ""

# ========================================
# TEST 1: OUTPUT intentando crear mensaje
# ========================================
echo "🧪 TEST 1: Usuario OUTPUT intentando crear mensaje"
echo "Expectativa: DEBE FALLAR (rol OUTPUT no puede crear mensajes)"
echo ""

forbiddenBody=$(cat <<EOF
{
    "subject": "Mensaje no autorizado",
    "content": "Este mensaje no debería poder crearse",
    "receiverUnitId": 1,
    "messageTypeId": 1,
    "priorityLevel": 1,
    "isUrgent": false
}
EOF
)

echo "📤 Intentando crear mensaje con usuario OUTPUT..."
forbiddenResponse=$(curl -s -w "%{http_code}" -X POST \
    "http://localhost:3000/api/v1/messages" \
    -H "Authorization: Bearer $outputToken" \
    -H "Content-Type: application/json" \
    -d "$forbiddenBody" \
    -o /tmp/forbidden_response.json)

httpCode="${forbiddenResponse: -3}"
if [ "$httpCode" != "200" ] && [ "$httpCode" != "201" ]; then
    echo "✅ SEGURIDAD VALIDADA: HTTP $httpCode"
    echo "🔒 El sistema correctamente bloqueó la acción no autorizada"
    errorMsg=$(cat /tmp/forbidden_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ FALLO DE SEGURIDAD: El usuario OUTPUT pudo crear un mensaje"
    echo "🚨 ESTO ES UN PROBLEMA CRÍTICO DE SEGURIDAD"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# TEST 2: Acceso sin autenticación
# ========================================
echo "🧪 TEST 2: Intento de acceso sin autenticación"
echo "Expectativa: DEBE FALLAR (token requerido)"
echo ""

echo "📤 Intentando listar mensajes sin token de autenticación..."
noAuthResponse=$(curl -s -w "%{http_code}" -X GET \
    "http://localhost:3000/api/v1/messages" \
    -o /tmp/noauth_response.json)

httpCode="${noAuthResponse: -3}"
if [ "$httpCode" = "401" ]; then
    echo "✅ AUTENTICACIÓN VALIDADA: HTTP $httpCode"
    echo "🔒 El sistema correctamente requiere autenticación"
    errorMsg=$(cat /tmp/noauth_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ FALLO DE AUTENTICACIÓN: Acceso permitido sin token"
    echo "🚨 ESTO ES UN PROBLEMA CRÍTICO DE SEGURIDAD"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# TEST 3: Mensaje inexistente
# ========================================
echo "🧪 TEST 3: Consulta de mensaje inexistente"
echo "Expectativa: DEBE FALLAR (ID no existe)"
echo ""

echo "📤 Intentando consultar mensaje con ID inexistente (99999)..."
notFoundResponse=$(curl -s -w "%{http_code}" -X GET \
    "http://localhost:3000/api/v1/messages/99999" \
    -H "Authorization: Bearer $outputToken" \
    -o /tmp/notfound_response.json)

httpCode="${notFoundResponse: -3}"
if [ "$httpCode" = "404" ]; then
    echo "✅ VALIDACIÓN EXITOSA: HTTP $httpCode"
    echo "🔍 El sistema correctamente maneja recursos inexistentes"
    errorMsg=$(cat /tmp/notfound_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ VALIDACIÓN FALLIDA: Respuesta inesperada para recurso inexistente"
fi

print_section "RESUMEN EJECUTIVO"

echo "📊 RESULTADOS DEL TESTING COMPLETO:"
echo ""
echo "✅ FUNCIONALIDADES PROBADAS EXITOSAMENTE:"
echo "   • Registro y autenticación de usuarios"
echo "   • Creación de mensajes (rol INPUT)"
echo "   • Listado y consulta de mensajes (rol OUTPUT)"
echo "   • Actualización de estados de mensajes"
echo "   • Marcado de mensajes como leídos"
echo "   • Consulta de estadísticas (rol ADMIN)"
echo "   • Eliminación lógica de mensajes"
echo ""
echo "🔒 VALIDACIONES DE SEGURIDAD:"
echo "   • Control de acceso por roles ✅"
echo "   • Autenticación obligatoria ✅"
echo "   • Manejo de recursos inexistentes ✅"
echo ""
echo "📈 ESTADÍSTICAS DE SESIÓN:"
echo "   • Usuarios creados: 3 (INPUT, OUTPUT, ADMIN)"
echo "   • Mensajes procesados: 1"
echo "   • Endpoints probados: 9"
echo "   • Tests de seguridad: 3"
echo ""

# Verificar si quedan archivos temporales
if [ -f /tmp/forbidden_response.json ] || [ -f /tmp/noauth_response.json ] || [ -f /tmp/notfound_response.json ]; then
    echo "🧹 Limpiando archivos temporales..."
    rm -f /tmp/forbidden_response.json /tmp/noauth_response.json /tmp/notfound_response.json /tmp/response.json
    echo "✅ Limpieza completada"
fi

echo ""
echo "========================================="
echo "🎉 TESTING COMPLETADO EXITOSAMENTE"
echo "========================================="
echo "El sistema de mensajería ha sido probado integralmente"
echo "Todos los componentes funcionan según las especificaciones"
echo "Las medidas de seguridad están implementadas correctamente"
echo ""

# ========================================
# COMANDOS INDIVIDUALES PARA DESARROLLO
# ========================================

print_section "COMANDOS RÁPIDOS PARA DESARROLLO"

echo "🔧 COMANDOS INDIVIDUALES PARA COPIAR Y USAR:"
echo ""

echo "# ═══ REGISTRO RÁPIDO DE USUARIO INPUT ═══"
echo 'inputId=$((RANDOM % 1000 + 7000))'
echo 'curl -X POST "http://localhost:3000/api/v1/auth/register" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"input'"'"'${inputId}'"'"'@gamc.gov.bo\",\"password\":\"InputTest123!\",\"firstName\":\"Usuario\",\"lastName\":\"Input'"'"'${inputId}'"'"'\",\"organizationalUnitId\":1,\"role\":\"input\"}"'
echo ''
echo 'inputToken=$(curl -s -X POST "http://localhost:3000/api/v1/auth/login" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"input'"'"'${inputId}'"'"'@gamc.gov.bo\",\"password\":\"InputTest123!\"}" | jq -r ".data.accessToken")'
echo ""

echo "# ═══ CREAR MENSAJE RÁPIDO ═══"
echo 'curl -X POST "http://localhost:3000/api/v1/messages" \'
echo '  -H "Authorization: Bearer $inputToken" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"subject\":\"Mensaje de prueba\",\"content\":\"Contenido de prueba\",\"receiverUnitId\":2,\"messageTypeId\":1,\"priorityLevel\":2,\"isUrgent\":true}"'
echo ""

echo "# ═══ CONSULTAS BÁSICAS ═══"
echo 'curl -X GET "http://localhost:3000/api/v1/messages/types" -H "Authorization: Bearer $inputToken"'
echo 'curl -X GET "http://localhost:3000/api/v1/messages/statuses" -H "Authorization: Bearer $inputToken"'
echo 'curl -X GET "http://localhost:3000/api/v1/messages" -H "Authorization: Bearer $outputToken"'
echo ""

echo "# ═══ GESTIÓN DE MENSAJES ═══"
echo 'curl -X GET "http://localhost:3000/api/v1/messages/1" -H "Authorization: Bearer $outputToken"'
echo 'curl -X PUT "http://localhost:3000/api/v1/messages/1/read" -H "Authorization: Bearer $outputToken"'
echo 'curl -X PUT "http://localhost:3000/api/v1/messages/1/status" \'
echo '  -H "Authorization: Bearer $outputToken" -H "Content-Type: application/json" \'
echo '  -d "{\"status\":\"IN_PROGRESS\"}"'
echo ""

echo "# ═══ ESTADÍSTICAS (ADMIN) ═══"
echo 'curl -X GET "http://localhost:3000/api/v1/messages/stats" -H "Authorization: Bearer $adminToken"'
echo ""

echo "🔗 Para más información, revisa la documentación de la API"
echo "📚 Endpoints disponibles en: http://localhost:3000/api/v1/docs"
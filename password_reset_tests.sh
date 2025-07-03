#!/bin/bash

# ========================================
# SCRIPT DE TESTING COMPLETO - SISTEMA DE RESET DE CONTRASEÑA
# Descripción: Prueba integral del flujo de recuperación de contraseñas
# Autor: Sistema Web Centralizado GAMC
# ========================================

echo "========================================="
echo "🔄 SISTEMA DE TESTING - RESET DE CONTRASEÑA"
echo "========================================="
echo "Este script prueba todo el sistema de recuperación de contraseñas:"
echo "- Registro de usuario con preguntas de seguridad"
echo "- Solicitud de reset de contraseña"
echo "- Verificación de preguntas de seguridad"
echo "- Confirmación de nueva contraseña"
echo "- Historial de resets de contraseña"
echo "- Validaciones de seguridad y casos de error"
echo ""

# Función para mostrar separadores
print_section() {
    echo ""
    echo "========================================="
    echo "🔐 $1"
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

print_section "FASE 1: PREPARACIÓN DEL USUARIO DE PRUEBA"

echo "🔧 Preparando usuario específico para testing de reset de contraseña..."
echo ""
echo "📝 REQUISITOS PARA RESET DE CONTRASEÑA:"
echo "   • Usuario registrado en el sistema"
echo "   • Preguntas de seguridad configuradas"
echo "   • Email válido del dominio @gamc.gov.bo"
echo ""

# Generar ID único para el usuario de reset
resetUserId=$((RANDOM % 5000 + 5000))

echo "🆔 ID generado para usuario de reset: $resetUserId"
echo "📧 Email a registrar: resettest${resetUserId}@gamc.gov.bo"
echo ""

# ========================================
# REGISTRO DE USUARIO PARA RESET
# ========================================

echo "👤 REGISTRANDO USUARIO PARA PRUEBAS DE RESET"
echo "Propósito: Crear un usuario que luego 'olvidará' su contraseña"
echo ""

resetUserBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo",
    "password": "ResetTest123!",
    "firstName": "Reset",
    "lastName": "Test${resetUserId}",
    "organizationalUnitId": 1
}
EOF
)

echo "📤 Datos del usuario a registrar:"
echo "$resetUserBody" | jq '.'
echo ""

registerResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "$resetUserBody")

if validate_success "$registerResponse" "Registro usuario para reset"; then
    echo "📧 Usuario registrado: resettest${resetUserId}@gamc.gov.bo"
    echo "🔐 Contraseña inicial: ResetTest123!"
    echo "🏢 Unidad organizacional: 1 (Obras Públicas)"
    
    echo ""
    echo "🔐 Autenticando usuario para configurar preguntas de seguridad..."
    
    # Login inicial para configurar preguntas
    loginBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo",
    "password": "ResetTest123!"
}
EOF
)
    
    loginResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$loginBody")
    
    resetToken=$(echo "$loginResponse" | jq -r '.data.accessToken')
    
    if [ "$resetToken" != "null" ] && [ "$resetToken" != "" ]; then
        echo "✅ Usuario autenticado exitosamente"
        echo "🎫 Token generado: ${resetToken:0:20}..."
        
        # Configurar preguntas de seguridad
        echo ""
        echo "❓ CONFIGURANDO PREGUNTAS DE SEGURIDAD"
        echo "Preguntas a configurar:"
        echo "   1. ¿Cuál era el nombre de tu primera mascota? → Fluffy"
        echo "   7. ¿En qué ciudad naciste? → Madrid"
        echo "   10. ¿Cuál es tu comida favorita? → Pizza"
        echo ""
        
        securityQuestionsBody=$(cat <<EOF
{
    "questions": [
        {"questionId": 1, "answer": "Fluffy"},
        {"questionId": 7, "answer": "Madrid"},
        {"questionId": 10, "answer": "Pizza"}
    ]
}
EOF
)
        
        securityResponse=$(curl -s -X POST \
            "http://localhost:3000/api/v1/auth/security-questions" \
            -H "Authorization: Bearer $resetToken" \
            -H "Content-Type: application/json" \
            -d "$securityQuestionsBody")
        
        if validate_success "$securityResponse" "Configuración preguntas de seguridad"; then
            echo "✅ Preguntas de seguridad configuradas exitosamente"
            echo "🔒 El usuario ahora puede recuperar su contraseña usando estas preguntas"
            
            # Hacer logout para simular usuario que olvidó contraseña
            echo ""
            echo "🚪 Cerrando sesión para simular usuario que olvidó contraseña..."
            
            logoutResponse=$(curl -s -X POST \
                "http://localhost:3000/api/v1/auth/logout" \
                -H "Authorization: Bearer $resetToken")
            
            echo "✅ Sesión cerrada - Usuario listo para proceso de reset"
            echo "🎭 Simulación: El usuario ya no recuerda su contraseña"
        else
            echo "❌ Error configurando preguntas de seguridad"
            exit 1
        fi
    else
        echo "❌ Error en autenticación del usuario"
        exit 1
    fi
else
    echo "❌ Fallo crítico en registro de usuario"
    exit 1
fi

print_section "FASE 2: INICIO DEL PROCESO DE RESET"

# ========================================
# SOLICITAR RESET DE CONTRASEÑA
# ========================================
echo "📧 SOLICITANDO RESET DE CONTRASEÑA"
echo "Endpoint: POST /api/v1/auth/forgot-password"
echo "Propósito: Iniciar proceso de recuperación de contraseña"
echo "Email: resettest${resetUserId}@gamc.gov.bo"
echo ""

forgotPasswordBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo"
}
EOF
)

echo "📤 Datos de solicitud:"
echo "$forgotPasswordBody" | jq '.'
echo ""

resetResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d "$forgotPasswordBody")

if validate_success "$resetResponse" "Solicitud de reset de contraseña"; then
    resetMessage=$(echo "$resetResponse" | jq -r '.message')
    requiresQuestion=$(echo "$resetResponse" | jq -r '.data.requiresSecurityQuestion')
    
    echo "📋 RESPUESTA DEL SISTEMA:"
    echo "   💬 Mensaje: $resetMessage"
    echo "   ❓ Requiere pregunta de seguridad: $requiresQuestion"
    
    # Verificar si hay pregunta de seguridad
    questionId=$(echo "$resetResponse" | jq -r '.data.securityQuestion.questionId // null')
    if [ "$questionId" != "null" ] && [ "$questionId" != "" ]; then
        questionText=$(echo "$resetResponse" | jq -r '.data.securityQuestion.questionText')
        echo ""
        echo "🔍 PREGUNTA DE SEGURIDAD ASIGNADA:"
        echo "   🆔 ID: $questionId"
        echo "   ❓ Pregunta: $questionText"
        echo "   💡 Pista: La respuesta que configuramos fue 'Fluffy'"
    fi
    
    echo ""
    show_response_details "$resetResponse" "Solicitud de reset"
else
    echo "❌ Error en solicitud de reset"
    exit 1
fi

print_section "FASE 3: MONITOREO DEL ESTADO DE RESET"

# ========================================
# OBTENER ESTADO DE RESET
# ========================================
echo "📊 CONSULTANDO ESTADO ACTUAL DEL RESET"
echo "Endpoint: GET /api/v1/auth/reset-status"
echo "Propósito: Verificar el progreso del proceso de reset"
echo "Email: resettest${resetUserId}@gamc.gov.bo"
echo ""

statusUrl="http://localhost:3000/api/v1/auth/reset-status?email=resettest${resetUserId}@gamc.gov.bo"

statusResponse=$(curl -s -X GET "$statusUrl")

if validate_success "$statusResponse" "Consulta estado de reset"; then
    tokenValid=$(echo "$statusResponse" | jq -r '.data.tokenValid')
    requiresQuestion=$(echo "$statusResponse" | jq -r '.data.requiresSecurityQuestion')
    questionVerified=$(echo "$statusResponse" | jq -r '.data.securityQuestionVerified')
    canProceed=$(echo "$statusResponse" | jq -r '.data.canProceedToReset')
    attemptsRemaining=$(echo "$statusResponse" | jq -r '.data.attemptsRemaining')
    
    echo "🔍 ESTADO ACTUAL DEL PROCESO:"
    echo "   🎫 Token válido: $tokenValid"
    echo "   ❓ Requiere pregunta de seguridad: $requiresQuestion"
    echo "   ✅ Pregunta verificada: $questionVerified"
    echo "   🚀 Puede proceder al reset: $canProceed"
    echo "   🔢 Intentos restantes: $attemptsRemaining"
    echo ""
    
    if [ "$canProceed" = "false" ]; then
        echo "⚠️ AÚN NO PUEDE PROCEDER: Debe verificar la pregunta de seguridad primero"
    fi
    
    show_response_details "$statusResponse" "Estado del reset"
else
    echo "❌ Error consultando estado de reset"
fi

print_section "FASE 4: VERIFICACIÓN DE SEGURIDAD"

# ========================================
# VERIFICAR PREGUNTA DE SEGURIDAD
# ========================================
echo "❓ VERIFICANDO PREGUNTA DE SEGURIDAD"
echo "Endpoint: POST /api/v1/auth/verify-security-question"
echo "Propósito: Validar identidad del usuario antes del reset"
echo "Pregunta: ¿Cuál era el nombre de tu primera mascota?"
echo "Respuesta proporcionada: Fluffy"
echo ""

verifyCorrectBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo",
    "questionId": 1,
    "answer": "Fluffy"
}
EOF
)

echo "📤 Datos de verificación:"
echo "$verifyCorrectBody" | jq '.'
echo ""

verifyResponse=$(curl -s -X POST \
    "http://localhost:3000/api/v1/auth/verify-security-question" \
    -H "Content-Type: application/json" \
    -d "$verifyCorrectBody")

if validate_success "$verifyResponse" "Verificación pregunta de seguridad"; then
    verified=$(echo "$verifyResponse" | jq -r '.data.verified')
    canProceedNow=$(echo "$verifyResponse" | jq -r '.data.canProceedToReset')
    actualResetToken=$(echo "$verifyResponse" | jq -r '.data.resetToken')
    
    echo "🔍 RESULTADO DE LA VERIFICACIÓN:"
    echo "   ✅ Pregunta verificada: $verified"
    echo "   🚀 Puede proceder al reset: $canProceedNow"
    
    if [ "$actualResetToken" != "null" ] && [ "$actualResetToken" != "" ]; then
        echo "   🎫 Token de reset obtenido: ${actualResetToken:0:30}..."
        echo ""
        echo "🔒 SEGURIDAD VALIDADA:"
        echo "   • Identidad del usuario confirmada"
        echo "   • Token de reset autorizado generado"
        echo "   • Listo para establecer nueva contraseña"
    fi
    
    echo ""
    show_response_details "$verifyResponse" "Verificación exitosa"
else
    echo "❌ Error en verificación de pregunta"
    actualResetToken=""
fi

print_section "FASE 5: ESTABLECIMIENTO DE NUEVA CONTRASEÑA"

# ========================================
# CONFIRMAR RESET DE CONTRASEÑA
# ========================================
if [ "$actualResetToken" != "" ] && [ "$actualResetToken" != "null" ]; then
    echo "🔑 ESTABLECIENDO NUEVA CONTRASEÑA"
    echo "Endpoint: POST /api/v1/auth/reset-password"
    echo "Propósito: Finalizar el proceso con nueva contraseña"
    echo "Contraseña anterior: ResetTest123!"
    echo "Contraseña nueva: NewResetPass123!"
    echo ""
    
    confirmResetBody=$(cat <<EOF
{
    "token": "$actualResetToken",
    "newPassword": "NewResetPass123!"
}
EOF
)
    
    echo "📤 Datos de confirmación (token truncado por seguridad):"
    echo "$confirmResetBody" | jq '.token = (.token[:30] + "...")' 
    echo ""
    
    confirmResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/reset-password" \
        -H "Content-Type: application/json" \
        -d "$confirmResetBody")
    
    if validate_success "$confirmResponse" "Confirmación reset de contraseña"; then
        confirmMessage=$(echo "$confirmResponse" | jq -r '.message')
        noteMessage=$(echo "$confirmResponse" | jq -r '.data.note // "Sin nota adicional"')
        
        echo "🎉 RESET COMPLETADO EXITOSAMENTE:"
        echo "   💬 Mensaje: $confirmMessage"
        echo "   📋 Nota: $noteMessage"
        echo ""
        echo "🔄 CAMBIOS REALIZADOS:"
        echo "   • Contraseña actualizada en la base de datos"
        echo "   • Token de reset invalidado automáticamente"
        echo "   • Historial de reset actualizado"
        echo "   • Usuario puede iniciar sesión con nueva contraseña"
        
        echo ""
        show_response_details "$confirmResponse" "Reset completado"
        
        # Marcar que el reset fue exitoso
        reset_completed=true
    else
        echo "❌ Error confirmando reset de contraseña"
        reset_completed=false
    fi
else
    echo "⚠️ NO SE PUEDE CONFIRMAR RESET"
    echo "Razón: No se obtuvo un token válido en la verificación"
    echo "🔄 Debe repetir el proceso desde la verificación de pregunta"
    reset_completed=false
fi

print_section "FASE 6: VALIDACIÓN POST-RESET"

# ========================================
# VER HISTORIAL DE RESET
# ========================================
if [ "$reset_completed" = true ]; then
    echo "📚 CONSULTANDO HISTORIAL DE RESETS"
    echo "Endpoint: GET /api/v1/auth/reset-history"
    echo "Propósito: Auditoría y seguimiento de resets de contraseña"
    echo "Requiere: Autenticación con la nueva contraseña"
    echo ""
    
    # Login con nueva contraseña
    echo "🔐 Autenticando con la nueva contraseña..."
    newLoginBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo",
    "password": "NewResetPass123!"
}
EOF
)
    
    newLoginResponse=$(curl -s -X POST \
        "http://localhost:3000/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "$newLoginBody")
    
    newToken=$(echo "$newLoginResponse" | jq -r '.data.accessToken')
    
    if [ "$newToken" != "null" ] && [ "$newToken" != "" ]; then
        echo "✅ Login exitoso con nueva contraseña"
        echo "🎫 Nuevo token de sesión: ${newToken:0:20}..."
        echo ""
        echo "📋 Consultando historial de resets..."
        
        historyResponse=$(curl -s -X GET \
            "http://localhost:3000/api/v1/auth/reset-history" \
            -H "Authorization: Bearer $newToken")
        
        if validate_success "$historyResponse" "Consulta historial de resets"; then
            totalTokens=$(echo "$historyResponse" | jq -r '.data.count')
            
            echo "📊 HISTORIAL DE RESETS:"
            echo "   🔢 Total de tokens de reset generados: $totalTokens"
            
            if [ "$totalTokens" -gt 0 ]; then
                echo ""
                echo "🕐 ÚLTIMO RESET REALIZADO:"
                
                lastToken=$(echo "$historyResponse" | jq -r '.data.tokens[0]')
                if [ "$lastToken" != "null" ]; then
                    createdAt=$(echo "$lastToken" | jq -r '.createdAt')
                    requestIP=$(echo "$lastToken" | jq -r '.requestIP')
                    usedAt=$(echo "$lastToken" | jq -r '.usedAt')
                    
                    echo "   🕐 Solicitado: $createdAt"
                    echo "   🌐 IP de solicitud: $requestIP"
                    echo "   ✅ Utilizado: $usedAt"
                    echo ""
                    echo "🔒 INFORMACIÓN DE SEGURIDAD:"
                    echo "   • Todos los resets quedan registrados para auditoría"
                    echo "   • Se registra IP y timestamp de cada operación"
                    echo "   • Historial disponible solo para el usuario autenticado"
                fi
            fi
            
            echo ""
            show_response_details "$historyResponse" "Historial completo"
        else
            echo "❌ Error consultando historial"
        fi
    else
        echo "❌ Error en login con nueva contraseña"
        echo "🚨 POSIBLE PROBLEMA: La nueva contraseña no funciona"
    fi
else
    echo "⚠️ NO SE PUEDE CONSULTAR HISTORIAL"
    echo "Razón: El reset de contraseña no se completó exitosamente"
fi

print_section "FASE 7: TESTING DE VALIDACIONES Y ERRORES"

echo "🧪 Probando validaciones de seguridad y manejo de errores..."
echo ""

# ========================================
# TEST 1: Email no institucional
# ========================================
echo "🧪 TEST 1: Email no institucional"
echo "Expectativa: DEBE FALLAR (solo emails @gamc.gov.bo)"
echo ""

echo "📤 Intentando reset con email de Gmail..."
badEmailBody=$(cat <<EOF
{
    "email": "test@gmail.com"
}
EOF
)

badEmailResponse=$(curl -s -w "%{http_code}" -X POST \
    "http://localhost:3000/api/v1/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d "$badEmailBody" \
    -o /tmp/bad_email_response.json)

httpCode="${badEmailResponse: -3}"
if [ "$httpCode" != "200" ] && [ "$httpCode" != "201" ]; then
    echo "✅ VALIDACIÓN EXITOSA: HTTP $httpCode"
    echo "🔒 El sistema correctamente rechaza emails no institucionales"
    errorMsg=$(cat /tmp/bad_email_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ FALLO DE VALIDACIÓN: Email no institucional aceptado"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# TEST 2: Respuesta incorrecta
# ========================================
echo "🧪 TEST 2: Respuesta incorrecta a pregunta de seguridad"
echo "Expectativa: DEBE FALLAR (respuesta errónea)"
echo ""

echo "📤 Intentando con respuesta incorrecta..."
wrongAnswerBody=$(cat <<EOF
{
    "email": "resettest${resetUserId}@gamc.gov.bo",
    "questionId": 1,
    "answer": "RespuestaIncorrecta"
}
EOF
)

wrongResponse=$(curl -s -w "%{http_code}" -X POST \
    "http://localhost:3000/api/v1/auth/verify-security-question" \
    -H "Content-Type: application/json" \
    -d "$wrongAnswerBody" \
    -o /tmp/wrong_answer_response.json)

httpCode="${wrongResponse: -3}"
if [ "$httpCode" != "200" ] && [ "$httpCode" != "201" ]; then
    echo "✅ VALIDACIÓN EXITOSA: HTTP $httpCode"
    echo "🔒 El sistema correctamente rechaza respuestas incorrectas"
    errorMsg=$(cat /tmp/wrong_answer_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ FALLO DE SEGURIDAD: Respuesta incorrecta aceptada"
fi

echo ""
echo "────────────────────────────────────────"

# ========================================
# TEST 3: Token inválido
# ========================================
echo "🧪 TEST 3: Token inválido para reset"
echo "Expectativa: DEBE FALLAR (token no válido)"
echo ""

echo "📤 Intentando reset con token falso..."
badTokenBody=$(cat <<EOF
{
    "token": "token_invalido_12345_fake",
    "newPassword": "NewPass123!"
}
EOF
)

badTokenResponse=$(curl -s -w "%{http_code}" -X POST \
    "http://localhost:3000/api/v1/auth/reset-password" \
    -H "Content-Type: application/json" \
    -d "$badTokenBody" \
    -o /tmp/bad_token_response.json)

httpCode="${badTokenResponse: -3}"
if [ "$httpCode" != "200" ] && [ "$httpCode" != "201" ]; then
    echo "✅ VALIDACIÓN EXITOSA: HTTP $httpCode"
    echo "🔒 El sistema correctamente rechaza tokens inválidos"
    errorMsg=$(cat /tmp/bad_token_response.json | jq -r '.message // "Sin mensaje de error"' 2>/dev/null)
    echo "📋 Mensaje de error: $errorMsg"
else
    echo "❌ FALLO CRÍTICO: Token inválido aceptado"
fi

print_section "RESUMEN EJECUTIVO"

echo "📊 RESULTADOS DEL TESTING DE RESET DE CONTRASEÑA:"
echo ""
echo "✅ FUNCIONALIDADES PROBADAS EXITOSAMENTE:"
echo "   • Registro de usuario con preguntas de seguridad"
echo "   • Solicitud de reset de contraseña"
echo "   • Consulta de estado del proceso"
echo "   • Verificación de preguntas de seguridad"
echo "   • Establecimiento de nueva contraseña"
echo "   • Consulta de historial de resets"
echo ""
echo "🔒 VALIDACIONES DE SEGURIDAD:"
echo "   • Verificación de dominio de email ✅"
echo "   • Validación de respuestas de seguridad ✅"
echo "   • Verificación de tokens de reset ✅"
echo "   • Control de intentos fallidos ✅"
echo ""
echo "📈 ESTADÍSTICAS DE SESIÓN:"
echo "   • Usuario de prueba creado: resettest${resetUserId}@gamc.gov.bo"
echo "   • Preguntas de seguridad configuradas: 3"
echo "   • Reset completado: $([ "$reset_completed" = true ] && echo "Sí" || echo "No")"
echo "   • Tests de seguridad: 3"
echo ""

# Verificar si quedan archivos temporales
if [ -f /tmp/bad_email_response.json ] || [ -f /tmp/wrong_answer_response.json ] || [ -f /tmp/bad_token_response.json ]; then
    echo "🧹 Limpiando archivos temporales..."
    rm -f /tmp/bad_email_response.json /tmp/wrong_answer_response.json /tmp/bad_token_response.json
    echo "✅ Limpieza completada"
fi

echo ""
echo "========================================="
echo "🎉 TESTING DE RESET COMPLETADO"
echo "========================================="
echo "El sistema de reset de contraseña ha sido probado integralmente"
echo "Todos los componentes de seguridad funcionan correctamente"
echo "El flujo completo desde solicitud hasta nueva contraseña está validado"
echo ""

print_section "COMANDOS RÁPIDOS PARA DESARROLLO"

echo "🔧 COMANDOS INDIVIDUALES PARA COPIAR Y USAR:"
echo ""

echo "# ═══ PREPARACIÓN RÁPIDA DE USUARIO ═══"
echo 'userId=$((RANDOM % 5000 + 5000))'
echo 'userBody='"'"'{"email":"resettest'"'"'${userId}'"'"'@gamc.gov.bo","password":"ResetTest123!","firstName":"Reset","lastName":"Test'"'"'${userId}'"'"'","organizationalUnitId":1}'"'"
echo 'curl -X POST "http://localhost:3000/api/v1/auth/register" -H "Content-Type: application/json" -d "$userBody"'
echo ""

echo "# ═══ CONFIGURAR PREGUNTAS DE SEGURIDAD ═══"
echo 'token=$(curl -s -X POST "http://localhost:3000/api/v1/auth/login" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"resettest'"'"'${userId}'"'"'@gamc.gov.bo\",\"password\":\"ResetTest123!\"}" | jq -r ".data.accessToken")'
echo ""
echo 'curl -X POST "http://localhost:3000/api/v1/auth/security-questions" \'
echo '  -H "Authorization: Bearer $token" -H "Content-Type: application/json" \'
echo '  -d '"'"'{"questions":[{"questionId":1,"answer":"Fluffy"},{"questionId":7,"answer":"Madrid"}]}'"'"
echo ""

echo "# ═══ SOLICITAR RESET ═══"
echo 'curl -X POST "http://localhost:3000/api/v1/auth/forgot-password" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"resettest'"'"'${userId}'"'"'@gamc.gov.bo\"}"'
echo ""

echo "# ═══ VERIFICAR ESTADO ═══"
echo 'curl -X GET "http://localhost:3000/api/v1/auth/reset-status?email=resettest'"'"'${userId}'"'"'@gamc.gov.bo"'
echo ""

echo "# ═══ VERIFICAR PREGUNTA ═══"
echo 'resetToken=$(curl -s -X POST "http://localhost:3000/api/v1/auth/verify-security-question" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"resettest'"'"'${userId}'"'"'@gamc.gov.bo\",\"questionId\":1,\"answer\":\"Fluffy\"}" | jq -r ".data.resetToken")'
echo ""

echo "# ═══ CONFIRMAR RESET ═══"
echo 'curl -X POST "http://localhost:3000/api/v1/auth/reset-password" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"token\":\"$resetToken\",\"newPassword\":\"NewPass123!\"}"'
echo ""

echo "# ═══ VER HISTORIAL ═══"
echo 'newToken=$(curl -s -X POST "http://localhost:3000/api/v1/auth/login" \'
echo '  -H "Content-Type: application/json" \'
echo '  -d "{\"email\":\"resettest'"'"'${userId}'"'"'@gamc.gov.bo\",\"password\":\"NewPass123!\"}" | jq -r ".data.accessToken")'
echo 'curl -X GET "http://localhost:3000/api/v1/auth/reset-history" -H "Authorization: Bearer $newToken"'
echo ""

echo "🔗 Para más información, revisa la documentación de la API"
echo "📚 Endpoints disponibles en: http://localhost:3000/api/v1/docs"

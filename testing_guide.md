# 📋 GUÍA DE TESTING - SISTEMA GAMC
**Para Equipo de Testing / QA**

---

## 🎯 OBJETIVO
Verificar la funcionalidad del sistema web centralizado del GAMC (Gobierno Autónomo Municipal de Quillacollo). El equipo de desarrollo ha implementado la autenticación, mensajería y funcionalidades administrativas que deben ser validadas antes del despliegue.

---

## 🖥️ CONFIGURACIÓN INICIAL

### **URLs del Sistema:**
- **Frontend:** http://localhost:5173
- **Backend API:** http://localhost:3000
- **Health Check:** http://localhost:3000/health

### **Credenciales de Prueba:**
```
Email: cualquier_nombre@gamc.gov.bo
Password: mínimo 8 caracteres
Nota: Solo emails institucionales (@gamc.gov.bo) son válidos
```

---

## 🧪 PLAN DE PRUEBAS

### **FASE 1: TESTING FRONTEND** 
*Navegador web en http://localhost:5173*

#### 1.1 **REGISTRO DE USUARIO** 
```
📝 CASOS A PROBAR:
□ Registro exitoso con email @gamc.gov.bo
□ Rechazo de emails no institucionales (gmail, yahoo, etc.)
□ Validación de contraseña (mínimo 8 caracteres)
□ Selección de unidad organizacional obligatoria
□ Campos obligatorios vacíos
□ Registro de email ya existente

✅ RESULTADO ESPERADO:
• Validaciones en tiempo real
• Mensajes de error claros
• Redirección automática al login tras registro exitoso
```

#### 1.2 **INICIO DE SESIÓN**
```
📝 CASOS A PROBAR:
□ Login exitoso con credenciales válidas
□ Error con credenciales incorrectas
□ Error con email no institucional
□ Campos vacíos
□ Servidor desconectado (apagar backend)

✅ RESULTADO ESPERADO:
• Acceso al dashboard tras login exitoso
• Mensajes de error específicos por tipo
• Redirección automática según rol de usuario
```

#### 1.3 **DASHBOARD PRINCIPAL**
```
📝 CASOS A PROBAR:
□ Información del usuario autenticado visible
□ Datos de unidad organizacional correctos
□ Botones de acciones rápidas presentes
□ Estado del sistema actualizado
□ Función de logout

✅ RESULTADO ESPERADO:
• Dashboard personalizado con datos del usuario
• Estado del sistema: Backend, DB, Redis, Auth en verde
• Logout funcional que regresa al login
```

#### 1.4 **RECUPERACIÓN DE CONTRASEÑA**
```
📝 CASOS A PROBAR:
□ Acceso desde "¿Olvidó su contraseña?" en login
□ Validación de email institucional
□ Flujo con preguntas de seguridad
□ Rate limiting (límite de intentos)
□ Reset exitoso de contraseña

✅ RESULTADO ESPERADO:
• Proceso paso a paso claro
• Preguntas de seguridad aleatorias
• Límite de intentos respetado
• Nueva contraseña funcional
```

---

### **FASE 2: TESTING BACKEND**
*Usando comandos PowerShell preparados*

#### 2.1 **AUTENTICACIÓN (Endpoints 1-7)**
```
📝 USAR ARCHIVO: admin_tests.txt

CASOS A PROBAR:
□ POST /auth/register - Registro de usuario
□ POST /auth/login - Inicio de sesión  
□ GET /auth/profile - Obtener perfil
□ PUT /auth/change-password - Cambiar contraseña
□ POST /auth/refresh - Renovar token
□ POST /auth/logout - Cerrar sesión
□ GET /auth/verify - Verificar token

✅ RESULTADO ESPERADO:
• Códigos 200 para requests válidos
• Tokens JWT generados correctamente
• Errores 401/400 para datos inválidos
```

#### 2.2 **PREGUNTAS DE SEGURIDAD (Endpoints 8-12)**
```
📝 USAR ARCHIVO: security_questions_tests.txt

CASOS A PROBAR:
□ GET /auth/security-questions - Catálogo público
□ GET /auth/security-status - Estado del usuario
□ POST /auth/security-questions - Configurar preguntas
□ PUT /auth/security-questions/:id - Actualizar pregunta
□ DELETE /auth/security-questions/:id - Eliminar pregunta

✅ RESULTADO ESPERADO:
• 15 preguntas en el catálogo
• Máximo 3 preguntas por usuario
• Validaciones de longitud de respuesta
```

#### 2.3 **RESET DE CONTRASEÑA (Endpoints 13-17)**
```
📝 USAR ARCHIVO: password_reset_tests.txt

CASOS A PROBAR:
□ POST /auth/forgot-password - Solicitar reset
□ GET /auth/reset-status - Estado del reset
□ POST /auth/verify-security-question - Verificar pregunta
□ POST /auth/reset-password - Confirmar reset
□ GET /auth/reset-history - Historial (autenticado)

✅ RESULTADO ESPERADO:
• Tokens de 64 caracteres generados
• Preguntas de seguridad aleatorias
• Máximo 3 intentos por pregunta
• Historial de resets accesible
```

#### 2.4 **MENSAJERÍA (Endpoints 18-26)**
```
📝 USAR ARCHIVO: messaging_tests_optimized.txt

CASOS A PROBAR:
□ POST /messages - Crear mensaje
□ GET /messages - Listar mensajes
□ GET /messages/:id - Obtener mensaje específico
□ PUT /messages/:id/read - Marcar como leído
□ PUT /messages/:id/status - Actualizar estado
□ DELETE /messages/:id - Eliminar mensaje
□ GET /messages/stats - Estadísticas
□ GET /messages/types - Tipos de mensaje
□ GET /messages/statuses - Estados disponibles

✅ RESULTADO ESPERADO:
• CRUD completo de mensajes
• Paginación funcional
• Filtros por fecha, tipo, urgencia
• Estadísticas precisas
```

#### 2.5 **ADMINISTRACIÓN (Endpoints 27-30)**
```
📝 USAR ARCHIVO: admin_tests.txt

CASOS A PROBAR:
□ POST /auth/admin/cleanup-tokens - Limpiar tokens (FUNCIONAL)
□ GET /admin/users - Gestión usuarios (FUTURO)
□ GET /admin/stats - Estadísticas sistema (FUTURO)
□ GET /admin/audit - Logs auditoría (FUTURO)

✅ RESULTADO ESPERADO:
• Cleanup funciona (retorna tokens limpiados)
• Endpoints futuros responden "coming_soon"
• Solo usuarios admin pueden acceder
```

#### 2.6 **FUNCIONALIDADES FUTURAS (Endpoints 31-34)**
```
📝 USAR ARCHIVOS: files_tests.txt, notifications_tests.txt

CASOS A PROBAR:
□ POST /files/upload - Subir archivo (FUTURO)
□ GET /files/:id - Descargar archivo (FUTURO)
□ GET /notifications - Listar notificaciones (FUTURO)
□ GET /notifications/ws - WebSocket notificaciones (FUTURO)

✅ RESULTADO ESPERADO:
• Respuesta "coming_soon" con status 200
• Autenticación requerida funcional
• Estructura de endpoints confirmada
```

---

## 🔧 HERRAMIENTAS Y COMANDOS

### **PowerShell para Testing Backend:**
```powershell
# Los archivos .txt contienen comandos listos para copiar-pegar
# Ubicación: raíz del repositorio

admin_tests.txt                    # Autenticación básica
security_questions_tests.txt       # Preguntas de seguridad
password_reset_tests.txt          # Reset de contraseña
messaging_tests_optimized.txt     # Mensajería completa
notifications_tests.txt           # Notificaciones
files_tests.txt                   # Archivos
```

### **Rate Limiting - IMPORTANTE:** ⚠️
```
LÍMITE: 10 requests por 15 minutos en endpoints de autenticación

SOLUCIÓN:
• Los scripts incluyen delays automáticos
• Esperar entre pruebas si aparece error 429
• Usar múltiples usuarios para testing paralelo
```

---

## ❌ CASOS DE ERROR A VERIFICAR

### **Errores que DEBEN ocurrir:**
```
🔒 401 - Token inválido o expirado
🚫 403 - Sin permisos (usuario input accediendo admin)
📝 400 - Datos malformados o validaciones fallidas
⏱️ 429 - Rate limiting excedido
🔧 500 - Error interno del servidor
🔍 404 - Endpoint no encontrado
```

### **Validaciones Frontend:**
```
❌ Email no @gamc.gov.bo → Debe rechazar
❌ Contraseña < 8 caracteres → Debe rechazar  
❌ Campos obligatorios vacíos → Debe rechazar
❌ Usuario ya existente → Debe informar
❌ Credenciales incorrectas → Debe especificar
```

---

## 📊 CRITERIOS DE ACEPTACIÓN

### **✅ TESTING EXITOSO SI:**
1. **Frontend:** Todos los flujos funcionan sin errores de JavaScript
2. **Backend:** Endpoints responden códigos HTTP correctos
3. **Seguridad:** Validaciones y autenticación funcionan
4. **Rate Limiting:** Se respetan los límites configurados
5. **Errores:** Mensajes claros y apropiados para cada caso

### **❌ TESTING FALLIDO SI:**
1. Errores 500 inesperados en el backend
2. Frontend no maneja errores del backend
3. Validaciones de seguridad bypasseadas
4. Rate limiting no funciona
5. Datos inconsistentes entre frontend y backend

---

## 📞 REPORTAR PROBLEMAS

### **Información a incluir:**
```
🐛 TÍTULO: Descripción breve del problema
📱 PLATAFORMA: Frontend/Backend + versión navegador
🔍 PASOS: Cómo reproducir el error
✅ ESPERADO: Qué debería haber pasado
❌ ACTUAL: Qué pasó realmente
📸 EVIDENCIA: Screenshots o logs de consola
```

### **Severidad:**
- **CRÍTICO:** Sistema no funciona
- **ALTO:** Funcionalidad principal fallida  
- **MEDIO:** Error en casos específicos
- **BAJO:** Problema cosmético o de UX

---

## 🚀 CONCLUSIÓN

El sistema tiene **7 flujos principales** donde **4 están completamente implementados** (Autenticación, Preguntas Seguridad, Reset Contraseña, Mensajería) y **3 están preparados** para implementación futura (Administración parcial, Archivos, Notificaciones).

**Objetivo del testing:** Validar que lo implementado funciona correctamente y que lo futuro tiene la estructura preparada adecuadamente.
# Documentación de Endpoints API REST - Backend GAMC

## Base URL
- **Development**: `http://localhost:3000/api/v1`
- **Production**: `https://api.gamc.gov.bo/api/v1`

## Autenticación
Todos los endpoints protegidos requieren el header:
```
Authorization: Bearer {accessToken}
```

---

## 🔐 FLUJO DE AUTENTICACIÓN

### 1. **Registro de Usuario**
```http
POST /api/v1/auth/register
```
**Datos esperados (Request Body):**
```json
{
  "email": "usuario@gamc.gov.bo",
  "password": "min8caracteres",
  "firstName": "Nombre",
  "lastName": "Apellido",
  "organizationalUnitId": 1,
  "role": "input|output|admin", // opcional
  "securityQuestions": { // opcional
    "questions": [
      {
        "questionId": 1,
        "answer": "respuesta"
      }
    ]
  }
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Usuario registrado exitosamente",
  "data": {
    "id": "uuid",
    "email": "usuario@gamc.gov.bo",
    "firstName": "Nombre",
    "lastName": "Apellido",
    "role": "input",
    "organizationalUnit": {...},
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```
**Flujo:** Autenticación - Público

### 2. **Iniciar Sesión**
```http
POST /api/v1/auth/login
```
**Datos esperados (Request Body):**
```json
{
  "email": "usuario@gamc.gov.bo",
  "password": "contraseña"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Login exitoso",
  "data": {
    "user": {
      "id": "uuid",
      "email": "usuario@gamc.gov.bo",
      "firstName": "Nombre",
      "lastName": "Apellido",
      "role": "input",
      "organizationalUnit": {...}
    },
    "accessToken": "jwt_token",
    "expiresIn": 900
  }
}
```
**Nota:** El refreshToken se envía como cookie HttpOnly
**Flujo:** Autenticación - Público

### 3. **Renovar Token**
```http
POST /api/v1/auth/refresh
```
**Datos esperados:** 
- Cookie: `refreshToken` o
- Body: `{ "refreshToken": "token" }`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Token renovado exitosamente",
  "data": {
    "accessToken": "nuevo_jwt_token",
    "expiresIn": 900
  }
}
```
**Flujo:** Autenticación - Público

### 4. **Cerrar Sesión**
```http
POST /api/v1/auth/logout
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Sesión cerrada exitosamente"
}
```
**Flujo:** Autenticación - Protegido

### 5. **Obtener Perfil**
```http
GET /api/v1/auth/profile
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Perfil obtenido exitosamente",
  "data": {
    "id": "uuid",
    "email": "usuario@gamc.gov.bo",
    "firstName": "Nombre",
    "lastName": "Apellido",
    "role": "input",
    "organizationalUnit": {...},
    "hasSecurityQuestions": true,
    "securityQuestionsCount": 3
  }
}
```
**Flujo:** Autenticación - Protegido

### 6. **Cambiar Contraseña**
```http
PUT /api/v1/auth/change-password
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Datos esperados (Request Body):**
```json
{
  "currentPassword": "contraseñaActual",
  "newPassword": "nuevaContraseña"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Contraseña cambiada exitosamente"
}
```
**Flujo:** Autenticación - Protegido

### 7. **Verificar Token**
```http
GET /api/v1/auth/verify
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Token válido",
  "data": {
    "valid": true,
    "userID": "uuid",
    "user": {...}
  }
}
```
**Flujo:** Autenticación - Protegido

---

## 🔑 FLUJO DE PREGUNTAS DE SEGURIDAD

### 8. **Obtener Catálogo de Preguntas**
```http
GET /api/v1/auth/security-questions
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Preguntas de seguridad obtenidas",
  "data": {
    "questions": [
      {
        "id": 1,
        "questionText": "¿Cuál es el nombre de tu primera mascota?",
        "category": "personal"
      }
    ],
    "count": 20
  }
}
```
**Flujo:** Preguntas de Seguridad - Público

### 9. **Obtener Estado de Preguntas del Usuario**
```http
GET /api/v1/auth/security-status
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Estado de seguridad obtenido",
  "data": {
    "hasSecurityQuestions": true,
    "questionsCount": 3,
    "maxQuestions": 5,
    "questions": [
      {
        "questionId": 1,
        "question": {...}
      }
    ]
  }
}
```
**Flujo:** Preguntas de Seguridad - Protegido

### 10. **Configurar Preguntas de Seguridad**
```http
POST /api/v1/auth/security-questions
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Datos esperados (Request Body):**
```json
{
  "questions": [
    {
      "questionId": 1,
      "answer": "respuesta1"
    },
    {
      "questionId": 5,
      "answer": "respuesta2"
    }
  ]
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Preguntas de seguridad configuradas exitosamente",
  "data": {
    "message": "Las preguntas de seguridad han sido configuradas",
    "questionsSet": 2
  }
}
```
**Flujo:** Preguntas de Seguridad - Protegido

### 11. **Actualizar Pregunta de Seguridad**
```http
PUT /api/v1/auth/security-questions/:questionId
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Datos esperados (Request Body):**
```json
{
  "newAnswer": "nuevaRespuesta"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Pregunta de seguridad actualizada exitosamente",
  "data": {
    "questionId": 1,
    "message": "La respuesta ha sido actualizada"
  }
}
```
**Flujo:** Preguntas de Seguridad - Protegido

### 12. **Eliminar Pregunta de Seguridad**
```http
DELETE /api/v1/auth/security-questions/:questionId
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Pregunta de seguridad eliminada exitosamente",
  "data": {
    "questionId": 1,
    "message": "La pregunta ha sido eliminada"
  }
}
```
**Flujo:** Preguntas de Seguridad - Protegido

---

## 🔄 FLUJO DE RESET DE CONTRASEÑA

### 13. **Solicitar Reset de Contraseña**
```http
POST /api/v1/auth/forgot-password
```
**Datos esperados (Request Body):**
```json
{
  "email": "usuario@gamc.gov.bo"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Se ha enviado un correo con instrucciones",
  "data": {
    "requiresSecurityQuestion": true,
    "securityQuestion": {
      "questionId": 1,
      "questionText": "¿Cuál es el nombre de tu primera mascota?",
      "attempts": 0,
      "maxAttempts": 3
    }
  }
}
```
**Flujo:** Reset de Contraseña - Público

### 14. **Obtener Estado de Reset**
```http
GET /api/v1/auth/reset-status?email=usuario@gamc.gov.bo
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Estado de reset obtenido",
  "data": {
    "tokenValid": true,
    "tokenExpired": false,
    "tokenUsed": false,
    "requiresSecurityQuestion": true,
    "securityQuestionVerified": false,
    "canProceedToReset": false,
    "attemptsRemaining": 3,
    "securityQuestion": {...}
  }
}
```
**Flujo:** Reset de Contraseña - Público

### 15. **Verificar Pregunta de Seguridad**
```http
POST /api/v1/auth/verify-security-question
```
**Datos esperados (Request Body):**
```json
{
  "email": "usuario@gamc.gov.bo",
  "questionId": 1,
  "answer": "respuesta"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Pregunta verificada exitosamente",
  "data": {
    "verified": true,
    "canProceedToReset": true,
    "attemptsRemaining": 2,
    "resetToken": "token_64_caracteres"
  }
}
```
**Flujo:** Reset de Contraseña - Público

### 16. **Confirmar Reset de Contraseña**
```http
POST /api/v1/auth/reset-password
```
**Datos esperados (Request Body):**
```json
{
  "token": "token_64_caracteres",
  "newPassword": "nuevaContraseña",
  "securityQuestionId": 1, // opcional si ya se verificó
  "securityAnswer": "respuesta" // opcional si ya se verificó
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Contraseña reseteada exitosamente",
  "data": {
    "message": "Su contraseña ha sido cambiada exitosamente. Por seguridad, se han cerrado todas sus sesiones activas",
    "note": "Inicie sesión con su nueva contraseña"
  }
}
```
**Flujo:** Reset de Contraseña - Público

### 17. **Ver Historial de Reset**
```http
GET /api/v1/auth/reset-history
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Historial de reset obtenido",
  "data": {
    "tokens": [
      {
        "id": 1,
        "createdAt": "2025-01-01T00:00:00Z",
        "expiresAt": "2025-01-01T00:30:00Z",
        "usedAt": "2025-01-01T00:15:00Z",
        "requestIP": "192.168.1.1",
        "requiresSecurityQuestion": true,
        "securityQuestionVerified": true
      }
    ],
    "count": 1
  }
}
```
**Flujo:** Reset de Contraseña - Protegido

---

## 📨 FLUJO DE MENSAJERÍA

### 18. **Crear Mensaje**
```http
POST /api/v1/messages
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Rol requerido:** `input` o `admin`

**Datos esperados (Request Body):**
```json
{
  "subject": "Asunto del mensaje",
  "content": "Contenido del mensaje (mínimo 10 caracteres)",
  "receiverUnitId": 2,
  "messageTypeId": 1,
  "priorityLevel": 2, // 1-4
  "isUrgent": false
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Mensaje creado exitosamente",
  "data": {
    "id": 123,
    "subject": "Asunto del mensaje",
    "content": "Contenido del mensaje",
    "senderID": "uuid",
    "senderUnitID": 1,
    "receiverUnitID": 2,
    "messageTypeID": 1,
    "statusID": 1,
    "priorityLevel": 2,
    "isUrgent": false,
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```
**Flujo:** Mensajería - Protegido

### 19. **Listar Mensajes**
```http
GET /api/v1/messages
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Parámetros de consulta (Query Params):**
- `unitId` (int): ID de unidad organizacional
- `messageType` (int): Tipo de mensaje
- `status` (int): Estado del mensaje
- `isUrgent` (bool): Solo mensajes urgentes
- `dateFrom` (string): Fecha desde (YYYY-MM-DD)
- `dateTo` (string): Fecha hasta (YYYY-MM-DD)
- `searchText` (string): Texto a buscar
- `page` (int): Página (default: 1)
- `limit` (int): Elementos por página (default: 20, max: 100)
- `sortBy` (string): Ordenar por (created_at, subject, priority_level)
- `sortOrder` (string): Orden (asc, desc)

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Mensajes obtenidos exitosamente",
  "data": {
    "messages": [...],
    "total": 150,
    "page": 1,
    "limit": 20,
    "totalPages": 8,
    "hasNext": true,
    "hasPrevious": false
  }
}
```
**Flujo:** Mensajería - Protegido

### 20. **Obtener Mensaje por ID**
```http
GET /api/v1/messages/:id
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Mensaje obtenido exitosamente",
  "data": {
    "id": 123,
    "subject": "Asunto",
    "content": "Contenido completo",
    "sender": {...},
    "senderUnit": {...},
    "receiverUnit": {...},
    "messageType": {...},
    "status": {...},
    "attachments": [],
    "readAt": null,
    "respondedAt": null,
    "createdAt": "2025-01-01T00:00:00Z"
  }
}
```
**Flujo:** Mensajería - Protegido

### 21. **Marcar Mensaje como Leído**
```http
PUT /api/v1/messages/:id/read
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Mensaje marcado como leído exitosamente"
}
```
**Flujo:** Mensajería - Protegido

### 22. **Actualizar Estado de Mensaje**
```http
PUT /api/v1/messages/:id/status
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Datos esperados (Request Body):**
```json
{
  "status": "IN_PROGRESS|RESPONDED|RESOLVED|ARCHIVED|CANCELLED"
}
```
**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Estado actualizado exitosamente"
}
```
**Flujo:** Mensajería - Protegido

### 23. **Eliminar Mensaje (Soft Delete)**
```http
DELETE /api/v1/messages/:id
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Nota:** Solo el creador o admin puede eliminar

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Mensaje eliminado exitosamente"
}
```
**Flujo:** Mensajería - Protegido

### 24. **Obtener Estadísticas de Mensajes**
```http
GET /api/v1/messages/stats
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Parámetros de consulta (Query Params):**
- `unitId` (int): ID de unidad (solo admin)

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Estadísticas obtenidas exitosamente",
  "data": {
    "totalMessages": 1250,
    "messagesByStatus": {
      "SENT": 100,
      "READ": 450,
      "IN_PROGRESS": 300,
      "RESPONDED": 250,
      "RESOLVED": 150
    },
    "messagesByType": {...},
    "urgentMessages": 45,
    "averageResponseTime": "2h 30m"
  }
}
```
**Flujo:** Mensajería - Protegido

### 25. **Obtener Tipos de Mensajes**
```http
GET /api/v1/messages/types
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Funcionalidad en desarrollo",
  "data": [
    "SOLICITUD",
    "URGENTE",
    "COORDINACION",
    "NOTIFICACION",
    "SEGUIMIENTO",
    "CONSULTA",
    "DIRECTRIZ"
  ]
}
```
**Flujo:** Mensajería - Protegido

### 26. **Obtener Estados de Mensajes**
```http
GET /api/v1/messages/statuses
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Funcionalidad en desarrollo",
  "data": [
    "SENT",
    "READ",
    "IN_PROGRESS",
    "RESPONDED",
    "RESOLVED",
    "ARCHIVED",
    "CANCELLED"
  ]
}
```
**Flujo:** Mensajería - Protegido

---

## 🛡️ FLUJO DE ADMINISTRACIÓN

### 27. **Limpiar Tokens Expirados**
```http
POST /api/v1/auth/admin/cleanup-tokens
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Rol requerido:** `admin`

**Respuesta exitosa (200):**
```json
{
  "success": true,
  "message": "Limpieza completada",
  "data": {
    "cleanedTokens": 15,
    "timestamp": "2025-01-01T00:00:00Z"
  }
}
```
**Flujo:** Administración - Admin Only

### 28. **Gestión de Usuarios (Futuro)**
```http
GET /api/v1/admin/users
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Rol requerido:** `admin`

**Respuesta actual (200):**
```json
{
  "message": "Admin users endpoint - Tarea 7.1 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Administración - Admin Only

### 29. **Estadísticas del Sistema (Futuro)**
```http
GET /api/v1/admin/stats
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Rol requerido:** `admin`

**Respuesta actual (200):**
```json
{
  "message": "Admin stats endpoint - Tarea 6.x pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Administración - Admin Only

### 30. **Logs de Auditoría (Futuro)**
```http
GET /api/v1/admin/audit
```
**Headers requeridos:** `Authorization: Bearer {token}`
**Rol requerido:** `admin`

**Respuesta actual (200):**
```json
{
  "message": "Admin audit endpoint - Tarea 7.3 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Administración - Admin Only

---

## 📁 FLUJO DE ARCHIVOS (Futuro)

### 31. **Subir Archivo**
```http
POST /api/v1/files/upload
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta actual (200):**
```json
{
  "message": "File upload endpoint - Tarea 4.3 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Archivos - Protegido

### 32. **Descargar Archivo**
```http
GET /api/v1/files/:id
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta actual (200):**
```json
{
  "message": "File download endpoint - Tarea 4.3 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Archivos - Protegido

---

## 🔔 FLUJO DE NOTIFICACIONES (Futuro)

### 33. **Listar Notificaciones**
```http
GET /api/v1/notifications
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta actual (200):**
```json
{
  "message": "Notifications endpoint - Tarea 4.4 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Notificaciones - Protegido

### 34. **WebSocket Notificaciones**
```http
GET /api/v1/notifications/ws
```
**Headers requeridos:** `Authorization: Bearer {token}`

**Respuesta actual (200):**
```json
{
  "message": "WebSocket notifications - Tarea 4.4 pendiente",
  "status": "coming_soon"
}
```
**Flujo:** Notificaciones - Protegido

---

## 🏥 FLUJO DE SISTEMA

### 35. **Información del Servicio**
```http
GET /
```
**Respuesta exitosa (200):**
```json
{
  "service": "GAMC Backend Auth Service",
  "version": "2.0.0",
  "environment": "development",
  "timestamp": "2025-01-01T00:00:00Z",
  "status": "operational",
  "endpoints": {
    "auth": "/api/v1/auth",
    "health": "/health",
    "docs": "coming soon"
  }
}
```
**Flujo:** Sistema - Público

### 36. **Health Check**
```http
GET /health
```
**Respuesta exitosa (200):**
```json
{
  "status": "healthy",
  "timestamp": "2025-01-01T00:00:00Z",
  "uptime": 86400,
  "services": {
    "database": "connected",
    "redis": "connected",
    "auth": "operational"
  },
  "version": "2.0.0",
  "environment": "development"
}
```
**Flujo:** Sistema - Público

---

## 🚫 RESPUESTAS DE ERROR COMUNES

### Error 400 - Bad Request
```json
{
  "success": false,
  "message": "Datos inválidos",
  "error": "Detalles específicos del error",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error 401 - Unauthorized
```json
{
  "success": false,
  "message": "No autorizado",
  "error": "Token inválido o expirado",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error 403 - Forbidden
```json
{
  "success": false,
  "message": "Acceso denegado",
  "error": "No tiene permisos para realizar esta acción",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error 404 - Not Found
```json
{
  "success": false,
  "message": "Recurso no encontrado",
  "error": "El recurso solicitado no existe",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error 429 - Too Many Requests
```json
{
  "success": false,
  "message": "Demasiadas peticiones",
  "error": "Intente de nuevo más tarde",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

### Error 500 - Internal Server Error
```json
{
  "success": false,
  "message": "Error interno del servidor",
  "error": "Contacte al administrador",
  "timestamp": "2025-01-01T00:00:00Z"
}
```

---

## 📝 NOTAS IMPORTANTES

1. **Autenticación**: Todos los endpoints protegidos requieren un token JWT válido en el header Authorization.

2. **Rate Limiting**: 
   - Global: 100 requests/15 minutos
   - Auth endpoints: 10 requests/15 minutos

3. **CORS**: Configurado para aceptar requests desde `http://localhost:5173` en desarrollo.

4. **Cookies**: El refresh token se envía como cookie HttpOnly para mayor seguridad.

5. **Roles**:
   - `admin`: Acceso total
   - `input`: Puede crear y enviar mensajes
   - `output`: Solo puede leer mensajes

6. **Validaciones**:
   - Email: Debe ser @gamc.gov.bo
   - Password: Mínimo 8 caracteres
   - Tokens: 64 caracteres hexadecimales

7. **Timezone**: Todas las fechas están en UTC, considerar zona horaria de Bolivia (UTC-4).

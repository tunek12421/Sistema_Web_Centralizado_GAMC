# 🛠️ Guía de Desarrollo - GAMC Sistema Web Centralizado

## 🚀 Configuración del Entorno de Desarrollo

### Prerrequisitos
- Node.js >= 18.0.0
- Docker >= 20.10
- Docker Compose >= 2.0
- Git

### Configuración Inicial

1. **Clonar el repositorio:**
```bash
git clone https://github.com/tu-usuario/sistema-web-centralizado-gamc.git
cd sistema-web-centralizado-gamc
```

2. **Configurar variables de entorno:**
```bash
# Copiar archivos de ejemplo
cp backend-auth/.env.example backend-auth/.env
cp frontend-auth/.env.example frontend-auth/.env

# Editar según necesidades
```

3. **Iniciar servicios de desarrollo:**
```bash
./gamc.sh start
```

## 🏗️ Arquitectura de Desarrollo

### Backend (Node.js + TypeScript)

#### Estructura de directorios
```
backend-auth/
├── src/
│   ├── controllers/     # Controladores de rutas
│   ├── middleware/      # Middleware personalizado
│   ├── routes/         # Definición de rutas
│   ├── services/       # Lógica de negocio
│   ├── config/         # Configuraciones
│   ├── types/          # Tipos de TypeScript
│   └── utils/          # Utilidades
├── prisma/
│   └── schema.prisma   # Esquema de base de datos
├── Dockerfile
└── package.json
```

#### Scripts de desarrollo
```bash
cd backend-auth

# Desarrollo con hot reload
npm run dev

# Generar cliente Prisma
npm run prisma:generate

# Aplicar migraciones
npm run prisma:push

# Compilar TypeScript
npm run build
```

#### Configuración de Base de Datos

Prisma se conecta automáticamente usando la variable `DATABASE_URL`. Para desarrollo local:

```bash
# Acceder a la base de datos
docker-compose exec postgres psql -U gamc_user gamc_system

# Aplicar cambios de esquema
cd backend-auth
npx prisma db push

# Ver datos en Prisma Studio
npx prisma studio
```

### Frontend (React + Vite + TypeScript)

#### Estructura de directorios
```
frontend-auth/
├── src/
│   ├── components/     # Componentes reutilizables
│   ├── pages/         # Páginas de la aplicación
│   ├── services/      # Servicios API
│   ├── store/         # Estado global (Redux)
│   ├── hooks/         # Hooks personalizados
│   ├── types/         # Tipos de TypeScript
│   └── styles/        # Estilos globales
├── public/            # Archivos estáticos
├── Dockerfile
└── package.json
```

#### Scripts de desarrollo
```bash
cd frontend-auth

# Desarrollo con hot reload
npm run dev

# Compilar para producción
npm run build

# Linter
npm run lint

# Preview de producción
npm run preview
```

#### Configuración de Vite

```typescript
// vite.config.ts
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    watch: {
      usePolling: true, // Para Docker
    },
    hmr: {
      port: 5173,
    }
  }
})
```

## 🔧 Herramientas de Desarrollo

### Hot Reload
- **Backend**: `tsx watch` recarga automáticamente en cambios
- **Frontend**: Vite HMR (Hot Module Replacement) para React

### Debugging

#### Backend (Node.js)
```bash
# Con debugger
NODE_OPTIONS="--inspect=0.0.0.0:9229" npm run dev

# En VS Code, agregar configuración launch.json:
{
  "type": "node",
  "request": "attach",
  "name": "Docker: Attach to Node",
  "port": 9229,
  "address": "localhost",
  "localRoot": "${workspaceFolder}/backend-auth",
  "remoteRoot": "/app"
}
```

#### Frontend (React)
```bash
# React DevTools disponible en navegador
# Redux DevTools para estado global
```

### Base de Datos

#### Prisma Studio
```bash
cd backend-auth
npx prisma studio
# Abrir http://localhost:5555
```

#### Acceso directo a PostgreSQL
```bash
# Via Docker
docker-compose exec postgres psql -U gamc_user gamc_system

# Via PgAdmin
# http://localhost:8080
```

#### Migrations
```bash
cd backend-auth

# Crear nueva migración
npx prisma migrate dev --name add_new_feature

# Resetear base de datos
npx prisma migrate reset

# Aplicar migraciones en producción
npx prisma migrate deploy
```

## 🧪 Testing

### Backend Testing
```bash
cd backend-auth

# Instalar dependencias de testing
npm install --save-dev jest @types/jest supertest

# Ejecutar tests
npm run test

# Coverage
npm run test:coverage
```

### Frontend Testing
```bash
cd frontend-auth

# Instalar dependencias de testing
npm install --save-dev vitest @testing-library/react

# Ejecutar tests
npm run test

# Tests en modo watch
npm run test:watch
```

### Testing de Integración
```bash
# Ejecutar suite completa con Docker
docker-compose -f docker-compose.yml -f docker-compose.test.yml up --abort-on-container-exit
```

## 📊 Monitoring y Logs

### Logs de Desarrollo
```bash
# Ver logs en tiempo real
./gamc.sh logs

# Logs específicos
./gamc.sh logs gamc-auth-backend
./gamc.sh logs gamc-auth-frontend

# Logs con timestamps
docker-compose logs --timestamps -f
```

### Health Checks
```bash
# Backend health
curl http://localhost:3000/health

# Verificar todos los servicios
./gamc.sh status
```

### Métricas de Redis
```bash
# Conectar a Redis
docker-compose exec redis redis-cli

# Ver estadísticas
INFO memory
INFO stats
```

## 🔒 Seguridad en Desarrollo

### Variables Sensibles
```bash
# Nunca commits archivos .env
# Usar .env.example como plantilla

# Para JWT secrets en desarrollo
openssl rand -hex 64
```

### CORS
```typescript
// backend: configuración CORS
app.use(cors({
  origin: process.env.CORS_ORIGIN || 'http://localhost:5173',
  credentials: true
}));
```

### Rate Limiting
```typescript
// backend: rate limiting para desarrollo
import rateLimit from 'express-rate-limit';

const limiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutos
  max: 100 // límite de requests
});
```

## 🚢 Deployment

### Build de Producción
```bash
# Construir imágenes
docker-compose build

# Crear imágenes optimizadas
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build
```

### Variables de Producción
```bash
# Crear .env.production
NODE_ENV=production
JWT_SECRET=your_super_secure_jwt_secret
DATABASE_URL=postgresql://user:pass@host:5432/db
REDIS_URL=redis://user:pass@host:6379/0
```

### Deployment Scripts
```bash
# Script de deploy
#!/bin/bash
set -e

echo "🚀 Deploying GAMC Sistema..."

# Pull latest code
git pull origin main

# Build images
docker-compose build --no-cache

# Run migrations
docker-compose exec backend-auth npx prisma migrate deploy

# Restart services
docker-compose down
docker-compose up -d

echo "✅ Deployment completed!"
```

## 🐛 Debugging y Troubleshooting

### Problemas Comunes

#### Puerto ocupado
```bash
# Linux/Mac
lsof -i :5173
kill -9 <PID>

# Windows
netstat -ano | findstr :5173
taskkill /PID <PID> /F
```

#### Docker build fails
```bash
# Limpiar cache
docker system prune -f
docker builder prune -f

# Rebuild sin cache
docker-compose build --no-cache
```

#### Base de datos no conecta
```bash
# Verificar logs
./gamc.sh logs postgres

# Verificar conectividad
docker-compose exec backend-auth npm run test-db-connection
```

#### Frontend no carga
```bash
# Verificar configuración Vite
cat frontend-auth/vite.config.ts

# Verificar variables de entorno
docker-compose exec frontend-auth env | grep VITE_
```

### Performance Debugging

#### Backend Performance
```bash
# Profiling con clinic.js
npm install -g clinic
clinic doctor -- node dist/server.js
```

#### Frontend Performance
```bash
# Bundle analyzer
npm install --save-dev vite-bundle-analyzer
npm run build:analyze
```

## 📚 Recursos Adicionales

### Documentación de APIs
- Swagger/OpenAPI disponible en: `http://localhost:3000/api/docs`
- Postman collection en: `docs/postman/`

### Diagramas de Arquitectura
- Diagramas en: `docs/architecture/`
- Flujos de autenticación: `docs/auth-flows.md`

### Code Style
- ESLint configurado para TypeScript
- Prettier para formateo automático
- Husky para pre-commit hooks

### Git Workflow
```bash
# Feature branch
git checkout -b feature/nueva-funcionalidad

# Commits descriptivos
git commit -m "feat: agregar autenticación JWT"
git commit -m "fix: corregir validación de email"
git commit -m "docs: actualizar README"

# Push y PR
git push origin feature/nueva-funcionalidad
```

---

**¿Necesitas ayuda? Contacta al equipo de desarrollo o crea un issue en GitHub.**

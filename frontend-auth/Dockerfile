# ========================================
# GAMC Frontend Auth - Dockerfile Multi-stage
# ========================================

# Etapa de build para producción
FROM node:18-alpine AS builder

WORKDIR /app

# Instalar dependencias
COPY package*.json ./
RUN npm install

# Copiar código fuente
COPY . .

# Build para producción
RUN npm run build

# Etapa de producción con nginx
FROM nginx:alpine AS production

# Copiar archivos build
COPY --from=builder /app/dist /usr/share/nginx/html

# Configuración nginx personalizada
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]

# Etapa de desarrollo - VERSIÓN CORREGIDA
FROM node:18-alpine AS development

# Crear usuario node con UID/GID específicos para evitar problemas de permisos
RUN addgroup -g 1001 -S nodejs && \
    adduser -S reactjs -u 1001 -G nodejs

WORKDIR /app

# Cambiar propietario del directorio de trabajo
RUN chown -R reactjs:nodejs /app

# Cambiar al usuario no-root
USER reactjs

# Copiar package files
COPY --chown=reactjs:nodejs package*.json ./

# Instalar dependencias como usuario reactjs
RUN npm install

# Copiar código fuente
COPY --chown=reactjs:nodejs . .

# Asegurar permisos de ejecución en node_modules
RUN chmod -R 755 node_modules/.bin/ 2>/dev/null || true

EXPOSE 5173

# IMPORTANTE: Agregar --host 0.0.0.0 para Docker
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
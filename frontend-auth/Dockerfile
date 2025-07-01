# ========================================
# GAMC Frontend Auth - Dockerfile Multi-stage
# ========================================

# Etapa de build para producción
FROM node:18-alpine AS builder

WORKDIR /app

# Instalar dependencias
COPY package*.json ./
RUN npm ci

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

# Etapa de desarrollo
FROM node:18-alpine AS development

WORKDIR /app

# Copiar package files
COPY package*.json ./
RUN npm ci

# Copiar código fuente
COPY . .

EXPOSE 5173

# IMPORTANTE: Agregar --host 0.0.0.0 para Docker
CMD ["npm", "run", "dev"]
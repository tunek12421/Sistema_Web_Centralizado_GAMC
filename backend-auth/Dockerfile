FROM node:18-alpine

WORKDIR /app

# Instalar dependencias necesarias para Prisma
RUN apk add --no-cache curl openssl libc6-compat

# Crear usuario no-root
RUN addgroup -g 1001 -S nodejs && \
    adduser -S nodejs -u 1001

# Copiar archivos de dependencias
COPY package*.json ./

# Instalar dependencias
RUN npm ci && npm cache clean --force

# Copiar código fuente
COPY . .

# Cambiar propiedad de archivos
RUN chown -R nodejs:nodejs /app

# Cambiar a usuario no-root
USER nodejs

EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

# Script de inicio que genera Prisma y ejecuta la app
CMD ["sh", "-c", "npx prisma generate && npx tsx src/server.ts"]
// ========================================
// GAMC Backend Auth - Configuración Base de Datos
// ========================================

import { PrismaClient } from '@prisma/client';

// Crear instancia de Prisma
const prisma = new PrismaClient({
  log: process.env.NODE_ENV === 'development' ? ['query', 'info', 'warn', 'error'] : ['error'],
  datasources: {
    db: {
      url: process.env.DATABASE_URL
    }
  }
});

// Configuración de conexión
export const databaseConfig = {
  url: process.env.DATABASE_URL || 'postgresql://gamc_user:gamc_password_2024@localhost:5432/gamc_system',
  connectTimeout: 30000,
  maxConnections: 20,
  retryAttempts: 3,
  retryDelay: 1000
};

// Hook para desconectar en shutdown
const gracefulShutdown = async () => {
  console.log('🔌 Cerrando conexión a PostgreSQL...');
  await prisma.$disconnect();
  process.exit(0);
};

process.on('SIGINT', gracefulShutdown);
process.on('SIGTERM', gracefulShutdown);

// Función para probar conexión
export const testDatabaseConnection = async (): Promise<boolean> => {
  try {
    await prisma.$connect();
    console.log('✅ Conexión a PostgreSQL establecida');
    return true;
  } catch (error) {
    console.error('❌ Error conectando a PostgreSQL:', error);
    return false;
  }
};

// Función para obtener estadísticas de la DB
export const getDatabaseStats = async () => {
  try {
    const users = await prisma.users.count();
    const messages = await prisma.messages.count();
    const organizationalUnits = await prisma.organizational_units.count();
    
    return {
      users,
      messages,
      organizationalUnits,
      connected: true
    };
  } catch (error) {
    console.error('Error obteniendo estadísticas DB:', error);
    return {
      users: 0,
      messages: 0,
      organizationalUnits: 0,
      connected: false
    };
  }
};

export default prisma;
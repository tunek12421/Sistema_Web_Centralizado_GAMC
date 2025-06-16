// ========================================
// GAMC Backend Auth - Configuración Redis
// ========================================

import { createClient, RedisClientType } from 'redis';

// Configuración de Redis
const redisConfig = {
  url: process.env.REDIS_URL || 'redis://:gamc_redis_password_2024@localhost:6379/0',
  socket: {
    connectTimeout: 10000,
    lazyConnect: true,
    reconnectStrategy: (retries: number) => {
      if (retries > 10) {
        console.error('❌ Redis: Máximo número de reconexiones alcanzado');
        return false;
      }
      return Math.min(retries * 100, 3000);
    }
  },
  database: 0 // DB 0 para sesiones
};

// Cliente principal para sesiones
export const redisClient: RedisClientType = createClient(redisConfig);

// Cliente para blacklist de JWT (DB 5)
export const redisBlacklistClient: RedisClientType = createClient({
  ...redisConfig,
  url: process.env.REDIS_URL?.replace('/0', '/5') || 'redis://:gamc_redis_password_2024@localhost:6379/5'
});

// Configuración de conexión
const setupRedisConnection = (client: RedisClientType, name: string) => {
  client.on('connect', () => {
    console.log(`✅ Redis ${name}: Conexión establecida`);
  });

  client.on('ready', () => {
    console.log(`🚀 Redis ${name}: Cliente listo`);
  });

  client.on('error', (error) => {
    console.error(`❌ Redis ${name} Error:`, error);
  });

  client.on('reconnecting', () => {
    console.log(`🔄 Redis ${name}: Reconectando...`);
  });

  client.on('end', () => {
    console.log(`📴 Redis ${name}: Conexión cerrada`);
  });
};

// Configurar eventos para ambos clientes
setupRedisConnection(redisClient, 'Sessions');
setupRedisConnection(redisBlacklistClient, 'Blacklist');

// Conectar clientes
export const connectRedis = async (): Promise<boolean> => {
  try {
    await redisClient.connect();
    await redisBlacklistClient.connect();
    
    // Probar conexiones
    await redisClient.ping();
    await redisBlacklistClient.ping();
    
    console.log('✅ Redis: Ambos clientes conectados exitosamente');
    return true;
  } catch (error) {
    console.error('❌ Redis: Error en conexión:', error);
    return false;
  }
};

// Desconectar clientes
export const disconnectRedis = async (): Promise<void> => {
  try {
    await redisClient.quit();
    await redisBlacklistClient.quit();
    console.log('📴 Redis: Desconectado exitosamente');
  } catch (error) {
    console.error('❌ Redis: Error al desconectar:', error);
  }
};

// Utilidades para sesiones
export const sessionUtils = {
  // Guardar sesión
  async saveSession(sessionId: string, sessionData: any, ttlSeconds = 7 * 24 * 60 * 60): Promise<boolean> {
    try {
      await redisClient.setEx(
        `session:${sessionId}`,
        ttlSeconds,
        JSON.stringify(sessionData)
      );
      return true;
    } catch (error) {
      console.error('Error guardando sesión:', error);
      return false;
    }
  },

  // Obtener sesión
  async getSession(sessionId: string): Promise<any | null> {
    try {
      const sessionData = await redisClient.get(`session:${sessionId}`);
      return sessionData ? JSON.parse(sessionData) : null;
    } catch (error) {
      console.error('Error obteniendo sesión:', error);
      return null;
    }
  },

  // Eliminar sesión
  async deleteSession(sessionId: string): Promise<boolean> {
    try {
      await redisClient.del(`session:${sessionId}`);
      return true;
    } catch (error) {
      console.error('Error eliminando sesión:', error);
      return false;
    }
  },

  // Extender sesión
  async extendSession(sessionId: string, ttlSeconds = 7 * 24 * 60 * 60): Promise<boolean> {
    try {
      await redisClient.expire(`session:${sessionId}`, ttlSeconds);
      return true;
    } catch (error) {
      console.error('Error extendiendo sesión:', error);
      return false;
    }
  },

  // Obtener todas las sesiones de un usuario
  async getUserSessions(userId: string): Promise<string[]> {
    try {
      const keys = await redisClient.keys('session:*');
      const sessions: string[] = [];

      for (const key of keys) {
        const sessionData = await redisClient.get(key);
        if (sessionData) {
          const data = JSON.parse(sessionData);
          if (data.userId === userId) {
            sessions.push(key.replace('session:', ''));
          }
        }
      }

      return sessions;
    } catch (error) {
      console.error('Error obteniendo sesiones de usuario:', error);
      return [];
    }
  }
};

// Utilidades para refresh tokens
export const refreshTokenUtils = {
  // Guardar refresh token
  async saveRefreshToken(userId: string, sessionId: string, token: string, ttlSeconds = 7 * 24 * 60 * 60): Promise<boolean> {
    try {
      await redisClient.setEx(
        `refresh:${userId}:${sessionId}`,
        ttlSeconds,
        token
      );
      return true;
    } catch (error) {
      console.error('Error guardando refresh token:', error);
      return false;
    }
  },

  // Obtener refresh token
  async getRefreshToken(userId: string, sessionId: string): Promise<string | null> {
    try {
      return await redisClient.get(`refresh:${userId}:${sessionId}`);
    } catch (error) {
      console.error('Error obteniendo refresh token:', error);
      return null;
    }
  },

  // Eliminar refresh token
  async deleteRefreshToken(userId: string, sessionId: string): Promise<boolean> {
    try {
      await redisClient.del(`refresh:${userId}:${sessionId}`);
      return true;
    } catch (error) {
      console.error('Error eliminando refresh token:', error);
      return false;
    }
  },

  // Eliminar todos los refresh tokens de un usuario
  async deleteAllUserRefreshTokens(userId: string): Promise<boolean> {
    try {
      const keys = await redisClient.keys(`refresh:${userId}:*`);
      if (keys.length > 0) {
        await redisClient.del(keys);
      }
      return true;
    } catch (error) {
      console.error('Error eliminando todos los refresh tokens:', error);
      return false;
    }
  }
};

// Utilidades para blacklist de JWT
export const jwtBlacklistUtils = {
  // Agregar token a blacklist
  async blacklistToken(jti: string, exp: number): Promise<boolean> {
    try {
      const ttl = exp - Math.floor(Date.now() / 1000);
      if (ttl > 0) {
        await redisBlacklistClient.setEx(`blacklist:${jti}`, ttl, 'revoked');
      }
      return true;
    } catch (error) {
      console.error('Error agregando token a blacklist:', error);
      return false;
    }
  },

  // Verificar si token está en blacklist
  async isTokenBlacklisted(jti: string): Promise<boolean> {
    try {
      const result = await redisBlacklistClient.exists(`blacklist:${jti}`);
      return result === 1;
    } catch (error) {
      console.error('Error verificando blacklist:', error);
      return false;
    }
  },

  // Limpiar tokens expirados (ejecutar periódicamente)
  async cleanExpiredTokens(): Promise<number> {
    try {
      const keys = await redisBlacklistClient.keys('blacklist:*');
      let cleaned = 0;

      for (const key of keys) {
        const ttl = await redisBlacklistClient.ttl(key);
        if (ttl <= 0) {
          await redisBlacklistClient.del(key);
          cleaned++;
        }
      }

      return cleaned;
    } catch (error) {
      console.error('Error limpiando tokens expirados:', error);
      return 0;
    }
  }
};

// Estadísticas de Redis
export const getRedisStats = async () => {
  try {
    const sessionKeys = await redisClient.keys('session:*');
    const refreshKeys = await redisClient.keys('refresh:*');
    const blacklistKeys = await redisBlacklistClient.keys('blacklist:*');

    const info = await redisClient.info('memory');
    const memoryUsed = info.match(/used_memory_human:(.+)/)?.[1]?.trim() || 'N/A';

    return {
      sessions: sessionKeys.length,
      refreshTokens: refreshKeys.length,
      blacklistedTokens: blacklistKeys.length,
      memoryUsed,
      connected: true
    };
  } catch (error) {
    console.error('Error obteniendo estadísticas Redis:', error);
    return {
      sessions: 0,
      refreshTokens: 0,
      blacklistedTokens: 0,
      memoryUsed: 'N/A',
      connected: false
    };
  }
};

// Limpieza automática cada hora
setInterval(async () => {
  const cleaned = await jwtBlacklistUtils.cleanExpiredTokens();
  if (cleaned > 0) {
    console.log(`🧹 Redis: Limpiados ${cleaned} tokens expirados de blacklist`);
  }
}, 60 * 60 * 1000);

// Hook para desconectar en shutdown
const gracefulShutdown = async () => {
  console.log('🔌 Cerrando conexiones Redis...');
  await disconnectRedis();
  process.exit(0);
};

process.on('SIGINT', gracefulShutdown);
process.on('SIGTERM', gracefulShutdown);
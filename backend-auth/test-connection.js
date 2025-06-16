const { PrismaClient } = require('@prisma/client');
const redis = require('redis');

async function testConnections() {
  console.log('🧪 Probando conexiones...');
  
  // Test PostgreSQL
  try {
    const prisma = new PrismaClient();
    await prisma.$connect();
    const users = await prisma.users.count();
    console.log(`✅ PostgreSQL: Conectado (${users} usuarios)`);
    await prisma.$disconnect();
  } catch (error) {
    console.error('❌ PostgreSQL:', error.message);
  }
  
  // Test Redis
  try {
    const client = redis.createClient({
      url: 'redis://:gamc_redis_password_2024@localhost:6379/0'
    });
    await client.connect();
    await client.ping();
    console.log('✅ Redis: Conectado');
    await client.quit();
  } catch (error) {
    console.error('❌ Redis:', error.message);
  }
}

testConnections();

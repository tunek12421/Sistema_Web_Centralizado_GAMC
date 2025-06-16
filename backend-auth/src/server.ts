// ========================================
// GAMC Backend Auth - Servidor Principal
// ========================================

import 'dotenv/config';
import App from './app';

const PORT = process.env.PORT || 3000;
const NODE_ENV = process.env.NODE_ENV || 'development';

async function startServer() {
  try {
    // Crear aplicación
    const app = new App();
    
    // Inicializar servicios
    await app.initialize();
    
    // Iniciar servidor
    const server = app.app.listen(PORT, () => {
      console.log('🎉 ========================================');
      console.log('🎉 GAMC Sistema de Autenticación');
      console.log('🎉 ========================================');
      console.log(`🚀 Servidor ejecutándose en puerto ${PORT}`);
      console.log(`🌍 Entorno: ${NODE_ENV}`);
      console.log(`📡 Health check: http://localhost:${PORT}/health`);
      console.log(`🔐 Auth API: http://localhost:${PORT}/api/v1/auth`);
      console.log('🎉 ========================================');
    });

    // Manejo graceful de shutdown
    const gracefulShutdown = (signal: string) => {
      console.log(`\n📴 Recibida señal ${signal}. Cerrando servidor...`);
      
      server.close(() => {
        console.log('✅ Servidor HTTP cerrado.');
        process.exit(0);
      });

      // Forzar cierre después de 10 segundos
      setTimeout(() => {
        console.error('❌ Forzando cierre del servidor...');
        process.exit(1);
      }, 10000);
    };

    // Escuchar señales de sistema
    process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
    process.on('SIGINT', () => gracefulShutdown('SIGINT'));

  } catch (error) {
    console.error('❌ Error fatal al iniciar servidor:', error);
    process.exit(1);
  }
}

// Iniciar servidor
startServer();
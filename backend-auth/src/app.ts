// ========================================
// GAMC Backend Auth - Aplicación Express
// ========================================

import express, { Application, Request, Response } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import morgan from 'morgan';
import rateLimit from 'express-rate-limit';
import cookieParser from 'cookie-parser';
import { connectRedis, getRedisStats } from './config/redis';
import { testDatabaseConnection, getDatabaseStats } from './config/database';
import authRoutes from './routes/authRoutes';

class App {
  public app: Application;

  constructor() {
    this.app = express();
    this.initializeMiddlewares();
    this.initializeRoutes();
    this.initializeErrorHandling();
  }

  private initializeMiddlewares(): void {
    // Seguridad
    this.app.use(helmet({
      crossOriginResourcePolicy: { policy: 'cross-origin' }
    }));

    // CORS
    this.app.use(cors({
      origin: process.env.CORS_ORIGIN || 'http://localhost:5173',
      credentials: true,
      methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
      allowedHeaders: ['Content-Type', 'Authorization'],
      exposedHeaders: ['X-Total-Count']
    }));

    // Rate limiting global
    const limiter = rateLimit({
      windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '900000'), // 15 minutos
      max: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100'), // límite de requests
      message: {
        success: false,
        message: 'Demasiadas peticiones desde esta IP, intente de nuevo más tarde.',
        timestamp: new Date().toISOString()
      },
      standardHeaders: true,
      legacyHeaders: false
    });
    this.app.use(limiter);

    // Logging
    if (process.env.NODE_ENV === 'development') {
      this.app.use(morgan('dev'));
    } else {
      this.app.use(morgan('combined'));
    }

    // Parsers
    this.app.use(express.json({ limit: '10mb' }));
    this.app.use(express.urlencoded({ extended: true, limit: '10mb' }));
    this.app.use(cookieParser());

    // Trust proxy (para obtener IP real detrás de proxy)
    this.app.set('trust proxy', 1);
  }

  private initializeRoutes(): void {
    const apiPrefix = process.env.API_PREFIX || '/api/v1';

    // Health check
    this.app.get('/health', async (req: Request, res: Response) => {
      try {
        const dbStats = await getDatabaseStats();
        const redisStats = await getRedisStats();

        res.status(200).json({
          success: true,
          message: 'GAMC Auth Service - Health Check',
          data: {
            status: 'healthy',
            timestamp: new Date().toISOString(),
            environment: process.env.NODE_ENV || 'development',
            version: '1.0.0',
            services: {
              database: {
                connected: dbStats.connected,
                users: dbStats.users,
                messages: dbStats.messages,
                organizationalUnits: dbStats.organizationalUnits
              },
              redis: {
                connected: redisStats.connected,
                sessions: redisStats.sessions,
                refreshTokens: redisStats.refreshTokens,
                blacklistedTokens: redisStats.blacklistedTokens,
                memoryUsed: redisStats.memoryUsed
              }
            }
          }
        });
      } catch (error) {
        res.status(503).json({
          success: false,
          message: 'Service Unavailable',
          timestamp: new Date().toISOString()
        });
      }
    });

    // Info básica
    this.app.get('/', (req: Request, res: Response) => {
      res.status(200).json({
        success: true,
        message: 'GAMC Sistema de Autenticación API',
        data: {
          version: '1.0.0',
          environment: process.env.NODE_ENV || 'development',
          timestamp: new Date().toISOString(),
          endpoints: {
            health: '/health',
            auth: `${apiPrefix}/auth`,
            docs: `${apiPrefix}/docs`
          }
        }
      });
    });

    // Rutas de autenticación
    this.app.use(`${apiPrefix}/auth`, authRoutes);

    // Ruta 404
    this.app.use('*', (req: Request, res: Response) => {
      res.status(404).json({
        success: false,
        message: 'Endpoint no encontrado',
        path: req.originalUrl,
        timestamp: new Date().toISOString()
      });
    });
  }

  private initializeErrorHandling(): void {
    // Manejo global de errores
    this.app.use((error: any, req: Request, res: Response, next: any) => {
      console.error('❌ Error no manejado:', error);

      // Error de validación de JSON
      if (error instanceof SyntaxError && 'body' in error) {
        return res.status(400).json({
          success: false,
          message: 'JSON malformado',
          timestamp: new Date().toISOString()
        });
      }

      // Error genérico
      res.status(error.status || 500).json({
        success: false,
        message: error.message || 'Error interno del servidor',
        timestamp: new Date().toISOString(),
        ...(process.env.NODE_ENV === 'development' && { stack: error.stack })
      });
    });

    // Manejo de promesas rechazadas no capturadas
    process.on('unhandledRejection', (reason, promise) => {
      console.error('❌ Unhandled Rejection at:', promise, 'reason:', reason);
    });

    // Manejo de excepciones no capturadas
    process.on('uncaughtException', (error) => {
      console.error('❌ Uncaught Exception:', error);
      process.exit(1);
    });
  }

  public async initialize(): Promise<void> {
    try {
      console.log('🚀 Inicializando GAMC Auth Service...');

      // Conectar a Redis
      const redisConnected = await connectRedis();
      if (!redisConnected) {
        throw new Error('No se pudo conectar a Redis');
      }

      // Probar conexión a base de datos
      const dbConnected = await testDatabaseConnection();
      if (!dbConnected) {
        throw new Error('No se pudo conectar a PostgreSQL');
      }

      console.log('✅ Servicios conectados exitosamente');

    } catch (error) {
      console.error('❌ Error en inicialización:', error);
      throw error;
    }
  }
}

export default App;
// ========================================
// GAMC Backend Auth - Middleware de Autenticación
// ========================================

import { Request, Response, NextFunction } from 'express';
import { verifyAccessToken, extractTokenFromHeader } from '../utils/jwt';
import { sessionUtils, jwtBlacklistUtils } from '../config/redis';
import { AuthService } from '../services/authService';
import { UserProfile } from '../types/auth';

// Extender Request para incluir usuario y sesión
declare global {
  namespace Express {
    interface Request {
      user?: UserProfile;
      sessionId?: string;
    }
  }
}

// Middleware principal de autenticación
export const authenticateToken = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
  try {
    // Extraer token del header Authorization
    const token = extractTokenFromHeader(req.get('Authorization'));

    if (!token) {
      res.status(401).json({
        success: false,
        message: 'Token de acceso requerido',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Verificar que el token no esté en blacklist
    const decoded = verifyAccessToken(token);
    if (!decoded) {
      res.status(401).json({
        success: false,
        message: 'Token inválido o expirado',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Verificar blacklist usando JTI si está disponible
    if (decoded.jti) {
      const isBlacklisted = await jwtBlacklistUtils.isTokenBlacklisted(decoded.jti);
      if (isBlacklisted) {
        res.status(401).json({
          success: false,
          message: 'Token revocado',
          timestamp: new Date().toISOString()
        });
        return;
      }
    }

    // Verificar que la sesión existe en Redis
    const sessionData = await sessionUtils.getSession(decoded.sessionId);
    if (!sessionData) {
      res.status(401).json({
        success: false,
        message: 'Sesión expirada',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Verificar que los datos del token coincidan con la sesión
    if (sessionData.userId !== decoded.userId || sessionData.email !== decoded.email) {
      res.status(401).json({
        success: false,
        message: 'Datos de sesión inconsistentes',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Obtener perfil actualizado del usuario
    const userProfile = await AuthService.getUserProfile(decoded.userId);
    if (!userProfile || !userProfile.isActive) {
      res.status(401).json({
        success: false,
        message: 'Usuario no encontrado o inactivo',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Actualizar última actividad en la sesión
    sessionData.lastActivity = new Date();
    await sessionUtils.saveSession(decoded.sessionId, sessionData);

    // Agregar usuario y sessionId al request
    req.user = userProfile;
    req.sessionId = decoded.sessionId;

    next();

  } catch (error) {
    console.error('Error en middleware de autenticación:', error);
    res.status(500).json({
      success: false,
      message: 'Error interno del servidor',
      timestamp: new Date().toISOString()
    });
  }
};

// Middleware opcional de autenticación (no falla si no hay token)
export const optionalAuth = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
  try {
    const token = extractTokenFromHeader(req.get('Authorization'));

    if (token) {
      const decoded = verifyAccessToken(token);
      if (decoded) {
        const sessionData = await sessionUtils.getSession(decoded.sessionId);
        if (sessionData) {
          const userProfile = await AuthService.getUserProfile(decoded.userId);
          if (userProfile && userProfile.isActive) {
            req.user = userProfile;
            req.sessionId = decoded.sessionId;
          }
        }
      }
    }

    next();
  } catch (error) {
    // En modo opcional, no fallar si hay errores
    console.error('Error en middleware de autenticación opcional:', error);
    next();
  }
};

// Middleware para verificar roles específicos
export const requireRole = (...allowedRoles: string[]) => {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) {
      res.status(401).json({
        success: false,
        message: 'Autenticación requerida',
        timestamp: new Date().toISOString()
      });
      return;
    }

    if (!allowedRoles.includes(req.user.role)) {
      res.status(403).json({
        success: false,
        message: 'Permisos insuficientes',
        timestamp: new Date().toISOString()
      });
      return;
    }

    next();
  };
};

// Middleware para verificar unidad organizacional
export const requireOrganizationalUnit = (...allowedUnitIds: number[]) => {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) {
      res.status(401).json({
        success: false,
        message: 'Autenticación requerida',
        timestamp: new Date().toISOString()
      });
      return;
    }

    if (!allowedUnitIds.includes(req.user.organizationalUnit.id)) {
      res.status(403).json({
        success: false,
        message: 'Acceso restringido a su unidad organizacional',
        timestamp: new Date().toISOString()
      });
      return;
    }

    next();
  };
};

// Middleware para verificar que el usuario es admin
export const requireAdmin = requireRole('admin');

// Middleware para verificar que el usuario puede enviar mensajes
export const requireInputRole = requireRole('admin', 'input');

// Middleware para verificar que el usuario puede ver mensajes
export const requireOutputRole = requireRole('admin', 'input', 'output');

// Middleware para verificar propiedad del recurso
export const requireOwnership = (getUserIdFromResource: (req: Request) => string) => {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) {
      res.status(401).json({
        success: false,
        message: 'Autenticación requerida',
        timestamp: new Date().toISOString()
      });
      return;
    }

    // Los admins pueden acceder a cualquier recurso
    if (req.user.role === 'admin') {
      next();
      return;
    }

    const resourceUserId = getUserIdFromResource(req);
    if (req.user.id !== resourceUserId) {
      res.status(403).json({
        success: false,
        message: 'Solo puede acceder a sus propios recursos',
        timestamp: new Date().toISOString()
      });
      return;
    }

    next();
  };
};

// Middleware para rate limiting por usuario
export const userRateLimit = (maxRequests: number, windowMs: number) => {
  const userRequests = new Map<string, { count: number; resetTime: number }>();

  return (req: Request, res: Response, next: NextFunction): void => {
    if (!req.user) {
      next();
      return;
    }

    const userId = req.user.id;
    const now = Date.now();
    const userLimit = userRequests.get(userId);

    if (!userLimit || now > userLimit.resetTime) {
      userRequests.set(userId, {
        count: 1,
        resetTime: now + windowMs
      });
      next();
      return;
    }

    if (userLimit.count >= maxRequests) {
      res.status(429).json({
        success: false,
        message: 'Demasiadas peticiones. Intente de nuevo más tarde.',
        timestamp: new Date().toISOString()
      });
      return;
    }

    userLimit.count++;
    next();
  };
};

// Middleware para logging de actividad de usuario
export const logUserActivity = (action: string) => {
  return (req: Request, res: Response, next: NextFunction): void => {
    if (req.user) {
      console.log(`📝 Actividad: ${req.user.email} - ${action} - ${req.method} ${req.path} - IP: ${req.ip}`);
    }
    next();
  };
};
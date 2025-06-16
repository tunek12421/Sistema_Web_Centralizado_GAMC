// ========================================
// GAMC Backend Auth - Utilidades JWT
// ========================================

import jwt from 'jsonwebtoken';
import { JWTPayload, RefreshTokenPayload } from '../types/auth';

// Configuración JWT
const JWT_SECRET = process.env.JWT_SECRET || 'gamc_jwt_secret_super_secure_2024_key_never_share';
const JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET || 'gamc_jwt_refresh_secret_super_secure_2024_key';
const JWT_EXPIRES_IN = process.env.JWT_EXPIRES_IN || '15m';
const JWT_REFRESH_EXPIRES_IN = process.env.JWT_REFRESH_EXPIRES_IN || '7d';

// Generar access token
export const generateAccessToken = (payload: JWTPayload): string => {
  return jwt.sign(
    {
      userId: payload.userId,
      email: payload.email,
      role: payload.role,
      organizationalUnitId: payload.organizationalUnitId,
      sessionId: payload.sessionId
    },
    JWT_SECRET,
    {
      expiresIn: JWT_EXPIRES_IN,
      issuer: 'gamc-auth',
      audience: 'gamc-system'
    }
  );
};

// Generar refresh token
export const generateRefreshToken = (payload: RefreshTokenPayload): string => {
  return jwt.sign(
    {
      userId: payload.userId,
      sessionId: payload.sessionId,
      tokenVersion: payload.tokenVersion || 1
    },
    JWT_REFRESH_SECRET,
    {
      expiresIn: JWT_REFRESH_EXPIRES_IN,
      issuer: 'gamc-auth',
      audience: 'gamc-refresh'
    }
  );
};

// Generar ambos tokens
export const generateTokens = (payload: JWTPayload): {
  accessToken: string;
  refreshToken: string;
} => {
  const accessToken = generateAccessToken(payload);
  const refreshToken = generateRefreshToken({
    userId: payload.userId,
    sessionId: payload.sessionId,
    tokenVersion: 1
  });

  return {
    accessToken,
    refreshToken
  };
};

// Verificar access token
export const verifyAccessToken = (token: string): JWTPayload | null => {
  try {
    const decoded = jwt.verify(token, JWT_SECRET, {
      issuer: 'gamc-auth',
      audience: 'gamc-system'
    }) as JWTPayload;

    return decoded;
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      console.error('JWT Error:', error.message);
    } else if (error instanceof jwt.TokenExpiredError) {
      console.error('JWT Expired:', error.message);
    } else if (error instanceof jwt.NotBeforeError) {
      console.error('JWT Not Active:', error.message);
    }
    return null;
  }
};

// Verificar refresh token
export const verifyRefreshToken = (token: string): RefreshTokenPayload | null => {
  try {
    const decoded = jwt.verify(token, JWT_REFRESH_SECRET, {
      issuer: 'gamc-auth',
      audience: 'gamc-refresh'
    }) as RefreshTokenPayload;

    return decoded;
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      console.error('Refresh JWT Error:', error.message);
    } else if (error instanceof jwt.TokenExpiredError) {
      console.error('Refresh JWT Expired:', error.message);
    }
    return null;
  }
};

// Decodificar token sin verificar (para debugging)
export const decodeToken = (token: string): any => {
  try {
    return jwt.decode(token);
  } catch (error) {
    console.error('Error decoding token:', error);
    return null;
  }
};

// Obtener tiempo de expiración del token
export const getTokenExpiration = (token: string): Date | null => {
  try {
    const decoded = jwt.decode(token) as any;
    if (decoded && decoded.exp) {
      return new Date(decoded.exp * 1000);
    }
    return null;
  } catch (error) {
    console.error('Error getting token expiration:', error);
    return null;
  }
};

// Verificar si el token está por expirar
export const isTokenExpiringSoon = (token: string, thresholdMinutes = 5): boolean => {
  try {
    const expiration = getTokenExpiration(token);
    if (!expiration) return true;

    const now = new Date();
    const thresholdTime = new Date(now.getTime() + thresholdMinutes * 60 * 1000);

    return expiration <= thresholdTime;
  } catch (error) {
    console.error('Error checking token expiration:', error);
    return true;
  }
};

// Extraer header Authorization
export const extractTokenFromHeader = (authHeader: string | undefined): string | null => {
  if (!authHeader) return null;
  
  const parts = authHeader.split(' ');
  if (parts.length !== 2 || parts[0] !== 'Bearer') {
    return null;
  }
  
  return parts[1];
};

// Crear JWT para reset de contraseña
export const generatePasswordResetToken = (userId: string, email: string): string => {
  return jwt.sign(
    {
      userId,
      email,
      type: 'password-reset'
    },
    JWT_SECRET,
    {
      expiresIn: '30m', // 30 minutos para reset
      issuer: 'gamc-auth',
      audience: 'gamc-password-reset'
    }
  );
};

// Verificar token de reset de contraseña
export const verifyPasswordResetToken = (token: string): {
  userId: string;
  email: string;
} | null => {
  try {
    const decoded = jwt.verify(token, JWT_SECRET, {
      issuer: 'gamc-auth',
      audience: 'gamc-password-reset'
    }) as any;

    if (decoded.type !== 'password-reset') {
      return null;
    }

    return {
      userId: decoded.userId,
      email: decoded.email
    };
  } catch (error) {
    console.error('Password reset token error:', error);
    return null;
  }
};

// Blacklist de tokens (para logout)
const tokenBlacklist = new Set<string>();

// Agregar token a blacklist
export const blacklistToken = (token: string): void => {
  const decoded = decodeToken(token);
  if (decoded && decoded.exp) {
    // Solo agregar si no ha expirado
    const now = Math.floor(Date.now() / 1000);
    if (decoded.exp > now) {
      tokenBlacklist.add(token);
    }
  }
};

// Verificar si token está en blacklist
export const isTokenBlacklisted = (token: string): boolean => {
  return tokenBlacklist.has(token);
};

// Limpiar tokens expirados de blacklist (ejecutar periódicamente)
export const cleanExpiredTokensFromBlacklist = (): void => {
  const now = Math.floor(Date.now() / 1000);
  
  for (const token of tokenBlacklist) {
    const decoded = decodeToken(token);
    if (!decoded || !decoded.exp || decoded.exp <= now) {
      tokenBlacklist.delete(token);
    }
  }
};

// Configuración de limpieza automática cada hora
setInterval(cleanExpiredTokensFromBlacklist, 60 * 60 * 1000);
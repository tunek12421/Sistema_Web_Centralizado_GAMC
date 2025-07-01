// src/utils/tokenValidation.ts
// Utilidades para validación y manejo de tokens de reset de contraseña
// Incluye extracción de URL, validaciones, timers y persistencia

import type { FieldValidation, PasswordResetTokenStatus } from '../types/passwordReset';
import { PASSWORD_RESET_CONFIG } from '../types/passwordReset';

// ========================================
// EXTRACCIÓN Y LIMPIEZA DE TOKENS DESDE URL
// ========================================

/**
 * Extrae el token de reset desde los parámetros de la URL
 * Busca en: ?token=xxx, #token=xxx, /reset/:token
 */
export const extractTokenFromURL = (): string | null => {
  try {
    const currentURL = window.location.href;
    
    // Método 1: Query parameter ?token=xxx
    const urlParams = new URLSearchParams(window.location.search);
    const queryToken = urlParams.get('token');
    if (queryToken && validateTokenFormat(queryToken).isValid) {
      return queryToken;
    }
    
    // Método 2: Hash parameter #token=xxx
    if (window.location.hash) {
      const hashParams = new URLSearchParams(window.location.hash.slice(1));
      const hashToken = hashParams.get('token');
      if (hashToken && validateTokenFormat(hashToken).isValid) {
        return hashToken;
      }
    }
    
    // Método 3: Path parameter /reset/:token
    const pathMatch = currentURL.match(/\/reset(?:-password)?\/([a-f0-9]{64})/i);
    if (pathMatch && pathMatch[1]) {
      return pathMatch[1];
    }
    
    // Método 4: Cualquier secuencia de 64 caracteres hex en la URL
    const tokenMatch = currentURL.match(/[a-f0-9]{64}/i);
    if (tokenMatch) {
      return tokenMatch[0];
    }
    
    return null;
  } catch (error) {
    console.error('Error extracting token from URL:', error);
    return null;
  }
};

/**
 * Limpia el token de la URL sin recargar la página
 * Útil después de procesar el token exitosamente
 */
export const cleanTokenFromURL = (): void => {
  try {
    const url = new URL(window.location.href);
    
    // Limpiar query parameter
    url.searchParams.delete('token');
    
    // Limpiar hash si es necesario
    if (url.hash.includes('token=')) {
      const hashParams = new URLSearchParams(url.hash.slice(1));
      hashParams.delete('token');
      url.hash = hashParams.toString();
    }
    
    // Actualizar URL sin recargar
    window.history.replaceState({}, document.title, url.toString());
  } catch (error) {
    console.error('Error cleaning token from URL:', error);
  }
};

// ========================================
// VALIDACIONES DE FORMATO Y ESTADO
// ========================================

/**
 * Valida el formato básico de un token de reset
 * Verifica longitud y caracteres hexadecimales
 */
export const validateTokenFormat = (token: string): FieldValidation => {
  if (!token || typeof token !== 'string') {
    return {
      isValid: false,
      message: 'Token es requerido',
      type: 'error'
    };
  }
  
  const trimmedToken = token.trim();
  
  if (trimmedToken.length === 0) {
    return {
      isValid: false,
      message: 'Token no puede estar vacío',
      type: 'error'
    };
  }
  
  if (trimmedToken.length !== PASSWORD_RESET_CONFIG.TOKEN_LENGTH) {
    return {
      isValid: false,
      message: `Token debe tener exactamente ${PASSWORD_RESET_CONFIG.TOKEN_LENGTH} caracteres`,
      type: 'error'
    };
  }
  
  if (!PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(trimmedToken)) {
    return {
      isValid: false,
      message: 'Token contiene caracteres inválidos. Solo se permiten números y letras a-f',
      type: 'error'
    };
  }
  
  return {
    isValid: true,
    message: '✓ Formato de token válido',
    type: 'success'
  };
};

/**
 * Valida el estado de un token (si está activo, expirado, usado, etc.)
 * Esta función haría una llamada al backend para verificar estado
 */
export const validateTokenState = async (token: string): Promise<FieldValidation & { status: PasswordResetTokenStatus }> => {
  try {
    // Primero validar formato
    const formatValidation = validateTokenFormat(token);
    if (!formatValidation.isValid) {
      return {
        ...formatValidation,
        status: 'invalid'
      };
    }
    
    // TODO: Implementar llamada al backend para verificar estado
    // const response = await apiClient.get(`/auth/reset-status/${token}`);
    
    // Por ahora, simular validación básica
    // En implementación real, esto vendría del backend
    const now = Date.now();
    const storedExpiry = getStoredTokenExpiry(token);
    
    if (storedExpiry && now > storedExpiry) {
      return {
        isValid: false,
        message: 'Token ha expirado. Solicite un nuevo reset de contraseña',
        type: 'error',
        status: 'expired'
      };
    }
    
    return {
      isValid: true,
      message: '✓ Token válido y activo',
      type: 'success',
      status: 'active'
    };
    
  } catch (error) {
    console.error('Error validating token state:', error);
    return {
      isValid: false,
      message: 'Error al verificar token. Intente nuevamente',
      type: 'error',
      status: 'invalid'
    };
  }
};

// ========================================
// MANEJO DE TIEMPO Y EXPIRACIÓN
// ========================================

/**
 * Calcula tiempo restante hasta que expire un token
 * Retorna tiempo en milisegundos, 0 si ya expiró
 */
export const getTokenTimeRemaining = (tokenOrExpiry: string | number): number => {
  try {
    let expiryTime: number;
    
    if (typeof tokenOrExpiry === 'string') {
      // Es un token, buscar su tiempo de expiración
      expiryTime = getStoredTokenExpiry(tokenOrExpiry) || 0;
    } else {
      // Es un timestamp de expiración
      expiryTime = tokenOrExpiry;
    }
    
    if (!expiryTime) return 0;
    
    const now = Date.now();
    const remaining = expiryTime - now;
    
    return Math.max(0, remaining);
  } catch (error) {
    console.error('Error calculating token time remaining:', error);
    return 0;
  }
};

/**
 * Formatea tiempo restante en formato legible
 * Ej: "25 minutos 30 segundos", "5 minutos", "30 segundos"
 */
export const formatTokenTimeRemaining = (timeMs: number): string => {
  if (timeMs <= 0) return 'Expirado';
  
  const totalSeconds = Math.floor(timeMs / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  
  if (minutes > 0) {
    if (seconds > 0) {
      return `${minutes} ${minutes === 1 ? 'minuto' : 'minutos'} ${seconds} ${seconds === 1 ? 'segundo' : 'segundos'}`;
    }
    return `${minutes} ${minutes === 1 ? 'minuto' : 'minutos'}`;
  }
  
  return `${seconds} ${seconds === 1 ? 'segundo' : 'segundos'}`;
};

/**
 * Verifica si un token está cerca de expirar (< 5 minutos restantes)
 */
export const isTokenNearExpiration = (tokenOrExpiry: string | number, warningThresholdMs: number = 5 * 60 * 1000): boolean => {
  const timeRemaining = getTokenTimeRemaining(tokenOrExpiry);
  return timeRemaining > 0 && timeRemaining <= warningThresholdMs;
};

// ========================================
// PERSISTENCIA LOCAL (sessionStorage)
// ========================================

const TOKEN_STORAGE_KEY = 'gamc_reset_token';
const TOKEN_EXPIRY_KEY = 'gamc_reset_token_expiry';

/**
 * Almacena token y su tiempo de expiración en sessionStorage
 * Útil para mantener estado durante navegación
 */
export const storeToken = (token: string, expiryMs?: number): void => {
  try {
    if (!validateTokenFormat(token).isValid) {
      throw new Error('Invalid token format');
    }
    
    sessionStorage.setItem(TOKEN_STORAGE_KEY, token);
    
    if (expiryMs) {
      sessionStorage.setItem(TOKEN_EXPIRY_KEY, expiryMs.toString());
    } else {
      // Default: 30 minutos desde ahora
      const defaultExpiry = Date.now() + PASSWORD_RESET_CONFIG.TOKEN_EXPIRY;
      sessionStorage.setItem(TOKEN_EXPIRY_KEY, defaultExpiry.toString());
    }
  } catch (error) {
    console.error('Error storing token:', error);
  }
};

/**
 * Recupera token almacenado desde sessionStorage
 * Retorna null si no existe o ha expirado
 */
export const retrieveStoredToken = (): string | null => {
  try {
    const token = sessionStorage.getItem(TOKEN_STORAGE_KEY);
    if (!token) return null;
    
    const expiry = sessionStorage.getItem(TOKEN_EXPIRY_KEY);
    if (expiry && Date.now() > parseInt(expiry)) {
      // Token expirado, limpiar storage
      clearStoredToken();
      return null;
    }
    
    return token;
  } catch (error) {
    console.error('Error retrieving stored token:', error);
    return null;
  }
};

/**
 * Obtiene tiempo de expiración de un token almacenado
 */
export const getStoredTokenExpiry = (token?: string): number | null => {
  try {
    const storedToken = sessionStorage.getItem(TOKEN_STORAGE_KEY);
    if (token && storedToken !== token) return null;
    
    const expiry = sessionStorage.getItem(TOKEN_EXPIRY_KEY);
    return expiry ? parseInt(expiry) : null;
  } catch (error) {
    console.error('Error getting stored token expiry:', error);
    return null;
  }
};

/**
 * Limpia token almacenado del sessionStorage
 */
export const clearStoredToken = (): void => {
  try {
    sessionStorage.removeItem(TOKEN_STORAGE_KEY);
    sessionStorage.removeItem(TOKEN_EXPIRY_KEY);
  } catch (error) {
    console.error('Error clearing stored token:', error);
  }
};

// ========================================
// UTILIDADES DE TIMERS PARA COMPONENTES
// ========================================

/**
 * Crea un timer que ejecuta callback cada segundo con tiempo restante
 * Útil para countdown en UI
 */
export const createTokenTimer = (
  tokenOrExpiry: string | number,
  onTick: (timeRemaining: number, formatted: string) => void,
  onExpired?: () => void
): { stop: () => void; getTimeRemaining: () => number } => {
  let intervalId: NodeJS.Timeout | null = null;
  
  const tick = () => {
    const timeRemaining = getTokenTimeRemaining(tokenOrExpiry);
    const formatted = formatTokenTimeRemaining(timeRemaining);
    
    onTick(timeRemaining, formatted);
    
    if (timeRemaining <= 0) {
      stop();
      if (onExpired) onExpired();
    }
  };
  
  const start = () => {
    // Ejecutar inmediatamente
    tick();
    // Después cada segundo
    intervalId = setInterval(tick, 1000);
  };
  
  const stop = () => {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
  };
  
  const getTimeRemaining = () => getTokenTimeRemaining(tokenOrExpiry);
  
  // Iniciar automáticamente
  start();
  
  return { stop, getTimeRemaining };
};

// ========================================
// HELPERS DE URL PARA DIFERENTES FLUJOS
// ========================================

/**
 * Genera URL para reset de contraseña con token
 */
export const generateResetURL = (token: string, baseURL?: string): string => {
  const base = baseURL || window.location.origin;
  return `${base}/reset-password?token=${token}`;
};

/**
 * Genera URL para solicitar nuevo reset
 */
export const generateForgotPasswordURL = (baseURL?: string): string => {
  const base = baseURL || window.location.origin;
  return `${base}/forgot-password`;
};

/**
 * Valida si la URL actual corresponde a un flujo de reset
 */
export const isResetPasswordURL = (): boolean => {
  const pathname = window.location.pathname;
  return pathname.includes('/reset') || 
         pathname.includes('/forgot') ||
         window.location.search.includes('token=') ||
         window.location.hash.includes('token=');
};
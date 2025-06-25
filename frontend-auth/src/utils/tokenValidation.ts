// src/utils/tokenValidation.ts
// Utilidades de validación de tokens para el módulo de reset de contraseña
// Compatible con los tokens generados por el backend GAMC

import { 
  FieldValidation, 
  PasswordResetTokenStatus,
  PasswordResetErrorType,
  PASSWORD_RESET_CONFIG, 
  PASSWORD_RESET_MESSAGES 
} from '../types/passwordReset';

// ========================================
// VALIDACIONES DE TOKEN
// ========================================

/**
 * Validación completa de formato de token
 * El backend genera tokens de 64 caracteres hexadecimales
 */
export const validateTokenFormat = (token: string): FieldValidation => {
  // Token vacío
  if (!token || !token.trim()) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_REQUIRED, 
      type: 'error' 
    };
  }

  // Limpiar espacios en blanco
  const cleanToken = token.trim();

  // Longitud exacta (64 caracteres)
  if (cleanToken.length !== PASSWORD_RESET_CONFIG.TOKEN_LENGTH) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_INVALID_FORMAT, 
      type: 'error' 
    };
  }

  // Solo caracteres hexadecimales válidos
  if (!PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(cleanToken)) {
    return { 
      isValid: false, 
      message: 'Token contiene caracteres inválidos (solo 0-9, a-f permitidos)', 
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
 * Validación de token con verificación de estado
 * Incluye checks de expiración y uso previo
 */
export const validateTokenState = (
  token: string,
  expiresAt?: string,
  usedAt?: string,
  isActive?: boolean
): FieldValidation => {
  // Primero validar formato
  const formatValidation = validateTokenFormat(token);
  if (!formatValidation.isValid) {
    return formatValidation;
  }

  // Verificar si está activo
  if (isActive === false) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_INVALID, 
      type: 'error' 
    };
  }

  // Verificar si ya fue usado
  if (usedAt) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_USED, 
      type: 'error' 
    };
  }

  // Verificar expiración
  if (expiresAt) {
    const expirationTime = new Date(expiresAt).getTime();
    const currentTime = Date.now();
    
    if (currentTime > expirationTime) {
      return { 
        isValid: false, 
        message: PASSWORD_RESET_MESSAGES.TOKEN_EXPIRED, 
        type: 'error' 
      };
    }
  }

  return { 
    isValid: true, 
    message: '✓ Token válido y activo', 
    type: 'success' 
  };
};

// ========================================
// EXTRACCIÓN DE TOKEN DESDE URL
// ========================================

/**
 * Extrae token desde parámetros de URL
 * Soporte para diferentes formatos: ?token=... o ?reset=... o hash #token
 */
export const extractTokenFromURL = (url?: string): string | null => {
  const currentUrl = url || window.location.href;
  
  try {
    const urlObj = new URL(currentUrl);
    
    // Intentar diferentes parámetros comunes
    const tokenParams = ['token', 'reset', 'resetToken', 'passwordReset'];
    
    for (const param of tokenParams) {
      const token = urlObj.searchParams.get(param);
      if (token && token.trim()) {
        return token.trim();
      }
    }
    
    // Intentar extraer desde hash (fragment)
    const hash = urlObj.hash.substring(1); // Remove #
    if (hash && hash.length === PASSWORD_RESET_CONFIG.TOKEN_LENGTH) {
      const hashValidation = validateTokenFormat(hash);
      if (hashValidation.isValid) {
        return hash;
      }
    }
    
    return null;
  } catch (error) {
    console.warn('Error al extraer token desde URL:', error);
    return null;
  }
};

/**
 * Extrae token desde localStorage (backup/cache)
 * Útil para recuperar tokens temporalmente almacenados
 */
export const extractTokenFromStorage = (key: string = 'passwordResetToken'): string | null => {
  try {
    const storedToken = localStorage.getItem(key);
    if (storedToken) {
      const validation = validateTokenFormat(storedToken);
      return validation.isValid ? storedToken : null;
    }
    return null;
  } catch (error) {
    console.warn('Error al extraer token desde localStorage:', error);
    return null;
  }
};

/**
 * Limpia token de URL para mejorar seguridad
 * Remueve el token de la URL visible sin recargar la página
 */
export const cleanTokenFromURL = (): void => {
  try {
    const url = new URL(window.location.href);
    const tokenParams = ['token', 'reset', 'resetToken', 'passwordReset'];
    let hasChanges = false;
    
    // Remover parámetros de token
    for (const param of tokenParams) {
      if (url.searchParams.has(param)) {
        url.searchParams.delete(param);
        hasChanges = true;
      }
    }
    
    // Remover hash si parece ser un token
    if (url.hash && url.hash.length === PASSWORD_RESET_CONFIG.TOKEN_LENGTH + 1) {
      url.hash = '';
      hasChanges = true;
    }
    
    // Actualizar URL si hubo cambios
    if (hasChanges) {
      window.history.replaceState({}, '', url.toString());
    }
  } catch (error) {
    console.warn('Error al limpiar token de URL:', error);
  }
};

// ========================================
// UTILIDADES DE ESTADO DE TOKEN
// ========================================

/**
 * Determina el estado actual de un token
 */
export const getTokenStatus = (
  token: string,
  expiresAt?: string,
  usedAt?: string,
  isActive?: boolean
): PasswordResetTokenStatus => {
  const formatValidation = validateTokenFormat(token);
  if (!formatValidation.isValid) {
    return PasswordResetTokenStatus.INVALID;
  }

  if (isActive === false) {
    return PasswordResetTokenStatus.INVALID;
  }

  if (usedAt) {
    return PasswordResetTokenStatus.USED;
  }

  if (expiresAt) {
    const expirationTime = new Date(expiresAt).getTime();
    const currentTime = Date.now();
    
    if (currentTime > expirationTime) {
      return PasswordResetTokenStatus.EXPIRED;
    }
  }

  return PasswordResetTokenStatus.ACTIVE;
};

/**
 * Calcula tiempo restante hasta expiración
 */
export const getTokenTimeRemaining = (expiresAt: string): number => {
  try {
    const expirationTime = new Date(expiresAt).getTime();
    const currentTime = Date.now();
    const remaining = expirationTime - currentTime;
    
    return Math.max(0, Math.ceil(remaining / 1000)); // segundos
  } catch (error) {
    console.warn('Error al calcular tiempo restante:', error);
    return 0;
  }
};

/**
 * Formatea tiempo restante en formato legible
 */
export const formatTokenTimeRemaining = (seconds: number): string => {
  if (seconds <= 0) return 'Expirado';
  
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  
  if (minutes > 0) {
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')} min`;
  }
  
  return `${remainingSeconds}s`;
};

/**
 * Verifica si un token está próximo a expirar (menos de 5 minutos)
 */
export const isTokenNearExpiration = (expiresAt: string, thresholdMinutes: number = 5): boolean => {
  const remainingSeconds = getTokenTimeRemaining(expiresAt);
  return remainingSeconds > 0 && remainingSeconds <= (thresholdMinutes * 60);
};

// ========================================
// GENERACIÓN DE ENLACES DE RESET
// ========================================

/**
 * Genera URL completa para reset de contraseña
 * Útil para testing o visualización
 */
export const generateResetURL = (
  token: string, 
  baseUrl?: string, 
  useHash?: boolean
): string => {
  const base = baseUrl || window.location.origin;
  const resetPath = '/reset-password';
  
  if (useHash) {
    return `${base}${resetPath}#${token}`;
  } else {
    return `${base}${resetPath}?token=${token}`;
  }
};

/**
 * Valida si una URL contiene un token de reset válido
 */
export const validateResetURL = (url: string): { isValid: boolean; token?: string; error?: string } => {
  try {
    const urlObj = new URL(url);
    const token = extractTokenFromURL(url);
    
    if (!token) {
      return { 
        isValid: false, 
        error: 'No se encontró token en la URL' 
      };
    }
    
    const validation = validateTokenFormat(token);
    if (!validation.isValid) {
      return { 
        isValid: false, 
        error: validation.message 
      };
    }
    
    return { 
      isValid: true, 
      token 
    };
  } catch (error) {
    return { 
      isValid: false, 
      error: 'URL inválida' 
    };
  }
};

// ========================================
// GESTIÓN DE TOKENS EN LOCALSTORAGE
// ========================================

/**
 * Guarda token temporalmente en localStorage
 * Útil para persistir token entre navegación
 */
export const storeToken = (
  token: string, 
  expiresAt?: string, 
  key: string = 'passwordResetToken'
): boolean => {
  try {
    const validation = validateTokenFormat(token);
    if (!validation.isValid) {
      console.warn('Intento de guardar token inválido:', validation.message);
      return false;
    }
    
    const tokenData = {
      token,
      expiresAt,
      storedAt: new Date().toISOString()
    };
    
    localStorage.setItem(key, JSON.stringify(tokenData));
    return true;
  } catch (error) {
    console.error('Error al guardar token:', error);
    return false;
  }
};

/**
 * Recupera token desde localStorage con validación
 */
export const retrieveStoredToken = (key: string = 'passwordResetToken'): {
  token?: string;
  expiresAt?: string;
  isValid: boolean;
  error?: string;
} => {
  try {
    const storedData = localStorage.getItem(key);
    if (!storedData) {
      return { isValid: false, error: 'No hay token almacenado' };
    }
    
    const tokenData = JSON.parse(storedData);
    const { token, expiresAt } = tokenData;
    
    if (!token) {
      return { isValid: false, error: 'Token faltante en datos almacenados' };
    }
    
    const validation = validateTokenState(token, expiresAt);
    if (!validation.isValid) {
      // Limpiar token inválido
      clearStoredToken(key);
      return { isValid: false, error: validation.message };
    }
    
    return { token, expiresAt, isValid: true };
  } catch (error) {
    console.error('Error al recuperar token almacenado:', error);
    clearStoredToken(key);
    return { isValid: false, error: 'Error al leer token almacenado' };
  }
};

/**
 * Limpia token de localStorage
 */
export const clearStoredToken = (key: string = 'passwordResetToken'): void => {
  try {
    localStorage.removeItem(key);
  } catch (error) {
    console.warn('Error al limpiar token almacenado:', error);
  }
};

// ========================================
// UTILIDADES DE DEBUGGING
// ========================================

/**
 * Función de debugging para analizar tokens
 * Solo disponible en desarrollo
 */
export const debugToken = (token: string): void => {
  if (process.env.NODE_ENV !== 'development') return;
  
  console.group('🔍 Debug Token Analysis');
  console.log('Token:', token);
  console.log('Length:', token.length);
  console.log('Format Valid:', validateTokenFormat(token));
  console.log('Pattern Match:', PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(token));
  console.log('Hex Characters:', /^[a-f0-9]+$/i.test(token));
  console.groupEnd();
};

/**
 * Genera token de prueba para desarrollo
 * Solo disponible en desarrollo
 */
export const generateTestToken = (): string => {
  if (process.env.NODE_ENV !== 'development') {
    throw new Error('generateTestToken solo disponible en desarrollo');
  }
  
  // Generar 64 caracteres hexadecimales aleatorios
  return Array.from({ length: 64 }, () => 
    Math.floor(Math.random() * 16).toString(16)
  ).join('');
};

// ========================================
// EXPORT POR DEFECTO
// ========================================

export default {
  // Validaciones principales
  validateTokenFormat,
  validateTokenState,
  
  // Extracción de tokens
  extractTokenFromURL,
  extractTokenFromStorage,
  cleanTokenFromURL,
  
  // Estado y tiempo
  getTokenStatus,
  getTokenTimeRemaining,
  formatTokenTimeRemaining,
  isTokenNearExpiration,
  
  // URLs y enlaces
  generateResetURL,
  validateResetURL,
  
  // Gestión de almacenamiento
  storeToken,
  retrieveStoredToken,
  clearStoredToken,
  
  // Debugging (desarrollo)
  debugToken,
  generateTestToken
};
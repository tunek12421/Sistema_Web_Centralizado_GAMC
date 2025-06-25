// src/types/passwordReset.ts
// Interfaces TypeScript para el módulo de reset de contraseña
// Compatible con el backend GAMC y reutiliza tipos existentes

import { ApiResponse, FieldValidation } from './auth';

// ========================================
// INTERFACES PRINCIPALES DE RESET
// ========================================

/**
 * Solicitud de reset de contraseña
 * Usada en: ForgotPasswordForm, passwordResetService.requestPasswordReset()
 */
export interface PasswordResetRequest {
  email: string; // Debe ser email institucional @gamc.gov.bo
}

/**
 * Confirmación de reset de contraseña
 * Usada en: ResetPasswordForm, passwordResetService.confirmPasswordReset()
 */
export interface PasswordResetConfirm {
  token: string;       // Token de 64 caracteres hex
  newPassword: string; // Nueva contraseña que cumple políticas
}

/**
 * Token de reset de contraseña (modelo completo del backend)
 * Usada en: Estado de admin, consultas de tokens
 */
export interface PasswordResetToken {
  id: number;
  userId: string;
  expiresAt: string;    // ISO string datetime
  createdAt: string;    // ISO string datetime
  usedAt?: string;      // ISO string datetime (opcional)
  requestIp: string;
  userAgent?: string;
  isActive: boolean;
  
  // Metadatos de auditoría
  emailSentAt?: string;    // ISO string datetime (opcional)
  emailOpenedAt?: string;  // ISO string datetime (opcional) 
  attemptsCount: number;
  
  // Relación con usuario (opcional, solo para admin)
  user?: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
  };
}

// ========================================
// RESPUESTAS ESPECÍFICAS DE LA API
// ========================================

/**
 * Respuesta de solicitud de reset exitosa
 * Siempre devuelve el mismo mensaje por seguridad
 */
export interface PasswordResetRequestResponse extends ApiResponse {
  data: {
    message: string; // "Si el email existe y es válido, recibirás un enlace..."
    note: string;    // "Solo emails @gamc.gov.bo pueden solicitar reset..."
  };
}

/**
 * Respuesta de confirmación de reset exitosa
 */
export interface PasswordResetConfirmResponse extends ApiResponse {
  data: {
    message: string; // "Su contraseña ha sido cambiada exitosamente..."
    note: string;    // "Inicie sesión con su nueva contraseña"
  };
}

/**
 * Respuesta de estado de tokens de reset (para usuarios autenticados)
 */
export interface PasswordResetStatusResponse extends ApiResponse {
  data: {
    tokens: PasswordResetToken[];
    count: number;
  };
}

/**
 * Respuesta de limpieza de tokens (solo admins)
 */
export interface PasswordResetCleanupResponse extends ApiResponse {
  data: {
    cleanedTokens: number;
    timestamp: string; // ISO string datetime
  };
}

// ========================================
// ENUMS Y CONSTANTES
// ========================================

/**
 * Estados posibles de un token de reset
 */
export enum PasswordResetTokenStatus {
  ACTIVE = 'active',      // Token válido y activo
  EXPIRED = 'expired',    // Token expirado (>30 min)
  USED = 'used',         // Token ya utilizado
  INVALID = 'invalid'     // Token inválido o no encontrado
}

/**
 * Tipos de errores específicos del reset de contraseña
 */
export enum PasswordResetErrorType {
  // Errores de solicitud
  EMAIL_NOT_INSTITUTIONAL = 'email_not_institutional',
  RATE_LIMIT_EXCEEDED = 'rate_limit_exceeded',
  EMAIL_NOT_FOUND = 'email_not_found',
  
  // Errores de confirmación
  TOKEN_INVALID = 'token_invalid',
  TOKEN_EXPIRED = 'token_expired',
  TOKEN_USED = 'token_used',
  PASSWORD_WEAK = 'password_weak',
  
  // Errores generales
  NETWORK_ERROR = 'network_error',
  SERVER_ERROR = 'server_error',
  VALIDATION_ERROR = 'validation_error'
}

// ========================================
// INTERFACES PARA COMPONENTES UI
// ========================================

/**
 * Estado del formulario de solicitud de reset
 * Usada en: ForgotPasswordForm, usePasswordReset hook
 */
export interface ForgotPasswordFormState {
  email: string;
  isLoading: boolean;
  isSubmitted: boolean;
  validation: FieldValidation;
  lastSubmissionTime?: number; // timestamp para rate limiting UI
}

/**
 * Estado del formulario de confirmación de reset
 * Usada en: ResetPasswordForm, usePasswordReset hook
 */
export interface ResetPasswordFormState {
  token: string;
  newPassword: string;
  confirmPassword: string;
  isLoading: boolean;
  isSubmitted: boolean;
  validation: {
    token: FieldValidation;
    password: FieldValidation;
    confirmPassword: FieldValidation;
  };
}

/**
 * Estado general del proceso de reset
 * Usada en: usePasswordReset hook, páginas completas
 */
export interface PasswordResetState {
  step: 'request' | 'email_sent' | 'confirm' | 'success' | 'error';
  email?: string;
  token?: string;
  error?: PasswordResetErrorType;
  errorMessage?: string;
  successMessage?: string;
  requestedAt?: number; // timestamp
  expiresAt?: number;   // timestamp
}

// ========================================
// INTERFACES PARA HOOKS Y SERVICIOS
// ========================================

/**
 * Opciones para el hook usePasswordReset
 */
export interface UsePasswordResetOptions {
  autoRedirect?: boolean;           // Redirect automático después del éxito
  redirectDelay?: number;          // Delay en ms para redirect (default: 3000)
  enableRateLimitUI?: boolean;     // Mostrar cooldown en UI
  onSuccess?: (step: 'request' | 'confirm') => void;
  onError?: (error: PasswordResetErrorType, message: string) => void;
}

/**
 * Resultado del hook usePasswordReset
 */
export interface UsePasswordResetResult {
  // Estado
  state: PasswordResetState;
  formStates: {
    forgot: ForgotPasswordFormState;
    reset: ResetPasswordFormState;
  };
  
  // Acciones
  requestReset: (email: string) => Promise<void>;
  confirmReset: (token: string, newPassword: string) => Promise<void>;
  validateToken: (token: string) => FieldValidation;
  validatePassword: (password: string) => FieldValidation;
  
  // Utilidades
  canSubmitRequest: boolean;
  canSubmitConfirm: boolean;
  timeUntilNextRequest: number; // segundos restantes para rate limit
  
  // Admin (solo para role admin)
  getResetStatus?: () => Promise<PasswordResetToken[]>;
  cleanupExpiredTokens?: () => Promise<number>;
}

// ========================================
// CLASES DE ERROR PERSONALIZADAS
// ========================================

export class PasswordResetError extends Error {
  constructor(
    public type: PasswordResetErrorType,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'PasswordResetError';
  }
}

// ========================================
// CONFIGURACIÓN Y CONSTANTES
// ========================================

/**
 * Configuración del módulo de reset de contraseña
 */
export const PASSWORD_RESET_CONFIG = {
  // Validaciones
  TOKEN_LENGTH: 64,
  MIN_PASSWORD_LENGTH: 8,
  ALLOWED_SPECIAL_CHARS: '@$!%*?&',
  
  // Timeouts (en milisegundos)
  REQUEST_TIMEOUT: 10000,        // 10 segundos
  TOKEN_EXPIRY: 30 * 60 * 1000, // 30 minutos
  RATE_LIMIT_WINDOW: 5 * 60 * 1000, // 5 minutos
  
  // UI
  AUTO_REDIRECT_DELAY: 3000,     // 3 segundos
  SUCCESS_MESSAGE_DURATION: 5000, // 5 segundos
  
  // Regex patterns
  EMAIL_PATTERN: /^[^\s@]+@gamc\.gov\.bo$/,
  TOKEN_PATTERN: /^[a-f0-9]{64}$/i,
  PASSWORD_PATTERN: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
  
  // Mensajes de error
  ERROR_MESSAGES: {
    [PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL]: 'Solo emails @gamc.gov.bo pueden solicitar reset',
    [PasswordResetErrorType.RATE_LIMIT_EXCEEDED]: 'Debe esperar 5 minutos entre solicitudes',
    [PasswordResetErrorType.TOKEN_INVALID]: 'Token de reset inválido',
    [PasswordResetErrorType.TOKEN_EXPIRED]: 'Token de reset expirado',
    [PasswordResetErrorType.TOKEN_USED]: 'Token ya utilizado',
    [PasswordResetErrorType.PASSWORD_WEAK]: 'Contraseña no cumple requisitos',
    [PasswordResetErrorType.NETWORK_ERROR]: 'Error de conexión',
    [PasswordResetErrorType.SERVER_ERROR]: 'Error del servidor',
    [PasswordResetErrorType.VALIDATION_ERROR]: 'Error de validación'
  }
} as const;

// ========================================
// UTILIDADES Y HELPERS
// ========================================

/**
 * Verifica si un string es un token válido
 */
export const isValidResetToken = (token: string): boolean => {
  return typeof token === 'string' && 
         token.length === PASSWORD_RESET_CONFIG.TOKEN_LENGTH &&
         PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(token);
};

/**
 * Obtiene el mensaje de error apropiado para un tipo de error
 */
export const getPasswordResetErrorMessage = (type: PasswordResetErrorType): string => {
  return PASSWORD_RESET_CONFIG.ERROR_MESSAGES[type] || 'Error desconocido';
};

/**
 * Valida un email para reset de contraseña
 */
export const validateResetEmail = (email: string): FieldValidation => {
  if (!email || !email.trim()) {
    return {
      isValid: false,
      message: 'Email es requerido',
      type: 'error'
    };
  }

  if (!PASSWORD_RESET_CONFIG.EMAIL_PATTERN.test(email.trim())) {
    return {
      isValid: false,
      message: 'Solo emails @gamc.gov.bo pueden solicitar reset',
      type: 'error'
    };
  }

  return {
    isValid: true,
    message: '✓ Email válido',
    type: 'success'
  };
};

/**
 * Calcula tiempo restante para rate limiting
 */
export const getRateLimitTimeRemaining = (): number => {
  try {
    const lastRequest = localStorage.getItem('gamc_last_reset_request');
    if (!lastRequest) return 0;
    
    const lastRequestTime = parseInt(lastRequest);
    const now = Date.now();
    const timeElapsed = now - lastRequestTime;
    const cooldownTime = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW;
    
    return Math.max(0, cooldownTime - timeElapsed);
  } catch {
    return 0;
  }
};
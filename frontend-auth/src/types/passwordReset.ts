// src/types/passwordReset.ts
// Interfaces TypeScript para el módulo de reset de contraseña
// Compatible con el backend GAMC y reutiliza tipos existentes

import { ApiResponse } from './auth';

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
 * Estado de validación de campo en tiempo real
 * Usada en: ForgotPasswordForm, ResetPasswordForm
 */
export interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

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
  PASSWORD_PATTERN: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/
} as const;

/**
 * Mensajes de error predefinidos para consistencia en UI
 */
export const PASSWORD_RESET_MESSAGES = {
  // Solicitud de reset
  EMAIL_REQUIRED: 'El email es requerido',
  EMAIL_INVALID_FORMAT: 'Formato de email inválido',
  EMAIL_NOT_INSTITUTIONAL: 'Solo emails @gamc.gov.bo pueden solicitar reset',
  RATE_LIMIT_EXCEEDED: 'Debe esperar 5 minutos entre solicitudes',
  REQUEST_SUCCESS: 'Si el email existe, recibirás un enlace de reset',
  
  // Confirmación de reset
  TOKEN_REQUIRED: 'El token es requerido',
  TOKEN_INVALID_FORMAT: 'Token debe tener exactamente 64 caracteres',
  TOKEN_EXPIRED: 'El token ha expirado. Solicite un nuevo reset',
  TOKEN_USED: 'Este token ya fue utilizado',
  TOKEN_INVALID: 'Token inválido',
  
  // Contraseña
  PASSWORD_REQUIRED: 'La nueva contraseña es requerida',
  PASSWORD_TOO_SHORT: 'Mínimo 8 caracteres',
  PASSWORD_MISSING_UPPERCASE: 'Debe contener al menos una mayúscula',
  PASSWORD_MISSING_LOWERCASE: 'Debe contener al menos una minúscula',
  PASSWORD_MISSING_NUMBER: 'Debe contener al menos un número',
  PASSWORD_MISSING_SPECIAL: 'Debe contener al menos un símbolo (@$!%*?&)',
  PASSWORD_CONFIRM_MISMATCH: 'Las contraseñas no coinciden',
  
  // Éxito
  RESET_SUCCESS: 'Contraseña cambiada exitosamente',
  REDIRECT_MESSAGE: 'Redirigiendo al login...',
  
  // Errores generales
  NETWORK_ERROR: 'Error de conexión. Verifique su conexión a internet',
  SERVER_ERROR: 'Error del servidor. Intente nuevamente',
  UNKNOWN_ERROR: 'Error inesperado. Contacte al soporte técnico'
} as const;
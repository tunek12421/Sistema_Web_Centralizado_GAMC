// src/types/passwordReset.ts
// Interfaces TypeScript para el módulo de reset de contraseña
// Compatible con el backend GAMC y reutiliza tipos existentes

// ========================================
// INTERFACES BÁSICAS
// ========================================

export interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
  timestamp: string;
}

// ========================================
// TIPOS SIMPLES
// ========================================

export type PasswordResetTokenStatus = 'active' | 'expired' | 'used' | 'invalid';

export type PasswordResetErrorType = 
  | 'email_not_institutional'
  | 'rate_limit_exceeded'
  | 'email_not_found'
  | 'token_invalid'
  | 'token_expired'
  | 'token_used'
  | 'password_weak'
  | 'network_error'
  | 'server_error'
  | 'validation_error';

// ========================================
// INTERFACES PRINCIPALES DE RESET
// ========================================

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirm {
  token: string;
  newPassword: string;
}

export interface PasswordResetToken {
  id: number;
  userId: string;
  expiresAt: string;
  createdAt: string;
  usedAt?: string;
  requestIp: string;
  userAgent?: string;
  isActive: boolean;
  emailSentAt?: string;
  emailOpenedAt?: string;
  attemptsCount: number;
  user?: {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
  };
}

// ========================================
// RESPUESTAS DE LA API
// ========================================

export interface PasswordResetRequestResponse extends ApiResponse {
  data: {
    message: string;
    note: string;
  };
}

export interface PasswordResetConfirmResponse extends ApiResponse {
  data: {
    message: string;
    note: string;
  };
}

export interface PasswordResetStatusResponse extends ApiResponse {
  data: {
    tokens: PasswordResetToken[];
    count: number;
  };
}

export interface PasswordResetCleanupResponse extends ApiResponse {
  data: {
    cleanedTokens: number;
    timestamp: string;
  };
}

// ========================================
// INTERFACES PARA COMPONENTES UI
// ========================================

export interface ForgotPasswordFormState {
  email: string;
  isLoading: boolean;
  isSubmitted: boolean;
  validation: FieldValidation;
  lastSubmissionTime?: number;
}

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

export interface PasswordResetState {
  step: 'request' | 'email_sent' | 'confirm' | 'success' | 'error';
  email?: string;
  token?: string;
  error?: PasswordResetErrorType;
  errorMessage?: string;
  successMessage?: string;
  requestedAt?: number;
  expiresAt?: number;
}

// ========================================
// INTERFACES PARA HOOKS Y SERVICIOS
// ========================================

export interface UsePasswordResetOptions {
  autoRedirect?: boolean;
  redirectDelay?: number;
  enableRateLimitUI?: boolean;
  onSuccess?: (step: 'request' | 'confirm') => void;
  onError?: (error: PasswordResetErrorType, message: string) => void;
}

export interface UsePasswordResetResult {
  state: PasswordResetState;
  formStates: {
    forgot: ForgotPasswordFormState;
    reset: ResetPasswordFormState;
  };
  requestReset: (email: string) => Promise<void>;
  confirmReset: (token: string, newPassword: string) => Promise<void>;
  validateToken: (token: string) => FieldValidation;
  validatePassword: (password: string) => FieldValidation;
  canSubmitRequest: boolean;
  canSubmitConfirm: boolean;
  timeUntilNextRequest: number;
  getResetStatus?: () => Promise<PasswordResetToken[]>;
  cleanupExpiredTokens?: () => Promise<number>;
}

// ========================================
// CLASES DE ERROR
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

export const PASSWORD_RESET_CONFIG = {
  TOKEN_LENGTH: 64,
  MIN_PASSWORD_LENGTH: 8,
  ALLOWED_SPECIAL_CHARS: '@$!%*?&',
  REQUEST_TIMEOUT: 10000,
  TOKEN_EXPIRY: 30 * 60 * 1000,
  RATE_LIMIT_WINDOW: 5 * 60 * 1000,
  AUTO_REDIRECT_DELAY: 3000,
  SUCCESS_MESSAGE_DURATION: 5000,
  EMAIL_PATTERN: /^[^\s@]+@gamc\.gov\.bo$/,
  TOKEN_PATTERN: /^[a-f0-9]{64}$/i,
  PASSWORD_PATTERN: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/
};

export const PASSWORD_RESET_MESSAGES = {
  email_not_institutional: 'Solo emails @gamc.gov.bo pueden solicitar reset',
  rate_limit_exceeded: 'Debe esperar 5 minutos entre solicitudes',
  token_invalid: 'Token de reset inválido',
  token_expired: 'Token de reset expirado',
  token_used: 'Token ya utilizado',
  password_weak: 'Contraseña no cumple requisitos',
  network_error: 'Error de conexión',
  server_error: 'Error del servidor',
  validation_error: 'Error de validación',
  email_not_found: 'Email no encontrado',
  EMAIL_REQUIRED: 'Email es requerido',
  EMAIL_INVALID_FORMAT: 'Formato de email inválido',
  PASSWORD_REQUIRED: 'Contraseña es requerida',
  PASSWORD_TOO_SHORT: 'Contraseña debe tener al menos 8 caracteres',
  PASSWORD_CONFIRM_MISMATCH: 'Las contraseñas no coinciden',
  TOKEN_REQUIRED: 'Token es requerido',
  TOKEN_INVALID_FORMAT: 'Formato de token inválido'
};

// ========================================
// UTILIDADES Y HELPERS
// ========================================

export const isValidResetToken = (token: string): boolean => {
  return typeof token === 'string' && 
         token.length === PASSWORD_RESET_CONFIG.TOKEN_LENGTH &&
         PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(token);
};

export const getPasswordResetErrorMessage = (type: PasswordResetErrorType): string => {
  const messages = {
    email_not_institutional: 'Solo emails @gamc.gov.bo pueden solicitar reset',
    rate_limit_exceeded: 'Debe esperar 5 minutos entre solicitudes',
    token_invalid: 'Token de reset inválido',
    token_expired: 'Token de reset expirado',
    token_used: 'Token ya utilizado',
    password_weak: 'Contraseña no cumple requisitos',
    network_error: 'Error de conexión',
    server_error: 'Error del servidor',
    validation_error: 'Error de validación',
    email_not_found: 'Email no encontrado'
  };
  return messages[type] || 'Error desconocido';
};

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
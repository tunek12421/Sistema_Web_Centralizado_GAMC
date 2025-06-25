// src/types/index.ts
// Archivo central de exportación de tipos
// Facilita las importaciones y evita conflictos

// ========================================
// TIPOS DE AUTENTICACIÓN
// ========================================
export {
  // Interfaces principales
  User,
  LoginCredentials,
  RegisterData,
  AuthResponse,
  ApiResponse,
  ApiError,
  
  // Interfaces de perfil
  UserProfile,
  ChangePasswordRequest,
  RefreshTokenResponse,
  TokenVerificationResponse,
  
  // Interfaces de validación
  FieldValidation,
  FormValidationState,
  
  // Enums y constantes
  UserRole,
  AuthErrorType,
  AUTH_CONFIG,
  
  // Guards de tipos
  isUser,
  isAuthResponse,
  isApiResponse
} from './auth';

// ========================================
// TIPOS DE PASSWORD RESET
// ========================================
export {
  // Interfaces principales
  PasswordResetRequest,
  PasswordResetConfirm,
  PasswordResetToken,
  
  // Respuestas de API
  PasswordResetRequestResponse,
  PasswordResetConfirmResponse,
  PasswordResetStatusResponse,
  PasswordResetCleanupResponse,
  
  // Estados de formularios
  ForgotPasswordFormState,
  ResetPasswordFormState,
  PasswordResetState,
  
  // Hooks y servicios
  UsePasswordResetOptions,
  UsePasswordResetResult,
  
  // Enums y constantes
  PasswordResetTokenStatus,
  PasswordResetErrorType,
  PASSWORD_RESET_CONFIG,
  
  // Clases de error
  PasswordResetError,
  
  // Utilidades
  isValidResetToken,
  getPasswordResetErrorMessage,
  validateResetEmail,
  getRateLimitTimeRemaining
} from './passwordReset';

// ========================================
// RE-EXPORTACIONES PARA COMPATIBILIDAD
// ========================================

// Importaciones legacy que otros archivos podrían estar usando
export type {
  User as UserType,
  AuthResponse as AuthResponseType,
  ApiResponse as ApiResponseType,
  FieldValidation as ValidationResult
} from './auth';

export type {
  PasswordResetRequest as ResetRequest,
  PasswordResetConfirm as ResetConfirm,
  PasswordResetError as ResetError
} from './passwordReset';
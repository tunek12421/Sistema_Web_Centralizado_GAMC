// src/utils/errorHandling.ts
// Utilidades centralizadas para manejo de errores en el módulo de password reset
// Convierte errores del backend en mensajes user-friendly y códigos específicos

import { PasswordResetErrorType, FieldValidation } from '../types/passwordReset';
import { PasswordResetError } from '../services/passwordResetService';
import { SecurityQuestionError, SecurityQuestionErrorType } from '../services/securityQuestionsService';

// ========================================
// MAPEO DE ERRORES HTTP A MENSAJES
// ========================================

/**
 * Mapea códigos de estado HTTP a mensajes user-friendly
 */
export const HTTP_ERROR_MESSAGES: Record<number, string> = {
  400: 'Datos inválidos. Verifique la información ingresada',
  401: 'Sesión expirada. Inicie sesión nuevamente',
  403: 'Solo usuarios con email @gamc.gov.bo pueden usar esta función',
  404: 'Recurso no encontrado. El enlace puede haber expirado',
  409: 'Conflicto con datos existentes',
  422: 'Los datos enviados no cumplen con los requisitos',
  429: 'Demasiadas solicitudes. Espere unos minutos antes de intentar nuevamente',
  500: 'Error interno del servidor. Intente nuevamente más tarde',
  502: 'Servidor no disponible temporalmente',
  503: 'Servicio en mantenimiento. Intente más tarde',
  504: 'Tiempo de espera agotado. Verifique su conexión'
};

/**
 * Mensajes específicos para errores de password reset
 */
export const PASSWORD_RESET_ERROR_MESSAGES: Record<PasswordResetErrorType, string> = {
  [PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL]: 'Solo emails @gamc.gov.bo pueden solicitar reset de contraseña',
  [PasswordResetErrorType.RATE_LIMIT_EXCEEDED]: 'Debe esperar 5 minutos entre solicitudes de reset',
  [PasswordResetErrorType.EMAIL_NOT_FOUND]: 'Si el email existe, recibirá instrucciones en breve',
  [PasswordResetErrorType.TOKEN_INVALID]: 'Token de reset inválido. Solicite un nuevo reset',
  [PasswordResetErrorType.TOKEN_EXPIRED]: 'Token expirado. Los tokens son válidos por 30 minutos',
  [PasswordResetErrorType.TOKEN_USED]: 'Este token ya fue utilizado. Solicite un nuevo reset',
  [PasswordResetErrorType.PASSWORD_WEAK]: 'La contraseña no cumple con los requisitos de seguridad',
  [PasswordResetErrorType.NETWORK_ERROR]: 'Error de conexión. Verifique su internet',
  [PasswordResetErrorType.SERVER_ERROR]: 'Error del servidor. Intente nuevamente',
  [PasswordResetErrorType.VALIDATION_ERROR]: 'Error en validación de datos'
};

/**
 * Mensajes específicos para errores de preguntas de seguridad
 */
export const SECURITY_QUESTION_ERROR_MESSAGES: Record<SecurityQuestionErrorType, string> = {
  [SecurityQuestionErrorType.QUESTION_NOT_FOUND]: 'La pregunta de seguridad no existe',
  [SecurityQuestionErrorType.INVALID_ANSWER]: 'Respuesta incorrecta',
  [SecurityQuestionErrorType.MAX_ATTEMPTS_REACHED]: 'Demasiados intentos fallidos. Proceso bloqueado',
  [SecurityQuestionErrorType.QUESTION_ALREADY_EXISTS]: 'Ya tiene configurada esta pregunta',
  [SecurityQuestionErrorType.MAX_QUESTIONS_LIMIT]: 'Límite máximo de preguntas alcanzado (3)',
  [SecurityQuestionErrorType.ANSWER_TOO_SHORT]: 'La respuesta debe tener al menos 2 caracteres',
  [SecurityQuestionErrorType.ANSWER_TOO_LONG]: 'La respuesta no puede exceder 100 caracteres',
  [SecurityQuestionErrorType.NETWORK_ERROR]: 'Error de conexión',
  [SecurityQuestionErrorType.VALIDATION_ERROR]: 'Error en validación'
};

// ========================================
// FUNCIONES DE ANÁLISIS DE ERRORES
// ========================================

/**
 * Analiza un error HTTP y extrae información relevante
 */
export interface ErrorAnalysis {
  type: 'network' | 'http' | 'validation' | 'security' | 'unknown';
  statusCode?: number;
  message: string;
  userMessage: string;
  retryable: boolean;
  severity: 'low' | 'medium' | 'high';
  context?: any;
}

/**
 * Analiza cualquier error y devuelve información estructurada
 */
export const analyzeError = (error: any, operation?: string): ErrorAnalysis => {
  // Error de red (sin respuesta del servidor)
  if (!error.response && (error.code === 'NETWORK_ERROR' || error.name === 'NetworkError')) {
    return {
      type: 'network',
      message: 'Error de conexión de red',
      userMessage: 'No se pudo conectar al servidor. Verifique su conexión a internet.',
      retryable: true,
      severity: 'medium'
    };
  }

  // Error HTTP con respuesta del servidor
  if (error.response) {
    const { status, data } = error.response;
    const baseMessage = HTTP_ERROR_MESSAGES[status] || `Error HTTP ${status}`;
    const serverMessage = data?.message || data?.error || '';

    return {
      type: 'http',
      statusCode: status,
      message: serverMessage || baseMessage,
      userMessage: getHttpErrorUserMessage(status, serverMessage, operation),
      retryable: isRetryableHttpError(status),
      severity: getHttpErrorSeverity(status),
      context: { operation, status, data }
    };
  }

  // Error personalizado de Password Reset
  if (error instanceof PasswordResetError) {
    return {
      type: 'validation',
      message: error.message,
      userMessage: PASSWORD_RESET_ERROR_MESSAGES[error.type] || error.message,
      retryable: isRetryablePasswordResetError(error.type),
      severity: getPasswordResetErrorSeverity(error.type),
      context: { type: error.type, operation }
    };
  }

  // Error personalizado de Security Questions
  if (error instanceof SecurityQuestionError) {
    return {
      type: 'security',
      message: error.message,
      userMessage: SECURITY_QUESTION_ERROR_MESSAGES[error.type] || error.message,
      retryable: isRetryableSecurityQuestionError(error.type),
      severity: getSecurityQuestionErrorSeverity(error.type),
      context: { type: error.type, operation }
    };
  }

  // Error desconocido
  return {
    type: 'unknown',
    message: error.message || 'Error desconocido',
    userMessage: 'Ha ocurrido un error inesperado. Intente nuevamente.',
    retryable: true,
    severity: 'medium',
    context: { operation, error: error.toString() }
  };
};

// ========================================
// FUNCIONES DE CLASIFICACIÓN
// ========================================

/**
 * Genera mensaje user-friendly para errores HTTP
 */
const getHttpErrorUserMessage = (status: number, serverMessage: string, operation?: string): string => {
  const baseMessage = HTTP_ERROR_MESSAGES[status];
  
  // Mensajes específicos según operación
  if (operation === 'forgot-password' && status === 403) {
    return 'Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña';
  }
  
  if (operation === 'verify-security-question' && status === 400) {
    return serverMessage || 'Respuesta incorrecta a la pregunta de seguridad';
  }
  
  if (operation === 'reset-password' && status === 400) {
    if (serverMessage.includes('token')) {
      return 'El enlace de reset es inválido o ha expirado';
    }
    if (serverMessage.includes('contraseña')) {
      return 'La nueva contraseña no cumple con los requisitos de seguridad';
    }
  }

  return serverMessage || baseMessage || `Error ${status}`;
};

/**
 * Determina si un error HTTP es reintentable
 */
const isRetryableHttpError = (status: number): boolean => {
  // Errores de servidor (500+) son reintentables
  if (status >= 500) return true;
  
  // Rate limiting es reintentable después de esperar
  if (status === 429) return true;
  
  // Timeouts son reintentables
  if (status === 408 || status === 504) return true;
  
  // Errores de cliente (4xx) generalmente no son reintentables
  return false;
};

/**
 * Determina la severidad de un error HTTP
 */
const getHttpErrorSeverity = (status: number): 'low' | 'medium' | 'high' => {
  if (status >= 500) return 'high'; // Errores de servidor
  if (status === 429) return 'medium'; // Rate limiting
  if (status === 403) return 'medium'; // Acceso denegado
  if (status === 401) return 'medium'; // No autorizado
  return 'low'; // Otros errores de cliente
};

/**
 * Determina si un error de password reset es reintentable
 */
const isRetryablePasswordResetError = (type: PasswordResetErrorType): boolean => {
  switch (type) {
    case PasswordResetErrorType.NETWORK_ERROR:
    case PasswordResetErrorType.SERVER_ERROR:
      return true;
    case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
      return true; // Después de esperar
    case PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL:
    case PasswordResetErrorType.TOKEN_INVALID:
    case PasswordResetErrorType.TOKEN_EXPIRED:
    case PasswordResetErrorType.TOKEN_USED:
      return false;
    default:
      return true;
  }
};

/**
 * Determina la severidad de un error de password reset
 */
const getPasswordResetErrorSeverity = (type: PasswordResetErrorType): 'low' | 'medium' | 'high' => {
  switch (type) {
    case PasswordResetErrorType.NETWORK_ERROR:
    case PasswordResetErrorType.SERVER_ERROR:
      return 'high';
    case PasswordResetErrorType.TOKEN_EXPIRED:
    case PasswordResetErrorType.TOKEN_USED:
    case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
      return 'medium';
    default:
      return 'low';
  }
};

/**
 * Determina si un error de security questions es reintentable
 */
const isRetryableSecurityQuestionError = (type: SecurityQuestionErrorType): boolean => {
  switch (type) {
    case SecurityQuestionErrorType.NETWORK_ERROR:
      return true;
    case SecurityQuestionErrorType.INVALID_ANSWER:
      return true; // Si quedan intentos
    case SecurityQuestionErrorType.MAX_ATTEMPTS_REACHED:
    case SecurityQuestionErrorType.QUESTION_ALREADY_EXISTS:
    case SecurityQuestionErrorType.MAX_QUESTIONS_LIMIT:
      return false;
    default:
      return true;
  }
};

/**
 * Determina la severidad de un error de security questions
 */
const getSecurityQuestionErrorSeverity = (type: SecurityQuestionErrorType): 'low' | 'medium' | 'high' => {
  switch (type) {
    case SecurityQuestionErrorType.MAX_ATTEMPTS_REACHED:
      return 'high';
    case SecurityQuestionErrorType.INVALID_ANSWER:
    case SecurityQuestionErrorType.NETWORK_ERROR:
      return 'medium';
    default:
      return 'low';
  }
};

// ========================================
// UTILIDADES DE UI PARA ERRORES
// ========================================

/**
 * Convierte un error en una validación de campo
 */
export const errorToFieldValidation = (error: any, defaultMessage?: string): FieldValidation => {
  const analysis = analyzeError(error);
  
  return {
    isValid: false,
    message: analysis.userMessage || defaultMessage || 'Error de validación',
    type: 'error'
  };
};

/**
 * Obtiene el icono apropiado para un tipo de error
 */
export const getErrorIcon = (analysis: ErrorAnalysis): string => {
  switch (analysis.type) {
    case 'network':
      return '🔌';
    case 'http':
      return analysis.statusCode && analysis.statusCode >= 500 ? '⚠️' : '❌';
    case 'validation':
      return '📝';
    case 'security':
      return '🔒';
    default:
      return '❓';
  }
};

/**
 * Obtiene las clases CSS para mostrar un error
 */
export const getErrorClasses = (analysis: ErrorAnalysis): string => {
  const baseClasses = 'p-3 rounded-lg text-sm border';
  
  switch (analysis.severity) {
    case 'high':
      return `${baseClasses} bg-red-50 text-red-800 border-red-200`;
    case 'medium':
      return `${baseClasses} bg-amber-50 text-amber-800 border-amber-200`;
    case 'low':
      return `${baseClasses} bg-blue-50 text-blue-800 border-blue-200`;
    default:
      return `${baseClasses} bg-gray-50 text-gray-800 border-gray-200`;
  }
};

/**
 * Formatea un error para mostrar al usuario con retry si es aplicable
 */
export const formatErrorForDisplay = (error: any, operation?: string): {
  message: string;
  icon: string;
  classes: string;
  retryable: boolean;
  severity: 'low' | 'medium' | 'high';
} => {
  const analysis = analyzeError(error, operation);
  
  return {
    message: analysis.userMessage,
    icon: getErrorIcon(analysis),
    classes: getErrorClasses(analysis),
    retryable: analysis.retryable,
    severity: analysis.severity
  };
};

// ========================================
// LOGGING DE ERRORES PARA DEBUGGING
// ========================================

/**
 * Registra un error con contexto para debugging
 */
export const logError = (error: any, operation?: string, context?: any): void => {
  const analysis = analyzeError(error, operation);
  
  console.group(`🐛 Error en ${operation || 'operación desconocida'}`);
  console.error('Error original:', error);
  console.log('Análisis:', analysis);
  console.log('Contexto adicional:', context);
  console.groupEnd();
  
  // En producción, aquí se enviaría a un servicio de logging
  // como Sentry, LogRocket, etc.
};

/**
 * Crea un handler de errores reutilizable para componentes
 */
export const createErrorHandler = (operation: string, onError?: (analysis: ErrorAnalysis) => void) => {
  return (error: any, context?: any) => {
    logError(error, operation, context);
    const analysis = analyzeError(error, operation);
    
    if (onError) {
      onError(analysis);
    }
    
    return analysis;
  };
};
// src/utils/passwordValidation.ts
// Utilidades de validación de contraseñas para el módulo de reset
// Compatible 100% con las validaciones del backend GAMC

import { 
  FieldValidation, 
  PASSWORD_RESET_CONFIG, 
  PASSWORD_RESET_MESSAGES 
} from '../types/passwordReset';

// ========================================
// VALIDACIONES DE EMAIL
// ========================================

/**
 * Valida email con énfasis en emails institucionales GAMC
 * Compatible con la validación del backend
 */
export const validateEmail = (email: string): FieldValidation => {
  // Email vacío
  if (!email.trim()) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.EMAIL_REQUIRED, 
      type: 'error' 
    };
  }

  // Formato básico de email
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.EMAIL_INVALID_FORMAT, 
      type: 'error' 
    };
  }

  // Email institucional GAMC (requerido para reset)
  if (PASSWORD_RESET_CONFIG.EMAIL_PATTERN.test(email)) {
    return { 
      isValid: true, 
      message: '✓ Email institucional GAMC válido', 
      type: 'success' 
    };
  }

  // Email válido pero no institucional
  return { 
    isValid: false, 
    message: PASSWORD_RESET_MESSAGES.EMAIL_NOT_INSTITUTIONAL, 
    type: 'error' 
  };
};

/**
 * Validación rápida solo para formato (sin restricción institucional)
 * Usada en contextos donde se permite cualquier email válido
 */
export const validateEmailFormat = (email: string): FieldValidation => {
  if (!email.trim()) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.EMAIL_REQUIRED, 
      type: 'error' 
    };
  }

  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.EMAIL_INVALID_FORMAT, 
      type: 'error' 
    };
  }

  if (email.includes('@gamc.gov.bo')) {
    return { 
      isValid: true, 
      message: '✓ Email institucional GAMC', 
      type: 'success' 
    };
  }

  return { 
    isValid: true, 
    message: '✓ Email válido (se recomienda usar @gamc.gov.bo)', 
    type: 'warning' 
  };
};

// ========================================
// VALIDACIONES DE CONTRASEÑA
// ========================================

/**
 * Validación completa de contraseña compatible con backend
 * CRÍTICO: Usar exactamente los mismos caracteres especiales que el backend
 */
export const validatePassword = (password: string): FieldValidation => {
  // Contraseña vacía
  if (!password) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.PASSWORD_REQUIRED, 
      type: 'error' 
    };
  }

  // Longitud mínima
  if (password.length < PASSWORD_RESET_CONFIG.MIN_PASSWORD_LENGTH) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.PASSWORD_TOO_SHORT, 
      type: 'error' 
    };
  }

  // Verificar requisitos específicos
  const hasUppercase = /[A-Z]/.test(password);
  const hasLowercase = /[a-z]/.test(password);
  const hasNumbers = /\d/.test(password);
  // CRÍTICO: Usar exactamente los mismos caracteres especiales que el backend
  const hasSpecialChars = /[@$!%*?&]/.test(password);

  const missing = [];
  if (!hasUppercase) missing.push('mayúscula');
  if (!hasLowercase) missing.push('minúscula');
  if (!hasNumbers) missing.push('número');
  if (!hasSpecialChars) missing.push('símbolo (@$!%*?&)');

  if (missing.length > 0) {
    return { 
      isValid: false, 
      message: `Falta: ${missing.join(', ')}`, 
      type: 'error' 
    };
  }

  // Contraseña válida
  return { 
    isValid: true, 
    message: '✓ Contraseña segura', 
    type: 'success' 
  };
};

/**
 * Validación de confirmación de contraseña
 */
export const validatePasswordConfirmation = (
  password: string, 
  confirmPassword: string
): FieldValidation => {
  if (!confirmPassword) {
    return { 
      isValid: false, 
      message: 'La confirmación de contraseña es requerida', 
      type: 'error' 
    };
  }

  if (password !== confirmPassword) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.PASSWORD_CONFIRM_MISMATCH, 
      type: 'error' 
    };
  }

  return { 
    isValid: true, 
    message: '✓ Las contraseñas coinciden', 
    type: 'success' 
  };
};

/**
 * Análisis detallado de la fortaleza de la contraseña
 * Retorna información detallada para feedback al usuario
 */
export const analyzePasswordStrength = (password: string) => {
  const analysis = {
    length: password.length,
    hasUppercase: /[A-Z]/.test(password),
    hasLowercase: /[a-z]/.test(password),
    hasNumbers: /\d/.test(password),
    hasSpecialChars: /[@$!%*?&]/.test(password),
    score: 0,
    level: 'weak' as 'weak' | 'medium' | 'strong' | 'very-strong',
    feedback: [] as string[]
  };

  // Calcular score
  if (analysis.length >= 8) analysis.score += 1;
  if (analysis.length >= 12) analysis.score += 1;
  if (analysis.hasUppercase) analysis.score += 1;
  if (analysis.hasLowercase) analysis.score += 1;
  if (analysis.hasNumbers) analysis.score += 1;
  if (analysis.hasSpecialChars) analysis.score += 1;

  // Determinar nivel
  if (analysis.score >= 6) analysis.level = 'very-strong';
  else if (analysis.score >= 5) analysis.level = 'strong';
  else if (analysis.score >= 3) analysis.level = 'medium';
  else analysis.level = 'weak';

  // Generar feedback
  if (analysis.length < 8) {
    analysis.feedback.push('Use al menos 8 caracteres');
  }
  if (!analysis.hasUppercase) {
    analysis.feedback.push('Incluya al menos una mayúscula');
  }
  if (!analysis.hasLowercase) {
    analysis.feedback.push('Incluya al menos una minúscula');
  }
  if (!analysis.hasNumbers) {
    analysis.feedback.push('Incluya al menos un número');
  }
  if (!analysis.hasSpecialChars) {
    analysis.feedback.push('Incluya al menos un símbolo (@$!%*?&)');
  }

  return analysis;
};

// ========================================
// VALIDACIONES DE NOMBRES (para registro)
// ========================================

/**
 * Validación de nombres y apellidos
 * Reutilizada del componente RegisterForm existente
 */
export const validateName = (name: string, fieldName: string): FieldValidation => {
  if (!name.trim()) {
    return { 
      isValid: false, 
      message: `${fieldName} es requerido`, 
      type: 'error' 
    };
  }

  if (name.length < 2) {
    return { 
      isValid: false, 
      message: `${fieldName} debe tener al menos 2 caracteres`, 
      type: 'error' 
    };
  }

  if (name.length > 50) {
    return { 
      isValid: false, 
      message: `${fieldName} no puede exceder 50 caracteres`, 
      type: 'error' 
    };
  }

  // Permitir letras, acentos, ñ y espacios
  const nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/;
  if (!nameRegex.test(name)) {
    return { 
      isValid: false, 
      message: 'Solo se permiten letras y espacios', 
      type: 'error' 
    };
  }

  return { 
    isValid: true, 
    message: `✓ ${fieldName} válido`, 
    type: 'success' 
  };
};

// ========================================
// UTILIDADES DE UI
// ========================================

/**
 * Obtiene las clases CSS para inputs según el estado de validación
 * Compatible con el sistema de estilos existente
 */
export const getInputClasses = (
  validation?: FieldValidation, 
  baseClasses?: string
): string => {
  const base = baseClasses || 'w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 transition-colors';

  if (!validation || validation.type === '') {
    return `${base} border-gray-300 focus:ring-blue-500`;
  }

  switch (validation.type) {
    case 'success':
      return `${base} border-green-300 focus:ring-green-500 bg-green-50`;
    case 'error':
      return `${base} border-red-300 focus:ring-red-500 bg-red-50`;
    case 'warning':
      return `${base} border-amber-300 focus:ring-amber-500 bg-amber-50`;
    default:
      return `${base} border-gray-300 focus:ring-blue-500`;
  }
};

/**
 * Obtiene las clases CSS para mensajes de validación
 */
export const getValidationMessageClasses = (validation: FieldValidation): string => {
  const colors = {
    success: 'text-green-600 bg-green-50 border-green-200',
    error: 'text-red-600 bg-red-50 border-red-200',
    warning: 'text-amber-600 bg-amber-50 border-amber-200'
  };

  return `text-xs mt-1 p-2 rounded border ${colors[validation.type] || 'text-gray-600 bg-gray-50 border-gray-200'}`;
};

/**
 * Obtiene el color para el indicador de fortaleza de contraseña
 */
export const getPasswordStrengthColor = (level: string): string => {
  switch (level) {
    case 'very-strong': return 'text-green-600';
    case 'strong': return 'text-green-500';
    case 'medium': return 'text-yellow-500';
    case 'weak': return 'text-red-500';
    default: return 'text-gray-400';
  }
};

/**
 * Obtiene el texto descriptivo para el nivel de fortaleza
 */
export const getPasswordStrengthText = (level: string): string => {
  switch (level) {
    case 'very-strong': return 'Muy segura';
    case 'strong': return 'Segura';
    case 'medium': return 'Moderada';
    case 'weak': return 'Débil';
    default: return 'Sin evaluar';
  }
};

// ========================================
// VALIDACIONES COMPUESTAS
// ========================================

/**
 * Validación completa para formulario de solicitud de reset
 */
export const validateForgotPasswordForm = (email: string) => {
  const emailValidation = validateEmail(email);
  
  return {
    email: emailValidation,
    isValid: emailValidation.isValid
  };
};

/**
 * Validación completa para formulario de confirmación de reset
 */
export const validateResetPasswordForm = (
  token: string, 
  newPassword: string, 
  confirmPassword: string
) => {
  const tokenValidation = validateToken(token);
  const passwordValidation = validatePassword(newPassword);
  const confirmValidation = validatePasswordConfirmation(newPassword, confirmPassword);
  
  return {
    token: tokenValidation,
    password: passwordValidation,
    confirmPassword: confirmValidation,
    isValid: tokenValidation.isValid && passwordValidation.isValid && confirmValidation.isValid
  };
};

// ========================================
// VALIDACIÓN DE TOKEN (adelanto)
// ========================================

/**
 * Validación básica de token (la completa estará en tokenValidation.ts)
 */
export const validateToken = (token: string): FieldValidation => {
  if (!token.trim()) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_REQUIRED, 
      type: 'error' 
    };
  }

  if (token.length !== PASSWORD_RESET_CONFIG.TOKEN_LENGTH) {
    return { 
      isValid: false, 
      message: PASSWORD_RESET_MESSAGES.TOKEN_INVALID_FORMAT, 
      type: 'error' 
    };
  }

  if (!PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(token)) {
    return { 
      isValid: false, 
      message: 'Token contiene caracteres inválidos', 
      type: 'error' 
    };
  }

  return { 
    isValid: true, 
    message: '✓ Token válido', 
    type: 'success' 
  };
};

// ========================================
// UTILIDADES DE RATE LIMITING
// ========================================

/**
 * Calcula el tiempo restante para poder hacer otra solicitud
 * Basado en localStorage para persistir entre recargas
 */
export const getTimeUntilNextRequest = (): number => {
  const lastRequestTime = localStorage.getItem('lastPasswordResetRequest');
  if (!lastRequestTime) return 0;

  const timeDiff = Date.now() - parseInt(lastRequestTime);
  const timeRemaining = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW - timeDiff;
  
  return Math.max(0, Math.ceil(timeRemaining / 1000)); // segundos
};

/**
 * Registra el timestamp de la última solicitud de reset
 */
export const recordPasswordResetRequest = (): void => {
  localStorage.setItem('lastPasswordResetRequest', Date.now().toString());
};

/**
 * Verifica si se puede hacer una nueva solicitud de reset
 */
export const canMakePasswordResetRequest = (): boolean => {
  return getTimeUntilNextRequest() === 0;
};

/**
 * Formatea el tiempo restante en formato legible
 */
export const formatTimeRemaining = (seconds: number): string => {
  if (seconds <= 0) return '';
  
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;
  
  if (minutes > 0) {
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  }
  
  return `${remainingSeconds}s`;
};

// ========================================
// EXPORT POR DEFECTO
// ========================================

export default {
  // Validaciones principales
  validateEmail,
  validateEmailFormat,
  validatePassword,
  validatePasswordConfirmation,
  validateName,
  validateToken,
  
  // Análisis
  analyzePasswordStrength,
  
  // Utilidades UI
  getInputClasses,
  getValidationMessageClasses,
  getPasswordStrengthColor,
  getPasswordStrengthText,
  
  // Validaciones compuestas
  validateForgotPasswordForm,
  validateResetPasswordForm,
  
  // Rate limiting
  getTimeUntilNextRequest,
  recordPasswordResetRequest,
  canMakePasswordResetRequest,
  formatTimeRemaining
};
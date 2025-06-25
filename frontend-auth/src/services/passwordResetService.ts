// src/services/passwordResetService.ts
// Servicio para todas las operaciones de reset de contraseña
// Integrado con preguntas de seguridad y arquitectura existente

import { apiClient } from './api';
import { 
  PasswordResetRequest,
  PasswordResetConfirm,
  PasswordResetRequestResponse,
  PasswordResetConfirmResponse,
  PasswordResetStatusResponse,
  PasswordResetCleanupResponse,
  PasswordResetToken,
  PasswordResetErrorType,
  PASSWORD_RESET_CONFIG 
} from '../types/passwordReset';
import { ApiResponse } from '../types/auth';

// ========================================
// NUEVOS TIPOS PARA PREGUNTAS DE SEGURIDAD
// ========================================

export interface SecurityQuestion {
  id: number;
  questionText: string;
  category: string;
}

export interface SecurityQuestionForReset {
  questionId: number;
  questionText: string;
  attempts: number;
  maxAttempts: number;
}

export interface PasswordResetInitResponse {
  success: boolean;
  message: string;
  requiresSecurityQuestion: boolean;
  securityQuestion?: SecurityQuestionForReset;
}

export interface PasswordResetVerifySecurityRequest {
  email: string;
  questionId: number;
  answer: string;
}

export interface PasswordResetVerifySecurityResponse {
  success: boolean;
  message: string;
  verified: boolean;
  canProceedToReset: boolean;
  attemptsRemaining: number;
  resetToken?: string; // Solo se devuelve si la verificación es exitosa
}

export interface PasswordResetStatusByEmailResponse {
  tokenValid: boolean;
  tokenExpired: boolean;
  tokenUsed: boolean;
  requiresSecurityQuestion: boolean;
  securityQuestionVerified: boolean;
  canProceedToReset: boolean;
  attemptsRemaining: number;
  securityQuestion?: SecurityQuestionForReset;
}

/**
 * Clase para manejar todas las operaciones de reset de contraseña
 * Incluye nuevo flujo con preguntas de seguridad
 */
class PasswordResetService {

  // ========================================
  // ENDPOINTS PÚBLICOS DE PREGUNTAS DE SEGURIDAD
  // ========================================

  /**
   * Obtener catálogo de preguntas de seguridad disponibles
   * GET /api/v1/auth/security-questions
   */
  async getSecurityQuestions(): Promise<SecurityQuestion[]> {
    try {
      const response = await apiClient.get<{ questions: SecurityQuestion[]; count: number }>(
        '/auth/security-questions',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response.questions;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'questions');
    }
  }

  /**
   * Solicitar reset de contraseña (NUEVO FLUJO)
   * POST /api/v1/auth/forgot-password
   * Ahora retorna información sobre preguntas de seguridad si aplica
   */
  async requestPasswordReset(request: PasswordResetRequest): Promise<PasswordResetInitResponse> {
    try {
      const response = await apiClient.post<PasswordResetInitResponse>(
        '/auth/forgot-password',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      // Registrar el timestamp de la solicitud para rate limiting
      this.recordResetRequest();

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'request');
    }
  }

  /**
   * Verificar pregunta de seguridad durante reset (NUEVO)
   * POST /api/v1/auth/verify-security-question
   * Retorna el token de reset si la verificación es exitosa
   */
  async verifySecurityQuestion(request: PasswordResetVerifySecurityRequest): Promise<PasswordResetVerifySecurityResponse> {
    try {
      const response = await apiClient.post<PasswordResetVerifySecurityResponse>(
        '/auth/verify-security-question',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'verify');
    }
  }

  /**
   * Obtener estado del proceso de reset por email (NUEVO)
   * GET /api/v1/auth/reset-status?email=user@gamc.gov.bo
   */
  async getPasswordResetStatusByEmail(email: string): Promise<PasswordResetStatusByEmailResponse> {
    try {
      const response = await apiClient.get<PasswordResetStatusByEmailResponse>(
        `/auth/reset-status?email=${encodeURIComponent(email)}`,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'status');
    }
  }

  /**
   * Confirmar reset de contraseña (ACTUALIZADO)
   * POST /api/v1/auth/reset-password
   * Ahora requiere token obtenido tras verificar pregunta de seguridad
   */
  async confirmPasswordReset(request: PasswordResetConfirm): Promise<PasswordResetConfirmResponse> {
    try {
      const response = await apiClient.post<PasswordResetConfirmResponse>(
        '/auth/reset-password',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      // Limpiar rate limiting tras reset exitoso
      this.clearRateLimitRecord();

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'confirm');
    }
  }

  // ========================================
  // ENDPOINTS PROTEGIDOS (REQUIEREN AUTH)
  // ========================================

  /**
   * Obtener historial de resets del usuario actual (NUEVO)
   * GET /api/v1/auth/reset-history
   * Requiere autenticación
   */
  async getPasswordResetHistory(): Promise<PasswordResetToken[]> {
    try {
      const response = await apiClient.get<{ tokens: PasswordResetToken[]; count: number }>(
        '/auth/reset-history',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response.tokens;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'history');
    }
  }

  // ========================================
  // ENDPOINTS DE ADMINISTRACIÓN (SOLO ADMINS)
  // ========================================

  /**
   * Limpiar tokens expirados (solo administradores)
   * POST /api/v1/auth/admin/cleanup-tokens
   */
  async cleanupExpiredTokens(): Promise<PasswordResetCleanupResponse> {
    try {
      const response = await apiClient.post<PasswordResetCleanupResponse>(
        '/auth/admin/cleanup-tokens',
        {},
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'cleanup');
    }
  }

  // ========================================
  // UTILIDADES Y VALIDACIONES
  // ========================================

  /**
   * Validar email antes de enviar solicitud
   */
  validateEmailForReset(email: string): { isValid: boolean; error?: string } {
    if (!email?.trim()) {
      return { isValid: false, error: 'Email es requerido' };
    }

    if (!PASSWORD_RESET_CONFIG.EMAIL_PATTERN.test(email.trim())) {
      return { 
        isValid: false, 
        error: 'Solo emails @gamc.gov.bo pueden solicitar reset de contraseña' 
      };
    }

    return { isValid: true };
  }

  /**
   * Validar respuesta de seguridad (NUEVO)
   */
  validateSecurityAnswer(answer: string): { isValid: boolean; error?: string } {
    if (!answer?.trim()) {
      return { isValid: false, error: 'La respuesta es requerida' };
    }

    const trimmedAnswer = answer.trim();

    if (trimmedAnswer.length < 2) {
      return { 
        isValid: false, 
        error: 'La respuesta debe tener al menos 2 caracteres' 
      };
    }

    if (trimmedAnswer.length > 100) {
      return { 
        isValid: false, 
        error: 'La respuesta no puede exceder 100 caracteres' 
      };
    }

    // Verificar que no sea solo caracteres repetidos
    if (/^(.)\1+$/.test(trimmedAnswer)) {
      return { 
        isValid: false, 
        error: 'La respuesta no puede ser solo caracteres repetidos' 
      };
    }

    // Verificar que no sea solo números
    if (/^\d+$/.test(trimmedAnswer)) {
      return { 
        isValid: false, 
        error: 'La respuesta no puede ser solo números' 
      };
    }

    return { isValid: true };
  }

  /**
   * Validar token antes de enviar confirmación
   */
  validateTokenForReset(token: string): { isValid: boolean; error?: string } {
    if (!token?.trim()) {
      return { isValid: false, error: 'Token es requerido' };
    }

    const cleanToken = token.trim();

    if (cleanToken.length !== PASSWORD_RESET_CONFIG.TOKEN_LENGTH) {
      return { 
        isValid: false, 
        error: `Token debe tener exactamente ${PASSWORD_RESET_CONFIG.TOKEN_LENGTH} caracteres` 
      };
    }

    if (!PASSWORD_RESET_CONFIG.TOKEN_PATTERN.test(cleanToken)) {
      return { 
        isValid: false, 
        error: 'Token contiene caracteres inválidos' 
      };
    }

    return { isValid: true };
  }

  /**
   * Validar nueva contraseña antes de enviar confirmación
   */
  validatePasswordForReset(password: string): { isValid: boolean; error?: string } {
    if (!password) {
      return { isValid: false, error: 'Nueva contraseña es requerida' };
    }

    if (password.length < PASSWORD_RESET_CONFIG.MIN_PASSWORD_LENGTH) {
      return { 
        isValid: false, 
        error: `Contraseña debe tener al menos ${PASSWORD_RESET_CONFIG.MIN_PASSWORD_LENGTH} caracteres` 
      };
    }

    if (!PASSWORD_RESET_CONFIG.PASSWORD_PATTERN.test(password)) {
      return { 
        isValid: false, 
        error: 'Contraseña debe contener mayúscula, minúscula, número y símbolo (@$!%*?&)' 
      };
    }

    return { isValid: true };
  }

  // ========================================
  // FLUJO COMPLETO DE RESET CON PREGUNTAS
  // ========================================

  /**
   * Ejecutar flujo completo de reset de contraseña
   * Maneja automáticamente preguntas de seguridad si son requeridas
   */
  async executeFullPasswordReset(
    email: string, 
    newPassword: string,
    onSecurityQuestionRequired?: (question: SecurityQuestionForReset) => Promise<string>
  ): Promise<PasswordResetConfirmResponse> {
    
    // Paso 1: Validar datos iniciales
    const emailValidation = this.validateEmailForReset(email);
    if (!emailValidation.isValid) {
      throw new PasswordResetError(
        PasswordResetErrorType.VALIDATION_ERROR,
        emailValidation.error!
      );
    }

    const passwordValidation = this.validatePasswordForReset(newPassword);
    if (!passwordValidation.isValid) {
      throw new PasswordResetError(
        PasswordResetErrorType.PASSWORD_WEAK,
        passwordValidation.error!
      );
    }

    // Paso 2: Solicitar reset
    const resetRequest = await this.requestPasswordReset({ email });
    
    let resetToken: string;

    // Paso 3: Manejar pregunta de seguridad si es requerida
    if (resetRequest.requiresSecurityQuestion && resetRequest.securityQuestion) {
      if (!onSecurityQuestionRequired) {
        throw new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          'Se requiere verificación de pregunta de seguridad pero no se proporcionó manejador'
        );
      }

      // Obtener respuesta del usuario
      const userAnswer = await onSecurityQuestionRequired(resetRequest.securityQuestion);
      
      // Validar respuesta
      const answerValidation = this.validateSecurityAnswer(userAnswer);
      if (!answerValidation.isValid) {
        throw new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          answerValidation.error!
        );
      }

      // Verificar pregunta de seguridad
      const verifyResponse = await this.verifySecurityQuestion({
        email,
        questionId: resetRequest.securityQuestion.questionId,
        answer: userAnswer
      });

      if (!verifyResponse.verified || !verifyResponse.resetToken) {
        throw new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          verifyResponse.message
        );
      }

      resetToken = verifyResponse.resetToken;
    } else {
      // Para usuarios sin preguntas de seguridad, el token debería venir por email
      // En este caso, necesitamos que el usuario proporcione el token manualmente
      throw new PasswordResetError(
        PasswordResetErrorType.VALIDATION_ERROR,
        'Token de reset enviado por email. Revise su correo institucional'
      );
    }

    // Paso 4: Confirmar reset con token obtenido
    return await this.confirmPasswordReset({
      token: resetToken,
      newPassword
    });
  }

  // ========================================
  // MANEJO DE ERRORES ESPECÍFICOS
  // ========================================

  /**
   * Maneja errores específicos de reset de contraseña
   */
  private handlePasswordResetError(error: any, operation: string): Error {
    // Error de red/timeout
    if (!error.response) {
      return new PasswordResetError(
        PasswordResetErrorType.NETWORK_ERROR,
        'Error de conexión. Verifique su conexión a internet y que el servidor esté ejecutándose.',
        error
      );
    }

    const { status, data } = error.response;
    const errorMessage = data?.message || data?.error || 'Error desconocido';

    // Mapear errores HTTP específicos según la operación
    switch (operation) {
      case 'request':
        return this.handleRequestErrors(status, errorMessage, data);
      
      case 'verify':
        return this.handleVerifyErrors(status, errorMessage, data);
      
      case 'confirm':
        return this.handleConfirmErrors(status, errorMessage, data);
      
      case 'status':
        return this.handleStatusErrors(status, errorMessage, data);
      
      case 'history':
        return this.handleHistoryErrors(status, errorMessage, data);
      
      case 'cleanup':
        return this.handleCleanupErrors(status, errorMessage, data);
      
      case 'questions':
        return this.handleQuestionsErrors(status, errorMessage, data);
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          errorMessage,
          error
        );
    }
  }

  /**
   * Maneja errores específicos de verificación de pregunta de seguridad (NUEVO)
   */
  private handleVerifyErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 400:
        if (message.includes('incorrecta') || message.includes('incorrect')) {
          return new PasswordResetError(
            PasswordResetErrorType.VALIDATION_ERROR,
            'Respuesta de seguridad incorrecta',
            { status, message, data }
          );
        }
        
        if (message.includes('no encontrado') || message.includes('not found')) {
          return new PasswordResetError(
            PasswordResetErrorType.VALIDATION_ERROR,
            'No hay solicitud de reset activa para este usuario',
            { status, message, data }
          );
        }
        
        if (message.includes('expirado') || message.includes('expired')) {
          return new PasswordResetError(
            PasswordResetErrorType.TOKEN_EXPIRED,
            'La solicitud de reset ha expirado. Solicite un nuevo reset',
            { status, message, data }
          );
        }
        
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          `Error de verificación: ${message}`,
          { status, message, data }
        );
      
      case 429:
        return new PasswordResetError(
          PasswordResetErrorType.RATE_LIMIT_EXCEEDED,
          'Demasiados intentos fallidos. Solicite un nuevo reset',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          `Error ${status}: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores de obtener preguntas de seguridad (NUEVO)
   */
  private handleQuestionsErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 500:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          'Error del servidor al cargar preguntas de seguridad',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.NETWORK_ERROR,
          `Error al cargar preguntas: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores de historial (NUEVO)
   */
  private handleHistoryErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 401:
        return new PasswordResetError(
          PasswordResetErrorType.NETWORK_ERROR,
          'Sesión expirada. Inicie sesión nuevamente',
          { status, message, data }
        );
      
      case 403:
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          'No tiene permisos para acceder al historial',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          `Error al obtener historial: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores específicos de solicitud de reset
   */
  private handleRequestErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 403:
        return new PasswordResetError(
          PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL,
          'Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña',
          { status, message, data }
        );
      
      case 429:
        return new PasswordResetError(
          PasswordResetErrorType.RATE_LIMIT_EXCEEDED,
          'Debe esperar 5 minutos entre solicitudes de reset',
          { status, message, data }
        );
      
      case 400:
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          `Datos inválidos: ${message}`,
          { status, message, data }
        );
      
      case 500:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          'Error del servidor. Intente nuevamente en unos momentos',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          `Error ${status}: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores específicos de confirmación de reset
   */
  private handleConfirmErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 400:
        if (message.includes('inválido') || message.includes('invalid')) {
          return new PasswordResetError(
            PasswordResetErrorType.TOKEN_INVALID,
            'Token de reset inválido',
            { status, message, data }
          );
        }
        
        if (message.includes('expirado') || message.includes('expired')) {
          return new PasswordResetError(
            PasswordResetErrorType.TOKEN_EXPIRED,
            'El token de reset ha expirado. Solicite un nuevo reset',
            { status, message, data }
          );
        }
        
        if (message.includes('usado') || message.includes('used')) {
          return new PasswordResetError(
            PasswordResetErrorType.TOKEN_USED,
            'Este token ya fue utilizado. Solicite un nuevo reset si es necesario',
            { status, message, data }
          );
        }
        
        if (message.includes('contraseña') || message.includes('password')) {
          return new PasswordResetError(
            PasswordResetErrorType.PASSWORD_WEAK,
            `Nueva contraseña inválida: ${message}`,
            { status, message, data }
          );
        }
        
        if (message.includes('verificar') || message.includes('pregunta')) {
          return new PasswordResetError(
            PasswordResetErrorType.VALIDATION_ERROR,
            'Debe verificar la pregunta de seguridad primero',
            { status, message, data }
          );
        }
        
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          `Error de validación: ${message}`,
          { status, message, data }
        );
      
      case 500:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          'Error del servidor. Intente nuevamente en unos momentos',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          `Error ${status}: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores de consulta de estado
   */
  private handleStatusErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 400:
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          'Email requerido para consultar estado',
          { status, message, data }
        );
      
      case 500:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          `Error al obtener estado: ${message}`,
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          `Error ${status}: ${message}`,
          { status, message, data }
        );
    }
  }

  /**
   * Maneja errores de limpieza de tokens (admin)
   */
  private handleCleanupErrors(status: number, message: string, data: any): Error {
    switch (status) {
      case 401:
        return new PasswordResetError(
          PasswordResetErrorType.NETWORK_ERROR,
          'Sesión expirada. Inicie sesión nuevamente',
          { status, message, data }
        );
      
      case 403:
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          'Solo administradores pueden realizar limpieza de tokens',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          `Error en limpieza: ${message}`,
          { status, message, data }
        );
    }
  }

  // ========================================
  // UTILIDADES DE ESTADO
  // ========================================

  /**
   * Verifica si un usuario puede solicitar reset (rate limiting)
   */
  canRequestReset(): { canRequest: boolean; timeRemaining?: number } {
    const lastRequestTime = localStorage.getItem('lastPasswordResetRequest');
    
    if (!lastRequestTime) {
      return { canRequest: true };
    }

    const timeDiff = Date.now() - parseInt(lastRequestTime);
    const timeRemaining = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW - timeDiff;

    if (timeRemaining <= 0) {
      return { canRequest: true };
    }

    return { 
      canRequest: false, 
      timeRemaining: Math.ceil(timeRemaining / 1000) 
    };
  }

  /**
   * Registra el timestamp de la última solicitud de reset
   */
  recordResetRequest(): void {
    localStorage.setItem('lastPasswordResetRequest', Date.now().toString());
  }

  /**
   * Limpia el registro de rate limiting
   */
  clearRateLimitRecord(): void {
    localStorage.removeItem('lastPasswordResetRequest');
  }

  /**
   * Guarda estado del proceso de reset en progreso (NUEVO)
   */
  saveResetProgress(email: string, step: 'requested' | 'security_verified', data?: any): void {
    const progressData = {
      email,
      step,
      timestamp: Date.now(),
      data
    };
    localStorage.setItem('passwordResetProgress', JSON.stringify(progressData));
  }

  /**
   * Obtiene estado del proceso de reset en progreso (NUEVO)
   */
  getResetProgress(): { email: string; step: string; timestamp: number; data?: any } | null {
    const progressData = localStorage.getItem('passwordResetProgress');
    if (!progressData) return null;

    try {
      const parsed = JSON.parse(progressData);
      // Verificar que no haya expirado (30 minutos)
      if (Date.now() - parsed.timestamp > 30 * 60 * 1000) {
        this.clearResetProgress();
        return null;
      }
      return parsed;
    } catch {
      this.clearResetProgress();
      return null;
    }
  }

  /**
   * Limpia estado del proceso de reset (NUEVO)
   */
  clearResetProgress(): void {
    localStorage.removeItem('passwordResetProgress');
  }
}

// ========================================
// CLASE DE ERROR PERSONALIZADA
// ========================================

/**
 * Error personalizado para operaciones de reset de contraseña
 */
export class PasswordResetError extends Error {
  public type: PasswordResetErrorType;
  public context?: any;

  constructor(type: PasswordResetErrorType, message: string, context?: any) {
    super(message);
    this.name = 'PasswordResetError';
    this.type = type;
    this.context = context;
  }

  /**
   * Verifica si el error es de un tipo específico
   */
  is(type: PasswordResetErrorType): boolean {
    return this.type === type;
  }

  /**
   * Obtiene mensaje de usuario amigable
   */
  getUserMessage(): string {
    switch (this.type) {
      case PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL:
        return 'Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña.';
      
      case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
        return 'Debe esperar 5 minutos entre solicitudes de reset.';
      
      case PasswordResetErrorType.TOKEN_EXPIRED:
        return 'El enlace de reset ha expirado. Solicite un nuevo reset.';
      
      case PasswordResetErrorType.TOKEN_USED:
        return 'Este enlace ya fue utilizado. Solicite un nuevo reset si es necesario.';
      
      case PasswordResetErrorType.TOKEN_INVALID:
        return 'Enlace de reset inválido. Verifique que copió correctamente la URL.';
      
      case PasswordResetErrorType.PASSWORD_WEAK:
        return 'La nueva contraseña no cumple con los requisitos de seguridad.';
      
      case PasswordResetErrorType.NETWORK_ERROR:
        return 'Error de conexión. Verifique su internet e intente nuevamente.';
      
      case PasswordResetErrorType.SERVER_ERROR:
        return 'Error del servidor. Intente nuevamente en unos momentos.';
      
      default:
        return this.message || 'Error inesperado. Contacte al soporte técnico.';
    }
  }
}

// ========================================
// INSTANCIA SINGLETON
// ========================================

export const passwordResetService = new PasswordResetService();

// Export por defecto
export default passwordResetService;
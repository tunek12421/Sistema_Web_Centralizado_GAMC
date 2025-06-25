// src/services/passwordResetService.ts
// Servicio para todas las operaciones de reset de contraseña
// Integrado con preguntas de seguridad y arquitectura existente

import { apiClient } from './api';
import type { ApiResponse } from '../types/index';
import type { 
  PasswordResetRequest,
  PasswordResetConfirm,
  PasswordResetConfirmResponse,
  PasswordResetCleanupResponse,
  PasswordResetToken,
  PasswordResetErrorType
} from '../types/passwordReset';
import { 
  PasswordResetError,
  PASSWORD_RESET_CONFIG 
} from '../types/passwordReset';

// ========================================
// INTERFACES ESPECÍFICAS PARA PREGUNTAS DE SEGURIDAD
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

// Exportar la clase de error para uso externo
export { PasswordResetError };
export type { PasswordResetErrorType };

// ========================================
// SERVICIO PRINCIPAL
// ========================================

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
      const response = await apiClient.get<ApiResponse<{ questions: SecurityQuestion[]; count: number }>>(
        '/auth/security-questions',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response.data!.questions;
    } catch (error: any) {
      throw this.handlePasswordResetError(error);
    }
  }

  /**
   * Solicitar reset de contraseña (NUEVO FLUJO)
   * POST /api/v1/auth/forgot-password
   * Ahora retorna información sobre preguntas de seguridad si aplica
   */
  async requestPasswordReset(request: PasswordResetRequest): Promise<PasswordResetInitResponse> {
    try {
      const response = await apiClient.post<ApiResponse<PasswordResetInitResponse>>(
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

      return response.data!;
    } catch (error: any) {
      throw this.handlePasswordResetError(error);
    }
  }

  /**
   * Verificar pregunta de seguridad durante reset (NUEVO)
   * POST /api/v1/auth/verify-security-question
   */
  async verifySecurityQuestion(request: PasswordResetVerifySecurityRequest): Promise<PasswordResetVerifySecurityResponse> {
    try {
      const response = await apiClient.post<ApiResponse<PasswordResetVerifySecurityResponse>>(
        '/auth/verify-security-question',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response.data!;
    } catch (error: any) {
      throw this.handlePasswordResetError(error);
    }
  }

  /**
   * Obtener estado de reset por email (NUEVO)
   * GET /api/v1/auth/reset-status?email=
   */
  async getPasswordResetStatusByEmail(email: string): Promise<PasswordResetStatusByEmailResponse> {
    try {
      const response = await apiClient.get<ApiResponse<PasswordResetStatusByEmailResponse>>(
        `/auth/reset-status?email=${encodeURIComponent(email)}`,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response.data!;
    } catch (error: any) {
      throw this.handlePasswordResetError(error);
    }
  }

  /**
   * Confirmar reset de contraseña
   * POST /api/v1/auth/reset-password
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
      throw this.handlePasswordResetError(error);
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
      const response = await apiClient.get<ApiResponse<{ tokens: PasswordResetToken[]; count: number }>>(
        '/auth/reset-history',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response.data!.tokens;
    } catch (error: any) {
      throw this.handlePasswordResetError(error);
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
      throw this.handlePasswordResetError(error);
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

    return { isValid: true };
  }

  /**
   * Validar nueva contraseña
   */
  validatePasswordForReset(password: string): { isValid: boolean; error?: string } {
    if (!password) {
      return { isValid: false, error: 'La contraseña es requerida' };
    }

    if (password.length < PASSWORD_RESET_CONFIG.MIN_PASSWORD_LENGTH) {
      return { 
        isValid: false, 
        error: `La contraseña debe tener al menos ${PASSWORD_RESET_CONFIG.MIN_PASSWORD_LENGTH} caracteres` 
      };
    }

    if (!PASSWORD_RESET_CONFIG.PASSWORD_PATTERN.test(password)) {
      return { 
        isValid: false, 
        error: 'La contraseña debe incluir mayúsculas, minúsculas, números y símbolos (@$!%*?&)' 
      };
    }

    return { isValid: true };
  }

  /**
   * Flujo completo de reset con manejo de preguntas de seguridad
   */
  async performCompletePasswordReset(
    email: string,
    newPassword: string,
    onSecurityQuestionRequired?: (question: SecurityQuestionForReset) => Promise<string>
  ): Promise<PasswordResetConfirmResponse> {
    
    // Paso 1: Validar datos iniciales
    const emailValidation = this.validateEmailForReset(email);
    if (!emailValidation.isValid) {
      throw new PasswordResetError(
        'validation_error',
        emailValidation.error!
      );
    }

    const passwordValidation = this.validatePasswordForReset(newPassword);
    if (!passwordValidation.isValid) {
      throw new PasswordResetError(
        'password_weak',
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
          'validation_error',
          'Se requiere verificación de pregunta de seguridad pero no se proporcionó manejador'
        );
      }

      // Obtener respuesta del usuario
      const userAnswer = await onSecurityQuestionRequired(resetRequest.securityQuestion);
      
      // Validar respuesta
      const answerValidation = this.validateSecurityAnswer(userAnswer);
      if (!answerValidation.isValid) {
        throw new PasswordResetError(
          'validation_error',
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
          'validation_error',
          verifyResponse.message
        );
      }

      resetToken = verifyResponse.resetToken;
    } else {
      // Para usuarios sin preguntas de seguridad, el token debería venir por email
      // En este caso, necesitamos que el usuario proporcione el token manualmente
      throw new PasswordResetError(
        'validation_error',
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
  private handlePasswordResetError(error: any): Error {
    // Error de red/timeout
    if (!error.response) {
      return new PasswordResetError(
        'network_error',
        'Error de conexión. Verifique su internet e intente nuevamente.'
      );
    }

    const { status, data } = error.response;

    // Mapear errores HTTP específicos
    switch (status) {
      case 400:
        if (data?.error?.includes('email no institucional') || data?.error?.includes('gamc.gov.bo')) {
          return new PasswordResetError(
            'email_not_institutional',
            'Solo emails @gamc.gov.bo pueden solicitar reset de contraseña'
          );
        }
        if (data?.error?.includes('token inválido') || data?.error?.includes('token expirado')) {
          return new PasswordResetError(
            'token_invalid',
            'Token de reset inválido o expirado'
          );
        }
        if (data?.error?.includes('contraseña débil') || data?.error?.includes('requisitos')) {
          return new PasswordResetError(
            'password_weak',
            'La contraseña no cumple con los requisitos de seguridad'
          );
        }
        break;

      case 429:
        return new PasswordResetError(
          'rate_limit_exceeded',
          'Demasiadas solicitudes. Espere 5 minutos antes de intentar nuevamente.'
        );

      case 404:
        return new PasswordResetError(
          'email_not_found',
          'Si el email existe, recibirá instrucciones para restablecer su contraseña'
        );

      case 500:
      default:
        return new PasswordResetError(
          'server_error',
          'Error interno del servidor. Intente nuevamente más tarde.'
        );
    }

    return new PasswordResetError(
      'validation_error',
      data?.message || data?.error || 'Error desconocido'
    );
  }

  // ========================================
  // UTILIDADES DE RATE LIMITING
  // ========================================

  /**
   * Registra una solicitud de reset para rate limiting
   */
  private recordResetRequest(): void {
    try {
      localStorage.setItem('gamc_last_reset_request', Date.now().toString());
    } catch (error) {
      console.warn('No se pudo registrar rate limiting:', error);
    }
  }

  /**
   * Limpia el registro de rate limiting
   */
  private clearRateLimitRecord(): void {
    try {
      localStorage.removeItem('gamc_last_reset_request');
    } catch (error) {
      console.warn('No se pudo limpiar rate limiting:', error);
    }
  }

  /**
   * Verifica si se puede hacer una nueva solicitud
   */
  canMakeNewRequest(): boolean {
    try {
      const lastRequest = localStorage.getItem('gamc_last_reset_request');
      if (!lastRequest) return true;

      const lastRequestTime = parseInt(lastRequest);
      const now = Date.now();
      const timeElapsed = now - lastRequestTime;
      
      return timeElapsed >= PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW;
    } catch {
      return true;
    }
  }

  /**
   * Obtiene tiempo restante hasta poder hacer nueva solicitud
   */
  getTimeUntilNextRequest(): number {
    try {
      const lastRequest = localStorage.getItem('gamc_last_reset_request');
      if (!lastRequest) return 0;

      const lastRequestTime = parseInt(lastRequest);
      const now = Date.now();
      const timeElapsed = now - lastRequestTime;
      const remaining = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW - timeElapsed;
      
      return Math.max(0, Math.ceil(remaining / 1000)); // Retornar en segundos
    } catch {
      return 0;
    }
  }
}

// Exportar instancia singleton
export const passwordResetService = new PasswordResetService();
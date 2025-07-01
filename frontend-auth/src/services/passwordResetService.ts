// src/services/passwordResetService.ts
// CORRECCIÓN: Servicio sin credentials para operaciones públicas de reset

import axios from 'axios';
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
  resetToken?: string;
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

  // Cliente HTTP específico para operaciones públicas (sin credentials)
  private publicClient = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3000/api/v1',
    withCredentials: false, // 🔧 CORRECCIÓN: Sin credentials para operaciones públicas
    timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Cliente HTTP para operaciones que requieren auth (importado desde api.ts)
  private async getAuthenticatedClient() {
    const { apiClient } = await import('./api');
    return apiClient;
  }

  // ========================================
  // ENDPOINTS PÚBLICOS DE PREGUNTAS DE SEGURIDAD
  // ========================================

  /**
   * Obtener catálogo de preguntas de seguridad disponibles
   * GET /api/v1/auth/security-questions
   */
  async getSecurityQuestions(): Promise<SecurityQuestion[]> {
    try {
      const response = await this.publicClient.get<ApiResponse<{ questions: SecurityQuestion[]; count: number }>>(
        '/auth/security-questions'
      );

      return response.data.data!.questions;
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
      console.log('🔧 Using public client for forgot-password request');
      
      const response = await this.publicClient.post<ApiResponse<PasswordResetInitResponse>>(
        '/auth/forgot-password',
        request
      );

      console.log('🔧 Public client response:', response.data);

      // Registrar el timestamp de la solicitud para rate limiting
      this.recordResetRequest();

      // 🔧 CORRECCIÓN: Retornar la estructura completa de data
      return response.data.data!;
    } catch (error: any) {
      console.error('🔧 Public client error:', error);
      throw this.handlePasswordResetError(error);
    }
  }

  /**
   * Verificar pregunta de seguridad durante reset (NUEVO)
   * POST /api/v1/auth/verify-security-question
   */
  async verifySecurityQuestion(request: PasswordResetVerifySecurityRequest): Promise<PasswordResetVerifySecurityResponse> {
    try {
      const response = await this.publicClient.post<ApiResponse<PasswordResetVerifySecurityResponse>>(
        '/auth/verify-security-question',
        request
      );

      return response.data.data!;
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
      const response = await this.publicClient.get<ApiResponse<PasswordResetStatusByEmailResponse>>(
        `/auth/reset-status?email=${encodeURIComponent(email)}`
      );

      return response.data.data!;
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
      const response = await this.publicClient.post<PasswordResetConfirmResponse>(
        '/auth/reset-password',
        request
      );

      // Limpiar rate limiting tras reset exitoso
      this.clearRateLimitRecord();

      return response.data;
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
      const apiClient = await this.getAuthenticatedClient();
      const response = await apiClient.get<ApiResponse<{ tokens: PasswordResetToken[]; count: number }>>(
        '/auth/reset-history'
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
      const apiClient = await this.getAuthenticatedClient();
      const response = await apiClient.post<PasswordResetCleanupResponse>(
        '/auth/admin/cleanup-tokens',
        {}
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
    if (!email || typeof email !== 'string') {
      return { isValid: false, error: 'Email es requerido' };
    }

    const trimmedEmail = email.trim();
    
    if (trimmedEmail.length === 0) {
      return { isValid: false, error: 'Email no puede estar vacío' };
    }

    // Validar formato básico
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(trimmedEmail)) {
      return { isValid: false, error: 'Formato de email inválido' };
    }

    // Validar dominio institucional
    if (!trimmedEmail.endsWith('@gamc.gov.bo')) {
      return { isValid: false, error: 'Solo se permiten emails @gamc.gov.bo' };
    }

    return { isValid: true };
  }

  /**
   * Validar contraseña antes de reset
   */
  validatePasswordForReset(password: string): { isValid: boolean; error?: string } {
    if (!password || typeof password !== 'string') {
      return { isValid: false, error: 'Contraseña es requerida' };
    }

    if (password.length < 8) {
      return { isValid: false, error: 'Contraseña debe tener al menos 8 caracteres' };
    }

    if (password.length > 128) {
      return { isValid: false, error: 'Contraseña no puede exceder 128 caracteres' };
    }

    // Validar complejidad básica
    const hasUpperCase = /[A-Z]/.test(password);
    const hasLowerCase = /[a-z]/.test(password);
    const hasNumbers = /\d/.test(password);
    const hasSpecialChar = /[!@#$%^&*(),.?":{}|<>]/.test(password);

    if (!hasUpperCase || !hasLowerCase || !hasNumbers || !hasSpecialChar) {
      return { 
        isValid: false, 
        error: 'Contraseña debe contener mayúsculas, minúsculas, números y símbolos' 
      };
    }

    return { isValid: true };
  }

  /**
   * Validar respuesta de pregunta de seguridad
   */
  validateSecurityAnswer(answer: string): { isValid: boolean; error?: string } {
    if (!answer || typeof answer !== 'string') {
      return { isValid: false, error: 'Respuesta es requerida' };
    }

    const trimmedAnswer = answer.trim();
    
    if (trimmedAnswer.length === 0) {
      return { isValid: false, error: 'Respuesta no puede estar vacía' };
    }

    if (trimmedAnswer.length < 1) {
      return { isValid: false, error: 'Respuesta debe tener al menos 1 caracter' };
    }

    if (trimmedAnswer.length > 100) {
      return { isValid: false, error: 'Respuesta no puede exceder 100 caracteres' };
    }

    return { isValid: true };
  }

  // ========================================
  // RATE LIMITING
  // ========================================

  recordResetRequest(): void {
    const timestamp = Date.now().toString();
    localStorage.setItem('lastPasswordResetRequest', timestamp);
  }

  clearRateLimitRecord(): void {
    localStorage.removeItem('lastPasswordResetRequest');
  }

  getRateLimitTimeRemaining(): number {
    const lastRequestTime = localStorage.getItem('lastPasswordResetRequest');
    if (!lastRequestTime) return 0;

    const timeDiff = Date.now() - parseInt(lastRequestTime);
    const timeRemaining = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW - timeDiff;
    
    return Math.max(0, Math.ceil(timeRemaining / 1000));
  }

  canMakeResetRequest(): boolean {
    return this.getRateLimitTimeRemaining() === 0;
  }

  // ========================================
  // MANEJO DE ERRORES ESPECÍFICOS
  // ========================================

  /**
   * Maneja errores específicos de reset de contraseña
   */
  private handlePasswordResetError(error: any): Error {
    console.log('🔧 Handling error:', error);

    // Error de red/timeout
    if (!error.response) {
      return new PasswordResetError(
        'network_error',
        'Error de conexión. Verifique su internet e intente nuevamente.'
      );
    }

    const { status, data } = error.response;
    
    switch (status) {
      case 400:
        return new PasswordResetError(
          'validation_error',
          data?.message || data?.error || 'Datos de entrada inválidos'
        );
      
      case 403:
        return new PasswordResetError(
          'email_not_institutional',
          'Solo usuarios con email @gamc.gov.bo pueden solicitar reset'
        );
      
      case 429:
        return new PasswordResetError(
          'rate_limit_exceeded',
          'Demasiadas solicitudes. Espere antes de intentar nuevamente'
        );
      
      case 404:
        return new PasswordResetError(
          'user_not_found',
          'Usuario no encontrado'
        );
      
      case 500:
        return new PasswordResetError(
          'server_error',
          'Error interno del servidor'
        );
      
      default:
        return new PasswordResetError(
          'unknown_error',
          data?.message || data?.error || 'Error inesperado'
        );
    }
  }
}

// Exportar instancia singleton
export const passwordResetService = new PasswordResetService();
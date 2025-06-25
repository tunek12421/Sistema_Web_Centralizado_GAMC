// src/services/passwordResetService.ts
// Servicio para todas las operaciones de reset de contraseña
// Integrado con la arquitectura existente de apiClient

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

/**
 * Clase para manejar todas las operaciones de reset de contraseña
 * Integrada con el sistema de interceptores existente para auth y refresh tokens
 */
class PasswordResetService {

  // ========================================
  // ENDPOINTS PÚBLICOS (SIN AUTENTICACIÓN)
  // ========================================

  /**
   * Solicitar reset de contraseña
   * POST /api/v1/auth/forgot-password
   */
  async requestPasswordReset(request: PasswordResetRequest): Promise<PasswordResetRequestResponse> {
    try {
      const response = await apiClient.post<PasswordResetRequestResponse>(
        '/auth/forgot-password',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          // No incluir Authorization header para endpoints públicos
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'request');
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
          // No incluir Authorization header para endpoints públicos
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'confirm');
    }
  }

  // ========================================
  // ENDPOINTS PROTEGIDOS (REQUIEREN AUTH)
  // ========================================

  /**
   * Obtener estado de tokens de reset del usuario actual
   * GET /api/v1/auth/reset-status
   * Requiere autenticación - el apiClient automáticamente incluye el Bearer token
   */
  async getPasswordResetStatus(): Promise<PasswordResetStatusResponse> {
    try {
      const response = await apiClient.get<PasswordResetStatusResponse>(
        '/auth/reset-status',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response;
    } catch (error: any) {
      throw this.handlePasswordResetError(error, 'status');
    }
  }

  // ========================================
  // ENDPOINTS DE ADMINISTRACIÓN (SOLO ADMINS)
  // ========================================

  /**
   * Limpiar tokens expirados (solo administradores)
   * POST /api/v1/auth/admin/cleanup-tokens
   * Requiere autenticación Y rol admin
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
   * Verifica formato y dominio institucional
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
  // MANEJO DE ERRORES ESPECÍFICOS
  // ========================================

  /**
   * Maneja errores específicos de reset de contraseña
   * Convierte errores HTTP en tipos de error específicos del dominio
   */
  private handlePasswordResetError(error: any, operation: 'request' | 'confirm' | 'status' | 'cleanup'): Error {
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
      
      case 'confirm':
        return this.handleConfirmErrors(status, errorMessage, data);
      
      case 'status':
        return this.handleStatusErrors(status, errorMessage, data);
      
      case 'cleanup':
        return this.handleCleanupErrors(status, errorMessage, data);
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.UNKNOWN_ERROR,
          errorMessage,
          error
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
        // Analizar mensaje específico para determinar tipo de error
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
      case 401:
        return new PasswordResetError(
          PasswordResetErrorType.NETWORK_ERROR,
          'Sesión expirada. Inicie sesión nuevamente',
          { status, message, data }
        );
      
      case 403:
        return new PasswordResetError(
          PasswordResetErrorType.VALIDATION_ERROR,
          'No tiene permisos para acceder a esta información',
          { status, message, data }
        );
      
      default:
        return new PasswordResetError(
          PasswordResetErrorType.SERVER_ERROR,
          `Error al obtener estado: ${message}`,
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
}

// ========================================
// CLASE DE ERROR PERSONALIZADA
// ========================================

/**
 * Error personalizado para operaciones de reset de contraseña
 * Incluye tipo específico y contexto adicional
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
// src/services/securityQuestionsService.ts
// Servicio API para gestión de preguntas de seguridad
// Maneja CRUD de preguntas, verificación durante reset y configuración usuario

import { apiClient } from './api';
import { ApiResponse } from '../types/auth';
import { PASSWORD_RESET_CONFIG } from '../types/passwordReset';

// ========================================
// INTERFACES ESPECÍFICAS
// ========================================

export interface SecurityQuestion {
  id: number;
  questionText: string;
  category: 'personal' | 'education' | 'professional' | 'preferences';
}

export interface SecurityQuestionResponse {
  id: number;
  questionText: string;
  category: string;
}

export interface UserSecurityQuestion {
  questionId: number;
  question: SecurityQuestion;
}

export interface SecurityQuestionForReset {
  questionId: number;
  questionText: string;
  attempts: number;
  maxAttempts: number;
}

export interface SecurityQuestionsStatus {
  hasSecurityQuestions: boolean;
  questionsCount: number;
  maxQuestions: number;
  questions: UserSecurityQuestion[];
}

export interface SetupSecurityQuestionsRequest {
  questions: Array<{
    questionId: number;
    answer: string;
  }>;
}

export interface UpdateSecurityQuestionRequest {
  newAnswer: string;
}

export interface VerifySecurityQuestionRequest {
  email: string;
  questionId: number;
  answer: string;
}

export interface VerifySecurityQuestionResponse {
  success: boolean;
  message: string;
  verified: boolean;
  canProceedToReset: boolean;
  attemptsRemaining: number;
  resetToken?: string;
}

export interface SecurityQuestionStats {
  totalQuestions: number;
  usersWithQuestions: number;
  usageByCategory: Record<string, number>;
  averageQuestionsPerUser: number;
}

// ========================================
// TIPOS DE ERRORES ESPECÍFICOS
// ========================================

export enum SecurityQuestionErrorType {
  QUESTION_NOT_FOUND = 'question_not_found',
  INVALID_ANSWER = 'invalid_answer',
  MAX_ATTEMPTS_REACHED = 'max_attempts_reached',
  QUESTION_ALREADY_EXISTS = 'question_already_exists',
  MAX_QUESTIONS_LIMIT = 'max_questions_limit',
  ANSWER_TOO_SHORT = 'answer_too_short',
  ANSWER_TOO_LONG = 'answer_too_long',
  NETWORK_ERROR = 'network_error',
  VALIDATION_ERROR = 'validation_error'
}

export class SecurityQuestionError extends Error {
  constructor(
    public type: SecurityQuestionErrorType,
    message: string,
    public details?: any
  ) {
    super(message);
    this.name = 'SecurityQuestionError';
  }
}

// ========================================
// SERVICIO PRINCIPAL
// ========================================

class SecurityQuestionsService {

  // ========================================
  // ENDPOINTS PÚBLICOS (Catálogo de preguntas)
  // ========================================

  /**
   * Obtiene el catálogo completo de preguntas de seguridad disponibles
   * GET /api/v1/auth/security-questions
   */
  async getSecurityQuestionsCatalog(): Promise<SecurityQuestion[]> {
    try {
      const response = await apiClient.get<ApiResponse<{ questions: SecurityQuestionResponse[]; count: number }>>(
        '/auth/security-questions',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      // Convertir respuesta del backend al formato esperado
      return response.data!.questions.map(q => ({
        id: q.id,
        questionText: q.questionText,
        category: q.category as any
      }));
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'get_catalog');
    }
  }

  /**
   * Obtiene preguntas organizadas por categoría
   */
  async getSecurityQuestionsByCategory(): Promise<Record<string, SecurityQuestion[]>> {
    const questions = await this.getSecurityQuestionsCatalog();
    
    return questions.reduce((acc, question) => {
      if (!acc[question.category]) {
        acc[question.category] = [];
      }
      acc[question.category].push(question);
      return acc;
    }, {} as Record<string, SecurityQuestion[]>);
  }

  // ========================================
  // ENDPOINTS PROTEGIDOS (Gestión usuario)
  // ========================================

  /**
   * Obtiene el estado de preguntas de seguridad del usuario actual
   * GET /api/v1/auth/security-status
   */
  async getUserSecurityStatus(): Promise<SecurityQuestionsStatus> {
    try {
      const response = await apiClient.get<ApiResponse<SecurityQuestionsStatus>>(
        '/auth/security-status',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response.data!;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'get_status');
    }
  }

  /**
   * Configura preguntas de seguridad para el usuario actual
   * POST /api/v1/auth/security-questions
   */
  async setupSecurityQuestions(request: SetupSecurityQuestionsRequest): Promise<UserSecurityQuestion[]> {
    try {
      // Validar entrada
      this.validateSetupRequest(request);

      const response = await apiClient.post<ApiResponse<{ configured: number; questions: UserSecurityQuestion[] }>>(
        '/auth/security-questions',
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );

      return response.data!.questions;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'setup');
    }
  }

  /**
   * Actualiza la respuesta de una pregunta de seguridad específica
   * PUT /api/v1/auth/security-questions/:questionId
   */
  async updateSecurityQuestion(questionId: number, request: UpdateSecurityQuestionRequest): Promise<void> {
    try {
      // Validar entrada
      this.validateUpdateRequest(request);

      await apiClient.put<ApiResponse<any>>(
        `/auth/security-questions/${questionId}`,
        request,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      );
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'update');
    }
  }

  /**
   * Elimina una pregunta de seguridad específica
   * DELETE /api/v1/auth/security-questions/:questionId
   */
  async removeSecurityQuestion(questionId: number): Promise<void> {
    try {
      await apiClient.delete<ApiResponse<any>>(
        `/auth/security-questions/${questionId}`,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'remove');
    }
  }

  // ========================================
  // ENDPOINTS DE VERIFICACIÓN (Reset de contraseña)
  // ========================================

  /**
   * Verifica respuesta a pregunta de seguridad durante reset de contraseña
   * POST /api/v1/auth/verify-security-question
   */
  async verifySecurityQuestion(request: VerifySecurityQuestionRequest): Promise<VerifySecurityQuestionResponse> {
    try {
      // Validar entrada
      this.validateVerifyRequest(request);

      const response = await apiClient.post<ApiResponse<VerifySecurityQuestionResponse>>(
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
      throw this.handleSecurityQuestionError(error, 'verify');
    }
  }

  // ========================================
  // ENDPOINTS ADMINISTRATIVOS
  // ========================================

  /**
   * Obtiene estadísticas de uso de preguntas de seguridad (solo admins)
   * GET /api/v1/auth/admin/security-questions/stats
   */
  async getSecurityQuestionsStats(): Promise<SecurityQuestionStats> {
    try {
      const response = await apiClient.get<ApiResponse<SecurityQuestionStats>>(
        '/auth/admin/security-questions/stats',
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );

      return response.data!;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error, 'stats');
    }
  }

  // ========================================
  // VALIDACIONES
  // ========================================

  /**
   * Valida solicitud de configuración de preguntas
   */
  private validateSetupRequest(request: SetupSecurityQuestionsRequest): void {
    if (!request.questions || !Array.isArray(request.questions)) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Las preguntas son requeridas y deben ser un array'
      );
    }

    if (request.questions.length === 0) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Debe configurar al menos una pregunta de seguridad'
      );
    }

    if (request.questions.length > 3) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.MAX_QUESTIONS_LIMIT,
        'No puede configurar más de 3 preguntas de seguridad'
      );
    }

    // Validar cada pregunta
    request.questions.forEach((q, index) => {
      if (!q.questionId || typeof q.questionId !== 'number') {
        throw new SecurityQuestionError(
          SecurityQuestionErrorType.VALIDATION_ERROR,
          `Pregunta ${index + 1}: ID de pregunta inválido`
        );
      }

      if (!q.answer || typeof q.answer !== 'string') {
        throw new SecurityQuestionError(
          SecurityQuestionErrorType.VALIDATION_ERROR,
          `Pregunta ${index + 1}: Respuesta es requerida`
        );
      }

      const trimmedAnswer = q.answer.trim();
      if (trimmedAnswer.length < 2) {
        throw new SecurityQuestionError(
          SecurityQuestionErrorType.ANSWER_TOO_SHORT,
          `Pregunta ${index + 1}: La respuesta debe tener al menos 2 caracteres`
        );
      }

      if (trimmedAnswer.length > 100) {
        throw new SecurityQuestionError(
          SecurityQuestionErrorType.ANSWER_TOO_LONG,
          `Pregunta ${index + 1}: La respuesta no puede exceder 100 caracteres`
        );
      }
    });

    // Verificar que no haya preguntas duplicadas
    const questionIds = request.questions.map(q => q.questionId);
    const uniqueIds = new Set(questionIds);
    if (uniqueIds.size !== questionIds.length) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.QUESTION_ALREADY_EXISTS,
        'No puede configurar la misma pregunta múltiples veces'
      );
    }
  }

  /**
   * Valida solicitud de actualización de pregunta
   */
  private validateUpdateRequest(request: UpdateSecurityQuestionRequest): void {
    if (!request.newAnswer || typeof request.newAnswer !== 'string') {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Nueva respuesta es requerida'
      );
    }

    const trimmedAnswer = request.newAnswer.trim();
    if (trimmedAnswer.length < 2) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.ANSWER_TOO_SHORT,
        'La respuesta debe tener al menos 2 caracteres'
      );
    }

    if (trimmedAnswer.length > 100) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.ANSWER_TOO_LONG,
        'La respuesta no puede exceder 100 caracteres'
      );
    }
  }

  /**
   * Valida solicitud de verificación de pregunta
   */
  private validateVerifyRequest(request: VerifySecurityQuestionRequest): void {
    if (!request.email || typeof request.email !== 'string') {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Email es requerido'
      );
    }

    if (!PASSWORD_RESET_CONFIG.EMAIL_PATTERN.test(request.email)) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Solo emails @gamc.gov.bo son válidos'
      );
    }

    if (!request.questionId || typeof request.questionId !== 'number') {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'ID de pregunta es requerido'
      );
    }

    if (!request.answer || typeof request.answer !== 'string') {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'Respuesta es requerida'
      );
    }

    const trimmedAnswer = request.answer.trim();
    if (trimmedAnswer.length < 1) {
      throw new SecurityQuestionError(
        SecurityQuestionErrorType.VALIDATION_ERROR,
        'La respuesta no puede estar vacía'
      );
    }
  }

  // ========================================
  // MANEJO DE ERRORES
  // ========================================

  /**
   * Maneja errores específicos de preguntas de seguridad
   */
  private handleSecurityQuestionError(error: any, operation: string): Error {
    // Error de red/timeout
    if (!error.response) {
      return new SecurityQuestionError(
        SecurityQuestionErrorType.NETWORK_ERROR,
        'Error de conexión. Verifique su internet e intente nuevamente.'
      );
    }

    const { status, data } = error.response;

    // Errores específicos del backend
    switch (status) {
      case 400:
        if (data?.error?.includes('pregunta no encontrada')) {
          return new SecurityQuestionError(
            SecurityQuestionErrorType.QUESTION_NOT_FOUND,
            'La pregunta de seguridad no existe o no está disponible'
          );
        }
        if (data?.error?.includes('respuesta incorrecta')) {
          return new SecurityQuestionError(
            SecurityQuestionErrorType.INVALID_ANSWER,
            data.error || 'Respuesta incorrecta'
          );
        }
        if (data?.error?.includes('máximo de preguntas')) {
          return new SecurityQuestionError(
            SecurityQuestionErrorType.MAX_QUESTIONS_LIMIT,
            'Ha alcanzado el límite máximo de preguntas de seguridad'
          );
        }
        if (data?.error?.includes('demasiados intentos')) {
          return new SecurityQuestionError(
            SecurityQuestionErrorType.MAX_ATTEMPTS_REACHED,
            'Demasiados intentos fallidos. El proceso ha sido bloqueado por seguridad.'
          );
        }
        break;

      case 401:
        return new SecurityQuestionError(
          SecurityQuestionErrorType.VALIDATION_ERROR,
          'Sesión expirada. Inicie sesión nuevamente.'
        );

      case 403:
        return new SecurityQuestionError(
          SecurityQuestionErrorType.VALIDATION_ERROR,
          'No tiene permisos para realizar esta acción'
        );

      case 429:
        return new SecurityQuestionError(
          SecurityQuestionErrorType.VALIDATION_ERROR,
          'Demasiadas solicitudes. Espere unos minutos antes de intentar nuevamente.'
        );

      case 500:
      default:
        return new SecurityQuestionError(
          SecurityQuestionErrorType.NETWORK_ERROR,
          'Error interno del servidor. Intente nuevamente más tarde.'
        );
    }

    // Error genérico
    return new SecurityQuestionError(
      SecurityQuestionErrorType.VALIDATION_ERROR,
      data?.message || data?.error || 'Error desconocido en preguntas de seguridad'
    );
  }

  // ========================================
  // UTILIDADES
  // ========================================

  /**
   * Normaliza respuesta de pregunta de seguridad para comparación
   */
  normalizeAnswer(answer: string): string {
    return answer.trim().toLowerCase();
  }

  /**
   * Obtiene categorías disponibles
   */
  getAvailableCategories(): Array<{ id: string; name: string; description: string }> {
    return [
      {
        id: 'personal',
        name: 'Personal',
        description: 'Información personal y familiar'
      },
      {
        id: 'education',
        name: 'Educación',
        description: 'Información sobre estudios y formación'
      },
      {
        id: 'professional',
        name: 'Profesional',
        description: 'Información sobre trabajo en el GAMC'
      },
      {
        id: 'preferences',
        name: 'Preferencias',
        description: 'Gustos personales y preferencias'
      }
    ];
  }

  /**
   * Verifica si una pregunta pertenece a una categoría específica
   */
  isQuestionInCategory(question: SecurityQuestion, category: string): boolean {
    return question.category === category;
  }
}

// Exportar instancia singleton
export const securityQuestionsService = new SecurityQuestionsService();
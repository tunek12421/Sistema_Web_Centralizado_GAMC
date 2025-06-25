// src/services/securityQuestionsService.ts
// Servicio API para gestión de preguntas de seguridad
// Maneja CRUD de preguntas, verificación durante reset y configuración usuario

import { apiClient } from './api';
import type { ApiResponse } from '../types/index';
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

export type SecurityQuestionErrorType = 
  | 'question_not_found'
  | 'invalid_answer'
  | 'max_attempts_reached'
  | 'question_already_exists'
  | 'max_questions_limit'
  | 'answer_too_short'
  | 'answer_too_long'
  | 'network_error'
  | 'validation_error';

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

      return response.data!.questions.map(q => ({
        id: q.id,
        questionText: q.questionText,
        category: q.category as any
      }));
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

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
      throw this.handleSecurityQuestionError(error);
    }
  }

  async setupSecurityQuestions(request: SetupSecurityQuestionsRequest): Promise<UserSecurityQuestion[]> {
    try {
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
      throw this.handleSecurityQuestionError(error);
    }
  }

  async updateSecurityQuestion(questionId: number, request: UpdateSecurityQuestionRequest): Promise<void> {
    try {
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
      throw this.handleSecurityQuestionError(error);
    }
  }

  async removeSecurityQuestion(questionId: number): Promise<void> {
    try {
      await apiClient.delete<ApiResponse<any>>(
        `/auth/security-questions/${questionId}`,
        {
          timeout: PASSWORD_RESET_CONFIG.REQUEST_TIMEOUT
        }
      );
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  async verifySecurityQuestion(request: VerifySecurityQuestionRequest): Promise<VerifySecurityQuestionResponse> {
    try {
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
      throw this.handleSecurityQuestionError(error);
    }
  }

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
      throw this.handleSecurityQuestionError(error);
    }
  }

  private validateSetupRequest(request: SetupSecurityQuestionsRequest): void {
    if (!request.questions || !Array.isArray(request.questions)) {
      throw new SecurityQuestionError(
        'validation_error',
        'Las preguntas son requeridas y deben ser un array'
      );
    }

    if (request.questions.length === 0) {
      throw new SecurityQuestionError(
        'validation_error',
        'Debe configurar al menos una pregunta de seguridad'
      );
    }

    if (request.questions.length > 3) {
      throw new SecurityQuestionError(
        'max_questions_limit',
        'No puede configurar más de 3 preguntas de seguridad'
      );
    }

    request.questions.forEach((q, index) => {
      if (!q.questionId || typeof q.questionId !== 'number') {
        throw new SecurityQuestionError(
          'validation_error',
          `Pregunta ${index + 1}: ID de pregunta inválido`
        );
      }

      if (!q.answer || typeof q.answer !== 'string') {
        throw new SecurityQuestionError(
          'validation_error',
          `Pregunta ${index + 1}: Respuesta es requerida`
        );
      }

      const trimmedAnswer = q.answer.trim();
      if (trimmedAnswer.length < 2) {
        throw new SecurityQuestionError(
          'answer_too_short',
          `Pregunta ${index + 1}: La respuesta debe tener al menos 2 caracteres`
        );
      }

      if (trimmedAnswer.length > 100) {
        throw new SecurityQuestionError(
          'answer_too_long',
          `Pregunta ${index + 1}: La respuesta no puede exceder 100 caracteres`
        );
      }
    });

    const questionIds = request.questions.map(q => q.questionId);
    const uniqueIds = new Set(questionIds);
    if (uniqueIds.size !== questionIds.length) {
      throw new SecurityQuestionError(
        'question_already_exists',
        'No puede configurar la misma pregunta múltiples veces'
      );
    }
  }

  private validateUpdateRequest(request: UpdateSecurityQuestionRequest): void {
    if (!request.newAnswer || typeof request.newAnswer !== 'string') {
      throw new SecurityQuestionError(
        'validation_error',
        'Nueva respuesta es requerida'
      );
    }

    const trimmedAnswer = request.newAnswer.trim();
    if (trimmedAnswer.length < 2) {
      throw new SecurityQuestionError(
        'answer_too_short',
        'La respuesta debe tener al menos 2 caracteres'
      );
    }

    if (trimmedAnswer.length > 100) {
      throw new SecurityQuestionError(
        'answer_too_long',
        'La respuesta no puede exceder 100 caracteres'
      );
    }
  }

  private validateVerifyRequest(request: VerifySecurityQuestionRequest): void {
    if (!request.email || typeof request.email !== 'string') {
      throw new SecurityQuestionError(
        'validation_error',
        'Email es requerido'
      );
    }

    if (!PASSWORD_RESET_CONFIG.EMAIL_PATTERN.test(request.email)) {
      throw new SecurityQuestionError(
        'validation_error',
        'Solo emails @gamc.gov.bo son válidos'
      );
    }

    if (!request.questionId || typeof request.questionId !== 'number') {
      throw new SecurityQuestionError(
        'validation_error',
        'ID de pregunta es requerido'
      );
    }

    if (!request.answer || typeof request.answer !== 'string') {
      throw new SecurityQuestionError(
        'validation_error',
        'Respuesta es requerida'
      );
    }

    const trimmedAnswer = request.answer.trim();
    if (trimmedAnswer.length < 1) {
      throw new SecurityQuestionError(
        'validation_error',
        'La respuesta no puede estar vacía'
      );
    }
  }

  private handleSecurityQuestionError(error: any): Error {
    if (!error.response) {
      return new SecurityQuestionError(
        'network_error',
        'Error de conexión. Verifique su internet e intente nuevamente.'
      );
    }

    const { status, data } = error.response;

    switch (status) {
      case 400:
        if (data?.error?.includes('pregunta no encontrada')) {
          return new SecurityQuestionError(
            'question_not_found',
            'La pregunta de seguridad no existe o no está disponible'
          );
        }
        if (data?.error?.includes('respuesta incorrecta')) {
          return new SecurityQuestionError(
            'invalid_answer',
            data.error || 'Respuesta incorrecta'
          );
        }
        if (data?.error?.includes('máximo de preguntas')) {
          return new SecurityQuestionError(
            'max_questions_limit',
            'Ha alcanzado el límite máximo de preguntas de seguridad'
          );
        }
        if (data?.error?.includes('demasiados intentos')) {
          return new SecurityQuestionError(
            'max_attempts_reached',
            'Demasiados intentos fallidos. El proceso ha sido bloqueado por seguridad.'
          );
        }
        break;

      case 401:
        return new SecurityQuestionError(
          'validation_error',
          'Sesión expirada. Inicie sesión nuevamente.'
        );

      case 403:
        return new SecurityQuestionError(
          'validation_error',
          'No tiene permisos para realizar esta acción'
        );

      case 429:
        return new SecurityQuestionError(
          'validation_error',
          'Demasiadas solicitudes. Espere unos minutos antes de intentar nuevamente.'
        );

      case 500:
      default:
        return new SecurityQuestionError(
          'network_error',
          'Error interno del servidor. Intente nuevamente más tarde.'
        );
    }

    return new SecurityQuestionError(
      'validation_error',
      data?.message || data?.error || 'Error desconocido en preguntas de seguridad'
    );
  }

  normalizeAnswer(answer: string): string {
    return answer.trim().toLowerCase();
  }

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

  isQuestionInCategory(question: SecurityQuestion, category: string): boolean {
    return question.category === category;
  }
}

export const securityQuestionsService = new SecurityQuestionsService();
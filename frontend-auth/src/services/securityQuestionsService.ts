// src/services/securityQuestionsService.ts
// CORRECCIÓN: Usar cliente sin credentials para operaciones públicas

import axios from 'axios';
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
  // OPERACIONES PÚBLICAS (sin autenticación)
  // ========================================

  async getSecurityQuestionsCatalog(): Promise<SecurityQuestion[]> {
    try {
      const response = await this.publicClient.get<ApiResponse<{ questions: SecurityQuestionResponse[]; count: number }>>(
        '/auth/security-questions'
      );

      return response.data.data!.questions.map(q => ({
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

  /**
   * 🔧 CORRECCIÓN: Verificar pregunta de seguridad durante reset (PÚBLICO)
   * POST /api/v1/auth/verify-security-question
   */
  async verifySecurityQuestion(request: VerifySecurityQuestionRequest): Promise<VerifySecurityQuestionResponse> {
    try {
      console.log('🔧 Using public client for verify-security-question');
      this.validateVerifyRequest(request);

      const response = await this.publicClient.post<ApiResponse<VerifySecurityQuestionResponse>>(
        '/auth/verify-security-question',
        request
      );

      console.log('🔧 Verify security question response:', response.data);
      return response.data.data!;
    } catch (error: any) {
      console.error('🔧 Verify security question error:', error);
      throw this.handleSecurityQuestionError(error);
    }
  }

  // ========================================
  // OPERACIONES PROTEGIDAS (requieren autenticación)
  // ========================================

  async getUserSecurityStatus(): Promise<SecurityQuestionsStatus> {
    try {
      const apiClient = await this.getAuthenticatedClient();
      const response = await apiClient.get<ApiResponse<SecurityQuestionsStatus>>(
        '/auth/security-status'
      );

      return response.data!;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  async setupSecurityQuestions(request: SetupSecurityQuestionsRequest): Promise<UserSecurityQuestion[]> {
    try {
      this.validateSetupRequest(request);

      const apiClient = await this.getAuthenticatedClient();
      const response = await apiClient.post<ApiResponse<{ configured: number; questions: UserSecurityQuestion[] }>>(
        '/auth/security-questions',
        request
      );

      return response.data!.questions;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  async updateSecurityQuestion(questionId: number, request: UpdateSecurityQuestionRequest): Promise<void> {
    try {
      this.validateUpdateRequest(request);

      const apiClient = await this.getAuthenticatedClient();
      await apiClient.put<ApiResponse<any>>(
        `/auth/security-questions/${questionId}`,
        request
      );
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  async removeSecurityQuestion(questionId: number): Promise<void> {
    try {
      const apiClient = await this.getAuthenticatedClient();
      await apiClient.delete<ApiResponse<any>>(
        `/auth/security-questions/${questionId}`
      );
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  async getSecurityQuestionsStats(): Promise<SecurityQuestionStats> {
    try {
      const apiClient = await this.getAuthenticatedClient();
      const response = await apiClient.get<ApiResponse<SecurityQuestionStats>>(
        '/auth/admin/security-questions/stats'
      );

      return response.data!;
    } catch (error: any) {
      throw this.handleSecurityQuestionError(error);
    }
  }

  // ========================================
  // VALIDACIONES
  // ========================================

  private validateSetupRequest(request: SetupSecurityQuestionsRequest): void {
    if (!request.questions || !Array.isArray(request.questions)) {
      throw new SecurityQuestionError('validation_error', 'Questions array is required');
    }

    if (request.questions.length === 0) {
      throw new SecurityQuestionError('validation_error', 'At least one question is required');
    }

    if (request.questions.length > 3) {
      throw new SecurityQuestionError('max_questions_limit', 'Maximum 3 questions allowed');
    }

    for (const question of request.questions) {
      if (!question.questionId || !question.answer) {
        throw new SecurityQuestionError('validation_error', 'Each question must have questionId and answer');
      }

      if (question.answer.trim().length < 1) {
        throw new SecurityQuestionError('answer_too_short', 'Answer must be at least 1 character');
      }

      if (question.answer.trim().length > 100) {
        throw new SecurityQuestionError('answer_too_long', 'Answer cannot exceed 100 characters');
      }
    }
  }

  private validateUpdateRequest(request: UpdateSecurityQuestionRequest): void {
    if (!request.newAnswer || typeof request.newAnswer !== 'string') {
      throw new SecurityQuestionError('validation_error', 'New answer is required');
    }

    if (request.newAnswer.trim().length < 1) {
      throw new SecurityQuestionError('answer_too_short', 'Answer must be at least 1 character');
    }

    if (request.newAnswer.trim().length > 100) {
      throw new SecurityQuestionError('answer_too_long', 'Answer cannot exceed 100 characters');
    }
  }

  private validateVerifyRequest(request: VerifySecurityQuestionRequest): void {
    if (!request.email || !request.questionId || !request.answer) {
      throw new SecurityQuestionError('validation_error', 'Email, questionId and answer are required');
    }

    if (!request.email.endsWith('@gamc.gov.bo')) {
      throw new SecurityQuestionError('validation_error', 'Only @gamc.gov.bo emails are allowed');
    }

    if (request.answer.trim().length < 1) {
      throw new SecurityQuestionError('answer_too_short', 'Answer must be at least 1 character');
    }
  }

  // ========================================
  // MANEJO DE ERRORES
  // ========================================

  private handleSecurityQuestionError(error: any): SecurityQuestionError {
    console.log('🔧 Handling security question error:', error);

    // Error de red/timeout
    if (!error.response) {
      return new SecurityQuestionError(
        'network_error',
        'Error de conexión. Verifique su internet e intente nuevamente.'
      );
    }

    const { status, data } = error.response;
    
    switch (status) {
      case 400:
        return new SecurityQuestionError(
          'validation_error',
          data?.message || data?.error || 'Datos de entrada inválidos'
        );
      
      case 403:
        return new SecurityQuestionError(
          'invalid_answer',
          'Respuesta incorrecta'
        );
      
      case 429:
        return new SecurityQuestionError(
          'max_attempts_reached',
          'Demasiados intentos. Intente más tarde'
        );
      
      case 404:
        return new SecurityQuestionError(
          'question_not_found',
          'Pregunta de seguridad no encontrada'
        );
      
      case 500:
        return new SecurityQuestionError(
          'network_error',
          'Error interno del servidor'
        );
      
      default:
        return new SecurityQuestionError(
          'validation_error',
          data?.message || data?.error || 'Error inesperado'
        );
    }
  }
}

// Exportar instancia singleton
export const securityQuestionsService = new SecurityQuestionsService();
// src/components/auth/SecurityQuestionForm.tsx
// Componente para verificar pregunta de seguridad durante reset de contraseña
// Incluye manejo de intentos, validaciones y UI de progreso

import React, { useState, useEffect, useRef } from 'react';
import { securityQuestionsService, SecurityQuestionError, VerifySecurityQuestionRequest } from '../../services/securityQuestionsService';
import { getInputClasses, getValidationMessageClasses } from '../../utils/passwordValidation';
import { FieldValidation } from '../../types/passwordReset';

interface SecurityQuestionData {
  questionId: number;
  questionText: string;
  attempts: number;
  maxAttempts: number;
}

interface SecurityQuestionFormProps {
  email: string;
  securityQuestion: SecurityQuestionData;
  onSuccess?: (resetToken: string) => void;
  onError?: (error: SecurityQuestionError) => void;
  onBack?: () => void;
  onNewReset?: () => void;
  autoFocus?: boolean;
}

const SecurityQuestionForm: React.FC<SecurityQuestionFormProps> = ({
  email,
  securityQuestion,
  onSuccess,
  onError,
  onBack,
  onNewReset,
  autoFocus = true
}) => {
  // Referencias para DOM
  const answerInputRef = useRef<HTMLInputElement>(null);

  // Estados del formulario
  const [answer, setAnswer] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error' | 'warning' | ''>('');

  // Estados de validación
  const [answerValidation, setAnswerValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });

  // Estados de progreso
  const [currentAttempts, setCurrentAttempts] = useState(securityQuestion.attempts);
  const [attemptsRemaining, setAttemptsRemaining] = useState(securityQuestion.maxAttempts - securityQuestion.attempts);
  const [isBlocked, setIsBlocked] = useState(false);

  // ========================================
  // EFECTOS Y INICIALIZACIÓN
  // ========================================

  // Auto-focus en el campo respuesta
  useEffect(() => {
    if (autoFocus && answerInputRef.current) {
      answerInputRef.current.focus();
    }
  }, [autoFocus]);

  // Verificar si ya está bloqueado
  useEffect(() => {
    if (attemptsRemaining <= 0) {
      setIsBlocked(true);
      setMessage('❌ Se han agotado los intentos disponibles. Solicite un nuevo reset de contraseña.');
      setMessageType('error');
    }
  }, [attemptsRemaining]);

  // ========================================
  // HANDLERS Y VALIDACIONES
  // ========================================

  /**
   * Valida respuesta de seguridad
   */
  const validateAnswer = (answerText: string): FieldValidation => {
    if (!answerText || typeof answerText !== 'string') {
      return {
        isValid: false,
        message: 'La respuesta es requerida',
        type: 'error'
      };
    }

    const trimmedAnswer = answerText.trim();
    
    if (trimmedAnswer.length === 0) {
      return {
        isValid: false,
        message: 'La respuesta no puede estar vacía',
        type: 'error'
      };
    }

    if (trimmedAnswer.length < 1) {
      return {
        isValid: false,
        message: 'La respuesta debe tener al menos 1 caracter',
        type: 'error'
      };
    }

    if (trimmedAnswer.length > 100) {
      return {
        isValid: false,
        message: 'La respuesta no puede exceder 100 caracteres',
        type: 'error'
      };
    }

    return {
      isValid: true,
      message: '✓ Respuesta válida',
      type: 'success'
    };
  };

  /**
   * Maneja cambios en el campo respuesta
   */
  const handleAnswerChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newAnswer = e.target.value;
    setAnswer(newAnswer);
    
    // Validar en tiempo real
    const validation = validateAnswer(newAnswer);
    setAnswerValidation(validation);
    
    // Limpiar mensajes previos
    if (message && messageType !== 'error') {
      setMessage('');
      setMessageType('');
    }
  };

  /**
   * Maneja envío del formulario
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Prevenir envío si está bloqueado o cargando
    if (isLoading || isBlocked || attemptsRemaining <= 0) return;

    setIsLoading(true);
    setMessage('');
    setMessageType('');

    try {
      // Validar respuesta final
      const validation = validateAnswer(answer);
      setAnswerValidation(validation);

      if (!validation.isValid) {
        setMessage(validation.message);
        setMessageType('error');
        return;
      }

      // Preparar solicitud
      const request: VerifySecurityQuestionRequest = {
        email: email.trim(),
        questionId: securityQuestion.questionId,
        answer: answer.trim()
      };

      // Realizar verificación
      const result = await securityQuestionsService.verifySecurityQuestion(request);
      
      if (result.verified && result.resetToken) {
        // Verificación exitosa
        setIsSubmitted(true);
        setMessage('✅ Respuesta correcta. Redirigiendo...');
        setMessageType('success');

        // Callback de éxito con token
        if (onSuccess) {
          onSuccess(result.resetToken);
        }

      } else {
        // Respuesta incorrecta
        const newAttempts = currentAttempts + 1;
        const newRemaining = result.attemptsRemaining;
        
        setCurrentAttempts(newAttempts);
        setAttemptsRemaining(newRemaining);
        
        if (newRemaining <= 0) {
          setIsBlocked(true);
          setMessage('❌ Se han agotado los intentos. El proceso ha sido bloqueado por seguridad.');
          setMessageType('error');
        } else {
          setMessage(`❌ Respuesta incorrecta. Intentos restantes: ${newRemaining}`);
          setMessageType('error');
        }
        
        // Limpiar campo de respuesta
        setAnswer('');
        setAnswerValidation({ isValid: false, message: '', type: '' });
        
        // Focus nuevamente
        if (answerInputRef.current) {
          answerInputRef.current.focus();
        }
      }

    } catch (error: any) {
      console.error('Error en verificación de pregunta:', error);
      
      let errorMessage = 'Error al verificar respuesta';
      
      if (error instanceof SecurityQuestionError) {
        errorMessage = error.message;
        
        // Manejar casos específicos
        if (error.type === 'max_attempts_reached') {
          setIsBlocked(true);
          setAttemptsRemaining(0);
        }
      } else if (error.response) {
        const { status, data } = error.response;
        
        switch (status) {
          case 400:
            if (data?.error?.includes('respuesta incorrecta')) {
              // Extraer intentos restantes del mensaje
              const match = data.error.match(/intentos restantes: (\d+)/);
              if (match) {
                const remaining = parseInt(match[1]);
                setAttemptsRemaining(remaining);
                setCurrentAttempts(securityQuestion.maxAttempts - remaining);
                
                if (remaining <= 0) {
                  setIsBlocked(true);
                }
              }
              errorMessage = data.error;
            } else {
              errorMessage = data?.message || data?.error || 'Error de validación';
            }
            break;
          case 429:
            errorMessage = 'Demasiados intentos. Espere antes de intentar nuevamente';
            break;
          default:
            errorMessage = data?.message || 'Error del servidor';
        }
      } else {
        errorMessage = 'Error de conexión. Verifique su internet e intente nuevamente';
      }

      setMessage('❌ ' + errorMessage);
      setMessageType('error');

      // Callback de error
      if (onError && error instanceof SecurityQuestionError) {
        onError(error);
      }

    } finally {
      setIsLoading(false);
    }
  };

  // ========================================
  // COMPONENTES DE UI
  // ========================================

  const ProgressIndicator = () => (
    <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
      <div className="flex items-center justify-between mb-3">
        <span className="text-blue-800 font-semibold text-sm">Verificación de Seguridad</span>
        <span className="text-blue-600 text-xs">
          Intento {currentAttempts + 1} de {securityQuestion.maxAttempts}
        </span>
      </div>
      
      {/* Barra de progreso */}
      <div className="w-full bg-blue-200 rounded-full h-2 mb-2">
        <div 
          className={`h-2 rounded-full transition-all duration-300 ${
            attemptsRemaining > 1 ? 'bg-blue-600' : 
            attemptsRemaining === 1 ? 'bg-yellow-500' : 'bg-red-500'
          }`}
          style={{ 
            width: `${Math.max(0, (attemptsRemaining / securityQuestion.maxAttempts) * 100)}%` 
          }}
        ></div>
      </div>
      
      <p className="text-blue-700 text-xs">
        {attemptsRemaining > 1 ? (
          `Quedan ${attemptsRemaining} intentos`
        ) : attemptsRemaining === 1 ? (
          '⚠️ Último intento disponible'
        ) : (
          '🔒 Sin intentos restantes'
        )}
      </p>
    </div>
  );

  const QuestionDisplay = () => (
    <div className="mb-6 p-4 bg-gray-50 border border-gray-200 rounded-lg">
      <h3 className="text-gray-800 font-semibold text-sm mb-2">Pregunta de Seguridad:</h3>
      <p className="text-gray-700 text-base italic">
        "{securityQuestion.questionText}"
      </p>
      <p className="text-gray-500 text-xs mt-2">
        💡 Ingrese su respuesta exactamente como la configuró
      </p>
    </div>
  );

  const BlockedMessage = () => {
    if (!isBlocked) return null;

    return (
      <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
        <div className="flex items-start">
          <span className="text-red-600 text-lg mr-3 mt-0.5">🔒</span>
          <div>
            <h4 className="text-red-800 font-semibold text-sm mb-2">
              Proceso bloqueado por seguridad
            </h4>
            <p className="text-red-700 text-sm mb-3">
              Ha agotado todos los intentos disponibles para verificar su pregunta de seguridad.
            </p>
            <div className="space-y-2">
              <button
                onClick={onNewReset}
                className="w-full py-2 px-4 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 transition-colors text-sm"
              >
                🔄 Solicitar nuevo reset
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // ========================================
  // RENDER PRINCIPAL
  // ========================================

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 to-purple-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-indigo-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">❓</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Verificación de Seguridad</h1>
          <p className="text-gray-600 mt-2">
            Responda su pregunta de seguridad para continuar
          </p>
        </div>

        {/* Progreso */}
        <ProgressIndicator />

        {/* Pregunta */}
        <QuestionDisplay />

        {/* Mensaje de bloqueo */}
        <BlockedMessage />

        {/* Formulario */}
        {!isBlocked && !isSubmitted && (
          <form onSubmit={handleSubmit} className="space-y-6">
            
            {/* Campo Respuesta */}
            <div>
              <label htmlFor="answer" className="block text-sm font-medium text-gray-700 mb-2">
                Su Respuesta
              </label>
              <input
                ref={answerInputRef}
                type="text"
                id="answer"
                name="answer"
                value={answer}
                onChange={handleAnswerChange}
                className={getInputClasses(answerValidation, 
                  "w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-colors"
                )}
                placeholder="Ingrese su respuesta..."
                disabled={isLoading}
                autoComplete="off"
                required
              />
              
              {/* Mensaje de validación */}
              {answerValidation.message && (
                <p className={getValidationMessageClasses(answerValidation, "text-sm mt-1")}>
                  {answerValidation.message}
                </p>
              )}
            </div>

            {/* Mensaje general */}
            {message && (
              <div className={`p-3 rounded-lg text-sm ${
                messageType === 'success' 
                  ? 'bg-green-50 text-green-700 border border-green-200'
                  : messageType === 'error' 
                  ? 'bg-red-50 text-red-700 border border-red-200' 
                  : 'bg-blue-50 text-blue-700 border border-blue-200'
              }`}>
                {message}
              </div>
            )}

            {/* Botón Submit */}
            <button
              type="submit"
              disabled={isLoading || !answerValidation.isValid}
              className={`w-full py-3 px-4 rounded-lg font-medium transition-all duration-200 ${
                isLoading || !answerValidation.isValid
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : 'bg-indigo-600 text-white hover:bg-indigo-700 active:transform active:scale-95'
              }`}
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                  Verificando...
                </div>
              ) : (
                'Verificar Respuesta'
              )}
            </button>

          </form>
        )}

        {/* Mensaje de éxito */}
        {isSubmitted && messageType === 'success' && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-center justify-center">
              <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-green-600 mr-3"></div>
              <span className="text-green-700 font-medium">Redirigiendo...</span>
            </div>
          </div>
        )}

        {/* Botones de navegación */}
        <div className="mt-8 space-y-3">
          
          {/* Email para referencia */}
          <div className="text-center p-3 bg-gray-50 rounded-lg">
            <span className="text-xs text-gray-500">Verificando para:</span>
            <p className="text-sm font-medium text-gray-700">{email}</p>
          </div>

          {/* Botón nuevo reset */}
          {(isBlocked || !isSubmitted) && onNewReset && (
            <button
              onClick={onNewReset}
              className="w-full py-2 px-4 text-purple-600 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors text-sm"
              disabled={isLoading}
            >
              🔄 Solicitar nuevo reset
            </button>
          )}

          {/* Botón volver */}
          {onBack && !isSubmitted && (
            <button
              onClick={onBack}
              className="w-full py-2 px-4 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors text-sm"
              disabled={isLoading}
            >
              ← Volver al formulario anterior
            </button>
          )}

          {/* Información adicional */}
          <div className="text-center pt-4 border-t border-gray-200">
            <p className="text-xs text-gray-500">
              Las respuestas son sensibles a mayúsculas/minúsculas
            </p>
          </div>

        </div>

      </div>
    </div>
  );
};

export default SecurityQuestionForm;
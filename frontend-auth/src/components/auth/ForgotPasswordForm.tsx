// src/components/auth/ForgotPasswordForm.tsx
// Formulario principal para solicitar reset de contraseña
// Incluye validación email, rate limiting UI y manejo de preguntas de seguridad

import React, { useState, useEffect, useRef } from 'react';
import { passwordResetService, PasswordResetError } from '../../services/passwordResetService';
import { validateEmail, getInputClasses, getValidationMessageClasses } from '../../utils/passwordValidation';
import type { FieldValidation, PasswordResetErrorType } from '../../types/passwordReset';
import { PASSWORD_RESET_CONFIG } from '../../types/passwordReset';

interface ForgotPasswordFormProps {
  onSuccess?: (email: string, requiresSecurityQuestion: boolean, securityQuestion?: any) => void;
  onError?: (error: PasswordResetError) => void;
  onBack?: () => void;
  onLoginRedirect?: () => void;
  initialEmail?: string;
  autoFocus?: boolean;
}

const ForgotPasswordForm: React.FC<ForgotPasswordFormProps> = ({
  onSuccess,
  onError,
  onBack,
  onLoginRedirect,
  initialEmail = '',
  autoFocus = true
}) => {
  // Referencias para DOM
  const emailInputRef = useRef<HTMLInputElement>(null);

  // Estados del formulario
  const [email, setEmail] = useState(initialEmail);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error' | 'warning' | ''>('');

  // Estados de validación
  const [emailValidation, setEmailValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });

  // Estados de rate limiting
  const [isRateLimited, setIsRateLimited] = useState(false);
  const [timeUntilNextRequest, setTimeUntilNextRequest] = useState(0);
  const [lastSubmissionTime, setLastSubmissionTime] = useState<number | null>(null);

  // ========================================
  // EFECTOS Y INICIALIZACIÓN
  // ========================================

  // Auto-focus en el campo email
  useEffect(() => {
    if (autoFocus && emailInputRef.current) {
      emailInputRef.current.focus();
    }
  }, [autoFocus]);

  // Validar email inicial si se proporciona
  useEffect(() => {
    if (initialEmail) {
      const validation = validateEmail(initialEmail);
      setEmailValidation(validation);
    }
  }, [initialEmail]);

  // Timer para rate limiting
  useEffect(() => {
    if (timeUntilNextRequest > 0) {
      const timer = setInterval(() => {
        setTimeUntilNextRequest(prev => {
          const newTime = prev - 1;
          if (newTime <= 0) {
            setIsRateLimited(false);
            return 0;
          }
          return newTime;
        });
      }, 1000);

      return () => clearInterval(timer);
    }
  }, [timeUntilNextRequest]);

  // ========================================
  // HANDLERS Y VALIDACIONES
  // ========================================

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setEmail(value);
    
    // Validar en tiempo real
    if (value.trim()) {
      const validation = validateEmail(value);
      setEmailValidation(validation);
    } else {
      setEmailValidation({ isValid: false, message: '', type: '' });
    }
    
    // Limpiar mensajes cuando el usuario empiece a escribir
    if (message && !isLoading) {
      setMessage('');
      setMessageType('');
    }
  };

  const recordRateLimit = () => {
    const now = Date.now();
    setLastSubmissionTime(now);
    setIsRateLimited(true);
    setTimeUntilNextRequest(PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW / 1000);
    
    // Persistir en localStorage para mantener límite entre recargas
    localStorage.setItem('lastPasswordResetRequest', now.toString());
  };

  const checkRateLimit = () => {
    const lastRequestTime = localStorage.getItem('lastPasswordResetRequest');
    if (!lastRequestTime) return false;

    const timeDiff = Date.now() - parseInt(lastRequestTime);
    if (timeDiff < PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW) {
      const remaining = Math.ceil((PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW - timeDiff) / 1000);
      setTimeUntilNextRequest(remaining);
      setIsRateLimited(true);
      return true;
    }

    return false;
  };

  // Verificar rate limit al cargar el componente
  useEffect(() => {
    checkRateLimit();
  }, []);

  /**
   * Maneja envío del formulario
   */
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Prevenir doble envío
    if (isLoading || isRateLimited) return;

    setIsLoading(true);
    setMessage('');
    setMessageType('');

    try {
      // Validar email final
      const validation = validateEmail(email);
      setEmailValidation(validation);

      if (!validation.isValid) {
        setMessage(validation.message);
        setMessageType('error');
        return;
      }

      // Realizar solicitud de reset
      const result = await passwordResetService.requestPasswordReset({ email: email.trim() });
      
      // Registrar solicitud para rate limiting
      recordRateLimit();
      
      // Marcar como enviado
      setIsSubmitted(true);
      
      // ✅ CORRECCIÓN FINAL: Verificar si hay preguntas de seguridad y manejar el flujo correctamente
      console.log('✅ Resultado del servicio corregido:', result);
      console.log('✅ requiresSecurityQuestion:', result.requiresSecurityQuestion);
      console.log('✅ securityQuestion:', result.securityQuestion);
      
      if (result.requiresSecurityQuestion && result.securityQuestion) {
        // Usuario tiene preguntas de seguridad configuradas
        console.log('✅ Usuario tiene preguntas de seguridad - transicionando a security-question');
        setMessage('✅ Debe responder una pregunta de seguridad para continuar');
        setMessageType('success');

        // Callback de éxito con datos de pregunta de seguridad
        if (onSuccess) {
          console.log('✅ Llamando onSuccess con datos de pregunta de seguridad...');
          onSuccess(
            email.trim(),
            true, // requiresSecurityQuestion
            {
              questionId: result.securityQuestion.questionId,
              questionText: result.securityQuestion.questionText,
              attempts: result.securityQuestion.attempts || 0,
              maxAttempts: result.securityQuestion.maxAttempts || 3
            }
          );
        }
      } else {
        // Usuario no tiene preguntas de seguridad o flujo por email
        console.log('✅ Sin preguntas de seguridad - flujo por email');
        setMessage('✅ ' + (result.message || 'Si el email existe, recibirá instrucciones para restablecer su contraseña'));
        setMessageType('success');

        // Callback de éxito para flujo por email
        if (onSuccess) {
          onSuccess(
            email.trim(),
            false, // no requiresSecurityQuestion
            undefined // no securityQuestion
          );
        }
      }

    } catch (error: any) {
      console.error('Error en solicitud de reset:', error);
      
      let errorMessage = 'Error al procesar solicitud';
      let errorType: PasswordResetErrorType = 'server_error';

      if (error instanceof PasswordResetError) {
        errorMessage = error.message;
        errorType = error.type;
      } else if (error.response) {
        // Error HTTP del backend
        const { status, data } = error.response;
        
        switch (status) {
          case 403:
            errorMessage = 'Solo usuarios con email @gamc.gov.bo pueden solicitar reset de contraseña';
            errorType = 'email_not_institutional';
            break;
          case 429:
            errorMessage = 'Demasiadas solicitudes. Espere 5 minutos antes de intentar nuevamente';
            errorType = 'rate_limit_exceeded';
            recordRateLimit(); // Activar rate limit local también
            break;
          default:
            errorMessage = data?.message || data?.error || 'Error del servidor';
        }
      } else if (error.code === 'NETWORK_ERROR' || !error.response) {
        errorMessage = 'Error de conexión. Verifique su internet e intente nuevamente.';
        errorType = 'network_error';
      }

      setMessage('❌ ' + errorMessage);
      setMessageType('error');

      // Callback de error
      if (onError) {
        onError(new PasswordResetError(errorType, errorMessage));
      }

    } finally {
      setIsLoading(false);
    }
  };

  // ========================================
  // CAMPOS COMPUTADOS
  // ========================================

  const isFormValid = emailValidation.isValid && !isRateLimited;
  const canSubmit = isFormValid && !isLoading && !isSubmitted;

  const formatTimeRemaining = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    
    if (minutes > 0) {
      return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    }
    
    return `${remainingSeconds}s`;
  };

  // ========================================
  // COMPONENTES AUXILIARES
  // ========================================

  const FieldValidationMessage: React.FC<{ validation: FieldValidation }> = ({ validation }) => {
    if (!validation.message) return null;

    return (
      <div className={getValidationMessageClasses(validation)}>
        {validation.message}
      </div>
    );
  };

  const RateLimitMessage: React.FC = () => {
    if (!isRateLimited || timeUntilNextRequest <= 0) return null;

    return (
      <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
        <div className="flex items-center text-amber-800">
          <span className="text-amber-600 text-lg mr-2">⏱️</span>
          <div>
            <div className="font-semibold text-sm">Límite de solicitudes alcanzado</div>
            <div className="text-xs">
              Próxima solicitud disponible en: <span className="font-mono">{formatTimeRemaining(timeUntilNextRequest)}</span>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // ========================================
  // RENDERIZADO PRINCIPAL
  // ========================================

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-purple-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl">🔑</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">¿Olvidaste tu contraseña?</h1>
          <p className="text-gray-600 mt-2">
            Ingresa tu email institucional para recibir instrucciones de restablecimiento
          </p>
        </div>

        {/* Rate Limit Warning */}
        <RateLimitMessage />

        {/* Formulario */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Campo Email */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Email Institucional <span className="text-red-500">*</span>
            </label>
            <input
              ref={emailInputRef}
              type="email"
              value={email}
              onChange={handleEmailChange}
              className={getInputClasses(emailValidation)}
              placeholder="usuario@gamc.gov.bo"
              disabled={isLoading || isRateLimited}
              required
            />
            <FieldValidationMessage validation={emailValidation} />
          </div>

          {/* Mensaje de estado */}
          {message && (
            <div className={`p-3 rounded-lg text-sm border ${
              messageType === 'success' 
                ? 'bg-green-50 text-green-800 border-green-200' 
                : messageType === 'error'
                  ? 'bg-red-50 text-red-800 border-red-200'
                  : messageType === 'warning'
                    ? 'bg-amber-50 text-amber-800 border-amber-200'
                    : 'bg-blue-50 text-blue-800 border-blue-200'
            }`}>
              {message}
            </div>
          )}

          {/* Botón de envío */}
          <button
            type="submit"
            disabled={!canSubmit}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              !canSubmit
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-purple-600 hover:bg-purple-700 text-white'
            }`}
          >
            {isLoading ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Enviando instrucciones...
              </div>
            ) : isRateLimited ? (
              `Esperar ${formatTimeRemaining(timeUntilNextRequest)}`
            ) : (
              'Enviar instrucciones'
            )}
          </button>
        </form>

        {/* Enlaces de navegación */}
        <div className="mt-6 text-center space-y-3">
          <button
            onClick={onBack}
            className="text-purple-600 hover:text-purple-700 text-sm font-medium transition-colors"
          >
            ← Volver al inicio de sesión
          </button>
          
          <div className="text-xs text-gray-500 border-t pt-3">
            ¿Necesita ayuda? Contacte al administrador del sistema
          </div>
        </div>
      </div>
    </div>
  );
};

export default ForgotPasswordForm;
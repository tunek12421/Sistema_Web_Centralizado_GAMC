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

  // Verificar rate limit al cargar
  useEffect(() => {
    checkRateLimit();
  }, []);

  // ========================================
  // HANDLERS Y VALIDACIONES
  // ========================================

  /**
   * Maneja cambios en el campo email
   */
  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newEmail = e.target.value;
    setEmail(newEmail);
    
    // Validar en tiempo real
    const validation = validateEmail(newEmail);
    setEmailValidation(validation);
    
    // Limpiar mensajes previos
    if (message) {
      setMessage('');
      setMessageType('');
    }
  };

  /**
   * Verifica rate limit desde localStorage
   */
  const checkRateLimit = () => {
    try {
      const lastRequestStr = localStorage.getItem('gamc_last_reset_request');
      if (lastRequestStr) {
        const lastRequestTime = parseInt(lastRequestStr);
        const now = Date.now();
        const timeSinceLastRequest = now - lastRequestTime;
        const cooldownTime = PASSWORD_RESET_CONFIG.RATE_LIMIT_WINDOW;

        if (timeSinceLastRequest < cooldownTime) {
          const remainingTime = Math.ceil((cooldownTime - timeSinceLastRequest) / 1000);
          setIsRateLimited(true);
          setTimeUntilNextRequest(remainingTime);
          setLastSubmissionTime(lastRequestTime);
        }
      }
    } catch (error) {
      console.error('Error checking rate limit:', error);
    }
  };

  /**
   * Registra nueva solicitud para rate limiting
   */
  const recordRateLimit = () => {
    try {
      const now = Date.now();
      localStorage.setItem('gamc_last_reset_request', now.toString());
      setLastSubmissionTime(now);
      setIsRateLimited(true);
      setTimeUntilNextRequest(300); // 5 minutos = 300 segundos
    } catch (error) {
      console.error('Error recording rate limit:', error);
    }
  };

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
      
      // Mostrar mensaje de éxito
      setMessage('✅ ' + (result.message || 'Si el email existe, recibirá instrucciones para restablecer su contraseña'));
      setMessageType('success');

      // Callback de éxito
      if (onSuccess) {
        onSuccess(
          email.trim(),
          result.requiresSecurityQuestion || false,
          result.securityQuestion
        );
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
        errorMessage = 'Error de conexión. Verifique su internet e intente nuevamente';
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

  /**
   * Formatea tiempo restante para UI
   */
  const formatTimeRemaining = (seconds: number): string => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    
    if (minutes > 0) {
      return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
    }
    return `${remainingSeconds}s`;
  };

  // ========================================
  // COMPONENTES DE UI
  // ========================================

  const RateLimitWarning = () => {
    if (!isRateLimited || timeUntilNextRequest <= 0) return null;

    return (
      <div className="mb-4 p-4 bg-amber-50 border border-amber-200 rounded-lg">
        <div className="flex items-start">
          <span className="text-amber-600 text-lg mr-3 mt-0.5">⏰</span>
          <div>
            <h4 className="text-amber-800 font-semibold text-sm mb-1">
              Esperando tiempo de seguridad
            </h4>
            <p className="text-amber-700 text-sm">
              Por seguridad, debe esperar <strong>{formatTimeRemaining(timeUntilNextRequest)}</strong> antes de solicitar otro reset.
            </p>
            {lastSubmissionTime && (
              <p className="text-amber-600 text-xs mt-1">
                Última solicitud: {new Date(lastSubmissionTime).toLocaleTimeString()}
              </p>
            )}
          </div>
        </div>
      </div>
    );
  };

  const SuccessMessage = () => {
    if (!isSubmitted || messageType !== 'success') return null;

    return (
      <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
        <div className="flex items-start">
          <span className="text-green-600 text-lg mr-3 mt-0.5">✅</span>
          <div>
            <h4 className="text-green-800 font-semibold text-sm mb-2">
              Solicitud enviada exitosamente
            </h4>
            <p className="text-green-700 text-sm mb-3">
              Si su email existe en nuestro sistema, recibirá instrucciones para restablecer su contraseña.
            </p>
            <div className="space-y-2">
              <div className="flex items-center justify-between bg-white p-3 rounded border">
                <span className="text-sm text-gray-600">Email:</span>
                <span className="text-sm font-medium text-gray-900">{email}</span>
              </div>
              <p className="text-green-600 text-xs">
                • Revise su bandeja de entrada y carpeta de spam<br/>
                • El enlace expira en 30 minutos<br/>
                • Solo emails @gamc.gov.bo son válidos
              </p>
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
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-purple-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">🔑</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">¿Olvidaste tu contraseña?</h1>
          <p className="text-gray-600 mt-2">
            Ingresa tu email institucional para recibir instrucciones de restablecimiento
          </p>
        </div>

        {/* Rate Limit Warning */}
        <RateLimitWarning />

        {/* Success Message */}
        <SuccessMessage />

        {/* Formulario */}
        {!isSubmitted && (
          <form onSubmit={handleSubmit} className="space-y-6">
            
            {/* Campo Email */}
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                Email Institucional
              </label>
              <input
                ref={emailInputRef}
                type="email"
                id="email"
                name="email"
                value={email}
                onChange={handleEmailChange}
                className={getInputClasses(emailValidation, 
                  "w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-colors"
                )}
                placeholder="usuario@gamc.gov.bo"
                disabled={isLoading || isRateLimited}
                autoComplete="email"
                required
              />
              
              {/* Mensaje de validación */}
              {emailValidation.message && (
                <p className={getValidationMessageClasses(emailValidation)}>
                  {emailValidation.message}
                </p>
              )}
            </div>

            {/* Mensaje general */}
            {message && messageType !== 'success' && (
              <div className={`p-3 rounded-lg text-sm ${
                messageType === 'error' 
                  ? 'bg-red-50 text-red-700 border border-red-200' 
                  : 'bg-blue-50 text-blue-700 border border-blue-200'
              }`}>
                {message}
              </div>
            )}

            {/* Botón Submit */}
            <button
              type="submit"
              disabled={isLoading || isRateLimited || !emailValidation.isValid}
              className={`w-full py-3 px-4 rounded-lg font-medium transition-all duration-200 ${
                isLoading || isRateLimited || !emailValidation.isValid
                  ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                  : 'bg-purple-600 text-white hover:bg-purple-700 active:transform active:scale-95'
              }`}
            >
              {isLoading ? (
                <div className="flex items-center justify-center">
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                  Procesando...
                </div>
              ) : isRateLimited ? (
                `Espere ${formatTimeRemaining(timeUntilNextRequest)}`
              ) : (
                'Enviar instrucciones'
              )}
            </button>

          </form>
        )}

        {/* Botones de navegación */}
        <div className="mt-8 space-y-3">
          
          {/* Botón volver al login */}
          {onLoginRedirect && (
            <button
              onClick={onLoginRedirect}
              className="w-full py-2 px-4 text-purple-600 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors text-sm"
              disabled={isLoading}
            >
              ← Volver al inicio de sesión
            </button>
          )}

          {/* Botón volver (genérico) */}
          {onBack && !onLoginRedirect && (
            <button
              onClick={onBack}
              className="w-full py-2 px-4 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors text-sm"
              disabled={isLoading}
            >
              ← Volver
            </button>
          )}

          {/* Información adicional */}
          <div className="text-center pt-4 border-t border-gray-200">
            <p className="text-xs text-gray-500">
              ¿Necesita ayuda? Contacte al administrador del sistema
            </p>
          </div>

        </div>

      </div>
    </div>
  );
};

export default ForgotPasswordForm;
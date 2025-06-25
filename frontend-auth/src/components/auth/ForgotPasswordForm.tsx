// src/components/auth/ForgotPasswordForm.tsx
// Componente para solicitar reset de contraseña
// Integrado con el sistema de validación y estilos existentes

import React, { useState, useEffect } from 'react';
import { passwordResetService, PasswordResetError } from '../../services/passwordResetService';
import { 
  validateEmail, 
  getInputClasses, 
  getValidationMessageClasses,
  canMakePasswordResetRequest,
  getTimeUntilNextRequest,
  recordPasswordResetRequest,
  formatTimeRemaining
} from '../../utils/passwordValidation';
import { 
  FieldValidation, 
  PasswordResetErrorType,
  PASSWORD_RESET_MESSAGES 
} from '../../types/passwordReset';

interface ForgotPasswordFormProps {
  onSuccess?: (email: string) => void;
  onBack?: () => void;
  onError?: (error: PasswordResetError) => void;
}

const ForgotPasswordForm: React.FC<ForgotPasswordFormProps> = ({ 
  onSuccess, 
  onBack, 
  onError 
}) => {
  // Estados del formulario
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error' | 'info' | ''>('');

  // Estados de validación
  const [emailValidation, setEmailValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });

  // Rate limiting
  const [canSubmit, setCanSubmit] = useState(true);
  const [timeRemaining, setTimeRemaining] = useState(0);

  // ========================================
  // EFECTOS Y VALIDACIONES
  // ========================================

  // Validación en tiempo real del email
  useEffect(() => {
    if (email.trim()) {
      const validation = validateEmail(email);
      setEmailValidation(validation);
    } else {
      setEmailValidation({ isValid: false, message: '', type: '' });
    }
  }, [email]);

  // Timer para rate limiting
  useEffect(() => {
    const updateRateLimit = () => {
      const canRequest = canMakePasswordResetRequest();
      const remaining = getTimeUntilNextRequest();
      
      setCanSubmit(canRequest);
      setTimeRemaining(remaining);
    };

    // Actualizar inmediatamente
    updateRateLimit();

    // Actualizar cada segundo si hay rate limit activo
    const interval = setInterval(updateRateLimit, 1000);

    return () => clearInterval(interval);
  }, []);

  // ========================================
  // MANEJADORES DE EVENTOS
  // ========================================

  const handleEmailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newEmail = e.target.value;
    setEmail(newEmail);
    
    // Limpiar mensaje cuando el usuario empiece a escribir
    if (message && !isLoading) {
      setMessage('');
      setMessageType('');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validaciones pre-envío
    if (!canSubmit) {
      setMessage(`⏰ Debe esperar ${formatTimeRemaining(timeRemaining)} antes de hacer otra solicitud`);
      setMessageType('error');
      return;
    }

    if (!emailValidation.isValid) {
      setMessage('❌ Complete el email correctamente antes de continuar');
      setMessageType('error');
      return;
    }

    setIsLoading(true);
    setMessage('');
    setMessageType('');

    try {
      // Validación adicional del servicio
      const validation = passwordResetService.validateEmailForReset(email);
      if (!validation.isValid) {
        throw new PasswordResetError(
          PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL,
          validation.error || 'Email inválido'
        );
      }

      // Enviar solicitud de reset
      const response = await passwordResetService.requestPasswordReset({ email: email.trim() });

      // Registrar solicitud para rate limiting
      recordPasswordResetRequest();
      passwordResetService.recordResetRequest();

      // Mostrar mensaje de éxito
      setMessage('✅ ' + response.data.message);
      setMessageType('success');
      setIsSubmitted(true);

      // Limpiar formulario
      setEmail('');
      setEmailValidation({ isValid: false, message: '', type: '' });

      // Callback de éxito
      if (onSuccess) {
        setTimeout(() => onSuccess(email.trim()), 2000);
      }

    } catch (error) {
      console.error('Error en solicitud de reset:', error);
      
      let errorMessage = 'Error inesperado';
      let errorType = PasswordResetErrorType.UNKNOWN_ERROR;

      if (error instanceof PasswordResetError) {
        errorMessage = error.getUserMessage();
        errorType = error.type;
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }

      // Mostrar mensaje específico según el tipo de error
      switch (errorType) {
        case PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL:
          setMessage('🏢 ' + errorMessage);
          break;
        case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
          setMessage('⏰ ' + errorMessage);
          break;
        case PasswordResetErrorType.NETWORK_ERROR:
          setMessage('🔌 ' + errorMessage);
          break;
        case PasswordResetErrorType.SERVER_ERROR:
          setMessage('⚠️ ' + errorMessage);
          break;
        default:
          setMessage('❌ ' + errorMessage);
      }

      setMessageType('error');

      // Callback de error
      if (onError && error instanceof PasswordResetError) {
        onError(error);
      }
    } finally {
      setIsLoading(false);
    }
  };

  // ========================================
  // COMPONENTES AUXILIARES
  // ========================================

  const FieldValidationMessage = ({ validation }: { validation: FieldValidation }) => {
    if (!validation || validation.type === '') return null;
    
    return (
      <div className={getValidationMessageClasses(validation)}>
        {validation.message}
      </div>
    );
  };

  const RateLimitMessage = () => {
    if (canSubmit || timeRemaining <= 0) return null;

    return (
      <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
        <div className="flex items-center text-amber-800">
          <span className="text-lg mr-2">⏰</span>
          <div>
            <p className="text-sm font-medium">Solicitud reciente detectada</p>
            <p className="text-xs">
              Podrá hacer otra solicitud en {formatTimeRemaining(timeRemaining)}
            </p>
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
            Ingresa tu email institucional y te enviaremos un enlace para restablecer tu contraseña
          </p>
        </div>

        {/* Rate Limit Warning */}
        <RateLimitMessage />

        {/* Success Message */}
        {isSubmitted && messageType === 'success' && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-start">
              <span className="text-green-600 text-xl mr-3 mt-0.5">📧</span>
              <div>
                <h3 className="text-green-800 font-semibold text-sm">¡Solicitud enviada!</h3>
                <p className="text-green-700 text-xs mt-1">
                  Si tu email existe en nuestro sistema, recibirás un enlace de reset en tu correo institucional en los próximos minutos.
                </p>
                <p className="text-green-600 text-xs mt-2 font-medium">
                  ✓ Revisa tu bandeja de entrada y carpeta de spam
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Formulario */}
        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Correo Electrónico Institucional <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              value={email}
              onChange={handleEmailChange}
              className={getInputClasses(emailValidation)}
              placeholder="usuario@gamc.gov.bo"
              disabled={isLoading}
              required
            />
            <FieldValidationMessage validation={emailValidation} />
            
            {/* Ayuda contextual */}
            {!email && (
              <p className="text-xs text-gray-500 mt-1">
                Solo emails @gamc.gov.bo pueden solicitar reset de contraseña
              </p>
            )}
          </div>

          {/* Mensaje de estado */}
          {message && (
            <div className={`p-3 rounded-lg text-sm border ${
              messageType === 'success' 
                ? 'bg-green-50 text-green-800 border-green-200' 
                : messageType === 'error'
                  ? 'bg-red-50 text-red-800 border-red-200'
                  : 'bg-blue-50 text-blue-800 border-blue-200'
            }`}>
              {message}
            </div>
          )}

          {/* Botón de envío */}
          <button
            type="submit"
            disabled={isLoading || !emailValidation.isValid || !canSubmit}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              isLoading || !emailValidation.isValid || !canSubmit
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-purple-600 hover:bg-purple-700 text-white'
            }`}
          >
            {isLoading ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Enviando solicitud...
              </div>
            ) : !canSubmit ? (
              `⏰ Espere ${formatTimeRemaining(timeRemaining)}`
            ) : (
              '📧 Enviar enlace de reset'
            )}
          </button>

          {/* Información adicional */}
          {!isSubmitted && (
            <div className="text-xs text-gray-500 space-y-1">
              <p>• El enlace de reset expira en 30 minutos</p>
              <p>• Solo puede solicitar un reset cada 5 minutos</p>
              <p>• Verifique su bandeja de entrada y carpeta de spam</p>
            </div>
          )}

          {/* Botón de volver */}
          {onBack && (
            <button
              type="button"
              onClick={onBack}
              disabled={isLoading}
              className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors disabled:opacity-50"
            >
              ← Volver al login
            </button>
          )}
        </form>

        {/* Footer con ayuda */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p className="mb-2">¿Problemas para acceder?</p>
            <p>Contacte al administrador del sistema:</p>
            <p className="font-medium text-gray-700">soporte@gamc.gov.bo</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ForgotPasswordForm;
// src/components/auth/ResetPasswordForm.tsx
// Componente para confirmar reset de contraseña con token + nueva password
// Integrado con validación de tokens y contraseñas del backend

import React, { useState, useEffect } from 'react';
import { passwordResetService, PasswordResetError } from '../../services/passwordResetService';
import { 
  validatePassword, 
  validatePasswordConfirmation,
  analyzePasswordStrength,
  getInputClasses, 
  getValidationMessageClasses,
  getPasswordStrengthColor,
  getPasswordStrengthText
} from '../../utils/passwordValidation';
import {
  validateTokenFormat,
  validateTokenState,
  getTokenTimeRemaining,
  formatTokenTimeRemaining,
  isTokenNearExpiration
} from '../../utils/tokenValidation';
import { 
  FieldValidation, 
  PasswordResetErrorType,
  PASSWORD_RESET_MESSAGES 
} from '../../types/passwordReset';

interface ResetPasswordFormProps {
  token: string;
  onSuccess?: (message: string) => void;
  onError?: (error: PasswordResetError) => void;
  onTokenExpired?: () => void;
  onBack?: () => void;
}

const ResetPasswordForm: React.FC<ResetPasswordFormProps> = ({ 
  token, 
  onSuccess, 
  onError, 
  onTokenExpired,
  onBack 
}) => {
  // Estados del formulario
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState<'success' | 'error' | 'warning' | ''>('');

  // Estados de validación
  const [tokenValidation, setTokenValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });
  const [passwordValidation, setPasswordValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });
  const [confirmValidation, setConfirmValidation] = useState<FieldValidation>({
    isValid: false,
    message: '',
    type: ''
  });

  // Estados de fortaleza de contraseña
  const [passwordStrength, setPasswordStrength] = useState<any>(null);

  // Estados de token
  const [tokenTimeRemaining, setTokenTimeRemaining] = useState<number>(0);
  const [isTokenNearExp, setIsTokenNearExp] = useState(false);

  // ========================================
  // EFECTOS Y VALIDACIONES
  // ========================================

  // Validación inicial del token
  useEffect(() => {
    if (token) {
      const validation = validateTokenFormat(token);
      setTokenValidation(validation);
      
      if (!validation.isValid) {
        setMessage(`🔗 ${validation.message}`);
        setMessageType('error');
      }
    } else {
      setTokenValidation({
        isValid: false,
        message: PASSWORD_RESET_MESSAGES.TOKEN_REQUIRED,
        type: 'error'
      });
      setMessage('🔗 Token de reset requerido');
      setMessageType('error');
    }
  }, [token]);

  // Validación en tiempo real de la nueva contraseña
  useEffect(() => {
    if (newPassword) {
      const validation = validatePassword(newPassword);
      const strength = analyzePasswordStrength(newPassword);
      
      setPasswordValidation(validation);
      setPasswordStrength(strength);
    } else {
      setPasswordValidation({ isValid: false, message: '', type: '' });
      setPasswordStrength(null);
    }
  }, [newPassword]);

  // Validación en tiempo real de confirmación
  useEffect(() => {
    if (confirmPassword) {
      const validation = validatePasswordConfirmation(newPassword, confirmPassword);
      setConfirmValidation(validation);
    } else {
      setConfirmValidation({ isValid: false, message: '', type: '' });
    }
  }, [newPassword, confirmPassword]);

  // Timer para mostrar tiempo restante del token (simulado - en producción vendría del backend)
  useEffect(() => {
    // En un caso real, esto vendría de la respuesta del backend
    // Por ahora simulamos 30 minutos desde que se carga el componente
    const startTime = Date.now();
    const expirationTime = startTime + (30 * 60 * 1000); // 30 minutos

    const updateTimer = () => {
      const remaining = Math.max(0, Math.ceil((expirationTime - Date.now()) / 1000));
      setTokenTimeRemaining(remaining);
      setIsTokenNearExp(remaining > 0 && remaining <= (5 * 60)); // Menos de 5 minutos

      if (remaining === 0 && onTokenExpired) {
        onTokenExpired();
      }
    };

    updateTimer();
    const interval = setInterval(updateTimer, 1000);

    return () => clearInterval(interval);
  }, [onTokenExpired]);

  // ========================================
  // MANEJADORES DE EVENTOS
  // ========================================

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setNewPassword(value);
    
    // Limpiar mensaje cuando el usuario empiece a escribir
    if (message && !isLoading) {
      setMessage('');
      setMessageType('');
    }
  };

  const handleConfirmChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setConfirmPassword(value);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validaciones pre-envío
    if (!tokenValidation.isValid) {
      setMessage('🔗 Token de reset inválido');
      setMessageType('error');
      return;
    }

    if (!passwordValidation.isValid) {
      setMessage('🔒 La nueva contraseña no cumple con los requisitos');
      setMessageType('error');
      return;
    }

    if (!confirmValidation.isValid) {
      setMessage('🔒 Las contraseñas no coinciden');
      setMessageType('error');
      return;
    }

    if (tokenTimeRemaining <= 0) {
      setMessage('⏰ El token ha expirado. Solicite un nuevo reset');
      setMessageType('error');
      if (onTokenExpired) onTokenExpired();
      return;
    }

    setIsLoading(true);
    setMessage('');
    setMessageType('');

    try {
      // Validaciones adicionales del servicio
      const tokenValidationService = passwordResetService.validateTokenForReset(token);
      if (!tokenValidationService.isValid) {
        throw new PasswordResetError(
          PasswordResetErrorType.TOKEN_INVALID,
          tokenValidationService.error || 'Token inválido'
        );
      }

      const passwordValidationService = passwordResetService.validatePasswordForReset(newPassword);
      if (!passwordValidationService.isValid) {
        throw new PasswordResetError(
          PasswordResetErrorType.PASSWORD_WEAK,
          passwordValidationService.error || 'Contraseña inválida'
        );
      }

      // Enviar confirmación de reset
      const response = await passwordResetService.confirmPasswordReset({
        token: token.trim(),
        newPassword: newPassword
      });

      // Mostrar mensaje de éxito
      setMessage('🎉 ' + response.data.message);
      setMessageType('success');
      setIsSubmitted(true);

      // Limpiar formulario por seguridad
      setNewPassword('');
      setConfirmPassword('');
      setPasswordStrength(null);

      // Callback de éxito
      if (onSuccess) {
        setTimeout(() => onSuccess(response.data.message), 2000);
      }

    } catch (error) {
      console.error('Error en confirmación de reset:', error);
      
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
        case PasswordResetErrorType.TOKEN_EXPIRED:
          setMessage('⏰ ' + errorMessage);
          if (onTokenExpired) setTimeout(onTokenExpired, 1000);
          break;
        case PasswordResetErrorType.TOKEN_USED:
          setMessage('🔄 ' + errorMessage);
          break;
        case PasswordResetErrorType.TOKEN_INVALID:
          setMessage('🔗 ' + errorMessage);
          break;
        case PasswordResetErrorType.PASSWORD_WEAK:
          setMessage('🔒 ' + errorMessage);
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

  const PasswordStrengthIndicator = () => {
    if (!passwordStrength || !newPassword) return null;

    const { level, feedback } = passwordStrength;
    const color = getPasswordStrengthColor(level);
    const text = getPasswordStrengthText(level);

    return (
      <div className="mt-2 p-2 bg-gray-50 rounded border">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs font-medium text-gray-600">Fortaleza:</span>
          <span className={`text-xs font-semibold ${color}`}>{text}</span>
        </div>
        
        {/* Barra de progreso */}
        <div className="w-full bg-gray-200 rounded-full h-1.5 mb-2">
          <div 
            className={`h-1.5 rounded-full transition-all duration-300 ${
              level === 'very-strong' ? 'bg-green-600 w-full' :
              level === 'strong' ? 'bg-green-500 w-4/5' :
              level === 'medium' ? 'bg-yellow-500 w-3/5' :
              'bg-red-500 w-2/5'
            }`}
          />
        </div>

        {/* Feedback específico */}
        {feedback.length > 0 && (
          <div className="text-xs text-gray-600">
            <span>Mejoras: </span>
            <span>{feedback.join(', ')}</span>
          </div>
        )}
      </div>
    );
  };

  const TokenExpirationWarning = () => {
    if (tokenTimeRemaining <= 0) {
      return (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center text-red-800">
            <span className="text-lg mr-2">⏰</span>
            <div>
              <p className="text-sm font-medium">Token expirado</p>
              <p className="text-xs">Este enlace de reset ya no es válido</p>
            </div>
          </div>
        </div>
      );
    }

    if (isTokenNearExp) {
      return (
        <div className="mb-4 p-3 bg-amber-50 border border-amber-200 rounded-lg">
          <div className="flex items-center text-amber-800">
            <span className="text-lg mr-2">⚠️</span>
            <div>
              <p className="text-sm font-medium">Enlace próximo a expirar</p>
              <p className="text-xs">
                Tiempo restante: {formatTokenTimeRemaining(tokenTimeRemaining)}
              </p>
            </div>
          </div>
        </div>
      );
    }

    return null;
  };

  // ========================================
  // RENDER PRINCIPAL
  // ========================================

  const isFormValid = tokenValidation.isValid && 
                     passwordValidation.isValid && 
                     confirmValidation.isValid &&
                     tokenTimeRemaining > 0;

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-green-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">🔄</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Restablecer contraseña</h1>
          <p className="text-gray-600 mt-2">
            Ingresa tu nueva contraseña para completar el restablecimiento
          </p>
        </div>

        {/* Token Expiration Warning */}
        <TokenExpirationWarning />

        {/* Success Message */}
        {isSubmitted && messageType === 'success' && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <div className="flex items-start">
              <span className="text-green-600 text-xl mr-3 mt-0.5">🎉</span>
              <div>
                <h3 className="text-green-800 font-semibold text-sm">¡Contraseña cambiada!</h3>
                <p className="text-green-700 text-xs mt-1">
                  Tu contraseña ha sido restablecida exitosamente. Por seguridad, se han cerrado todas tus sesiones activas.
                </p>
                <p className="text-green-600 text-xs mt-2 font-medium">
                  ✓ Serás redirigido al login en unos momentos
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Formulario */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Nueva contraseña */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Nueva contraseña <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <input
                type={showPassword ? "text" : "password"}
                value={newPassword}
                onChange={handlePasswordChange}
                className={getInputClasses(passwordValidation)}
                placeholder="Mínimo 8 caracteres con @$!%*?&"
                disabled={isLoading || tokenTimeRemaining <= 0}
                required
                minLength={8}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                disabled={isLoading}
              >
                {showPassword ? '🙈' : '👁️'}
              </button>
            </div>
            <FieldValidationMessage validation={passwordValidation} />
            <PasswordStrengthIndicator />
          </div>

          {/* Confirmar contraseña */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Confirmar nueva contraseña <span className="text-red-500">*</span>
            </label>
            <input
              type={showPassword ? "text" : "password"}
              value={confirmPassword}
              onChange={handleConfirmChange}
              className={getInputClasses(confirmValidation)}
              placeholder="Repetir la nueva contraseña"
              disabled={isLoading || tokenTimeRemaining <= 0}
              required
            />
            <FieldValidationMessage validation={confirmValidation} />
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
            disabled={isLoading || !isFormValid}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              isLoading || !isFormValid
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-green-600 hover:bg-green-700 text-white'
            }`}
          >
            {isLoading ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Restableciendo contraseña...
              </div>
            ) : tokenTimeRemaining <= 0 ? (
              '⏰ Token expirado'
            ) : (
              '🔄 Restablecer contraseña'
            )}
          </button>

          {/* Información adicional */}
          {!isSubmitted && tokenTimeRemaining > 0 && (
            <div className="text-xs text-gray-500 space-y-1">
              <p>• La contraseña debe tener al menos 8 caracteres</p>
              <p>• Debe incluir mayúscula, minúscula, número y símbolo (@$!%*?&)</p>
              <p>• Al cambiar la contraseña se cerrarán todas las sesiones activas</p>
              {tokenTimeRemaining > 0 && (
                <p>• Este enlace expira en: {formatTokenTimeRemaining(tokenTimeRemaining)}</p>
              )}
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
              ← Solicitar nuevo reset
            </button>
          )}
        </form>

        {/* Footer con ayuda */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p className="mb-2">¿Problemas con el restablecimiento?</p>
            <p>Contacte al administrador del sistema:</p>
            <p className="font-medium text-gray-700">soporte@gamc.gov.bo</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResetPasswordForm;
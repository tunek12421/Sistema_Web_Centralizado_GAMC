// src/pages/ForgotPasswordPage.tsx
// Página completa para solicitar reset de contraseña
// Orquesta el hook usePasswordReset y el componente ForgotPasswordForm

import React, { useEffect, useState } from 'react';
import ForgotPasswordForm from '../components/auth/ForgotPasswordForm';
import { usePasswordReset } from '../hooks/usePasswordReset';
import { PasswordResetError, PasswordResetErrorType } from '../types/passwordReset';

interface ForgotPasswordPageProps {
  onBack?: () => void;
  onLoginRedirect?: () => void;
  onSuccess?: (email: string) => void;
}

const ForgotPasswordPage: React.FC<ForgotPasswordPageProps> = ({
  onBack,
  onLoginRedirect,
  onSuccess
}) => {
  // Estados locales para UI
  const [showSuccessView, setShowSuccessView] = useState(false);
  const [processedEmail, setProcessedEmail] = useState<string>('');
  const [hasError, setHasError] = useState(false);
  const [errorDetails, setErrorDetails] = useState<PasswordResetError | null>(null);

  // Hook personalizado para manejar lógica de reset
  const {
    state,
    formStates,
    requestReset,
    canSubmitRequest,
    timeUntilNextRequest
  } = usePasswordReset({
    autoRedirect: false, // En esta página no queremos auto-redirect
    enableRateLimitUI: true,
    onSuccess: (step) => {
      if (step === 'request') {
        setShowSuccessView(true);
        setHasError(false);
        
        // Callback al componente padre si está disponible
        if (onSuccess) {
          onSuccess(processedEmail);
        }
      }
    },
    onError: (error) => {
      console.error('Error en proceso de reset:', error);
      setHasError(true);
      setErrorDetails(error);
      setShowSuccessView(false);
    }
  });

  // ========================================
  // EFECTOS
  // ========================================

  // Detectar si ya hay un proceso de reset en curso
  useEffect(() => {
    if (state.step === 'email_sent') {
      setShowSuccessView(true);
      setProcessedEmail(state.email || '');
    } else if (state.step === 'error') {
      setHasError(true);
      setShowSuccessView(false);
    }
  }, [state.step, state.email]);

  // ========================================
  // MANEJADORES DE EVENTOS
  // ========================================

  const handleRequestReset = async (email: string) => {
    try {
      setProcessedEmail(email);
      setHasError(false);
      setErrorDetails(null);
      
      await requestReset(email);
      
      // El hook ya maneja el cambio de estado a través del callback onSuccess
    } catch (error) {
      // El error ya se maneja en el callback onError del hook
      console.error('Error en solicitud de reset:', error);
    }
  };

  const handleTryAgain = () => {
    setShowSuccessView(false);
    setHasError(false);
    setErrorDetails(null);
    setProcessedEmail('');
  };

  const handleGoToLogin = () => {
    if (onLoginRedirect) {
      onLoginRedirect();
    }
  };

  // ========================================
  // COMPONENTES AUXILIARES
  // ========================================

  const SuccessView = () => (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header de éxito */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-green-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">📧</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">¡Solicitud enviada!</h1>
          <p className="text-gray-600 mt-2">
            Revisa tu correo electrónico para continuar
          </p>
        </div>

        {/* Mensaje de confirmación */}
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-start">
            <span className="text-green-600 text-xl mr-3 mt-0.5">✅</span>
            <div>
              <h3 className="text-green-800 font-semibold text-sm mb-1">Solicitud procesada</h3>
              <p className="text-green-700 text-xs">
                {state.successMessage || 'Si el email existe y es válido, recibirás un enlace de reset en tu correo institucional.'}
              </p>
              {processedEmail && (
                <p className="text-green-600 text-xs mt-2 font-mono">
                  📧 {processedEmail}
                </p>
              )}
            </div>
          </div>
        </div>

        {/* Instrucciones */}
        <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-blue-800 font-semibold text-sm mb-2">Próximos pasos:</h3>
          <ol className="text-blue-700 text-xs space-y-1 list-decimal list-inside">
            <li>Revisa tu bandeja de entrada en {processedEmail}</li>
            <li>Busca también en tu carpeta de spam o correo no deseado</li>
            <li>Haz clic en el enlace del email para restablecer tu contraseña</li>
            <li>El enlace expira en 30 minutos por seguridad</li>
          </ol>
        </div>

        {/* Rate limit info */}
        {timeUntilNextRequest > 0 && (
          <div className="mb-6 p-3 bg-amber-50 border border-amber-200 rounded-lg">
            <div className="flex items-center text-amber-800">
              <span className="text-lg mr-2">⏰</span>
              <div>
                <p className="text-sm font-medium">Límite de solicitudes activo</p>
                <p className="text-xs">
                  Próxima solicitud disponible en: {Math.ceil(timeUntilNextRequest / 60)} minuto(s)
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Acciones */}
        <div className="space-y-3">
          <button
            onClick={handleGoToLogin}
            className="w-full py-3 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold"
          >
            🔐 Ir al login
          </button>

          <button
            onClick={handleTryAgain}
            disabled={timeUntilNextRequest > 0}
            className={`w-full py-2 px-4 rounded-lg transition-colors text-sm ${
              timeUntilNextRequest > 0
                ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            {timeUntilNextRequest > 0 
              ? `⏰ Espere ${Math.ceil(timeUntilNextRequest / 60)} min para nueva solicitud`
              : '🔄 Solicitar reset para otra cuenta'
            }
          </button>
        </div>

        {/* Footer con soporte */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p className="mb-2">¿No recibiste el email?</p>
            <div className="space-y-1">
              <p>• Verifica que el email sea correcto</p>
              <p>• Revisa tu carpeta de spam</p>
              <p>• Contacta soporte: <span className="font-medium text-gray-700">soporte@gamc.gov.bo</span></p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  const ErrorView = () => (
    <div className="min-h-screen bg-gradient-to-br from-red-50 to-orange-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header de error */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-red-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">⚠️</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Error en la solicitud</h1>
          <p className="text-gray-600 mt-2">
            No se pudo procesar tu solicitud de reset
          </p>
        </div>

        {/* Mensaje de error específico */}
        <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-start">
            <span className="text-red-600 text-xl mr-3 mt-0.5">❌</span>
            <div>
              <h3 className="text-red-800 font-semibold text-sm mb-1">
                {getErrorTitle(errorDetails?.type)}
              </h3>
              <p className="text-red-700 text-xs">
                {errorDetails?.getUserMessage() || state.errorMessage || 'Error inesperado'}
              </p>
            </div>
          </div>
        </div>

        {/* Soluciones sugeridas */}
        <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-blue-800 font-semibold text-sm mb-2">Posibles soluciones:</h3>
          <ul className="text-blue-700 text-xs space-y-1">
            {getErrorSolutions(errorDetails?.type).map((solution, index) => (
              <li key={index}>• {solution}</li>
            ))}
          </ul>
        </div>

        {/* Acciones de recuperación */}
        <div className="space-y-3">
          <button
            onClick={handleTryAgain}
            disabled={!canSubmitRequest && errorDetails?.type === PasswordResetErrorType.RATE_LIMIT_EXCEEDED}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              !canSubmitRequest && errorDetails?.type === PasswordResetErrorType.RATE_LIMIT_EXCEEDED
                ? 'bg-gray-400 cursor-not-allowed text-white'
                : 'bg-purple-600 hover:bg-purple-700 text-white'
            }`}
          >
            {!canSubmitRequest && errorDetails?.type === PasswordResetErrorType.RATE_LIMIT_EXCEEDED
              ? `⏰ Espere ${Math.ceil(timeUntilNextRequest / 60)} min`
              : '🔄 Intentar nuevamente'
            }
          </button>

          <button
            onClick={handleGoToLogin}
            className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
          >
            ← Volver al login
          </button>
        </div>

        {/* Footer con soporte */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p className="mb-2">¿Persiste el problema?</p>
            <p>Contacte al soporte técnico:</p>
            <p className="font-medium text-gray-700">soporte@gamc.gov.bo</p>
          </div>
        </div>
      </div>
    </div>
  );

  // ========================================
  // FUNCIONES AUXILIARES
  // ========================================

  const getErrorTitle = (errorType?: PasswordResetErrorType): string => {
    switch (errorType) {
      case PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL:
        return 'Email no autorizado';
      case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
        return 'Demasiadas solicitudes';
      case PasswordResetErrorType.NETWORK_ERROR:
        return 'Error de conexión';
      case PasswordResetErrorType.SERVER_ERROR:
        return 'Error del servidor';
      default:
        return 'Error inesperado';
    }
  };

  const getErrorSolutions = (errorType?: PasswordResetErrorType): string[] => {
    switch (errorType) {
      case PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL:
        return [
          'Use únicamente emails con dominio @gamc.gov.bo',
          'Verifique que escribió correctamente su email institucional',
          'Contacte al administrador si no tiene email institucional'
        ];
      case PasswordResetErrorType.RATE_LIMIT_EXCEEDED:
        return [
          'Espere 5 minutos antes de hacer otra solicitud',
          'Revise su email mientras tanto por si ya recibió el enlace',
          'El límite es por seguridad del sistema'
        ];
      case PasswordResetErrorType.NETWORK_ERROR:
        return [
          'Verifique su conexión a internet',
          'Compruebe que el servidor esté disponible',
          'Intente nuevamente en unos momentos'
        ];
      case PasswordResetErrorType.SERVER_ERROR:
        return [
          'El problema es temporal del servidor',
          'Intente nuevamente en unos minutos',
          'Contacte soporte si persiste el error'
        ];
      default:
        return [
          'Intente nuevamente en unos momentos',
          'Verifique su conexión a internet',
          'Contacte soporte técnico si persiste'
        ];
    }
  };

  // ========================================
  // RENDER PRINCIPAL
  // ========================================

  // Mostrar vista de éxito si la solicitud fue exitosa
  if (showSuccessView) {
    return <SuccessView />;
  }

  // Mostrar vista de error si hay un error persistente
  if (hasError && errorDetails) {
    return <ErrorView />;
  }

  // Mostrar formulario principal
  return (
    <ForgotPasswordForm
      onSuccess={handleRequestReset}
      onBack={onBack}
      onError={(error) => {
        setHasError(true);
        setErrorDetails(error);
      }}
    />
  );
};

export default ForgotPasswordPage;
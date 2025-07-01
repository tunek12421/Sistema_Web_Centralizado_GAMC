// src/pages/ResetPasswordPage.tsx
// Página completa para confirmar reset de contraseña con token
// Maneja extracción de token desde URL y todo el flujo de confirmación

import React, { useEffect, useState } from 'react';
import ResetPasswordForm from '../components/auth/ResetPasswordForm';
import PasswordResetSuccess from '../components/auth/PasswordResetSuccess';
import { usePasswordReset } from '../hooks/usePasswordReset';
import { extractTokenFromURL, validateTokenFormat } from '../utils/tokenValidation';
import { PasswordResetError, PasswordResetErrorType } from '../types/passwordReset';

interface ResetPasswordPageProps {
  onLoginRedirect?: () => void;
  onNewReset?: () => void;
  onBack?: () => void;
  token?: string; // Token puede venir como prop o extraerse de URL
}

const ResetPasswordPage: React.FC<ResetPasswordPageProps> = ({
  onLoginRedirect,
  onNewReset,
  onBack,
  token: propToken
}) => {
  // Estados locales para UI
  const [currentToken, setCurrentToken] = useState<string>('');
  const [isTokenValid, setIsTokenValid] = useState<boolean>(false);
  const [showSuccessView, setShowSuccessView] = useState(false);
  const [hasError, setHasError] = useState(false);
  const [errorDetails, setErrorDetails] = useState<PasswordResetError | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [successMessage, setSuccessMessage] = useState<string>('');

  // Hook personalizado para manejar lógica de reset
  const {
    state,
    formStates,
    confirmReset,
    canSubmitConfirm,
    validateToken
  } = usePasswordReset({
    autoRedirect: false, // Manejamos el redirect manualmente en esta página
    enableRateLimitUI: false, // No necesario en la confirmación
    onSuccess: (step) => {
      if (step === 'confirm') {
        setShowSuccessView(true);
        setHasError(false);
        setSuccessMessage(state.successMessage || 'Contraseña restablecida exitosamente');
      }
    },
    onError: (error) => {
      console.error('Error en confirmación de reset:', error);
      setHasError(true);
      setErrorDetails(error);
      setShowSuccessView(false);
    }
  });

  // ========================================
  // EFECTOS DE INICIALIZACIÓN
  // ========================================

  // Inicialización: extraer y validar token
  useEffect(() => {
    const initializeToken = () => {
      setIsLoading(true);

      // Obtener token de prop o URL
      let token = propToken || extractTokenFromURL();

      if (!token) {
        // No hay token disponible
        setHasError(true);
        setErrorDetails(new PasswordResetError(
          PasswordResetErrorType.TOKEN_INVALID,
          'No se encontró token de reset. Verifique que copió correctamente la URL del email.'
        ));
        setIsLoading(false);
        return;
      }

      // Validar formato del token
      const validation = validateTokenFormat(token);
      if (!validation.isValid) {
        setHasError(true);
        setErrorDetails(new PasswordResetError(
          PasswordResetErrorType.TOKEN_INVALID,
          validation.message
        ));
        setIsLoading(false);
        return;
      }

      // Token válido
      setCurrentToken(token);
      setIsTokenValid(true);
      setHasError(false);
      setErrorDetails(null);
      setIsLoading(false);
    };

    initializeToken();
  }, [propToken]);

  // Efectos para detectar cambios de estado del hook
  useEffect(() => {
    if (state.step === 'success') {
      setShowSuccessView(true);
      setSuccessMessage(state.successMessage || 'Contraseña restablecida exitosamente');
    } else if (state.step === 'error') {
      setHasError(true);
      setShowSuccessView(false);
    }
  }, [state.step, state.successMessage]);

  // ========================================
  // MANEJADORES DE EVENTOS
  // ========================================

  const handleConfirmReset = async (token: string, newPassword: string) => {
    try {
      setHasError(false);
      setErrorDetails(null);
      
      await confirmReset(token, newPassword);
      
      // El hook ya maneja el cambio de estado a través del callback onSuccess
    } catch (error) {
      // El error ya se maneja en el callback onError del hook
      console.error('Error en confirmación de reset:', error);
    }
  };

  const handleTokenExpired = () => {
    setHasError(true);
    setErrorDetails(new PasswordResetError(
      PasswordResetErrorType.TOKEN_EXPIRED,
      'El enlace de reset ha expirado. Solicite un nuevo reset de contraseña.'
    ));
    setShowSuccessView(false);
  };

  const handleTryAgain = () => {
    if (onNewReset) {
      onNewReset();
    }
  };

  const handleGoToLogin = () => {
    if (onLoginRedirect) {
      onLoginRedirect();
    }
  };

  const handleBackToRequest = () => {
    if (onBack) {
      onBack();
    } else if (onNewReset) {
      onNewReset();
    }
  };

  // ========================================
  // COMPONENTES AUXILIARES
  // ========================================

  const LoadingView = () => (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-purple-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center">
          <div className="mx-auto h-16 w-16 bg-blue-600 rounded-full flex items-center justify-center mb-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-white"></div>
          </div>
          <h1 className="text-xl font-bold text-gray-900 mb-2">Verificando enlace...</h1>
          <p className="text-gray-600 text-sm">
            Validando el token de reset de contraseña
          </p>
        </div>
      </div>
    </div>
  );

  const InvalidTokenView = () => (
    <div className="min-h-screen bg-gradient-to-br from-red-50 to-orange-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header de error */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-red-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">🔗</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Enlace inválido</h1>
          <p className="text-gray-600 mt-2">
            El enlace de reset de contraseña no es válido
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
                {errorDetails?.getUserMessage() || 'Enlace de reset inválido o corrupto'}
              </p>
            </div>
          </div>
        </div>

        {/* Posibles causas */}
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
          <h3 className="text-amber-800 font-semibold text-sm mb-2">Posibles causas:</h3>
          <ul className="text-amber-700 text-xs space-y-1">
            {getErrorCauses(errorDetails?.type).map((cause, index) => (
              <li key={index}>• {cause}</li>
            ))}
          </ul>
        </div>

        {/* Soluciones */}
        <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-blue-800 font-semibold text-sm mb-2">¿Qué puedes hacer?</h3>
          <ul className="text-blue-700 text-xs space-y-1">
            <li>• Solicita un nuevo enlace de reset de contraseña</li>
            <li>• Verifica que copiaste correctamente la URL del email</li>
            <li>• Asegúrate de usar el enlace más reciente</li>
            <li>• Contacta soporte si el problema persiste</li>
          </ul>
        </div>

        {/* Acciones */}
        <div className="space-y-3">
          <button
            onClick={handleTryAgain}
            className="w-full py-3 px-4 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors font-semibold"
          >
            🔄 Solicitar nuevo reset
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
            <p className="mb-2">¿Necesitas ayuda?</p>
            <p>Contacta soporte: <span className="font-medium text-gray-700">soporte@gamc.gov.bo</span></p>
          </div>
        </div>
      </div>
    </div>
  );

  const ExpiredTokenView = () => (
    <div className="min-h-screen bg-gradient-to-br from-amber-50 to-orange-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header de expiración */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-amber-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">⏰</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Enlace expirado</h1>
          <p className="text-gray-600 mt-2">
            Este enlace de reset ya no es válido
          </p>
        </div>

        {/* Mensaje de expiración */}
        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
          <div className="flex items-start">
            <span className="text-amber-600 text-xl mr-3 mt-0.5">⏰</span>
            <div>
              <h3 className="text-amber-800 font-semibold text-sm mb-1">Tiempo agotado</h3>
              <p className="text-amber-700 text-xs">
                Los enlaces de reset expiran después de 30 minutos por seguridad. 
                Necesitas solicitar un nuevo enlace.
              </p>
            </div>
          </div>
        </div>

        {/* Información de seguridad */}
        <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <h3 className="text-blue-800 font-semibold text-sm mb-2">¿Por qué expiran los enlaces?</h3>
          <ul className="text-blue-700 text-xs space-y-1">
            <li>• Protegen tu cuenta contra acceso no autorizado</li>
            <li>• Reducen el riesgo si alguien más accede a tu email</li>
            <li>• Garantizan que uses enlaces recientes y seguros</li>
          </ul>
        </div>

        {/* Acciones */}
        <div className="space-y-3">
          <button
            onClick={handleTryAgain}
            className="w-full py-3 px-4 bg-amber-600 text-white rounded-lg hover:bg-amber-700 transition-colors font-semibold"
          >
            🔑 Solicitar nuevo enlace
          </button>

          <button
            onClick={handleGoToLogin}
            className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
          >
            ← Volver al login
          </button>
        </div>

        {/* Footer */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p>El nuevo enlace será válido por 30 minutos más</p>
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
      case PasswordResetErrorType.TOKEN_INVALID:
        return 'Token inválido';
      case PasswordResetErrorType.TOKEN_EXPIRED:
        return 'Token expirado';
      case PasswordResetErrorType.TOKEN_USED:
        return 'Token ya utilizado';
      default:
        return 'Enlace inválido';
    }
  };

  const getErrorCauses = (errorType?: PasswordResetErrorType): string[] => {
    switch (errorType) {
      case PasswordResetErrorType.TOKEN_INVALID:
        return [
          'El enlace fue copiado incorrectamente',
          'La URL está incompleta o dañada',
          'El token contiene caracteres inválidos'
        ];
      case PasswordResetErrorType.TOKEN_EXPIRED:
        return [
          'El enlace tiene más de 30 minutos',
          'Demasiado tiempo desde que se envió el email',
          'El token venció por seguridad'
        ];
      case PasswordResetErrorType.TOKEN_USED:
        return [
          'Este enlace ya fue utilizado anteriormente',
          'La contraseña ya fue cambiada con este token',
          'Cada enlace solo se puede usar una vez'
        ];
      default:
        return [
          'Problema con el formato del enlace',
          'Token corrupto o inválido',
          'Error en la transmisión del enlace'
        ];
    }
  };

  // ========================================
  // RENDER PRINCIPAL
  // ========================================

  // Mostrar loading mientras se valida el token
  if (isLoading) {
    return <LoadingView />;
  }

  // Mostrar vista de éxito si el reset fue completado
  if (showSuccessView) {
    return (
      <PasswordResetSuccess
        message={successMessage}
        autoRedirect={true}
        redirectDelay={3000}
        onLoginRedirect={handleGoToLogin}
        onNewReset={handleTryAgain}
      />
    );
  }

  // Mostrar vista de error específica para token expirado
  if (hasError && errorDetails?.type === PasswordResetErrorType.TOKEN_EXPIRED) {
    return <ExpiredTokenView />;
  }

  // Mostrar vista de error para token inválido o otros problemas
  if (hasError || !isTokenValid) {
    return <InvalidTokenView />;
  }

  // Mostrar formulario de reset con token válido
  return (
    <ResetPasswordForm
      token={currentToken}
      onSuccess={handleConfirmReset}
      onError={(error) => {
        setHasError(true);
        setErrorDetails(error);
      }}
      onTokenExpired={handleTokenExpired}
      onBack={handleBackToRequest}
    />
  );
};

export default ResetPasswordPage;
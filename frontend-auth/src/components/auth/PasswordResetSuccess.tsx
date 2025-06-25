// src/components/auth/PasswordResetSuccess.tsx
// Componente de éxito para mostrar confirmación de reset completado
// Con countdown automático y opciones de navegación manual

import React, { useState, useEffect } from 'react';
import { PASSWORD_RESET_CONFIG } from '../../types/passwordReset';

interface PasswordResetSuccessProps {
  message?: string;
  email?: string;
  autoRedirect?: boolean;
  redirectDelay?: number;
  onLoginRedirect?: () => void;
  onNewReset?: () => void;
  onDashboard?: () => void;
}

const PasswordResetSuccess: React.FC<PasswordResetSuccessProps> = ({
  message = "Su contraseña ha sido restablecida exitosamente",
  email,
  autoRedirect = true,
  redirectDelay = PASSWORD_RESET_CONFIG.AUTO_REDIRECT_DELAY,
  onLoginRedirect,
  onNewReset,
  onDashboard
}) => {
  // Estados del componente
  const [countdown, setCountdown] = useState(Math.ceil(redirectDelay / 1000));
  const [isRedirecting, setIsRedirecting] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  // ========================================
  // EFECTOS
  // ========================================

  // Countdown automático para redirect
  useEffect(() => {
    if (!autoRedirect) return;

    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          setIsRedirecting(true);
          if (onLoginRedirect) {
            setTimeout(onLoginRedirect, 500);
          }
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [autoRedirect, onLoginRedirect]);

  // ========================================
  // MANEJADORES DE EVENTOS
  // ========================================

  const handleLoginNow = () => {
    if (onLoginRedirect) {
      setIsRedirecting(true);
      onLoginRedirect();
    }
  };

  const handleNewReset = () => {
    if (onNewReset) {
      onNewReset();
    }
  };

  const handleGoToDashboard = () => {
    if (onDashboard) {
      setIsRedirecting(true);
      onDashboard();
    }
  };

  // ========================================
  // COMPONENTES AUXILIARES
  // ========================================

  const CountdownDisplay = () => {
    if (!autoRedirect || countdown <= 0) return null;

    return (
      <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
        <div className="flex items-center justify-center text-blue-800">
          <div className="text-center">
            <div className="flex items-center justify-center mb-2">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
              <span className="text-sm font-medium">Redirigiendo automáticamente</span>
            </div>
            <div className="text-2xl font-bold">{countdown}</div>
            <div className="text-xs">segundos</div>
          </div>
        </div>
      </div>
    );
  };

  const SecurityNotice = () => (
    <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-lg">
      <div className="flex items-start">
        <span className="text-amber-600 text-lg mr-3 mt-0.5">🔒</span>
        <div>
          <h3 className="text-amber-800 font-semibold text-sm mb-1">Medidas de seguridad aplicadas</h3>
          <ul className="text-amber-700 text-xs space-y-1">
            <li>• Todas sus sesiones activas han sido cerradas</li>
            <li>• Deberá iniciar sesión con su nueva contraseña</li>
            <li>• El enlace de reset ya no es válido</li>
            <li>• Su actividad ha sido registrada por seguridad</li>
          </ul>
        </div>
      </div>
    </div>
  );

  const AdditionalOptions = () => (
    <div className="space-y-3">
      <div className="text-center">
        <button
          onClick={() => setShowDetails(!showDetails)}
          className="text-sm text-gray-500 hover:text-gray-700 transition-colors"
        >
          {showDetails ? '▼ Ocultar detalles' : '▶ Ver más opciones'}
        </button>
      </div>

      {showDetails && (
        <div className="space-y-3 pt-4 border-t border-gray-200">
          {/* Opción: Nuevo reset para otra cuenta */}
          <button
            onClick={handleNewReset}
            className="w-full py-2 px-4 bg-purple-100 text-purple-700 rounded-lg hover:bg-purple-200 transition-colors text-sm"
            disabled={isRedirecting}
          >
            🔑 Restablecer otra contraseña
          </button>

          {/* Opción: Ir al dashboard directamente (si está logueado) */}
          {onDashboard && (
            <button
              onClick={handleGoToDashboard}
              className="w-full py-2 px-4 bg-green-100 text-green-700 rounded-lg hover:bg-green-200 transition-colors text-sm"
              disabled={isRedirecting}
            >
              🏠 Ir al dashboard
            </button>
          )}

          {/* Información del email (si está disponible) */}
          {email && (
            <div className="p-3 bg-gray-50 rounded-lg">
              <div className="text-xs text-gray-600">
                <span className="font-medium">Cuenta actualizada:</span>
                <br />
                <span className="font-mono">{email}</span>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );

  const QuickTips = () => (
    <div className="mt-8 p-4 bg-gray-50 rounded-lg">
      <h4 className="text-sm font-medium text-gray-700 mb-2">💡 Consejos de seguridad</h4>
      <ul className="text-xs text-gray-600 space-y-1">
        <li>• Use una contraseña única para su cuenta GAMC</li>
        <li>• No comparta sus credenciales con terceros</li>
        <li>• Cierre sesión al usar computadoras compartidas</li>
        <li>• Contacte a soporte si detecta actividad sospechosa</li>
      </ul>
    </div>
  );

  // ========================================
  // RENDER PRINCIPAL
  // ========================================

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        {/* Header de éxito */}
        <div className="text-center mb-8">
          <div className="mx-auto h-20 w-20 bg-green-600 rounded-full flex items-center justify-center mb-4 relative">
            <span className="text-white text-3xl font-bold">✓</span>
            {/* Animación de éxito */}
            <div className="absolute inset-0 bg-green-600 rounded-full animate-ping opacity-25"></div>
          </div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">¡Contraseña restablecida!</h1>
          <p className="text-gray-600">
            {message}
          </p>
        </div>

        {/* Mensaje de éxito principal */}
        <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-start">
            <span className="text-green-600 text-xl mr-3 mt-0.5">🎉</span>
            <div>
              <h3 className="text-green-800 font-semibold text-sm mb-1">¡Operación completada!</h3>
              <p className="text-green-700 text-xs">
                Su nueva contraseña ha sido configurada exitosamente. Ya puede acceder al sistema con sus nuevas credenciales.
              </p>
            </div>
          </div>
        </div>

        {/* Countdown automático */}
        <CountdownDisplay />

        {/* Aviso de seguridad */}
        <SecurityNotice />

        {/* Botones de acción principales */}
        <div className="space-y-3 mb-6">
          <button
            onClick={handleLoginNow}
            disabled={isRedirecting}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              isRedirecting
                ? 'bg-gray-400 cursor-not-allowed text-white'
                : 'bg-blue-600 hover:bg-blue-700 text-white'
            }`}
          >
            {isRedirecting ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Redirigiendo...
              </div>
            ) : (
              '🔐 Iniciar sesión ahora'
            )}
          </button>

          {countdown > 0 && autoRedirect && (
            <button
              onClick={() => setCountdown(0)}
              className="w-full py-2 px-4 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors text-sm"
              disabled={isRedirecting}
            >
              ⏹️ Cancelar redirección automática
            </button>
          )}
        </div>

        {/* Opciones adicionales */}
        <AdditionalOptions />

        {/* Consejos de seguridad */}
        <QuickTips />

        {/* Footer con soporte */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <div className="text-center text-xs text-gray-500">
            <p className="mb-2">¿Necesita ayuda adicional?</p>
            <div className="space-y-1">
              <p>📞 Soporte técnico: ext. 123</p>
              <p>📧 Email: <span className="font-medium text-gray-700">soporte@gamc.gov.bo</span></p>
              <p>🕐 Horario: Lunes a Viernes, 8:00 - 17:00</p>
            </div>
          </div>
        </div>

        {/* Indicador de estado */}
        {isRedirecting && (
          <div className="fixed inset-0 bg-black bg-opacity-25 flex items-center justify-center z-50">
            <div className="bg-white p-6 rounded-lg shadow-xl">
              <div className="flex items-center">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mr-3"></div>
                <span className="text-gray-700">Redirigiendo al login...</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default PasswordResetSuccess;
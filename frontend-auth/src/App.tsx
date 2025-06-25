// src/App.tsx (VERSIÓN COMPLETA)
// Aplicación principal con routing completo para recuperación de contraseñas
// Incluye flujo completo: forgot-password → security-question → reset-password → success

import React, { useState, useEffect } from 'react';
import LoginForm from './components/auth/LoginForm';
import RegisterForm from './components/auth/RegisterForm';
import DashboardPage from './pages/DashboardPage';
import ForgotPasswordForm from './components/auth/ForgotPasswordForm';
import SecurityQuestionForm from './components/auth/SecurityQuestionForm';
import ResetPasswordForm from './components/auth/ResetPasswordForm';
import PasswordResetSuccess from './components/auth/PasswordResetSuccess';
import { extractTokenFromURL, cleanTokenFromURL } from './utils/tokenValidation';
import { PasswordResetError } from './services/passwordResetService';
import { SecurityQuestionError } from './services/securityQuestionsService';

// Tipos de vistas disponibles
type ViewType = 
  | 'home' 
  | 'login' 
  | 'register' 
  | 'dashboard' 
  | 'forgot-password'
  | 'security-question'
  | 'reset-password'
  | 'reset-success';

// Interface del usuario
interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: 'admin' | 'input' | 'output';
  organizationalUnit: {
    id: number;
    name: string;
    code: string;
  };
}

// Interface para datos del proceso de reset
interface ResetProcessData {
  email: string;
  requiresSecurityQuestion: boolean;
  securityQuestion?: {
    questionId: number;
    questionText: string;
    attempts: number;
    maxAttempts: number;
  };
  resetToken?: string;
}

const App: React.FC = () => {
  // Estados principales
  const [currentView, setCurrentView] = useState<ViewType>('home');
  const [user, setUser] = useState<User | null>(null);
  const [fromRegistration, setFromRegistration] = useState(false);
  
  // Estados del proceso de reset
  const [resetProcessData, setResetProcessData] = useState<ResetProcessData | null>(null);
  const [resetSuccessMessage, setResetSuccessMessage] = useState<string>('');

  // ========================================
  // EFECTOS DE INICIALIZACIÓN
  // ========================================

  // Verificar sesión activa y token de reset en URL
  useEffect(() => {
    // 1. Verificar si hay una sesión activa
    const checkAuthStatus = () => {
      const token = localStorage.getItem('accessToken');
      const userData = localStorage.getItem('user');
      
      if (token && userData) {
        try {
          const parsedUser = JSON.parse(userData);
          setUser(parsedUser);
          setCurrentView('dashboard');
          return true; // Sesión activa encontrada
        } catch (error) {
          console.error('Error parsing user data:', error);
          // Limpiar datos corruptos
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      }
      return false; // No hay sesión activa
    };

    // 2. Verificar si hay un token de reset en la URL
    const checkResetToken = () => {
      const token = extractTokenFromURL();
      if (token) {
        console.log('Token de reset detectado en URL:', token.substring(0, 10) + '...');
        // Ir directamente al formulario de reset con token
        setCurrentView('reset-password');
        setResetProcessData({
          email: '', // Se podría extraer del token o solicitar al usuario
          requiresSecurityQuestion: false,
          resetToken: token
        });
        // Limpiar token de URL por seguridad
        cleanTokenFromURL();
        return true; // Token encontrado
      }
      return false; // No hay token
    };

    // Ejecutar verificaciones
    const hasActiveSession = checkAuthStatus();
    const hasResetToken = checkResetToken();

    // Si no hay sesión activa ni token de reset, quedarse en home
    if (!hasActiveSession && !hasResetToken) {
      setCurrentView('home');
    }
  }, []);

  // ========================================
  // HANDLERS DE AUTENTICACIÓN
  // ========================================

  /**
   * Maneja login exitoso
   */
  const handleLoginSuccess = (userData: User) => {
    setUser(userData);
    setCurrentView('dashboard');
    setFromRegistration(false);
    // Limpiar datos de reset si existen
    setResetProcessData(null);
  };

  /**
   * Maneja registro exitoso
   */
  const handleRegistrationSuccess = () => {
    setFromRegistration(true);
    setCurrentView('login');
  };

  /**
   * Maneja logout
   */
  const handleLogout = () => {
    setUser(null);
    setCurrentView('home');
    setFromRegistration(false);
    setResetProcessData(null);
  };

  // ========================================
  // HANDLERS DE NAVEGACIÓN
  // ========================================

  /**
   * Navega a una vista específica
   */
  const goToView = (view: ViewType) => {
    setCurrentView(view);
    if (view !== 'login') {
      setFromRegistration(false);
    }
    if (!view.includes('reset') && !view.includes('forgot')) {
      setResetProcessData(null);
    }
  };

  /**
   * Inicia proceso de reset de contraseña
   */
  const handleForgotPassword = () => {
    setResetProcessData(null);
    setCurrentView('forgot-password');
  };

  // ========================================
  // HANDLERS DEL PROCESO DE RESET
  // ========================================

  /**
   * Maneja éxito en solicitud de reset (forgot-password)
   */
  const handleForgotPasswordSuccess = (
    email: string, 
    requiresSecurityQuestion: boolean, 
    securityQuestion?: any
  ) => {
    const processData: ResetProcessData = {
      email,
      requiresSecurityQuestion,
      securityQuestion
    };

    setResetProcessData(processData);

    if (requiresSecurityQuestion && securityQuestion) {
      // Ir a verificación de pregunta de seguridad
      setCurrentView('security-question');
    } else {
      // Usuario debe revisar su email (flujo futuro)
      // Por ahora mostrar mensaje de éxito
      alert('✅ Revise su email institucional para continuar con el reset');
      setCurrentView('login');
    }
  };

  /**
   * Maneja error en solicitud de reset
   */
  const handleForgotPasswordError = (error: PasswordResetError) => {
    console.error('Error en forgot password:', error);
    // El error ya se maneja en el componente ForgotPasswordForm
  };

  /**
   * Maneja éxito en verificación de pregunta de seguridad
   */
  const handleSecurityQuestionSuccess = (resetToken: string) => {
    if (!resetProcessData) {
      console.error('No hay datos del proceso de reset');
      return;
    }

    // Actualizar datos con el token obtenido
    const updatedData: ResetProcessData = {
      ...resetProcessData,
      resetToken
    };

    setResetProcessData(updatedData);
    setCurrentView('reset-password');
  };

  /**
   * Maneja error en verificación de pregunta de seguridad
   */
  const handleSecurityQuestionError = (error: SecurityQuestionError) => {
    console.error('Error en security question:', error);
    // El error ya se maneja en el componente SecurityQuestionForm
  };

  /**
   * Maneja éxito en confirmación de reset de contraseña
   */
  const handleResetPasswordSuccess = (message: string) => {
    setResetSuccessMessage(message || 'Contraseña restablecida exitosamente');
    setCurrentView('reset-success');
    // Limpiar datos del proceso
    setResetProcessData(null);
  };

  /**
   * Maneja error en confirmación de reset de contraseña
   */
  const handleResetPasswordError = (error: PasswordResetError) => {
    console.error('Error en reset password:', error);
    // El error ya se maneja en el componente ResetPasswordForm
  };

  /**
   * Maneja finalización exitosa del proceso completo
   */
  const handleResetSuccessComplete = () => {
    setResetSuccessMessage('');
    setResetProcessData(null);
    setCurrentView('login');
  };

  /**
   * Maneja solicitud de nuevo reset desde cualquier punto
   */
  const handleNewReset = () => {
    setResetProcessData(null);
    setResetSuccessMessage('');
    setCurrentView('forgot-password');
  };

  // ========================================
  // COMPONENTES DE ERROR PARA DEBUGGING
  // ========================================

  const ErrorBoundary: React.FC<{ children: React.ReactNode; error?: string }> = ({ children, error }) => {
    if (error) {
      return (
        <div className="min-h-screen bg-red-50 flex items-center justify-center p-4">
          <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
            <h1 className="text-xl font-bold text-red-600 mb-4">Error en la aplicación</h1>
            <p className="text-gray-700 mb-4">{error}</p>
            <button 
              onClick={() => window.location.reload()} 
              className="bg-red-600 text-white px-4 py-2 rounded-lg"
            >
              Recargar página
            </button>
          </div>
        </div>
      );
    }
    return <>{children}</>;
  };

  // ========================================
  // RENDER DE VISTAS
  // ========================================

  // Vista principal (Home)
  if (currentView === 'home') {
    return (
      <ErrorBoundary>
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
          <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
            <div className="mb-6">
              <div className="mx-auto h-16 w-16 bg-blue-600 rounded-full flex items-center justify-center mb-4">
                <span className="text-white text-2xl font-bold">G</span>
              </div>
              <h1 className="text-3xl font-bold text-gray-900">GAMC</h1>
              <p className="text-gray-600">Sistema Web Centralizado</p>
            </div>
            
            <div className="space-y-4">
              {/* Estados del sistema */}
              <div className="bg-green-50 border border-green-200 rounded-lg p-4">
                <p className="text-green-800 font-semibold">✅ Sistema: Funcionando</p>
                <p className="text-green-600 text-sm">Frontend + Backend integrados</p>
              </div>
              
              {/* Botones principales */}
              <div className="grid grid-cols-2 gap-4">
                <button 
                  onClick={() => goToView('register')}
                  className="bg-green-600 text-white py-3 px-4 rounded-lg hover:bg-green-700 transition-colors font-semibold"
                >
                  📝 Registrarse
                </button>
                <button 
                  onClick={() => goToView('login')}
                  className="bg-blue-600 text-white py-3 px-4 rounded-lg hover:bg-blue-700 transition-colors font-semibold"
                >
                  🔐 Iniciar Sesión
                </button>
              </div>

              {/* Reset de contraseña */}
              <button 
                onClick={handleForgotPassword}
                className="w-full bg-purple-600 text-white py-2 px-4 rounded-lg hover:bg-purple-700 transition-colors text-sm"
              >
                🔑 ¿Olvidaste tu contraseña?
              </button>
              
              {/* Enlaces adicionales */}
              <div className="flex gap-2">
                <button 
                  onClick={() => window.open('http://localhost:3000/health', '_blank')}
                  className="flex-1 bg-gray-600 text-white py-2 px-4 rounded-lg hover:bg-gray-700 transition-colors text-sm"
                >
                  📊 API Status
                </button>
                <button 
                  onClick={() => window.open('http://localhost:3000/api/v1/auth/security-questions', '_blank')}
                  className="flex-1 bg-indigo-600 text-white py-2 px-4 rounded-lg hover:bg-indigo-700 transition-colors text-sm"
                >
                  ❓ Preguntas
                </button>
              </div>

              {/* Información del sistema */}
              <div className="mt-6 pt-4 border-t border-gray-200">
                <h3 className="text-sm font-medium text-gray-700 mb-2">Módulos Implementados</h3>
                <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="font-medium">✅ Autenticación</div>
                    <div>Login/Register</div>
                  </div>
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="font-medium">✅ Reset Password</div>
                    <div>Con preguntas</div>
                  </div>
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="font-medium">🚧 Mensajería</div>
                    <div>En desarrollo</div>
                  </div>
                  <div className="bg-gray-50 p-2 rounded">
                    <div className="font-medium">🚧 Archivos</div>
                    <div>Pendiente</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </ErrorBoundary>
    );
  }

  // Vista de registro
  if (currentView === 'register') {
    return (
      <ErrorBoundary>
        <RegisterForm 
          onRegistrationSuccess={handleRegistrationSuccess}
          onBack={() => goToView('home')}
        />
      </ErrorBoundary>
    );
  }

  // Vista de login
  if (currentView === 'login') {
    return (
      <ErrorBoundary>
        <LoginForm 
          onLoginSuccess={handleLoginSuccess}
          onBack={() => goToView('home')}
          onForgotPassword={handleForgotPassword}
          fromRegistration={fromRegistration}
        />
      </ErrorBoundary>
    );
  }

  // Vista de forgot password
  if (currentView === 'forgot-password') {
    return (
      <ErrorBoundary>
        <ForgotPasswordForm
          onSuccess={handleForgotPasswordSuccess}
          onError={handleForgotPasswordError}
          onBack={() => goToView('login')}
          onLoginRedirect={() => goToView('login')}
          autoFocus={true}
        />
      </ErrorBoundary>
    );
  }

  // Vista de verificación de pregunta de seguridad
  if (currentView === 'security-question') {
    if (!resetProcessData || !resetProcessData.securityQuestion) {
      return (
        <ErrorBoundary error="Error: No hay datos de pregunta de seguridad disponibles">
          <div>Error en datos de seguridad</div>
        </ErrorBoundary>
      );
    }

    return (
      <ErrorBoundary>
        <SecurityQuestionForm
          email={resetProcessData.email}
          securityQuestion={resetProcessData.securityQuestion}
          onSuccess={handleSecurityQuestionSuccess}
          onError={handleSecurityQuestionError}
          onBack={() => goToView('forgot-password')}
          onNewReset={handleNewReset}
          autoFocus={true}
        />
      </ErrorBoundary>
    );
  }

  // Vista de reset de contraseña
  if (currentView === 'reset-password') {
    if (!resetProcessData?.resetToken) {
      return (
        <ErrorBoundary error="Error: No hay token de reset disponible">
          <div>Error en token de reset</div>
        </ErrorBoundary>
      );
    }

    return (
      <ErrorBoundary>
        <ResetPasswordForm
          token={resetProcessData.resetToken}
          onSuccess={handleResetPasswordSuccess}
          onError={handleResetPasswordError}
          onTokenExpired={handleNewReset}
          onBack={() => goToView('forgot-password')}
        />
      </ErrorBoundary>
    );
  }

  // Vista de éxito en reset
  if (currentView === 'reset-success') {
    return (
      <ErrorBoundary>
        <PasswordResetSuccess
          message={resetSuccessMessage}
          email={resetProcessData?.email}
          onLoginRedirect={handleResetSuccessComplete}
          onNewReset={handleNewReset}
          autoRedirect={true}
          redirectDelay={5000}
        />
      </ErrorBoundary>
    );
  }

  // Vista de dashboard
  if (currentView === 'dashboard' && user) {
    return (
      <ErrorBoundary>
        <DashboardPage 
          onLogout={handleLogout}
        />
      </ErrorBoundary>
    );
  }

  // Fallback - loading o error
  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Cargando aplicación...</p>
          <p className="text-gray-400 text-sm mt-2">
            Vista actual: {currentView}
          </p>
        </div>
      </div>
    </ErrorBoundary>
  );
};

export default App;
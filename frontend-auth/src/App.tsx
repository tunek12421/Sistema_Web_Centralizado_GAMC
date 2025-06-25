import React, { useState, useEffect } from 'react';
import LoginForm from './components/auth/LoginForm';
import RegisterForm from './components/auth/RegisterForm';
import DashboardPage from './pages/DashboardPage';
import ForgotPasswordForm from './components/auth/ForgotPasswordForm';
import SecurityQuestionForm from './components/auth/SecurityQuestionForm';
import ResetPasswordForm from './components/auth/ResetPasswordForm';
import PasswordResetSuccess from './components/auth/PasswordResetSuccess';

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
    // 1. Verificar si hay token de reset en la URL
    const checkForResetToken = () => {
      const urlParams = new URLSearchParams(window.location.search);
      const token = urlParams.get('token');
      
      if (token && token.length === 64) {
        // Hay un token válido en la URL, ir directamente a reset-password
        setResetProcessData({
          email: '',
          requiresSecurityQuestion: false,
          resetToken: token
        });
        setCurrentView('reset-password');
        return true;
      }
      return false;
    };

    // 2. Verificar sesión activa
    const checkAuthStatus = () => {
      const accessToken = localStorage.getItem('accessToken');
      const userData = localStorage.getItem('user');
      
      if (accessToken && userData) {
        try {
          const parsedUser = JSON.parse(userData);
          setUser(parsedUser);
          setCurrentView('dashboard');
          return true;
        } catch (error) {
          console.error('Error parsing user data:', error);
          // Limpiar datos corruptos
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      }
      return false;
    };

    // Ejecutar verificaciones en orden de prioridad
    if (!checkForResetToken()) {
      checkAuthStatus();
    }
  }, []);

  // ========================================
  // HANDLERS DE NAVEGACIÓN PRINCIPAL
  // ========================================

  // Manejar login exitoso
  const handleLoginSuccess = (userData: User) => {
    setUser(userData);
    setCurrentView('dashboard');
    setFromRegistration(false);
  };

  // Manejar registro exitoso
  const handleRegistrationSuccess = () => {
    setFromRegistration(true);
    setCurrentView('login');
  };

  // Manejar logout
  const handleLogout = () => {
    setUser(null);
    setCurrentView('home');
    setFromRegistration(false);
    setResetProcessData(null);
    // Los tokens ya se limpian en el componente DashboardPage
  };

  // Ir a vista específica
  const goToView = (view: ViewType) => {
    setCurrentView(view);
    if (view !== 'login') {
      setFromRegistration(false);
    }
    if (view === 'home') {
      setResetProcessData(null);
    }
  };

  // ========================================
  // HANDLERS DEL PROCESO DE RESET
  // ========================================

  // Manejar solicitud de reset exitosa
  const handleForgotPasswordSuccess = (email: string, data: any) => {
    console.log('Reset solicitado para:', email, data);
    
    setResetProcessData({
      email: email,
      requiresSecurityQuestion: data.requiresSecurityQuestion || false,
      securityQuestion: data.securityQuestion,
      resetToken: data.resetToken
    });

    if (data.requiresSecurityQuestion && data.securityQuestion) {
      setCurrentView('security-question');
    } else if (data.resetToken) {
      setCurrentView('reset-password');
    } else {
      // Email enviado, volver al login con mensaje
      setResetSuccessMessage(`Se han enviado instrucciones de recuperación a ${email}`);
      setCurrentView('login');
    }
  };

  // Manejar verificación de pregunta de seguridad exitosa
  const handleSecurityQuestionSuccess = (resetToken: string) => {
    if (resetProcessData) {
      setResetProcessData({
        ...resetProcessData,
        resetToken: resetToken
      });
    }
    setCurrentView('reset-password');
  };

  // Manejar reset de contraseña exitoso
  const handleResetPasswordSuccess = () => {
    setResetSuccessMessage('¡Contraseña restablecida exitosamente!');
    setCurrentView('reset-success');
  };

  // Manejar finalización del proceso de reset
  const handleResetComplete = () => {
    setResetProcessData(null);
    setResetSuccessMessage('');
    setCurrentView('login');
  };

  // ========================================
  // RENDERIZADO DE VISTAS
  // ========================================

  // Vista principal (Home)
  if (currentView === 'home') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8 text-center">
          <div className="mb-6">
            <div className="mx-auto h-16 w-16 bg-blue-600 rounded-full flex items-center justify-center mb-4">
              <span className="text-white text-2xl font-bold">G</span>
            </div>
            <h1 className="text-3xl font-bold text-gray-900">GAMC</h1>
            <p className="text-gray-600">Sistema de Autenticación</p>
          </div>
          
          <div className="space-y-4">
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <p className="text-green-800 font-semibold">✅ Frontend: Funcionando</p>
            </div>
            <div className="bg-green-50 border border-green-200 rounded-lg p-4">
              <p className="text-green-800 font-semibold">✅ Backend: Funcionando</p>
              <p className="text-green-600 text-sm">Puerto 3000</p>
            </div>
            
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
                🔐 Login
              </button>
            </div>
            
            <button 
              onClick={() => window.open('http://localhost:3000/health', '_blank')}
              className="w-full bg-gray-600 text-white py-2 px-4 rounded-lg hover:bg-gray-700 transition-colors"
            >
              📊 Health Check API
            </button>

            {/* Información adicional */}
            <div className="mt-6 pt-4 border-t border-gray-200">
              <h3 className="text-sm font-medium text-gray-700 mb-2">Sistema Web Centralizado</h3>
              <div className="grid grid-cols-2 gap-2 text-xs text-gray-600">
                <div className="bg-gray-50 p-2 rounded">
                  <div className="font-medium">Versión</div>
                  <div>1.0.0</div>
                </div>
                <div className="bg-gray-50 p-2 rounded">
                  <div className="font-medium">Entorno</div>
                  <div>Desarrollo</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Vista de registro
  if (currentView === 'register') {
    return (
      <RegisterForm 
        onRegistrationSuccess={handleRegistrationSuccess}
        onBack={() => goToView('home')}
      />
    );
  }

  // Vista de login
  if (currentView === 'login') {
    return (
      <LoginForm 
        onLoginSuccess={handleLoginSuccess}
        onBack={() => goToView('home')}
        onForgotPassword={() => goToView('forgot-password')}
        fromRegistration={fromRegistration}
      />
    );
  }

  // ========================================
  // VISTAS DEL PROCESO DE RESET
  // ========================================

  // Vista de solicitud de reset (forgot password)
  if (currentView === 'forgot-password') {
    return (
      <ForgotPasswordForm 
        onSuccess={handleForgotPasswordSuccess}
        onBack={() => goToView('login')}
        onLoginRedirect={() => goToView('login')}
        onError={(error) => {
          console.error('Error en forgot password:', error);
          // Manejar error si es necesario
        }}
      />
    );
  }

  // Vista de pregunta de seguridad
  if (currentView === 'security-question' && resetProcessData?.securityQuestion) {
    return (
      <SecurityQuestionForm 
        email={resetProcessData.email}
        securityQuestion={resetProcessData.securityQuestion}
        onSuccess={handleSecurityQuestionSuccess}
        onBack={() => goToView('forgot-password')}
        onError={(error) => {
          console.error('Error en security question:', error);
          // Volver a forgot password en caso de error
          setCurrentView('forgot-password');
        }}
      />
    );
  }

  // Vista de reset de contraseña
  if (currentView === 'reset-password' && resetProcessData?.resetToken) {
    return (
      <ResetPasswordForm 
        token={resetProcessData.resetToken}
        onSuccess={handleResetPasswordSuccess}
        onBack={() => goToView('forgot-password')}
        onError={(error) => {
          console.error('Error en reset password:', error);
          // Volver a forgot password en caso de error
          setCurrentView('forgot-password');
        }}
      />
    );
  }

  // Vista de éxito en reset
  if (currentView === 'reset-success') {
    return (
      <PasswordResetSuccess 
        message={resetSuccessMessage}
        onLoginRedirect={handleResetComplete}
      />
    );
  }

  // Vista de dashboard
  if (currentView === 'dashboard' && user) {
    return (
      <DashboardPage 
        onLogout={handleLogout}
      />
    );
  }

  // Fallback - no debería llegar aquí nunca
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-600">Cargando aplicación...</p>
      </div>
    </div>
  );
};

export default App;
// src/App.tsx (VERSIÓN SIMPLIFICADA)
// Agregamos solo forgot-password sin importaciones complejas

import React, { useState, useEffect } from 'react';
import LoginForm from './components/auth/LoginForm';
import RegisterForm from './components/auth/RegisterForm';
import DashboardPage from './pages/DashboardPage';

type ViewType = 'home' | 'login' | 'register' | 'dashboard' | 'forgot-password';

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

const App: React.FC = () => {
  const [currentView, setCurrentView] = useState<ViewType>('home');
  const [user, setUser] = useState<User | null>(null);
  const [fromRegistration, setFromRegistration] = useState(false);

  // Verificar si hay una sesión activa al cargar la aplicación
  useEffect(() => {
    const checkAuthStatus = () => {
      const token = localStorage.getItem('accessToken');
      const userData = localStorage.getItem('user');
      
      if (token && userData) {
        try {
          const parsedUser = JSON.parse(userData);
          setUser(parsedUser);
          setCurrentView('dashboard');
        } catch (error) {
          console.error('Error parsing user data:', error);
          // Limpiar datos corruptos
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          localStorage.removeItem('user');
        }
      }
    };

    checkAuthStatus();
  }, []);

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
    // Los tokens ya se limpian en el componente DashboardPage
  };

  // Ir a vista específica
  const goToView = (view: ViewType) => {
    setCurrentView(view);
    if (view !== 'login') {
      setFromRegistration(false);
    }
  };

  // ⭐ NUEVO: Manejar forgot password
  const handleForgotPassword = () => {
    setCurrentView('forgot-password');
  };

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

            {/* ⭐ NUEVO: Botón para testing de reset password */}
            <button 
              onClick={() => goToView('forgot-password')}
              className="w-full bg-purple-600 text-white py-2 px-4 rounded-lg hover:bg-purple-700 transition-colors text-sm"
            >
              🔑 ¿Olvidaste tu contraseña?
            </button>
            
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
        onForgotPassword={handleForgotPassword} // ⭐ NUEVO
        fromRegistration={fromRegistration}
      />
    );
  }

  // ⭐ NUEVO: Vista de forgot password (simplificada)
  if (currentView === 'forgot-password') {
    return <SimpleForgotPasswordPage onBack={() => goToView('login')} />;
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

// ⭐ NUEVO: Componente simplificado para forgot password
const SimpleForgotPasswordPage: React.FC<{ onBack: () => void }> = ({ onBack }) => {
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    try {
      const response = await fetch('http://localhost:3000/api/v1/auth/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email }),
      });

      const result = await response.json();

      if (response.ok) {
        setMessage('✅ ' + (result.data?.message || 'Email de reset enviado'));
        setIsSubmitted(true);
        setEmail('');
      } else {
        setMessage('❌ ' + (result.message || 'Error al enviar email'));
      }
    } catch (error) {
      setMessage('❌ Error de conexión');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-purple-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">🔑</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">¿Olvidaste tu contraseña?</h1>
          <p className="text-gray-600 mt-2">
            Ingresa tu email institucional para recibir un enlace de reset
          </p>
        </div>

        {isSubmitted ? (
          <div className="space-y-6">
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <div className="flex items-start">
                <span className="text-green-600 text-xl mr-3 mt-0.5">📧</span>
                <div>
                  <h3 className="text-green-800 font-semibold text-sm">¡Email enviado!</h3>
                  <p className="text-green-700 text-xs mt-1">
                    Revisa tu correo institucional para el enlace de reset
                  </p>
                </div>
              </div>
            </div>
            
            <button
              onClick={onBack}
              className="w-full py-3 px-4 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-semibold"
            >
              🔐 Volver al login
            </button>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email institucional <span className="text-red-500">*</span>
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500"
                placeholder="usuario@gamc.gov.bo"
                required
                disabled={loading}
              />
              <p className="text-xs text-gray-500 mt-1">
                Solo emails @gamc.gov.bo pueden solicitar reset
              </p>
            </div>

            {message && (
              <div className={`p-3 rounded-lg text-sm border ${
                message.includes('✅') 
                  ? 'bg-green-50 text-green-800 border-green-200' 
                  : 'bg-red-50 text-red-800 border-red-200'
              }`}>
                {message}
              </div>
            )}

            <button
              type="submit"
              disabled={loading || !email}
              className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
                loading || !email
                  ? 'bg-gray-400 cursor-not-allowed text-white' 
                  : 'bg-purple-600 hover:bg-purple-700 text-white'
              }`}
            >
              {loading ? '🔄 Enviando...' : '📧 Enviar enlace de reset'}
            </button>

            <button
              type="button"
              onClick={onBack}
              disabled={loading}
              className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
            >
              ← Volver al login
            </button>
          </form>
        )}
      </div>
    </div>
  );
};

export default App;
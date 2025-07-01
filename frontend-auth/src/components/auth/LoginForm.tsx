import React, { useState, useEffect } from 'react';

interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

interface LoginFormProps {
  onLoginSuccess: (userData: any) => void;
  onBack: () => void;
  onForgotPassword?: () => void; // ⭐ NUEVO: Callback para ir a forgot password
  fromRegistration?: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({ 
  onLoginSuccess, 
  onBack, 
  onForgotPassword, // ⭐ NUEVO
  fromRegistration = false 
}) => {
  const [loginData, setLoginData] = useState({ email: '', password: '' });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [showRegistrationMessage, setShowRegistrationMessage] = useState(fromRegistration);
  
  // Estados para validaciones en tiempo real
  const [loginValidation, setLoginValidation] = useState<{[key: string]: FieldValidation}>({});

  // Función para validar email
  const validateEmail = (email: string): FieldValidation => {
    if (!email) {
      return { isValid: false, message: 'El email es requerido', type: 'error' };
    }
    
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return { isValid: false, message: 'Formato de email inválido', type: 'error' };
    }
    
    // Preferencia por emails institucionales
    if (email.includes('@gamc.gov.bo')) {
      return { isValid: true, message: '✓ Email institucional GAMC', type: 'success' };
    }
    
    return { isValid: true, message: '✓ Email válido (se recomienda usar @gamc.gov.bo)', type: 'warning' };
  };

  // Validaciones para login
  useEffect(() => {
    const validations: {[key: string]: FieldValidation} = {};
    
    validations.email = validateEmail(loginData.email);
    
    if (!loginData.password) {
      validations.password = { isValid: false, message: 'La contraseña es requerida', type: 'error' };
    } else {
      validations.password = { isValid: true, message: '✓ Contraseña ingresada', type: 'success' };
    }
    
    setLoginValidation(validations);
  }, [loginData]);

  // Función mejorada de login
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    // Validación final antes del envío
    const emailValidation = validateEmail(loginData.email);
    if (!emailValidation.isValid || !loginData.password) {
      setMessage('❌ Complete todos los campos correctamente antes de continuar');
      setLoading(false);
      return;
    }

    try {
      const response = await fetch('http://localhost:3000/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(loginData),
      });

      // Verificar si la respuesta es exitosa
      if (response.ok) {
        const result = await response.json();
        
        if (result.success) {
          setMessage(`🎉 ¡Bienvenido ${result.data.user.firstName}! Login exitoso`);
          
          // Guardar tokens y datos del usuario
          localStorage.setItem('accessToken', result.data.accessToken);
          localStorage.setItem('refreshToken', result.data.refreshToken);
          localStorage.setItem('user', JSON.stringify(result.data.user));
          
          // Simular redirección al dashboard después de 1 segundo
          setTimeout(() => {
            setMessage('🚀 Redirigiendo al dashboard...');
            setTimeout(() => {
              onLoginSuccess(result.data.user);
            }, 1000);
          }, 1000);
        } else {
          setMessage(`❌ ${result.message || 'Error en el login'}`);
        }
      } else {
        // Manejar errores HTTP específicos
        const result = await response.json().catch(() => ({}));
        
        switch (response.status) {
          case 401:
            setMessage('🔐 Credenciales incorrectas: Verifique su email y contraseña');
            break;
          case 403:
            setMessage('⛔ Cuenta inactiva: Contacte al administrador del sistema');
            break;
          case 400:
            setMessage('📝 Datos inválidos: Verifique el formato de email y contraseña');
            break;
          case 500:
            setMessage('⚠️ Error del servidor: Intente nuevamente en unos momentos');
            break;
          default:
            setMessage(`❌ Error ${response.status}: ${result.message || 'Error desconocido'}`);
        }
      }
    } catch (error: any) {
      // Solo mostrar error de conexión si realmente es un error de red
      console.error('Error de red:', error);
      setMessage('🔌 Error de conexión: Verifique que el servidor esté ejecutándose en puerto 3000');
    } finally {
      setLoading(false);
    }
  };

  // ⭐ NUEVO: Manejador para forgot password
  const handleForgotPassword = () => {
    if (onForgotPassword) {
      onForgotPassword();
    }
  };

  // Componente para mostrar validación de campo
  const FieldValidationMessage = ({ validation }: { validation?: FieldValidation }) => {
    if (!validation || validation.type === '') return null;
    
    const colors = {
      success: 'text-green-600 bg-green-50 border-green-200',
      error: 'text-red-600 bg-red-50 border-red-200',
      warning: 'text-amber-600 bg-amber-50 border-amber-200'
    };
    
    return (
      <div className={`text-xs mt-1 p-2 rounded border ${colors[validation.type]}`}>
        {validation.message}
      </div>
    );
  };

  // Función para obtener clases de input según validación
  const getInputClasses = (validation?: FieldValidation, baseClasses: string = '') => {
    const base = baseClasses || 'w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 transition-colors';
    
    if (!validation || validation.type === '') {
      return `${base} border-gray-300 focus:ring-blue-500`;
    }
    
    switch (validation.type) {
      case 'success':
        return `${base} border-green-300 focus:ring-green-500 bg-green-50`;
      case 'error':
        return `${base} border-red-300 focus:ring-red-500 bg-red-50`;
      case 'warning':
        return `${base} border-amber-300 focus:ring-amber-500 bg-amber-50`;
      default:
        return `${base} border-gray-300 focus:ring-blue-500`;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-blue-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">🔐</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Iniciar Sesión</h1>
          <p className="text-gray-600">Sistema GAMC</p>
          
          {showRegistrationMessage && (
            <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 text-sm">
                ✅ Cuenta creada exitosamente. Ingrese sus credenciales para continuar.
              </p>
            </div>
          )}
        </div>

        <form onSubmit={handleLogin} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Correo Electrónico <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              value={loginData.email}
              onChange={(e) => {
                setLoginData({ ...loginData, email: e.target.value });
                setShowRegistrationMessage(false); // Limpiar mensaje al empezar a escribir
              }}
              className={getInputClasses(loginValidation.email)}
              placeholder="usuario@gamc.gov.bo"
              required
            />
            <FieldValidationMessage validation={loginValidation.email} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Contraseña <span className="text-red-500">*</span>
            </label>
            <input
              type="password"
              value={loginData.password}
              onChange={(e) => {
                setLoginData({ ...loginData, password: e.target.value });
                setShowRegistrationMessage(false); // Limpiar mensaje al empezar a escribir
              }}
              className={getInputClasses(loginValidation.password)}
              placeholder="Tu contraseña"
              required
            />
            <FieldValidationMessage validation={loginValidation.password} />
          </div>

          {/* ⭐ NUEVO: Enlace de forgot password */}
          {onForgotPassword && (
            <div className="text-right">
              <button
                type="button"
                onClick={handleForgotPassword}
                className="text-sm text-blue-600 hover:text-blue-800 transition-colors"
                disabled={loading}
              >
                ¿Olvidaste tu contraseña?
              </button>
            </div>
          )}

          {message && (
            <div className={`p-3 rounded-lg text-sm border ${
              message.includes('🎉') || message.includes('🚀')
                ? 'bg-green-50 text-green-800 border-green-200' 
                : 'bg-red-50 text-red-800 border-red-200'
            }`}>
              {message}
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !loginValidation.email?.isValid || !loginData.password}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              loading || !loginValidation.email?.isValid || !loginData.password
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-blue-600 hover:bg-blue-700 text-white'
            }`}
          >
            {loading ? '🔄 Iniciando sesión...' : '🔐 Iniciar Sesión'}
          </button>

          <button
            type="button"
            onClick={() => {
              onBack();
              setShowRegistrationMessage(false);
            }}
            className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
          >
            ← Volver
          </button>
        </form>
      </div>
    </div>
  );
};

export default LoginForm;
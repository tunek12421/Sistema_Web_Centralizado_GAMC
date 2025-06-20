import React, { useState, useEffect } from 'react';

interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

interface RegisterFormProps {
  onRegistrationSuccess: () => void;
  onBack: () => void;
}

const RegisterForm: React.FC<RegisterFormProps> = ({ onRegistrationSuccess, onBack }) => {
  const [registerData, setRegisterData] = useState({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    organizationalUnitId: 1
  });
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  
  // Estados para validaciones en tiempo real
  const [registerValidation, setRegisterValidation] = useState<{[key: string]: FieldValidation}>({});
  const [isFormValid, setIsFormValid] = useState(false);

  // Unidades organizacionales
  const units = [
    { id: 1, name: 'Administración' },
    { id: 2, name: 'Obras Públicas' },
    { id: 3, name: 'Monitoreo' },
    { id: 4, name: 'Movilidad Urbana' },
    { id: 5, name: 'Gobierno Electrónico' },
    { id: 6, name: 'Prensa e Imagen' },
    { id: 7, name: 'Tecnología' }
  ];

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

  // Función para validar contraseña - CORREGIDA para sincronizar con backend
  const validatePassword = (password: string): FieldValidation => {
    if (!password) {
      return { isValid: false, message: 'La contraseña es requerida', type: 'error' };
    }
    
    if (password.length < 8) {
      return { isValid: false, message: 'Mínimo 8 caracteres', type: 'error' };
    }
    
    const hasUppercase = /[A-Z]/.test(password);
    const hasLowercase = /[a-z]/.test(password);
    const hasNumbers = /\d/.test(password);
    // CORRECCIÓN: Usar exactamente los mismos caracteres especiales que el backend
    const hasSymbols = /[@$!%*?&]/.test(password);
    
    const missing = [];
    if (!hasUppercase) missing.push('mayúscula');
    if (!hasLowercase) missing.push('minúscula');
    if (!hasNumbers) missing.push('número');
    if (!hasSymbols) missing.push('símbolo (@$!%*?&)');
    
    if (missing.length > 0) {
      return { 
        isValid: false, 
        message: `Falta: ${missing.join(', ')}`, 
        type: 'error' 
      };
    }
    
    return { isValid: true, message: '✓ Contraseña segura', type: 'success' };
  };

  // Función para validar nombres
  const validateName = (name: string, fieldName: string): FieldValidation => {
    if (!name) {
      return { isValid: false, message: `${fieldName} es requerido`, type: 'error' };
    }
    
    if (name.length < 2) {
      return { isValid: false, message: `${fieldName} debe tener al menos 2 caracteres`, type: 'error' };
    }
    
    if (name.length > 50) {
      return { isValid: false, message: `${fieldName} no puede exceder 50 caracteres`, type: 'error' };
    }
    
    const nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/;
    if (!nameRegex.test(name)) {
      return { isValid: false, message: 'Solo se permiten letras y espacios', type: 'error' };
    }
    
    return { isValid: true, message: `✓ ${fieldName} válido`, type: 'success' };
  };

  // Validaciones en tiempo real para registro
  useEffect(() => {
    const validations: {[key: string]: FieldValidation} = {};
    
    validations.firstName = validateName(registerData.firstName, 'Nombre');
    validations.lastName = validateName(registerData.lastName, 'Apellido');
    validations.email = validateEmail(registerData.email);
    validations.password = validatePassword(registerData.password);
    
    setRegisterValidation(validations);
    
    // Verificar si todo el formulario es válido
    const allValid = Object.values(validations).every(v => v.isValid);
    setIsFormValid(allValid);
  }, [registerData]);

  // Función mejorada de registro
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    // Validación final antes del envío
    if (!isFormValid) {
      setMessage('❌ Complete todos los campos correctamente antes de continuar');
      setLoading(false);
      return;
    }

    try {
      console.log('🔄 Enviando datos:', registerData);
      
      const response = await fetch('http://localhost:3000/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(registerData),
      });

      console.log('📡 Respuesta recibida:', {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok
      });

      // Leer el contenido de la respuesta
      const result = await response.json();
      console.log('✅ JSON parseado:', result);

      // Verificar si la respuesta es exitosa
      if (response.ok && result.success) {
        // CORRECCIÓN: result.data ES directamente el user, no result.data.user
        const unitName = units.find(u => u.id === registerData.organizationalUnitId)?.name;
        setMessage(`🎉 ¡Registro exitoso! Usuario ${result.data.firstName} creado en ${unitName}`);
        
        // Limpiar formulario
        setRegisterData({
          email: '',
          password: '',
          firstName: '',
          lastName: '',
          organizationalUnitId: 1
        });
        
        // Secuencia de mensajes y redirección
        setTimeout(() => {
          setMessage('✅ Cuenta creada. ¿Desea iniciar sesión ahora?');
        }, 2000);
        
        setTimeout(() => {
          setMessage('🚀 Redirigiendo al login...');
        }, 4000);
        
        setTimeout(() => {
          onRegistrationSuccess();
        }, 5000);
        
      } else {
        // Manejar errores HTTP específicos - MEJORADO
        switch (response.status) {
          case 409:
            setMessage('👤 Este email ya está registrado. ¿Desea iniciar sesión en su lugar?');
            break;
          case 400:
            if (result.message?.includes('ya existe')) {
              setMessage('👤 El usuario ya existe: Este email ya está registrado en el sistema');
            } else if (result.message?.includes('organizacional')) {
              setMessage('🏢 Unidad organizacional inválida. Seleccione una opción válida');
            } else if (result.message?.includes('password') || result.error?.includes('contraseña')) {
              setMessage(`🔒 ${result.error || result.message}`);
            } else {
              setMessage(`📝 Datos inválidos: ${result.message || 'Verifique los campos del formulario'}`);
            }
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
      console.error('❌ Error de red completo:', error);
      setMessage('🔌 Error de conexión: Verifique que el servidor esté ejecutándose en puerto 3000');
    } finally {
      setLoading(false);
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
      return `${base} border-gray-300 focus:ring-green-500`;
    }
    
    switch (validation.type) {
      case 'success':
        return `${base} border-green-300 focus:ring-green-500 bg-green-50`;
      case 'error':
        return `${base} border-red-300 focus:ring-red-500 bg-red-50`;
      case 'warning':
        return `${base} border-amber-300 focus:ring-amber-500 bg-amber-50`;
      default:
        return `${base} border-gray-300 focus:ring-green-500`;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 to-blue-100 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-green-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">📝</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Crear Cuenta</h1>
          <p className="text-gray-600">Sistema GAMC</p>
        </div>

        <form onSubmit={handleRegister} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Nombre <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={registerData.firstName}
                onChange={(e) => setRegisterData({ ...registerData, firstName: e.target.value })}
                className={getInputClasses(registerValidation.firstName)}
                placeholder="Tu nombre"
                required
              />
              <FieldValidationMessage validation={registerValidation.firstName} />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Apellido <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={registerData.lastName}
                onChange={(e) => setRegisterData({ ...registerData, lastName: e.target.value })}
                className={getInputClasses(registerValidation.lastName)}
                placeholder="Tu apellido"
                required
              />
              <FieldValidationMessage validation={registerValidation.lastName} />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Email <span className="text-red-500">*</span>
            </label>
            <input
              type="email"
              value={registerData.email}
              onChange={(e) => setRegisterData({ ...registerData, email: e.target.value })}
              className={getInputClasses(registerValidation.email)}
              placeholder="usuario@gamc.gov.bo"
              required
            />
            <FieldValidationMessage validation={registerValidation.email} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Contraseña <span className="text-red-500">*</span>
            </label>
            <input
              type="password"
              value={registerData.password}
              onChange={(e) => setRegisterData({ ...registerData, password: e.target.value })}
              className={getInputClasses(registerValidation.password)}
              placeholder="Mínimo 8 caracteres con @$!%*?&"
              required
              minLength={8}
            />
            <FieldValidationMessage validation={registerValidation.password} />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Unidad Organizacional <span className="text-red-500">*</span>
            </label>
            <select
              value={registerData.organizationalUnitId}
              onChange={(e) => setRegisterData({ ...registerData, organizationalUnitId: parseInt(e.target.value) })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
              required
            >
              {units.map(unit => (
                <option key={unit.id} value={unit.id}>{unit.name}</option>
              ))}
            </select>
            <p className="text-xs text-gray-500 mt-1">Seleccione la unidad donde trabajará</p>
          </div>

          {message && (
            <div className={`p-3 rounded-lg text-sm border ${
              message.includes('🎉') || message.includes('✅') 
                ? 'bg-green-50 text-green-800 border-green-200' 
                : message.includes('👤') 
                  ? 'bg-blue-50 text-blue-800 border-blue-200'
                  : 'bg-red-50 text-red-800 border-red-200'
            }`}>
              {message}
            </div>
          )}

          <button
            type="submit"
            disabled={loading || !isFormValid}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              loading || !isFormValid
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-green-600 hover:bg-green-700 text-white'
            }`}
          >
            {loading ? '🔄 Creando cuenta...' : '📝 Crear Cuenta'}
          </button>

          {!isFormValid && (
            <p className="text-xs text-center text-gray-500">
              Complete todos los campos correctamente para continuar
            </p>
          )}

          <button
            type="button"
            onClick={onBack}
            className="w-full py-2 px-4 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors"
          >
            ← Volver
          </button>
        </form>
      </div>
    </div>
  );
};

export default RegisterForm;
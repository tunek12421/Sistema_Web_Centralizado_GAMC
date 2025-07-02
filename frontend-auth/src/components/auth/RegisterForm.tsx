import React, { useState, useEffect } from 'react';

interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

interface SecurityQuestion {
  id: number;
  questionText: string;
  category: string;
}

interface UserSecurityQuestion {
  questionId: number;
  answer: string;
}

interface RegisterFormProps {
  onRegistrationSuccess: () => void;
  onBack: () => void;
}

// 🔧 NUEVA INTERFAZ para unidades organizacionales
interface OrganizationalUnit {
  id: number;
  name: string;
  code: string;
  isActive: boolean;
}

const RegisterForm: React.FC<RegisterFormProps> = ({ onRegistrationSuccess, onBack }) => {
  const [registerData, setRegisterData] = useState({
    email: '',
    password: '',
    firstName: '',
    lastName: '',
    organizationalUnitId: 0 // 🔧 CAMBIADO: Inicializar en 0 hasta cargar las unidades
  });

  // Estados para preguntas de seguridad
  const [securityQuestions, setSecurityQuestions] = useState<SecurityQuestion[]>([]);
  const [userSecurityQuestions, setUserSecurityQuestions] = useState<UserSecurityQuestion[]>([]);
  const [showSecurityQuestions, setShowSecurityQuestions] = useState(false);
  const [loadingQuestions, setLoadingQuestions] = useState(false);

  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  
  // Estados para validaciones en tiempo real
  const [registerValidation, setRegisterValidation] = useState<{[key: string]: FieldValidation}>({});
  const [securityValidation, setSecurityValidation] = useState<{[key: string]: FieldValidation}>({});
  const [isFormValid, setIsFormValid] = useState(false);

  // 🔧 NUEVO: Estados para unidades organizacionales dinámicas
  const [units, setUnits] = useState<OrganizationalUnit[]>([]);
  const [loadingUnits, setLoadingUnits] = useState(true);

  // 🔧 NUEVO: useEffect para cargar unidades organizacionales
  useEffect(() => {
    const loadOrganizationalUnits = async () => {
      try {
        setLoadingUnits(true);
        
        console.log('🔄 Cargando unidades organizacionales...');
        const response = await fetch('http://localhost:3000/api/v1/organizational-units');
        const result = await response.json();
        
        if (response.ok && result.success && result.data) {
          console.log('✅ Unidades cargadas:', result.data);
          setUnits(result.data);
          
          // 🔧 NUEVO: Configurar la primera unidad como valor por defecto
          if (result.data.length > 0) {
            setRegisterData(prev => ({ 
              ...prev, 
              organizationalUnitId: result.data[0].id 
            }));
          }
        } else {
          // 🔧 FALLBACK: Usar datos hardcodeados con IDs correctos (sin mostrar error)
          const fallbackUnits = [
            { id: 1, name: 'Obras Públicas', code: 'OBRAS_PUBLICAS', isActive: true },
            { id: 2, name: 'Monitoreo', code: 'MONITOREO', isActive: true },
            { id: 3, name: 'Movilidad Urbana', code: 'MOVILIDAD_URBANA', isActive: true },
            { id: 4, name: 'Gobierno Electrónico', code: 'GOBIERNO_ELECTRONICO', isActive: true },
            { id: 5, name: 'Prensa e Imagen', code: 'PRENSA_IMAGEN', isActive: true },
            { id: 6, name: 'Tecnología', code: 'TECNOLOGIA', isActive: true },
            { id: 7, name: 'Administración', code: 'ADMINISTRACION', isActive: true }
          ];
          setUnits(fallbackUnits);
          setRegisterData(prev => ({ ...prev, organizationalUnitId: fallbackUnits[0].id }));
        }
      } catch (error) {
        // 🔧 FALLBACK: Usar datos hardcodeados con IDs correctos (sin mostrar error)
        const fallbackUnits = [
          { id: 1, name: 'Obras Públicas', code: 'OBRAS_PUBLICAS', isActive: true },
          { id: 2, name: 'Monitoreo', code: 'MONITOREO', isActive: true },
          { id: 3, name: 'Movilidad Urbana', code: 'MOVILIDAD_URBANA', isActive: true },
          { id: 4, name: 'Gobierno Electrónico', code: 'GOBIERNO_ELECTRONICO', isActive: true },
          { id: 5, name: 'Prensa e Imagen', code: 'PRENSA_IMAGEN', isActive: true },
          { id: 6, name: 'Tecnología', code: 'TECNOLOGIA', isActive: true },
          { id: 7, name: 'Administración', code: 'ADMINISTRACION', isActive: true }
        ];
        setUnits(fallbackUnits);
        setRegisterData(prev => ({ ...prev, organizationalUnitId: fallbackUnits[0].id }));
      } finally {
        setLoadingUnits(false);
      }
    };

    loadOrganizationalUnits();
  }, []);

  // Cargar preguntas de seguridad disponibles
  const loadSecurityQuestions = async () => {
    if (securityQuestions.length > 0) return; // Ya están cargadas

    setLoadingQuestions(true);
    try {
      const response = await fetch('http://localhost:3000/api/v1/auth/security-questions');
      const result = await response.json();
      
      if (response.ok && result.success) {
        setSecurityQuestions(result.data.questions);
      } else {
        console.error('Error al cargar preguntas de seguridad:', result.message);
      }
    } catch (error) {
      console.error('Error de conexión al cargar preguntas:', error);
    } finally {
      setLoadingQuestions(false);
    }
  };

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

  // Función para validar contraseña
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
    
    if (name.trim() === '') {
      return { isValid: false, message: `${fieldName} no puede contener solo espacios en blanco`, type: 'error' };
    }
    
    const trimmedName = name.trim();
    if (trimmedName.length < 2) {
      return { isValid: false, message: `${fieldName} debe tener al menos 2 caracteres`, type: 'error' };
    }
    
    if (trimmedName.length > 50) {
      return { isValid: false, message: `${fieldName} no puede exceder 50 caracteres`, type: 'error' };
    }
    
    if (/\s{2,}/.test(name)) {
      return { isValid: false, message: `${fieldName} no puede tener espacios múltiples consecutivos`, type: 'error' };
    }
    
    if (name !== trimmedName) {
      return { isValid: false, message: `${fieldName} no puede empezar o terminar con espacios`, type: 'error' };
    }
    
    const nameRegex = /^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/;
    if (!nameRegex.test(trimmedName)) {
      return { isValid: false, message: 'Solo se permiten letras y espacios', type: 'error' };
    }
    
    if (/^(.)\1+$/.test(trimmedName.replace(/\s/g, ''))) {
      return { isValid: false, message: `${fieldName} no puede ser solo caracteres repetidos`, type: 'error' };
    }
    
    return { isValid: true, message: `✓ ${fieldName} válido`, type: 'success' };
  };

  // Función para validar respuesta de seguridad
  const validateSecurityAnswer = (answer: string): FieldValidation => {
    if (!answer) {
      return { isValid: false, message: 'La respuesta es requerida', type: 'error' };
    }

    const trimmedAnswer = answer.trim();
    if (trimmedAnswer.length < 2) {
      return { isValid: false, message: 'La respuesta debe tener al menos 2 caracteres', type: 'error' };
    }

    if (trimmedAnswer.length > 100) {
      return { isValid: false, message: 'La respuesta no puede exceder 100 caracteres', type: 'error' };
    }

    if (trimmedAnswer === '') {
      return { isValid: false, message: 'La respuesta no puede estar vacía', type: 'error' };
    }

    // Verificar que no sea solo caracteres repetidos
    if (/^(.)\1+$/.test(trimmedAnswer)) {
      return { isValid: false, message: 'La respuesta no puede ser solo caracteres repetidos', type: 'error' };
    }

    // Verificar que no sea solo números
    if (/^\d+$/.test(trimmedAnswer)) {
      return { isValid: false, message: 'La respuesta no puede ser solo números', type: 'error' };
    }

    return { isValid: true, message: '✓ Respuesta válida', type: 'success' };
  };

  // Función para manejar cambios en campos de nombre
  const handleNameChange = (value: string, field: 'firstName' | 'lastName') => {
    const cleanedValue = value.replace(/\s{2,}/g, ' ');
    setRegisterData({ 
      ...registerData, 
      [field]: cleanedValue 
    });
  };

  // Función para agregar pregunta de seguridad
  const addSecurityQuestion = () => {
    if (userSecurityQuestions.length < 3) {
      setUserSecurityQuestions([
        ...userSecurityQuestions,
        { questionId: 0, answer: '' }
      ]);
    }
  };

  // Función para remover pregunta de seguridad
  const removeSecurityQuestion = (index: number) => {
    setUserSecurityQuestions(userSecurityQuestions.filter((_, i) => i !== index));
  };

  // Función para actualizar pregunta de seguridad
  const updateSecurityQuestion = (index: number, field: 'questionId' | 'answer', value: string | number) => {
    const updated = [...userSecurityQuestions];
    updated[index] = { ...updated[index], [field]: value };
    setUserSecurityQuestions(updated);
  };

  // Obtener preguntas disponibles (no seleccionadas)
  const getAvailableQuestions = (currentIndex: number) => {
    const selectedIds = userSecurityQuestions
      .map((q, i) => i !== currentIndex ? q.questionId : 0)
      .filter(id => id > 0);
    
    return securityQuestions.filter(q => !selectedIds.includes(q.id));
  };

  // 🔧 ACTUALIZADO: Validaciones que incluyen verificación de unidad organizacional
  useEffect(() => {
    const validations: {[key: string]: FieldValidation} = {};
    
    validations.firstName = validateName(registerData.firstName, 'Nombre');
    validations.lastName = validateName(registerData.lastName, 'Apellido');
    validations.email = validateEmail(registerData.email);
    validations.password = validatePassword(registerData.password);
    
    // 🔧 NUEVO: Validar unidad organizacional
    if (registerData.organizationalUnitId === 0) {
      validations.organizationalUnit = {
        isValid: false,
        message: 'Seleccione una unidad organizacional',
        type: 'error'
      };
    } else {
      validations.organizationalUnit = {
        isValid: true,
        message: '✓ Unidad seleccionada',
        type: 'success'
      };
    }
    
    setRegisterValidation(validations);
    
    // Verificar si todo el formulario básico es válido
    const basicFormValid = Object.values(validations).every(v => v.isValid);
    
    // Validar preguntas de seguridad si están habilitadas
    let securityValid = true;
    const securityValidations: {[key: string]: FieldValidation} = {};
    
    if (showSecurityQuestions && userSecurityQuestions.length > 0) {
      userSecurityQuestions.forEach((q, index) => {
        // Validar que se haya seleccionado una pregunta
        if (q.questionId === 0) {
          securityValidations[`question_${index}`] = {
            isValid: false,
            message: 'Seleccione una pregunta',
            type: 'error'
          };
          securityValid = false;
        } else {
          securityValidations[`question_${index}`] = {
            isValid: true,
            message: '✓ Pregunta seleccionada',
            type: 'success'
          };
        }

        // Validar respuesta
        const answerValidation = validateSecurityAnswer(q.answer);
        securityValidations[`answer_${index}`] = answerValidation;
        if (!answerValidation.isValid) {
          securityValid = false;
        }
      });
    }
    
    setSecurityValidation(securityValidations);
    setIsFormValid(basicFormValid && securityValid);
  }, [registerData, userSecurityQuestions, showSecurityQuestions]);

  // Función mejorada de registro
  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage('');

    if (!isFormValid) {
      setMessage('❌ Complete todos los campos correctamente antes de continuar');
      setLoading(false);
      return;
    }

    try {
      const requestBody: any = { ...registerData };

      // Agregar preguntas de seguridad si están configuradas
      if (showSecurityQuestions && userSecurityQuestions.length > 0) {
        requestBody.securityQuestions = {
          questions: userSecurityQuestions.filter(q => q.questionId > 0 && q.answer.trim())
        };
      }

      console.log('🔄 Enviando datos:', requestBody);
      
      const response = await fetch('http://localhost:3000/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(requestBody),
      });

      const result = await response.json();
      console.log('📡 Respuesta recibida:', result);

      if (response.ok && result.success) {
        // 🔧 ACTUALIZADO: Buscar nombre de unidad desde el estado dinámico
        const unitName = units.find(u => u.id === registerData.organizationalUnitId)?.name || 'Unidad desconocida';
        const securityInfo = showSecurityQuestions && userSecurityQuestions.length > 0 
          ? ` con ${userSecurityQuestions.length} preguntas de seguridad` 
          : '';
        
        setMessage(`🎉 ¡Registro exitoso! Usuario ${result.data.firstName} creado en ${unitName}${securityInfo}`);
        
        // Limpiar formulario
        setRegisterData({
          email: '',
          password: '',
          firstName: '',
          lastName: '',
          organizationalUnitId: units.length > 0 ? units[0].id : 0
        });
        setUserSecurityQuestions([]);
        setShowSecurityQuestions(false);
        
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
        // Manejar errores HTTP específicos
        switch (response.status) {
          case 409:
            setMessage('👤 Este email ya está registrado. ¿Desea iniciar sesión en su lugar?');
            break;
          case 400:
            if (result.message?.includes('ya existe')) {
              setMessage('👤 El usuario ya existe: Este email ya está registrado en el sistema');
            } else if (result.message?.includes('organizacional')) {
              setMessage('🏢 Unidad organizacional inválida. Seleccione una opción válida');
            } else if (result.message?.includes('preguntas de seguridad')) {
              setMessage(`🔒 Error en preguntas de seguridad: ${result.message}`);
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
      <div className="max-w-lg w-full bg-white rounded-lg shadow-lg p-8">
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-green-600 rounded-full flex items-center justify-center mb-4">
            <span className="text-white text-2xl font-bold">📝</span>
          </div>
          <h1 className="text-2xl font-bold text-gray-900">Crear Cuenta</h1>
          <p className="text-gray-600">Sistema GAMC</p>
        </div>

        <form onSubmit={handleRegister} className="space-y-4">
          {/* Campos básicos */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Nombre <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={registerData.firstName}
                onChange={(e) => handleNameChange(e.target.value, 'firstName')}
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
                onChange={(e) => handleNameChange(e.target.value, 'lastName')}
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

          {/* 🔧 ACTUALIZADO: Select de unidades organizacionales dinámico */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Unidad Organizacional <span className="text-red-500">*</span>
            </label>
            <select
              value={registerData.organizationalUnitId}
              onChange={(e) => setRegisterData({ ...registerData, organizationalUnitId: parseInt(e.target.value) })}
              className={getInputClasses(registerValidation.organizationalUnit)}
              required
              disabled={loadingUnits}
            >
              <option value={0}>
                {loadingUnits ? '🔄 Cargando unidades...' : 'Seleccione una unidad organizacional'}
              </option>
              {units.map(unit => (
                <option key={unit.id} value={unit.id}>
                  {unit.name}
                </option>
              ))}
            </select>
            
            {/* 🔧 NUEVO: Mostrar estado de carga */}
            {loadingUnits && (
              <p className="text-xs text-blue-600 mt-1">
                🔄 Cargando unidades organizacionales...
              </p>
            )}
            
            {!loadingUnits && units.length > 0 && (
              <p className="text-xs text-gray-500 mt-1">
                ✅ Seleccione la unidad donde trabajará ({units.length} disponibles)
              </p>
            )}
            
            <FieldValidationMessage validation={registerValidation.organizationalUnit} />
          </div>

          {/* Sección de preguntas de seguridad */}
          <div className="border-t pt-4">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h3 className="text-sm font-medium text-gray-700">Preguntas de Seguridad</h3>
                <p className="text-xs text-gray-500">Opcional - Mejora la seguridad de tu cuenta</p>
              </div>
              <button
                type="button"
                onClick={() => {
                  setShowSecurityQuestions(!showSecurityQuestions);
                  if (!showSecurityQuestions) {
                    loadSecurityQuestions();
                  }
                }}
                className={`px-3 py-1 text-xs rounded-full transition-colors ${
                  showSecurityQuestions 
                    ? 'bg-green-100 text-green-800 hover:bg-green-200' 
                    : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
              >
                {showSecurityQuestions ? '🔒 Configurando' : '🔓 Configurar'}
              </button>
            </div>

            {showSecurityQuestions && (
              <div className="space-y-4 bg-gray-50 p-4 rounded-lg">
                {loadingQuestions ? (
                  <div className="text-center py-4">
                    <span className="text-gray-500">🔄 Cargando preguntas...</span>
                  </div>
                ) : (
                  <>
                    {userSecurityQuestions.map((userQuestion, index) => (
                      <div key={index} className="bg-white p-3 rounded-lg border">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm font-medium text-gray-700">
                            Pregunta {index + 1}
                          </span>
                          <button
                            type="button"
                            onClick={() => removeSecurityQuestion(index)}
                            className="text-red-500 hover:text-red-700 text-xs"
                          >
                            ❌ Eliminar
                          </button>
                        </div>
                        
                        <div className="space-y-2">
                          <select
                            value={userQuestion.questionId}
                            onChange={(e) => updateSecurityQuestion(index, 'questionId', parseInt(e.target.value))}
                            className={getInputClasses(securityValidation[`question_${index}`], 'text-sm')}
                          >
                            <option value={0}>Seleccione una pregunta...</option>
                            {getAvailableQuestions(index).map(q => (
                              <option key={q.id} value={q.id}>{q.questionText}</option>
                            ))}
                          </select>
                          <FieldValidationMessage validation={securityValidation[`question_${index}`]} />
                          
                          <input
                            type="text"
                            value={userQuestion.answer}
                            onChange={(e) => updateSecurityQuestion(index, 'answer', e.target.value)}
                            placeholder="Su respuesta..."
                            className={getInputClasses(securityValidation[`answer_${index}`], 'text-sm')}
                          />
                          <FieldValidationMessage validation={securityValidation[`answer_${index}`]} />
                        </div>
                      </div>
                    ))}

                    {userSecurityQuestions.length < 3 && (
                      <button
                        type="button"
                        onClick={addSecurityQuestion}
                        className="w-full py-2 px-4 border-2 border-dashed border-gray-300 rounded-lg text-gray-500 hover:border-green-400 hover:text-green-600 transition-colors text-sm"
                      >
                        ➕ Agregar pregunta de seguridad ({userSecurityQuestions.length}/3)
                      </button>
                    )}

                    {userSecurityQuestions.length > 0 && (
                      <div className="text-xs text-gray-500 bg-blue-50 p-2 rounded border-l-4 border-blue-400">
                        💡 Las preguntas de seguridad son opcionales pero recomendadas. Te ayudarán a recuperar tu cuenta si olvidas tu contraseña.
                      </div>
                    )}
                  </>
                )}
              </div>
            )}
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
            disabled={loading || !isFormValid || loadingUnits}
            className={`w-full py-3 px-4 rounded-lg font-semibold transition-colors ${
              loading || !isFormValid || loadingUnits
                ? 'bg-gray-400 cursor-not-allowed text-white' 
                : 'bg-green-600 hover:bg-green-700 text-white'
            }`}
          >
            {loading ? '🔄 Creando cuenta...' : loadingUnits ? '⏳ Cargando...' : '📝 Crear Cuenta'}
          </button>

          {(!isFormValid || loadingUnits) && (
            <p className="text-xs text-center text-gray-500">
              {loadingUnits 
                ? 'Cargando información necesaria...' 
                : 'Complete todos los campos correctamente para continuar'
              }
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
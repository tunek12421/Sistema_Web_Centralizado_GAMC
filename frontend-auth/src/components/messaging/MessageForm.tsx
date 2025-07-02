// frontend-auth/src/components/messaging/MessageForm.tsx
import React, { useState, useEffect } from 'react';
import { messageService } from '../../services/messageService';
import type { CreateMessageRequest } from '../../services/messageService';

interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

interface MessageFormProps {
  onSuccess: () => void;
  onCancel: () => void;
}

// 🔧 NUEVA INTERFAZ para unidades organizacionales
interface OrganizationalUnit {
  id: number;
  name: string;
  code: string;
  isActive: boolean;
}

const MessageForm: React.FC<MessageFormProps> = ({ onSuccess, onCancel }) => {
  const [formData, setFormData] = useState<CreateMessageRequest>({
    subject: '',
    content: '',
    receiverUnitId: 0, // 🔧 CAMBIADO: Inicializar en 0 hasta cargar las unidades
    messageTypeId: 1,
    priorityLevel: 1,
    isUrgent: false
  });

  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [validation, setValidation] = useState<{[key: string]: FieldValidation}>({});
  const [messageTypes, setMessageTypes] = useState<string[]>([]);
  const [loadingTypes, setLoadingTypes] = useState(true);

  // 🔧 NUEVO: Estados para unidades organizacionales dinámicas
  const [organizationalUnits, setOrganizationalUnits] = useState<OrganizationalUnit[]>([]);
  const [loadingUnits, setLoadingUnits] = useState(true);

  // 🔧 NUEVO: useEffect para cargar unidades organizacionales
  useEffect(() => {
    const loadOrganizationalUnits = async () => {
      try {
        setLoadingUnits(true);
        
        console.log('🔄 Cargando unidades organizacionales para mensajes...');
        const response = await fetch('http://localhost:3000/api/v1/organizational-units');
        const result = await response.json();
        
        if (response.ok && result.success && result.data) {
          console.log('✅ Unidades cargadas para mensajes:', result.data);
          setOrganizationalUnits(result.data);
        } else {
          // 🔧 FALLBACK: Usar datos hardcodeados con IDs correctos
          const fallbackUnits = [
            { id: 1, name: 'Obras Públicas', code: 'OBRAS_PUBLICAS', isActive: true },
            { id: 2, name: 'Monitoreo', code: 'MONITOREO', isActive: true },
            { id: 3, name: 'Movilidad Urbana', code: 'MOVILIDAD_URBANA', isActive: true },
            { id: 4, name: 'Gobierno Electrónico', code: 'GOBIERNO_ELECTRONICO', isActive: true },
            { id: 5, name: 'Prensa e Imagen', code: 'PRENSA_IMAGEN', isActive: true },
            { id: 6, name: 'Tecnología', code: 'TECNOLOGIA', isActive: true },
            { id: 7, name: 'Administración', code: 'ADMINISTRACION', isActive: true }
          ];
          setOrganizationalUnits(fallbackUnits);
        }
      } catch (error) {
        // 🔧 FALLBACK: Usar datos hardcodeados con IDs correctos
        const fallbackUnits = [
          { id: 1, name: 'Obras Públicas', code: 'OBRAS_PUBLICAS', isActive: true },
          { id: 2, name: 'Monitoreo', code: 'MONITOREO', isActive: true },
          { id: 3, name: 'Movilidad Urbana', code: 'MOVILIDAD_URBANA', isActive: true },
          { id: 4, name: 'Gobierno Electrónico', code: 'GOBIERNO_ELECTRONICO', isActive: true },
          { id: 5, name: 'Prensa e Imagen', code: 'PRENSA_IMAGEN', isActive: true },
          { id: 6, name: 'Tecnología', code: 'TECNOLOGIA', isActive: true },
          { id: 7, name: 'Administración', code: 'ADMINISTRACION', isActive: true }
        ];
        setOrganizationalUnits(fallbackUnits);
      } finally {
        setLoadingUnits(false);
      }
    };

    loadOrganizationalUnits();
  }, []);

  // Cargar tipos de mensajes al montar
  useEffect(() => {
    loadMessageTypes();
  }, []);

  // Validaciones en tiempo real
  useEffect(() => {
    validateForm();
  }, [formData, organizationalUnits]);

  // Cargar tipos de mensajes
  const loadMessageTypes = async () => {
    try {
      const types = await messageService.getMessageTypes();
      setMessageTypes(types);
    } catch (err) {
      console.error('Error loading message types:', err);
      setMessage('⚠️ Error al cargar tipos de mensaje. Usando valores por defecto.');
    } finally {
      setLoadingTypes(false);
    }
  };

  // Validar formulario
  const validateForm = () => {
    const newValidation: {[key: string]: FieldValidation} = {};

    // Validar asunto
    if (!formData.subject.trim()) {
      newValidation.subject = {
        isValid: false,
        message: 'El asunto es requerido',
        type: 'error'
      };
    } else if (formData.subject.length < 5) {
      newValidation.subject = {
        isValid: false,
        message: 'El asunto debe tener al menos 5 caracteres',
        type: 'warning'
      };
    } else if (formData.subject.length > 100) {
      newValidation.subject = {
        isValid: false,
        message: 'El asunto no puede superar los 100 caracteres',
        type: 'error'
      };
    } else {
      newValidation.subject = {
        isValid: true,
        message: `✓ Asunto válido (${formData.subject.length}/100)`,
        type: 'success'
      };
    }

    // Validar contenido
    if (!formData.content.trim()) {
      newValidation.content = {
        isValid: false,
        message: 'El contenido es requerido',
        type: 'error'
      };
    } else if (formData.content.length < 10) {
      newValidation.content = {
        isValid: false,
        message: 'El contenido debe tener al menos 10 caracteres',
        type: 'warning'
      };
    } else if (formData.content.length > 2000) {
      newValidation.content = {
        isValid: false,
        message: 'El contenido no puede superar los 2000 caracteres',
        type: 'error'
      };
    } else {
      newValidation.content = {
        isValid: true,
        message: `✓ Contenido válido (${formData.content.length}/2000)`,
        type: 'success'
      };
    }

    // 🔧 ACTUALIZADO: Validar unidad receptora usando datos dinámicos
    if (formData.receiverUnitId === 0) {
      newValidation.receiverUnitId = {
        isValid: false,
        message: 'Debe seleccionar una unidad destinataria',
        type: 'error'
      };
    } else {
      const unit = organizationalUnits.find(u => u.id === formData.receiverUnitId);
      newValidation.receiverUnitId = {
        isValid: true,
        message: `✓ Destinatario: ${unit?.name || 'Unidad válida'}`,
        type: 'success'
      };
    }

    // Validar tipo de mensaje
    if (formData.messageTypeId < 1 || formData.messageTypeId > messageTypes.length) {
      newValidation.messageTypeId = {
        isValid: false,
        message: 'Debe seleccionar un tipo de mensaje válido',
        type: 'error'
      };
    } else {
      const typeName = messageTypes[formData.messageTypeId - 1] || 'Tipo válido';
      newValidation.messageTypeId = {
        isValid: true,
        message: `✓ Tipo: ${typeName}`,
        type: 'success'
      };
    }

    setValidation(newValidation);
  };

  // Verificar si el formulario es válido
  const isFormValid = () => {
    return Object.values(validation).every(v => v.isValid) &&
           formData.subject.trim() &&
           formData.content.trim() &&
           formData.receiverUnitId > 0 &&
           !loadingUnits; // 🔧 NUEVO: No permitir envío mientras se cargan las unidades
  };

  // Manejar cambios en el formulario
  const handleInputChange = (field: keyof CreateMessageRequest, value: any) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  // Enviar formulario
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!isFormValid()) {
      setMessage('❌ Por favor complete todos los campos correctamente');
      return;
    }

    setLoading(true);
    setMessage('');

    try {
      // 🔧 NUEVO: Log de datos antes de enviar para debugging
      console.log('🔄 Enviando mensaje con datos:', formData);
      const selectedUnit = organizationalUnits.find(u => u.id === formData.receiverUnitId);
      console.log('📍 Unidad seleccionada:', selectedUnit);
      
      const createdMessage = await messageService.createMessage(formData);
      
      setMessage('✅ Mensaje enviado exitosamente');
      
      // Limpiar formulario
      setFormData({
        subject: '',
        content: '',
        receiverUnitId: 0,
        messageTypeId: 1,
        priorityLevel: 1,
        isUrgent: false
      });

      // Notificar éxito después de un breve delay
      setTimeout(() => {
        onSuccess();
      }, 1500);

    } catch (err: any) {
      console.error('Error creating message:', err);
      setMessage(`❌ Error al enviar mensaje: ${err.message}`);
    } finally {
      setLoading(false);
    }
  };

  // Componente para mostrar validación
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

  // Obtener clases CSS para inputs
  const getInputClasses = (field: string) => {
    const fieldValidation = validation[field];
    const baseClasses = 'w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent';
    
    if (!fieldValidation) return `${baseClasses} border-gray-300`;
    
    switch (fieldValidation.type) {
      case 'success':
        return `${baseClasses} border-green-300 bg-green-50`;
      case 'error':
        return `${baseClasses} border-red-300 bg-red-50`;
      case 'warning':
        return `${baseClasses} border-amber-300 bg-amber-50`;
      default:
        return `${baseClasses} border-gray-300`;
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-lg">
      <div className="border-b border-gray-200 px-6 py-4">
        <h2 className="text-xl font-semibold text-gray-900">Nuevo Mensaje</h2>
        <p className="text-sm text-gray-600 mt-1">
          Complete la información para enviar un mensaje a otra unidad organizacional
        </p>
      </div>

      <form onSubmit={handleSubmit} className="p-6">
        {/* Mensaje de estado */}
        {message && (
          <div className={`mb-4 p-3 rounded-lg ${
            message.includes('✅') ? 'bg-green-50 border border-green-200 text-green-700' :
            message.includes('❌') ? 'bg-red-50 border border-red-200 text-red-700' :
            'bg-blue-50 border border-blue-200 text-blue-700'
          }`}>
            {message}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Asunto */}
          <div className="md:col-span-2">
            <label htmlFor="subject" className="block text-sm font-medium text-gray-700 mb-2">
              Asunto *
            </label>
            <input
              type="text"
              id="subject"
              value={formData.subject}
              onChange={(e) => handleInputChange('subject', e.target.value)}
              placeholder="Ingrese el asunto del mensaje"
              className={getInputClasses('subject')}
              maxLength={100}
            />
            <FieldValidationMessage validation={validation.subject} />
          </div>

          {/* 🔧 ACTUALIZADO: Unidad destinataria dinámica */}
          <div>
            <label htmlFor="receiverUnitId" className="block text-sm font-medium text-gray-700 mb-2">
              Unidad destinataria *
            </label>
            <select
              id="receiverUnitId"
              value={formData.receiverUnitId}
              onChange={(e) => handleInputChange('receiverUnitId', parseInt(e.target.value))}
              className={getInputClasses('receiverUnitId')}
              disabled={loadingUnits}
            >
              <option value={0}>
                {loadingUnits ? '🔄 Cargando unidades...' : 'Seleccione una unidad'}
              </option>
              {organizationalUnits.map(unit => (
                <option key={unit.id} value={unit.id}>
                  {unit.name} ({unit.code})
                </option>
              ))}
            </select>
            
            {/* 🔧 NUEVO: Mostrar estado de carga */}
            {loadingUnits && (
              <p className="text-xs text-blue-600 mt-1">
                🔄 Cargando unidades organizacionales...
              </p>
            )}
            
            {!loadingUnits && organizationalUnits.length > 0 && (
              <p className="text-xs text-gray-500 mt-1">
                ✅ {organizationalUnits.length} unidades disponibles
              </p>
            )}
            
            <FieldValidationMessage validation={validation.receiverUnitId} />
          </div>

          {/* Tipo de mensaje */}
          <div>
            <label htmlFor="messageTypeId" className="block text-sm font-medium text-gray-700 mb-2">
              Tipo de mensaje *
            </label>
            <select
              id="messageTypeId"
              value={formData.messageTypeId}
              onChange={(e) => handleInputChange('messageTypeId', parseInt(e.target.value))}
              className={getInputClasses('messageTypeId')}
              disabled={loadingTypes}
            >
              {loadingTypes ? (
                <option>Cargando tipos...</option>
              ) : (
                messageTypes.map((type, index) => (
                  <option key={index} value={index + 1}>
                    {type}
                  </option>
                ))
              )}
            </select>
            <FieldValidationMessage validation={validation.messageTypeId} />
          </div>

          {/* Prioridad */}
          <div>
            <label htmlFor="priorityLevel" className="block text-sm font-medium text-gray-700 mb-2">
              Nivel de prioridad
            </label>
            <select
              id="priorityLevel"
              value={formData.priorityLevel}
              onChange={(e) => handleInputChange('priorityLevel', parseInt(e.target.value))}
              className={getInputClasses('priorityLevel')}
            >
              <option value={1}>🟢 Baja (1)</option>
              <option value={2}>🟡 Media (2)</option>
              <option value={3}>🟠 Alta (3)</option>
              <option value={4}>🔴 Crítica (4)</option>
              <option value={5}>🚨 Máxima (5)</option>
            </select>
          </div>

          {/* Urgente */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Marcas especiales
            </label>
            <div className="flex items-center">
              <input
                type="checkbox"
                id="isUrgent"
                checked={formData.isUrgent}
                onChange={(e) => handleInputChange('isUrgent', e.target.checked)}
                className="h-4 w-4 text-red-600 focus:ring-red-500 border-gray-300 rounded"
              />
              <label htmlFor="isUrgent" className="ml-2 text-sm text-gray-700">
                🚨 Marcar como urgente
              </label>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              Los mensajes urgentes se destacan visualmente y requieren atención inmediata
            </p>
          </div>
        </div>

        {/* Contenido */}
        <div className="mt-6">
          <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
            Contenido del mensaje *
          </label>
          <textarea
            id="content"
            value={formData.content}
            onChange={(e) => handleInputChange('content', e.target.value)}
            placeholder="Escriba aquí el contenido completo del mensaje..."
            rows={6}
            className={getInputClasses('content')}
            maxLength={2000}
          />
          <FieldValidationMessage validation={validation.content} />
        </div>

        {/* Preview del mensaje */}
        {formData.subject.trim() && formData.content.trim() && (
          <div className="mt-6 p-4 bg-gray-50 rounded-lg border">
            <h4 className="text-sm font-medium text-gray-700 mb-2">Vista previa:</h4>
            <div className="bg-white p-3 rounded border">
              <div className="flex items-center justify-between mb-2">
                <h5 className="font-medium text-gray-900">{formData.subject}</h5>
                <div className="flex items-center space-x-2">
                  {formData.isUrgent && <span className="text-red-600 text-sm">🚨 URGENTE</span>}
                  <span className="text-xs text-gray-500">
                    Prioridad: {formData.priorityLevel}
                  </span>
                </div>
              </div>
              <p className="text-sm text-gray-700 whitespace-pre-wrap">
                {formData.content}
              </p>
              {/* 🔧 NUEVO: Mostrar unidad de destino en preview */}
              {formData.receiverUnitId > 0 && (
                <div className="mt-2 pt-2 border-t text-xs text-gray-500">
                  📍 Destinatario: {organizationalUnits.find(u => u.id === formData.receiverUnitId)?.name}
                </div>
              )}
            </div>
          </div>
        )}

        {/* Botones */}
        <div className="mt-6 flex justify-end space-x-3">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
            disabled={loading}
          >
            Cancelar
          </button>
          <button
            type="submit"
            disabled={!isFormValid() || loading}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center"
          >
            {loading ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Enviando...
              </>
            ) : loadingUnits ? (
              <>
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Cargando...
              </>
            ) : (
              <>
                <span className="mr-2">📤</span>
                Enviar Mensaje
              </>
            )}
          </button>
        </div>
        
        {/* 🔧 NUEVO: Mensaje informativo mientras se carga */}
        {loadingUnits && (
          <p className="text-xs text-center text-gray-500 mt-2">
            Cargando información de unidades organizacionales...
          </p>
        )}
        
        {!isFormValid() && !loadingUnits && (
          <p className="text-xs text-center text-gray-500 mt-2">
            Complete todos los campos correctamente para enviar el mensaje
          </p>
        )}
      </form>
    </div>
  );
};

export default MessageForm;
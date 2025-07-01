// frontend-auth/src/components/messaging/MessageDetail.tsx
import React, { useState, useEffect } from 'react';
import { messageService } from '../../services/messageService';
import type { Message } from '../../services/messageService';

interface MessageDetailProps {
  message: Message;
  onBack: () => void;
  onMessageUpdated: () => void;
}

const MessageDetail: React.FC<MessageDetailProps> = ({ 
  message: initialMessage, 
  onBack, 
  onMessageUpdated 
}) => {
  const [message, setMessage] = useState<Message>(initialMessage);
  const [loading, setLoading] = useState(false);
  const [statusLoading, setStatusLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState('');
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [availableStatuses, setAvailableStatuses] = useState<string[]>([]);
  const [user, setUser] = useState<any>(null);

  // Cargar datos del usuario
  useEffect(() => {
    const userData = localStorage.getItem('user');
    if (userData) {
      setUser(JSON.parse(userData));
    }
  }, []);

  // Cargar estados disponibles
  useEffect(() => {
    loadAvailableStatuses();
    markAsReadIfNeeded();
  }, []);

  // Cargar estados disponibles
  const loadAvailableStatuses = async () => {
    try {
      const statuses = await messageService.getMessageStatuses();
      setAvailableStatuses(statuses);
    } catch (err) {
      console.error('Error loading statuses:', err);
    }
  };

  // Marcar como leído automáticamente si no está leído
  const markAsReadIfNeeded = async () => {
    if (!message.readAt) {
      try {
        await messageService.markAsRead(message.id);
        setMessage(prev => ({
          ...prev,
          readAt: new Date().toISOString()
        }));
        onMessageUpdated();
      } catch (err) {
        console.error('Error marking as read:', err);
      }
    }
  };

  // Actualizar estado del mensaje
  const updateMessageStatus = async (newStatus: string) => {
    setStatusLoading(true);
    setActionMessage('');

    try {
      await messageService.updateStatus(message.id, newStatus);
      
      // Actualizar estado local
      setMessage(prev => ({
        ...prev,
        status: {
          ...prev.status,
          name: newStatus
        }
      }));

      setActionMessage(`✅ Estado actualizado a: ${newStatus}`);
      onMessageUpdated();
      
      // Limpiar mensaje después de 3 segundos
      setTimeout(() => setActionMessage(''), 3000);
      
    } catch (err: any) {
      setActionMessage(`❌ Error al actualizar estado: ${err.message}`);
    } finally {
      setStatusLoading(false);
    }
  };

  // Eliminar mensaje
  const deleteMessage = async () => {
    setLoading(true);
    setActionMessage('');

    try {
      await messageService.deleteMessage(message.id);
      setActionMessage('✅ Mensaje eliminado exitosamente');
      
      // Volver a la lista después de 2 segundos
      setTimeout(() => {
        onMessageUpdated();
        onBack();
      }, 2000);
      
    } catch (err: any) {
      setActionMessage(`❌ Error al eliminar mensaje: ${err.message}`);
      setLoading(false);
    }
  };

  // Formatear fecha completa
  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'No disponible';
    
    const date = new Date(dateString);
    return date.toLocaleString('es-BO', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  // Obtener icono y color según prioridad
  const getPriorityInfo = (priority: number, isUrgent: boolean) => {
    if (isUrgent) return { icon: '🚨', color: 'text-red-600', bg: 'bg-red-100', label: 'URGENTE' };
    if (priority >= 4) return { icon: '🔴', color: 'text-red-600', bg: 'bg-red-100', label: 'Crítica' };
    if (priority === 3) return { icon: '🟠', color: 'text-orange-600', bg: 'bg-orange-100', label: 'Alta' };
    if (priority === 2) return { icon: '🟡', color: 'text-yellow-600', bg: 'bg-yellow-100', label: 'Media' };
    return { icon: '🟢', color: 'text-green-600', bg: 'bg-green-100', label: 'Baja' };
  };

  // Obtener icono de estado
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'SENT': return '📤';
      case 'READ': return '👁️';
      case 'IN_PROGRESS': return '🔄';
      case 'RESPONDED': return '💬';
      case 'RESOLVED': return '✅';
      case 'ARCHIVED': return '📋';
      case 'CANCELLED': return '❌';
      default: return '📄';
    }
  };

  // Verificar si el usuario puede editar el mensaje
  const canEditMessage = () => {
    return user && (
      user.role === 'admin' || 
      user.id === message.sender.id
    );
  };

  // Verificar si el usuario puede eliminar el mensaje
  const canDeleteMessage = () => {
    return user && (
      user.role === 'admin' || 
      user.id === message.sender.id
    );
  };

  const priorityInfo = getPriorityInfo(message.priorityLevel, message.isUrgent);

  return (
    <div className="space-y-6">
      {/* Mensaje de acción */}
      {actionMessage && (
        <div className={`p-4 rounded-lg ${
          actionMessage.includes('✅') ? 'bg-green-50 border border-green-200 text-green-700' :
          'bg-red-50 border border-red-200 text-red-700'
        }`}>
          {actionMessage}
        </div>
      )}

      {/* Header del mensaje */}
      <div className="bg-white rounded-lg shadow-lg overflow-hidden">
        <div className="border-b border-gray-200 px-6 py-4">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <div className="flex items-center mb-2">
                {message.isUrgent && (
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 mr-2">
                    🚨 URGENTE
                  </span>
                )}
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${priorityInfo.bg} ${priorityInfo.color}`}>
                  {priorityInfo.icon} Prioridad {priorityInfo.label}
                </span>
              </div>
              
              <h1 className="text-2xl font-bold text-gray-900 mb-2">
                {message.subject}
              </h1>
              
              <div className="flex items-center text-sm text-gray-600 space-x-4">
                <span className="flex items-center">
                  <span className="mr-1">{getStatusIcon(message.status.name)}</span>
                  Estado: {message.status.name}
                </span>
                <span>
                  ID: {message.id}
                </span>
                <span>
                  Tipo: {message.messageType.name}
                </span>
              </div>
            </div>

            <button
              onClick={onBack}
              className="bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 transition-colors flex items-center ml-4"
            >
              <span className="mr-2">←</span>
              Volver
            </button>
          </div>
        </div>

        {/* Información del remitente y destinatario */}
        <div className="px-6 py-4 bg-gray-50">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Remitente</h3>
              <div className="bg-white p-3 rounded border">
                <p className="font-medium text-gray-900">
                  {message.sender.firstName} {message.sender.lastName}
                </p>
                <p className="text-sm text-gray-600">{message.sender.email}</p>
                <p className="text-sm text-gray-600">
                  {message.senderUnit.name} ({message.senderUnit.code})
                </p>
              </div>
            </div>

            <div>
              <h3 className="text-sm font-medium text-gray-700 mb-2">Destinatario</h3>
              <div className="bg-white p-3 rounded border">
                <p className="font-medium text-gray-900">
                  {message.receiverUnit.name}
                </p>
                <p className="text-sm text-gray-600">
                  Código: {message.receiverUnit.code}
                </p>
                <p className="text-sm text-gray-600">
                  ID Unidad: {message.receiverUnit.id}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Contenido del mensaje */}
      <div className="bg-white rounded-lg shadow-lg p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Contenido del mensaje</h3>
        <div className="prose max-w-none">
          <div className="bg-gray-50 p-4 rounded border whitespace-pre-wrap text-gray-700">
            {message.content}
          </div>
        </div>
      </div>

      {/* Acciones y metadatos */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Metadatos */}
        <div className="lg:col-span-2 bg-white rounded-lg shadow-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Información adicional</h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <p className="text-sm font-medium text-gray-700">Fecha de creación</p>
              <p className="text-sm text-gray-600">{formatDate(message.createdAt)}</p>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-700">Última actualización</p>
              <p className="text-sm text-gray-600">{formatDate(message.updatedAt)}</p>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-700">Fecha de lectura</p>
              <p className="text-sm text-gray-600">
                {message.readAt ? formatDate(message.readAt) : 'No leído aún'}
              </p>
            </div>
            
            <div>
              <p className="text-sm font-medium text-gray-700">Fecha de respuesta</p>
              <p className="text-sm text-gray-600">
                {message.respondedAt ? formatDate(message.respondedAt) : 'Sin responder'}
              </p>
            </div>
          </div>

          {/* Archivos adjuntos */}
          <div className="mt-6">
            <p className="text-sm font-medium text-gray-700 mb-2">Archivos adjuntos</p>
            {message.attachments && message.attachments.length > 0 ? (
              <div className="space-y-2">
                {message.attachments.map((attachment, index) => (
                  <div key={index} className="flex items-center p-2 bg-gray-50 rounded border">
                    <span className="mr-2">📎</span>
                    <span className="text-sm text-gray-700">{attachment.name}</span>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 italic">Sin archivos adjuntos</p>
            )}
          </div>
        </div>

        {/* Panel de acciones */}
        <div className="bg-white rounded-lg shadow-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Acciones</h3>
          
          {/* Cambiar estado */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Cambiar estado
            </label>
            <select
              value={message.status.name}
              onChange={(e) => updateMessageStatus(e.target.value)}
              disabled={statusLoading}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              {availableStatuses.map((status, index) => (
                <option key={index} value={status}>
                  {getStatusIcon(status)} {status}
                </option>
              ))}
            </select>
            
            {statusLoading && (
              <div className="mt-2 flex items-center text-sm text-gray-500">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                Actualizando estado...
              </div>
            )}
          </div>

          {/* Botones de acción */}
          <div className="space-y-3">
            {canDeleteMessage() && (
              <>
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  disabled={loading}
                  className="w-full bg-red-600 text-white py-2 px-4 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center"
                >
                  {loading ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Eliminando...
                    </>
                  ) : (
                    <>
                      <span className="mr-2">🗑️</span>
                      Eliminar Mensaje
                    </>
                  )}
                </button>

                {/* Confirmación de eliminación */}
                {showDeleteConfirm && (
                  <div className="border border-red-200 bg-red-50 p-3 rounded">
                    <p className="text-sm text-red-800 mb-3">
                      ¿Está seguro de que desea eliminar este mensaje? Esta acción no se puede deshacer.
                    </p>
                    <div className="flex space-x-2">
                      <button
                        onClick={deleteMessage}
                        className="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700"
                      >
                        Sí, eliminar
                      </button>
                      <button
                        onClick={() => setShowDeleteConfirm(false)}
                        className="bg-gray-300 text-gray-700 px-3 py-1 rounded text-sm hover:bg-gray-400"
                      >
                        Cancelar
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}

            <button
              onClick={() => window.print()}
              className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors flex items-center justify-center"
            >
              <span className="mr-2">🖨️</span>
              Imprimir
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MessageDetail;
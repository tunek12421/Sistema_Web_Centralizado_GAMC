// frontend-auth/src/components/messaging/MessageList.tsx
import React, { useState, useEffect } from 'react';
import { messageService } from '../../services/messageService';
import type { Message, MessageFilters } from '../../services/messageService';

interface MessageListProps {
  onMessageSelect: (message: Message) => void;
  onCreateMessage: () => void;
  refreshTrigger?: number;
}

const MessageList: React.FC<MessageListProps> = ({ 
  onMessageSelect, 
  onCreateMessage,
  refreshTrigger = 0 
}) => {
  // Estados principales
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalMessages, setTotalMessages] = useState(0);

  // Estados para filtros
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [urgentFilter, setUrgentFilter] = useState<string>('');
  const [messageTypes, setMessageTypes] = useState<string[]>([]);
  const [messageStatuses, setMessageStatuses] = useState<string[]>([]);
  
  // Estado del usuario actual
  const [currentUser, setCurrentUser] = useState<any>(null);

  // Filtros base
  const [filters, setFilters] = useState<MessageFilters>({
    page: 1,
    limit: 10,
    sortBy: 'created_at',
    sortOrder: 'desc'
  });

  // Cargar datos iniciales
  useEffect(() => {
    // Cargar datos del usuario actual
    const userData = localStorage.getItem('user');
    if (userData) {
      try {
        const user = JSON.parse(userData);
        setCurrentUser(user);
      } catch (e) {
        console.error('Error parseando datos de usuario:', e);
      }
    }
    
    loadInitialData();
  }, []);

  // Cargar mensajes cuando cambian los filtros o refreshTrigger
  useEffect(() => {
    loadMessages();
  }, [filters, refreshTrigger]);

  // Función para cargar datos iniciales
  const loadInitialData = async () => {
    try {
      // Cargar en paralelo
      await Promise.allSettled([
        loadMessageTypes(),
        loadMessageStatuses()
      ]);
    } catch (error) {
      console.error('Error en carga inicial:', error);
    }
  };

  // Cargar tipos de mensajes
  const loadMessageTypes = async () => {
    try {
      const types = await messageService.getMessageTypes();
      setMessageTypes(types);
    } catch (err) {
      console.error('Error cargando tipos de mensajes:', err);
      // Usar fallback
      setMessageTypes(['Información', 'Coordinación', 'Urgente']);
    }
  };

  // Cargar estados de mensajes
  const loadMessageStatuses = async () => {
    try {
      const statuses = await messageService.getMessageStatuses();
      setMessageStatuses(statuses);
    } catch (err) {
      console.error('Error cargando estados de mensajes:', err);
      // Usar fallback
      setMessageStatuses(['Enviado', 'Leído', 'En Proceso', 'Resuelto']);
    }
  };

  // Función principal de carga de mensajes
  const loadMessages = async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await messageService.getMessages(filters);
      
      if (!response || !Array.isArray(response.messages)) {
        throw new Error('Formato de respuesta inválido');
      }
      
      // Actualizar estado
      setMessages(response.messages);
      setCurrentPage(response.page);
      setTotalPages(response.totalPages);
      setTotalMessages(response.total);
      
    } catch (err: any) {
      console.error('Error en loadMessages:', err);
      setError(err.message || 'Error al cargar mensajes');
    } finally {
      setLoading(false);
    }
  };

  // Funciones de filtro rápido
  const applyQuickFilter = (filterType: 'received' | 'sent' | 'all') => {
    if (!currentUser?.organizationalUnitId) {
      return;
    }

    const unitId = currentUser.organizationalUnitId;

    switch (filterType) {
      case 'received':
        setFilters(prev => ({
          ...prev,
          receiverUnitId: unitId,
          senderUnitId: undefined,
          page: 1
        }));
        break;
        
      case 'sent':
        setFilters(prev => ({
          ...prev,
          senderUnitId: unitId,
          receiverUnitId: undefined,
          page: 1
        }));
        break;
        
      case 'all':
        setFilters(prev => ({
          ...prev,
          receiverUnitId: undefined,
          senderUnitId: undefined,
          page: 1
        }));
        break;
    }
  };

  // Aplicar filtros
  const applyFilters = () => {
    const newFilters: MessageFilters = {
      ...filters,
      page: 1, // Reset page when applying filters
      searchText: searchText.trim() || undefined,
      status: statusFilter ? parseInt(statusFilter) : undefined,
      isUrgent: urgentFilter === 'true' ? true : urgentFilter === 'false' ? false : undefined
    };
    
    setFilters(newFilters);
  };

  // Limpiar filtros
  const clearFilters = () => {
    setSearchText('');
    setStatusFilter('');
    setUrgentFilter('');
    
    const baseFilters: MessageFilters = {
      page: 1,
      limit: 10,
      sortBy: 'created_at',
      sortOrder: 'desc'
    };
    
    setFilters(baseFilters);
  };

  // Cambiar página
  const changePage = (newPage: number) => {
    setFilters({ ...filters, page: newPage });
  };

  // Marcar mensaje como leído
  const handleMarkAsRead = async (message: Message, event: React.MouseEvent) => {
    event.stopPropagation();
    
    if (message.readAt) {
      return;
    }
    
    try {
      await messageService.markAsRead(message.id);
      
      // Update local state
      setMessages(messages.map(m => 
        m.id === message.id 
          ? { ...m, readAt: new Date().toISOString() }
          : m
      ));
      
    } catch (err: any) {
      console.error(`Error marcando mensaje ${message.id} como leído:`, err);
    }
  };

  // Determinar si es mensaje enviado o recibido
  const isMessageSent = (message: Message): boolean => {
    return currentUser && message.sender.id === currentUser.id;
  };

  // Obtener información de dirección del mensaje
  const getMessageDirection = (message: Message) => {
    if (isMessageSent(message)) {
      return {
        direction: 'sent',
        label: 'Para',
        unitName: message.receiverUnit.name,
        unitCode: message.receiverUnit.code,
        icon: '📤'
      };
    } else {
      return {
        direction: 'received',
        label: 'De',
        unitName: message.senderUnit.name,
        unitCode: message.senderUnit.code,
        personName: `${message.sender.firstName} ${message.sender.lastName}`,
        icon: '📥'
      };
    }
  };

  // Formatear fecha
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('es-BO', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  // Obtener clase CSS según prioridad
  const getPriorityClass = (priority: number, isUrgent: boolean) => {
    if (isUrgent) return 'border-l-4 border-red-500 bg-red-50';
    if (priority >= 3) return 'border-l-4 border-orange-500 bg-orange-50';
    if (priority === 2) return 'border-l-4 border-yellow-500 bg-yellow-50';
    return 'border-l-4 border-green-500 bg-green-50';
  };

  // Obtener icono de estado
  const getStatusIcon = (status: string) => {
    switch (status.toUpperCase()) {
      case 'SENT': 
      case 'ENVIADO': 
        return '📤';
      case 'READ': 
      case 'LEÍDO': 
        return '👁️';
      case 'IN_PROGRESS': 
      case 'EN_PROCESO': 
        return '🔄';
      case 'RESPONDED': 
      case 'RESPONDIDO': 
        return '💬';
      case 'RESOLVED': 
      case 'RESUELTO': 
        return '✅';
      case 'ARCHIVED': 
      case 'ARCHIVADO': 
        return '📋';
      case 'CANCELLED': 
      case 'CANCELADO': 
        return '❌';
      default: 
        return '📄';
    }
  };

  // Renderizado con loading
  if (loading && messages.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
            <span className="mt-3 text-gray-600 block">Cargando mensajes...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header con filtros */}
      <div className="border-b border-gray-200 p-6">
        <div className="flex justify-between items-center mb-4">
          <div className="flex items-center space-x-4">
            <h2 className="text-xl font-semibold text-gray-900">
              Mensajes ({totalMessages})
            </h2>
            
            {/* Botones de filtro rápido */}
            {currentUser?.organizationalUnitId && (
              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-600">Vista:</span>
                <button
                  onClick={() => applyQuickFilter('received')}
                  className={`px-3 py-1 rounded text-sm transition-colors ${
                    filters.receiverUnitId === currentUser.organizationalUnitId && !filters.senderUnitId
                      ? 'bg-blue-600 text-white' 
                      : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                  }`}
                >
                  📥 Recibidos
                </button>
                
                <button
                  onClick={() => applyQuickFilter('sent')}
                  className={`px-3 py-1 rounded text-sm transition-colors ${
                    filters.senderUnitId === currentUser.organizationalUnitId && !filters.receiverUnitId
                      ? 'bg-green-600 text-white' 
                      : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                  }`}
                >
                  📤 Enviados
                </button>
                
                <button
                  onClick={() => applyQuickFilter('all')}
                  className={`px-3 py-1 rounded text-sm transition-colors ${
                    !filters.receiverUnitId && !filters.senderUnitId
                      ? 'bg-purple-600 text-white' 
                      : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                  }`}
                >
                  🌐 Todos
                </button>
              </div>
            )}
          </div>
          
          <button
            onClick={onCreateMessage}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors flex items-center"
          >
            <span className="mr-2">✉️</span>
            Nuevo Mensaje
          </button>
        </div>

        {/* Filtros */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
          <input
            type="text"
            placeholder="Buscar mensajes..."
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Todos los estados</option>
            {messageStatuses.map((status, index) => (
              <option key={index} value={index + 1}>
                {getStatusIcon(status)} {status}
              </option>
            ))}
          </select>

          <select
            value={urgentFilter}
            onChange={(e) => setUrgentFilter(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Todas las prioridades</option>
            <option value="true">🚨 Solo urgentes</option>
            <option value="false">📝 No urgentes</option>
          </select>

          <div className="flex gap-2">
            <button
              onClick={applyFilters}
              className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors flex-1"
            >
              Filtrar
            </button>
            <button
              onClick={clearFilters}
              className="bg-gray-500 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition-colors"
            >
              Limpiar
            </button>
          </div>
        </div>
      </div>

      {/* Lista de mensajes */}
      <div className="divide-y divide-gray-200">
        {error && (
          <div className="p-4 bg-red-50 border-l-4 border-red-500 text-red-700">
            <p className="font-medium">Error:</p>
            <p>{error}</p>
            <button
              onClick={loadMessages}
              className="mt-2 text-sm text-red-600 hover:text-red-800 underline"
            >
              🔄 Reintentar
            </button>
          </div>
        )}

        {messages.length === 0 && !loading ? (
          <div className="p-8 text-center text-gray-500">
            <span className="text-4xl mb-4 block">📭</span>
            <p className="text-lg mb-2">No hay mensajes</p>
            <p>Los mensajes aparecerán aquí cuando los reciba</p>
            
            {/* Sugerencias útiles */}
            <div className="mt-4 p-3 bg-blue-50 rounded-lg text-sm">
              <p className="font-medium text-blue-800">💡 Sugerencias:</p>
              <ul className="text-blue-700 mt-1">
                {(filters.receiverUnitId || filters.senderUnitId) && (
                  <li>• Prueba hacer clic en "🌐 Todos" para ver todos los mensajes</li>
                )}
                {filters.searchText && (
                  <li>• Prueba con términos de búsqueda diferentes</li>
                )}
                <li>• Verifica que hay mensajes en el sistema</li>
                <li>• Usa los botones de filtro rápido arriba</li>
              </ul>
              <button
                onClick={clearFilters}
                className="mt-2 text-blue-600 hover:text-blue-800 underline"
              >
                Limpiar todos los filtros
              </button>
            </div>
          </div>
        ) : (
          messages.map((message) => {
            const messageDirection = getMessageDirection(message);
            
            return (
              <div
                key={message.id}
                onClick={() => onMessageSelect(message)}
                className={`p-4 hover:bg-gray-50 cursor-pointer transition-colors ${
                  getPriorityClass(message.priorityLevel, message.isUrgent)
                } ${!message.readAt && !isMessageSent(message) ? 'font-semibold' : ''}`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center mb-2">
                      <span className="text-sm text-gray-600 mr-2">
                        {messageDirection.icon}
                      </span>
                      
                      {message.isUrgent && (
                        <span className="text-red-600 mr-2 text-sm font-bold">🚨 URGENTE</span>
                      )}
                      <h3 className="text-sm font-medium text-gray-900 truncate">
                        {message.subject}
                      </h3>
                    </div>
                    
                    <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                      {message.content}
                    </p>
                    
                    <div className="flex items-center text-xs text-gray-500">
                      <span>
                        {messageDirection.label}: {messageDirection.unitName}
                        {messageDirection.personName && ` (${messageDirection.personName})`}
                      </span>
                      <span className="mx-2">•</span>
                      <span>{formatDate(message.createdAt)}</span>
                      <span className="mx-2">•</span>
                      <span>ID: {message.id}</span>
                    </div>
                  </div>

                  <div className="flex items-center ml-4">
                    <div className="text-right mr-3">
                      <div className="flex items-center text-sm text-gray-600">
                        <span className="mr-1">
                          {getStatusIcon(message.status.name)}
                        </span>
                        <span>{message.status.name}</span>
                      </div>
                      
                      {!message.readAt && !isMessageSent(message) && (
                        <button
                          onClick={(e) => handleMarkAsRead(message, e)}
                          className="text-xs text-blue-600 hover:text-blue-800 mt-1"
                        >
                          Marcar como leído
                        </button>
                      )}
                    </div>
                    
                    <div className="text-gray-400">
                      <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clipRule="evenodd" />
                      </svg>
                    </div>
                  </div>
                </div>
              </div>
            );
          })
        )}

        {loading && messages.length > 0 && (
          <div className="p-4 text-center text-gray-500">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div>
            <span className="mt-2 block">Actualizando...</span>
          </div>
        )}
      </div>

      {/* Paginación */}
      {totalPages > 1 && (
        <div className="border-t border-gray-200 px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-700">
              Página {currentPage} de {totalPages} ({totalMessages} mensajes total)
            </div>
            
            <div className="flex space-x-2">
              <button
                onClick={() => changePage(currentPage - 1)}
                disabled={currentPage === 1}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Anterior
              </button>
              
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const pageNum = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i;
                return (
                  <button
                    key={pageNum}
                    onClick={() => changePage(pageNum)}
                    className={`px-3 py-1 border rounded-md text-sm ${
                      pageNum === currentPage
                        ? 'bg-blue-600 text-white border-blue-600'
                        : 'border-gray-300 text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    {pageNum}
                  </button>
                );
              })}
              
              <button
                onClick={() => changePage(currentPage + 1)}
                disabled={currentPage === totalPages}
                className="px-3 py-1 border border-gray-300 rounded-md text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Siguiente
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default MessageList;
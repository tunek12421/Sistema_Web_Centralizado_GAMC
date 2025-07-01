// frontend-auth/src/components/messaging/MessageList.tsx
import React, { useState, useEffect } from 'react';
import { messageService } from '../../services/messageService';
import type { Message, MessageFilters } from '../../services/messageService';

interface MessageListProps {
  onMessageSelect: (message: Message) => void;
  onCreateMessage: () => void;
  refreshTrigger?: number; // Para forzar refresh desde componente padre
}

const MessageList: React.FC<MessageListProps> = ({ 
  onMessageSelect, 
  onCreateMessage,
  refreshTrigger = 0 
}) => {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalMessages, setTotalMessages] = useState(0);
  const [filters, setFilters] = useState<MessageFilters>({
    page: 1,
    limit: 10,
    sortBy: 'created_at',
    sortOrder: 'desc'
  });

  // Estados para filtros
  const [searchText, setSearchText] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [urgentFilter, setUrgentFilter] = useState<string>('');
  const [messageTypes, setMessageTypes] = useState<string[]>([]);
  const [messageStatuses, setMessageStatuses] = useState<string[]>([]);

  // Cargar datos iniciales
  useEffect(() => {
    loadMessageTypes();
    loadMessageStatuses();
  }, []);

  // Cargar mensajes cuando cambian los filtros o refreshTrigger
  useEffect(() => {
    loadMessages();
  }, [filters, refreshTrigger]);

  // Cargar tipos de mensajes
  const loadMessageTypes = async () => {
    try {
      const types = await messageService.getMessageTypes();
      setMessageTypes(types);
    } catch (err) {
      console.error('Error loading message types:', err);
    }
  };

  // Cargar estados de mensajes
  const loadMessageStatuses = async () => {
    try {
      const statuses = await messageService.getMessageStatuses();
      setMessageStatuses(statuses);
    } catch (err) {
      console.error('Error loading message statuses:', err);
    }
  };

  // Cargar mensajes
  const loadMessages = async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await messageService.getMessages(filters);
      setMessages(response.messages);
      setCurrentPage(response.page);
      setTotalPages(response.totalPages);
      setTotalMessages(response.total);
    } catch (err: any) {
      setError(err.message || 'Error al cargar mensajes');
      console.error('Error loading messages:', err);
    } finally {
      setLoading(false);
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
    setFilters({
      page: 1,
      limit: 10,
      sortBy: 'created_at',
      sortOrder: 'desc'
    });
  };

  // Cambiar página
  const changePage = (newPage: number) => {
    setFilters({ ...filters, page: newPage });
  };

  // Marcar mensaje como leído
  const handleMarkAsRead = async (message: Message, event: React.MouseEvent) => {
    event.stopPropagation(); // Prevent message selection
    
    if (message.readAt) return; // Already read
    
    try {
      await messageService.markAsRead(message.id);
      // Update local state
      setMessages(messages.map(m => 
        m.id === message.id 
          ? { ...m, readAt: new Date().toISOString() }
          : m
      ));
    } catch (err: any) {
      console.error('Error marking as read:', err);
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

  if (loading && messages.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <span className="ml-3 text-gray-600">Cargando mensajes...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header con filtros */}
      <div className="border-b border-gray-200 p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-semibold text-gray-900">
            Mensajes ({totalMessages})
          </h2>
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
          </div>
        )}

        {messages.length === 0 && !loading ? (
          <div className="p-8 text-center text-gray-500">
            <span className="text-4xl mb-4 block">📭</span>
            <p className="text-lg mb-2">No hay mensajes</p>
            <p>Los mensajes aparecerán aquí cuando los reciba</p>
          </div>
        ) : (
          messages.map((message) => (
            <div
              key={message.id}
              onClick={() => onMessageSelect(message)}
              className={`p-4 hover:bg-gray-50 cursor-pointer transition-colors ${
                getPriorityClass(message.priorityLevel, message.isUrgent)
              } ${!message.readAt ? 'font-semibold' : ''}`}
            >
              <div className="flex items-start justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center mb-2">
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
                    <span>De: {message.sender.firstName} {message.sender.lastName}</span>
                    <span className="mx-2">•</span>
                    <span>{message.senderUnit.name}</span>
                    <span className="mx-2">•</span>
                    <span>{formatDate(message.createdAt)}</span>
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
                    
                    {!message.readAt && (
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
          ))
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
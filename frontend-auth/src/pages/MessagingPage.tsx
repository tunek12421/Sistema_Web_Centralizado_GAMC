// frontend-auth/src/pages/MessagingPage.tsx
import React, { useState, useEffect } from 'react';
import MessageList from '../components/messaging/MessageList';
import MessageForm from '../components/messaging/MessageForm';
import MessageDetail from '../components/messaging/MessageDetail';
import type { Message } from '../services/messageService';

interface MessagingPageProps {
  onBack: () => void;
}

type MessagingView = 'list' | 'create' | 'detail';

// 🔧 NUEVA INTERFAZ para estadísticas
interface MessageStats {
  totalMessages: number;
  readMessages: number;
  inProgressMessages: number;
  urgentMessages: number;
  messagesByStatus: {
    [key: string]: number;
  };
  averageResponseTime: string;
}

const MessagingPage: React.FC<MessagingPageProps> = ({ onBack }) => {
  const [currentView, setCurrentView] = useState<MessagingView>('list');
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [user, setUser] = useState<any>(null);
  
  // 🔧 NUEVO: Estados para estadísticas
  const [stats, setStats] = useState<MessageStats>({
    totalMessages: 0,
    readMessages: 0,
    inProgressMessages: 0,
    urgentMessages: 0,
    messagesByStatus: {},
    averageResponseTime: '0 horas'
  });
  const [loadingStats, setLoadingStats] = useState(true);
  const [statsError, setStatsError] = useState<string>('');

  // Cargar datos del usuario
  useEffect(() => {
    const userData = localStorage.getItem('user');
    if (userData) {
      setUser(JSON.parse(userData));
    }
  }, []);

  // 🔧 NUEVO: Cargar estadísticas cuando se monta el componente o se actualiza
  useEffect(() => {
    if (currentView === 'list') {
      loadMessageStats();
    }
  }, [currentView, refreshTrigger]);

  // 🔧 NUEVA FUNCIÓN: Cargar estadísticas desde la API
  const loadMessageStats = async () => {
    try {
      setLoadingStats(true);
      setStatsError('');

      // 🔧 CORREGIDO: Buscar token con diferentes claves posibles
      let token = localStorage.getItem('accessToken') || 
                  localStorage.getItem('token') || 
                  localStorage.getItem('authToken');
      
      // 🔧 DEBUGGING: Mostrar información de autenticación
      console.log('🔐 Debugging autenticación:');
      console.log('- accessToken:', localStorage.getItem('accessToken') ? 'Existe' : 'No existe');
      console.log('- token:', localStorage.getItem('token') ? 'Existe' : 'No existe');
      console.log('- user:', localStorage.getItem('user') ? 'Existe' : 'No existe');
      console.log('- Todas las claves en localStorage:', Object.keys(localStorage));
      
      if (!token) {
        throw new Error('No hay token de autenticación. Inicie sesión nuevamente.');
      }

      console.log('🔄 Cargando estadísticas de mensajes...');
      const response = await fetch('http://localhost:3000/api/v1/messages/stats', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error(`Error ${response.status}: ${response.statusText}`);
      }

      const result = await response.json();
      console.log('📊 Estadísticas recibidas:', result);

      if (result.success && result.data) {
        // Procesar datos de estadísticas
        const statsData = result.data;
        setStats({
          totalMessages: statsData.totalMessages || 0,
          readMessages: statsData.messagesByStatus?.READ || 0,
          inProgressMessages: statsData.messagesByStatus?.IN_PROGRESS || 0,
          urgentMessages: statsData.urgentMessages || 0,
          messagesByStatus: statsData.messagesByStatus || {},
          averageResponseTime: statsData.averageResponseTime || '0 horas'
        });
      } else {
        console.warn('⚠️ Respuesta sin datos válidos:', result);
        setStatsError('Datos de estadísticas no válidos');
      }
    } catch (error: any) {
      console.error('❌ Error cargando estadísticas:', error);
      setStatsError(error.message || 'Error al cargar estadísticas');
      
      // 🔧 FALLBACK: Usar estadísticas básicas calculadas desde la lista de mensajes
      try {
        await loadBasicStats();
      } catch (fallbackError) {
        console.error('❌ Error en fallback de estadísticas:', fallbackError);
      }
    } finally {
      setLoadingStats(false);
    }
  };

  // 🔧 NUEVA FUNCIÓN: Fallback - cargar estadísticas básicas desde la lista de mensajes
  const loadBasicStats = async () => {
    try {
      // 🔧 CORREGIDO: Usar el mismo método para obtener token
      let token = localStorage.getItem('accessToken') || 
                  localStorage.getItem('token') || 
                  localStorage.getItem('authToken');
                  
      if (!token) return;

      const response = await fetch('http://localhost:3000/api/v1/messages?page=1&limit=100', {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const result = await response.json();
        if (result.success && result.data && result.data.messages) {
          const messages = result.data.messages;
          
          // Calcular estadísticas básicas
          const total = messages.length;
          const read = messages.filter((m: any) => m.readAt).length;
          const inProgress = messages.filter((m: any) => m.status?.code === 'IN_PROGRESS').length;
          const urgent = messages.filter((m: any) => m.isUrgent).length;

          setStats({
            totalMessages: total,
            readMessages: read,
            inProgressMessages: inProgress,
            urgentMessages: urgent,
            messagesByStatus: {},
            averageResponseTime: 'N/A'
          });

          console.log('📊 Estadísticas básicas calculadas:', { total, read, inProgress, urgent });
        }
      }
    } catch (error) {
      console.error('❌ Error en estadísticas básicas:', error);
    }
  };

  // Forzar refresh de la lista de mensajes
  const triggerRefresh = () => {
    setRefreshTrigger(prev => prev + 1);
  };

  // Manejar selección de mensaje
  const handleMessageSelect = (message: Message) => {
    setSelectedMessage(message);
    setCurrentView('detail');
  };

  // Manejar creación exitosa de mensaje
  const handleCreateSuccess = () => {
    setCurrentView('list');
    triggerRefresh(); // Refresh la lista para mostrar el nuevo mensaje
  };

  // Manejar cancelación de creación
  const handleCreateCancel = () => {
    setCurrentView('list');
  };

  // Volver a la lista desde detalle
  const handleBackToList = () => {
    setSelectedMessage(null);
    setCurrentView('list');
    triggerRefresh(); // Refresh por si hubo cambios en el mensaje
  };

  // Mostrar crear mensaje
  const showCreateMessage = () => {
    setCurrentView('create');
  };

  // Obtener título según la vista actual
  const getPageTitle = () => {
    switch (currentView) {
      case 'create':
        return 'Crear Nuevo Mensaje';
      case 'detail':
        return selectedMessage ? `Mensaje: ${selectedMessage.subject}` : 'Detalle del Mensaje';
      default:
        return 'Sistema de Mensajería';
    }
  };

  // Obtener breadcrumb
  const getBreadcrumb = () => {
    const items = [
      { label: 'Dashboard', onClick: onBack },
      { label: 'Mensajería', onClick: currentView !== 'list' ? () => setCurrentView('list') : undefined }
    ];

    if (currentView === 'create') {
      items.push({ label: 'Nuevo Mensaje', onClick: undefined });
    } else if (currentView === 'detail' && selectedMessage) {
      items.push({ label: 'Detalle del Mensaje', onClick: undefined });
    }

    return items;
  };

  // 🔧 NUEVA FUNCIÓN: Obtener valor de estadística con fallback
  const getStatValue = (value: number, loading: boolean = false) => {
    if (loading) return '⏳';
    if (statsError) return '❌';
    return value.toString();
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div>
              <nav className="flex space-x-2 text-sm text-gray-500 mb-1">
                {getBreadcrumb().map((item, index) => (
                  <React.Fragment key={index}>
                    {index > 0 && <span>›</span>}
                    {item.onClick ? (
                      <button
                        onClick={item.onClick}
                        className="hover:text-gray-700 transition-colors"
                      >
                        {item.label}
                      </button>
                    ) : (
                      <span className="text-gray-900 font-medium">{item.label}</span>
                    )}
                  </React.Fragment>
                ))}
              </nav>
              <h1 className="text-2xl font-bold text-gray-900">
                {getPageTitle()}
              </h1>
            </div>

            <div className="flex items-center space-x-4">
              {/* Información del usuario */}
              {user && (
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900">
                    {user.firstName} {user.lastName}
                  </p>
                  <p className="text-xs text-gray-500">
                    {user.organizationalUnit?.name}
                  </p>
                </div>
              )}

              {/* Botón de regreso */}
              <button
                onClick={onBack}
                className="bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700 transition-colors flex items-center"
              >
                <span className="mr-2">🏠</span>
                Volver al Dashboard
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Contenido principal */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Vista de lista de mensajes */}
        {currentView === 'list' && (
          <div className="space-y-6">
            {/* 🔧 ACTUALIZADO: Estadísticas dinámicas */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex items-center">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <span className="text-blue-600 text-xl">📨</span>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">Total</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {getStatValue(stats.totalMessages, loadingStats)}
                    </p>
                    {loadingStats && (
                      <p className="text-xs text-blue-500">Cargando...</p>
                    )}
                    {statsError && (
                      <p className="text-xs text-red-500">Error al cargar</p>
                    )}
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex items-center">
                  <div className="p-2 bg-green-100 rounded-lg">
                    <span className="text-green-600 text-xl">👁️</span>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">Leídos</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {getStatValue(stats.readMessages, loadingStats)}
                    </p>
                    {!loadingStats && !statsError && stats.totalMessages > 0 && (
                      <p className="text-xs text-gray-500">
                        {Math.round((stats.readMessages / stats.totalMessages) * 100)}% del total
                      </p>
                    )}
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex items-center">
                  <div className="p-2 bg-yellow-100 rounded-lg">
                    <span className="text-yellow-600 text-xl">🔄</span>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">En proceso</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {getStatValue(stats.inProgressMessages, loadingStats)}
                    </p>
                    {!loadingStats && !statsError && stats.totalMessages > 0 && (
                      <p className="text-xs text-gray-500">
                        {Math.round((stats.inProgressMessages / stats.totalMessages) * 100)}% del total
                      </p>
                    )}
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex items-center">
                  <div className="p-2 bg-red-100 rounded-lg">
                    <span className="text-red-600 text-xl">🚨</span>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">Urgentes</p>
                    <p className="text-2xl font-bold text-gray-900">
                      {getStatValue(stats.urgentMessages, loadingStats)}
                    </p>
                    {!loadingStats && !statsError && stats.totalMessages > 0 && (
                      <p className="text-xs text-gray-500">
                        {Math.round((stats.urgentMessages / stats.totalMessages) * 100)}% del total
                      </p>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* 🔧 NUEVO: Mostrar información adicional de estadísticas */}
            {!loadingStats && !statsError && stats.averageResponseTime && stats.averageResponseTime !== '0 horas' && (
              <div className="bg-white rounded-lg shadow p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-600">Tiempo promedio de respuesta</p>
                    <p className="text-lg font-bold text-blue-600">{stats.averageResponseTime}</p>
                  </div>
                  <div className="text-right">
                    <button
                      onClick={loadMessageStats}
                      className="text-sm text-blue-600 hover:text-blue-800 transition-colors"
                      disabled={loadingStats}
                    >
                      🔄 Actualizar estadísticas
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Lista de mensajes */}
            <MessageList
              onMessageSelect={handleMessageSelect}
              onCreateMessage={showCreateMessage}
              refreshTrigger={refreshTrigger}
            />
          </div>
        )}

        {/* Vista de creación de mensaje */}
        {currentView === 'create' && (
          <MessageForm
            onSuccess={handleCreateSuccess}
            onCancel={handleCreateCancel}
          />
        )}

        {/* Vista de detalle de mensaje */}
        {currentView === 'detail' && selectedMessage && (
          <MessageDetail
            message={selectedMessage}
            onBack={handleBackToList}
            onMessageUpdated={triggerRefresh}
          />
        )}
      </div>

      {/* Footer */}
      <footer className="bg-white border-t mt-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-between items-center text-sm text-gray-500">
            <div>
              Sistema de Mensajería - GAMC Quillacollo
            </div>
            <div>
              Versión 1.0 - {new Date().getFullYear()}
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default MessagingPage;
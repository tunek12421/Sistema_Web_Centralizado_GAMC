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

const MessagingPage: React.FC<MessagingPageProps> = ({ onBack }) => {
  const [currentView, setCurrentView] = useState<MessagingView>('list');
  const [selectedMessage, setSelectedMessage] = useState<Message | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [user, setUser] = useState<any>(null);

  // Cargar datos del usuario
  useEffect(() => {
    const userData = localStorage.getItem('user');
    if (userData) {
      setUser(JSON.parse(userData));
    }
  }, []);

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
            {/* Estadísticas rápidas */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              <div className="bg-white rounded-lg shadow p-6">
                <div className="flex items-center">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <span className="text-blue-600 text-xl">📨</span>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-600">Total</p>
                    <p className="text-2xl font-bold text-gray-900">--</p>
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
                    <p className="text-2xl font-bold text-gray-900">--</p>
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
                    <p className="text-2xl font-bold text-gray-900">--</p>
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
                    <p className="text-2xl font-bold text-gray-900">--</p>
                  </div>
                </div>
              </div>
            </div>

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
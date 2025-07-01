import React, { useEffect, useState } from 'react';

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
  isActive: boolean;
  lastLogin?: Date;
  createdAt: Date;
}

// ✨ ACTUALIZADO - Agregada prop onGoToMessaging
interface DashboardProps {
  onLogout: () => void;
  onGoToMessaging: () => void; // ✨ NUEVO PROP
}

const DashboardPage: React.FC<DashboardProps> = ({ onLogout, onGoToMessaging }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Cargar datos del usuario desde localStorage
    const userData = localStorage.getItem('user');
    if (userData) {
      setUser(JSON.parse(userData));
    }
    setLoading(false);
  }, []);

  const handleLogout = async () => {
    try {
      // Limpiar localStorage
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('user');
      
      // Llamar al logout del backend (opcional)
      const token = localStorage.getItem('accessToken');
      if (token) {
        await fetch('http://localhost:3000/api/v1/auth/logout', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        });
      }
      
      onLogout();
    } catch (error) {
      console.error('Error en logout:', error);
      // Limpiar de todas formas
      onLogout();
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Cargando dashboard...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600">Error: No se pudo cargar la información del usuario</p>
          <button 
            onClick={onLogout}
            className="mt-4 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
          >
            Volver al login
          </button>
        </div>
      </div>
    );
  }

  const getRoleDisplayName = (role: string) => {
    switch (role) {
      case 'admin': return 'Administrador';
      case 'input': return 'Usuario de Entrada';
      case 'output': return 'Usuario de Salida';
      default: return role;
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'admin': return 'bg-purple-100 text-purple-800';
      case 'input': return 'bg-green-100 text-green-800';
      case 'output': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <div className="h-8 w-8 bg-blue-600 rounded-full flex items-center justify-center mr-3">
                <span className="text-white text-sm font-bold">G</span>
              </div>
              <h1 className="text-xl font-semibold text-gray-900">GAMC Dashboard</h1>
            </div>
            
            <div className="flex items-center space-x-4">
              <div className="text-sm text-gray-600">
                Bienvenido, <span className="font-medium">{user.firstName}</span>
              </div>
              <button
                onClick={handleLogout}
                className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors text-sm"
              >
                🚪 Cerrar Sesión
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          
          {/* Welcome Card */}
          <div className="bg-white overflow-hidden shadow rounded-lg mb-6">
            <div className="bg-gradient-to-r from-blue-500 to-blue-600 px-6 py-4">
              <h2 className="text-xl font-bold text-white">
                ¡Bienvenido al Sistema GAMC! 🎉
              </h2>
              <p className="text-blue-100 mt-1">
                Has iniciado sesión exitosamente en el sistema de gestión municipal
              </p>
            </div>
            
            <div className="px-6 py-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div>
                  <h3 className="text-lg font-medium text-gray-900 mb-3">Información Personal</h3>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Nombre completo:</span>
                      <span className="font-medium">{user.firstName} {user.lastName}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Email:</span>
                      <span className="font-medium">{user.email}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Rol:</span>
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${getRoleColor(user.role)}`}>
                        {getRoleDisplayName(user.role)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Estado:</span>
                      <span className="text-green-600 font-medium">✅ Activo</span>
                    </div>
                  </div>
                </div>
                
                <div>
                  <h3 className="text-lg font-medium text-gray-900 mb-3">Unidad Organizacional</h3>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Unidad:</span>
                      <span className="font-medium">{user.organizationalUnit.name}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Código:</span>
                      <span className="font-medium text-gray-700">{user.organizationalUnit.code}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Miembro desde:</span>
                      <span className="font-medium">
                        {new Date(user.createdAt).toLocaleDateString('es-BO')}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
            
            {/* ✨ CARD DE MENSAJES - ACTUALIZADO para ser funcional */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="h-10 w-10 bg-green-100 rounded-full flex items-center justify-center">
                      <span className="text-green-600 text-xl">📝</span>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Mensajes</h3>
                    <p className="text-sm text-gray-600">Gestionar comunicaciones</p>
                  </div>
                </div>
                <div className="mt-4">
                  {/* ✨ ACTUALIZADO - Hacer funcional el botón */}
                  <button 
                    onClick={onGoToMessaging}
                    className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 transition-colors flex items-center justify-center"
                  >
                    <span className="mr-2">📨</span>
                    Ir a Mensajería
                  </button>
                </div>
                
                {/* ✨ NUEVO - Información adicional según el rol */}
                <div className="mt-3 text-xs text-gray-500">
                  {user.role === 'input' && (
                    <p>✓ Puede crear y enviar mensajes</p>
                  )}
                  {user.role === 'output' && (
                    <p>✓ Puede recibir y gestionar mensajes</p>
                  )}
                  {user.role === 'admin' && (
                    <p>✓ Acceso completo al sistema de mensajería</p>
                  )}
                </div>
              </div>
            </div>

            {/* Card de Reportes - Sin cambios */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="h-10 w-10 bg-blue-100 rounded-full flex items-center justify-center">
                      <span className="text-blue-600 text-xl">📊</span>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Reportes</h3>
                    <p className="text-sm text-gray-600">Generar informes</p>
                  </div>
                </div>
                <div className="mt-4">
                  <button className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 transition-colors">
                    Ver Reportes
                  </button>
                </div>
              </div>
            </div>

            {/* Card de Configuración - Sin cambios */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="h-10 w-10 bg-purple-100 rounded-full flex items-center justify-center">
                      <span className="text-purple-600 text-xl">⚙️</span>
                    </div>
                  </div>
                  <div className="ml-4">
                    <h3 className="text-lg font-medium text-gray-900">Configuración</h3>
                    <p className="text-sm text-gray-600">Ajustar preferencias</p>
                  </div>
                </div>
                <div className="mt-4">
                  <button className="w-full bg-purple-600 text-white py-2 px-4 rounded-lg hover:bg-purple-700 transition-colors">
                    Configurar
                  </button>
                </div>
              </div>
            </div>
          </div>

          {/* System Status */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Estado del Sistema</h3>
            </div>
            <div className="px-6 py-4">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">✅</div>
                  <div className="text-sm text-gray-600 mt-1">Backend</div>
                  <div className="text-xs text-green-600">Operativo</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">✅</div>
                  <div className="text-sm text-gray-600 mt-1">Base de Datos</div>
                  <div className="text-xs text-green-600">Conectada</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">✅</div>
                  <div className="text-sm text-gray-600 mt-1">Redis</div>
                  <div className="text-xs text-green-600">Activo</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">✅</div>
                  <div className="text-sm text-gray-600 mt-1">Autenticación</div>
                  <div className="text-xs text-green-600">Funcionando</div>
                </div>
              </div>
            </div>
          </div>

        </div>
      </main>
    </div>
  );
};

export default DashboardPage;
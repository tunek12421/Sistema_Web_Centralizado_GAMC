// frontend-auth/src/services/messageService.ts
// 🔧 VERSIÓN CORREGIDA CON DEBUG DETALLADO

// ✅ INTERFACES CORREGIDAS PARA COINCIDIR CON LA API
export interface Message {
  id: number;
  subject: string;
  content: string;
  sender: {
    id: string;
    firstName: string;
    lastName: string;
    email: string;
  };
  senderUnit: {
    id: number;
    name: string;
    code: string;
  };
  receiverUnit: {
    id: number;
    name: string;
    code: string;
  };
  messageType: {
    id: number;
    name: string;
  };
  status: {
    id: number;
    name: string;
    code: string; // 🔧 AGREGADO: code del status
  };
  priorityLevel: number;
  isUrgent: boolean;
  attachments: any[];
  readAt: string | null;
  respondedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface MessageListResponse {
  messages: Message[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
  hasNext: boolean;
  hasPrevious: boolean;
}

export interface CreateMessageRequest {
  subject: string;
  content: string;
  receiverUnitId: number;
  messageTypeId: number;
  priorityLevel: number;
  isUrgent: boolean;
}

// 🔧 CORREGIDO: Filtros que coinciden exactamente con la API
export interface MessageFilters {
  receiverUnitId?: number;  // ✅ CAMBIO: de unitId a receiverUnitId
  senderUnitId?: number;    // ✅ AGREGADO: filtro por unidad emisora
  messageType?: number;
  status?: number;
  isUrgent?: boolean;
  dateFrom?: string;
  dateTo?: string;
  searchText?: string;
  page?: number;
  limit?: number;
  sortBy?: 'created_at' | 'subject' | 'priority_level';
  sortOrder?: 'asc' | 'desc';
}

export interface MessageStats {
  totalMessages: number;
  messagesByStatus: Record<string, number>;
  messagesByType: Record<string, number>;
  urgentMessages: number;
  averageResponseTime: string;
}

// 🔧 MAPEO de nombres de estados a IDs (basado en la respuesta de la API)
const STATUS_NAME_TO_ID: { [key: string]: number } = {
  'DRAFT': 1,
  'SENT': 2,
  'READ': 3,
  'IN_PROGRESS': 4,
  'RESPONDED': 5,
  'RESOLVED': 6,
  'ARCHIVED': 7,
  'CANCELLED': 8
};

class MessageService {
  private baseURL = 'http://localhost:3000/api/v1';
  private debugMode = true; // 🔧 FLAG para debug detallado

  // 🔧 HELPER para logging detallado
  private log(message: string, data?: any) {
    if (this.debugMode) {
      console.log(`🔧 MessageService: ${message}`, data || '');
    }
  }

  private logError(message: string, error?: any) {
    console.error(`❌ MessageService: ${message}`, error || '');
  }

  // 🔧 CORREGIDO: Helper para obtener headers con autorización
  private getAuthHeaders(): Record<string, string> {
    this.log('🔐 Obteniendo headers de autenticación...');
    
    // 🔧 BÚSQUEDA EXHAUSTIVA de tokens
    const tokenSources = {
      accessToken: localStorage.getItem('accessToken'),
      token: localStorage.getItem('token'),
      authToken: localStorage.getItem('authToken')
    };
    
    this.log('🔍 Tokens disponibles:', tokenSources);
    
    // Usar el primer token disponible
    const token = tokenSources.accessToken || tokenSources.token || tokenSources.authToken;
    
    if (!token) {
      this.logError('❌ No hay token de autenticación disponible');
      this.log('📦 localStorage completo:', Object.keys(localStorage));
      throw new Error('No hay token de autenticación disponible. Por favor inicie sesión.');
    }
    
    this.log(`✅ Token encontrado: ${token.substring(0, 20)}...`);
    
    const headers = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    };
    
    this.log('📋 Headers preparados:', {
      'Content-Type': headers['Content-Type'],
      'Authorization': `Bearer ${token.substring(0, 20)}...`
    });
    
    return headers;
  }

  // 🔧 CORREGIDO: Helper para manejar respuestas con debug detallado
  private async handleResponse<T>(response: Response): Promise<T> {
    this.log(`📡 Respuesta recibida: ${response.status} ${response.statusText}`);
    this.log(`📋 Headers de respuesta:`, Object.fromEntries(response.headers.entries()));
    
    if (!response.ok) {
      const errorText = await response.text();
      this.logError(`❌ Error HTTP ${response.status}:`, errorText);
      
      let errorData: any = {};
      try {
        errorData = JSON.parse(errorText);
      } catch (e) {
        errorData = { message: errorText };
      }
      
      // 🔧 MANEJO ESPECÍFICO para diferentes códigos de error
      if (response.status === 401) {
        this.logError('🔒 Token expirado o inválido - limpiando localStorage');
        localStorage.removeItem('accessToken');
        localStorage.removeItem('token');
        localStorage.removeItem('authToken');
        throw new Error('Sesión expirada. Por favor inicie sesión nuevamente.');
      }
      
      if (response.status === 403) {
        throw new Error('Acceso denegado. No tiene permisos para esta acción.');
      }
      
      if (response.status === 404) {
        throw new Error('Recurso no encontrado.');
      }
      
      throw new Error(errorData.message || `Error ${response.status}: ${response.statusText}`);
    }
    
    const responseText = await response.text();
    this.log(`📦 Texto de respuesta completo:`, responseText);
    
    let data: any;
    try {
      data = JSON.parse(responseText);
      this.log(`✅ JSON parseado exitosamente:`, data);
    } catch (e) {
      this.logError('❌ Error parseando JSON:', e);
      throw new Error('Respuesta inválida del servidor');
    }
    
    if (!data.success) {
      this.logError('❌ Respuesta indica falla:', data);
      throw new Error(data.message || 'Error en la respuesta del servidor');
    }
    
    this.log(`📤 Retornando data:`, data.data);
    return data.data;
  }

  // 🔧 MÉTODO DE DEBUG para probar conectividad
  async testConnection(): Promise<boolean> {
    try {
      this.log('🧪 Probando conexión con el servidor...');
      
      const response = await fetch(`${this.baseURL}/../health`, {
        method: 'GET'
      });
      
      this.log(`🌐 Health check response: ${response.status}`);
      
      if (response.ok) {
        const healthData = await response.json();
        this.log('✅ Servidor saludable:', healthData);
        return true;
      } else {
        this.logError('❌ Servidor no responde correctamente:', response.status);
        return false;
      }
    } catch (error) {
      this.logError('❌ Error de conectividad:', error);
      return false;
    }
  }

  // 🔧 CORREGIDO: Obtener tipos de mensajes disponibles
  async getMessageTypes(): Promise<string[]> {
    this.log('🔄 Cargando tipos de mensajes...');
    
    try {
      const response = await fetch(`${this.baseURL}/messages/types`, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });
      
      const result = await this.handleResponse<string[]>(response);
      this.log('✅ Tipos de mensajes cargados:', result);
      return result;
    } catch (error) {
      this.logError('❌ Error cargando tipos de mensajes:', error);
      throw error;
    }
  }

  // 🔧 CORREGIDO: Obtener estados de mensajes disponibles
  async getMessageStatuses(): Promise<string[]> {
    this.log('🔄 Cargando estados de mensajes...');
    
    try {
      const response = await fetch(`${this.baseURL}/messages/statuses`, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });
      
      const result = await this.handleResponse<string[]>(response);
      this.log('✅ Estados de mensajes cargados:', result);
      return result;
    } catch (error) {
      this.logError('❌ Error cargando estados de mensajes:', error);
      throw error;
    }
  }

  // 🔧 COMPLETAMENTE REESCRITO: Listar mensajes con filtros y paginación
  async getMessages(filters: MessageFilters = {}): Promise<MessageListResponse> {
    this.log('🔄 Iniciando carga de mensajes...');
    this.log('📋 Filtros recibidos:', filters);
    
    const queryParams = new URLSearchParams();
    
    // 🔧 MAPEO EXHAUSTIVO de filtros
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        queryParams.append(key, value.toString());
        this.log(`🔗 Agregando filtro: ${key} = ${value}`);
      }
    });

    const url = `${this.baseURL}/messages${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    this.log('📡 URL completa:', url);
    
    try {
      this.log('🚀 Enviando request...');
      
      const response = await fetch(url, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });
      
      // 🔧 RECIBIR RESPUESTA RAW de la API
      const apiResponse = await this.handleResponse<any>(response);
      this.log('📦 Respuesta de API recibida:', apiResponse);
      
      // 🔧 MAPEO CORRECTO de la respuesta de API al formato esperado por el frontend
      const mappedResponse: MessageListResponse = {
        messages: apiResponse.messages || [],
        total: apiResponse.total || 0,
        page: apiResponse.page || 1,
        limit: apiResponse.limit || 10,
        totalPages: apiResponse.totalPages || 1,
        hasNext: (apiResponse.page || 1) < (apiResponse.totalPages || 1),
        hasPrevious: (apiResponse.page || 1) > 1
      };
      
      this.log('✅ Respuesta mapeada para frontend:', mappedResponse);
      this.log(`📊 Resumen: ${mappedResponse.messages.length} mensajes de ${mappedResponse.total} total, página ${mappedResponse.page}/${mappedResponse.totalPages}`);
      
      return mappedResponse;
      
    } catch (error) {
      this.logError('❌ Error completo en getMessages:', error);
      throw error;
    }
  }

  // 🔧 NUEVO: Método específico para obtener mensajes por unidad
  async getMessagesByUnit(unitId: number, asReceiver: boolean = true): Promise<MessageListResponse> {
    this.log(`🏢 Obteniendo mensajes para unidad ${unitId} como ${asReceiver ? 'receptor' : 'emisor'}`);
    
    const filters: MessageFilters = asReceiver 
      ? { receiverUnitId: unitId, page: 1, limit: 10 }
      : { senderUnitId: unitId, page: 1, limit: 10 };
      
    return this.getMessages(filters);
  }

  // Obtener un mensaje específico por ID
  async getMessageById(id: number): Promise<Message> {
    this.log(`🔍 Obteniendo mensaje ID: ${id}`);
    
    try {
      const response = await fetch(`${this.baseURL}/messages/${id}`, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });
      
      const result = await this.handleResponse<Message>(response);
      this.log('✅ Mensaje obtenido:', result);
      return result;
    } catch (error) {
      this.logError(`❌ Error obteniendo mensaje ${id}:`, error);
      throw error;
    }
  }

  // Crear un nuevo mensaje
  async createMessage(messageData: CreateMessageRequest): Promise<Message> {
    this.log('📝 Creando nuevo mensaje:', messageData);
    
    try {
      const response = await fetch(`${this.baseURL}/messages`, {
        method: 'POST',
        headers: this.getAuthHeaders(),
        body: JSON.stringify(messageData)
      });
      
      const result = await this.handleResponse<Message>(response);
      this.log('✅ Mensaje creado exitosamente:', result);
      return result;
    } catch (error) {
      this.logError('❌ Error creando mensaje:', error);
      throw error;
    }
  }

  // Marcar mensaje como leído
  async markAsRead(id: number): Promise<void> {
    this.log(`👁️ Marcando mensaje ${id} como leído`);
    
    try {
      const response = await fetch(`${this.baseURL}/messages/${id}/read`, {
        method: 'PUT',
        headers: this.getAuthHeaders()
      });
      
      await this.handleResponse<void>(response);
      this.log(`✅ Mensaje ${id} marcado como leído`);
    } catch (error) {
      this.logError(`❌ Error marcando mensaje ${id} como leído:`, error);
      throw error;
    }
  }

  // 🔧 CORREGIDO: Actualizar estado de mensaje
  async updateStatus(id: number, statusName: string): Promise<void> {
    this.log(`🔄 Actualizando estado de mensaje ${id} a: ${statusName}`);
    
    // Convertir nombre del estado a ID
    const statusId = STATUS_NAME_TO_ID[statusName];
    if (!statusId) {
      this.logError(`❌ Estado inválido: ${statusName}`);
      throw new Error(`Estado inválido: ${statusName}`);
    }
    
    this.log(`📋 Enviando statusId: ${statusId} para ${statusName}`);
    
    try {
      const response = await fetch(`${this.baseURL}/messages/${id}/status`, {
        method: 'PUT',
        headers: this.getAuthHeaders(),
        body: JSON.stringify({ statusId })
      });
      
      await this.handleResponse<void>(response);
      this.log(`✅ Estado de mensaje ${id} actualizado a ${statusName}`);
    } catch (error) {
      this.logError(`❌ Error actualizando estado de mensaje ${id}:`, error);
      throw error;
    }
  }

  // Eliminar mensaje (soft delete)
  async deleteMessage(id: number): Promise<void> {
    this.log(`🗑️ Eliminando mensaje ${id}`);
    
    try {
      const response = await fetch(`${this.baseURL}/messages/${id}`, {
        method: 'DELETE',
        headers: this.getAuthHeaders()
      });
      
      await this.handleResponse<void>(response);
      this.log(`✅ Mensaje ${id} eliminado`);
    } catch (error) {
      this.logError(`❌ Error eliminando mensaje ${id}:`, error);
      throw error;
    }
  }

  // Obtener estadísticas de mensajes
  async getMessageStats(unitId?: number): Promise<MessageStats> {
    this.log('📊 Obteniendo estadísticas de mensajes...', { unitId });
    
    const url = unitId 
      ? `${this.baseURL}/messages/stats?unitId=${unitId}`
      : `${this.baseURL}/messages/stats`;
    
    this.log('📡 URL de estadísticas:', url);
    
    try {
      const response = await fetch(url, {
        method: 'GET',
        headers: this.getAuthHeaders()
      });
      
      const result = await this.handleResponse<MessageStats>(response);
      this.log('✅ Estadísticas obtenidas:', result);
      return result;
    } catch (error) {
      this.logError('❌ Error obteniendo estadísticas:', error);
      throw error;
    }
  }

  // Búsqueda de mensajes con texto
  async searchMessages(searchText: string, filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    this.log(`🔍 Buscando mensajes con texto: "${searchText}"`);
    
    return this.getMessages({
      ...filters,
      searchText,
      page: 1,
      limit: 20
    });
  }

  // Obtener mensajes urgentes
  async getUrgentMessages(filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    this.log('🚨 Obteniendo mensajes urgentes');
    
    return this.getMessages({
      ...filters,
      isUrgent: true,
      sortBy: 'created_at',
      sortOrder: 'desc'
    });
  }

  // Obtener mensajes por estado específico
  async getMessagesByStatus(status: number, filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    this.log(`📋 Obteniendo mensajes con estado: ${status}`);
    
    return this.getMessages({
      ...filters,
      status,
      sortBy: 'created_at',
      sortOrder: 'desc'
    });
  }

  // 🔧 MÉTODO DE DEBUG COMPLETO
  async debugFullFlow(): Promise<void> {
    this.log('🧪 === INICIANDO DEBUG COMPLETO ===');
    
    try {
      // 1. Test de conectividad
      this.log('1️⃣ Probando conectividad...');
      const isConnected = await this.testConnection();
      this.log(`✅ Conectividad: ${isConnected ? 'OK' : 'FALLO'}`);
      
      // 2. Test de autenticación
      this.log('2️⃣ Probando autenticación...');
      const headers = this.getAuthHeaders();
      this.log('✅ Headers de auth obtenidos');
      
      // 3. Test de obtener mensajes sin filtros
      this.log('3️⃣ Probando obtener mensajes sin filtros...');
      const allMessages = await this.getMessages();
      this.log(`✅ Mensajes sin filtros: ${allMessages.messages.length} encontrados`);
      
      // 4. Test con filtros
      this.log('4️⃣ Probando con filtros...');
      const filteredMessages = await this.getMessages({ receiverUnitId: 1 });
      this.log(`✅ Mensajes con filtro receiverUnitId=1: ${filteredMessages.messages.length} encontrados`);
      
      this.log('🎉 === DEBUG COMPLETO EXITOSO ===');
      
    } catch (error) {
      this.logError('💥 === DEBUG COMPLETO FALLÓ ===', error);
      throw error;
    }
  }
}

// Exportar instancia singleton
export const messageService = new MessageService();

// Export default para compatibilidad
export default messageService;
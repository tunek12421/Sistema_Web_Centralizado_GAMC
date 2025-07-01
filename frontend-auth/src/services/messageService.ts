// frontend-auth/src/services/messageService.ts

// ✅ TODAS LAS INTERFACES EXPORTADAS CORRECTAMENTE
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

export interface MessageFilters {
  unitId?: number;
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

class MessageService {
  private baseURL = 'http://localhost:3000/api/v1';

  // Helper para obtener headers con autorización
  private getAuthHeaders(): Record<string, string> {
    const token = localStorage.getItem('accessToken');
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    };
  }

  // Helper para manejar respuestas de la API
  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `Error ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    if (!data.success) {
      throw new Error(data.message || 'Error en la respuesta del servidor');
    }
    
    return data.data;
  }

  // Obtener tipos de mensajes disponibles
  async getMessageTypes(): Promise<string[]> {
    const response = await fetch(`${this.baseURL}/messages/types`, {
      method: 'GET',
      headers: this.getAuthHeaders()
    });
    
    return this.handleResponse<string[]>(response);
  }

  // Obtener estados de mensajes disponibles
  async getMessageStatuses(): Promise<string[]> {
    const response = await fetch(`${this.baseURL}/messages/statuses`, {
      method: 'GET',
      headers: this.getAuthHeaders()
    });
    
    return this.handleResponse<string[]>(response);
  }

  // Listar mensajes con filtros y paginación
  async getMessages(filters: MessageFilters = {}): Promise<MessageListResponse> {
    const queryParams = new URLSearchParams();
    
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        queryParams.append(key, value.toString());
      }
    });

    const url = `${this.baseURL}/messages${queryParams.toString() ? `?${queryParams.toString()}` : ''}`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: this.getAuthHeaders()
    });
    
    return this.handleResponse<MessageListResponse>(response);
  }

  // Obtener un mensaje específico por ID
  async getMessageById(id: number): Promise<Message> {
    const response = await fetch(`${this.baseURL}/messages/${id}`, {
      method: 'GET',
      headers: this.getAuthHeaders()
    });
    
    return this.handleResponse<Message>(response);
  }

  // Crear un nuevo mensaje
  async createMessage(messageData: CreateMessageRequest): Promise<Message> {
    const response = await fetch(`${this.baseURL}/messages`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(messageData)
    });
    
    return this.handleResponse<Message>(response);
  }

  // Marcar mensaje como leído
  async markAsRead(id: number): Promise<void> {
    const response = await fetch(`${this.baseURL}/messages/${id}/read`, {
      method: 'PUT',
      headers: this.getAuthHeaders()
    });
    
    await this.handleResponse<void>(response);
  }

  // Actualizar estado de mensaje
  async updateStatus(id: number, status: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/messages/${id}/status`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify({ status })
    });
    
    await this.handleResponse<void>(response);
  }

  // Eliminar mensaje (soft delete)
  async deleteMessage(id: number): Promise<void> {
    const response = await fetch(`${this.baseURL}/messages/${id}`, {
      method: 'DELETE',
      headers: this.getAuthHeaders()
    });
    
    await this.handleResponse<void>(response);
  }

  // Obtener estadísticas de mensajes
  async getMessageStats(unitId?: number): Promise<MessageStats> {
    const url = unitId 
      ? `${this.baseURL}/messages/stats?unitId=${unitId}`
      : `${this.baseURL}/messages/stats`;
    
    const response = await fetch(url, {
      method: 'GET',
      headers: this.getAuthHeaders()
    });
    
    return this.handleResponse<MessageStats>(response);
  }

  // Búsqueda de mensajes con texto
  async searchMessages(searchText: string, filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    return this.getMessages({
      ...filters,
      searchText,
      page: 1,
      limit: 20
    });
  }

  // Obtener mensajes urgentes
  async getUrgentMessages(filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    return this.getMessages({
      ...filters,
      isUrgent: true,
      sortBy: 'created_at',
      sortOrder: 'desc'
    });
  }

  // Obtener mensajes por estado específico
  async getMessagesByStatus(status: number, filters: Partial<MessageFilters> = {}): Promise<MessageListResponse> {
    return this.getMessages({
      ...filters,
      status,
      sortBy: 'created_at',
      sortOrder: 'desc'
    });
  }
}

// Exportar instancia singleton
export const messageService = new MessageService();

// Export default para compatibilidad
export default messageService;
import { apiClient } from './api';

// Definiciones locales para authService
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
  lastLogin?: string;
  createdAt: string;
}

interface LoginCredentials {
  email: string;
  password: string;
}

interface RegisterData {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  organizationalUnitId: number;
  role?: 'admin' | 'input' | 'output';
}

interface AuthResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
  timestamp: string;
}

class AuthService {
  async login(credentials: LoginCredentials): Promise<AuthResponse> {
    const response = await apiClient.post<ApiResponse<AuthResponse>>('/auth/login', credentials);
    return response.data!;
  }

  async register(userData: RegisterData): Promise<User> {
    const response = await apiClient.post<ApiResponse<User>>('/auth/register', userData);
    return response.data!;
  }

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout');
  }

  async refreshToken(): Promise<{ accessToken: string }> {
    const response = await apiClient.post<ApiResponse<{ accessToken: string }>>('/auth/refresh');
    return response.data!;
  }

  async getCurrentUser(): Promise<User> {
    const response = await apiClient.get<ApiResponse<User>>('/auth/profile');
    return response.data!;
  }

  async verifyToken(): Promise<{ valid: boolean; user: User }> {
    const response = await apiClient.get<ApiResponse<{ valid: boolean; user: User }>>('/auth/verify');
    return response.data!;
  }

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await apiClient.put('/auth/change-password', {
      currentPassword,
      newPassword
    });
  }
}

export const authService = new AuthService();
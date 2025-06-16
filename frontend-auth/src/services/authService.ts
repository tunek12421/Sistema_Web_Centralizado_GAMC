import { apiClient } from './api';
import { LoginCredentials, RegisterData, AuthResponse, User, ApiResponse } from '../types/auth';

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
// index.ts independiente - sin importaciones problemáticas

// Definiciones locales de tipos principales
export interface User {
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

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface RegisterData {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  organizationalUnitId: number;
  role?: 'admin' | 'input' | 'output';
}

export interface AuthResponse {
  user: User;
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
  timestamp: string;
}

export interface FieldValidation {
  isValid: boolean;
  message: string;
  type: 'success' | 'error' | 'warning' | '';
}

// Re-exportar desde passwordReset.ts (que funciona)
export * from './passwordReset';
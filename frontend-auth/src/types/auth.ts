// src/types/auth.ts
// Versión corregida con tipos strict TypeScript

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

export interface ApiError {
  success: false;
  message: string;
  error: string;
  timestamp: string;
  status?: number;
}

export interface RefreshTokenResponse {
  accessToken: string;
  expiresIn: number;
}

export interface TokenVerificationResponse {
  valid: boolean;
  user: User;
  expiresAt: string;
}

export interface UserProfile extends User {
  preferences?: {
    language: string;
    timezone: string;
    notifications: boolean;
  };
  permissions?: string[];
  lastPasswordChange?: string;
  securityQuestionsCount?: number;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

export interface FormValidationState<T> {
  fields: { [K in keyof T]: FieldValidation };
  isFormValid: boolean;
  hasErrors: boolean;
}

// Convertir enum a const object
export const UserRole = {
  ADMIN: 'admin',
  INPUT: 'input',
  OUTPUT: 'output'
} as const;

export type UserRole = typeof UserRole[keyof typeof UserRole];

export const AuthErrorType = {
  INVALID_CREDENTIALS: 'invalid_credentials',
  USER_NOT_FOUND: 'user_not_found',
  USER_INACTIVE: 'user_inactive',
  EMAIL_ALREADY_EXISTS: 'email_already_exists',
  WEAK_PASSWORD: 'weak_password',
  TOKEN_EXPIRED: 'token_expired',
  TOKEN_INVALID: 'token_invalid',
  RATE_LIMIT_EXCEEDED: 'rate_limit_exceeded',
  NETWORK_ERROR: 'network_error',
  SERVER_ERROR: 'server_error'
} as const;

export type AuthErrorType = typeof AuthErrorType[keyof typeof AuthErrorType];

export const AUTH_CONFIG = {
  LOGIN_TIMEOUT: 10000,
  REFRESH_TIMEOUT: 5000,
  ACCESS_TOKEN_KEY: 'accessToken',
  REFRESH_TOKEN_KEY: 'refreshToken',
  USER_DATA_KEY: 'user',
  ACCESS_TOKEN_EXPIRY_MINUTES: 15,
  REFRESH_TOKEN_EXPIRY_DAYS: 7,
  MAX_LOGIN_ATTEMPTS: 5,
  LOGIN_COOLDOWN_MINUTES: 15
} as const;

export const isUser = (obj: any): obj is User => {
  return obj &&
    typeof obj.id === 'string' &&
    typeof obj.email === 'string' &&
    typeof obj.firstName === 'string' &&
    typeof obj.lastName === 'string' &&
    ['admin', 'input', 'output'].includes(obj.role);
};

export const isAuthResponse = (obj: any): obj is AuthResponse => {
  return obj &&
    typeof obj.accessToken === 'string' &&
    typeof obj.refreshToken === 'string' &&
    isUser(obj.user);
};

export const isApiResponse = <T>(obj: any): obj is ApiResponse<T> => {
  return obj &&
    typeof obj.success === 'boolean' &&
    typeof obj.message === 'string' &&
    typeof obj.timestamp === 'string';
};
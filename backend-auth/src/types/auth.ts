// ========================================
// GAMC Backend Auth - Tipos de Autenticación
// ========================================

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  organizationalUnitId: number;
  role?: 'admin' | 'input' | 'output';
}

export interface AuthResponse {
  success: boolean;
  message: string;
  data?: {
    user: UserProfile;
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  };
}

export interface UserProfile {
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

export interface JWTPayload {
  userId: string;
  email: string;
  role: string;
  organizationalUnitId: number;
  sessionId: string;
  iat?: number;
  exp?: number;
}

export interface RefreshTokenPayload {
  userId: string;
  sessionId: string;
  tokenVersion: number;
  iat?: number;
  exp?: number;
}

export interface SessionData {
  userId: string;
  email: string;
  role: string;
  organizationalUnitId: number;
  sessionId: string;
  createdAt: Date;
  lastActivity: Date;
  ipAddress?: string;
  userAgent?: string;
}

export interface PasswordResetRequest {
  email: string;
}

export interface PasswordResetConfirm {
  token: string;
  newPassword: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

export interface LogoutRequest {
  refreshToken?: string;
  logoutAll?: boolean;
}

// Request extendido con usuario autenticado
export interface AuthenticatedRequest {
  user: UserProfile;
  sessionId: string;
}

// Respuesta estándar de API
export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data?: T;
  error?: string;
  timestamp: string;
}
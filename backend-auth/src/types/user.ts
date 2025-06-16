export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: 'admin' | 'input' | 'output';
  organizationalUnitId: number;
  isActive: boolean;
  lastLogin?: Date;
  createdAt: Date;
  updatedAt: Date;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  organizationalUnitId: number;
  role?: 'admin' | 'input' | 'output';
}

export interface UpdateUserRequest {
  firstName?: string;
  lastName?: string;
  organizationalUnitId?: number;
  role?: 'admin' | 'input' | 'output';
  isActive?: boolean;
}
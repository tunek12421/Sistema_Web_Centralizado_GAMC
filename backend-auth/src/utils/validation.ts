// ========================================
// GAMC Backend Auth - Esquemas de Validación
// ========================================

import { z } from 'zod';

// Validación para login
export const loginSchema = z.object({
  email: z
    .string()
    .email('Formato de email inválido')
    .min(1, 'Email es requerido')
    .max(100, 'Email demasiado largo'),
  password: z
    .string()
    .min(1, 'Contraseña es requerida')
    .max(255, 'Contraseña demasiado larga')
});

// Validación para registro
export const registerSchema = z.object({
  email: z
    .string()
    .email('Formato de email inválido')
    .min(1, 'Email es requerido')
    .max(100, 'Email demasiado largo'),
  password: z
    .string()
    .min(8, 'La contraseña debe tener al menos 8 caracteres')
    .max(255, 'Contraseña demasiado larga')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
      'La contraseña debe contener al menos: 1 minúscula, 1 mayúscula, 1 número y 1 carácter especial'
    ),
  firstName: z
    .string()
    .min(1, 'Nombre es requerido')
    .max(50, 'Nombre demasiado largo')
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, 'Nombre solo debe contener letras y espacios'),
  lastName: z
    .string()
    .min(1, 'Apellido es requerido')
    .max(50, 'Apellido demasiado largo')
    .regex(/^[a-zA-ZÀ-ÿ\s]+$/, 'Apellido solo debe contener letras y espacios'),
  organizationalUnitId: z
    .number()
    .int('ID de unidad debe ser un entero')
    .positive('ID de unidad debe ser positivo'),
  role: z
    .enum(['admin', 'input', 'output'])
    .optional()
    .default('output')
});

// Validación para cambio de contraseña
export const changePasswordSchema = z.object({
  currentPassword: z
    .string()
    .min(1, 'Contraseña actual es requerida'),
  newPassword: z
    .string()
    .min(8, 'La nueva contraseña debe tener al menos 8 caracteres')
    .max(255, 'Nueva contraseña demasiado larga')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
      'La nueva contraseña debe contener al menos: 1 minúscula, 1 mayúscula, 1 número y 1 carácter especial'
    )
}).refine(
  (data) => data.currentPassword !== data.newPassword,
  {
    message: 'La nueva contraseña debe ser diferente a la actual',
    path: ['newPassword']
  }
);

// Validación para reset de contraseña
export const passwordResetRequestSchema = z.object({
  email: z
    .string()
    .email('Formato de email inválido')
    .min(1, 'Email es requerido')
});

export const passwordResetConfirmSchema = z.object({
  token: z
    .string()
    .min(1, 'Token es requerido'),
  newPassword: z
    .string()
    .min(8, 'La contraseña debe tener al menos 8 caracteres')
    .max(255, 'Contraseña demasiado larga')
    .regex(
      /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]/,
      'La contraseña debe contener al menos: 1 minúscula, 1 mayúscula, 1 número y 1 carácter especial'
    )
});

// Validación para refresh token
export const refreshTokenSchema = z.object({
  refreshToken: z
    .string()
    .min(1, 'Refresh token es requerido')
});

// Tipos inferidos de los esquemas
export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;
export type PasswordResetRequestInput = z.infer<typeof passwordResetRequestSchema>;
export type PasswordResetConfirmInput = z.infer<typeof passwordResetConfirmSchema>;
export type RefreshTokenInput = z.infer<typeof refreshTokenSchema>;

// Función helper para validar datos
export const validateInput = <T>(schema: z.ZodSchema<T>, data: unknown): {
  success: boolean;
  data?: T;
  errors?: string[];
} => {
  try {
    const result = schema.parse(data);
    return {
      success: true,
      data: result
    };
  } catch (error) {
    if (error instanceof z.ZodError) {
      return {
        success: false,
        errors: error.errors.map(err => err.message)
      };
    }
    return {
      success: false,
      errors: ['Error de validación desconocido']
    };
  }
};

// Sanitización de strings
export const sanitizeString = (str: string): string => {
  return str.trim().replace(/\s+/g, ' ');
};

// Validación de email personalizada
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

// Validación de contraseña fuerte
export const isStrongPassword = (password: string): {
  isValid: boolean;
  reasons: string[];
} => {
  const reasons: string[] = [];
  
  if (password.length < 8) {
    reasons.push('Debe tener al menos 8 caracteres');
  }
  
  if (!/[a-z]/.test(password)) {
    reasons.push('Debe contener al menos una letra minúscula');
  }
  
  if (!/[A-Z]/.test(password)) {
    reasons.push('Debe contener al menos una letra mayúscula');
  }
  
  if (!/\d/.test(password)) {
    reasons.push('Debe contener al menos un número');
  }
  
  if (!/[@$!%*?&]/.test(password)) {
    reasons.push('Debe contener al menos un carácter especial (@$!%*?&)');
  }
  
  return {
    isValid: reasons.length === 0,
    reasons
  };
};
// ========================================
// GAMC Backend Auth - Servicio de Autenticación
// ========================================

import bcrypt from 'bcryptjs';
import { v4 as uuidv4 } from 'uuid';
import prisma from '../config/database';
import { redisClient } from '../config/redis';
import { generateTokens, verifyRefreshToken } from '../utils/jwt';
import { 
  LoginRequest, 
  RegisterRequest, 
  AuthResponse, 
  UserProfile, 
  SessionData,
  ChangePasswordRequest 
} from '../types/auth';

export class AuthService {
  
  // Función auxiliar para generar username único basado en el contexto GAMC
  private static async generateUniqueUsername(firstName: string, lastName: string, email: string, organizationalUnitId: number): Promise<string> {
    try {
      // Obtener código de la unidad organizacional para usar como prefijo
      const orgUnit = await prisma.organizational_units.findUnique({
        where: { id: organizationalUnitId },
        select: { code: true }
      });

      const unitCode = orgUnit?.code?.toLowerCase() || 'gamc';
      
      // Limpiar nombres (sin espacios, caracteres especiales, acentos)
      const cleanFirstName = firstName.toLowerCase()
        .normalize("NFD").replace(/[\u0300-\u036f]/g, "")
        .replace(/[^a-z0-9]/g, '');
      
      const cleanLastName = lastName.toLowerCase()
        .normalize("NFD").replace(/[\u0300-\u036f]/g, "")
        .replace(/[^a-z0-9]/g, '');

      // Estrategia 1: unidad.nombre.apellido (ej: obras.juan.perez)
      let username = `${unitCode}.${cleanFirstName}.${cleanLastName}`;
      
      // Verificar si existe
      let existingUser = await prisma.users.findUnique({ where: { username } });
      
      // Estrategia 2: usar parte del email si el formato es institucional
      if (existingUser && email.includes('@gamc.gov.bo')) {
        username = email.split('@')[0].toLowerCase();
        existingUser = await prisma.users.findUnique({ where: { username } });
      }
      
      // Estrategia 3: agregar números incrementales
      if (existingUser) {
        const baseUsername = `${unitCode}.${cleanFirstName}.${cleanLastName}`;
        let counter = 1;
        
        do {
          username = `${baseUsername}${counter}`;
          existingUser = await prisma.users.findUnique({ where: { username } });
          counter++;
        } while (existingUser && counter <= 99);
        
        // Si después de 99 intentos aún hay conflicto, usar timestamp
        if (existingUser) {
          username = `${baseUsername}_${Date.now().toString().slice(-6)}`;
        }
      }
      
      return username;
      
    } catch (error) {
      console.error('Error generando username:', error);
      // Fallback: usar timestamp para garantizar unicidad
      return `user_${Date.now()}`;
    }
  }

  // Login de usuario
  static async login(credentials: LoginRequest, ipAddress?: string, userAgent?: string): Promise<AuthResponse> {
    try {
      const { email, password } = credentials;

      // Buscar usuario con unidad organizacional
      const user = await prisma.users.findUnique({
        where: { email, is_active: true },
        include: {
          organizational_units: {
            select: {
              id: true,
              name: true,
              code: true
            }
          }
        }
      });

      if (!user) {
        return {
          success: false,
          message: 'Credenciales inválidas'
        };
      }

      // Verificar contraseña
      const isPasswordValid = await bcrypt.compare(password, user.password_hash);
      if (!isPasswordValid) {
        return {
          success: false,
          message: 'Credenciales inválidas'
        };
      }

      // Generar ID de sesión
      const sessionId = uuidv4();

      // Crear sesión en Redis
      const sessionData: SessionData = {
        userId: user.id,
        email: user.email,
        role: user.role,
        organizationalUnitId: user.organizational_unit_id!,
        sessionId,
        createdAt: new Date(),
        lastActivity: new Date(),
        ipAddress,
        userAgent
      };

      await redisClient.setEx(
        `session:${sessionId}`,
        7 * 24 * 60 * 60, // 7 días en segundos
        JSON.stringify(sessionData)
      );

      // Generar tokens JWT
      const { accessToken, refreshToken } = generateTokens({
        userId: user.id,
        email: user.email,
        role: user.role,
        organizationalUnitId: user.organizational_unit_id!,
        sessionId
      });

      // Guardar refresh token en Redis
      await redisClient.setEx(
        `refresh:${user.id}:${sessionId}`,
        7 * 24 * 60 * 60, // 7 días
        refreshToken
      );

      // Actualizar último login
      await prisma.users.update({
        where: { id: user.id },
        data: { last_login: new Date() }
      });

      // Crear perfil de usuario para respuesta
      const userProfile: UserProfile = {
        id: user.id,
        email: user.email,
        firstName: user.first_name,
        lastName: user.last_name,
        role: user.role as 'admin' | 'input' | 'output',
        organizationalUnit: {
          id: user.organizational_units!.id,
          name: user.organizational_units!.name,
          code: user.organizational_units!.code
        },
        isActive: user.is_active,
        lastLogin: new Date(),
        createdAt: user.created_at
      };

      return {
        success: true,
        message: 'Login exitoso',
        data: {
          user: userProfile,
          accessToken,
          refreshToken,
          expiresIn: 15 * 60 // 15 minutos en segundos
        }
      };

    } catch (error) {
      console.error('Error en login:', error);
      return {
        success: false,
        message: 'Error interno del servidor'
      };
    }
  }

  // Registro de usuario
  static async register(userData: RegisterRequest): Promise<AuthResponse> {
    try {
      const { email, password, firstName, lastName, organizationalUnitId, role = 'output' } = userData;

      // Verificar si el usuario ya existe
      const existingUser = await prisma.users.findUnique({
        where: { email }
      });

      if (existingUser) {
        return {
          success: false,
          message: 'El usuario ya existe'
        };
      }

      // Verificar que existe la unidad organizacional
      const orgUnit = await prisma.organizational_units.findUnique({
        where: { id: organizationalUnitId, is_active: true }
      });

      if (!orgUnit) {
        return {
          success: false,
          message: 'Unidad organizacional no válida'
        };
      }

      // Generar username único basado en el contexto GAMC
      const username = await this.generateUniqueUsername(firstName, lastName, email, organizationalUnitId);

      // Hashear contraseña
      const passwordHash = await bcrypt.hash(password, 12);

      // Crear usuario - CAMPO USERNAME INCLUIDO
      const newUser = await prisma.users.create({
        data: {
          id: uuidv4(),
          username, // ← CAMPO AGREGADO - SOLUCIÓN AL ERROR
          email,
          password_hash: passwordHash,
          first_name: firstName,
          last_name: lastName,
          role,
          organizational_unit_id: organizationalUnitId,
          is_active: true,
          password_changed_at: new Date()
        },
        include: {
          organizational_units: {
            select: {
              id: true,
              name: true,
              code: true
            }
          }
        }
      });

      // Log en consola en lugar de base de datos
      console.log(`✅ Usuario registrado exitosamente:`);
      console.log(`   Username: ${username}`);
      console.log(`   Email: ${email}`);
      console.log(`   Unidad: ${orgUnit.name}`);
      console.log(`   Role: ${role}`);

      // Crear perfil de usuario para respuesta
      const userProfile: UserProfile = {
        id: newUser.id,
        email: newUser.email,
        firstName: newUser.first_name,
        lastName: newUser.last_name,
        role: newUser.role as 'admin' | 'input' | 'output',
        organizationalUnit: {
          id: newUser.organizational_units!.id,
          name: newUser.organizational_units!.name,
          code: newUser.organizational_units!.code
        },
        isActive: newUser.is_active,
        createdAt: newUser.created_at
      };

      return {
        success: true,
        message: `Usuario registrado exitosamente con username: ${username}`,
        data: {
          user: userProfile,
          accessToken: '',
          refreshToken: '',
          expiresIn: 0
        }
      };

    } catch (error) {
      console.error('Error en registro:', error);
      return {
        success: false,
        message: 'Error interno del servidor'
      };
    }
  }

  // Refresh token
  static async refreshToken(refreshToken: string): Promise<AuthResponse> {
    try {
      // Verificar refresh token
      const payload = verifyRefreshToken(refreshToken);
      if (!payload) {
        return {
          success: false,
          message: 'Token de refresh inválido'
        };
      }

      // Verificar que el refresh token existe en Redis
      const storedToken = await redisClient.get(`refresh:${payload.userId}:${payload.sessionId}`);
      if (storedToken !== refreshToken) {
        return {
          success: false,
          message: 'Token de refresh inválido'
        };
      }

      // Obtener datos de sesión
      const sessionData = await redisClient.get(`session:${payload.sessionId}`);
      if (!sessionData) {
        return {
          success: false,
          message: 'Sesión expirada'
        };
      }

      const session: SessionData = JSON.parse(sessionData);

      // Generar nuevos tokens
      const { accessToken, refreshToken: newRefreshToken } = generateTokens({
        userId: session.userId,
        email: session.email,
        role: session.role,
        organizationalUnitId: session.organizationalUnitId,
        sessionId: session.sessionId
      });

      // Actualizar refresh token en Redis
      await redisClient.setEx(
        `refresh:${session.userId}:${session.sessionId}`,
        7 * 24 * 60 * 60,
        newRefreshToken
      );

      return {
        success: true,
        message: 'Token renovado exitosamente',
        data: {
          user: {} as UserProfile, // Se puede omitir en refresh
          accessToken,
          refreshToken: newRefreshToken,
          expiresIn: 15 * 60
        }
      };

    } catch (error) {
      console.error('Error en refresh token:', error);
      return {
        success: false,
        message: 'Error interno del servidor'
      };
    }
  }

  // Logout
  static async logout(userId: string, sessionId?: string, logoutAll = false): Promise<{ success: boolean; message: string }> {
    try {
      if (logoutAll) {
        // Logout de todas las sesiones del usuario
        const keys = await redisClient.keys(`session:*`);
        const sessions = await Promise.all(
          keys.map(async (key) => {
            const data = await redisClient.get(key);
            return data ? { key, data: JSON.parse(data) } : null;
          })
        );

        const userSessions = sessions.filter(s => s && s.data.userId === userId);
        
        for (const session of userSessions) {
          if (session) {
            await redisClient.del(session.key);
            await redisClient.del(`refresh:${userId}:${session.data.sessionId}`);
          }
        }
      } else if (sessionId) {
        // Logout de sesión específica
        await redisClient.del(`session:${sessionId}`);
        await redisClient.del(`refresh:${userId}:${sessionId}`);
      }

      return {
        success: true,
        message: logoutAll ? 'Logout de todas las sesiones exitoso' : 'Logout exitoso'
      };

    } catch (error) {
      console.error('Error en logout:', error);
      return {
        success: false,
        message: 'Error interno del servidor'
      };
    }
  }

  // Cambiar contraseña
  static async changePassword(userId: string, passwordData: ChangePasswordRequest): Promise<{ success: boolean; message: string }> {
    try {
      const { currentPassword, newPassword } = passwordData;

      // Obtener usuario actual
      const user = await prisma.users.findUnique({
        where: { id: userId, is_active: true }
      });

      if (!user) {
        return {
          success: false,
          message: 'Usuario no encontrado'
        };
      }

      // Verificar contraseña actual
      const isCurrentPasswordValid = await bcrypt.compare(currentPassword, user.password_hash);
      if (!isCurrentPasswordValid) {
        return {
          success: false,
          message: 'Contraseña actual incorrecta'
        };
      }

      // Hashear nueva contraseña
      const newPasswordHash = await bcrypt.hash(newPassword, 12);

      // Actualizar contraseña
      await prisma.users.update({
        where: { id: userId },
        data: {
          password_hash: newPasswordHash,
          password_changed_at: new Date()
        }
      });

      return {
        success: true,
        message: 'Contraseña cambiada exitosamente'
      };

    } catch (error) {
      console.error('Error cambiando contraseña:', error);
      return {
        success: false,
        message: 'Error interno del servidor'
      };
    }
  }

  // Obtener perfil de usuario
  static async getUserProfile(userId: string): Promise<UserProfile | null> {
    try {
      const user = await prisma.users.findUnique({
        where: { id: userId, is_active: true },
        include: {
          organizational_units: {
            select: {
              id: true,
              name: true,
              code: true
            }
          }
        }
      });

      if (!user) return null;

      return {
        id: user.id,
        email: user.email,
        firstName: user.first_name,
        lastName: user.last_name,
        role: user.role as 'admin' | 'input' | 'output',
        organizationalUnit: {
          id: user.organizational_units!.id,
          name: user.organizational_units!.name,
          code: user.organizational_units!.code
        },
        isActive: user.is_active,
        lastLogin: user.last_login,
        createdAt: user.created_at
      };

    } catch (error) {
      console.error('Error obteniendo perfil:', error);
      return null;
    }
  }
}
// ========================================
// GAMC Backend Auth - Controlador de Autenticación
// ========================================

import { Request, Response } from 'express';
import { AuthService } from '../services/authService';
import { loginSchema, registerSchema, changePasswordSchema } from '../utils/validation';
import { ApiResponse } from '../types/auth';

export class AuthController {

  // POST /api/v1/auth/login
  static async login(req: Request, res: Response): Promise<void> {
    try {
      // Validar datos de entrada
      const validation = loginSchema.safeParse(req.body);
      if (!validation.success) {
        res.status(400).json({
          success: false,
          message: 'Datos de entrada inválidos',
          error: validation.error.errors[0].message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      const { email, password } = validation.data;
      const ipAddress = req.ip || req.connection.remoteAddress;
      const userAgent = req.get('User-Agent');

      // Ejecutar login
      const result = await AuthService.login(
        { email, password },
        ipAddress,
        userAgent
      );

      if (result.success && result.data) {
        // Configurar cookie HttpOnly para refresh token
        res.cookie('refreshToken', result.data.refreshToken, {
          httpOnly: true,
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'lax',
          maxAge: 7 * 24 * 60 * 60 * 1000 // 7 días
        });

        res.status(200).json({
          success: true,
          message: result.message,
          data: {
            user: result.data.user,
            accessToken: result.data.accessToken,
            expiresIn: result.data.expiresIn
          },
          timestamp: new Date().toISOString()
        } as ApiResponse);
      } else {
        res.status(401).json({
          success: false,
          message: result.message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      }

    } catch (error) {
      console.error('Error en login controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // POST /api/v1/auth/register
  static async register(req: Request, res: Response): Promise<void> {
    try {
      // Validar datos de entrada
      const validation = registerSchema.safeParse(req.body);
      if (!validation.success) {
        res.status(400).json({
          success: false,
          message: 'Datos de entrada inválidos',
          error: validation.error.errors[0].message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Ejecutar registro
      const result = await AuthService.register(validation.data);

      if (result.success) {
        res.status(201).json({
          success: true,
          message: result.message,
          data: result.data?.user,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      } else {
        res.status(400).json({
          success: false,
          message: result.message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      }

    } catch (error) {
      console.error('Error en register controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // POST /api/v1/auth/refresh
  static async refreshToken(req: Request, res: Response): Promise<void> {
    try {
      const refreshToken = req.cookies.refreshToken || req.body.refreshToken;

      if (!refreshToken) {
        res.status(401).json({
          success: false,
          message: 'Refresh token requerido',
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Ejecutar refresh
      const result = await AuthService.refreshToken(refreshToken);

      if (result.success && result.data) {
        // Actualizar cookie
        res.cookie('refreshToken', result.data.refreshToken, {
          httpOnly: true,
          secure: process.env.NODE_ENV === 'production',
          sameSite: 'lax',
          maxAge: 7 * 24 * 60 * 60 * 1000
        });

        res.status(200).json({
          success: true,
          message: result.message,
          data: {
            accessToken: result.data.accessToken,
            expiresIn: result.data.expiresIn
          },
          timestamp: new Date().toISOString()
        } as ApiResponse);
      } else {
        res.status(401).json({
          success: false,
          message: result.message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      }

    } catch (error) {
      console.error('Error en refresh controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // POST /api/v1/auth/logout
  static async logout(req: Request, res: Response): Promise<void> {
    try {
      const user = (req as any).user;
      const sessionId = (req as any).sessionId;
      const logoutAll = req.body.logoutAll === true;

      if (!user) {
        res.status(401).json({
          success: false,
          message: 'Usuario no autenticado',
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Ejecutar logout
      const result = await AuthService.logout(user.id, sessionId, logoutAll);

      // Limpiar cookie
      res.clearCookie('refreshToken');

      res.status(200).json({
        success: true,
        message: result.message,
        timestamp: new Date().toISOString()
      } as ApiResponse);

    } catch (error) {
      console.error('Error en logout controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // GET /api/v1/auth/profile
  static async getProfile(req: Request, res: Response): Promise<void> {
    try {
      const user = (req as any).user;

      if (!user) {
        res.status(401).json({
          success: false,
          message: 'Usuario no autenticado',
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Obtener perfil actualizado
      const profile = await AuthService.getUserProfile(user.id);

      if (profile) {
        res.status(200).json({
          success: true,
          message: 'Perfil obtenido exitosamente',
          data: profile,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      } else {
        res.status(404).json({
          success: false,
          message: 'Perfil no encontrado',
          timestamp: new Date().toISOString()
        } as ApiResponse);
      }

    } catch (error) {
      console.error('Error en getProfile controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // PUT /api/v1/auth/change-password
  static async changePassword(req: Request, res: Response): Promise<void> {
    try {
      const user = (req as any).user;

      if (!user) {
        res.status(401).json({
          success: false,
          message: 'Usuario no autenticado',
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Validar datos
      const validation = changePasswordSchema.safeParse(req.body);
      if (!validation.success) {
        res.status(400).json({
          success: false,
          message: 'Datos de entrada inválidos',
          error: validation.error.errors[0].message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      // Ejecutar cambio de contraseña
      const result = await AuthService.changePassword(user.id, validation.data);

      if (result.success) {
        res.status(200).json({
          success: true,
          message: result.message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      } else {
        res.status(400).json({
          success: false,
          message: result.message,
          timestamp: new Date().toISOString()
        } as ApiResponse);
      }

    } catch (error) {
      console.error('Error en changePassword controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }

  // GET /api/v1/auth/verify
  static async verifyToken(req: Request, res: Response): Promise<void> {
    try {
      const user = (req as any).user;

      if (!user) {
        res.status(401).json({
          success: false,
          message: 'Token inválido',
          timestamp: new Date().toISOString()
        } as ApiResponse);
        return;
      }

      res.status(200).json({
        success: true,
        message: 'Token válido',
        data: {
          valid: true,
          user: user
        },
        timestamp: new Date().toISOString()
      } as ApiResponse);

    } catch (error) {
      console.error('Error en verifyToken controller:', error);
      res.status(500).json({
        success: false,
        message: 'Error interno del servidor',
        timestamp: new Date().toISOString()
      } as ApiResponse);
    }
  }
}
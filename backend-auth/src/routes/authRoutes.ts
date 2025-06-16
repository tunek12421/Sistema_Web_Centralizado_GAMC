// ========================================
// GAMC Backend Auth - Rutas de Autenticación
// ========================================

import { Router } from 'express';
import { AuthController } from '../controllers/authController';
import { authenticateToken, logUserActivity, userRateLimit } from '../middleware/auth';

const router = Router();

// Rate limiting específico para auth
const authRateLimit = userRateLimit(10, 15 * 60 * 1000); // 10 requests por 15 minutos

/**
 * @route   POST /api/v1/auth/login
 * @desc    Iniciar sesión de usuario
 * @access  Public
 * @body    { email: string, password: string }
 */
router.post('/login', 
  authRateLimit,
  logUserActivity('LOGIN_ATTEMPT'),
  AuthController.login
);

/**
 * @route   POST /api/v1/auth/register
 * @desc    Registrar nuevo usuario
 * @access  Public
 * @body    { email, password, firstName, lastName, organizationalUnitId, role? }
 */
router.post('/register',
  authRateLimit,
  logUserActivity('REGISTER_ATTEMPT'),
  AuthController.register
);

/**
 * @route   POST /api/v1/auth/refresh
 * @desc    Renovar access token usando refresh token
 * @access  Public
 * @body    { refreshToken?: string } (o cookie)
 */
router.post('/refresh',
  authRateLimit,
  AuthController.refreshToken
);

/**
 * @route   POST /api/v1/auth/logout
 * @desc    Cerrar sesión del usuario
 * @access  Private
 * @body    { logoutAll?: boolean }
 */
router.post('/logout',
  authenticateToken,
  logUserActivity('LOGOUT'),
  AuthController.logout
);

/**
 * @route   GET /api/v1/auth/profile
 * @desc    Obtener perfil del usuario autenticado
 * @access  Private
 */
router.get('/profile',
  authenticateToken,
  logUserActivity('GET_PROFILE'),
  AuthController.getProfile
);

/**
 * @route   PUT /api/v1/auth/change-password
 * @desc    Cambiar contraseña del usuario
 * @access  Private
 * @body    { currentPassword: string, newPassword: string }
 */
router.put('/change-password',
  authenticateToken,
  authRateLimit,
  logUserActivity('CHANGE_PASSWORD'),
  AuthController.changePassword
);

/**
 * @route   GET /api/v1/auth/verify
 * @desc    Verificar validez del token de acceso
 * @access  Private
 */
router.get('/verify',
  authenticateToken,
  AuthController.verifyToken
);

export default router;
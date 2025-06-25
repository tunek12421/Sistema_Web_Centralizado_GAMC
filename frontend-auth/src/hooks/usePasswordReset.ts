// src/hooks/usePasswordReset.ts
// Custom hook para manejar todo el flujo de reset de contraseña
// Centraliza la lógica de estado, validaciones y llamadas a API

import { useState, useEffect, useCallback, useRef } from 'react';
import { passwordResetService, PasswordResetError } from '../services/passwordResetService';
import { 
  validateEmail,
  validatePassword,
  validatePasswordConfirmation,
  canMakePasswordResetRequest,
  getTimeUntilNextRequest,
  recordPasswordResetRequest,
  formatTimeRemaining
} from '../utils/passwordValidation';
import {
  extractTokenFromURL,
  validateTokenFormat,
  getTokenTimeRemaining,
  formatTokenTimeRemaining,
  cleanTokenFromURL,
  storeToken,
  retrieveStoredToken,
  clearStoredToken
} from '../utils/tokenValidation';
import {
  PasswordResetState,
  ForgotPasswordFormState,
  ResetPasswordFormState,
  UsePasswordResetOptions,
  UsePasswordResetResult,
  PasswordResetErrorType,
  PasswordResetToken,
  PASSWORD_RESET_CONFIG
} from '../types/passwordReset';

/**
 * Custom hook para manejar el flujo completo de reset de contraseña
 * Incluye gestión de estado, validaciones, timers y llamadas a API
 */
export const usePasswordReset = (options: UsePasswordResetOptions = {}): UsePasswordResetResult => {
  const {
    autoRedirect = true,
    redirectDelay = PASSWORD_RESET_CONFIG.AUTO_REDIRECT_DELAY,
    enableRateLimitUI = true,
    onSuccess,
    onError
  } = options;

  // Referencias para cleanup
  const rateLimitTimerRef = useRef<NodeJS.Timeout | null>(null);
  const tokenTimerRef = useRef<NodeJS.Timeout | null>(null);
  const redirectTimerRef = useRef<NodeJS.Timeout | null>(null);

  // ========================================
  // ESTADO PRINCIPAL
  // ========================================

  const [state, setState] = useState<PasswordResetState>({
    step: 'request',
    error: undefined,
    errorMessage: undefined,
    successMessage: undefined
  });

  const [forgotFormState, setForgotFormState] = useState<ForgotPasswordFormState>({
    email: '',
    isLoading: false,
    isSubmitted: false,
    validation: { isValid: false, message: '', type: '' }
  });

  const [resetFormState, setResetFormState] = useState<ResetPasswordFormState>({
    token: '',
    newPassword: '',
    confirmPassword: '',
    isLoading: false,
    isSubmitted: false,
    validation: {
      token: { isValid: false, message: '', type: '' },
      password: { isValid: false, message: '', type: '' },
      confirmPassword: { isValid: false, message: '', type: '' }
    }
  });

  // Estados de timers
  const [timeUntilNextRequest, setTimeUntilNextRequest] = useState(0);
  const [tokenTimeRemaining, setTokenTimeRemaining] = useState(0);

  // ========================================
  // EFECTOS PARA INICIALIZACIÓN
  // ========================================

  // Inicialización: detectar token en URL y restaurar estado
  useEffect(() => {
    const initializeFromURL = () => {
      const urlToken = extractTokenFromURL();
      
      if (urlToken) {
        // Hay token en URL, ir a paso de confirmación
        const tokenValidation = validateTokenFormat(urlToken);
        
        setResetFormState(prev => ({
          ...prev,
          token: urlToken,
          validation: {
            ...prev.validation,
            token: tokenValidation
          }
        }));

        setState(prev => ({
          ...prev,
          step: 'confirm',
          token: urlToken
        }));

        // Limpiar token de URL por seguridad
        cleanTokenFromURL();
        
        // Almacenar token temporalmente
        storeToken(urlToken);
      } else {
        // No hay token en URL, verificar localStorage
        const storedTokenData = retrieveStoredToken();
        
        if (storedTokenData.isValid && storedTokenData.token) {
          setResetFormState(prev => ({
            ...prev,
            token: storedTokenData.token!,
            validation: {
              ...prev.validation,
              token: { isValid: true, message: '✓ Token válido', type: 'success' }
            }
          }));

          setState(prev => ({
            ...prev,
            step: 'confirm',
            token: storedTokenData.token,
            expiresAt: storedTokenData.expiresAt ? new Date(storedTokenData.expiresAt).getTime() : undefined
          }));
        }
      }
    };

    initializeFromURL();
  }, []);

  // Timer para rate limiting
  useEffect(() => {
    if (!enableRateLimitUI) return;

    const updateRateLimit = () => {
      const remaining = getTimeUntilNextRequest();
      setTimeUntilNextRequest(remaining);
    };

    updateRateLimit();
    rateLimitTimerRef.current = setInterval(updateRateLimit, 1000);

    return () => {
      if (rateLimitTimerRef.current) {
        clearInterval(rateLimitTimerRef.current);
      }
    };
  }, [enableRateLimitUI]);

  // Timer para expiración de token
  useEffect(() => {
    if (state.step !== 'confirm' || !state.expiresAt) return;

    const updateTokenTimer = () => {
      const remaining = Math.max(0, Math.ceil((state.expiresAt! - Date.now()) / 1000));
      setTokenTimeRemaining(remaining);

      if (remaining === 0) {
        setState(prev => ({ ...prev, step: 'error', error: PasswordResetErrorType.TOKEN_EXPIRED }));
      }
    };

    updateTokenTimer();
    tokenTimerRef.current = setInterval(updateTokenTimer, 1000);

    return () => {
      if (tokenTimerRef.current) {
        clearInterval(tokenTimerRef.current);
      }
    };
  }, [state.step, state.expiresAt]);

  // ========================================
  // VALIDACIONES EN TIEMPO REAL
  // ========================================

  // Validación de email en forgot form
  useEffect(() => {
    if (forgotFormState.email.trim()) {
      const validation = validateEmail(forgotFormState.email);
      setForgotFormState(prev => ({ ...prev, validation }));
    } else {
      setForgotFormState(prev => ({ 
        ...prev, 
        validation: { isValid: false, message: '', type: '' } 
      }));
    }
  }, [forgotFormState.email]);

  // Validación de campos en reset form
  useEffect(() => {
    const tokenValidation = resetFormState.token 
      ? validateTokenFormat(resetFormState.token)
      : { isValid: false, message: '', type: '' };

    const passwordValidation = resetFormState.newPassword
      ? validatePassword(resetFormState.newPassword)
      : { isValid: false, message: '', type: '' };

    const confirmValidation = resetFormState.confirmPassword
      ? validatePasswordConfirmation(resetFormState.newPassword, resetFormState.confirmPassword)
      : { isValid: false, message: '', type: '' };

    setResetFormState(prev => ({
      ...prev,
      validation: {
        token: tokenValidation,
        password: passwordValidation,
        confirmPassword: confirmValidation
      }
    }));
  }, [resetFormState.token, resetFormState.newPassword, resetFormState.confirmPassword]);

  // ========================================
  // ACCIONES PRINCIPALES
  // ========================================

  const requestReset = useCallback(async (email: string): Promise<void> => {
    // Validaciones previas
    if (!canMakePasswordResetRequest()) {
      const error = new PasswordResetError(
        PasswordResetErrorType.RATE_LIMIT_EXCEEDED,
        `Debe esperar ${formatTimeRemaining(timeUntilNextRequest)} antes de hacer otra solicitud`
      );
      
      if (onError) onError(error);
      throw error;
    }

    const emailValidation = validateEmail(email);
    if (!emailValidation.isValid) {
      const error = new PasswordResetError(
        PasswordResetErrorType.EMAIL_NOT_INSTITUTIONAL,
        emailValidation.message
      );
      
      if (onError) onError(error);
      throw error;
    }

    // Actualizar estado de loading
    setForgotFormState(prev => ({ 
      ...prev, 
      email, 
      isLoading: true 
    }));

    setState(prev => ({ 
      ...prev, 
      email, 
      error: undefined, 
      errorMessage: undefined 
    }));

    try {
      const response = await passwordResetService.requestPasswordReset({ email });

      // Registrar solicitud para rate limiting
      recordPasswordResetRequest();
      passwordResetService.recordResetRequest();

      // Actualizar estado de éxito
      setForgotFormState(prev => ({ 
        ...prev, 
        isLoading: false, 
        isSubmitted: true 
      }));

      setState(prev => ({ 
        ...prev, 
        step: 'email_sent', 
        requestedAt: Date.now(),
        successMessage: response.data.message
      }));

      if (onSuccess) onSuccess('request');

    } catch (error) {
      console.error('Error en solicitud de reset:', error);
      
      setForgotFormState(prev => ({ 
        ...prev, 
        isLoading: false 
      }));

      const resetError = error instanceof PasswordResetError 
        ? error 
        : new PasswordResetError(PasswordResetErrorType.UNKNOWN_ERROR, 'Error inesperado');

      setState(prev => ({ 
        ...prev, 
        step: 'error', 
        error: resetError.type, 
        errorMessage: resetError.getUserMessage() 
      }));

      if (onError) onError(resetError);
      throw resetError;
    }
  }, [timeUntilNextRequest, onSuccess, onError]);

  const confirmReset = useCallback(async (token: string, newPassword: string): Promise<void> => {
    // Validaciones previas
    const tokenValidation = validateTokenFormat(token);
    if (!tokenValidation.isValid) {
      const error = new PasswordResetError(
        PasswordResetErrorType.TOKEN_INVALID,
        tokenValidation.message
      );
      
      if (onError) onError(error);
      throw error;
    }

    const passwordValidation = validatePassword(newPassword);
    if (!passwordValidation.isValid) {
      const error = new PasswordResetError(
        PasswordResetErrorType.PASSWORD_WEAK,
        passwordValidation.message
      );
      
      if (onError) onError(error);
      throw error;
    }

    // Verificar expiración de token
    if (tokenTimeRemaining <= 0) {
      const error = new PasswordResetError(
        PasswordResetErrorType.TOKEN_EXPIRED,
        'El token ha expirado'
      );
      
      if (onError) onError(error);
      throw error;
    }

    // Actualizar estado de loading
    setResetFormState(prev => ({ 
      ...prev, 
      token, 
      newPassword, 
      isLoading: true 
    }));

    setState(prev => ({ 
      ...prev, 
      error: undefined, 
      errorMessage: undefined 
    }));

    try {
      const response = await passwordResetService.confirmPasswordReset({
        token,
        newPassword
      });

      // Limpiar token almacenado
      clearStoredToken();

      // Actualizar estado de éxito
      setResetFormState(prev => ({ 
        ...prev, 
        isLoading: false, 
        isSubmitted: true,
        // Limpiar contraseñas por seguridad
        newPassword: '',
        confirmPassword: ''
      }));

      setState(prev => ({ 
        ...prev, 
        step: 'success',
        successMessage: response.data.message
      }));

      if (onSuccess) onSuccess('confirm');

      // Auto-redirect si está habilitado
      if (autoRedirect) {
        redirectTimerRef.current = setTimeout(() => {
          // El componente padre debe manejar la redirección
          if (onSuccess) onSuccess('confirm');
        }, redirectDelay);
      }

    } catch (error) {
      console.error('Error en confirmación de reset:', error);
      
      setResetFormState(prev => ({ 
        ...prev, 
        isLoading: false 
      }));

      const resetError = error instanceof PasswordResetError 
        ? error 
        : new PasswordResetError(PasswordResetErrorType.UNKNOWN_ERROR, 'Error inesperado');

      setState(prev => ({ 
        ...prev, 
        step: 'error', 
        error: resetError.type, 
        errorMessage: resetError.getUserMessage() 
      }));

      if (onError) onError(resetError);
      throw resetError;
    }
  }, [tokenTimeRemaining, onSuccess, onError, autoRedirect, redirectDelay]);

  // ========================================
  // FUNCIONES AUXILIARES
  // ========================================

  const validateToken = useCallback((token: string) => {
    return validateTokenFormat(token);
  }, []);

  const validateNewPassword = useCallback((password: string) => {
    return validatePassword(password);
  }, []);

  // Estados computados
  const canSubmitRequest = forgotFormState.validation.isValid && timeUntilNextRequest === 0;
  const canSubmitConfirm = resetFormState.validation.token.isValid && 
                          resetFormState.validation.password.isValid && 
                          resetFormState.validation.confirmPassword.isValid &&
                          tokenTimeRemaining > 0;

  // ========================================
  // FUNCIONES ADMIN (SI ESTÁN DISPONIBLES)
  // ========================================

  const getResetStatus = useCallback(async (): Promise<PasswordResetToken[]> => {
    try {
      const response = await passwordResetService.getPasswordResetStatus();
      return response.data.tokens;
    } catch (error) {
      console.error('Error al obtener estado de reset:', error);
      throw error;
    }
  }, []);

  const cleanupExpiredTokens = useCallback(async (): Promise<number> => {
    try {
      const response = await passwordResetService.cleanupExpiredTokens();
      return response.data.cleanedTokens;
    } catch (error) {
      console.error('Error en limpieza de tokens:', error);
      throw error;
    }
  }, []);

  // ========================================
  // CLEANUP
  // ========================================

  useEffect(() => {
    return () => {
      if (rateLimitTimerRef.current) clearInterval(rateLimitTimerRef.current);
      if (tokenTimerRef.current) clearInterval(tokenTimerRef.current);
      if (redirectTimerRef.current) clearTimeout(redirectTimerRef.current);
    };
  }, []);

  // ========================================
  // RETURN DEL HOOK
  // ========================================

  return {
    // Estado
    state,
    formStates: {
      forgot: forgotFormState,
      reset: resetFormState
    },

    // Acciones principales
    requestReset,
    confirmReset,
    validateToken,
    validatePassword: validateNewPassword,

    // Utilidades
    canSubmitRequest,
    canSubmitConfirm,
    timeUntilNextRequest,

    // Funciones admin (opcionales)
    getResetStatus,
    cleanupExpiredTokens
  };
};

export default usePasswordReset;
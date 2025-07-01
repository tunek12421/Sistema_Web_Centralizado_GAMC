// src/constants/passwordPolicies.ts
// Configuración centralizada de políticas de contraseñas y seguridad
// Sincronizado con el backend para mantener consistencia

// ========================================
// POLÍTICAS DE CONTRASEÑAS
// ========================================

export const PASSWORD_POLICIES = {
  // Longitud
  MIN_LENGTH: 8,
  MAX_LENGTH: 128,
  
  // Caracteres requeridos
  REQUIRES_UPPERCASE: true,
  REQUIRES_LOWERCASE: true,
  REQUIRES_NUMBERS: true,
  REQUIRES_SPECIAL_CHARS: true,
  
  // Caracteres especiales permitidos
  ALLOWED_SPECIAL_CHARS: '@$!%*?&',
  
  // Patrones de validación
  PATTERNS: {
    UPPERCASE: /[A-Z]/,
    LOWERCASE: /[a-z]/,
    NUMBERS: /\d/,
    SPECIAL_CHARS: /[@$!%*?&]/,
    FULL_VALIDATION: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/
  },
  
  // Patrones prohibidos
  FORBIDDEN_PATTERNS: [
    /(.)\1{2,}/, // 3 o más caracteres repetidos consecutivos
    /123456/,    // Secuencias numéricas comunes
    /abcdef/,    // Secuencias alfabéticas
    /qwerty/i,   // Patrones de teclado
    /password/i, // Palabra "password"
    /gamc/i,     // Nombre de la institución
    /admin/i     // Palabras administrativas
  ],
  
  // Mensajes de validación
  VALIDATION_MESSAGES: {
    TOO_SHORT: `La contraseña debe tener al menos 8 caracteres`,
    TOO_LONG: `La contraseña no puede exceder 128 caracteres`,
    NEEDS_UPPERCASE: 'Debe incluir al menos una letra mayúscula (A-Z)',
    NEEDS_LOWERCASE: 'Debe incluir al menos una letra minúscula (a-z)',
    NEEDS_NUMBERS: 'Debe incluir al menos un número (0-9)',
    NEEDS_SPECIAL: 'Debe incluir al menos un símbolo (@$!%*?&)',
    FORBIDDEN_PATTERN: 'Contiene un patrón no permitido',
    INVALID_CHARS: 'Contiene caracteres no permitidos'
  }
} as const;

// ========================================
// CONFIGURACIÓN DE FORTALEZA DE CONTRASEÑA
// ========================================

export const PASSWORD_STRENGTH = {
  // Niveles de fortaleza
  LEVELS: {
    VERY_WEAK: 0,
    WEAK: 1,
    FAIR: 2,
    GOOD: 3,
    STRONG: 4,
    VERY_STRONG: 5
  },
  
  // Etiquetas para cada nivel
  LABELS: {
    0: 'Muy débil',
    1: 'Débil',
    2: 'Regular',
    3: 'Buena',
    4: 'Fuerte',
    5: 'Muy fuerte'
  },
  
  // Colores para UI
  COLORS: {
    0: '#ef4444', // red-500
    1: '#f97316', // orange-500
    2: '#eab308', // yellow-500
    3: '#22c55e', // green-500
    4: '#06b6d4', // cyan-500
    5: '#8b5cf6'  // violet-500
  },
  
  // Clases CSS para barras de progreso
  CSS_CLASSES: {
    0: 'bg-red-500',
    1: 'bg-orange-500',
    2: 'bg-yellow-500',
    3: 'bg-green-500',
    4: 'bg-cyan-500',
    5: 'bg-violet-500'
  },
  
  // Puntuación por criterio
  SCORING: {
    BASE_LENGTH: 1,        // Punto por cumplir longitud mínima
    EXTRA_LENGTH: 0.5,     // 0.5 puntos por cada 4 chars extra
    UPPERCASE: 1,          // Punto por mayúsculas
    LOWERCASE: 1,          // Punto por minúsculas
    NUMBERS: 1,            // Punto por números
    SPECIAL_CHARS: 1,      // Punto por símbolos
    VARIETY_BONUS: 0.5,    // Bonus por variedad de caracteres
    LENGTH_BONUS: 1        // Bonus por contraseñas >12 chars
  }
} as const;

// ========================================
// CONFIGURACIÓN DE PREGUNTAS DE SEGURIDAD
// ========================================

export const SECURITY_QUESTIONS_POLICIES = {
  // Límites
  MIN_QUESTIONS_PER_USER: 1,
  MAX_QUESTIONS_PER_USER: 3,
  RECOMMENDED_QUESTIONS: 3,
  
  // Validación de respuestas
  ANSWER_MIN_LENGTH: 2,
  ANSWER_MAX_LENGTH: 100,
  
  // Intentos durante reset
  MAX_VERIFICATION_ATTEMPTS: 3,
  VERIFICATION_LOCKOUT_MINUTES: 30,
  
  // Categorías disponibles
  CATEGORIES: {
    PERSONAL: {
      id: 'personal',
      name: 'Personal',
      description: 'Información personal y familiar',
      icon: '👤',
      color: 'blue'
    },
    EDUCATION: {
      id: 'education',
      name: 'Educación',
      description: 'Estudios y formación académica',
      icon: '🎓',
      color: 'green'
    },
    PROFESSIONAL: {
      id: 'professional',
      name: 'Profesional',
      description: 'Trabajo y experiencia en GAMC',
      icon: '💼',
      color: 'purple'
    },
    PREFERENCES: {
      id: 'preferences',
      name: 'Preferencias',
      description: 'Gustos personales y favoritos',
      icon: '❤️',
      color: 'pink'
    }
  },
  
  // Mensajes de validación
  VALIDATION_MESSAGES: {
    ANSWER_TOO_SHORT: 'La respuesta debe tener al menos 2 caracteres',
    ANSWER_TOO_LONG: 'La respuesta no puede exceder 100 caracteres',
    ANSWER_REQUIRED: 'La respuesta es requerida',
    QUESTION_REQUIRED: 'Debe seleccionar una pregunta',
    DUPLICATE_QUESTION: 'Ya tiene configurada esta pregunta',
    MAX_QUESTIONS_REACHED: 'Ha alcanzado el límite máximo de 3 preguntas',
    MIN_QUESTIONS_REQUIRED: 'Debe configurar al menos 1 pregunta de seguridad'
  }
} as const;

// ========================================
// CONFIGURACIÓN DE RATE LIMITING
// ========================================

export const RATE_LIMITING = {
  // Password reset requests
  PASSWORD_RESET: {
    MAX_REQUESTS: 1,
    WINDOW_MINUTES: 5,
    MESSAGE: 'Debe esperar 5 minutos entre solicitudes de reset'
  },
  
  // Security question verification
  SECURITY_VERIFICATION: {
    MAX_ATTEMPTS: 3,
    WINDOW_MINUTES: 30,
    MESSAGE: 'Demasiados intentos fallidos. Intente en 30 minutos'
  },
  
  // Login attempts
  LOGIN_ATTEMPTS: {
    MAX_ATTEMPTS: 5,
    WINDOW_MINUTES: 15,
    MESSAGE: 'Demasiados intentos de login. Espere 15 minutos'
  },
  
  // Registration attempts
  REGISTRATION: {
    MAX_ATTEMPTS: 3,
    WINDOW_MINUTES: 60,
    MESSAGE: 'Límite de registros alcanzado. Intente en 1 hora'
  }
} as const;

// ========================================
// CONFIGURACIÓN DE TOKENS
// ========================================

export const TOKEN_POLICIES = {
  // Password reset tokens
  RESET_TOKEN: {
    LENGTH: 64,
    EXPIRY_MINUTES: 30,
    FORMAT: 'hex',
    PATTERN: /^[a-f0-9]{64}$/i
  },
  
  // JWT tokens
  JWT: {
    ACCESS_TOKEN_MINUTES: 15,
    REFRESH_TOKEN_DAYS: 7,
    ALGORITHM: 'HS256'
  },
  
  // Session tokens
  SESSION: {
    IDLE_TIMEOUT_MINUTES: 30,
    ABSOLUTE_TIMEOUT_HOURS: 8
  }
} as const;

// ========================================
// CONFIGURACIÓN DE EMAILS
// ========================================

export const EMAIL_POLICIES = {
  // Dominios permitidos para usuarios institucionales
  INSTITUTIONAL_DOMAINS: ['gamc.gov.bo'],
  
  // Patrones de validación
  PATTERNS: {
    INSTITUTIONAL: /^[^\s@]+@gamc\.gov\.bo$/,
    GENERAL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  },
  
  // Mensajes
  VALIDATION_MESSAGES: {
    REQUIRED: 'El email es requerido',
    INVALID_FORMAT: 'Formato de email inválido',
    NOT_INSTITUTIONAL: 'Solo emails @gamc.gov.bo están permitidos',
    ALREADY_EXISTS: 'Este email ya está registrado'
  }
} as const;

// ========================================
// CONFIGURACIÓN DE UI/UX
// ========================================

export const UI_CONFIG = {
  // Timeouts y delays
  TIMEOUTS: {
    SUCCESS_MESSAGE_MS: 5000,
    ERROR_MESSAGE_MS: 8000,
    AUTO_REDIRECT_MS: 3000,
    TOAST_DURATION_MS: 4000,
    LOADING_MIN_MS: 500
  },
  
  // Animaciones
  ANIMATIONS: {
    FADE_IN_MS: 300,
    SLIDE_IN_MS: 250,
    BOUNCE_DURATION_MS: 600,
    PULSE_INTERVAL_MS: 1000
  },
  
  // Breakpoints para responsive
  BREAKPOINTS: {
    SM: '640px',
    MD: '768px',
    LG: '1024px',
    XL: '1280px'
  },
  
  // Colores del tema
  THEME_COLORS: {
    PRIMARY: '#2563eb',      // blue-600
    SECONDARY: '#7c3aed',    // violet-600
    SUCCESS: '#16a34a',      // green-600
    WARNING: '#ea580c',      // orange-600
    ERROR: '#dc2626',        // red-600
    INFO: '#0891b2'          // cyan-600
  }
} as const;

// ========================================
// CONFIGURACIÓN DE LOGGING Y DEBUGGING
// ========================================

export const LOGGING_CONFIG = {
  // Niveles de log
  LEVELS: {
    ERROR: 0,
    WARN: 1,
    INFO: 2,
    DEBUG: 3
  },
  
  // Contextos de operaciones
  OPERATIONS: {
    LOGIN: 'login',
    REGISTER: 'register',
    FORGOT_PASSWORD: 'forgot-password',
    VERIFY_SECURITY_QUESTION: 'verify-security-question',
    RESET_PASSWORD: 'reset-password',
    SETUP_QUESTIONS: 'setup-security-questions',
    UPDATE_QUESTION: 'update-security-question',
    REMOVE_QUESTION: 'remove-security-question'
  },
  
  // Configuración para desarrollo
  DEV_CONFIG: {
    CONSOLE_ENABLED: true,
    NETWORK_LOGS: true,
    PERFORMANCE_LOGS: true,
    ERROR_STACK_TRACES: true
  }
} as const;

// ========================================
// CONFIGURACIÓN DE AMBIENTE
// ========================================

export const ENVIRONMENT_CONFIG = {
  // URLs base según ambiente
  API_BASE_URLS: {
    development: 'http://localhost:3000/api/v1',
    staging: 'https://staging-api.gamc.gov.bo/api/v1',
    production: 'https://api.gamc.gov.bo/api/v1'
  },
  
  // Configuración por ambiente
  ENVIRONMENTS: {
    development: {
      DEBUG: true,
      MOCK_DELAYS: false,
      CONSOLE_LOGS: true
    },
    staging: {
      DEBUG: true,
      MOCK_DELAYS: false,
      CONSOLE_LOGS: false
    },
    production: {
      DEBUG: false,
      MOCK_DELAYS: false,
      CONSOLE_LOGS: false
    }
  }
} as const;

// ========================================
// VALIDACIONES COMBINADAS
// ========================================

/**
 * Obtiene configuración activa según el ambiente
 */
export const getActiveConfig = () => {
  const env = process.env.NODE_ENV || 'development';
  return {
    ...ENVIRONMENT_CONFIG.ENVIRONMENTS[env as keyof typeof ENVIRONMENT_CONFIG.ENVIRONMENTS],
    API_BASE_URL: ENVIRONMENT_CONFIG.API_BASE_URLS[env as keyof typeof ENVIRONMENT_CONFIG.API_BASE_URLS]
  };
};

/**
 * Verifica si una contraseña cumple todas las políticas
 */
export const validatePasswordPolicy = (password: string): {
  valid: boolean;
  errors: string[];
  score: number;
} => {
  const errors: string[] = [];
  let score = 0;

  // Longitud
  if (password.length < PASSWORD_POLICIES.MIN_LENGTH) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.TOO_SHORT);
  } else {
    score += PASSWORD_STRENGTH.SCORING.BASE_LENGTH;
  }

  if (password.length > PASSWORD_POLICIES.MAX_LENGTH) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.TOO_LONG);
  }

  // Caracteres requeridos
  if (!PASSWORD_POLICIES.PATTERNS.UPPERCASE.test(password)) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.NEEDS_UPPERCASE);
  } else {
    score += PASSWORD_STRENGTH.SCORING.UPPERCASE;
  }

  if (!PASSWORD_POLICIES.PATTERNS.LOWERCASE.test(password)) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.NEEDS_LOWERCASE);
  } else {
    score += PASSWORD_STRENGTH.SCORING.LOWERCASE;
  }

  if (!PASSWORD_POLICIES.PATTERNS.NUMBERS.test(password)) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.NEEDS_NUMBERS);
  } else {
    score += PASSWORD_STRENGTH.SCORING.NUMBERS;
  }

  if (!PASSWORD_POLICIES.PATTERNS.SPECIAL_CHARS.test(password)) {
    errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.NEEDS_SPECIAL);
  } else {
    score += PASSWORD_STRENGTH.SCORING.SPECIAL_CHARS;
  }

  // Patrones prohibidos
  for (const pattern of PASSWORD_POLICIES.FORBIDDEN_PATTERNS) {
    if (pattern.test(password)) {
      errors.push(PASSWORD_POLICIES.VALIDATION_MESSAGES.FORBIDDEN_PATTERN);
      break;
    }
  }

  return {
    valid: errors.length === 0,
    errors,
    score: Math.min(score, PASSWORD_STRENGTH.LEVELS.VERY_STRONG)
  };
};
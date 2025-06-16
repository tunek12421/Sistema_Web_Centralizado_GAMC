export const jwtConfig = {
  secret: process.env.JWT_SECRET || 'gamc_jwt_secret_super_secure_2024_key_never_share',
  refreshSecret: process.env.JWT_REFRESH_SECRET || 'gamc_jwt_refresh_secret_super_secure_2024_key',
  expiresIn: process.env.JWT_EXPIRES_IN || '15m',
  refreshExpiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d',
  issuer: 'gamc-auth',
  audience: 'gamc-system'
};
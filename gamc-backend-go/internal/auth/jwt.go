// internal/auth/jwt.go
package auth

import (
	"fmt"
	"time"

	"gamc-backend-go/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims representa los claims del JWT
type JWTClaims struct {
	UserID               string `json:"userId"`
	Email                string `json:"email"`
	Role                 string `json:"role"`
	OrganizationalUnitID int    `json:"organizationalUnitId"`
	SessionID            string `json:"sessionId"`
	JTI                  string `json:"jti,omitempty"` // JWT ID para blacklist
	jwt.RegisteredClaims
}

// RefreshTokenClaims representa los claims del refresh token
type RefreshTokenClaims struct {
	UserID       string `json:"userId"`
	SessionID    string `json:"sessionId"`
	TokenVersion int    `json:"tokenVersion"`
	jwt.RegisteredClaims
}

// JWTService maneja la generación y validación de tokens JWT
type JWTService struct {
	config *config.Config
}

// NewJWTService crea una nueva instancia del servicio JWT
func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{config: cfg}
}

// GenerateAccessToken genera un access token
func (j *JWTService) GenerateAccessToken(userID, email, role string, orgUnitID int, sessionID string) (string, error) {
	now := time.Now()
	jti := uuid.New().String()

	claims := &JWTClaims{
		UserID:               userID,
		Email:                email,
		Role:                 role,
		OrganizationalUnitID: orgUnitID,
		SessionID:            sessionID,
		JTI:                  jti,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.JWTIssuer,
			Audience:  []string{j.config.JWTAudience},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWTExpiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.JWTSecret))
}

// GenerateRefreshToken genera un refresh token
func (j *JWTService) GenerateRefreshToken(userID, sessionID string) (string, error) {
	now := time.Now()

	claims := &RefreshTokenClaims{
		UserID:       userID,
		SessionID:    sessionID,
		TokenVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.JWTIssuer,
			Audience:  []string{"gamc-refresh"},
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.config.JWTRefreshExpiresIn)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.JWTRefreshSecret))
}

// GenerateTokens genera ambos tokens (access y refresh)
func (j *JWTService) GenerateTokens(userID, email, role string, orgUnitID int, sessionID string) (accessToken, refreshToken string, err error) {
	accessToken, err = j.GenerateAccessToken(userID, email, role, orgUnitID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = j.GenerateRefreshToken(userID, sessionID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// VerifyAccessToken verifica y parsea un access token
func (j *JWTService) VerifyAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Verificar issuer y audience
		if claims.Issuer != j.config.JWTIssuer {
			return nil, fmt.Errorf("invalid issuer")
		}

		validAudience := false
		for _, aud := range claims.Audience {
			if aud == j.config.JWTAudience {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, fmt.Errorf("invalid audience")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// VerifyRefreshToken verifica y parsea un refresh token
func (j *JWTService) VerifyRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWTRefreshSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		// Verificar issuer y audience
		if claims.Issuer != j.config.JWTIssuer {
			return nil, fmt.Errorf("invalid issuer")
		}

		validAudience := false
		for _, aud := range claims.Audience {
			if aud == "gamc-refresh" {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, fmt.Errorf("invalid audience")
		}

		return claims, nil
	}

	return nil, fmt.Errorf("invalid refresh token")
}

// ExtractTokenFromHeader extrae el token del header Authorization
func ExtractTokenFromHeader(authHeader string) string {
	const bearerPrefix = "Bearer "
	if len(authHeader) > len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
		return authHeader[len(bearerPrefix):]
	}
	return ""
}

// GetTokenExpiration obtiene la fecha de expiración de un token
func (j *JWTService) GetTokenExpiration(tokenString string) (*time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		if claims.ExpiresAt != nil {
			exp := claims.ExpiresAt.Time
			return &exp, nil
		}
	}

	return nil, fmt.Errorf("no expiration found in token")
}

// IsTokenExpiringSoon verifica si el token expira pronto
func (j *JWTService) IsTokenExpiringSoon(tokenString string, threshold time.Duration) (bool, error) {
	exp, err := j.GetTokenExpiration(tokenString)
	if err != nil {
		return true, err
	}

	return time.Now().Add(threshold).After(*exp), nil
}

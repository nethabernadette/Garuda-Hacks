package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"garuda-hacks/backend/users"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultAccessTokenTTL = time.Hour
	jwtIssuer             = "harvestlink"
)

type TokenManager struct {
	secret         []byte
	accessTokenTTL time.Duration
}

type Claims struct {
	UserID string         `json:"user_id"`
	Email  string         `json:"email"`
	Role   users.UserRole `json:"role"`
	jwt.RegisteredClaims
}

func NewTokenManagerFromEnv() (*TokenManager, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, ErrJWTSecretNotSet
	}

	ttl := defaultAccessTokenTTL
	if rawTTL := os.Getenv("JWT_ACCESS_TOKEN_TTL_SECONDS"); rawTTL != "" {
		seconds, err := strconv.Atoi(rawTTL)
		if err != nil || seconds <= 0 {
			return nil, errors.New("JWT_ACCESS_TOKEN_TTL_SECONDS must be a positive integer")
		}
		ttl = time.Duration(seconds) * time.Second
	}

	return NewTokenManager(secret, ttl), nil
}

func NewTokenManager(secret string, accessTokenTTL time.Duration) *TokenManager {
	return &TokenManager{
		secret:         []byte(secret),
		accessTokenTTL: accessTokenTTL,
	}
}

func (m *TokenManager) AccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}

func (m *TokenManager) GenerateAccessToken(user *users.User) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jwtIssuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if !claims.Role.IsValid() {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

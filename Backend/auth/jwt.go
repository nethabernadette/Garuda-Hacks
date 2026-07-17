package auth

import (
	"time"

	"garuda-hacks/backend/users"
	"github.com/golang-jwt/jwt/v5"
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
	config, err := LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	return NewTokenManager(config.JWTSecret, config.AccessTokenTTL), nil
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
	parser := jwt.NewParser(
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(jwtIssuer),
	)

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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
	if claims.UserID == "" || claims.Email == "" || claims.Subject == "" || claims.Subject != claims.UserID {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

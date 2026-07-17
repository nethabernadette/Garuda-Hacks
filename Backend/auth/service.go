package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"garuda-hacks/backend/users"
)

type Service struct {
	repository Repository
	tokens     *TokenManager
}

func NewService(repository Repository, tokenManager *TokenManager) *Service {
	return &Service{
		repository: repository,
		tokens:     tokenManager,
	}
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	normalizeRegisterRequest(&req)
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.repository.FindUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrCredentialsNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateEmail
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Role:         req.Role,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	}

	if err := s.repository.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{User: NewUserResponse(user)}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	normalizeLoginRequest(&req)
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	user, err := s.repository.FindUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrCredentialsNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, err
	}

	accessToken, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshTokenStr, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour) // 30 days expiry
	rt := &RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenStr,
		ExpiresAt: expiresAt,
	}

	if err := s.repository.SaveRefreshToken(ctx, rt); err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokens.AccessTokenTTL().Seconds()),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*LoginResponse, error) {
	if req.RefreshToken == "" {
		return nil, ErrInvalidToken
	}

	rt, err := s.repository.FindRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if rt.IsExpired() || rt.IsRevoked() {
		return nil, ErrInvalidToken
	}

	user, err := s.repository.FindUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	accessToken, err := s.tokens.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rt.Token,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.tokens.AccessTokenTTL().Seconds()),
	}, nil
}

func (s *Service) Logout(ctx context.Context, tokenStr string) error {
	if tokenStr == "" {
		return ErrInvalidToken
	}
	return s.repository.RevokeRefreshToken(ctx, tokenStr)
}

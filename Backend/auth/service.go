package auth

import (
	"context"
	"errors"

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

	return &LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.tokens.AccessTokenTTL().Seconds()),
	}, nil
}

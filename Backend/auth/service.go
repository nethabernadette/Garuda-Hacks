package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"garuda-hacks/backend/users"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users  users.Repository
	tokens *TokenManager
}

func NewService(userRepository users.Repository, tokenManager *TokenManager) *Service {
	return &Service{
		users:  userRepository,
		tokens: tokenManager,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	normalizeRegisterRequest(&req)
	if err := validateRegisterRequest(req); err != nil {
		return nil, err
	}

	existing, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, users.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateEmail
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &users.User{
		Role:         req.Role,
		CompanyName:  req.CompanyName,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Phone:        req.Phone,
		City:         req.City,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{User: NewUserResponse(user)}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	normalizeLoginRequest(&req)
	if err := validateLoginRequest(req); err != nil {
		return nil, err
	}

	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
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

func normalizeRegisterRequest(req *RegisterRequest) {
	req.Role = users.UserRole(strings.ToUpper(strings.TrimSpace(string(req.Role))))
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)
	req.City = strings.TrimSpace(req.City)
}

func normalizeLoginRequest(req *LoginRequest) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
}

func validateRegisterRequest(req RegisterRequest) error {
	if !req.Role.IsValid() {
		return ErrInvalidRole
	}
	if req.CompanyName == "" {
		return ErrRequiredCompanyName
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	if err := validatePassword(req.Password); err != nil {
		return err
	}
	if req.Phone == "" {
		return ErrRequiredPhone
	}
	if req.City == "" {
		return ErrRequiredCity
	}

	return nil
}

func validateLoginRequest(req LoginRequest) error {
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	return validatePassword(req.Password)
}

func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

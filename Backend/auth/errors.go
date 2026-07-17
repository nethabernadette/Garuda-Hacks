package auth

import "errors"

var (
	ErrDuplicateEmail     = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrMissingToken       = errors.New("authorization token is required")
	ErrForbidden          = errors.New("insufficient permissions")
	ErrJWTSecretNotSet    = errors.New("JWT_SECRET is required")
	ErrInvalidJWTConfig   = errors.New("JWT configuration is invalid")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrRequiredEmail      = errors.New("email is required")
	ErrInvalidEmail       = errors.New("email is invalid")
	ErrRequiredPassword   = errors.New("password is required")
	ErrInvalidPassword    = errors.New("password must be at least 8 characters")
	ErrRequiredRole       = errors.New("role is required")
	ErrInvalidRole        = errors.New("role is invalid")
)

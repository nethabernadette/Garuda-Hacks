package auth

import "errors"

var (
	ErrDuplicateEmail      = errors.New("email already registered")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidToken        = errors.New("invalid token")
	ErrMissingToken        = errors.New("authorization token is required")
	ErrForbidden           = errors.New("insufficient permissions")
	ErrJWTSecretNotSet     = errors.New("JWT_SECRET is required")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInvalidEmail        = errors.New("email is invalid")
	ErrInvalidPassword     = errors.New("password must be at least 8 characters")
	ErrInvalidRole         = errors.New("role is invalid")
	ErrRequiredCompanyName = errors.New("company_name is required")
	ErrRequiredPhone       = errors.New("phone is required")
	ErrRequiredCity        = errors.New("city is required")
)

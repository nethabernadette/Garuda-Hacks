package users

import "errors"

var (
	ErrUnauthorized        = errors.New("authentication is required")
	ErrForbidden           = errors.New("insufficient permissions")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInvalidUserID       = errors.New("user id is invalid")
	ErrRequiredCompanyName = errors.New("company_name cannot be empty")
	ErrRequiredPhone       = errors.New("phone cannot be empty")
	ErrRequiredCity        = errors.New("city cannot be empty")
)

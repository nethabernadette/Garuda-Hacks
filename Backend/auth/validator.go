package auth

import (
	"net/mail"
	"strings"

	"garuda-hacks/backend/users"
)

func normalizeRegisterRequest(req *RegisterRequest) {
	req.Role = users.UserRole(strings.ToUpper(strings.TrimSpace(string(req.Role))))
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
}

func normalizeLoginRequest(req *LoginRequest) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
}

func validateRegisterRequest(req RegisterRequest) error {
	if strings.TrimSpace(string(req.Role)) == "" {
		return ErrRequiredRole
	}
	if !req.Role.IsValid() {
		return ErrInvalidRole
	}
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	return validatePassword(req.Password)
}

func validateLoginRequest(req LoginRequest) error {
	if err := validateEmail(req.Email); err != nil {
		return err
	}
	return validatePassword(req.Password)
}

func validateEmail(email string) error {
	if email == "" {
		return ErrRequiredEmail
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrRequiredPassword
	}
	if len(password) < 8 {
		return ErrInvalidPassword
	}
	return nil
}

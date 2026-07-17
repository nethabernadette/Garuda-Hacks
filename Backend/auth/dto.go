package auth

import "garuda-hacks/backend/users"

type RegisterRequest struct {
	Role     users.UserRole `json:"role" validate:"required,oneof=BUYER PRODUCER ADMIN"`
	Email    string         `json:"email" validate:"required,email"`
	Password string         `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID    string         `json:"id"`
	Role  users.UserRole `json:"role"`
	Email string         `json:"email"`
}

type RegisterResponse struct {
	User UserResponse `json:"user"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewUserResponse(user *users.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Role:  user.Role,
		Email: user.Email,
	}
}

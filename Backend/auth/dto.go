package auth

import "garuda-hacks/backend/users"

type RegisterRequest struct {
	Role        users.UserRole `json:"role"`
	CompanyName string         `json:"company_name"`
	Email       string         `json:"email"`
	Password    string         `json:"password"`
	Phone       string         `json:"phone"`
	City        string         `json:"city"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          string         `json:"id"`
	Role        users.UserRole `json:"role"`
	CompanyName string         `json:"company_name"`
	Email       string         `json:"email"`
	Phone       string         `json:"phone"`
	City        string         `json:"city"`
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
	var profile users.UserProfile
	if user.Profile != nil {
		profile = *user.Profile
	}

	return UserResponse{
		ID:          user.ID,
		Role:        user.Role,
		CompanyName: profile.CompanyName,
		Email:       user.Email,
		Phone:       profile.Phone,
		City:        profile.City,
	}
}

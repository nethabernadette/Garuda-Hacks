package users

type Principal struct {
	UserID string
	Role   UserRole
}

type ProfileResponse struct {
	ID          string   `json:"id"`
	Role        UserRole `json:"role"`
	CompanyName string   `json:"company_name"`
	Email       string   `json:"email"`
	Phone       string   `json:"phone"`
	City        string   `json:"city"`
}

type UpdateProfileRequest struct {
	CompanyName *string `json:"company_name"`
	Phone       *string `json:"phone"`
	City        *string `json:"city"`
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func NewProfileResponse(user *User) ProfileResponse {
	var profile UserProfile
	if user.Profile != nil {
		profile = *user.Profile
	}

	return ProfileResponse{
		ID:          user.ID,
		Role:        user.Role,
		CompanyName: profile.CompanyName,
		Email:       user.Email,
		Phone:       profile.Phone,
		City:        profile.City,
	}
}

func NewProfileResponses(records []User) []ProfileResponse {
	responses := make([]ProfileResponse, 0, len(records))
	for i := range records {
		responses = append(responses, NewProfileResponse(&records[i]))
	}

	return responses
}

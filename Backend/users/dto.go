package users

import "time"

type Principal struct {
	UserID string
	Role   UserRole
}

type ProfileResponse struct {
	ID                string                    `json:"id"`
	Role              UserRole                  `json:"role"`
	Email             string                    `json:"email"`
	CompanyName       string                    `json:"company_name"`
	Phone             string                    `json:"phone"`
	City              string                    `json:"city"`
	BusinessType      string                    `json:"business_type,omitempty"`
	ProductCategory   string                    `json:"product_category,omitempty"`
	Capacity          string                    `json:"capacity,omitempty"`
	MOQ               string                    `json:"moq,omitempty"`
	Certifications    string                    `json:"certifications,omitempty"`
	DeliveryArea      string                    `json:"delivery_area,omitempty"`
	Availability      string                    `json:"availability,omitempty"`
	PurchaseFrequency string                    `json:"purchase_frequency,omitempty"`
	Verification      *NIBVerificationResponse  `json:"verification,omitempty"`
}

type UpdateProfileRequest struct {
	CompanyName       *string `json:"company_name" validate:"omitempty,min=1"`
	Phone             *string `json:"phone" validate:"omitempty,min=1"`
	City              *string `json:"city" validate:"omitempty,min=1"`
	BusinessType      *string `json:"business_type"`
	ProductCategory   *string `json:"product_category"`
	Capacity          *string `json:"capacity"`
	MOQ               *string `json:"moq"`
	Certifications    *string `json:"certifications"`
	DeliveryArea      *string `json:"delivery_area"`
	Availability      *string `json:"availability"`
	PurchaseFrequency *string `json:"purchase_frequency"`
}

type NIBVerificationRequest struct {
	NIBNumber string `json:"nib_number" validate:"required"`
}

type ReviewNIBVerificationRequest struct {
	Status          NIBVerificationStatus `json:"status" validate:"required,oneof=VERIFIED REJECTED"`
	RejectionReason string                `json:"rejection_reason,omitempty"`
}

type NIBVerificationResponse struct {
	ID              string                `json:"id"`
	UserID          string                `json:"user_id"`
	NIBNumber       string                `json:"nib_number"`
	Status          NIBVerificationStatus `json:"status"`
	VerifiedAt      *time.Time            `json:"verified_at,omitempty"`
	RejectedAt      *time.Time            `json:"rejected_at,omitempty"`
	RejectionReason string                `json:"rejection_reason,omitempty"`
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
		ID:                user.ID,
		Role:              user.Role,
		Email:             user.Email,
		CompanyName:       profile.CompanyName,
		Phone:             profile.Phone,
		City:              profile.City,
		BusinessType:      profile.BusinessType,
		ProductCategory:   profile.ProductCategory,
		Capacity:          profile.Capacity,
		MOQ:               profile.MOQ,
		Certifications:    profile.Certifications,
		DeliveryArea:      profile.DeliveryArea,
		Availability:      profile.Availability,
		PurchaseFrequency: profile.PurchaseFrequency,
		Verification:      NewNIBVerificationResponse(user.Verification),
	}
}

func NewProfileResponses(records []User) []ProfileResponse {
	responses := make([]ProfileResponse, 0, len(records))
	for i := range records {
		responses = append(responses, NewProfileResponse(&records[i]))
	}

	return responses
}

func NewNIBVerificationResponse(record *NIBVerification) *NIBVerificationResponse {
	if record == nil {
		return nil
	}

	return &NIBVerificationResponse{
		ID:              record.ID,
		UserID:          record.UserID,
		NIBNumber:       record.NIBNumber,
		Status:          record.Status,
		VerifiedAt:      record.VerifiedAt,
		RejectedAt:      record.RejectedAt,
		RejectionReason: record.RejectionReason,
	}
}

func NewNIBVerificationResponses(records []NIBVerification) []NIBVerificationResponse {
	responses := make([]NIBVerificationResponse, 0, len(records))
	for i := range records {
		response := NewNIBVerificationResponse(&records[i])
		if response != nil {
			responses = append(responses, *response)
		}
	}

	return responses
}

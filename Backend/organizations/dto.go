package organizations

import "time"

type CreateOrgRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateOrgRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type OrganizationResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type OrganizationMemberResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Role           string    `json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

type TransferOwnershipRequest struct {
	NewOwnerID string `json:"new_owner_id" validate:"required"`
}

func NewOrganizationResponse(org *Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		OwnerID:     org.OwnerID,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
	}
}

func NewOrganizationMemberResponse(member *OrganizationMember) OrganizationMemberResponse {
	return OrganizationMemberResponse{
		ID:             member.ID,
		OrganizationID: member.OrganizationID,
		UserID:         member.UserID,
		Role:           string(member.Role),
		JoinedAt:       member.JoinedAt,
	}
}

func NewOrganizationMemberResponses(members []OrganizationMember) []OrganizationMemberResponse {
	responses := make([]OrganizationMemberResponse, 0, len(members))
	for _, m := range members {
		responses = append(responses, NewOrganizationMemberResponse(&m))
	}
	return responses
}

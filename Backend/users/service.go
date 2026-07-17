package users

import (
	"context"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetCurrentProfile(ctx context.Context, principal Principal) (*ProfileResponse, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}

	user, err := s.repository.FindByID(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	response := NewProfileResponse(user)
	return &response, nil
}

func (s *Service) GetUserByID(ctx context.Context, principal Principal, id string) (*ProfileResponse, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if id == "" {
		return nil, ErrInvalidUserID
	}
	if principal.Role != RoleAdmin && principal.UserID != id {
		return nil, ErrForbidden
	}

	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := NewProfileResponse(user)
	return &response, nil
}

func (s *Service) ListUsers(ctx context.Context, principal Principal) ([]ProfileResponse, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if principal.Role != RoleAdmin {
		return nil, ErrForbidden
	}

	records, err := s.repository.List(ctx)
	if err != nil {
		return nil, err
	}

	return NewProfileResponses(records), nil
}

func (s *Service) UpdateCurrentProfile(ctx context.Context, principal Principal, req UpdateProfileRequest) (*ProfileResponse, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}

	return s.UpdateProfile(ctx, principal, principal.UserID, req)
}

func (s *Service) UpdateProfile(ctx context.Context, principal Principal, id string, req UpdateProfileRequest) (*ProfileResponse, error) {
	if principal.UserID == "" {
		return nil, ErrUnauthorized
	}
	if id == "" {
		return nil, ErrInvalidUserID
	}
	if principal.Role != RoleAdmin && principal.UserID != id {
		return nil, ErrForbidden
	}

	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := applyProfileUpdates(user, req); err != nil {
		return nil, err
	}

	if err := s.repository.UpdateProfile(ctx, user); err != nil {
		return nil, err
	}

	response := NewProfileResponse(user)
	return &response, nil
}

func applyProfileUpdates(user *User, req UpdateProfileRequest) error {
	if user.Profile == nil {
		user.Profile = &UserProfile{UserID: user.ID}
	}

	if req.CompanyName != nil {
		value := strings.TrimSpace(*req.CompanyName)
		if value == "" {
			return ErrRequiredCompanyName
		}
		user.Profile.CompanyName = value
	}

	if req.Phone != nil {
		value := strings.TrimSpace(*req.Phone)
		if value == "" {
			return ErrRequiredPhone
		}
		user.Profile.Phone = value
	}

	if req.City != nil {
		value := strings.TrimSpace(*req.City)
		if value == "" {
			return ErrRequiredCity
		}
		user.Profile.City = value
	}

	return nil
}

package organizations

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, userID string, req CreateOrgRequest) (*OrganizationResponse, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return nil, ErrRequiredOrgName
	}

	org := &Organization{
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		OwnerID:     userID,
	}

	if err := s.repository.Create(ctx, org); err != nil {
		return nil, err
	}

	res := NewOrganizationResponse(org)
	return &res, nil
}

func (s *Service) FindByID(ctx context.Context, id string) (*OrganizationResponse, error) {
	if id == "" {
		return nil, ErrInvalidOrgRequest
	}

	org, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := NewOrganizationResponse(org)
	return &res, nil
}

func (s *Service) Update(ctx context.Context, id string, userID string, req UpdateOrgRequest) (*OrganizationResponse, error) {
	if id == "" {
		return nil, ErrInvalidOrgRequest
	}

	membership, err := s.repository.GetMembership(ctx, id, userID)
	if err != nil {
		return nil, ErrUnauthorizedAction
	}

	if membership.Role != OrgRoleOwner && membership.Role != OrgRoleAdmin {
		return nil, ErrUnauthorizedAction
	}

	org, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrRequiredOrgName
		}
		org.Name = name
	}

	if req.Description != nil {
		org.Description = strings.TrimSpace(*req.Description)
	}

	if err := s.repository.Update(ctx, org); err != nil {
		return nil, err
	}

	res := NewOrganizationResponse(org)
	return &res, nil
}

func (s *Service) Join(ctx context.Context, id string, userID string) (*OrganizationMemberResponse, error) {
	if id == "" {
		return nil, ErrInvalidOrgRequest
	}

	// Check if already member
	existing, err := s.repository.GetMembership(ctx, id, userID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}

	// Double check if organization exists
	_, err = s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	member := &OrganizationMember{
		OrganizationID: id,
		UserID:         userID,
		Role:           OrgRoleMember,
	}

	if err := s.repository.AddMember(ctx, member); err != nil {
		return nil, err
	}

	res := NewOrganizationMemberResponse(member)
	return &res, nil
}

func (s *Service) Leave(ctx context.Context, id string, userID string) error {
	if id == "" {
		return ErrInvalidOrgRequest
	}

	membership, err := s.repository.GetMembership(ctx, id, userID)
	if err != nil {
		return ErrNotMember
	}

	if membership.Role == OrgRoleOwner {
		return ErrCannotLeaveOwner
	}

	return s.repository.RemoveMember(ctx, id, userID)
}

func (s *Service) TransferOwnership(ctx context.Context, id string, userID string, newOwnerID string) (*OrganizationResponse, error) {
	if id == "" || newOwnerID == "" {
		return nil, ErrInvalidOrgRequest
	}

	membership, err := s.repository.GetMembership(ctx, id, userID)
	if err != nil {
		return nil, ErrUnauthorizedAction
	}

	if membership.Role != OrgRoleOwner {
		return nil, ErrUnauthorizedAction
	}

	if userID == newOwnerID {
		return nil, ErrAlreadyOwner
	}

	// New owner must be a member
	_, err = s.repository.GetMembership(ctx, id, newOwnerID)
	if err != nil {
		return nil, errors.New("new owner must be a member of the organization first")
	}

	if err := s.repository.UpdateOwner(ctx, id, newOwnerID); err != nil {
		return nil, err
	}

	org, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	res := NewOrganizationResponse(org)
	return &res, nil
}

func (s *Service) ListMembers(ctx context.Context, id string) ([]OrganizationMemberResponse, error) {
	if id == "" {
		return nil, ErrInvalidOrgRequest
	}

	members, err := s.repository.ListMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	return NewOrganizationMemberResponses(members), nil
}

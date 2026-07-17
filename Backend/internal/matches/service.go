package matches

import (
	"context"
	"strings"

	"garuda-hacks/backend/users"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateInterest(ctx context.Context, userID string, role users.UserRole, req InterestRequest) (*MatchResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}

	buyerID := ""
	producerID := ""
	switch role {
	case users.RoleBuyer:
		buyerID = userID
		id, err := s.resolveProducerID(ctx, userID, req)
		if err != nil {
			return nil, err
		}
		producerID = id
	case users.RoleProducer, users.RoleFarmer:
		producerID = userID
		id, err := s.resolveBuyerID(ctx, userID, req)
		if err != nil {
			return nil, err
		}
		buyerID = id
	default:
		return nil, ErrUnsupportedRole
	}

	if buyerID == "" || producerID == "" {
		return nil, ErrPartnerRequired
	}
	if buyerID == producerID {
		return nil, ErrCannotMatchSelf
	}

	match, err := s.repository.FindOrCreate(ctx, buyerID, producerID)
	if err != nil {
		return nil, err
	}
	response := newMatchResponse(match)
	return &response, nil
}

func (s *Service) GetMatch(ctx context.Context, userID string, matchID string) (*MatchResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(matchID) == "" {
		return nil, ErrInvalidMatchID
	}
	match, err := s.repository.FindByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return nil, ErrForbidden
	}
	response := newMatchResponse(match)
	return &response, nil
}

func (s *Service) ListMatches(ctx context.Context, userID string, page int, limit int) ([]MatchResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	records, err := s.repository.ListForUser(ctx, userID, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}
	return newMatchResponses(records), nil
}

func (s *Service) resolveProducerID(ctx context.Context, currentUserID string, req InterestRequest) (string, error) {
	if strings.TrimSpace(req.SupplyPostID) != "" {
		return s.repository.FindProducerIDBySupplyPostID(ctx, req.SupplyPostID)
	}
	if strings.TrimSpace(req.PartnerID) != "" {
		user, err := s.repository.FindUserByID(ctx, req.PartnerID)
		if err != nil {
			return "", err
		}
		if user.Role != users.RoleProducer && user.Role != users.RoleFarmer {
			return "", ErrUnsupportedRole
		}
		return user.ID, nil
	}
	return s.repository.FindFirstActiveSupplyProducerID(ctx, currentUserID)
}

func (s *Service) resolveBuyerID(ctx context.Context, currentUserID string, req InterestRequest) (string, error) {
	if strings.TrimSpace(req.DemandPostID) != "" {
		return s.repository.FindBuyerIDByDemandPostID(ctx, req.DemandPostID)
	}
	if strings.TrimSpace(req.PartnerID) != "" {
		user, err := s.repository.FindUserByID(ctx, req.PartnerID)
		if err != nil {
			return "", err
		}
		if user.Role != users.RoleBuyer {
			return "", ErrUnsupportedRole
		}
		return user.ID, nil
	}
	return s.repository.FindFirstOpenDemandBuyerID(ctx, currentUserID)
}

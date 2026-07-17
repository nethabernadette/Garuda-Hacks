package agreement

import (
	"context"
	"errors"
	"strings"

	"garuda-hacks/backend/users"
)

var (
	ErrUnauthorized            = errors.New("authentication is required")
	ErrForbidden               = errors.New("forbidden: user is not part of this match")
	ErrInvalidMatchID          = errors.New("match id is invalid")
	ErrInvalidAgreementID      = errors.New("agreement id is invalid")
	ErrInvalidAgreementItemID  = errors.New("agreement item id is invalid")
	ErrMatchNotFound           = errors.New("match not found")
	ErrAgreementNotFound       = errors.New("agreement not found")
	ErrAgreementItemNotFound   = errors.New("agreement item not found")
	ErrActiveAgreementExists   = errors.New("active agreement already exists for this match")
	ErrAgreementNotEditable    = errors.New("agreement is not editable")
	ErrAgreementNeedsItems     = errors.New("agreement must have at least one item")
	ErrContactHidden           = errors.New("contact information is hidden until agreement is confirmed")
	ErrContactNotFound         = errors.New("contact information not found")
	ErrInvalidRequest          = errors.New("invalid request")
	ErrRequiredProductName     = errors.New("product_name is required")
	ErrInvalidQuantity         = errors.New("quantity must be greater than zero")
	ErrRequiredUnit            = errors.New("unit is required")
	ErrInvalidUnitPrice        = errors.New("unit_price must be greater than zero")
	ErrRequiredCurrency        = errors.New("currency is required")
	ErrRequiredDeliveryDate    = errors.New("delivery_date is required")
	ErrRequiredDeliveryAddress = errors.New("delivery_address is required")
	ErrRequiredPaymentTerms    = errors.New("payment_terms is required")
)

// Service contains agreement business logic.
type Service struct {
	repository Repository
}

// NewService creates an agreement service.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// CreateAgreement creates a draft agreement for a match.
func (s *Service) CreateAgreement(ctx context.Context, userID string, req CreateAgreementRequest) (*AgreementResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if err := validateAgreementRequest(req); err != nil {
		return nil, err
	}
	match, err := s.authorizeMatchAccess(ctx, req.MatchID, userID)
	if err != nil {
		return nil, err
	}

	items := agreementItemsFromRequests(req.Items)
	agreement := &Agreement{
		MatchID:   match.ID,
		CreatedBy: userID,
		Status:    AgreementStatusDraft,
		Items:     items,
	}
	if err := s.repository.CreateAgreement(ctx, agreement); err != nil {
		return nil, err
	}

	response := newAgreementResponse(agreement)
	return &response, nil
}

// GetAgreement returns an agreement if the user belongs to its match.
func (s *Service) GetAgreement(ctx context.Context, userID string, agreementID string) (*AgreementResponse, error) {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	response := newAgreementResponse(agreement)
	return &response, nil
}

// UpdateAgreement replaces a draft agreement's items.
func (s *Service) UpdateAgreement(ctx context.Context, userID string, agreementID string, req UpdateAgreementRequest) (*AgreementResponse, error) {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	if err := ensureEditable(agreement); err != nil {
		return nil, err
	}
	if err := validateUpdateAgreementRequest(req); err != nil {
		return nil, err
	}

	items := agreementItemsFromRequests(req.Items)
	agreement.BuyerConfirmed = false
	agreement.ProducerConfirmed = false
	agreement.Status = AgreementStatusDraft
	if err := s.repository.ReplaceAgreementItems(ctx, agreement, items); err != nil {
		return nil, err
	}

	updated, err := s.repository.FindAgreementByID(ctx, agreement.ID)
	if err != nil {
		return nil, err
	}
	response := newAgreementResponse(updated)
	return &response, nil
}

// CancelAgreement cancels an editable agreement.
func (s *Service) CancelAgreement(ctx context.Context, userID string, agreementID string) error {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return err
	}
	if agreement.Status == AgreementStatusConfirmed {
		return ErrAgreementNotEditable
	}
	if agreement.Status == AgreementStatusCancelled {
		return nil
	}

	agreement.Status = AgreementStatusCancelled
	return s.repository.SaveAgreement(ctx, agreement)
}

// ConfirmAgreement records the current user's confirmation.
func (s *Service) ConfirmAgreement(ctx context.Context, userID string, agreementID string, _ ConfirmAgreementRequest) (*AgreementResponse, error) {
	agreement, match, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	if agreement.Status == AgreementStatusCancelled {
		return nil, ErrAgreementNotEditable
	}
	if len(agreement.Items) == 0 {
		return nil, ErrAgreementNeedsItems
	}

	if userID == match.BuyerID {
		agreement.BuyerConfirmed = true
	}
	if userID == match.ProducerID {
		agreement.ProducerConfirmed = true
	}

	if agreement.BuyerConfirmed && agreement.ProducerConfirmed {
		agreement.Status = AgreementStatusConfirmed
	} else {
		agreement.Status = AgreementStatusPending
	}

	if err := s.repository.SaveAgreement(ctx, agreement); err != nil {
		return nil, err
	}

	response := newAgreementResponse(agreement)
	return &response, nil
}

// ListItems returns agreement items.
func (s *Service) ListItems(ctx context.Context, userID string, agreementID string) ([]AgreementItemResponse, error) {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	items, err := s.repository.ListAgreementItems(ctx, agreement.ID)
	if err != nil {
		return nil, err
	}
	return newAgreementItemResponses(items), nil
}

// AddItem adds an item to a draft agreement.
func (s *Service) AddItem(ctx context.Context, userID string, agreementID string, req AgreementItemRequest) (*AgreementItemResponse, error) {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	if err := ensureEditable(agreement); err != nil {
		return nil, err
	}
	if err := validateItemRequest(req); err != nil {
		return nil, err
	}

	item := agreementItemFromRequest(req)
	item.AgreementID = agreement.ID
	if err := s.repository.CreateAgreementItem(ctx, &item); err != nil {
		return nil, err
	}

	response := newAgreementItemResponse(&item)
	return &response, nil
}

// UpdateItem updates a draft agreement item.
func (s *Service) UpdateItem(ctx context.Context, userID string, agreementID string, itemID string, req AgreementItemRequest) (*AgreementItemResponse, error) {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	if err := ensureEditable(agreement); err != nil {
		return nil, err
	}
	if strings.TrimSpace(itemID) == "" {
		return nil, ErrInvalidAgreementItemID
	}
	if err := validateItemRequest(req); err != nil {
		return nil, err
	}

	item, err := s.repository.FindAgreementItemByID(ctx, agreement.ID, itemID)
	if err != nil {
		return nil, err
	}

	updated := agreementItemFromRequest(req)
	item.ProductName = updated.ProductName
	item.Quantity = updated.Quantity
	item.Unit = updated.Unit
	item.UnitPrice = updated.UnitPrice
	item.Currency = updated.Currency
	item.DeliveryDate = updated.DeliveryDate
	item.DeliveryAddress = updated.DeliveryAddress
	item.PaymentTerms = updated.PaymentTerms
	item.Specification = updated.Specification
	item.AdditionalNotes = updated.AdditionalNotes

	if err := s.repository.UpdateAgreementItem(ctx, item); err != nil {
		return nil, err
	}

	response := newAgreementItemResponse(item)
	return &response, nil
}

// DeleteItem deletes a draft agreement item.
func (s *Service) DeleteItem(ctx context.Context, userID string, agreementID string, itemID string) error {
	agreement, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return err
	}
	if err := ensureEditable(agreement); err != nil {
		return err
	}
	if strings.TrimSpace(itemID) == "" {
		return ErrInvalidAgreementItemID
	}

	item, err := s.repository.FindAgreementItemByID(ctx, agreement.ID, itemID)
	if err != nil {
		return err
	}
	itemCount, err := s.repository.CountAgreementItems(ctx, agreement.ID)
	if err != nil {
		return err
	}
	if itemCount <= 1 {
		return ErrAgreementNeedsItems
	}
	return s.repository.DeleteAgreementItem(ctx, item)
}

// GetContact returns both parties' contact information for confirmed agreements.
func (s *Service) GetContact(ctx context.Context, userID string, agreementID string) (*ContactResponse, error) {
	agreement, match, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	if agreement.Status != AgreementStatusConfirmed {
		return nil, ErrContactHidden
	}

	buyer, err := s.repository.FindUserContactByID(ctx, match.BuyerID)
	if err != nil {
		return nil, err
	}
	producer, err := s.repository.FindUserContactByID(ctx, match.ProducerID)
	if err != nil {
		return nil, err
	}

	return &ContactResponse{
		AgreementID: agreement.ID,
		MatchID:     agreement.MatchID,
		Buyer:       contactFromUser(buyer),
		Producer:    contactFromUser(producer),
	}, nil
}

func (s *Service) authorizedAgreement(ctx context.Context, userID string, agreementID string) (*Agreement, *matchRecord, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, nil, ErrUnauthorized
	}
	if strings.TrimSpace(agreementID) == "" {
		return nil, nil, ErrInvalidAgreementID
	}

	agreement, err := s.repository.FindAgreementByID(ctx, agreementID)
	if err != nil {
		return nil, nil, err
	}
	match, err := s.authorizeMatchAccess(ctx, agreement.MatchID, userID)
	if err != nil {
		return nil, nil, err
	}
	return agreement, match, nil
}

func (s *Service) authorizeMatchAccess(ctx context.Context, matchID string, userID string) (*matchRecord, error) {
	if strings.TrimSpace(matchID) == "" {
		return nil, ErrInvalidMatchID
	}

	match, err := s.repository.FindMatchByID(ctx, matchID)
	if err != nil {
		return nil, err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return nil, ErrForbidden
	}
	return match, nil
}

func ensureEditable(agreement *Agreement) error {
	if agreement.Status != AgreementStatusDraft {
		return ErrAgreementNotEditable
	}
	return nil
}

func contactFromUser(user *users.User) ContactPartyResponse {
	response := ContactPartyResponse{
		UserID: user.ID,
		Email:  user.Email,
	}
	if user.Profile != nil {
		response.CompanyName = user.Profile.CompanyName
		response.PhoneNumber = user.Profile.Phone
		response.BusinessAddress = user.Profile.DeliveryArea
		if response.BusinessAddress == "" {
			response.BusinessAddress = user.Profile.City
		}
	}
	return response
}

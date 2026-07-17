package document

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	agreementmodule "garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/users"
)

var (
	ErrUnauthorized       = errors.New("authentication is required")
	ErrForbidden          = errors.New("forbidden: user is not part of this agreement")
	ErrInvalidAgreementID = errors.New("agreement id is invalid")
	ErrMatchNotFound      = errors.New("match not found")
	ErrAgreementNotReady  = errors.New("agreement must be confirmed by both parties before document or contact can be revealed")
	ErrAgreementCancelled = errors.New("agreement has been cancelled")
	ErrContactNotFound    = errors.New("contact information not found")
)

// Service contains procurement document and contact reveal business logic.
type Service struct {
	repository Repository
	generator  Generator
	now        func() time.Time
}

// NewService creates a document service.
func NewService(repository Repository, generator Generator) *Service {
	if generator == nil {
		generator = NewHTMLGenerator()
	}
	return &Service{
		repository: repository,
		generator:  generator,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

// GetDocument returns the structured procurement summary plus printable HTML.
func (s *Service) GetDocument(ctx context.Context, userID string, agreementID string) (*DocumentResponse, error) {
	summary, err := s.GetSummary(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	html, err := s.generator.RenderProcurementSummary(*summary)
	if err != nil {
		return nil, err
	}
	return &DocumentResponse{
		DocumentNumber: summary.DocumentNumber,
		AgreementID:    summary.AgreementID,
		GeneratedDate:  summary.GeneratedDate,
		Summary:        *summary,
		HTML:           html,
	}, nil
}

// GetHTML returns the printable procurement summary HTML.
func (s *Service) GetHTML(ctx context.Context, userID string, agreementID string) (string, error) {
	summary, err := s.GetSummary(ctx, userID, agreementID)
	if err != nil {
		return "", err
	}
	return s.generator.RenderProcurementSummary(*summary)
}

// GetSummary returns a structured procurement summary for a confirmed agreement.
func (s *Service) GetSummary(ctx context.Context, userID string, agreementID string) (*ProcurementSummaryResponse, error) {
	agreement, match, err := s.confirmedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}

	buyer, err := s.repository.FindUserContactByID(ctx, match.BuyerID)
	if err != nil {
		return nil, ErrContactNotFound
	}
	producer, err := s.repository.FindUserContactByID(ctx, match.ProducerID)
	if err != nil {
		return nil, ErrContactNotFound
	}

	documentNumber, err := s.documentNumber(ctx, agreement)
	if err != nil {
		return nil, err
	}

	items := make([]ProcurementSummaryItemResponse, 0, len(agreement.Items))
	total := 0.0
	currency := "IDR"
	deliveryAddress := ""
	paymentTerms := ""
	notes := make([]string, 0)
	for i := range agreement.Items {
		item := agreement.Items[i]
		lineTotal := item.Quantity * item.UnitPrice
		total += lineTotal
		if item.Currency != "" {
			currency = item.Currency
		}
		if deliveryAddress == "" {
			deliveryAddress = item.DeliveryAddress
		}
		if paymentTerms == "" {
			paymentTerms = item.PaymentTerms
		}
		if strings.TrimSpace(item.AdditionalNotes) != "" {
			notes = append(notes, strings.TrimSpace(item.AdditionalNotes))
		}
		items = append(items, ProcurementSummaryItemResponse{
			ProductName:     item.ProductName,
			Quantity:        item.Quantity,
			Unit:            item.Unit,
			UnitPrice:       item.UnitPrice,
			Currency:        item.Currency,
			TotalValue:      lineTotal,
			Specifications:  item.Specification,
			DeliveryDate:    item.DeliveryDate,
			DeliveryAddress: item.DeliveryAddress,
			PaymentTerms:    item.PaymentTerms,
			AdditionalNotes: item.AdditionalNotes,
		})
	}

	return &ProcurementSummaryResponse{
		DocumentNumber:                documentNumber,
		GeneratedDate:                 s.now(),
		AgreementID:                   agreement.ID,
		ProducerCompany:               companyName(producer),
		BuyerCompany:                  companyName(buyer),
		ProductList:                   items,
		TotalValue:                    total,
		Currency:                      currency,
		DeliveryAddress:               deliveryAddress,
		PaymentTerms:                  paymentTerms,
		AdditionalNotes:               strings.Join(notes, "; "),
		AgreementStatus:               string(agreement.Status),
		ProducerConfirmationTimestamp: agreement.ProducerConfirmedAt,
		BuyerConfirmationTimestamp:    agreement.BuyerConfirmedAt,
	}, nil
}

// GetContact returns both parties' company contact information after confirmation.
func (s *Service) GetContact(ctx context.Context, userID string, agreementID string) (*ContactResponse, error) {
	agreement, match, err := s.confirmedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}

	buyer, err := s.repository.FindUserContactByID(ctx, match.BuyerID)
	if err != nil {
		return nil, ErrContactNotFound
	}
	producer, err := s.repository.FindUserContactByID(ctx, match.ProducerID)
	if err != nil {
		return nil, ErrContactNotFound
	}

	return &ContactResponse{
		AgreementID: agreement.ID,
		MatchID:     agreement.MatchID,
		Buyer:       contactFromUser(buyer),
		Producer:    contactFromUser(producer),
	}, nil
}

func (s *Service) confirmedAgreement(ctx context.Context, userID string, agreementID string) (*agreementmodule.Agreement, *MatchRecord, error) {
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
	match, err := s.repository.FindMatchByID(ctx, agreement.MatchID)
	if err != nil {
		return nil, nil, err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return nil, nil, ErrForbidden
	}
	if agreement.Status == agreementmodule.AgreementStatusCancelled {
		return nil, nil, ErrAgreementCancelled
	}
	if agreement.Status != agreementmodule.AgreementStatusConfirmed || !agreement.BuyerConfirmed || !agreement.ProducerConfirmed {
		return nil, nil, ErrAgreementNotReady
	}
	return agreement, match, nil
}

func (s *Service) documentNumber(ctx context.Context, agreement *agreementmodule.Agreement) (string, error) {
	year := agreement.CreatedAt.Year()
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	sequence, err := s.repository.CountAgreementsInYearThrough(ctx, start, end, agreement.CreatedAt)
	if err != nil {
		return "", err
	}
	if sequence < 1 {
		sequence = 1
	}
	return fmt.Sprintf("RFQ-%d-%06d", year, sequence), nil
}

func contactFromUser(user *users.User) ContactPartyResponse {
	response := ContactPartyResponse{
		UserID: user.ID,
		Email:  user.Email,
	}
	if user.Profile != nil {
		response.CompanyName = user.Profile.CompanyName
		response.PhoneNumber = user.Profile.Phone
		response.BusinessAddress = firstNonEmpty(user.Profile.DeliveryArea, user.Profile.City)
		response.BusinessRepresentative = user.Profile.CompanyName
	}
	return response
}

func companyName(user *users.User) string {
	if user != nil && user.Profile != nil && strings.TrimSpace(user.Profile.CompanyName) != "" {
		return strings.TrimSpace(user.Profile.CompanyName)
	}
	if user != nil {
		return user.Email
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

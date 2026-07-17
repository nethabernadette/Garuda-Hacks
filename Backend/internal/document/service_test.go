package document

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	agreementmodule "garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/users"
)

type fakeRepository struct {
	agreement *agreementmodule.Agreement
	match     *MatchRecord
	buyer     *users.User
	producer  *users.User
}

func (r *fakeRepository) FindAgreementByID(ctx context.Context, id string) (*agreementmodule.Agreement, error) {
	if r.agreement == nil || r.agreement.ID != id {
		return nil, agreementmodule.ErrAgreementNotFound
	}
	return r.agreement, nil
}

func (r *fakeRepository) FindMatchByID(ctx context.Context, id string) (*MatchRecord, error) {
	if r.match == nil || r.match.ID != id {
		return nil, ErrMatchNotFound
	}
	return r.match, nil
}

func (r *fakeRepository) FindUserContactByID(ctx context.Context, userID string) (*users.User, error) {
	if r.buyer != nil && r.buyer.ID == userID {
		return r.buyer, nil
	}
	if r.producer != nil && r.producer.ID == userID {
		return r.producer, nil
	}
	return nil, ErrContactNotFound
}

func (r *fakeRepository) CountAgreementsInYearThrough(ctx context.Context, yearStart time.Time, yearEnd time.Time, createdAt time.Time) (int64, error) {
	return 7, nil
}

func TestGetSummaryRequiresBothConfirmations(t *testing.T) {
	repo := confirmedFixture()
	repo.agreement.ProducerConfirmed = false
	repo.agreement.ProducerConfirmedAt = nil
	service := NewService(repo, NewHTMLGenerator())

	_, err := service.GetSummary(context.Background(), "buyer-1", "agreement-1")
	if !errors.Is(err, ErrAgreementNotReady) {
		t.Fatalf("expected ErrAgreementNotReady, got %v", err)
	}
}

func TestGetSummaryBuildsProcurementSummary(t *testing.T) {
	service := NewService(confirmedFixture(), NewHTMLGenerator())

	summary, err := service.GetSummary(context.Background(), "buyer-1", "agreement-1")
	if err != nil {
		t.Fatalf("unexpected summary error: %v", err)
	}
	if summary.DocumentNumber != "RFQ-2026-000007" {
		t.Fatalf("unexpected document number %q", summary.DocumentNumber)
	}
	if summary.TotalValue != 500000 {
		t.Fatalf("unexpected total value %v", summary.TotalValue)
	}
	if summary.BuyerCompany != "Buyer Co" || summary.ProducerCompany != "Producer Co" {
		t.Fatalf("unexpected parties: buyer=%q producer=%q", summary.BuyerCompany, summary.ProducerCompany)
	}
}

func TestGeneratorRendersPrintableHTML(t *testing.T) {
	service := NewService(confirmedFixture(), NewHTMLGenerator())

	html, err := service.GetHTML(context.Background(), "producer-1", "agreement-1")
	if err != nil {
		t.Fatalf("unexpected html error: %v", err)
	}
	if !strings.Contains(html, "<!doctype html>") || !strings.Contains(html, "Procurement Summary") || !strings.Contains(html, "RFQ-2026-000007") {
		t.Fatalf("html does not contain expected printable document content")
	}
}

func TestContactRevealRequiresParticipant(t *testing.T) {
	service := NewService(confirmedFixture(), NewHTMLGenerator())

	_, err := service.GetContact(context.Background(), "outsider", "agreement-1")
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func confirmedFixture() *fakeRepository {
	buyerConfirmedAt := time.Date(2026, 7, 10, 9, 0, 0, 0, time.UTC)
	producerConfirmedAt := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	return &fakeRepository{
		agreement: &agreementmodule.Agreement{
			ID:                  "agreement-1",
			MatchID:             "match-1",
			Status:              agreementmodule.AgreementStatusConfirmed,
			BuyerConfirmed:      true,
			ProducerConfirmed:   true,
			BuyerConfirmedAt:    &buyerConfirmedAt,
			ProducerConfirmedAt: &producerConfirmedAt,
			CreatedAt:           time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC),
			Items: []agreementmodule.AgreementItem{{
				ProductName:     "Rice",
				Quantity:        100,
				Unit:            "kg",
				UnitPrice:       5000,
				Currency:        "IDR",
				DeliveryDate:    time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
				DeliveryAddress: "Jakarta",
				PaymentTerms:    "Net 14",
				Specification:   "Food grade",
			}},
		},
		match: &MatchRecord{ID: "match-1", BuyerID: "buyer-1", ProducerID: "producer-1"},
		buyer: &users.User{
			ID:    "buyer-1",
			Email: "buyer@example.com",
			Profile: &users.UserProfile{
				CompanyName:  "Buyer Co",
				Phone:        "+621",
				City:         "Jakarta",
				DeliveryArea: "Jakarta",
			},
		},
		producer: &users.User{
			ID:    "producer-1",
			Email: "producer@example.com",
			Profile: &users.UserProfile{
				CompanyName:  "Producer Co",
				Phone:        "+622",
				City:         "Bandung",
				DeliveryArea: "West Java",
			},
		},
	}
}

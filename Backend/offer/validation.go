package offer

import (
	"errors"
	"time"
)

var (
	ErrOfferNotFound     = errors.New("offer not found")
	ErrOfferForbidden    = errors.New("producer is not allowed to modify this offer")
	ErrDuplicateOffer    = errors.New("producer already submitted an offer for this demand group")
	ErrOfferCancelled    = errors.New("cancelled offer cannot be modified")
	ErrDemandGroupIDZero = errors.New("demand group id is required")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func validateOffer(offer *Offer) error {
	if offer.DemandGroupID == 0 {
		return ErrDemandGroupIDZero
	}
	if offer.OfferedPrice <= 0 {
		return ValidationError{Message: "price must be greater than zero"}
	}
	if isBeforeToday(offer.EstimatedDeliveryDate) {
		return ValidationError{Message: "estimated delivery cannot be before today"}
	}
	if offer.Status == "" {
		offer.Status = OfferStatusPending
	}
	return nil
}

func isBeforeToday(value time.Time) bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	date := value.In(now.Location())
	startOfDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, now.Location())
	return startOfDate.Before(today)
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit, (page - 1) * limit
}

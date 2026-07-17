package offer

import "context"

type OfferService interface {
	Create(ctx context.Context, producerID uint, req CreateOfferRequest) (*Offer, error)
	Update(ctx context.Context, producerID uint, id uint, req UpdateOfferRequest) (*Offer, error)
	Cancel(ctx context.Context, producerID uint, id uint) error
	GetByID(ctx context.Context, id uint) (*Offer, error)
	ListProducerOffers(ctx context.Context, producerID uint, page int, limit int) ([]Offer, error)
	ListByDemandGroup(ctx context.Context, demandGroupID uint, page int, limit int) ([]Offer, error)
}

type offerService struct {
	repository OfferRepository
}

func NewOfferService(repository OfferRepository) OfferService {
	return &offerService{repository: repository}
}

func (s *offerService) Create(ctx context.Context, producerID uint, req CreateOfferRequest) (*Offer, error) {
	offer := &Offer{
		DemandGroupID: req.DemandGroupID,
		ProducerID:    producerID,
		OfferedPrice:  req.OfferedPrice,
		Notes:         req.Notes,
		Status:        OfferStatusPending,
	}
	if req.EstimatedDeliveryDate != nil {
		offer.EstimatedDeliveryDate = req.EstimatedDeliveryDate.Time
	}

	if err := validateOffer(offer); err != nil {
		return nil, err
	}
	if err := s.repository.CreateIfNoDuplicate(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *offerService) Update(ctx context.Context, producerID uint, id uint, req UpdateOfferRequest) (*Offer, error) {
	offer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if offer.ProducerID != producerID {
		return nil, ErrOfferForbidden
	}
	if offer.Status == OfferStatusCancelled {
		return nil, ErrOfferCancelled
	}

	if req.OfferedPrice != nil {
		offer.OfferedPrice = *req.OfferedPrice
	}
	if req.EstimatedDeliveryDate != nil {
		offer.EstimatedDeliveryDate = req.EstimatedDeliveryDate.Time
	}
	if req.Notes != nil {
		offer.Notes = *req.Notes
	}

	if err := validateOffer(offer); err != nil {
		return nil, err
	}
	if err := s.repository.Update(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *offerService) Cancel(ctx context.Context, producerID uint, id uint) error {
	offer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if offer.ProducerID != producerID {
		return ErrOfferForbidden
	}
	if offer.Status == OfferStatusCancelled {
		return nil
	}

	offer.Status = OfferStatusCancelled
	return s.repository.Update(ctx, offer)
}

func (s *offerService) GetByID(ctx context.Context, id uint) (*Offer, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *offerService) ListProducerOffers(ctx context.Context, producerID uint, page int, limit int) ([]Offer, error) {
	normalizedLimit, offset := normalizePagination(page, limit)
	return s.repository.ListByProducer(ctx, producerID, normalizedLimit, offset)
}

func (s *offerService) ListByDemandGroup(ctx context.Context, demandGroupID uint, page int, limit int) ([]Offer, error) {
	normalizedLimit, offset := normalizePagination(page, limit)
	return s.repository.ListByDemandGroup(ctx, demandGroupID, normalizedLimit, offset)
}

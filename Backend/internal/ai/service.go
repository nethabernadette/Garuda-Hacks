package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"garuda-hacks/backend/internal/agreement"
	"garuda-hacks/backend/internal/chat"
	"garuda-hacks/backend/posts"
	"garuda-hacks/backend/users"
)

const (
	defaultCandidateLimit = 20
	historyLimit          = 50
	messageLimit          = 80
)

// Service contains AI orchestration logic.
type Service struct {
	repository Repository
	groq       *GroqClient
}

// NewService creates an AI service.
func NewService(repository Repository, groq *GroqClient) *Service {
	return &Service{repository: repository, groq: groq}
}

// TrackSearch stores non-sensitive search history for future personalization.
func (s *Service) TrackSearch(ctx context.Context, userID string, req SearchHistoryInput) (*SearchHistoryResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, ErrInvalidRequest
	}
	record := &SearchHistory{
		UserID:   userID,
		Query:    query,
		Category: strings.TrimSpace(req.Category),
		Location: strings.TrimSpace(req.Location),
	}
	if err := s.repository.CreateSearchHistory(ctx, record); err != nil {
		return nil, err
	}
	return &SearchHistoryResponse{
		ID:        record.ID,
		Query:     record.Query,
		Category:  record.Category,
		Location:  record.Location,
		CreatedAt: record.CreatedAt,
	}, nil
}

// Recommendations returns homepage recommendations using Groq with deterministic fallback.
func (s *Service) Recommendations(ctx context.Context, userID string) (*RecommendationResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	searches, err := s.repository.ListSearchHistory(ctx, userID, historyLimit)
	if err != nil {
		return nil, err
	}
	agreements, err := s.repository.ListAgreementsForUser(ctx, userID, historyLimit)
	if err != nil {
		return nil, err
	}
	candidates, err := s.recommendationCandidates(ctx, user)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"user_profile":        publicUserProfile(user),
		"search_history":      searchHistoryPayload(searches),
		"agreement_history":   agreementHistoryPayload(agreements),
		"eligible_candidates": candidates,
		"backend_aggregates":  buildAggregates(user, searches, agreements),
	}

	response := fallbackRecommendation(user, searches, agreements, candidates)
	if s.groq != nil {
		var aiResponse RecommendationResponse
		if err := s.groq.CompleteJSON(ctx, recommendationPrompt, payload, &aiResponse); err == nil {
			response = sanitizeRecommendation(aiResponse, response, candidateIDSet(candidates))
		}
	}
	return &response, nil
}

// Matchmaking ranks eligible opposite-side posts for the authenticated user's latest or requested post.
func (s *Service) Matchmaking(ctx context.Context, userID string, role users.UserRole, req MatchmakingRequest) (*MatchmakingResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	limit := normalizeLimit(req.Limit)
	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var source interface{}
	var sourceID string
	var sourceType string
	var candidates []map[string]interface{}
	var evaluations []matchEvaluation

	switch role {
	case users.RoleBuyer:
		post, err := s.buyerSourcePost(ctx, userID, req.PostID)
		if err != nil {
			return nil, err
		}
		source = demandPostPayload(*post)
		sourceID = post.ID
		sourceType = "demand_post"
		records, err := s.repository.ListSupplyCandidates(ctx, userID, limit)
		if err != nil {
			return nil, err
		}
		evaluations, err = s.evaluateSupplyCandidates(ctx, *post, user, records)
		if err != nil {
			return nil, err
		}
		candidates = matchEvaluationPayloads(evaluations)
	case users.RoleProducer, users.RoleFarmer:
		post, err := s.producerSourcePost(ctx, userID, req.PostID)
		if err != nil {
			return nil, err
		}
		source = supplyPostPayload(*post)
		sourceID = post.ID
		sourceType = "supply_post"
		records, err := s.repository.ListDemandCandidates(ctx, userID, limit)
		if err != nil {
			return nil, err
		}
		evaluations, err = s.evaluateDemandCandidates(ctx, *post, user, records)
		if err != nil {
			return nil, err
		}
		candidates = matchEvaluationPayloads(evaluations)
	default:
		return nil, ErrForbidden
	}

	payload := map[string]interface{}{
		"user_profile":              publicUserProfile(user),
		"source_post":               source,
		"deterministic_evaluations": candidates,
	}
	response := fallbackMatchmaking(sourceID, sourceType, evaluations)
	if s.groq != nil {
		var aiResponse MatchmakingResponse
		if err := s.groq.CompleteJSON(ctx, matchmakingPrompt, payload, &aiResponse); err == nil {
			response = sanitizeMatchmaking(aiResponse, response, sourceID, sourceType, evaluationIDSet(evaluations), evaluationScoreByID(evaluations), evaluationByID(evaluations))
		}
	}
	return &response, nil
}

// VerifyAgreement summarizes negotiation and verifies whether contact reveal is safe.
func (s *Service) VerifyAgreement(ctx context.Context, userID string, agreementID string) (*AgreementVerificationResponse, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(agreementID) == "" {
		return nil, ErrMissingAgreementID
	}

	record, err := s.repository.FindAgreementByID(ctx, agreementID)
	if err != nil {
		return nil, err
	}
	match, err := s.repository.FindMatchByID(ctx, record.MatchID)
	if err != nil {
		return nil, err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return nil, ErrForbidden
	}
	messages, err := s.repository.ListMessagesByMatchID(ctx, record.MatchID, messageLimit)
	if err != nil {
		return nil, err
	}

	storedSubmission := submissionFromAgreement(record)
	storedComparison := compareAgreementSubmissions(storedSubmission, storedSubmission)
	payload := map[string]interface{}{
		"buyer_submission":    storedSubmission,
		"producer_submission": storedSubmission,
		"backend_comparison":  storedComparison,
		"messages":            messagePayload(messages),
		"rules": map[string]interface{}{
			"contact_reveal_requires_status":             "CONFIRMED",
			"contact_reveal_requires_buyer_confirmed":    true,
			"contact_reveal_requires_producer_confirmed": true,
		},
	}
	response := fallbackAgreementVerification(record, messages)
	if s.groq != nil {
		var aiResponse AgreementVerificationResponse
		if err := s.groq.CompleteJSON(ctx, agreementVerificationPrompt, payload, &aiResponse); err == nil {
			response = sanitizeAgreementVerification(aiResponse, response, record)
		}
	}
	return &response, nil
}

// CompareAgreementSubmissions compares buyer and producer submissions for the same agreement.
func (s *Service) CompareAgreementSubmissions(ctx context.Context, userID string, agreementID string, req AgreementVerificationRequest) (*AgreementVerificationResponse, error) {
	record, match, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}

	deterministic := compareAgreementSubmissions(req.BuyerSubmission, req.ProducerSubmission)
	payload := map[string]interface{}{
		"agreement_id":           record.ID,
		"match_id":               record.MatchID,
		"buyer_submission":       normalizeSubmission(req.BuyerSubmission),
		"producer_submission":    normalizeSubmission(req.ProducerSubmission),
		"backend_comparison":     deterministic,
		"buyer_final_confirm":    req.BuyerFinalConfirm,
		"producer_final_confirm": req.ProducerFinalConfirm,
		"rules": map[string]interface{}{
			"ai_may_only_recommend":     true,
			"backend_controls_unlock":   true,
			"buyer_is_party":            match.BuyerID != "",
			"producer_is_party":         match.ProducerID != "",
			"agreement_status":          record.Status,
			"stored_buyer_confirmed":    record.BuyerConfirmed,
			"stored_producer_confirmed": record.ProducerConfirmed,
		},
	}

	response := deterministic
	response.AgreementID = record.ID
	response.MatchID = record.MatchID
	if s.groq != nil {
		var aiResponse AgreementVerificationResponse
		if err := s.groq.CompleteJSON(ctx, agreementVerificationPrompt, payload, &aiResponse); err == nil {
			response = sanitizeSubmissionVerification(aiResponse, deterministic)
			response.AgreementID = record.ID
			response.MatchID = record.MatchID
		}
	}
	response.CanRevealContact = canRevealAfterVerification(record, req, deterministic, response)
	if !response.CanRevealContact && response.NextAction == "" {
		response.NextAction = "Resolve conflicting or missing terms, then ask both parties to final confirm."
	}
	return &response, nil
}

// SummarizeNegotiation summarizes messages from the agreement's match into a structured draft.
func (s *Service) SummarizeNegotiation(ctx context.Context, userID string, agreementID string) (*NegotiationSummaryResponse, error) {
	record, _, err := s.authorizedAgreement(ctx, userID, agreementID)
	if err != nil {
		return nil, err
	}
	messages, err := s.repository.ListMessagesByMatchID(ctx, record.MatchID, messageLimit)
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"agreement_id": record.ID,
		"match_id":     record.MatchID,
		"messages":     messagePayload(messages),
		"limit":        messageLimit,
	}
	response := fallbackNegotiationSummary(record, messages)
	if s.groq != nil {
		var aiResponse NegotiationSummaryResponse
		if err := s.groq.CompleteJSON(ctx, negotiationSummaryPrompt, payload, &aiResponse); err == nil {
			response = sanitizeNegotiationSummary(aiResponse, response)
		}
	}
	return &response, nil
}

func (s *Service) authorizedAgreement(ctx context.Context, userID string, agreementID string) (*agreement.Agreement, *matchRecord, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, nil, ErrUnauthorized
	}
	if strings.TrimSpace(agreementID) == "" {
		return nil, nil, ErrMissingAgreementID
	}
	record, err := s.repository.FindAgreementByID(ctx, agreementID)
	if err != nil {
		return nil, nil, err
	}
	match, err := s.repository.FindMatchByID(ctx, record.MatchID)
	if err != nil {
		return nil, nil, err
	}
	if match.BuyerID != userID && match.ProducerID != userID {
		return nil, nil, ErrForbidden
	}
	return record, match, nil
}

func (s *Service) recommendationCandidates(ctx context.Context, user *users.User) ([]map[string]interface{}, error) {
	switch user.Role {
	case users.RoleBuyer:
		records, err := s.repository.ListSupplyCandidates(ctx, user.ID, defaultCandidateLimit)
		if err != nil {
			return nil, err
		}
		return supplyCandidatePayloads(records), nil
	case users.RoleProducer, users.RoleFarmer:
		records, err := s.repository.ListDemandCandidates(ctx, user.ID, defaultCandidateLimit)
		if err != nil {
			return nil, err
		}
		return demandCandidatePayloads(records), nil
	default:
		return []map[string]interface{}{}, nil
	}
}

func (s *Service) buyerSourcePost(ctx context.Context, userID string, postID string) (*posts.DemandPost, error) {
	if strings.TrimSpace(postID) != "" {
		post, err := s.repository.FindDemandPost(ctx, postID)
		if err != nil {
			return nil, err
		}
		if post.BuyerID != userID {
			return nil, ErrForbidden
		}
		return post, nil
	}
	records, err := s.repository.ListUserDemandPosts(ctx, userID, 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, ErrPostNotFound
	}
	return &records[0], nil
}

func (s *Service) producerSourcePost(ctx context.Context, userID string, postID string) (*posts.SupplyPost, error) {
	if strings.TrimSpace(postID) != "" {
		post, err := s.repository.FindSupplyPost(ctx, postID)
		if err != nil {
			return nil, err
		}
		if post.ProducerID != userID {
			return nil, ErrForbidden
		}
		return post, nil
	}
	records, err := s.repository.ListUserSupplyPosts(ctx, userID, 1)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, ErrPostNotFound
	}
	return &records[0], nil
}

func normalizeLimit(limit int) int {
	if limit < 1 {
		return defaultCandidateLimit
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func publicUserProfile(user *users.User) map[string]interface{} {
	profile := map[string]interface{}{}
	if user.Profile != nil {
		profile = map[string]interface{}{
			"company_name":       user.Profile.CompanyName,
			"city":               user.Profile.City,
			"business_type":      user.Profile.BusinessType,
			"product_category":   user.Profile.ProductCategory,
			"capacity":           user.Profile.Capacity,
			"moq":                user.Profile.MOQ,
			"certifications":     user.Profile.Certifications,
			"delivery_area":      user.Profile.DeliveryArea,
			"availability":       user.Profile.Availability,
			"purchase_frequency": user.Profile.PurchaseFrequency,
		}
	}
	return map[string]interface{}{
		"user_id": user.ID,
		"role":    user.Role,
		"profile": profile,
	}
}

func searchHistoryPayload(records []SearchHistory) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		payload = append(payload, map[string]interface{}{
			"query":      record.Query,
			"category":   record.Category,
			"location":   record.Location,
			"created_at": record.CreatedAt,
		})
	}
	return payload
}

func agreementHistoryPayload(records []agreement.Agreement) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		payload = append(payload, agreementPayload(record))
	}
	return payload
}

func agreementPayload(record agreement.Agreement) map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(record.Items))
	for _, item := range record.Items {
		items = append(items, map[string]interface{}{
			"product_name":     item.ProductName,
			"quantity":         item.Quantity,
			"unit":             item.Unit,
			"unit_price":       item.UnitPrice,
			"currency":         item.Currency,
			"delivery_date":    item.DeliveryDate,
			"delivery_address": item.DeliveryAddress,
			"payment_terms":    item.PaymentTerms,
			"specification":    item.Specification,
			"additional_notes": item.AdditionalNotes,
		})
	}
	return map[string]interface{}{
		"agreement_id":          record.ID,
		"match_id":              record.MatchID,
		"status":                record.Status,
		"buyer_confirmed":       record.BuyerConfirmed,
		"producer_confirmed":    record.ProducerConfirmed,
		"buyer_confirmed_at":    record.BuyerConfirmedAt,
		"producer_confirmed_at": record.ProducerConfirmedAt,
		"created_at":            record.CreatedAt,
		"items":                 items,
	}
}

func messagePayload(records []chat.Message) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		payload = append(payload, map[string]interface{}{
			"sender_id":  record.SenderID,
			"message":    record.Message,
			"created_at": record.CreatedAt,
		})
	}
	return payload
}

func supplyCandidatePayloads(records []posts.SupplyPost) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		payload = append(payload, supplyPostPayload(record))
	}
	return payload
}

func demandCandidatePayloads(records []posts.DemandPost) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(records))
	for _, record := range records {
		payload = append(payload, demandPostPayload(record))
	}
	return payload
}

func supplyPostPayload(record posts.SupplyPost) map[string]interface{} {
	return map[string]interface{}{
		"entity_id":              record.ID,
		"entity_type":            "supply_post",
		"producer_id":            record.ProducerID,
		"product_name":           record.ProductName,
		"category":               record.Category,
		"subcategory":            record.Subcategory,
		"quantity":               record.Quantity,
		"unit":                   record.Unit,
		"minimum_order_quantity": record.MinimumOrderQuantity,
		"price_min":              record.PriceMin,
		"price_max":              record.PriceMax,
		"location":               record.Location,
		"delivery_area":          record.DeliveryArea,
		"availability_status":    record.AvailabilityStatus,
		"available_from":         record.AvailableFrom,
		"available_until":        record.AvailableUntil,
		"created_at":             record.CreatedAt,
	}
}

func demandPostPayload(record posts.DemandPost) map[string]interface{} {
	return map[string]interface{}{
		"entity_id":               record.ID,
		"entity_type":             "demand_post",
		"buyer_id":                record.BuyerID,
		"product_name":            record.ProductName,
		"category":                record.Category,
		"subcategory":             record.Subcategory,
		"quantity":                record.Quantity,
		"unit":                    record.Unit,
		"budget_min":              record.BudgetMin,
		"budget_max":              record.BudgetMax,
		"delivery_location":       record.DeliveryLocation,
		"needed_date":             record.NeededDate,
		"frequency":               record.Frequency,
		"additional_requirements": record.AdditionalRequirements,
		"created_at":              record.CreatedAt,
	}
}

func buildAggregates(user *users.User, searches []SearchHistory, agreements []agreement.Agreement) map[string]interface{} {
	categoryCounts := map[string]int{}
	for _, search := range searches {
		addCount(categoryCounts, search.Category)
		addCount(categoryCounts, search.Query)
	}
	totalQty := 0.0
	totalPrice := 0.0
	itemCount := 0
	completed := 0
	for _, record := range agreements {
		if record.Status == agreement.AgreementStatusConfirmed {
			completed++
		}
		for _, item := range record.Items {
			addCount(categoryCounts, item.ProductName)
			totalQty += item.Quantity
			totalPrice += item.UnitPrice
			itemCount++
		}
	}
	averageQty := 0.0
	averagePrice := 0.0
	if itemCount > 0 {
		averageQty = totalQty / float64(itemCount)
		averagePrice = totalPrice / float64(itemCount)
	}
	return map[string]interface{}{
		"user_id":                   user.ID,
		"category_or_product_count": categoryCounts,
		"agreement_count":           len(agreements),
		"completed_agreement_count": completed,
		"average_quantity":          averageQty,
		"average_unit_price":        averagePrice,
	}
}

func addCount(counts map[string]int, value string) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value != "" {
		counts[value]++
	}
}

func candidateIDSet(candidates []map[string]interface{}) map[string]struct{} {
	ids := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		id, _ := candidate["entity_id"].(string)
		if id != "" {
			ids[id] = struct{}{}
		}
	}
	return ids
}

type matchEvaluation struct {
	PartnerID                string
	EntityID                 string
	EntityType               string
	Score                    int
	Status                   string
	MatchedFields            []string
	MissingOrConflictingReqs []string
	Reasoning                []string
	SuggestedAction          string
	DistanceKM               *float64
	CriticalConstraintFailed bool
	PublicData               map[string]interface{}
}

func (s *Service) evaluateSupplyCandidates(ctx context.Context, demand posts.DemandPost, buyer *users.User, supplies []posts.SupplyPost) ([]matchEvaluation, error) {
	evaluations := make([]matchEvaluation, 0, len(supplies))
	for _, supply := range supplies {
		producer, err := s.repository.FindUserByID(ctx, supply.ProducerID)
		if err != nil {
			if err == ErrUserNotFound {
				producer = &users.User{ID: supply.ProducerID, Role: users.RoleProducer}
			} else {
				return nil, err
			}
		}
		evaluations = append(evaluations, evaluateBuyerProducerMatch(demand, supply, buyer, producer))
	}
	sortEvaluations(evaluations)
	return evaluations, nil
}

func (s *Service) evaluateDemandCandidates(ctx context.Context, supply posts.SupplyPost, producer *users.User, demands []posts.DemandPost) ([]matchEvaluation, error) {
	evaluations := make([]matchEvaluation, 0, len(demands))
	for _, demand := range demands {
		buyer, err := s.repository.FindUserByID(ctx, demand.BuyerID)
		if err != nil {
			if err == ErrUserNotFound {
				buyer = &users.User{ID: demand.BuyerID, Role: users.RoleBuyer}
			} else {
				return nil, err
			}
		}
		evaluation := evaluateBuyerProducerMatch(demand, supply, buyer, producer)
		evaluation.PartnerID = demand.BuyerID
		evaluation.EntityID = demand.ID
		evaluation.EntityType = "demand_post"
		evaluation.PublicData = publicBuyerData(buyer, demand)
		evaluations = append(evaluations, evaluation)
	}
	sortEvaluations(evaluations)
	return evaluations, nil
}

func evaluateBuyerProducerMatch(demand posts.DemandPost, supply posts.SupplyPost, buyer *users.User, producer *users.User) matchEvaluation {
	evaluation := matchEvaluation{
		PartnerID:       supply.ProducerID,
		EntityID:        supply.ID,
		EntityType:      "supply_post",
		PublicData:      publicPartnerData(producer, supply),
		SuggestedAction: "Review the match details before starting mutual interest.",
	}
	addWeightedResult(&evaluation, "product_category", 25, categoryCompatible(demand, supply), true, "Product category is compatible.", "Product category is not compatible.")
	addWeightedResult(&evaluation, "producer_capacity", 15, supply.Quantity >= demand.Quantity, true, "Producer capacity can satisfy buyer quantity.", "Producer capacity is below buyer required quantity.")
	addWeightedResult(&evaluation, "moq", 10, supply.MinimumOrderQuantity <= 0 || supply.MinimumOrderQuantity <= demand.Quantity, true, "Producer MOQ fits buyer quantity.", "Producer MOQ exceeds buyer required quantity.")
	addWeightedResult(&evaluation, "certifications", 15, certificationsCompatible(requiredCertifications(demand, buyer), producerCertifications(producer)), true, "Required certifications are available or not specified.", "Required certifications are missing.")
	addWeightedResult(&evaluation, "delivery_area", 10, deliveryCompatible(supply, demand), true, "Producer can serve buyer delivery area.", "Producer delivery area does not cover buyer location.")
	addWeightedResult(&evaluation, "availability", 10, availabilityCompatible(supply, demand), true, "Producer availability fits buyer timeline.", "Producer availability does not fit buyer timeline.")
	addWeightedResult(&evaluation, "business_type", 5, businessTypeCompatible(buyer, producer), false, "Business types are compatible or unspecified.", "Business type compatibility is unclear.")
	addWeightedResult(&evaluation, "purchase_frequency", 5, purchaseFrequencyCompatible(demand, buyer, producer), false, "Purchase frequency is compatible or unspecified.", "Purchase frequency compatibility is unclear.")

	evaluation.MissingOrConflictingReqs = append(evaluation.MissingOrConflictingReqs, "distance unavailable: no coordinate data exists in current schema")
	evaluation.Status = statusForScore(evaluation.Score)
	if evaluation.CriticalConstraintFailed && evaluation.Status == "excellent_match" {
		evaluation.Status = "partial_match"
		if evaluation.Score > 74 {
			evaluation.Score = 74
		}
	}
	if evaluation.Score >= 75 {
		evaluation.SuggestedAction = "Candidate is strong enough for direct outreach through the agreement flow."
	}
	if evaluation.Score < 50 {
		evaluation.SuggestedAction = "Do not prioritize this partner unless requirements change."
	}
	return evaluation
}

func addWeightedResult(evaluation *matchEvaluation, field string, weight int, passed bool, critical bool, passReason string, failReason string) {
	if passed {
		evaluation.Score += weight
		evaluation.MatchedFields = append(evaluation.MatchedFields, field)
		evaluation.Reasoning = append(evaluation.Reasoning, passReason)
		return
	}
	evaluation.MissingOrConflictingReqs = append(evaluation.MissingOrConflictingReqs, field)
	evaluation.Reasoning = append(evaluation.Reasoning, failReason)
	if critical {
		evaluation.CriticalConstraintFailed = true
	}
}

func categoryCompatible(demand posts.DemandPost, supply posts.SupplyPost) bool {
	return equalText(demand.Category, supply.Category) ||
		equalText(demand.Subcategory, supply.Subcategory) ||
		equalText(demand.ProductName, supply.ProductName) ||
		containsText(supply.ProductName, demand.ProductName) ||
		containsText(demand.ProductName, supply.ProductName)
}

func requiredCertifications(demand posts.DemandPost, buyer *users.User) []string {
	values := splitCSV(demand.AdditionalRequirements)
	if buyer != nil && buyer.Profile != nil {
		values = append(values, splitCSV(buyer.Profile.Certifications)...)
	}
	return normalizeTerms(values)
}

func producerCertifications(producer *users.User) []string {
	if producer == nil || producer.Profile == nil {
		return []string{}
	}
	return normalizeTerms(splitCSV(producer.Profile.Certifications))
}

func certificationsCompatible(required []string, available []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(available) == 0 {
		return false
	}
	for _, req := range required {
		found := false
		for _, cert := range available {
			if containsText(cert, req) || containsText(req, cert) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func deliveryCompatible(supply posts.SupplyPost, demand posts.DemandPost) bool {
	return containsText(supply.DeliveryArea, demand.DeliveryLocation) ||
		containsText(demand.DeliveryLocation, supply.Location) ||
		equalText(supply.Location, demand.DeliveryLocation)
}

func availabilityCompatible(supply posts.SupplyPost, demand posts.DemandPost) bool {
	if demand.NeededDate == nil {
		return true
	}
	if supply.AvailableFrom != nil && supply.AvailableFrom.After(*demand.NeededDate) {
		return false
	}
	if supply.AvailableUntil != nil && supply.AvailableUntil.Before(*demand.NeededDate) {
		return false
	}
	if strings.TrimSpace(supply.AvailabilityStatus) != "" && !strings.EqualFold(supply.AvailabilityStatus, "available") {
		return false
	}
	return true
}

func businessTypeCompatible(buyer *users.User, producer *users.User) bool {
	if buyer == nil || producer == nil || buyer.Profile == nil || producer.Profile == nil {
		return true
	}
	buyerType := strings.TrimSpace(buyer.Profile.BusinessType)
	producerType := strings.TrimSpace(producer.Profile.BusinessType)
	return buyerType == "" || producerType == "" || !strings.EqualFold(buyerType, producerType)
}

func purchaseFrequencyCompatible(demand posts.DemandPost, buyer *users.User, producer *users.User) bool {
	frequency := strings.TrimSpace(demand.Frequency)
	if frequency == "" {
		return true
	}
	if buyer != nil && buyer.Profile != nil && equalText(frequency, buyer.Profile.PurchaseFrequency) {
		return true
	}
	if producer != nil && producer.Profile != nil && strings.TrimSpace(producer.Profile.Availability) != "" {
		return true
	}
	return false
}

func publicPartnerData(producer *users.User, supply posts.SupplyPost) map[string]interface{} {
	data := map[string]interface{}{
		"post_id":      supply.ID,
		"product_name": supply.ProductName,
		"category":     supply.Category,
		"location":     supply.Location,
	}
	if producer != nil && producer.Profile != nil {
		data["company_name"] = producer.Profile.CompanyName
		data["city"] = producer.Profile.City
		data["business_type"] = producer.Profile.BusinessType
		data["product_category"] = producer.Profile.ProductCategory
		data["certifications"] = producer.Profile.Certifications
	}
	return data
}

func publicBuyerData(buyer *users.User, demand posts.DemandPost) map[string]interface{} {
	data := map[string]interface{}{
		"post_id":            demand.ID,
		"product_name":       demand.ProductName,
		"category":           demand.Category,
		"delivery_location":  demand.DeliveryLocation,
		"purchase_frequency": demand.Frequency,
	}
	if buyer != nil && buyer.Profile != nil {
		data["company_name"] = buyer.Profile.CompanyName
		data["city"] = buyer.Profile.City
		data["business_type"] = buyer.Profile.BusinessType
		data["product_category"] = buyer.Profile.ProductCategory
	}
	return data
}

func statusForScore(score int) string {
	switch {
	case score >= 90:
		return "excellent_match"
	case score >= 75:
		return "good_match"
	case score >= 50:
		return "partial_match"
	default:
		return "poor_match"
	}
}

func normalizeTerms(values []string) []string {
	normalized := []string{}
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value != "" {
			normalized = appendUnique(normalized, value)
		}
	}
	return normalized
}

func equalText(a string, b string) bool {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	return a != "" && b != "" && strings.EqualFold(a, b)
}

func containsText(haystack string, needle string) bool {
	haystack = strings.ToLower(strings.TrimSpace(haystack))
	needle = strings.ToLower(strings.TrimSpace(needle))
	return haystack != "" && needle != "" && strings.Contains(haystack, needle)
}

func matchEvaluationPayloads(evaluations []matchEvaluation) []map[string]interface{} {
	payload := make([]map[string]interface{}, 0, len(evaluations))
	for _, evaluation := range evaluations {
		payload = append(payload, map[string]interface{}{
			"partner_id":                          evaluation.PartnerID,
			"entity_id":                           evaluation.EntityID,
			"entity_type":                         evaluation.EntityType,
			"base_compatibility_score":            evaluation.Score,
			"match_status":                        evaluation.Status,
			"matched_fields":                      evaluation.MatchedFields,
			"missing_or_conflicting_requirements": evaluation.MissingOrConflictingReqs,
			"distance_km":                         evaluation.DistanceKM,
			"critical_constraint_failed":          evaluation.CriticalConstraintFailed,
			"public_data":                         evaluation.PublicData,
		})
	}
	return payload
}

func evaluationIDSet(evaluations []matchEvaluation) map[string]struct{} {
	ids := make(map[string]struct{}, len(evaluations))
	for _, evaluation := range evaluations {
		ids[evaluation.PartnerID] = struct{}{}
	}
	return ids
}

func evaluationScoreByID(evaluations []matchEvaluation) map[string]int {
	scores := make(map[string]int, len(evaluations))
	for _, evaluation := range evaluations {
		scores[evaluation.PartnerID] = evaluation.Score
	}
	return scores
}

func evaluationByID(evaluations []matchEvaluation) map[string]matchEvaluation {
	records := make(map[string]matchEvaluation, len(evaluations))
	for _, evaluation := range evaluations {
		records[evaluation.PartnerID] = evaluation
	}
	return records
}

func sortEvaluations(evaluations []matchEvaluation) {
	sort.SliceStable(evaluations, func(i, j int) bool {
		return evaluations[i].Score > evaluations[j].Score
	})
}

func fallbackRecommendation(user *users.User, searches []SearchHistory, agreements []agreement.Agreement, candidates []map[string]interface{}) RecommendationResponse {
	summary := ProfileSummary{}
	if user.Profile != nil {
		summary.PreferredProductCategories = splitCSV(user.Profile.ProductCategory)
		summary.PreferredCertifications = splitCSV(user.Profile.Certifications)
		summary.PreferredDeliveryAreas = splitCSV(user.Profile.DeliveryArea)
		if user.Profile.PurchaseFrequency != "" {
			frequency := user.Profile.PurchaseFrequency
			summary.PurchaseFrequency = &frequency
		}
	}
	for _, search := range searches {
		if search.Query != "" {
			summary.RecentInterests = appendUnique(summary.RecentInterests, search.Query)
		}
		if search.Category != "" {
			summary.PreferredProductCategories = appendUnique(summary.PreferredProductCategories, search.Category)
		}
		if search.Location != "" {
			summary.PreferredDeliveryAreas = appendUnique(summary.PreferredDeliveryAreas, search.Location)
		}
	}
	for _, record := range agreements {
		for _, item := range record.Items {
			summary.RecentInterests = appendUnique(summary.RecentInterests, item.ProductName)
			summary.PreferredCapacityOrQty = widenRange(summary.PreferredCapacityOrQty, item.Quantity)
		}
	}

	recommendations := make([]RecommendationItem, 0, len(candidates))
	for i, candidate := range candidates {
		if i >= 10 {
			break
		}
		id, _ := candidate["entity_id"].(string)
		entityType, _ := candidate["entity_type"].(string)
		if id == "" {
			continue
		}
		score := 60
		reasons := []string{"Candidate is available in the eligible backend list."}
		category, _ := candidate["category"].(string)
		if containsFold(summary.PreferredProductCategories, category) {
			score += 20
			reasons = append(reasons, "Category matches the user's observed preference.")
		}
		recommendations = append(recommendations, RecommendationItem{
			EntityID:           id,
			EntityType:         entityType,
			CompatibilityScore: clampScore(score),
			Reasoning:          reasons,
			EvidenceSource:     []string{"search_history", "agreement_history"},
		})
	}
	return RecommendationResponse{ProfileSummary: summary, Recommendations: recommendations}
}

func fallbackMatchmaking(sourceID string, sourceType string, evaluations []matchEvaluation) MatchmakingResponse {
	matches := make([]MatchCandidate, 0, len(evaluations))
	for i, evaluation := range evaluations {
		if i >= 10 {
			break
		}
		if evaluation.PartnerID == "" {
			continue
		}
		matches = append(matches, MatchCandidate{
			PartnerID:                evaluation.PartnerID,
			EntityID:                 evaluation.EntityID,
			EntityType:               evaluation.EntityType,
			CompatibilityScore:       evaluation.Score,
			MatchStatus:              evaluation.Status,
			MatchedFields:            evaluation.MatchedFields,
			MissingOrConflictingReqs: evaluation.MissingOrConflictingReqs,
			Reasoning:                evaluation.Reasoning,
			SuggestedAction:          evaluation.SuggestedAction,
			PublicData:               evaluation.PublicData,
		})
	}
	response := MatchmakingResponse{
		SourcePostID:      sourceID,
		SourceType:        sourceType,
		SuggestedPartners: matches,
	}
	if len(matches) > 0 {
		top := matches[0]
		response.CompatibilityScore = top.CompatibilityScore
		response.MatchStatus = top.MatchStatus
		response.MatchedFields = top.MatchedFields
		response.MissingOrConflictingRequirements = top.MissingOrConflictingReqs
		response.Reasoning = top.Reasoning
		response.SuggestedAction = top.SuggestedAction
	}
	return response
}

func fallbackAgreementVerification(record *agreement.Agreement, messages []chat.Message) AgreementVerificationResponse {
	canReveal := record.Status == agreement.AgreementStatusConfirmed && record.BuyerConfirmed && record.ProducerConfirmed
	status := "needs_confirmation"
	if canReveal {
		status = "verified"
	}
	normalized := submissionFromAgreement(record)
	fieldComparison := map[string]AgreementFieldResult{
		"buyer_company":      {Status: "needs_review", Reason: "Buyer company is not stored on agreement items."},
		"producer_company":   {Status: "needs_review", Reason: "Producer company is not stored on agreement items."},
		"product":            {Status: statusFromBool(len(record.Items) > 0), Reason: "Product is sourced from stored agreement items."},
		"quantity_and_unit":  {Status: statusFromBool(len(record.Items) > 0), Reason: "Quantity and unit are sourced from stored agreement items."},
		"price_and_currency": {Status: statusFromBool(len(record.Items) > 0), Reason: "Price and currency are sourced from stored agreement items."},
		"delivery_terms":     {Status: statusFromBool(len(record.Items) > 0), Reason: "Delivery terms are sourced from stored agreement items."},
		"payment_terms":      {Status: statusFromBool(len(record.Items) > 0), Reason: "Payment terms are sourced from stored agreement items."},
		"additional_terms":   {Status: "needs_review", Reason: "Additional terms should be compared against producer submission or negotiation summary."},
	}
	terms := make([]string, 0, len(record.Items))
	for _, item := range record.Items {
		terms = append(terms, fmt.Sprintf("%s %.2f %s at %.2f %s", item.ProductName, item.Quantity, item.Unit, item.UnitPrice, item.Currency))
	}
	highlights := []string{}
	if len(messages) > 0 {
		highlights = append(highlights, fmt.Sprintf("%d negotiation messages reviewed.", len(messages)))
	}
	return AgreementVerificationResponse{
		AgreementID:           record.ID,
		MatchID:               record.MatchID,
		CanRevealContact:      canReveal,
		RecommendedApproval:   canReveal,
		OverallStatus:         statusFromBool(canReveal),
		VerificationStatus:    status,
		FieldComparison:       fieldComparison,
		Conflicts:             []string{},
		MissingInformation:    missingTerms(record),
		NormalizedAgreement:   normalized,
		Summary:               "Agreement terms were reviewed against confirmation status and recorded line items.",
		ConfirmedTerms:        terms,
		MissingTerms:          missingTerms(record),
		NegotiationHighlights: highlights,
		Risks:                 []string{},
		RecommendedNextSteps:  nextSteps(canReveal),
	}
}

func compareAgreementSubmissions(buyer AgreementSubmission, producer AgreementSubmission) AgreementVerificationResponse {
	buyer = normalizeSubmission(buyer)
	producer = normalizeSubmission(producer)
	fields := map[string]AgreementFieldResult{}
	conflicts := []string{}
	missing := []string{}

	compareField(fields, &conflicts, &missing, "buyer_company", companyCompatible(buyer.BuyerCompany, producer.BuyerCompany), buyer.BuyerCompany, producer.BuyerCompany)
	compareField(fields, &conflicts, &missing, "producer_company", companyCompatible(buyer.ProducerCompany, producer.ProducerCompany), buyer.ProducerCompany, producer.ProducerCompany)
	compareField(fields, &conflicts, &missing, "product", productCompatible(buyer.Product, producer.Product), buyer.Product, producer.Product)
	compareField(fields, &conflicts, &missing, "quantity_and_unit", quantityUnitCompatible(buyer, producer), quantityUnitString(buyer), quantityUnitString(producer))
	compareField(fields, &conflicts, &missing, "price_and_currency", priceCompatible(buyer, producer), priceString(buyer), priceString(producer))
	compareField(fields, &conflicts, &missing, "delivery_terms", textCompatible(buyer.DeliveryArea, producer.DeliveryArea) && textCompatible(buyer.DeliverySchedule, producer.DeliverySchedule), buyer.DeliveryArea+" "+buyer.DeliverySchedule, producer.DeliveryArea+" "+producer.DeliverySchedule)
	compareField(fields, &conflicts, &missing, "payment_terms", textCompatible(buyer.PaymentTerms, producer.PaymentTerms), buyer.PaymentTerms, producer.PaymentTerms)
	compareField(fields, &conflicts, &missing, "additional_terms", termsCompatible(buyer.AdditionalTerms, producer.AdditionalTerms), strings.Join(buyer.AdditionalTerms, "; "), strings.Join(producer.AdditionalTerms, "; "))

	overall := "match"
	if len(missing) > 0 {
		overall = "needs_review"
	}
	if len(conflicts) > 0 {
		overall = "mismatch"
	}
	recommended := overall == "match"
	nextAction := "Both parties can final-confirm the verified agreement."
	if !recommended {
		nextAction = "Resolve conflicts or missing information before final confirmation."
	}

	return AgreementVerificationResponse{
		RecommendedApproval: recommended,
		OverallStatus:       overall,
		FieldComparison:     fields,
		Conflicts:           conflicts,
		MissingInformation:  missing,
		NormalizedAgreement: mergeSubmission(buyer, producer),
		Summary:             "Buyer and producer submissions were compared across critical agreement fields.",
		NextAction:          nextAction,
	}
}

func compareField(fields map[string]AgreementFieldResult, conflicts *[]string, missing *[]string, name string, matched bool, buyerValue string, producerValue string) {
	buyerValue = strings.TrimSpace(buyerValue)
	producerValue = strings.TrimSpace(producerValue)
	if buyerValue == "" || producerValue == "" {
		fields[name] = AgreementFieldResult{Status: "missing_information", Reason: "Value is missing from one or both submissions."}
		*missing = appendUnique(*missing, name)
		return
	}
	if matched {
		fields[name] = AgreementFieldResult{Status: "match", Reason: "Values are compatible after backend normalization."}
		return
	}
	fields[name] = AgreementFieldResult{Status: "mismatch", Reason: "Values differ after backend normalization."}
	*conflicts = appendUnique(*conflicts, name)
}

func normalizeSubmission(submission AgreementSubmission) AgreementSubmission {
	submission.BuyerCompany = strings.TrimSpace(submission.BuyerCompany)
	submission.ProducerCompany = strings.TrimSpace(submission.ProducerCompany)
	submission.Product = strings.TrimSpace(submission.Product)
	submission.Unit = strings.ToLower(strings.TrimSpace(submission.Unit))
	submission.Currency = strings.ToUpper(strings.TrimSpace(submission.Currency))
	submission.DeliveryArea = strings.TrimSpace(submission.DeliveryArea)
	submission.DeliverySchedule = strings.TrimSpace(submission.DeliverySchedule)
	submission.PaymentTerms = strings.TrimSpace(submission.PaymentTerms)
	submission.Certifications = normalizeTerms(submission.Certifications)
	submission.QualityRequirements = normalizeTerms(submission.QualityRequirements)
	submission.AdditionalTerms = normalizeTerms(submission.AdditionalTerms)
	return submission
}

func mergeSubmission(buyer AgreementSubmission, producer AgreementSubmission) AgreementSubmission {
	return AgreementSubmission{
		BuyerCompany:        firstNonEmptyText(buyer.BuyerCompany, producer.BuyerCompany),
		ProducerCompany:     firstNonEmptyText(buyer.ProducerCompany, producer.ProducerCompany),
		Product:             firstNonEmptyText(buyer.Product, producer.Product),
		Quantity:            firstFloat(buyer.Quantity, producer.Quantity),
		Unit:                firstNonEmptyText(buyer.Unit, producer.Unit),
		AgreedUnitPrice:     firstFloat(buyer.AgreedUnitPrice, producer.AgreedUnitPrice),
		AgreedTotalPrice:    firstFloat(buyer.AgreedTotalPrice, producer.AgreedTotalPrice),
		Currency:            firstNonEmptyText(buyer.Currency, producer.Currency),
		DeliveryArea:        firstNonEmptyText(buyer.DeliveryArea, producer.DeliveryArea),
		DeliverySchedule:    firstNonEmptyText(buyer.DeliverySchedule, producer.DeliverySchedule),
		PaymentTerms:        firstNonEmptyText(buyer.PaymentTerms, producer.PaymentTerms),
		Certifications:      appendUniqueList(buyer.Certifications, producer.Certifications),
		QualityRequirements: appendUniqueList(buyer.QualityRequirements, producer.QualityRequirements),
		AdditionalTerms:     appendUniqueList(buyer.AdditionalTerms, producer.AdditionalTerms),
	}
}

func submissionFromAgreement(record *agreement.Agreement) AgreementSubmission {
	if len(record.Items) == 0 {
		return AgreementSubmission{}
	}
	item := record.Items[0]
	total := item.Quantity * item.UnitPrice
	return AgreementSubmission{
		Product:          item.ProductName,
		Quantity:         &item.Quantity,
		Unit:             item.Unit,
		AgreedUnitPrice:  &item.UnitPrice,
		AgreedTotalPrice: &total,
		Currency:         item.Currency,
		DeliveryArea:     item.DeliveryAddress,
		DeliverySchedule: item.DeliveryDate.Format("2006-01-02"),
		PaymentTerms:     item.PaymentTerms,
		AdditionalTerms:  splitCSV(item.AdditionalNotes),
	}
}

func canRevealAfterVerification(record *agreement.Agreement, req AgreementVerificationRequest, deterministic AgreementVerificationResponse, aiResponse AgreementVerificationResponse) bool {
	return record.Status == agreement.AgreementStatusConfirmed &&
		record.BuyerConfirmed &&
		record.ProducerConfirmed &&
		req.BuyerFinalConfirm &&
		req.ProducerFinalConfirm &&
		deterministic.OverallStatus == "match" &&
		deterministic.RecommendedApproval &&
		aiResponse.RecommendedApproval &&
		aiResponse.OverallStatus == "match"
}

func sanitizeSubmissionVerification(aiResponse AgreementVerificationResponse, deterministic AgreementVerificationResponse) AgreementVerificationResponse {
	if aiResponse.OverallStatus == "" {
		aiResponse.OverallStatus = deterministic.OverallStatus
	}
	if deterministic.OverallStatus != "match" {
		aiResponse.RecommendedApproval = false
		aiResponse.OverallStatus = deterministic.OverallStatus
	}
	if len(aiResponse.FieldComparison) == 0 {
		aiResponse.FieldComparison = deterministic.FieldComparison
	}
	if len(deterministic.Conflicts) > 0 {
		aiResponse.Conflicts = appendUniqueList(aiResponse.Conflicts, deterministic.Conflicts)
	}
	if len(deterministic.MissingInformation) > 0 {
		aiResponse.MissingInformation = appendUniqueList(aiResponse.MissingInformation, deterministic.MissingInformation)
	}
	if aiResponse.NormalizedAgreement.Product == "" {
		aiResponse.NormalizedAgreement = deterministic.NormalizedAgreement
	}
	if aiResponse.Summary == "" {
		aiResponse.Summary = deterministic.Summary
	}
	if aiResponse.NextAction == "" {
		aiResponse.NextAction = deterministic.NextAction
	}
	return aiResponse
}

func fallbackNegotiationSummary(record *agreement.Agreement, messages []chat.Message) NegotiationSummaryResponse {
	draft := submissionFromAgreement(record)
	summary := "No negotiation messages were available for summarization."
	if len(messages) > 0 {
		summary = fmt.Sprintf("%d negotiation messages are available for AI summarization.", len(messages))
	}
	agreed := []string{}
	if draft.Product != "" {
		agreed = append(agreed, "Stored agreement item: "+draft.Product)
	}
	return NegotiationSummaryResponse{
		AgreementDraft:  draft,
		AgreedTerms:     agreed,
		UnresolvedTerms: []string{},
		Evidence:        []NegotiationEvidence{},
		Summary:         summary,
	}
}

func sanitizeNegotiationSummary(aiResponse NegotiationSummaryResponse, fallback NegotiationSummaryResponse) NegotiationSummaryResponse {
	if aiResponse.Summary == "" {
		aiResponse.Summary = fallback.Summary
	}
	if aiResponse.AgreementDraft.Product == "" && fallback.AgreementDraft.Product != "" {
		aiResponse.AgreementDraft = fallback.AgreementDraft
	}
	if aiResponse.AgreedTerms == nil {
		aiResponse.AgreedTerms = []string{}
	}
	if aiResponse.UnresolvedTerms == nil {
		aiResponse.UnresolvedTerms = []string{}
	}
	if aiResponse.Evidence == nil {
		aiResponse.Evidence = []NegotiationEvidence{}
	}
	return aiResponse
}

func companyCompatible(a string, b string) bool {
	return normalizeCompany(a) == normalizeCompany(b)
}

func normalizeCompany(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(".", "", ",", "", "-", " ", "_", " ")
	value = replacer.Replace(value)
	parts := strings.Fields(value)
	filtered := []string{}
	for _, part := range parts {
		if part == "pt" || part == "cv" || part == "tbk" {
			continue
		}
		filtered = append(filtered, part)
	}
	return strings.Join(filtered, " ")
}

func productCompatible(a string, b string) bool {
	return equalText(a, b)
}

func quantityUnitCompatible(a AgreementSubmission, b AgreementSubmission) bool {
	if a.Quantity == nil || b.Quantity == nil || !equalText(a.Unit, b.Unit) {
		return false
	}
	return almostEqual(*a.Quantity, *b.Quantity)
}

func priceCompatible(a AgreementSubmission, b AgreementSubmission) bool {
	if !equalText(a.Currency, b.Currency) {
		return false
	}
	aUnit := normalizedUnitPrice(a)
	bUnit := normalizedUnitPrice(b)
	if aUnit == nil || bUnit == nil {
		return false
	}
	return almostEqual(*aUnit, *bUnit)
}

func normalizedUnitPrice(submission AgreementSubmission) *float64 {
	if submission.AgreedUnitPrice != nil {
		return submission.AgreedUnitPrice
	}
	if submission.AgreedTotalPrice != nil && submission.Quantity != nil && *submission.Quantity > 0 {
		value := *submission.AgreedTotalPrice / *submission.Quantity
		return &value
	}
	return nil
}

func textCompatible(a string, b string) bool {
	return equalText(a, b) || containsText(a, b) || containsText(b, a)
}

func termsCompatible(a []string, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for _, term := range a {
		found := false
		for _, other := range b {
			if textCompatible(term, other) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func quantityUnitString(submission AgreementSubmission) string {
	if submission.Quantity == nil {
		return ""
	}
	return fmt.Sprintf("%.4f %s", *submission.Quantity, submission.Unit)
}

func priceString(submission AgreementSubmission) string {
	unit := normalizedUnitPrice(submission)
	if unit == nil {
		return ""
	}
	return fmt.Sprintf("%.4f %s", *unit, submission.Currency)
}

func statusFromBool(value bool) string {
	if value {
		return "match"
	}
	return "needs_review"
}

func almostEqual(a float64, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= 0.01
}

func firstNonEmptyText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstFloat(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func appendUniqueList(base []string, extra []string) []string {
	result := append([]string{}, base...)
	for _, value := range extra {
		result = appendUnique(result, value)
	}
	return result
}

func sanitizeRecommendation(aiResponse RecommendationResponse, fallback RecommendationResponse, allowed map[string]struct{}) RecommendationResponse {
	if len(aiResponse.ProfileSummary.PreferredProductCategories) == 0 {
		aiResponse.ProfileSummary = fallback.ProfileSummary
	}
	cleaned := make([]RecommendationItem, 0, len(aiResponse.Recommendations))
	seen := map[string]struct{}{}
	for _, item := range aiResponse.Recommendations {
		if _, ok := allowed[item.EntityID]; !ok {
			continue
		}
		if _, ok := seen[item.EntityID]; ok {
			continue
		}
		seen[item.EntityID] = struct{}{}
		item.CompatibilityScore = clampScore(item.CompatibilityScore)
		if item.EntityType == "" {
			item.EntityType = "product_or_partner"
		}
		cleaned = append(cleaned, item)
	}
	if len(cleaned) == 0 {
		aiResponse.Recommendations = fallback.Recommendations
	} else {
		aiResponse.Recommendations = cleaned
	}
	return aiResponse
}

func sanitizeMatchmaking(aiResponse MatchmakingResponse, fallback MatchmakingResponse, sourceID string, sourceType string, allowed map[string]struct{}, maxScores map[string]int, evaluations map[string]matchEvaluation) MatchmakingResponse {
	aiResponse.SourcePostID = sourceID
	aiResponse.SourceType = sourceType
	cleaned := make([]MatchCandidate, 0, len(aiResponse.SuggestedPartners))
	seen := map[string]struct{}{}
	for _, item := range aiResponse.SuggestedPartners {
		if item.PartnerID == "" {
			item.PartnerID = item.EntityID
		}
		if _, ok := allowed[item.PartnerID]; !ok {
			continue
		}
		if _, ok := seen[item.PartnerID]; ok {
			continue
		}
		seen[item.PartnerID] = struct{}{}
		maxScore := maxScores[item.PartnerID]
		if item.CompatibilityScore <= 0 || item.CompatibilityScore > maxScore {
			item.CompatibilityScore = maxScore
		}
		item.CompatibilityScore = clampScore(item.CompatibilityScore)
		item.MatchStatus = statusForScore(item.CompatibilityScore)
		if evaluation, ok := evaluations[item.PartnerID]; ok {
			item.EntityID = evaluation.EntityID
			item.EntityType = evaluation.EntityType
			item.MatchedFields = evaluation.MatchedFields
			item.MissingOrConflictingReqs = evaluation.MissingOrConflictingReqs
			item.SuggestedAction = evaluation.SuggestedAction
			item.PublicData = evaluation.PublicData
			if len(item.Reasoning) == 0 {
				item.Reasoning = evaluation.Reasoning
			}
		}
		cleaned = append(cleaned, item)
	}
	if len(cleaned) == 0 {
		return fallback
	}
	aiResponse.SuggestedPartners = cleaned
	sort.SliceStable(aiResponse.SuggestedPartners, func(i, j int) bool {
		return aiResponse.SuggestedPartners[i].CompatibilityScore > aiResponse.SuggestedPartners[j].CompatibilityScore
	})
	top := aiResponse.SuggestedPartners[0]
	if aiResponse.CompatibilityScore <= 0 || aiResponse.CompatibilityScore > top.CompatibilityScore {
		aiResponse.CompatibilityScore = top.CompatibilityScore
	}
	aiResponse.MatchStatus = statusForScore(aiResponse.CompatibilityScore)
	if len(aiResponse.MatchedFields) == 0 {
		aiResponse.MatchedFields = top.MatchedFields
	}
	if len(aiResponse.MissingOrConflictingRequirements) == 0 {
		aiResponse.MissingOrConflictingRequirements = top.MissingOrConflictingReqs
	}
	if len(aiResponse.Reasoning) == 0 {
		aiResponse.Reasoning = top.Reasoning
	}
	if aiResponse.SuggestedAction == "" {
		aiResponse.SuggestedAction = top.SuggestedAction
	}
	return aiResponse
}

func sanitizeAgreementVerification(aiResponse AgreementVerificationResponse, fallback AgreementVerificationResponse, record *agreement.Agreement) AgreementVerificationResponse {
	canReveal := record.Status == agreement.AgreementStatusConfirmed && record.BuyerConfirmed && record.ProducerConfirmed
	aiResponse.AgreementID = record.ID
	aiResponse.MatchID = record.MatchID
	aiResponse.CanRevealContact = canReveal
	if aiResponse.VerificationStatus == "" {
		aiResponse.VerificationStatus = fallback.VerificationStatus
	}
	if aiResponse.Summary == "" {
		aiResponse.Summary = fallback.Summary
	}
	if !canReveal && len(aiResponse.RecommendedNextSteps) == 0 {
		aiResponse.RecommendedNextSteps = fallback.RecommendedNextSteps
	}
	return aiResponse
}

func splitCSV(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
	result := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = appendUnique(result, part)
		}
	}
	return result
}

func appendUnique(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return values
		}
	}
	return append(values, value)
}

func containsFold(values []string, value string) bool {
	for _, existing := range values {
		if strings.EqualFold(existing, value) {
			return true
		}
	}
	return false
}

func widenRange(summary RangeSummary, value float64) RangeSummary {
	if value <= 0 {
		return summary
	}
	if summary.Min == nil || value < *summary.Min {
		v := value
		summary.Min = &v
	}
	if summary.Max == nil || value > *summary.Max {
		v := value
		summary.Max = &v
	}
	return summary
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func missingTerms(record *agreement.Agreement) []string {
	missing := []string{}
	if len(record.Items) == 0 {
		return []string{"agreement items"}
	}
	for _, item := range record.Items {
		if strings.TrimSpace(item.ProductName) == "" {
			missing = appendUnique(missing, "product name")
		}
		if item.Quantity <= 0 {
			missing = appendUnique(missing, "quantity")
		}
		if item.UnitPrice <= 0 {
			missing = appendUnique(missing, "unit price")
		}
		if strings.TrimSpace(item.DeliveryAddress) == "" {
			missing = appendUnique(missing, "delivery address")
		}
		if strings.TrimSpace(item.PaymentTerms) == "" {
			missing = appendUnique(missing, "payment terms")
		}
	}
	return missing
}

func nextSteps(canReveal bool) []string {
	if canReveal {
		return []string{"Contact reveal is available for both confirmed parties."}
	}
	return []string{"Ask both parties to confirm the agreement before revealing contact information."}
}

const recommendationPrompt = `You are the personalized recommendation engine for Foodlink, a platform connecting food producers and business buyers.

Analyze the supplied user profile, search history, agreement history, backend aggregates, and eligible candidate list.

Important rules:
1. Search history represents interest only. It does not prove that a transaction occurred.
2. Agreement history represents stronger evidence of actual business needs and preferences.
3. Prioritize confirmed or completed agreements over search history.
4. Recent activity may indicate new interests, but repeated confirmed agreements are more reliable.
5. Never invent products, quantities, locations, certifications, prices, partners, or preferences.
6. Only recommend candidates supplied in the eligible candidate list.
7. Do not expose private contact information.
8. Return valid JSON only.

Return exactly this shape:
{
  "profile_summary": {
    "preferred_product_categories": [],
    "recent_interests": [],
    "preferred_capacity_or_quantity": {"min": null, "max": null},
    "preferred_moq": {"min": null, "max": null},
    "preferred_certifications": [],
    "preferred_delivery_areas": [],
    "purchase_frequency": null
  },
  "recommendations": [
    {
      "entity_id": "",
      "entity_type": "product_or_partner",
      "compatibility_score": 0,
      "reasoning": [],
      "evidence_source": ["search_history", "agreement_history"]
    }
  ]
}`

const matchmakingPrompt = `You are the buyer-producer matching explanation engine for Foodlink.

The backend has already calculated deterministic compatibility values. Do not replace or contradict valid numeric calculations supplied by the backend.

Evaluate whether the buyer and producer are suitable business partners based on:
- Product category
- Required quantity
- Producer capacity
- MOQ
- Certifications
- Delivery area
- Availability
- Business type
- Purchase frequency
- Distance

Rules:
1. Never invent missing information.
2. Do not calculate geographic distance yourself. Use the supplied distance value.
3. Product mismatch, insufficient capacity, unsupported delivery area, or missing mandatory certification are critical issues.
4. Producer MOQ must not exceed buyer required quantity.
5. Shorter distance is preferred, but distance must not override critical business requirements.
6. Only return partner IDs supplied by the backend.
7. Do not expose private addresses, phone numbers, emails, or other private profile data.
8. Return JSON only.
9. Never raise compatibility_score above base_compatibility_score supplied by the backend.

Return exactly this shape:
{
  "compatibility_score": 0,
  "match_status": "excellent_match | good_match | partial_match | poor_match",
  "matched_fields": [],
  "missing_or_conflicting_requirements": [],
  "reasoning": [],
  "suggested_action": "",
  "suggested_partners": [
    {
      "partner_id": "",
      "compatibility_score": 0,
      "reasoning": []
    }
  ]
}`

const agreementVerificationPrompt = `You are the agreement verification engine for Foodlink.

Compare the buyer submission and producer submission.

The goal is to determine whether both parties agreed to the same business terms.

Compare semantic meaning, not exact wording.

Ignore harmless differences such as:
- Capitalization
- Punctuation
- Minor grammar differences
- Word order
- Common equivalent expressions

Do not ignore meaningful differences involving:
- Product type, grade, variety, size, quality, or condition
- Quantity or unit
- Price, currency, tax inclusion, or price basis
- Delivery date, location, schedule, or responsibility
- Payment timing or payment method
- Cancellation rules
- Quality requirements
- Certification requirements
- Penalties, returns, refunds, or dispute terms

Rules:
1. Never invent missing terms.
2. If a critical value is missing from either side, mark it as missing_information.
3. If two terms may be compatible but are ambiguous, mark them as needs_review.
4. Do not approve based only on similar wording.
5. The agreement can only be approved when all critical fields match and additional terms are logically compatible.
6. Do not expose private data in the reasoning.
7. Return valid JSON only.
8. AI only recommends verification; backend controls profile unlock.

Return exactly this shape:
{
  "recommended_approval": false,
  "overall_status": "match | partial_match | mismatch | needs_review",
  "field_comparison": {
    "buyer_company": {"status": "match | mismatch | missing_information | needs_review", "reason": ""},
    "producer_company": {"status": "match | mismatch | missing_information | needs_review", "reason": ""},
    "product": {"status": "", "reason": ""},
    "quantity_and_unit": {"status": "", "reason": ""},
    "price_and_currency": {"status": "", "reason": ""},
    "delivery_terms": {"status": "", "reason": ""},
    "payment_terms": {"status": "", "reason": ""},
    "additional_terms": {"status": "", "reason": ""}
  },
  "conflicts": [],
  "missing_information": [],
  "normalized_agreement": {
    "buyer_company": "",
    "producer_company": "",
    "product": "",
    "quantity": null,
    "unit": "",
    "agreed_unit_price": null,
    "agreed_total_price": null,
    "currency": "",
    "delivery_area": "",
    "delivery_schedule": "",
    "payment_terms": "",
    "additional_terms": []
  },
  "summary": "",
  "next_action": ""
}`

const negotiationSummaryPrompt = `You are the negotiation summarization engine for Foodlink.

Summarize the supplied buyer-producer negotiation into a structured agreement draft.

Rules:
1. Use only information explicitly stated in the negotiation.
2. Use the latest mutually accepted value when a term was revised.
3. A proposal from only one party is not automatically an agreed term.
4. Only mark a term as agreed when the conversation provides evidence that both parties accepted it.
5. Put unclear or one-sided proposals under unresolved_terms.
6. Do not invent missing information.
7. Preserve important conditions such as product grade, delivery schedule, payment deadline, certification, return policy, and penalties.
8. Return JSON only.

Return exactly this shape:
{
  "agreement_draft": {
    "buyer_company": "",
    "producer_company": "",
    "product": "",
    "quantity": null,
    "unit": "",
    "agreed_unit_price": null,
    "agreed_total_price": null,
    "currency": "",
    "delivery_area": "",
    "delivery_schedule": "",
    "payment_terms": "",
    "certifications": [],
    "quality_requirements": [],
    "additional_terms": []
  },
  "agreed_terms": [],
  "unresolved_terms": [],
  "evidence": [
    {
      "field": "",
      "buyer_evidence": "",
      "producer_evidence": ""
    }
  ],
  "summary": ""
}`

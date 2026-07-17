package ai

import "time"

// JSONResponse matches the response envelope used by existing modules.
type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type RangeSummary struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
}

type ProfileSummary struct {
	PreferredProductCategories []string     `json:"preferred_product_categories"`
	RecentInterests            []string     `json:"recent_interests"`
	PreferredCapacityOrQty     RangeSummary `json:"preferred_capacity_or_quantity"`
	PreferredMOQ               RangeSummary `json:"preferred_moq"`
	PreferredCertifications    []string     `json:"preferred_certifications"`
	PreferredDeliveryAreas     []string     `json:"preferred_delivery_areas"`
	PurchaseFrequency          *string      `json:"purchase_frequency"`
}

type RecommendationItem struct {
	EntityID           string   `json:"entity_id"`
	EntityType         string   `json:"entity_type"`
	CompatibilityScore int      `json:"compatibility_score"`
	Reasoning          []string `json:"reasoning"`
	EvidenceSource     []string `json:"evidence_source"`
}

type RecommendationResponse struct {
	ProfileSummary  ProfileSummary       `json:"profile_summary"`
	Recommendations []RecommendationItem `json:"recommendations"`
}

type MatchmakingRequest struct {
	PostID string `json:"post_id,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type MatchCandidate struct {
	PartnerID                string                 `json:"partner_id"`
	EntityID                 string                 `json:"entity_id,omitempty"`
	EntityType               string                 `json:"entity_type,omitempty"`
	CompatibilityScore       int                    `json:"compatibility_score"`
	MatchStatus              string                 `json:"match_status"`
	MatchedFields            []string               `json:"matched_fields"`
	MissingOrConflictingReqs []string               `json:"missing_or_conflicting_requirements"`
	Reasoning                []string               `json:"reasoning"`
	SuggestedAction          string                 `json:"suggested_action"`
	PublicData               map[string]interface{} `json:"public_data,omitempty"`
}

type MatchmakingResponse struct {
	SourcePostID                     string           `json:"source_post_id"`
	SourceType                       string           `json:"source_type"`
	CompatibilityScore               int              `json:"compatibility_score"`
	MatchStatus                      string           `json:"match_status"`
	MatchedFields                    []string         `json:"matched_fields"`
	MissingOrConflictingRequirements []string         `json:"missing_or_conflicting_requirements"`
	Reasoning                        []string         `json:"reasoning"`
	SuggestedAction                  string           `json:"suggested_action"`
	SuggestedPartners                []MatchCandidate `json:"suggested_partners"`
}

type AgreementVerificationResponse struct {
	AgreementID           string                          `json:"agreement_id"`
	MatchID               string                          `json:"match_id"`
	CanRevealContact      bool                            `json:"can_reveal_contact"`
	RecommendedApproval   bool                            `json:"recommended_approval"`
	OverallStatus         string                          `json:"overall_status"`
	VerificationStatus    string                          `json:"verification_status,omitempty"`
	FieldComparison       map[string]AgreementFieldResult `json:"field_comparison"`
	Conflicts             []string                        `json:"conflicts"`
	MissingInformation    []string                        `json:"missing_information"`
	NormalizedAgreement   AgreementSubmission             `json:"normalized_agreement"`
	Summary               string                          `json:"summary"`
	NextAction            string                          `json:"next_action"`
	ConfirmedTerms        []string                        `json:"confirmed_terms,omitempty"`
	MissingTerms          []string                        `json:"missing_terms,omitempty"`
	NegotiationHighlights []string                        `json:"negotiation_highlights,omitempty"`
	Risks                 []string                        `json:"risks,omitempty"`
	RecommendedNextSteps  []string                        `json:"recommended_next_steps,omitempty"`
}

type AgreementVerificationRequest struct {
	BuyerSubmission      AgreementSubmission `json:"buyer_submission"`
	ProducerSubmission   AgreementSubmission `json:"producer_submission"`
	BuyerFinalConfirm    bool                `json:"buyer_final_confirm"`
	ProducerFinalConfirm bool                `json:"producer_final_confirm"`
}

type AgreementSubmission struct {
	BuyerCompany        string   `json:"buyer_company"`
	ProducerCompany     string   `json:"producer_company"`
	Product             string   `json:"product"`
	Quantity            *float64 `json:"quantity"`
	Unit                string   `json:"unit"`
	AgreedUnitPrice     *float64 `json:"agreed_unit_price"`
	AgreedTotalPrice    *float64 `json:"agreed_total_price"`
	Currency            string   `json:"currency"`
	DeliveryArea        string   `json:"delivery_area"`
	DeliverySchedule    string   `json:"delivery_schedule"`
	PaymentTerms        string   `json:"payment_terms"`
	Certifications      []string `json:"certifications,omitempty"`
	QualityRequirements []string `json:"quality_requirements,omitempty"`
	AdditionalTerms     []string `json:"additional_terms"`
}

type AgreementFieldResult struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type NegotiationSummaryResponse struct {
	AgreementDraft  AgreementSubmission   `json:"agreement_draft"`
	AgreedTerms     []string              `json:"agreed_terms"`
	UnresolvedTerms []string              `json:"unresolved_terms"`
	Evidence        []NegotiationEvidence `json:"evidence"`
	Summary         string                `json:"summary"`
}

type NegotiationEvidence struct {
	Field            string `json:"field"`
	BuyerEvidence    string `json:"buyer_evidence"`
	ProducerEvidence string `json:"producer_evidence"`
}

type SearchHistoryInput struct {
	Query    string `json:"query"`
	Category string `json:"category,omitempty"`
	Location string `json:"location,omitempty"`
}

type SearchHistoryResponse struct {
	ID        string    `json:"id"`
	Query     string    `json:"query"`
	Category  string    `json:"category,omitempty"`
	Location  string    `json:"location,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

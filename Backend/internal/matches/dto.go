package matches

import "time"

type InterestRequest struct {
	PartnerID    string `json:"partner_id,omitempty"`
	SupplyPostID string `json:"supply_post_id,omitempty"`
	DemandPostID string `json:"demand_post_id,omitempty"`
}

type MatchResponse struct {
	ID         string      `json:"id"`
	BuyerID    string      `json:"buyer_id"`
	ProducerID string      `json:"producer_id"`
	Status     MatchStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func newMatchResponse(match *Match) MatchResponse {
	return MatchResponse{
		ID:         match.ID,
		BuyerID:    match.BuyerID,
		ProducerID: match.ProducerID,
		Status:     match.Status,
		CreatedAt:  match.CreatedAt,
		UpdatedAt:  match.UpdatedAt,
	}
}

func newMatchResponses(records []Match) []MatchResponse {
	responses := make([]MatchResponse, 0, len(records))
	for i := range records {
		responses = append(responses, newMatchResponse(&records[i]))
	}
	return responses
}

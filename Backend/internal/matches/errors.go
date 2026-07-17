package matches

import "errors"

var (
	ErrUnauthorized    = errors.New("authentication is required")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidRequest  = errors.New("invalid request")
	ErrPartnerRequired = errors.New("partner_id, supply_post_id, or demand_post_id is required")
	ErrPartnerNotFound = errors.New("partner not found")
	ErrInvalidMatchID  = errors.New("match id is invalid")
	ErrMatchNotFound   = errors.New("match not found")
	ErrCannotMatchSelf = errors.New("cannot create a match with yourself")
	ErrUnsupportedRole = errors.New("user role cannot create this match")
)

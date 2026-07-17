package ai

import "errors"

var (
	ErrUnauthorized       = errors.New("authentication is required")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrUserNotFound       = errors.New("user not found")
	ErrPostNotFound       = errors.New("post not found")
	ErrAgreementNotFound  = errors.New("agreement not found")
	ErrMatchNotFound      = errors.New("match not found")
	ErrAIUnavailable      = errors.New("ai provider is unavailable")
	ErrInvalidAIResponse  = errors.New("ai provider returned invalid response")
	ErrAgreementNotReady  = errors.New("agreement is not ready for contact reveal")
	ErrMissingAgreementID = errors.New("agreement id is required")
)

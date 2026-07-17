package users

import "errors"

var (
	ErrUnauthorized        = errors.New("authentication is required")
	ErrForbidden           = errors.New("insufficient permissions")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrInvalidUserID       = errors.New("user id is invalid")
	ErrVerificationNotFound = errors.New("verification not found")
	ErrRequiredCompanyName = errors.New("company_name cannot be empty")
	ErrRequiredPhone       = errors.New("phone cannot be empty")
	ErrRequiredCity        = errors.New("city cannot be empty")
	ErrRequiredNIBNumber   = errors.New("nib_number is required")
	ErrInvalidNIBNumber    = errors.New("nib_number is invalid")
	ErrInvalidVerificationStatus = errors.New("verification status is invalid")
	ErrRequiredRejectionReason = errors.New("rejection_reason is required when rejecting verification")
)

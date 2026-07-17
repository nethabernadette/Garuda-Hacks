package users

import (
	"strings"
	"unicode"
)

func validateProfileRequest(req UpdateProfileRequest, requireCoreFields bool) error {
	if requireCoreFields {
		if req.CompanyName == nil {
			return ErrRequiredCompanyName
		}
		if req.Phone == nil {
			return ErrRequiredPhone
		}
		if req.City == nil {
			return ErrRequiredCity
		}
	}

	if req.CompanyName != nil && strings.TrimSpace(*req.CompanyName) == "" {
		return ErrRequiredCompanyName
	}
	if req.Phone != nil && strings.TrimSpace(*req.Phone) == "" {
		return ErrRequiredPhone
	}
	if req.City != nil && strings.TrimSpace(*req.City) == "" {
		return ErrRequiredCity
	}

	return nil
}

func validateNIBVerificationRequest(req NIBVerificationRequest) error {
	nibNumber := strings.TrimSpace(req.NIBNumber)
	if nibNumber == "" {
		return ErrRequiredNIBNumber
	}
	if len(nibNumber) != 13 {
		return ErrInvalidNIBNumber
	}
	for _, value := range nibNumber {
		if !unicode.IsDigit(value) {
			return ErrInvalidNIBNumber
		}
	}

	return nil
}

func validateReviewNIBVerificationRequest(req ReviewNIBVerificationRequest) error {
	switch req.Status {
	case NIBVerificationStatusVerified:
		return nil
	case NIBVerificationStatusRejected:
		if strings.TrimSpace(req.RejectionReason) == "" {
			return ErrRequiredRejectionReason
		}
		return nil
	default:
		return ErrInvalidVerificationStatus
	}
}

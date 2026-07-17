package organizations

import "errors"

var (
	ErrOrgNotFound          = errors.New("organization not found")
	ErrAlreadyMember        = errors.New("user is already a member of this organization")
	ErrNotMember            = errors.New("user is not a member of this organization")
	ErrUnauthorizedAction   = errors.New("unauthorized action inside organization")
	ErrInvalidOrgRequest    = errors.New("invalid organization request")
	ErrRequiredOrgName      = errors.New("organization name is required")
	ErrMembershipNotFound   = errors.New("membership not found")
	ErrAlreadyOwner         = errors.New("user is already the owner of this organization")
	ErrCannotLeaveOwner     = errors.New("owner cannot leave organization without transferring ownership first")
)

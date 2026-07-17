package organizations

func validateCreateOrgRequest(req CreateOrgRequest) error {
	if req.Name == "" {
		return ErrRequiredOrgName
	}
	return nil
}

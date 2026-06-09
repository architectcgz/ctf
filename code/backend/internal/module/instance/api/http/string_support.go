package http

func normalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

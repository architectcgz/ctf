package commands

func normalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

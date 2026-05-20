package mapperhelper

func NormalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

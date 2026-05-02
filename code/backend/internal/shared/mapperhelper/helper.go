package mapperhelper

import "time"

func NormalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func CopyTimeToPtr(value time.Time) *time.Time {
	copied := value
	return &copied
}

func SingleString(value string) []string {
	return []string{value}
}

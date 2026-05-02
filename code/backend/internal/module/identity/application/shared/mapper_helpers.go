package shared

import (
	"time"

	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

func NormalizeOptionalString(value string) *string {
	return commonmapper.NormalizeOptionalString(value)
}

func SingleRole(role string) []string {
	return commonmapper.SingleString(role)
}

func CopyTimeToPtr(value time.Time) *time.Time {
	return commonmapper.CopyTimeToPtr(value)
}

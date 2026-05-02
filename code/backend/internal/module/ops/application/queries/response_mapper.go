package queries

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type notificationResponseMapper interface {
	// goverter:ignore Content
	// goverter:ignore Unread
	ToNotificationInfo(source model.Notification) dto.NotificationInfo
}

var notificationMapper notificationResponseMapper

func notificationContent(content string) *string {
	return commonmapper.NormalizeOptionalString(content)
}

func CopyTime(value time.Time) time.Time {
	return value
}

func CopyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

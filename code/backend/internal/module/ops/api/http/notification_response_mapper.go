package http

import (
	"time"

	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./notification_response_mapper_gen.go
// goverter:output:package :http
type notificationHTTPResponseMapper interface {
	ToNotificationInfo(source opsqry.NotificationInfo) NotificationInfo
	ToNotificationInfos(source []opsqry.NotificationInfo) []NotificationInfo

	ToAdminNotificationPublishResp(source opscmd.AdminNotificationPublishResp) AdminNotificationPublishResp
	ToAdminNotificationPublishRespPtr(source *opscmd.AdminNotificationPublishResp) *AdminNotificationPublishResp
}

var notificationResponseMapper notificationHTTPResponseMapper

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

func toNotificationInfo(source opsqry.NotificationInfo) NotificationInfo {
	return notificationResponseMapper.ToNotificationInfo(source)
}

func toNotificationInfos(source []opsqry.NotificationInfo) []NotificationInfo {
	return notificationResponseMapper.ToNotificationInfos(source)
}

func toAdminNotificationPublishResp(source *opscmd.AdminNotificationPublishResp) *AdminNotificationPublishResp {
	return notificationResponseMapper.ToAdminNotificationPublishRespPtr(source)
}

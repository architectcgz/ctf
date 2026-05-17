package http

import (
	opscmd "ctf-platform/internal/module/ops/application/commands"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type OpsRequestMapper interface {
	ToPublishAdminNotificationInput(source AdminNotificationPublishReq) opscmd.PublishAdminNotificationInput
}

var opsRequestMapper OpsRequestMapper

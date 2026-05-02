package commands

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :commands
type adminUserResponseMapper interface {
	// goverter:ignore Name
	// goverter:ignore Email
	// goverter:ignore StudentNo
	// goverter:ignore TeacherNo
	// goverter:ignore ClassName
	// goverter:ignore Roles
	// goverter:ignore UpdatedAt
	ToAdminUserResp(source model.User) dto.AdminUserResp
}

var adminUserMapper adminUserResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

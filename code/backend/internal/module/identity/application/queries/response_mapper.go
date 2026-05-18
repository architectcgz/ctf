package queries

import (
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type adminUserResponseMapper interface {
	// goverter:ignore Avatar
	// goverter:ignore Name
	// goverter:ignore ClassName
	ToProfileUserBase(source identitycontracts.User) identitycontracts.ProfileUser
	ToProfileUserBasePtr(source *identitycontracts.User) *identitycontracts.ProfileUser

	// goverter:ignore Name
	// goverter:ignore Email
	// goverter:ignore StudentNo
	// goverter:ignore TeacherNo
	// goverter:ignore ClassName
	// goverter:ignore Roles
	// goverter:ignore UpdatedAt
	ToAdminUserResp(source identitycontracts.User) identitycontracts.AdminUser
	ToAdminUserRespPtr(source *identitycontracts.User) *identitycontracts.AdminUser
}

var adminUserMapper adminUserResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

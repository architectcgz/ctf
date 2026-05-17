package http

import (
	"time"

	"ctf-platform/internal/dto"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type IdentityResponseMapper interface {
	ToAdminUserResp(source identitycontracts.AdminUser) dto.AdminUserResp
	ToAdminUserRespPtr(source *identitycontracts.AdminUser) *dto.AdminUserResp
	ToAdminUserResps(source []identitycontracts.AdminUser) []dto.AdminUserResp
	ToImportUsersResp(source identitycontracts.ImportUsersResult) dto.ImportUsersResp
	ToImportUsersRespPtr(source *identitycontracts.ImportUsersResult) *dto.ImportUsersResp
}

var identityResponseMapper IdentityResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

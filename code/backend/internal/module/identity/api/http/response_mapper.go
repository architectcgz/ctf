package http

import (
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type IdentityResponseMapper interface {
	ToAdminUserResp(source identitycontracts.AdminUser) AdminUserResp
	ToAdminUserRespPtr(source *identitycontracts.AdminUser) *AdminUserResp
	ToAdminUserResps(source []identitycontracts.AdminUser) []AdminUserResp
	ToImportUsersResp(source identitycontracts.ImportUsersResult) ImportUsersResp
	ToImportUsersRespPtr(source *identitycontracts.ImportUsersResult) *ImportUsersResp
}

var identityResponseMapper IdentityResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

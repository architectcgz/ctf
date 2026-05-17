package http

import (
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type IdentityRequestMapper interface {
	ToAdminUserListQuery(source AdminUserQuery) identitycontracts.AdminUserListQuery
	ToCreateUserInput(source CreateAdminUserReq) identitycontracts.CreateUserInput
	ToUpdateUserInput(source UpdateAdminUserReq) identitycontracts.UpdateUserInput
}

var identityRequestMapper IdentityRequestMapper

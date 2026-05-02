package http

import (
	"ctf-platform/internal/dto"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type IdentityRequestMapper interface {
	ToCreateUserInput(source dto.CreateAdminUserReq) identitycontracts.CreateUserInput
	ToUpdateUserInput(source dto.UpdateAdminUserReq) identitycontracts.UpdateUserInput
}

var identityRequestMapper IdentityRequestMapper

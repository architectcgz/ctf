package http

import (
	"ctf-platform/internal/dto"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type AuthRequestMapper interface {
	ToRegisterInput(source dto.RegisterReq) authcmd.RegisterInput
	ToLoginInput(source dto.LoginReq) authcmd.LoginInput
	ToChangePasswordInput(source dto.ChangePasswordReq) identitycontracts.ChangePasswordInput
}

var authRequestMapper AuthRequestMapper

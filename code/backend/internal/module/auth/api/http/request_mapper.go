package http

import (
	authcmd "ctf-platform/internal/module/auth/application/commands"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type AuthRequestMapper interface {
	ToRegisterInput(source RegisterReq) authcmd.RegisterInput
	ToLoginInput(source LoginReq) authcmd.LoginInput
	ToChangePasswordInput(source ChangePasswordReq) identitycontracts.ChangePasswordInput
}

var authRequestMapper AuthRequestMapper

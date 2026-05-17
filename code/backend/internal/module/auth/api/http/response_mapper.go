package http

import (
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type authHTTPResponseMapper interface {
	ToAuthUser(source identitycontracts.ProfileUser) AuthUser
	ToAuthUserPtr(source *identitycontracts.ProfileUser) *AuthUser

	ToLoginResp(source authcmd.LoginResp) LoginResp
	ToLoginRespPtr(source *authcmd.LoginResp) *LoginResp

	ToCASStatusResp(source authqry.CASStatusResp) CASStatusResp
	ToCASStatusRespPtr(source *authqry.CASStatusResp) *CASStatusResp

	ToCASLoginResp(source authqry.CASLoginResp) CASLoginResp
	ToCASLoginRespPtr(source *authqry.CASLoginResp) *CASLoginResp
}

var authResponseMapper authHTTPResponseMapper

func toAuthUser(source *identitycontracts.ProfileUser) *AuthUser {
	return authResponseMapper.ToAuthUserPtr(source)
}

func toLoginResp(source *authcmd.LoginResp) *LoginResp {
	return authResponseMapper.ToLoginRespPtr(source)
}

func toCASStatusResp(source *authqry.CASStatusResp) *CASStatusResp {
	return authResponseMapper.ToCASStatusRespPtr(source)
}

func toCASLoginResp(source *authqry.CASLoginResp) *CASLoginResp {
	return authResponseMapper.ToCASLoginRespPtr(source)
}

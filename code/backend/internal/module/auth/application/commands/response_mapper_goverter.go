package commands

import (
	"time"
)

type authUserSource struct {
	ID       int64
	Username string
	Role     string
}

type loginRespSource struct {
	User AuthUser
}

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type authCommandResponseMapper interface {
	// goverter:ignore Avatar
	// goverter:ignore Name
	// goverter:ignore ClassName
	ToAuthUserBase(source authUserSource) AuthUser
	ToAuthUserBasePtr(source *authUserSource) *AuthUser

	ToLoginResp(source loginRespSource) LoginResp
	ToLoginRespPtr(source loginRespSource) *LoginResp
}

var authCommandResponseMapperInst authCommandResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

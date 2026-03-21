package identity

import (
	authModule "ctf-platform/internal/module/auth"
	jwtpkg "ctf-platform/pkg/jwt"
)

type Authenticator interface {
	authModule.TokenService
	ParseAccessToken(token string) (*jwtpkg.Claims, error)
}

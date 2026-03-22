package identity

import (
	authcontracts "ctf-platform/internal/module/auth/contracts"
	jwtpkg "ctf-platform/pkg/jwt"
)

type Authenticator interface {
	authcontracts.TokenService
	ParseAccessToken(token string) (*jwtpkg.Claims, error)
}

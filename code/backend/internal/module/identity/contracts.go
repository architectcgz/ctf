package identity

import jwtpkg "ctf-platform/pkg/jwt"

type Authenticator interface {
	ParseAccessToken(token string) (*jwtpkg.Claims, error)
}

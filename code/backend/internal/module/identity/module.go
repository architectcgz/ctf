package identity

import (
	authModule "ctf-platform/internal/module/auth"
	jwtpkg "ctf-platform/pkg/jwt"
)

type Module struct {
	authModule.TokenService
}

func NewModule(tokenService authModule.TokenService) *Module {
	return &Module{TokenService: tokenService}
}

func (m *Module) ParseAccessToken(token string) (*jwtpkg.Claims, error) {
	claims, err := m.ParseToken(token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != jwtpkg.TokenTypeAccess {
		return nil, jwtpkg.ErrInvalidToken
	}
	return claims, nil
}

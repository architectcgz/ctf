package identity

import (
	authcontracts "ctf-platform/internal/module/auth/contracts"
	jwtpkg "ctf-platform/pkg/jwt"
)

type Module struct {
	authcontracts.TokenService
}

func NewModule(tokenService authcontracts.TokenService) *Module {
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

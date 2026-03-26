package commands

import (
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	jwtpkg "ctf-platform/pkg/jwt"
)

type AuthenticatorService struct {
	authcontracts.TokenService
}

var _ identitycontracts.Authenticator = (*AuthenticatorService)(nil)

func NewAuthenticatorService(tokenService authcontracts.TokenService) *AuthenticatorService {
	return &AuthenticatorService{TokenService: tokenService}
}

func (s *AuthenticatorService) ParseAccessToken(token string) (*jwtpkg.Claims, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != jwtpkg.TokenTypeAccess {
		return nil, jwtpkg.ErrInvalidToken
	}
	return claims, nil
}

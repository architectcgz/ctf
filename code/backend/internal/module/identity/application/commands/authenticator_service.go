package commands

import (
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type AuthenticatorService struct {
	authcontracts.TokenService
}

var _ identitycontracts.Authenticator = (*AuthenticatorService)(nil)

func NewAuthenticatorService(tokenService authcontracts.TokenService) *AuthenticatorService {
	return &AuthenticatorService{TokenService: tokenService}
}

package http

import (
	"ctf-platform/internal/dto"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

func toAuthUser(source *identitycontracts.ProfileUser) *dto.AuthUser {
	if source == nil {
		return nil
	}

	return &dto.AuthUser{
		ID:        source.ID,
		Username:  source.Username,
		Role:      source.Role,
		Avatar:    source.Avatar,
		Name:      source.Name,
		ClassName: source.ClassName,
	}
}

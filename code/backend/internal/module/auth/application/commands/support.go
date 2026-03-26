package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func buildAuthUser(user *model.User) dto.AuthUser {
	var name *string
	if user.Name != "" {
		name = &user.Name
	}
	var className *string
	if user.ClassName != "" {
		className = &user.ClassName
	}

	return dto.AuthUser{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Name:      name,
		ClassName: className,
	}
}

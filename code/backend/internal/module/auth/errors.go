package auth

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrRoleNotFound   = errors.New("role not found")
	ErrUsernameExists = errors.New("username exists")
	ErrEmailExists    = errors.New("email exists")
)

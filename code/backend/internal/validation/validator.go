package validation

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const UsernameTag = "ctf_username"

var (
	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	registerOnce    sync.Once
)

func Register() error {
	var registerErr error
	registerOnce.Do(func() {
		engine, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}

		registerErr = engine.RegisterValidation(UsernameTag, func(field validator.FieldLevel) bool {
			return IsValidUsername(field.Field().String())
		})
	})
	return registerErr
}

func IsValidUsername(value string) bool {
	return usernamePattern.MatchString(value)
}

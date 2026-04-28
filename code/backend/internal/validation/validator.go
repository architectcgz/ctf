package validation

import (
	"regexp"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const UsernameTag = "ctf_username"
const ImageNameTag = "ctf_image_name"
const ImageTagTag = "ctf_image_tag"

var (
	usernamePattern  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	imageNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:(?:[._/-]|:[0-9]+/)[a-z0-9]+)*$`)
	imageTagPattern  = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
	registerOnce     sync.Once
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
		if registerErr != nil {
			return
		}
		registerErr = engine.RegisterValidation(ImageNameTag, func(field validator.FieldLevel) bool {
			return IsValidImageName(field.Field().String())
		})
		if registerErr != nil {
			return
		}
		registerErr = engine.RegisterValidation(ImageTagTag, func(field validator.FieldLevel) bool {
			return IsValidImageTag(field.Field().String())
		})
	})
	return registerErr
}

func IsValidUsername(value string) bool {
	return usernamePattern.MatchString(value)
}

func IsValidImageName(value string) bool {
	return imageNamePattern.MatchString(value)
}

func IsValidImageTag(value string) bool {
	return imageTagPattern.MatchString(value)
}

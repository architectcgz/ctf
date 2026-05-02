//go:build !goverter

package commands

func init() {
	adminUserMapper = &adminUserResponseMapperImpl{}
}

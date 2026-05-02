//go:build !goverter

package commands

func init() {
	authCommandResponseMapperInst = &authCommandResponseMapperImpl{}
}

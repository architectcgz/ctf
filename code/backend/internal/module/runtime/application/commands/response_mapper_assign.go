//go:build !goverter

package commands

func init() {
	runtimeResponseMapper = &instanceResponseMapperImpl{}
}

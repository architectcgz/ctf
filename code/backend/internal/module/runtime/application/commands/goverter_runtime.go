//go:build !goverter

package commands

func init() {
	runtimeCommandResponseMapperInst = &RuntimeCommandResponseMapperImpl{}
}

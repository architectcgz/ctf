//go:build !goverter

package commands

func init() {
	challengeCommandResponseMapperInst = &challengeCommandResponseMapperImpl{}
}

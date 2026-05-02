//go:build !goverter

package domain

func init() {
	challengeResponseMapperInst = &challengeResponseMapperImpl{}
}

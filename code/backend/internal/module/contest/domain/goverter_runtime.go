//go:build !goverter

package domain

func init() {
	contestResponseMapperInst = &ContestResponseMapperImpl{}
}

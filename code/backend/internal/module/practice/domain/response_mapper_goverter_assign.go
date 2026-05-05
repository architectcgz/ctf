//go:build !goverter

package domain

func init() {
	practiceResponseMapperInst = &practiceResponseMapperImpl{}
}

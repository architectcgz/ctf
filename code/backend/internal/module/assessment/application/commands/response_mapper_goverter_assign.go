//go:build !goverter

package commands

func init() {
	assessmentCommandResponseMapperInst = &assessmentCommandResponseMapperImpl{}
}

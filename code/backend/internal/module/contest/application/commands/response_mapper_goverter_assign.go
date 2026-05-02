//go:build !goverter

package commands

func init() {
	contestResponseMapperInst = &contestResponseMapperImpl{}
}

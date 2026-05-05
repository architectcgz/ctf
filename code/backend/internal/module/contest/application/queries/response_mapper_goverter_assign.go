//go:build !goverter

package queries

func init() {
	contestQueryResponseMapperInst = &contestQueryResponseMapperImpl{}
}

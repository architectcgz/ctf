//go:build !goverter

package queries

func init() {
	authQueryResponseMapperInst = &authQueryResponseMapperImpl{}
}

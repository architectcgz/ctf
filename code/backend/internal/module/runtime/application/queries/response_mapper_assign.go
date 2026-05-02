//go:build !goverter

package queries

func init() {
	runtimeResponseMapper = &instanceResponseMapperImpl{}
}

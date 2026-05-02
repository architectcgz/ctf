//go:build !goverter

package http

func init() {
	contestRequestMapper = &ContestRequestMapperImpl{}
}

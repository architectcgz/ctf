//go:build !goverter

package http

func init() {
	authRequestMapper = &AuthRequestMapperImpl{}
}

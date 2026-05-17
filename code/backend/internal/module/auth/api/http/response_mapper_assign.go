//go:build !goverter

package http

func init() {
	authResponseMapper = &authHTTPResponseMapperImpl{}
}

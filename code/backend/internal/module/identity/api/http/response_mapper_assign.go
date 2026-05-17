//go:build !goverter

package http

func init() {
	identityResponseMapper = &IdentityResponseMapperImpl{}
}

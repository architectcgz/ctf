//go:build !goverter

package http

func init() {
	identityRequestMapper = &IdentityRequestMapperImpl{}
}

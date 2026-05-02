//go:build !goverter

package http

func init() {
	opsRequestMapper = &OpsRequestMapperImpl{}
}

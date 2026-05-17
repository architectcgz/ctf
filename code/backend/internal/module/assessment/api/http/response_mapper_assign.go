//go:build !goverter

package http

func init() {
	assessmentResponseMapper = &assessmentResponseMapperContractImpl{}
}

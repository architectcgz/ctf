//go:build !goverter

package http

func init() {
	practiceResponseMapper = &practiceResponseMapperContractImpl{}
}

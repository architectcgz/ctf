//go:build !goverter

package http

func init() {
	dashboardResponseMapper = &dashboardResponseMapperContractImpl{}
}

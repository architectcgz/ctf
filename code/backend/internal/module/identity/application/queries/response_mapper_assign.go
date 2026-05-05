//go:build !goverter

package queries

func init() {
	adminUserMapper = &adminUserResponseMapperImpl{}
}

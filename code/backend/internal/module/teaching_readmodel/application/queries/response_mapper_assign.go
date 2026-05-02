//go:build !goverter

package queries

func init() {
	teachingReadmodelMapper = &teachingReadmodelResponseMapperImpl{}
}

//go:build !goverter

package queries

func init() {
	teachingQueryMapper = &teachingQueryResponseMapperImpl{}
}

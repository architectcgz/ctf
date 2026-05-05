//go:build !goverter

package queries

func init() {
	practiceQueryResponseMapperInst = &practiceQueryResponseMapperImpl{}
}

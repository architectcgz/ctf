//go:build !goverter

package queries

func init() {
	challengeQueryResponseMapperInst = &challengeQueryResponseMapperImpl{}
}

//go:build !goverter

package queries

func init() {
	teacherAWDReviewMapper = &teacherAWDReviewResponseMapperImpl{}
}

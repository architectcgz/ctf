//go:build !goverter

package http

func init() {
	assessmentRequestMapper = &AssessmentRequestMapperImpl{}
}

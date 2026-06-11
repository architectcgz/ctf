//go:build !goverter

package queries

func init() {
	teachingAnalysisMapper = &teachingAnalysisResponseMapperImpl{}
}

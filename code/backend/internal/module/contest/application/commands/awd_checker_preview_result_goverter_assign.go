//go:build !goverter

package commands

func init() {
	awdPreviewResultMapper = &awdCheckerPreviewResultMapperImpl{}
}

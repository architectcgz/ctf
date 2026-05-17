package commands

import (
	contestdomain "ctf-platform/internal/module/contest/domain"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend ConvertAny
// goverter:output:file ./awd_checker_preview_result_goverter_gen.go
// goverter:output:package :commands
type awdCheckerPreviewResultMapper interface {
	ToStringAnyMap(source map[string]any) map[string]any
	ToDomain(source AWDCheckerPreviewResp) contestdomain.AWDCheckerPreviewResult
	ToDomainPtr(source *AWDCheckerPreviewResp) *contestdomain.AWDCheckerPreviewResult
	ToDTO(source contestdomain.AWDCheckerPreviewResult) AWDCheckerPreviewResp
	ToDTOPtr(source *contestdomain.AWDCheckerPreviewResult) *AWDCheckerPreviewResp
}

var awdPreviewResultMapper awdCheckerPreviewResultMapper

func ConvertAny(source any) any {
	return source
}

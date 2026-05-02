package commands

import (
	"ctf-platform/internal/dto"
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
	ToDomain(source dto.AWDCheckerPreviewResp) contestdomain.AWDCheckerPreviewResult
	ToDTO(source contestdomain.AWDCheckerPreviewResult) dto.AWDCheckerPreviewResp
}

var awdPreviewResultMapper awdCheckerPreviewResultMapper

func ConvertAny(source any) any {
	return source
}

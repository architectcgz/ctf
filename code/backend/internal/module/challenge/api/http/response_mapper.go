package http

import (
	"ctf-platform/internal/dto"
	"time"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type ChallengeResponseMapper interface {
	ToTagResp(source *dto.TagResp) *TagResp
	ToTagRespList(source []*dto.TagResp) []*TagResp
	ToImageResp(source *dto.ImageResp) *ImageResp
	ToImageRespList(source []*dto.ImageResp) []*ImageResp
	ToFlagResp(source *dto.FlagResp) *FlagResp
}

var challengeResponseMapper ChallengeResponseMapper

func CopyTime(source time.Time) time.Time {
	return source
}

func CopyTimePtr(source *time.Time) *time.Time {
	if source == nil {
		return nil
	}
	value := *source
	return &value
}

func mapImagePageResult(source *dto.PageResult[*dto.ImageResp]) *PageResult[*ImageResp] {
	if source == nil {
		return nil
	}

	return &PageResult[*ImageResp]{
		List:  challengeResponseMapper.ToImageRespList(source.List),
		Total: source.Total,
		Page:  source.Page,
		Size:  source.Size,
	}
}

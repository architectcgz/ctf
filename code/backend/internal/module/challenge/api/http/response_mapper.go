package http

import (
	"ctf-platform/internal/dto"
	"time"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:extend CopyAnyMap
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type ChallengeResponseMapper interface {
	ToChallengeResp(source *dto.ChallengeResp) *ChallengeResp
	ToChallengeRespList(source []*dto.ChallengeResp) []*ChallengeResp
	ToChallengeListItemResp(source *dto.ChallengeListItem) *ChallengeListItem
	ToChallengeListItemRespList(source []*dto.ChallengeListItem) []*ChallengeListItem
	ToChallengeDetailResp(source *dto.ChallengeDetailResp) *ChallengeDetailResp
	ToTagResp(source *dto.TagResp) *TagResp
	ToTagRespList(source []*dto.TagResp) []*TagResp
	ToImageResp(source *dto.ImageResp) *ImageResp
	ToImageRespList(source []*dto.ImageResp) []*ImageResp
	ToFlagResp(source *dto.FlagResp) *FlagResp
	ToChallengeTopologyResp(source *dto.ChallengeTopologyResp) *ChallengeTopologyResp
	ToEnvironmentTemplateResp(source *dto.EnvironmentTemplateResp) *EnvironmentTemplateResp
	ToEnvironmentTemplateRespList(source []*dto.EnvironmentTemplateResp) []*EnvironmentTemplateResp
	ToAWDChallengeResp(source *dto.AWDChallengeResp) *AWDChallengeResp
	ToAWDChallengeRespList(source []*dto.AWDChallengeResp) []*AWDChallengeResp
	ToAWDChallengeImportPreviewResp(source *dto.AWDChallengeImportPreviewResp) *AWDChallengeImportPreviewResp
	ToAWDChallengeImportPreviewRespList(source []dto.AWDChallengeImportPreviewResp) []AWDChallengeImportPreviewResp
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

func CopyAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
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

func mapChallengePageResult(source *dto.PageResult[*dto.ChallengeResp]) *PageResult[*ChallengeResp] {
	if source == nil {
		return nil
	}

	return &PageResult[*ChallengeResp]{
		List:  challengeResponseMapper.ToChallengeRespList(source.List),
		Total: source.Total,
		Page:  source.Page,
		Size:  source.Size,
	}
}

func mapChallengeListItemPageResult(source *dto.PageResult[*dto.ChallengeListItem]) *PageResult[*ChallengeListItem] {
	if source == nil {
		return nil
	}

	return &PageResult[*ChallengeListItem]{
		List:  challengeResponseMapper.ToChallengeListItemRespList(source.List),
		Total: source.Total,
		Page:  source.Page,
		Size:  source.Size,
	}
}

func mapAWDChallengePageResult(source *dto.AWDChallengePageResp) *AWDChallengePageResp {
	if source == nil {
		return nil
	}

	return &AWDChallengePageResp{
		Items: challengeResponseMapper.ToAWDChallengeRespList(source.Items),
		Total: source.Total,
		Page:  source.Page,
		Size:  source.Size,
	}
}

func mapAWDChallengeImportCommitResp(source *dto.AWDChallengeResp) *AWDChallengeImportCommitResp {
	return &AWDChallengeImportCommitResp{Challenge: challengeResponseMapper.ToAWDChallengeResp(source)}
}

func mapChallengeImportCommitResp(source *dto.ChallengeResp) *ChallengeImportCommitResp {
	return &ChallengeImportCommitResp{Challenge: challengeResponseMapper.ToChallengeResp(source)}
}

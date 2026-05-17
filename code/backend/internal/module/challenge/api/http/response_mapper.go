package http

import (
	"ctf-platform/internal/dto"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
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
	ToChallengePageResult(source *dto.PageResult[*dto.ChallengeResp]) *PageResult[*ChallengeResp]
	ToChallengeListItemResp(source *dto.ChallengeListItem) *ChallengeListItem
	ToChallengeListItemRespList(source []*dto.ChallengeListItem) []*ChallengeListItem
	ToChallengeListItemPageResult(source *dto.PageResult[*dto.ChallengeListItem]) *PageResult[*ChallengeListItem]
	ToChallengeDetailResp(source *dto.ChallengeDetailResp) *ChallengeDetailResp
	ToChallengePublishCheckJobResp(source *dto.ChallengePublishCheckJobResp) *ChallengePublishCheckJobResp
	ToChallengeSelfCheckResp(source *dto.ChallengeSelfCheckResp) *ChallengeSelfCheckResp
	ToChallengeImportPreviewResp(source *dto.ChallengeImportPreviewResp) *ChallengeImportPreviewResp
	ToChallengeImportPreviewRespList(source []dto.ChallengeImportPreviewResp) []ChallengeImportPreviewResp
	// goverter:map . Challenge
	ToChallengeImportCommitResp(source *dto.ChallengeResp) *ChallengeImportCommitResp
	ToChallengePackageExportResp(source *dto.ChallengePackageExportResp) *ChallengePackageExportResp
	ToTagResp(source *challengecontracts.TagResp) *TagResp
	ToTagRespList(source []*challengecontracts.TagResp) []*TagResp
	ToImageResp(source *challengecontracts.ImageResp) *ImageResp
	ToImageRespList(source []*challengecontracts.ImageResp) []*ImageResp
	ToImagePageResult(source *challengecontracts.PageResult[*challengecontracts.ImageResp]) *PageResult[*ImageResp]
	ToFlagResp(source *challengecontracts.FlagResp) *FlagResp
	ToChallengeTopologyResp(source *challengecontracts.ChallengeTopologyResp) *ChallengeTopologyResp
	ToEnvironmentTemplateResp(source *challengecontracts.EnvironmentTemplateResp) *EnvironmentTemplateResp
	ToEnvironmentTemplateRespList(source []*challengecontracts.EnvironmentTemplateResp) []*EnvironmentTemplateResp
	ToAWDChallengeResp(source *dto.AWDChallengeResp) *AWDChallengeResp
	ToAWDChallengeRespList(source []*dto.AWDChallengeResp) []*AWDChallengeResp
	ToAWDChallengePageResult(source *dto.AWDChallengePageResp) *AWDChallengePageResp
	ToAWDChallengeImportPreviewResp(source *dto.AWDChallengeImportPreviewResp) *AWDChallengeImportPreviewResp
	ToAWDChallengeImportPreviewRespList(source []dto.AWDChallengeImportPreviewResp) []AWDChallengeImportPreviewResp
	// goverter:map . Challenge
	ToAWDChallengeImportCommitResp(source *dto.AWDChallengeResp) *AWDChallengeImportCommitResp
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

func toChallengeResp(source *dto.ChallengeResp) *ChallengeResp {
	return challengeResponseMapper.ToChallengeResp(source)
}

func toChallengeDetailResp(source *dto.ChallengeDetailResp) *ChallengeDetailResp {
	return challengeResponseMapper.ToChallengeDetailResp(source)
}

func toChallengePublishCheckJobResp(source *dto.ChallengePublishCheckJobResp) *ChallengePublishCheckJobResp {
	return challengeResponseMapper.ToChallengePublishCheckJobResp(source)
}

func toChallengeSelfCheckResp(source *dto.ChallengeSelfCheckResp) *ChallengeSelfCheckResp {
	return challengeResponseMapper.ToChallengeSelfCheckResp(source)
}

func toChallengeImportPreviewResp(source *dto.ChallengeImportPreviewResp) *ChallengeImportPreviewResp {
	return challengeResponseMapper.ToChallengeImportPreviewResp(source)
}

func toChallengeImportPreviewRespList(source []dto.ChallengeImportPreviewResp) []ChallengeImportPreviewResp {
	return challengeResponseMapper.ToChallengeImportPreviewRespList(source)
}

func toChallengePackageExportResp(source *dto.ChallengePackageExportResp) *ChallengePackageExportResp {
	return challengeResponseMapper.ToChallengePackageExportResp(source)
}

func toTagResp(source *challengecontracts.TagResp) *TagResp {
	return challengeResponseMapper.ToTagResp(source)
}

func toTagRespList(source []*challengecontracts.TagResp) []*TagResp {
	return challengeResponseMapper.ToTagRespList(source)
}

func toImageResp(source *challengecontracts.ImageResp) *ImageResp {
	return challengeResponseMapper.ToImageResp(source)
}

func toFlagResp(source *challengecontracts.FlagResp) *FlagResp {
	return challengeResponseMapper.ToFlagResp(source)
}

func toChallengeTopologyResp(source *challengecontracts.ChallengeTopologyResp) *ChallengeTopologyResp {
	return challengeResponseMapper.ToChallengeTopologyResp(source)
}

func toEnvironmentTemplateResp(source *challengecontracts.EnvironmentTemplateResp) *EnvironmentTemplateResp {
	return challengeResponseMapper.ToEnvironmentTemplateResp(source)
}

func toEnvironmentTemplateRespList(source []*challengecontracts.EnvironmentTemplateResp) []*EnvironmentTemplateResp {
	return challengeResponseMapper.ToEnvironmentTemplateRespList(source)
}

func toAWDChallengeResp(source *dto.AWDChallengeResp) *AWDChallengeResp {
	return challengeResponseMapper.ToAWDChallengeResp(source)
}

func toAWDChallengeImportPreviewResp(source *dto.AWDChallengeImportPreviewResp) *AWDChallengeImportPreviewResp {
	return challengeResponseMapper.ToAWDChallengeImportPreviewResp(source)
}

func toAWDChallengeImportPreviewRespList(source []dto.AWDChallengeImportPreviewResp) []AWDChallengeImportPreviewResp {
	return challengeResponseMapper.ToAWDChallengeImportPreviewRespList(source)
}

func toImagePageResult(source *challengecontracts.PageResult[*challengecontracts.ImageResp]) *PageResult[*ImageResp] {
	return challengeResponseMapper.ToImagePageResult(source)
}

func toChallengePageResult(source *dto.PageResult[*dto.ChallengeResp]) *PageResult[*ChallengeResp] {
	return challengeResponseMapper.ToChallengePageResult(source)
}

func toChallengeListItemPageResult(source *dto.PageResult[*dto.ChallengeListItem]) *PageResult[*ChallengeListItem] {
	return challengeResponseMapper.ToChallengeListItemPageResult(source)
}

func toAWDChallengePageResult(source *dto.AWDChallengePageResp) *AWDChallengePageResp {
	return challengeResponseMapper.ToAWDChallengePageResult(source)
}

func toAWDChallengeImportCommitResp(source *dto.AWDChallengeResp) *AWDChallengeImportCommitResp {
	return challengeResponseMapper.ToAWDChallengeImportCommitResp(source)
}

func toChallengeImportCommitResp(source *dto.ChallengeResp) *ChallengeImportCommitResp {
	return challengeResponseMapper.ToChallengeImportCommitResp(source)
}

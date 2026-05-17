package http

import (
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
	ToChallengeResp(source *challengecontracts.ChallengeResp) *ChallengeResp
	ToChallengeRespList(source []*challengecontracts.ChallengeResp) []*ChallengeResp
	ToChallengePageResult(source *challengecontracts.PageResult[*challengecontracts.ChallengeResp]) *PageResult[*ChallengeResp]
	ToChallengeListItemResp(source *challengecontracts.ChallengeListItem) *ChallengeListItem
	ToChallengeListItemRespList(source []*challengecontracts.ChallengeListItem) []*ChallengeListItem
	ToChallengeListItemPageResult(source *challengecontracts.PageResult[*challengecontracts.ChallengeListItem]) *PageResult[*ChallengeListItem]
	ToChallengeDetailResp(source *challengecontracts.ChallengeDetailResp) *ChallengeDetailResp
	ToChallengePublishCheckJobResp(source *challengecontracts.ChallengePublishCheckJobResp) *ChallengePublishCheckJobResp
	ToChallengeSelfCheckResp(source *challengecontracts.ChallengeSelfCheckResp) *ChallengeSelfCheckResp
	ToChallengeImportPreviewResp(source *challengecontracts.ChallengeImportPreviewResp) *ChallengeImportPreviewResp
	ToChallengeImportPreviewRespList(source []challengecontracts.ChallengeImportPreviewResp) []ChallengeImportPreviewResp
	// goverter:map . Challenge
	ToChallengeImportCommitResp(source *challengecontracts.ChallengeResp) *ChallengeImportCommitResp
	ToChallengePackageExportResp(source *challengecontracts.ChallengePackageExportResp) *ChallengePackageExportResp
	ToTagResp(source *challengecontracts.TagResp) *TagResp
	ToTagRespList(source []*challengecontracts.TagResp) []*TagResp
	ToImageResp(source *challengecontracts.ImageResp) *ImageResp
	ToImageRespList(source []*challengecontracts.ImageResp) []*ImageResp
	ToImagePageResult(source *challengecontracts.PageResult[*challengecontracts.ImageResp]) *PageResult[*ImageResp]
	ToFlagResp(source *challengecontracts.FlagResp) *FlagResp
	ToChallengeTopologyResp(source *challengecontracts.ChallengeTopologyResp) *ChallengeTopologyResp
	ToEnvironmentTemplateResp(source *challengecontracts.EnvironmentTemplateResp) *EnvironmentTemplateResp
	ToEnvironmentTemplateRespList(source []*challengecontracts.EnvironmentTemplateResp) []*EnvironmentTemplateResp
	ToAWDChallengeResp(source *challengecontracts.AWDChallengeResp) *AWDChallengeResp
	ToAWDChallengeRespList(source []*challengecontracts.AWDChallengeResp) []*AWDChallengeResp
	ToAWDChallengePageResult(source *challengecontracts.AWDChallengePageResp) *AWDChallengePageResp
	ToAWDChallengeImportPreviewResp(source *challengecontracts.AWDChallengeImportPreviewResp) *AWDChallengeImportPreviewResp
	ToAWDChallengeImportPreviewRespList(source []challengecontracts.AWDChallengeImportPreviewResp) []AWDChallengeImportPreviewResp
	// goverter:map . Challenge
	ToAWDChallengeImportCommitResp(source *challengecontracts.AWDChallengeResp) *AWDChallengeImportCommitResp
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

func toChallengeResp(source *challengecontracts.ChallengeResp) *ChallengeResp {
	return challengeResponseMapper.ToChallengeResp(source)
}

func toChallengeDetailResp(source *challengecontracts.ChallengeDetailResp) *ChallengeDetailResp {
	return challengeResponseMapper.ToChallengeDetailResp(source)
}

func toChallengePublishCheckJobResp(source *challengecontracts.ChallengePublishCheckJobResp) *ChallengePublishCheckJobResp {
	return challengeResponseMapper.ToChallengePublishCheckJobResp(source)
}

func toChallengeSelfCheckResp(source *challengecontracts.ChallengeSelfCheckResp) *ChallengeSelfCheckResp {
	return challengeResponseMapper.ToChallengeSelfCheckResp(source)
}

func toChallengeImportPreviewResp(source *challengecontracts.ChallengeImportPreviewResp) *ChallengeImportPreviewResp {
	return challengeResponseMapper.ToChallengeImportPreviewResp(source)
}

func toChallengeImportPreviewRespList(source []challengecontracts.ChallengeImportPreviewResp) []ChallengeImportPreviewResp {
	return challengeResponseMapper.ToChallengeImportPreviewRespList(source)
}

func toChallengePackageExportResp(source *challengecontracts.ChallengePackageExportResp) *ChallengePackageExportResp {
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

func toAWDChallengeResp(source *challengecontracts.AWDChallengeResp) *AWDChallengeResp {
	return challengeResponseMapper.ToAWDChallengeResp(source)
}

func toAWDChallengeImportPreviewResp(source *challengecontracts.AWDChallengeImportPreviewResp) *AWDChallengeImportPreviewResp {
	return challengeResponseMapper.ToAWDChallengeImportPreviewResp(source)
}

func toAWDChallengeImportPreviewRespList(source []challengecontracts.AWDChallengeImportPreviewResp) []AWDChallengeImportPreviewResp {
	return challengeResponseMapper.ToAWDChallengeImportPreviewRespList(source)
}

func toImagePageResult(source *challengecontracts.PageResult[*challengecontracts.ImageResp]) *PageResult[*ImageResp] {
	return challengeResponseMapper.ToImagePageResult(source)
}

func toChallengePageResult(source *challengecontracts.PageResult[*challengecontracts.ChallengeResp]) *PageResult[*ChallengeResp] {
	return challengeResponseMapper.ToChallengePageResult(source)
}

func toChallengeListItemPageResult(source *challengecontracts.PageResult[*challengecontracts.ChallengeListItem]) *PageResult[*ChallengeListItem] {
	return challengeResponseMapper.ToChallengeListItemPageResult(source)
}

func toAWDChallengePageResult(source *challengecontracts.AWDChallengePageResp) *AWDChallengePageResp {
	return challengeResponseMapper.ToAWDChallengePageResult(source)
}

func toAWDChallengeImportCommitResp(source *challengecontracts.AWDChallengeResp) *AWDChallengeImportCommitResp {
	return challengeResponseMapper.ToAWDChallengeImportCommitResp(source)
}

func toChallengeImportCommitResp(source *challengecontracts.ChallengeResp) *ChallengeImportCommitResp {
	return challengeResponseMapper.ToChallengeImportCommitResp(source)
}

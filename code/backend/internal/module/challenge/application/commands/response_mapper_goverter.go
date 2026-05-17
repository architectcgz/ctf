package commands

import (
	"time"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengedomain "ctf-platform/internal/module/challenge/domain"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type challengeCommandResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore Active
	// goverter:ignore Result
	ToChallengePublishCheckJobRespBase(source model.ChallengePublishCheckJob) challengecontracts.ChallengePublishCheckJobResp
	ToChallengePublishCheckJobRespBasePtr(source *model.ChallengePublishCheckJob) *challengecontracts.ChallengePublishCheckJobResp

	ToChallengeImportAttachmentResps(source []challengedomain.ParsedChallengePackageAttachment) []challengecontracts.ChallengeImportAttachmentResp

	// goverter:ignore ID
	ToChallengeHintAdminRespFromParsed(source challengedomain.ParsedChallengePackageHint) challengecontracts.ChallengeHintAdminResp
	ToChallengeHintAdminRespFromParseds(source []challengedomain.ParsedChallengePackageHint) []challengecontracts.ChallengeHintAdminResp

	// goverter:ignore ID
	// goverter:ignore FileName
	// goverter:ignore Attachments
	// goverter:ignore Hints
	// goverter:ignore Flag
	// goverter:ignore Runtime
	// goverter:ignore ImageDelivery
	// goverter:ignore Extensions
	// goverter:ignore Topology
	// goverter:ignore PackageFiles
	// goverter:ignore Warnings
	// goverter:ignore CreatedAt
	ToChallengeImportPreviewRespBase(source challengedomain.ParsedChallengePackage) challengecontracts.ChallengeImportPreviewResp
	ToChallengeImportPreviewRespBasePtr(source *challengedomain.ParsedChallengePackage) *challengecontracts.ChallengeImportPreviewResp

	// goverter:ignore ID
	// goverter:ignore FileName
	// goverter:ignore CheckerConfig
	// goverter:ignore FlagConfig
	// goverter:ignore AccessConfig
	// goverter:ignore RuntimeConfig
	// goverter:ignore ImageDelivery
	// goverter:ignore Warnings
	// goverter:ignore CreatedAt
	ToAWDChallengeImportPreviewRespBase(source challengedomain.ParsedAWDChallengePackage) challengecontracts.AWDChallengeImportPreviewResp
	ToAWDChallengeImportPreviewRespBasePtr(source *challengedomain.ParsedAWDChallengePackage) *challengecontracts.AWDChallengeImportPreviewResp

	// goverter:map ID RevisionID
	// goverter:ignore ChallengeID
	// goverter:ignore FileName
	// goverter:ignore DownloadURL
	ToChallengePackageExportRespBase(source model.ChallengePackageRevision) challengecontracts.ChallengePackageExportResp
	ToChallengePackageExportRespBasePtr(source *model.ChallengePackageRevision) *challengecontracts.ChallengePackageExportResp
}

var challengeCommandResponseMapperInst challengeCommandResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

func CopyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

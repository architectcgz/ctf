package domain

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :domain
type challengeResponseMapper interface {
	// goverter:ignore Hints
	ToChallengeRespBase(source model.Challenge) dto.ChallengeResp
	ToChallengeRespBasePtr(source *model.Challenge) *dto.ChallengeResp

	// goverter:ignore CheckerConfig FlagConfig AccessConfig RuntimeConfig
	ToAWDChallengeRespBase(source model.AWDChallenge) dto.AWDChallengeResp
	ToAWDChallengeRespBasePtr(source *model.AWDChallenge) *dto.AWDChallengeResp

	ToChallengeHintAdminResp(source model.ChallengeHint) dto.ChallengeHintAdminResp
	ToChallengeHintAdminRespPtr(source *model.ChallengeHint) *dto.ChallengeHintAdminResp

	// goverter:ignore SizeFormatted
	ToImageRespBase(source model.Image) dto.ImageResp
	ToImageRespBasePtr(source *model.Image) *dto.ImageResp

	ToTagResp(source model.Tag) dto.TagResp
	ToTagRespPtr(source *model.Tag) *dto.TagResp
	ToAdminChallengeWriteupResp(source model.ChallengeWriteup) dto.AdminChallengeWriteupResp
	ToAdminChallengeWriteupRespPtr(source *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp
	ToSubmissionWriteupResp(source model.SubmissionWriteup) dto.SubmissionWriteupResp
	ToSubmissionWriteupRespPtr(source *model.SubmissionWriteup) *dto.SubmissionWriteupResp

	// goverter:map Submission.ID ID
	// goverter:map Submission.UserID UserID
	// goverter:map Submission.ChallengeID ChallengeID
	// goverter:map Submission.Title Title
	// goverter:map Submission.SubmissionStatus SubmissionStatus
	// goverter:map Submission.VisibilityStatus VisibilityStatus
	// goverter:map Submission.IsRecommended IsRecommended
	// goverter:map Submission.PublishedAt PublishedAt
	// goverter:map Submission.UpdatedAt UpdatedAt
	// goverter:ignore ContentPreview
	ToTeacherSubmissionWriteupItemRespBase(source challengeports.TeacherSubmissionWriteupRecord) dto.TeacherSubmissionWriteupItemResp
	ToTeacherSubmissionWriteupItemRespBasePtr(source *challengeports.TeacherSubmissionWriteupRecord) *dto.TeacherSubmissionWriteupItemResp

	// goverter:map Submission SubmissionWriteupResp
	ToTeacherSubmissionWriteupDetailResp(source challengeports.TeacherSubmissionWriteupRecord) dto.TeacherSubmissionWriteupDetailResp
	ToTeacherSubmissionWriteupDetailRespPtr(source challengeports.TeacherSubmissionWriteupRecord) *dto.TeacherSubmissionWriteupDetailResp

	// goverter:ignore ID
	ToRecommendedChallengeSolutionRespBase(source challengeports.RecommendedSolutionRecord) dto.RecommendedChallengeSolutionResp
	ToRecommendedChallengeSolutionRespBasePtr(source *challengeports.RecommendedSolutionRecord) *dto.RecommendedChallengeSolutionResp

	ToChallengePackageRevisionResp(source model.ChallengePackageRevision) dto.ChallengePackageRevisionResp
	ToChallengePackageRevisionRespPtr(source *model.ChallengePackageRevision) *dto.ChallengePackageRevisionResp

	// goverter:map Submission.ID ID
	// goverter:map Submission.ChallengeID ChallengeID
	// goverter:map Submission.UserID UserID
	// goverter:map Submission.Title Title
	// goverter:map Submission.Content Content
	// goverter:map Submission.SubmissionStatus SubmissionStatus
	// goverter:map Submission.VisibilityStatus VisibilityStatus
	// goverter:map Submission.IsRecommended IsRecommended
	// goverter:map Submission.PublishedAt PublishedAt
	// goverter:map Submission.UpdatedAt UpdatedAt
	// goverter:ignore ContentPreview
	ToCommunityChallengeSolutionRespBase(source challengeports.CommunitySolutionRecord) dto.CommunityChallengeSolutionResp
	ToCommunityChallengeSolutionRespBasePtr(source *challengeports.CommunitySolutionRecord) *dto.CommunityChallengeSolutionResp
}

var challengeResponseMapperInst challengeResponseMapper

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

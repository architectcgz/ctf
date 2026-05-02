package domain

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :domain
type challengeResponseMapper interface {
	ToChallengeHintAdminResp(source model.ChallengeHint) dto.ChallengeHintAdminResp
	ToChallengeHintAdminRespPtr(source *model.ChallengeHint) *dto.ChallengeHintAdminResp
	ToTagResp(source model.Tag) dto.TagResp
	ToTagRespPtr(source *model.Tag) *dto.TagResp
	ToAdminChallengeWriteupResp(source model.ChallengeWriteup) dto.AdminChallengeWriteupResp
	ToAdminChallengeWriteupRespPtr(source *model.ChallengeWriteup) *dto.AdminChallengeWriteupResp
	ToSubmissionWriteupResp(source model.SubmissionWriteup) dto.SubmissionWriteupResp
	ToSubmissionWriteupRespPtr(source *model.SubmissionWriteup) *dto.SubmissionWriteupResp
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

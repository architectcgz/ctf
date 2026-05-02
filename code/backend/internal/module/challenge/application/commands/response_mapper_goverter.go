package commands

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
// goverter:output:package :commands
type challengeCommandResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore Active
	// goverter:ignore Result
	ToChallengePublishCheckJobRespBase(source model.ChallengePublishCheckJob) dto.ChallengePublishCheckJobResp
	ToChallengePublishCheckJobRespBasePtr(source *model.ChallengePublishCheckJob) *dto.ChallengePublishCheckJobResp
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

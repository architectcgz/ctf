package queries

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type challengeQueryResponseMapper interface {
	ToChallengeHintResp(source model.ChallengeHint) dto.ChallengeHintResp
	ToChallengeHintRespPtr(source *model.ChallengeHint) *dto.ChallengeHintResp
	ToChallengeHintRespsPtr(source []*model.ChallengeHint) []*dto.ChallengeHintResp

	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeListItemBase(source model.Challenge) dto.ChallengeListItem
	ToChallengeListItemBasePtr(source *model.Challenge) *dto.ChallengeListItem

	// goverter:ignore NeedTarget
	// goverter:ignore Hints
	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeDetailRespBase(source model.Challenge) dto.ChallengeDetailResp
	ToChallengeDetailRespBasePtr(source *model.Challenge) *dto.ChallengeDetailResp

	// goverter:ignore Configured
	ToFlagRespBase(source model.Challenge) dto.FlagResp
	ToFlagRespBasePtr(source *model.Challenge) *dto.FlagResp

	// goverter:ignore RequiresSpoilerWarning
	ToChallengeWriteupRespBase(source model.ChallengeWriteup) dto.ChallengeWriteupResp
	ToChallengeWriteupRespBasePtr(source *model.ChallengeWriteup) *dto.ChallengeWriteupResp
}

var challengeQueryResponseMapperInst challengeQueryResponseMapper

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

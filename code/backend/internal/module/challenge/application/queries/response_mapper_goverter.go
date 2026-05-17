package queries

import (
	"time"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type challengeQueryResponseMapper interface {
	ToChallengeHintResp(source model.ChallengeHint) challengecontracts.ChallengeHintResp
	ToChallengeHintRespPtr(source *model.ChallengeHint) *challengecontracts.ChallengeHintResp
	ToChallengeHintRespsPtr(source []*model.ChallengeHint) []*challengecontracts.ChallengeHintResp

	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeListItemBase(source model.Challenge) challengecontracts.ChallengeListItem
	ToChallengeListItemBasePtr(source *model.Challenge) *challengecontracts.ChallengeListItem

	// goverter:ignore NeedTarget
	// goverter:ignore Hints
	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeDetailRespBase(source model.Challenge) challengecontracts.ChallengeDetailResp
	ToChallengeDetailRespBasePtr(source *model.Challenge) *challengecontracts.ChallengeDetailResp

	// goverter:ignore Configured
	ToFlagRespBase(source model.Challenge) challengecontracts.FlagResp
	ToFlagRespBasePtr(source *model.Challenge) *challengecontracts.FlagResp

	// goverter:ignore RequiresSpoilerWarning
	ToChallengeWriteupRespBase(source model.ChallengeWriteup) challengecontracts.ChallengeWriteupResp
	ToChallengeWriteupRespBasePtr(source *model.ChallengeWriteup) *challengecontracts.ChallengeWriteupResp
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

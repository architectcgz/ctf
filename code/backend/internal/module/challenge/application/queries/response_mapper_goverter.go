package queries

import (
	"time"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type challengeQueryResponseMapper interface {
	ToChallengeHintAdminResp(source challengeentity.ChallengeHint) challengecontracts.ChallengeHintAdminResp
	ToChallengeHintAdminRespPtr(source *challengeentity.ChallengeHint) *challengecontracts.ChallengeHintAdminResp
	ToChallengeHintAdminRespsPtr(source []*challengeentity.ChallengeHint) []*challengecontracts.ChallengeHintAdminResp

	// goverter:ignore Hints
	ToChallengeRespBase(source challengeports.ChallengeReadModel) challengecontracts.ChallengeResp
	ToChallengeRespBasePtr(source *challengeports.ChallengeReadModel) *challengecontracts.ChallengeResp

	ToChallengeHintResp(source challengeentity.ChallengeHint) challengecontracts.ChallengeHintResp
	ToChallengeHintRespPtr(source *challengeentity.ChallengeHint) *challengecontracts.ChallengeHintResp
	ToChallengeHintRespsPtr(source []*challengeentity.ChallengeHint) []*challengecontracts.ChallengeHintResp

	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeListItemBase(source challengeports.ChallengeReadModel) challengecontracts.ChallengeListItem
	ToChallengeListItemBasePtr(source *challengeports.ChallengeReadModel) *challengecontracts.ChallengeListItem

	// goverter:ignore NeedTarget
	// goverter:ignore Hints
	// goverter:ignore SolvedCount
	// goverter:ignore TotalAttempts
	// goverter:ignore IsSolved
	ToChallengeDetailRespBase(source challengeports.ChallengeReadModel) challengecontracts.ChallengeDetailResp
	ToChallengeDetailRespBasePtr(source *challengeports.ChallengeReadModel) *challengecontracts.ChallengeDetailResp

	// goverter:ignore RequiresSpoilerWarning
	ToChallengeWriteupRespBase(source challengeentity.ChallengeWriteup) challengecontracts.ChallengeWriteupResp
	ToChallengeWriteupRespBasePtr(source *challengeentity.ChallengeWriteup) *challengecontracts.ChallengeWriteupResp
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

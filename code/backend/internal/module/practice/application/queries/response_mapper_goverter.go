package queries

import (
	"time"

	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type practiceQueryResponseMapper interface {
	// goverter:ignore Username
	ToUserScoreInfoBase(source model.UserScore) practicecontracts.UserScoreInfo
	ToUserScoreInfoBasePtr(source *model.UserScore) *practicecontracts.UserScoreInfo

	// goverter:ignore Rank
	// goverter:ignore Username
	// goverter:ignore ClassName
	ToRankingItemBase(source model.UserScore) practicecontracts.RankingItem
	ToRankingItemBasePtr(source *model.UserScore) *practicecontracts.RankingItem
}

var practiceQueryResponseMapperInst practiceQueryResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

package queries

import (
	"time"

	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type practiceQueryResponseMapper interface {
	// goverter:ignore Username
	ToUserScoreInfoBase(source practiceentity.UserScore) practicecontracts.UserScoreInfo
	ToUserScoreInfoBasePtr(source *practiceentity.UserScore) *practicecontracts.UserScoreInfo

	// goverter:ignore Rank
	// goverter:ignore Username
	// goverter:ignore ClassName
	ToRankingItemBase(source practiceentity.UserScore) practicecontracts.RankingItem
	ToRankingItemBasePtr(source *practiceentity.UserScore) *practicecontracts.RankingItem
}

var practiceQueryResponseMapperInst practiceQueryResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

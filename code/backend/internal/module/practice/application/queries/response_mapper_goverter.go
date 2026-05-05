package queries

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type practiceQueryResponseMapper interface {
	// goverter:ignore Username
	ToUserScoreInfoBase(source model.UserScore) dto.UserScoreInfo
	ToUserScoreInfoBasePtr(source *model.UserScore) *dto.UserScoreInfo

	// goverter:ignore Rank
	// goverter:ignore Username
	// goverter:ignore ClassName
	ToRankingItemBase(source model.UserScore) dto.RankingItem
	ToRankingItemBasePtr(source *model.UserScore) *dto.RankingItem
}

var practiceQueryResponseMapperInst practiceQueryResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

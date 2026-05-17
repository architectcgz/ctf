package http

import (
	"ctf-platform/internal/dto"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type assessmentResponseMapperContract interface {
	ToRecommendationWeakDimension(source assessmentcontracts.RecommendationWeakDimension) dto.RecommendationWeakDimension
	ToRecommendationResp(source assessmentcontracts.Recommendation) dto.RecommendationResp
	ToRecommendationRespPtr(source *assessmentcontracts.Recommendation) *dto.RecommendationResp
	ToChallengeRecommendation(source assessmentcontracts.ChallengeRecommendation) dto.ChallengeRecommendation
	ToChallengeRecommendationPtr(source *assessmentcontracts.ChallengeRecommendation) *dto.ChallengeRecommendation
}

var assessmentResponseMapper assessmentResponseMapperContract

package http

import (
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type assessmentResponseMapperContract interface {
	ToRecommendationWeakDimension(source assessmentcontracts.RecommendationWeakDimension) RecommendationWeakDimension
	ToRecommendationResp(source assessmentcontracts.Recommendation) RecommendationResp
	ToRecommendationRespPtr(source *assessmentcontracts.Recommendation) *RecommendationResp
	ToChallengeRecommendation(source assessmentcontracts.ChallengeRecommendation) ChallengeRecommendation
	ToChallengeRecommendationPtr(source *assessmentcontracts.ChallengeRecommendation) *ChallengeRecommendation
}

var assessmentResponseMapper assessmentResponseMapperContract

package http

import (
	"ctf-platform/internal/dto"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type ChallengeRequestMapper interface {
	ToCreateAWDChallengeInput(source dto.CreateAWDChallengeReq) challengecmd.CreateAWDChallengeInput
	ToUpdateAWDChallengeInput(source dto.UpdateAWDChallengeReq) challengecmd.UpdateAWDChallengeInput
	ToListAWDChallengesInput(source dto.AWDChallengeQuery) challengeqry.ListAWDChallengesInput
	ToCreateImageInput(source dto.CreateImageReq) challengecmd.CreateImageInput
	ToUpdateImageInput(source dto.UpdateImageReq) challengecmd.UpdateImageInput
	ToListImagesInput(source dto.ImageQuery) challengeqry.ListImagesInput
}

var challengeRequestMapper ChallengeRequestMapper

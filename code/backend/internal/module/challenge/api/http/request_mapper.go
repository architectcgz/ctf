package http

import (
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type ChallengeRequestMapper interface {
	ToCreateChallengeInput(source CreateChallengeReq) challengecmd.CreateChallengeInput
	ToUpdateChallengeInput(source UpdateChallengeReq) challengecmd.UpdateChallengeInput
	ToCreateAWDChallengeInput(source CreateAWDChallengeReq) challengecmd.CreateAWDChallengeInput
	ToUpdateAWDChallengeInput(source UpdateAWDChallengeReq) challengecmd.UpdateAWDChallengeInput
	ToCreateTagInput(source CreateTagReq) challengecmd.CreateTagInput
	ToSaveChallengeTopologyInput(source SaveChallengeTopologyReq) challengecmd.SaveChallengeTopologyInput
	ToUpsertEnvironmentTemplateInput(source UpsertEnvironmentTemplateReq) challengecmd.UpsertEnvironmentTemplateInput
	ToUpsertOfficialWriteupInput(source challengecontracts.UpsertChallengeWriteupReq) challengecmd.UpsertOfficialWriteupInput
	ToUpsertSubmissionWriteupInput(source challengecontracts.UpsertSubmissionWriteupReq) challengecmd.UpsertSubmissionWriteupInput
	ToChallengeQuery(source ChallengeQuery) challengecontracts.ChallengeQuery
	ToListAWDChallengesInput(source AWDChallengeQuery) challengeqry.ListAWDChallengesInput
	ToCreateImageInput(source CreateImageReq) challengecmd.CreateImageInput
	ToUpdateImageInput(source UpdateImageReq) challengecmd.UpdateImageInput
	ToListImagesInput(source ImageQuery) challengeqry.ListImagesInput
}

var challengeRequestMapper ChallengeRequestMapper

package commands

import (
	"time"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type adminAWDInstanceItemRespSource struct {
	TeamID    int64
	ServiceID int64
	Instance  *instancecontracts.InstanceResp
}

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type practiceCommandResponseMapper interface {
	// goverter:map ID TeamID
	// goverter:map Name TeamName
	ToAdminAWDInstanceTeamResp(source model.Team) AdminAWDInstanceTeamResp
	ToAdminAWDInstanceTeamRespPtr(source *model.Team) *AdminAWDInstanceTeamResp

	// goverter:map ID ServiceID
	ToAdminAWDInstanceServiceResp(source model.ContestAWDService) AdminAWDInstanceServiceResp
	ToAdminAWDInstanceServiceRespPtr(source *model.ContestAWDService) *AdminAWDInstanceServiceResp

	// goverter:ignore Status
	// goverter:ignore Answer
	ToChallengeSubmissionRecordRespBase(source model.Submission) ChallengeSubmissionRecordResp
	ToChallengeSubmissionRecordRespBasePtr(source *model.Submission) *ChallengeSubmissionRecordResp

	ToAdminAWDInstanceItemResp(source adminAWDInstanceItemRespSource) AdminAWDInstanceItemResp
	ToAdminAWDInstanceItemRespPtr(source adminAWDInstanceItemRespSource) *AdminAWDInstanceItemResp
}

var practiceCommandResponseMapperInst practiceCommandResponseMapper

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

package commands

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type adminAWDInstanceItemRespSource struct {
	TeamID    int64
	ServiceID int64
	Instance  *instancecontracts.InstanceResp
}

type challengeSubmissionRecordRespSource struct {
	ID          int64
	SubmittedAt time.Time
}

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type practiceCommandResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore Answer
	ToChallengeSubmissionRecordRespBase(source challengeSubmissionRecordRespSource) ChallengeSubmissionRecordResp
	ToChallengeSubmissionRecordRespBasePtr(source *challengeSubmissionRecordRespSource) *ChallengeSubmissionRecordResp

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

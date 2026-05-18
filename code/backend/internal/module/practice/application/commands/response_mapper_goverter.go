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

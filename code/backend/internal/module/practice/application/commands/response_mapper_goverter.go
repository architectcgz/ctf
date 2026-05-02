package commands

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type adminAWDInstanceItemRespSource struct {
	TeamID    int64
	ServiceID int64
	Instance  *dto.InstanceResp
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
	ToAdminAWDInstanceTeamResp(source model.Team) dto.AdminAWDInstanceTeamResp
	ToAdminAWDInstanceTeamRespPtr(source *model.Team) *dto.AdminAWDInstanceTeamResp

	// goverter:map ID ServiceID
	ToAdminAWDInstanceServiceResp(source model.ContestAWDService) dto.AdminAWDInstanceServiceResp
	ToAdminAWDInstanceServiceRespPtr(source *model.ContestAWDService) *dto.AdminAWDInstanceServiceResp

	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore Answer
	// goverter:ignore ReviewerName
	ToTeacherManualReviewSubmissionDetailRespBase(source model.Submission) dto.TeacherManualReviewSubmissionDetailResp
	ToTeacherManualReviewSubmissionDetailRespBasePtr(source *model.Submission) *dto.TeacherManualReviewSubmissionDetailResp

	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore AnswerPreview
	ToTeacherManualReviewSubmissionItemRespBase(source model.Submission) dto.TeacherManualReviewSubmissionItemResp
	ToTeacherManualReviewSubmissionItemRespBasePtr(source *model.Submission) *dto.TeacherManualReviewSubmissionItemResp

	// goverter:ignore Status
	// goverter:ignore Answer
	ToChallengeSubmissionRecordRespBase(source model.Submission) dto.ChallengeSubmissionRecordResp
	ToChallengeSubmissionRecordRespBasePtr(source *model.Submission) *dto.ChallengeSubmissionRecordResp

	ToAdminAWDInstanceItemResp(source adminAWDInstanceItemRespSource) dto.AdminAWDInstanceItemResp
	ToAdminAWDInstanceItemRespPtr(source adminAWDInstanceItemRespSource) *dto.AdminAWDInstanceItemResp
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

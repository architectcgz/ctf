package commands

import (
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type practiceCommandResponseMapper interface {
	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore Answer
	// goverter:ignore ReviewerName
	ToTeacherManualReviewSubmissionDetailRespBase(source model.Submission) dto.TeacherManualReviewSubmissionDetailResp

	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore AnswerPreview
	ToTeacherManualReviewSubmissionItemRespBase(source model.Submission) dto.TeacherManualReviewSubmissionItemResp

	// goverter:ignore Status
	// goverter:ignore Answer
	ToChallengeSubmissionRecordRespBase(source model.Submission) dto.ChallengeSubmissionRecordResp
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

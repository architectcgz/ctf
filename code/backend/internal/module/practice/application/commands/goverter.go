package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ./zz_generated.goverter.go
// goverter:skipCopySameType
type PracticeCommandResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore Answer
	ToChallengeSubmissionRecordRespBase(source model.Submission) *dto.ChallengeSubmissionRecordResp
	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore ReviewerName
	// goverter:ignore Answer
	ToTeacherManualReviewSubmissionDetailRespBase(source model.Submission) *dto.TeacherManualReviewSubmissionDetailResp
	// goverter:ignore StudentUsername
	// goverter:ignore StudentName
	// goverter:ignore ClassName
	// goverter:ignore ChallengeTitle
	// goverter:ignore AnswerPreview
	ToTeacherManualReviewSubmissionItemRespBase(source model.Submission) *dto.TeacherManualReviewSubmissionItemResp
}

var practiceCommandResponseMapperInst PracticeCommandResponseMapper

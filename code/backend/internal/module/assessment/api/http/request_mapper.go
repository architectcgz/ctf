package http

import (
	"ctf-platform/internal/dto"
	assessmentcommands "ctf-platform/internal/module/assessment/application/commands"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type AssessmentRequestMapper interface {
	ToCreatePersonalReportInput(source dto.CreatePersonalReportReq) assessmentcommands.CreatePersonalReportInput
	ToCreateClassReportInput(source dto.CreateClassReportReq) assessmentcommands.CreateClassReportInput
	ToCreateContestExportInput(source dto.CreateContestExportReq) assessmentcommands.CreateContestExportInput
	ToCreateStudentReviewArchiveInput(source dto.CreateStudentReviewArchiveReq) assessmentcommands.CreateStudentReviewArchiveInput
	ToCreateTeacherAWDReviewExportInput(source dto.CreateTeacherAWDReviewExportReq) assessmentcommands.CreateTeacherAWDReviewExportInput
}

var assessmentRequestMapper AssessmentRequestMapper

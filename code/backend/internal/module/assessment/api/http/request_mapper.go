package http

import (
	assessmentcommands "ctf-platform/internal/module/assessment/application/commands"
	assessmentqueries "ctf-platform/internal/module/assessment/application/queries"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./request_mapper_gen.go
// goverter:output:package :http
type AssessmentRequestMapper interface {
	ToCreatePersonalReportInput(source CreatePersonalReportReq) assessmentcommands.CreatePersonalReportInput
	ToCreateClassReportInput(source CreateClassReportReq) assessmentcommands.CreateClassReportInput
	ToCreateContestExportInput(source CreateContestExportReq) assessmentcommands.CreateContestExportInput
	ToCreateStudentReviewArchiveInput(source CreateStudentReviewArchiveReq) assessmentcommands.CreateStudentReviewArchiveInput
	ToCreateTeacherAWDReviewExportInput(source CreateTeacherAWDReviewExportReq) assessmentcommands.CreateTeacherAWDReviewExportInput
	ToGetTeacherAWDReviewArchiveInput(source GetTeacherAWDReviewArchiveReq) assessmentqueries.GetTeacherAWDReviewArchiveInput
}

var assessmentRequestMapper AssessmentRequestMapper

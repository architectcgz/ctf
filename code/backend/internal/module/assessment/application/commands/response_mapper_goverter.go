package commands

import assessmententity "ctf-platform/internal/module/assessment/entity"

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type assessmentCommandResponseMapper interface {
	// goverter:map ID ReportID
	// goverter:ignore DownloadURL
	// goverter:ignore ExpiresAt
	// goverter:ignore ErrorMessage
	ToReportExportDataBase(source assessmententity.Report) ReportExportData
	ToReportExportDataBasePtr(source *assessmententity.Report) *ReportExportData
}

var assessmentCommandResponseMapperInst assessmentCommandResponseMapper

package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ./zz_generated.goverter.go
// goverter:skipCopySameType
type AssessmentCommandResponseMapper interface {
	// goverter:map ID ReportID
	// goverter:ignore DownloadURL
	// goverter:ignore ExpiresAt
	// goverter:ignore ErrorMessage
	ToReportExportDataBase(source *model.Report) *dto.ReportExportData
}

var assessmentCommandResponseMapperInst AssessmentCommandResponseMapper

package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :commands
type assessmentCommandResponseMapper interface {
	// goverter:map ID ReportID
	// goverter:ignore DownloadURL
	// goverter:ignore ExpiresAt
	// goverter:ignore ErrorMessage
	ToReportExportDataBase(source model.Report) dto.ReportExportData
	ToReportExportDataBasePtr(source *model.Report) *dto.ReportExportData
}

var assessmentCommandResponseMapperInst assessmentCommandResponseMapper

package domain

import (
	"strings"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type ReportDownload struct {
	Path        string
	FileName    string
	ContentType string
}

type ReportUser struct {
	ID        int64
	Username  string
	ClassName string
	Role      string
}

type PersonalReportStats struct {
	TotalScore    int
	TotalSolved   int
	TotalAttempts int
	Rank          int
}

type ReportDimensionStat struct {
	Dimension string
	Solved    int
	Total     int
}

type ClassDimensionAverage struct {
	Dimension string
	AvgScore  float64
}

type ClassTopStudent struct {
	UserID     int64
	Username   string
	TotalScore int
	Rank       int
}

func ValidateReportAccess(report *model.Report, requesterID int64, role string) error {
	if role == model.RoleAdmin {
		return nil
	}
	if report.UserID == nil || *report.UserID != requesterID {
		return errcode.ErrForbidden
	}
	return nil
}

func FillMissingDimensionAverages(rows []ClassDimensionAverage) []ClassDimensionAverage {
	index := make(map[string]float64, len(rows))
	for _, row := range rows {
		index[row.Dimension] = row.AvgScore
	}

	filled := make([]ClassDimensionAverage, 0, len(model.AllDimensions))
	for _, dimension := range model.AllDimensions {
		filled = append(filled, ClassDimensionAverage{
			Dimension: dimension,
			AvgScore:  index[dimension],
		})
	}
	return filled
}

func NormalizeReportConfig(cfg config.ReportConfig) config.ReportConfig {
	if strings.TrimSpace(cfg.StorageDir) == "" {
		cfg.StorageDir = "storage/exports"
	}
	if cfg.DefaultFormat != model.ReportFormatPDF && cfg.DefaultFormat != model.ReportFormatExcel {
		cfg.DefaultFormat = model.ReportFormatPDF
	}
	if cfg.PersonalTimeout <= 0 {
		cfg.PersonalTimeout = 30 * time.Second
	}
	if cfg.ClassTimeout <= 0 {
		cfg.ClassTimeout = 2 * time.Minute
	}
	if cfg.FileTTL <= 0 {
		cfg.FileTTL = 7 * 24 * time.Hour
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 2
	}
	return cfg
}

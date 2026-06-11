package commands

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	assessmentreporting "ctf-platform/internal/module/assessment/application/reporting"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	"ctf-platform/internal/teaching/classwindow"
)

func (s *ReportService) renderReport(filePath, format string, data any) error {
	switch format {
	case assessmententity.ReportFormatJSON:
		return assessmentreporting.WriteJSONReport(filePath, data)
	case assessmententity.ReportFormatExcel:
		switch payload := data.(type) {
		case *personalReportData:
			return assessmentreporting.WritePersonalExcel(filePath, payload)
		case *classReportData:
			return assessmentreporting.WriteClassExcel(filePath, payload)
		}
	default:
		switch payload := data.(type) {
		case *personalReportData:
			return assessmentreporting.WritePersonalPDF(filePath, payload)
		case *classReportData:
			return assessmentreporting.WriteClassPDF(filePath, payload)
		}
	}
	return apperror.ErrInternal.WithCause(fmt.Errorf("unsupported report payload"))
}

func (s *ReportService) reportFilePath(reportID int64, reportType, format string) (string, error) {
	storageDir := filepath.Clean(s.config.StorageDir)
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		return "", err
	}
	extension := reportFileExtension(format)
	fileName := fmt.Sprintf("%s-%d-%d.%s", reportType, reportID, time.Now().Unix(), extension)
	return filepath.Join(storageDir, fileName), nil
}

func (s *ReportService) safeReportPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absStorage, err := filepath.Abs(s.config.StorageDir)
	if err != nil {
		return "", err
	}
	prefix := absStorage + string(os.PathSeparator)
	if absPath != absStorage && !strings.HasPrefix(absPath, prefix) {
		return "", fmt.Errorf("unsafe path")
	}
	return absPath, nil
}

func (s *ReportService) normalizeFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case assessmententity.ReportFormatJSON:
		return assessmententity.ReportFormatJSON
	case assessmententity.ReportFormatExcel:
		return assessmententity.ReportFormatExcel
	case assessmententity.ReportFormatPDF:
		return assessmententity.ReportFormatPDF
	default:
		return s.config.DefaultFormat
	}
}

func (s *ReportService) normalizeArchiveFormat(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatJSON) {
		return assessmententity.ReportFormatJSON
	}
	return assessmententity.ReportFormatJSON
}

func (s *ReportService) parseClassWindow(req CreateClassReportInput) (classwindow.Range, error) {
	window, err := classwindow.Parse(reportNow(), req.FromDate, req.ToDate)
	if err != nil {
		return classwindow.Range{}, apperror.ErrInvalidParams.WithMessage(err.Error())
	}
	return window, nil
}

func reportFileExtension(format string) string {
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatJSON) {
		return "json"
	}
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatZIP) {
		return "zip"
	}
	if strings.EqualFold(strings.TrimSpace(format), assessmententity.ReportFormatExcel) {
		return "xlsx"
	}
	return "pdf"
}

func reportOutputFormat(report *assessmententity.Report) string {
	if report == nil {
		return assessmententity.ReportFormatPDF
	}
	switch report.Type {
	case assessmententity.ReportTypeAWDReviewArchive:
		return assessmententity.ReportFormatZIP
	case assessmententity.ReportTypeAWDReviewReport:
		return assessmententity.ReportFormatPDF
	default:
		return report.Format
	}
}

func reportDownloadFileName(report *assessmententity.Report) string {
	return fmt.Sprintf("%s-report-%d.%s", report.Type, report.ID, reportFileExtension(reportOutputFormat(report)))
}

func reportContentType(format string) string {
	return mime.TypeByExtension("." + reportFileExtension(format))
}

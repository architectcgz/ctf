package reporting

import (
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"ctf-platform/internal/module/assessment/application/reportassets"
)

const reportPDFFontFamily = "report-cjk"

func registerReportPDFFonts(pdf *gofpdf.Fpdf) error {
	pdf.AddUTF8FontFromBytes(reportPDFFontFamily, "", reportassets.ReportPDFRegularFont())
	pdf.AddUTF8FontFromBytes(reportPDFFontFamily, "B", reportassets.ReportPDFBoldFont())
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("register report pdf fonts: %w", err)
	}
	return nil
}

func setReportPDFFont(pdf *gofpdf.Fpdf, style string, size float64) {
	pdf.SetFont(reportPDFFontFamily, style, size)
}

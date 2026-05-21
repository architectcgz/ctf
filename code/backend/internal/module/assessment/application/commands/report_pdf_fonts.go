package commands

import (
	_ "embed"
	"fmt"

	"github.com/jung-kurt/gofpdf"
)

const reportPDFFontFamily = "report-cjk"

//go:embed reportfonts/NotoSansCJKsc-Regular.ttf
var reportPDFFontRegular []byte

func registerReportPDFFonts(pdf *gofpdf.Fpdf) error {
	pdf.AddUTF8FontFromBytes(reportPDFFontFamily, "", reportPDFFontRegular)
	pdf.AddUTF8FontFromBytes(reportPDFFontFamily, "B", reportPDFFontRegular)
	if err := pdf.Error(); err != nil {
		return fmt.Errorf("register report pdf fonts: %w", err)
	}
	return nil
}

func setReportPDFFont(pdf *gofpdf.Fpdf, style string, size float64) {
	pdf.SetFont(reportPDFFontFamily, style, size)
}

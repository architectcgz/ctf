package reporting

import (
	"strings"

	"github.com/jung-kurt/gofpdf"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"ctf-platform/internal/shared/taxonomy"
)

func newReportPDF() (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	if err := registerReportPDFFonts(pdf); err != nil {
		return nil, err
	}
	pdf.SetMargins(16, 16, 16)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddPage()
	setReportPDFFont(pdf, "", 12)
	return pdf, nil
}

func addReportTitle(pdf *gofpdf.Fpdf, title string) {
	setReportPDFFont(pdf, "B", 20)
	pdf.SetTextColor(24, 39, 75)
	pdf.CellFormat(0, 13, sanitizePDFText(title), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(180, 180, 180)
	pdf.Line(16, pdf.GetY(), 194, pdf.GetY())
	pdf.Ln(6)
}

func addSummaryBlock(pdf *gofpdf.Fpdf, lines []summaryLine) {
	addReportSectionTitle(pdf, "摘要")
	for _, line := range lines {
		setReportPDFFont(pdf, "B", 11)
		pdf.SetTextColor(45, 57, 84)
		pdf.CellFormat(45, 7, sanitizePDFText(line.Label), "0", 0, "L", false, 0, "")
		setReportPDFFont(pdf, "", 11)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(0, 7, sanitizePDFText(line.Value), "0", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addReportBulletSection(pdf *gofpdf.Fpdf, title string, items []string) {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		filtered = append(filtered, sanitizePDFText(item))
	}
	if len(filtered) == 0 {
		return
	}

	ensurePDFSpace(pdf, 16+float64(len(filtered))*8)
	addReportSectionTitle(pdf, title)
	setReportPDFFont(pdf, "", 11)
	for _, item := range filtered {
		ensurePDFSpace(pdf, 10)
		pdf.MultiCell(0, 7, sanitizePDFText("• "+item), "", "L", false)
	}
	pdf.Ln(2)
}

func addReportSectionTitle(pdf *gofpdf.Fpdf, title string) {
	ensurePDFSpace(pdf, 12)
	setReportPDFFont(pdf, "B", 15)
	pdf.SetFillColor(235, 240, 250)
	pdf.SetTextColor(24, 39, 75)
	pdf.CellFormat(0, 9, sanitizePDFText(title), "", 1, "L", true, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(1.5)
}

func writePDFTableHeader(pdf *gofpdf.Fpdf, headers []string) {
	pdf.SetFillColor(220, 230, 241)
	setReportPDFFont(pdf, "B", 10)
	widths := []float64{70, 40, 40}
	if len(headers) == 3 && headers[0] == "排名" {
		widths = []float64{30, 90, 30}
	}
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

func writePDFCustomTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFillColor(220, 230, 241)
	setReportPDFFont(pdf, "B", 10)
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

func writePDFTableRow(pdf *gofpdf.Fpdf, widths []float64, values []string) {
	for idx, value := range values {
		align := "L"
		if idx > 0 {
			align = "C"
		}
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(value), "1", 0, align, false, 0, "")
	}
	pdf.Ln(-1)
}

func ensurePDFSpace(pdf *gofpdf.Fpdf, needed float64) {
	if pdf.GetY()+needed <= 280 {
		return
	}
	pdf.AddPage()
}

func sanitizePDFText(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "-"
	}
	value = strings.ReplaceAll(value, "\r\n", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}

func localizeReportTerms(values []string) []string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, localizeReportTerm(value))
	}
	return items
}

func localizeReportTerm(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "":
		return ""
	case "final":
		return "最终快照"
	case contestcontracts.ContestStatusDraft:
		return "草稿"
	case contestcontracts.ContestStatusRegistration:
		return "报名中"
	case contestcontracts.ContestStatusFrozen:
		return "冻结中"
	case contestcontracts.ContestStatusEnded:
		return "已结束"
	case contestcontracts.AWDRoundStatusPending:
		return "待开始"
	case contestcontracts.ContestStatusRunning:
		return "进行中"
	case contestcontracts.AWDRoundStatusFinished:
		return "已完成"
	case contestcontracts.AWDServiceStatusUp:
		return "正常"
	case contestcontracts.AWDServiceStatusDown:
		return "下线"
	case contestcontracts.AWDServiceStatusCompromised:
		return "失陷"
	case contestcontracts.AWDAttackTypeFlagCapture:
		return "夺旗"
	case contestcontracts.AWDAttackTypeServiceExploit:
		return "服务利用"
	case contestcontracts.AWDAttackSourceManual:
		return "手工记录"
	case contestcontracts.AWDAttackSourceSubmission:
		return "提交记录"
	case contestcontracts.AWDTrafficSourceRuntimeProxy:
		return "代理流量"
	case "critical":
		return "高"
	case "warning":
		return "中"
	case "good":
		return "低"
	case taxonomy.DimensionWeb:
		return "Web"
	case taxonomy.DimensionPwn:
		return "Pwn"
	case taxonomy.DimensionReverse:
		return "逆向"
	case taxonomy.DimensionCrypto:
		return "密码"
	case taxonomy.DimensionMisc:
		return "综合"
	case taxonomy.DimensionForensics:
		return "取证"
	case taxonomy.DifficultyBeginner:
		return "入门"
	case taxonomy.DifficultyEasy:
		return "简单"
	case taxonomy.DifficultyMedium:
		return "中等"
	case taxonomy.DifficultyHard:
		return "困难"
	case taxonomy.DifficultyInsane:
		return "极难"
	default:
		return value
	}
}

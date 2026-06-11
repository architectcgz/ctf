package reporting

import (
	"fmt"
	"strings"

	"github.com/jung-kurt/gofpdf"

	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

func WritePersonalPDF(filePath string, data *PersonalReportData) error {
	pdf, err := newReportPDF()
	if err != nil {
		return err
	}
	addReportTitle(pdf, "个人训练报告")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "用户名", Value: sanitizePDFText(data.User.Username)},
		{Label: "班级", Value: sanitizePDFText(data.User.ClassName)},
		{Label: "总分", Value: fmt.Sprintf("%d", data.Stats.TotalScore)},
		{Label: "排名", Value: fmt.Sprintf("%d", data.Stats.Rank)},
		{Label: "解出题数", Value: fmt.Sprintf("%d", data.Stats.TotalSolved)},
		{Label: "尝试次数", Value: fmt.Sprintf("%d", data.Stats.TotalAttempts)},
	})
	addDimensionChart(pdf, "能力画像", skillProfileChartRows(data.SkillProfile))
	addDimensionStatsTable(pdf, "维度明细", data.DimensionStats)
	return pdf.OutputFileAndClose(filePath)
}

func WriteClassPDF(filePath string, data *ClassReportData) error {
	pdf, err := newReportPDF()
	if err != nil {
		return err
	}
	addReportTitle(pdf, "班级训练报告")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "班级", Value: sanitizePDFText(data.ClassName)},
		{Label: "统计窗口", Value: fmt.Sprintf("%s 至 %s（%d 天）", data.Window.FromDate, data.Window.ToDate, data.Window.Days)},
		{Label: "学生总数", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "平均分", Value: fmt.Sprintf("%.2f", data.AverageScore)},
		{Label: "活跃率", Value: fmt.Sprintf("%.0f%%", reviewSummaryActiveRate(data.Summary))},
		{Label: "近期事件数", Value: fmt.Sprintf("%d", reviewSummaryRecentEvents(data.Summary))},
	})
	addAverageChart(pdf, "维度平均分", data.DimensionAverages)
	addClassTrendTable(pdf, "趋势快照", data.Trend)
	addDistributionTable(pdf, "分类分布", data.CategoryDistribution)
	addDistributionTable(pdf, "难度分布", data.DifficultyDistribution)
	addTopStudentsTable(pdf, "班级前列学生", data.TopStudents)
	addContestMigrationSection(pdf, data.ContestMigration)
	addClassReviewOutlineTable(pdf, data.Review)
	return pdf.OutputFileAndClose(filePath)
}

func addDimensionChart(pdf *gofpdf.Fpdf, title string, rows []chartRow) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	for _, row := range rows {
		ensurePDFSpace(pdf, 12)
		setReportPDFFont(pdf, "", 10)
		pdf.CellFormat(28, 7, sanitizePDFText(localizeReportTerm(row.Label)), "", 0, "L", false, 0, "")
		x := pdf.GetX()
		y := pdf.GetY() + 2
		pdf.SetFillColor(232, 236, 241)
		pdf.Rect(x, y, 100, 4, "F")
		pdf.SetFillColor(79, 129, 189)
		pdf.Rect(x, y, 100*row.Value, 4, "F")
		pdf.SetX(x + 104)
		pdf.CellFormat(0, 7, fmt.Sprintf("%.0f%%", row.Value*100), "", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addAverageChart(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassDimensionAverage) {
	chartRows := make([]chartRow, 0, len(rows))
	for _, row := range rows {
		chartRows = append(chartRows, chartRow{Label: row.Dimension, Value: row.AvgScore})
	}
	addDimensionChart(pdf, title, chartRows)
}

func addDimensionStatsTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ReportDimensionStat) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"维度", "解出题数", "总题数"})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(70, 7, sanitizePDFText(localizeReportTerm(row.Dimension)), "1", 0, "L", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Solved), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%d", row.Total), "1", 1, "C", false, 0, "")
	}
}

func addTopStudentsTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassTopStudent) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFTableHeader(pdf, []string{"排名", "用户名", "分数"})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(90, 7, sanitizePDFText(row.Username), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 7, fmt.Sprintf("%d", row.TotalScore), "1", 1, "C", false, 0, "")
	}
}

func addClassTrendTable(pdf *gofpdf.Fpdf, title string, trend *ClassReportTrend) {
	points := safeTrendPoints(trend)
	if len(points) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"日期", "活跃学生", "事件数", "解题数"}, []float64{42, 36, 36, 36})
	setReportPDFFont(pdf, "", 10)
	for _, point := range points {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{42, 36, 36, 36}, []string{
			point.Date,
			fmt.Sprintf("%d", point.ActiveStudentCount),
			fmt.Sprintf("%d", point.EventCount),
			fmt.Sprintf("%d", point.SolveCount),
		})
	}
}

func addDistributionTable(pdf *gofpdf.Fpdf, title string, rows []assessmentdomain.ClassDistributionStat) {
	if len(rows) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, sanitizePDFText(title), "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"项目", "解出学生数", "已覆盖", "总数"}, []float64{46, 44, 36, 36})
	setReportPDFFont(pdf, "", 10)
	for _, row := range rows {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{46, 44, 36, 36}, []string{
			localizeReportTerm(row.Key),
			fmt.Sprintf("%d", row.SolvedStudents),
			fmt.Sprintf("%d", row.CoveredChallenges),
			fmt.Sprintf("%d", row.TotalChallenges),
		})
	}
}

func addContestMigrationSection(pdf *gofpdf.Fpdf, summary assessmentdomain.ClassContestMigrationSummary) {
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "竞赛迁移情况", "", 1, "L", false, 0, "")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "参赛学生数", Value: fmt.Sprintf("%d", summary.ParticipatingStudents)},
		{Label: "成功学生数", Value: fmt.Sprintf("%d", summary.SuccessfulStudents)},
		{Label: "攻击事件数", Value: fmt.Sprintf("%d", summary.AttackCount)},
		{Label: "成功事件数", Value: fmt.Sprintf("%d", summary.SuccessCount)},
		{Label: "成功维度", Value: strings.Join(localizeReportTerms(summary.SuccessDimensions), "、")},
	})
}

func addClassReviewOutlineTable(pdf *gofpdf.Fpdf, review *ClassReportReview) {
	items := safeReviewItems(review)
	if len(items) == 0 {
		return
	}
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "复盘提要", "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"编码", "级别", "维度", "学生"}, []float64{42, 32, 36, 68})
	setReportPDFFont(pdf, "", 10)
	for _, item := range items {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{42, 32, 36, 68}, []string{
			item.Code,
			localizeReportTerm(item.Severity),
			localizeReportTerm(item.Dimension),
			reviewStudentNames(item.Students),
		})
	}
}

type chartRow struct {
	Label string
	Value float64
}

func skillProfileChartRows(dimensions []*assessmentcontracts.SkillDimension) []chartRow {
	rows := make([]chartRow, 0, len(dimensions))
	for _, dimension := range dimensions {
		rows = append(rows, chartRow{
			Label: dimension.Dimension,
			Value: dimension.Score,
		})
	}
	return rows
}

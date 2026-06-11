package reporting

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"

	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

func WritePersonalExcel(filePath string, data *PersonalReportData) error {
	file := excelize.NewFile()
	defer file.Close()

	summarySheet := "Summary"
	file.SetSheetName("Sheet1", summarySheet)
	detailsSheet := "Dimensions"
	file.NewSheet(detailsSheet)

	headerStyle := mustNewExcelStyle(file, &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#D9E2F3"}},
	})

	writePairs(file, summarySheet, []summaryLine{
		{Label: "Username", Value: data.User.Username},
		{Label: "Class", Value: data.User.ClassName},
		{Label: "Total Score", Value: fmt.Sprintf("%d", data.Stats.TotalScore)},
		{Label: "Rank", Value: fmt.Sprintf("%d", data.Stats.Rank)},
		{Label: "Solved", Value: fmt.Sprintf("%d", data.Stats.TotalSolved)},
		{Label: "Attempts", Value: fmt.Sprintf("%d", data.Stats.TotalAttempts)},
	}, headerStyle)

	file.SetCellValue(summarySheet, "A10", "Dimension")
	file.SetCellValue(summarySheet, "B10", "Score")
	file.SetCellStyle(summarySheet, "A10", "B10", headerStyle)
	for idx, dimension := range data.SkillProfile {
		row := idx + 11
		file.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), dimension.Dimension)
		file.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), dimension.Score)
	}

	file.SetCellValue(detailsSheet, "A1", "Dimension")
	file.SetCellValue(detailsSheet, "B1", "Solved")
	file.SetCellValue(detailsSheet, "C1", "Total")
	file.SetCellStyle(detailsSheet, "A1", "C1", headerStyle)
	for idx, stat := range data.DimensionStats {
		row := idx + 2
		file.SetCellValue(detailsSheet, fmt.Sprintf("A%d", row), stat.Dimension)
		file.SetCellValue(detailsSheet, fmt.Sprintf("B%d", row), stat.Solved)
		file.SetCellValue(detailsSheet, fmt.Sprintf("C%d", row), stat.Total)
	}

	if len(data.SkillProfile) > 0 {
		_ = file.AddChart(summarySheet, "D2", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{{
				Name:       fmt.Sprintf("%s!$B$10", summarySheet),
				Categories: fmt.Sprintf("%s!$A$11:$A$%d", summarySheet, len(data.SkillProfile)+10),
				Values:     fmt.Sprintf("%s!$B$11:$B$%d", summarySheet, len(data.SkillProfile)+10),
			}},
			Title:  []excelize.RichTextRun{{Text: "Skill Profile"}},
			Legend: excelize.ChartLegend{Position: "bottom"},
		})
	}

	setReportSheetLayout(file, summarySheet)
	setReportSheetLayout(file, detailsSheet)
	return file.SaveAs(filePath)
}

func WriteClassExcel(filePath string, data *ClassReportData) error {
	file := excelize.NewFile()
	defer file.Close()

	summarySheet := "Summary"
	file.SetSheetName("Sheet1", summarySheet)
	trendSheet := "Trend"
	reviewSheet := "Review"
	categorySheet := "Category"
	difficultySheet := "Difficulty"
	migrationSheet := "Migration"
	topSheet := "TopStudents"
	file.NewSheet(trendSheet)
	file.NewSheet(reviewSheet)
	file.NewSheet(categorySheet)
	file.NewSheet(difficultySheet)
	file.NewSheet(migrationSheet)
	file.NewSheet(topSheet)

	headerStyle := mustNewExcelStyle(file, &excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FCE4D6"}},
	})

	writePairs(file, summarySheet, []summaryLine{
		{Label: "Class", Value: data.ClassName},
		{Label: "Window", Value: fmt.Sprintf("%s ~ %s (%d days)", data.Window.FromDate, data.Window.ToDate, data.Window.Days)},
		{Label: "Total Students", Value: fmt.Sprintf("%d", data.TotalStudents)},
		{Label: "Average Score", Value: fmt.Sprintf("%.2f", data.AverageScore)},
		{Label: "Active Rate", Value: fmt.Sprintf("%.0f%%", reviewSummaryActiveRate(data.Summary))},
		{Label: "Recent Events", Value: fmt.Sprintf("%d", reviewSummaryRecentEvents(data.Summary))},
	}, headerStyle)

	file.SetCellValue(summarySheet, "A9", "Dimension")
	file.SetCellValue(summarySheet, "B9", "Average Score")
	file.SetCellStyle(summarySheet, "A9", "B9", headerStyle)
	for idx, dimension := range data.DimensionAverages {
		row := idx + 10
		file.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), dimension.Dimension)
		file.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), dimension.AvgScore)
	}

	file.SetCellValue(trendSheet, "A1", "Date")
	file.SetCellValue(trendSheet, "B1", "Active Students")
	file.SetCellValue(trendSheet, "C1", "Events")
	file.SetCellValue(trendSheet, "D1", "Solves")
	file.SetCellStyle(trendSheet, "A1", "D1", headerStyle)
	for idx, point := range safeTrendPoints(data.Trend) {
		row := idx + 2
		file.SetCellValue(trendSheet, fmt.Sprintf("A%d", row), point.Date)
		file.SetCellValue(trendSheet, fmt.Sprintf("B%d", row), point.ActiveStudentCount)
		file.SetCellValue(trendSheet, fmt.Sprintf("C%d", row), point.EventCount)
		file.SetCellValue(trendSheet, fmt.Sprintf("D%d", row), point.SolveCount)
	}

	writeDistributionSheet(file, categorySheet, headerStyle, data.CategoryDistribution)
	writeDistributionSheet(file, difficultySheet, headerStyle, data.DifficultyDistribution)
	writeReviewSheet(file, reviewSheet, headerStyle, data.Review)
	writeContestMigrationSheet(file, migrationSheet, headerStyle, data.ContestMigration)

	file.SetCellValue(topSheet, "A1", "Rank")
	file.SetCellValue(topSheet, "B1", "Username")
	file.SetCellValue(topSheet, "C1", "Total Score")
	file.SetCellStyle(topSheet, "A1", "C1", headerStyle)
	for idx, student := range data.TopStudents {
		row := idx + 2
		file.SetCellValue(topSheet, fmt.Sprintf("A%d", row), student.Rank)
		file.SetCellValue(topSheet, fmt.Sprintf("B%d", row), student.Username)
		file.SetCellValue(topSheet, fmt.Sprintf("C%d", row), student.TotalScore)
	}

	if len(data.DimensionAverages) > 0 {
		_ = file.AddChart(summarySheet, "D2", &excelize.Chart{
			Type: excelize.Col,
			Series: []excelize.ChartSeries{{
				Name:       fmt.Sprintf("%s!$B$9", summarySheet),
				Categories: fmt.Sprintf("%s!$A$10:$A$%d", summarySheet, len(data.DimensionAverages)+9),
				Values:     fmt.Sprintf("%s!$B$10:$B$%d", summarySheet, len(data.DimensionAverages)+9),
			}},
			Title:  []excelize.RichTextRun{{Text: "Dimension Average"}},
			Legend: excelize.ChartLegend{Position: "bottom"},
		})
	}

	setReportSheetLayout(file, summarySheet)
	setReportSheetLayout(file, trendSheet)
	setReportSheetLayout(file, reviewSheet)
	setReportSheetLayout(file, categorySheet)
	setReportSheetLayout(file, difficultySheet)
	setReportSheetLayout(file, migrationSheet)
	setReportSheetLayout(file, topSheet)
	return file.SaveAs(filePath)
}

func mustNewExcelStyle(file *excelize.File, style *excelize.Style) int {
	styleID, _ := file.NewStyle(style)
	return styleID
}

func writePairs(file *excelize.File, sheet string, rows []summaryLine, headerStyle int) {
	for idx, row := range rows {
		line := idx + 1
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), row.Label)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), row.Value)
		file.SetCellStyle(sheet, fmt.Sprintf("A%d", line), fmt.Sprintf("A%d", line), headerStyle)
	}
}

func setReportSheetLayout(file *excelize.File, sheet string) {
	file.SetColWidth(sheet, "A", "A", 22)
	file.SetColWidth(sheet, "B", "E", 18)
}

func writeDistributionSheet(file *excelize.File, sheet string, headerStyle int, rows []assessmentdomain.ClassDistributionStat) {
	file.SetCellValue(sheet, "A1", "Key")
	file.SetCellValue(sheet, "B1", "Solved Students")
	file.SetCellValue(sheet, "C1", "Covered Challenges")
	file.SetCellValue(sheet, "D1", "Total Challenges")
	file.SetCellStyle(sheet, "A1", "D1", headerStyle)
	for idx, row := range rows {
		line := idx + 2
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), row.Key)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), row.SolvedStudents)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", line), row.CoveredChallenges)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", line), row.TotalChallenges)
	}
}

func writeReviewSheet(file *excelize.File, sheet string, headerStyle int, review *ClassReportReview) {
	file.SetCellValue(sheet, "A1", "Code")
	file.SetCellValue(sheet, "B1", "Severity")
	file.SetCellValue(sheet, "C1", "Dimension")
	file.SetCellValue(sheet, "D1", "Students")
	file.SetCellValue(sheet, "E1", "Summary")
	file.SetCellValue(sheet, "F1", "Evidence")
	file.SetCellValue(sheet, "G1", "Action")
	file.SetCellStyle(sheet, "A1", "G1", headerStyle)
	for idx, item := range safeReviewItems(review) {
		line := idx + 2
		file.SetCellValue(sheet, fmt.Sprintf("A%d", line), item.Code)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", line), item.Severity)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", line), item.Dimension)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", line), reviewStudentNames(item.Students))
		file.SetCellValue(sheet, fmt.Sprintf("E%d", line), item.Summary)
		file.SetCellValue(sheet, fmt.Sprintf("F%d", line), item.Evidence)
		file.SetCellValue(sheet, fmt.Sprintf("G%d", line), item.Action)
	}
}

func writeContestMigrationSheet(file *excelize.File, sheet string, headerStyle int, summary assessmentdomain.ClassContestMigrationSummary) {
	writePairs(file, sheet, []summaryLine{
		{Label: "Participating Students", Value: fmt.Sprintf("%d", summary.ParticipatingStudents)},
		{Label: "Successful Students", Value: fmt.Sprintf("%d", summary.SuccessfulStudents)},
		{Label: "Attack Events", Value: fmt.Sprintf("%d", summary.AttackCount)},
		{Label: "Success Events", Value: fmt.Sprintf("%d", summary.SuccessCount)},
		{Label: "Success Dimensions", Value: strings.Join(summary.SuccessDimensions, ", ")},
	}, headerStyle)
}

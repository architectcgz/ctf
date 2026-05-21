package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	"ctf-platform/internal/apperror"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
)

type awdReviewArchiveManifest struct {
	GeneratedAt       time.Time `json:"generated_at"`
	SnapshotType      string    `json:"snapshot_type"`
	ContestID         int64     `json:"contest_id"`
	ContestTitle      string    `json:"contest_title"`
	RoundCount        int       `json:"round_count"`
	TeamCount         int       `json:"team_count"`
	HasSelectedRound  bool      `json:"has_selected_round"`
	SelectedRound     *int      `json:"selected_round,omitempty"`
	RequestedByUserID int64     `json:"requested_by_user_id"`
}

func RenderAWDReviewArchiveZip(targetPath string, archive *assessmentqry.TeacherAWDReviewArchiveResp) error {
	if archive == nil {
		return apperror.ErrInternal.WithCause(fmt.Errorf("nil awd review archive"))
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	manifest := awdReviewArchiveManifest{
		GeneratedAt:       archive.GeneratedAt,
		SnapshotType:      archive.Scope.SnapshotType,
		ContestID:         archive.Contest.ID,
		ContestTitle:      archive.Contest.Title,
		RoundCount:        len(archive.Rounds),
		TeamCount:         len(extractAWDReviewTeams(archive)),
		HasSelectedRound:  archive.SelectedRound != nil,
		RequestedByUserID: archive.Scope.RequestedBy,
	}
	if archive.SelectedRound != nil {
		manifest.SelectedRound = &archive.SelectedRound.Round.RoundNumber
	}

	if err := writeZIPJSONFile(writer, "manifest.json", manifest); err != nil {
		return err
	}
	if err := writeZIPJSONFile(writer, "overview.json", archive.Overview); err != nil {
		return err
	}
	if err := writeZIPJSONFile(writer, "rounds.json", archive.Rounds); err != nil {
		return err
	}
	if err := writeZIPJSONFile(writer, "teams.json", extractAWDReviewTeams(archive)); err != nil {
		return err
	}
	if archive.SelectedRound != nil {
		if err := writeZIPJSONFile(writer, "selected-round.json", archive.SelectedRound); err != nil {
			return err
		}
	}
	return nil
}

func RenderAWDReviewReportPDF(targetPath string, archive *assessmentqry.TeacherAWDReviewArchiveResp) error {
	if archive == nil {
		return apperror.ErrInternal.WithCause(fmt.Errorf("nil awd review archive"))
	}

	pdf, err := newReportPDF()
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	addReportTitle(pdf, "教师 AWD 复盘报告")

	overview := archive.Overview
	if overview == nil {
		overview = &assessmentqry.TeacherAWDReviewOverviewResp{}
	}

	addSummaryBlock(pdf, []summaryLine{
		{Label: "赛事", Value: sanitizePDFText(archive.Contest.Title)},
		{Label: "快照类型", Value: sanitizePDFText(localizeReportTerm(archive.Scope.SnapshotType))},
		{Label: "状态", Value: sanitizePDFText(localizeReportTerm(archive.Contest.Status))},
		{Label: "轮次数", Value: fmt.Sprintf("%d", len(archive.Rounds))},
		{Label: "队伍数", Value: fmt.Sprintf("%d", overview.TeamCount)},
		{Label: "服务数", Value: fmt.Sprintf("%d", overview.ServiceCount)},
		{Label: "攻击数", Value: fmt.Sprintf("%d", overview.AttackCount)},
		{Label: "流量数", Value: fmt.Sprintf("%d", overview.TrafficCount)},
	})

	addAWDReviewKeyRoundSection(pdf, archive.Rounds)
	addAWDReviewRoundsTable(pdf, archive.Rounds)
	if archive.SelectedRound != nil {
		addAWDReviewSelectedRoundBlock(pdf, archive.SelectedRound)
		addAWDReviewServiceInsightSection(pdf, archive.SelectedRound)
		addAWDReviewAttackInsightSection(pdf, archive.SelectedRound)
		addAWDReviewSuggestionSection(pdf, archive.Rounds, archive.SelectedRound)
	}

	return pdf.OutputFileAndClose(targetPath)
}

func writeZIPJSONFile(writer *zip.Writer, name string, payload any) error {
	entry, err := writer.Create(name)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}

	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	content = append(content, '\n')
	if _, err := entry.Write(content); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func extractAWDReviewTeams(archive *assessmentqry.TeacherAWDReviewArchiveResp) []assessmentqry.TeacherAWDReviewTeamResp {
	if archive == nil || archive.SelectedRound == nil {
		return []assessmentqry.TeacherAWDReviewTeamResp{}
	}
	return archive.SelectedRound.Teams
}

func addAWDReviewRoundsTable(pdf *gofpdf.Fpdf, rounds []assessmentqry.TeacherAWDReviewRoundResp) {
	if len(rounds) == 0 {
		return
	}

	ensurePDFSpace(pdf, 20+float64(len(rounds))*8)
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "轮次概览", "", 1, "L", false, 0, "")

	headers := []string{"轮次", "状态", "服务", "攻击", "流量"}
	widths := []float64{24, 46, 32, 32, 32}
	setReportPDFFont(pdf, "B", 10)
	pdf.SetFillColor(230, 230, 230)
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	setReportPDFFont(pdf, "", 10)
	for _, round := range rounds {
		ensurePDFSpace(pdf, 8)
		values := []string{
			fmt.Sprintf("%d", round.RoundNumber),
			sanitizePDFText(localizeReportTerm(round.Status)),
			fmt.Sprintf("%d", round.ServiceCount),
			fmt.Sprintf("%d", round.AttackCount),
			fmt.Sprintf("%d", round.TrafficCount),
		}
		for idx, value := range values {
			pdf.CellFormat(widths[idx], 7, value, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.Ln(4)
}

func addAWDReviewSelectedRoundBlock(pdf *gofpdf.Fpdf, selected *assessmentqry.TeacherAWDSelectedRoundResp) {
	if selected == nil {
		return
	}

	ensurePDFSpace(pdf, 36)
	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "选中轮次摘要", "", 1, "L", false, 0, "")
	setReportPDFFont(pdf, "", 11)

	lines := []summaryLine{
		{Label: "轮次", Value: fmt.Sprintf("%d", selected.Round.RoundNumber)},
		{Label: "队伍数", Value: fmt.Sprintf("%d", len(selected.Teams))},
		{Label: "服务数", Value: fmt.Sprintf("%d", len(selected.Services))},
		{Label: "攻击数", Value: fmt.Sprintf("%d", len(selected.Attacks))},
		{Label: "流量数", Value: fmt.Sprintf("%d", len(selected.Traffic))},
	}
	for _, line := range lines {
		pdf.CellFormat(34, 7, sanitizePDFText(line.Label), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 7, sanitizePDFText(line.Value), "", 1, "L", false, 0, "")
	}
	pdf.Ln(3)
}

func addAWDReviewKeyRoundSection(pdf *gofpdf.Fpdf, rounds []assessmentqry.TeacherAWDReviewRoundResp) {
	if len(rounds) == 0 {
		return
	}

	roundsCopy := append([]assessmentqry.TeacherAWDReviewRoundResp(nil), rounds...)
	sort.SliceStable(roundsCopy, func(i, j int) bool {
		left := roundsCopy[i]
		right := roundsCopy[j]
		leftWeight := left.AttackCount*100 + left.TrafficCount*10 + left.ServiceCount
		rightWeight := right.AttackCount*100 + right.TrafficCount*10 + right.ServiceCount
		if leftWeight == rightWeight {
			return left.RoundNumber < right.RoundNumber
		}
		return leftWeight > rightWeight
	})

	highlights := roundsCopy
	if len(highlights) > 3 {
		highlights = highlights[:3]
	}

	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "关键轮次", "", 1, "L", false, 0, "")
	writePDFCustomTableHeader(pdf, []string{"轮次", "主要现象", "服务", "攻击", "流量"}, []float64{20, 72, 26, 26, 26})
	setReportPDFFont(pdf, "", 10)
	for _, round := range highlights {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{20, 72, 26, 26, 26}, []string{
			fmt.Sprintf("%d", round.RoundNumber),
			describeAWDRound(round),
			fmt.Sprintf("%d", round.ServiceCount),
			fmt.Sprintf("%d", round.AttackCount),
			fmt.Sprintf("%d", round.TrafficCount),
		})
	}
	pdf.Ln(4)
}

func addAWDReviewServiceInsightSection(pdf *gofpdf.Fpdf, selected *assessmentqry.TeacherAWDSelectedRoundResp) {
	if selected == nil || len(selected.Services) == 0 {
		return
	}

	upCount := 0
	abnormalCount := 0
	totalAttackReceived := 0
	topPressure := selected.Services[0]
	serviceRows := append([]assessmentqry.TeacherAWDReviewServiceResp(nil), selected.Services...)
	for _, item := range serviceRows {
		totalAttackReceived += item.AttackReceived
		if item.ServiceStatus == "up" {
			upCount++
		}
		if item.ServiceStatus != "up" {
			abnormalCount++
		}
		if serviceRiskScore(item) > serviceRiskScore(topPressure) {
			topPressure = item
		}
	}

	sort.SliceStable(serviceRows, func(i, j int) bool {
		left := serviceRiskScore(serviceRows[i])
		right := serviceRiskScore(serviceRows[j])
		if left == right {
			return serviceRows[i].TeamName < serviceRows[j].TeamName
		}
		return left > right
	})

	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "服务稳定性", "", 1, "L", false, 0, "")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "正常服务数", Value: fmt.Sprintf("%d", upCount)},
		{Label: "异常服务数", Value: fmt.Sprintf("%d", abnormalCount)},
		{Label: "累计受击次数", Value: fmt.Sprintf("%d", totalAttackReceived)},
		{Label: "重点服务", Value: fmt.Sprintf("%s / %s", sanitizePDFText(topPressure.TeamName), sanitizePDFText(topPressure.AWDChallengeTitle))},
	})

	writePDFCustomTableHeader(pdf, []string{"队伍", "服务", "状态", "受击", "防守分"}, []float64{34, 52, 32, 24, 30})
	setReportPDFFont(pdf, "", 10)
	limit := minInt(len(serviceRows), 5)
	for _, item := range serviceRows[:limit] {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{34, 52, 32, 24, 30}, []string{
			item.TeamName,
			item.AWDChallengeTitle,
			localizeReportTerm(item.ServiceStatus),
			fmt.Sprintf("%d", item.AttackReceived),
			fmt.Sprintf("%d", item.DefenseScore),
		})
	}
	pdf.Ln(4)
}

func addAWDReviewAttackInsightSection(pdf *gofpdf.Fpdf, selected *assessmentqry.TeacherAWDSelectedRoundResp) {
	if selected == nil {
		return
	}

	totalAttacks := len(selected.Attacks)
	successCount := 0
	manualCount := 0
	submissionCount := 0
	totalScore := 0
	topAttacker := ""
	topAttackerScore := 0
	attackerScores := make(map[string]int)
	for _, item := range selected.Attacks {
		if item.IsSuccess {
			successCount++
		}
		totalScore += item.ScoreGained
		switch item.Source {
		case "manual_attack_log":
			manualCount++
		case "submission":
			submissionCount++
		}
		attackerScores[item.AttackerTeamName] += item.ScoreGained
		if attackerScores[item.AttackerTeamName] > topAttackerScore {
			topAttacker = item.AttackerTeamName
			topAttackerScore = attackerScores[item.AttackerTeamName]
		}
	}
	successRate := "0%"
	if totalAttacks > 0 {
		successRate = fmt.Sprintf("%.0f%%", float64(successCount)*100/float64(totalAttacks))
	}

	trafficSummary := summarizeTraffic(selected.Traffic)

	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "攻击有效性", "", 1, "L", false, 0, "")
	addSummaryBlock(pdf, []summaryLine{
		{Label: "攻击总数", Value: fmt.Sprintf("%d", totalAttacks)},
		{Label: "成功率", Value: successRate},
		{Label: "累计得分", Value: fmt.Sprintf("%d", totalScore)},
		{Label: "手工记录", Value: fmt.Sprintf("%d", manualCount)},
		{Label: "提交记录", Value: fmt.Sprintf("%d", submissionCount)},
		{Label: "主要得分队伍", Value: fallbackString(topAttacker, "-")},
		{Label: "高频访问路径", Value: trafficSummary},
	})

	if totalAttacks == 0 {
		return
	}

	attackRows := append([]assessmentqry.TeacherAWDReviewAttackResp(nil), selected.Attacks...)
	sort.SliceStable(attackRows, func(i, j int) bool {
		if attackRows[i].ScoreGained == attackRows[j].ScoreGained {
			if attackRows[i].IsSuccess == attackRows[j].IsSuccess {
				return attackRows[i].CreatedAt.Before(attackRows[j].CreatedAt)
			}
			return attackRows[i].IsSuccess
		}
		return attackRows[i].ScoreGained > attackRows[j].ScoreGained
	})

	writePDFCustomTableHeader(pdf, []string{"攻击方", "目标方", "类型", "结果", "得分"}, []float64{34, 34, 44, 28, 28})
	setReportPDFFont(pdf, "", 10)
	limit := minInt(len(attackRows), 5)
	for _, item := range attackRows[:limit] {
		ensurePDFSpace(pdf, 8)
		writePDFTableRow(pdf, []float64{34, 34, 44, 28, 28}, []string{
			item.AttackerTeamName,
			item.VictimTeamName,
			localizeReportTerm(item.AttackType),
			boolLabel(item.IsSuccess),
			fmt.Sprintf("%d", item.ScoreGained),
		})
	}
	pdf.Ln(4)
}

func addAWDReviewSuggestionSection(pdf *gofpdf.Fpdf, rounds []assessmentqry.TeacherAWDReviewRoundResp, selected *assessmentqry.TeacherAWDSelectedRoundResp) {
	if selected == nil {
		return
	}

	suggestions := buildAWDReviewSuggestions(rounds, selected)
	if len(suggestions) == 0 {
		return
	}

	setReportPDFFont(pdf, "B", 14)
	pdf.CellFormat(0, 8, "复盘建议", "", 1, "L", false, 0, "")
	setReportPDFFont(pdf, "", 11)
	for _, item := range suggestions {
		ensurePDFSpace(pdf, 10)
		pdf.MultiCell(0, 7, sanitizePDFText("• "+item), "", "L", false)
	}
}

func describeAWDRound(round assessmentqry.TeacherAWDReviewRoundResp) string {
	switch {
	case round.AttackCount > 0:
		return "存在有效攻防事件，适合优先回看"
	case round.TrafficCount > 0:
		return "有访问流量但攻击较少，适合核对探测过程"
	case round.ServiceCount > 0:
		return "主要表现为服务状态变化"
	default:
		return "事件较少，可作为稳定轮次参考"
	}
}

func serviceRiskScore(item assessmentqry.TeacherAWDReviewServiceResp) int {
	score := item.AttackReceived*100 + item.DefenseScore + item.SLAScore + item.AttackScore
	switch item.ServiceStatus {
	case "compromised":
		score += 10000
	case "down":
		score += 8000
	case "up":
		score += 1000
	default:
		score += 2000
	}
	return score
}

func summarizeTraffic(items []assessmentqry.TeacherAWDReviewTrafficResp) string {
	if len(items) == 0 {
		return "-"
	}
	counts := make(map[string]int)
	bestPath := ""
	bestCount := 0
	for _, item := range items {
		path := strings.TrimSpace(item.Path)
		if path == "" {
			path = "/"
		}
		counts[path]++
		if counts[path] > bestCount {
			bestPath = path
			bestCount = counts[path]
		}
	}
	return fmt.Sprintf("%s（%d 次）", bestPath, bestCount)
}

func buildAWDReviewSuggestions(rounds []assessmentqry.TeacherAWDReviewRoundResp, selected *assessmentqry.TeacherAWDSelectedRoundResp) []string {
	suggestions := make([]string, 0, 3)
	if keyRound := hottestRound(rounds); keyRound != nil && (keyRound.AttackCount > 0 || keyRound.TrafficCount > 0 || keyRound.ServiceCount > 0) {
		suggestions = append(suggestions, fmt.Sprintf("优先回看第 %d 轮，这一轮的过程事件最集中，适合作为课堂复盘的主样本。", keyRound.RoundNumber))
	}

	riskyService := topRiskyService(selected.Services)
	if riskyService != nil && riskyService.AttackReceived > 0 {
		suggestions = append(suggestions, fmt.Sprintf("重点检查 %s 队的 %s 服务，该服务在本轮受击 %d 次，适合结合日志和修补过程做服务稳定性复盘。", riskyService.TeamName, riskyService.AWDChallengeTitle, riskyService.AttackReceived))
	}

	totalAttacks := len(selected.Attacks)
	successCount := 0
	for _, item := range selected.Attacks {
		if item.IsSuccess {
			successCount++
		}
	}
	switch {
	case totalAttacks == 0 && len(selected.Traffic) > 0:
		suggestions = append(suggestions, "本轮存在访问流量但没有形成有效攻击，适合复盘探测链路和利用链为何没有继续转化。")
	case totalAttacks > 0 && successCount == 0:
		suggestions = append(suggestions, "本轮有攻击尝试但没有形成有效得分，适合复盘 payload、目标识别和提交时机是否存在问题。")
	case totalAttacks > 0 && successCount < totalAttacks:
		suggestions = append(suggestions, "本轮攻击已出现成功样本，但整体成功率还不稳定，建议把成功与失败样本对照整理。")
	}

	return suggestions
}

func hottestRound(rounds []assessmentqry.TeacherAWDReviewRoundResp) *assessmentqry.TeacherAWDReviewRoundResp {
	if len(rounds) == 0 {
		return nil
	}
	best := rounds[0]
	bestWeight := best.AttackCount*100 + best.TrafficCount*10 + best.ServiceCount
	for _, item := range rounds[1:] {
		weight := item.AttackCount*100 + item.TrafficCount*10 + item.ServiceCount
		if weight > bestWeight {
			best = item
			bestWeight = weight
		}
	}
	return &best
}

func topRiskyService(items []assessmentqry.TeacherAWDReviewServiceResp) *assessmentqry.TeacherAWDReviewServiceResp {
	if len(items) == 0 {
		return nil
	}
	best := items[0]
	bestScore := serviceRiskScore(best)
	for _, item := range items[1:] {
		score := serviceRiskScore(item)
		if score > bestScore {
			best = item
			bestScore = score
		}
	}
	return &best
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func boolLabel(ok bool) string {
	if ok {
		return "成功"
	}
	return "失败"
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

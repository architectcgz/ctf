package commands

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
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

func RenderAWDReviewArchiveZip(targetPath string, archive *dto.TeacherAWDReviewArchiveResp) error {
	if archive == nil {
		return errcode.ErrInternal.WithCause(fmt.Errorf("nil awd review archive"))
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
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

func RenderAWDReviewReportPDF(targetPath string, archive *dto.TeacherAWDReviewArchiveResp) error {
	if archive == nil {
		return errcode.ErrInternal.WithCause(fmt.Errorf("nil awd review archive"))
	}

	pdf := newReportPDF()
	addReportTitle(pdf, "Teacher AWD Review Report")

	overview := archive.Overview
	if overview == nil {
		overview = &dto.TeacherAWDReviewOverviewResp{}
	}

	addSummaryBlock(pdf, []summaryLine{
		{Label: "Contest", Value: sanitizePDFText(archive.Contest.Title)},
		{Label: "Snapshot", Value: sanitizePDFText(archive.Scope.SnapshotType)},
		{Label: "Status", Value: sanitizePDFText(archive.Contest.Status)},
		{Label: "Rounds", Value: fmt.Sprintf("%d", len(archive.Rounds))},
		{Label: "Teams", Value: fmt.Sprintf("%d", overview.TeamCount)},
		{Label: "Services", Value: fmt.Sprintf("%d", overview.ServiceCount)},
		{Label: "Attacks", Value: fmt.Sprintf("%d", overview.AttackCount)},
		{Label: "Traffic", Value: fmt.Sprintf("%d", overview.TrafficCount)},
	})

	addAWDReviewRoundsTable(pdf, archive.Rounds)
	if archive.SelectedRound != nil {
		addAWDReviewSelectedRoundBlock(pdf, archive.SelectedRound)
	}

	return pdf.OutputFileAndClose(targetPath)
}

func writeZIPJSONFile(writer *zip.Writer, name string, payload any) error {
	entry, err := writer.Create(name)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	content = append(content, '\n')
	if _, err := entry.Write(content); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func extractAWDReviewTeams(archive *dto.TeacherAWDReviewArchiveResp) []dto.TeacherAWDReviewTeamResp {
	if archive == nil || archive.SelectedRound == nil {
		return []dto.TeacherAWDReviewTeamResp{}
	}
	return archive.SelectedRound.Teams
}

func addAWDReviewRoundsTable(pdf *gofpdf.Fpdf, rounds []dto.TeacherAWDReviewRoundResp) {
	if len(rounds) == 0 {
		return
	}

	ensurePDFSpace(pdf, 20+float64(len(rounds))*8)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "Rounds", "", 1, "L", false, 0, "")

	headers := []string{"Round", "Status", "Services", "Attacks", "Traffic"}
	widths := []float64{24, 46, 32, 32, 32}
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(230, 230, 230)
	for idx, header := range headers {
		pdf.CellFormat(widths[idx], 7, sanitizePDFText(header), "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 10)
	for _, round := range rounds {
		ensurePDFSpace(pdf, 8)
		values := []string{
			fmt.Sprintf("%d", round.RoundNumber),
			sanitizePDFText(round.Status),
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

func addAWDReviewSelectedRoundBlock(pdf *gofpdf.Fpdf, selected *dto.TeacherAWDSelectedRoundResp) {
	if selected == nil {
		return
	}

	ensurePDFSpace(pdf, 36)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(0, 8, "Selected Round", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)

	lines := []summaryLine{
		{Label: "Round", Value: fmt.Sprintf("%d", selected.Round.RoundNumber)},
		{Label: "Teams", Value: fmt.Sprintf("%d", len(selected.Teams))},
		{Label: "Services", Value: fmt.Sprintf("%d", len(selected.Services))},
		{Label: "Attacks", Value: fmt.Sprintf("%d", len(selected.Attacks))},
		{Label: "Traffic", Value: fmt.Sprintf("%d", len(selected.Traffic))},
	}
	for _, line := range lines {
		pdf.CellFormat(34, 7, sanitizePDFText(line.Label), "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 7, sanitizePDFText(line.Value), "", 1, "L", false, 0, "")
	}
}

package commands

import (
	"context"
	"ctf-platform/internal/config"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"fmt"
	"testing"
	"time"
)

type testAWDReviewExportBuilder struct {
	wait    <-chan struct{}
	archive *assessmentqry.TeacherAWDReviewArchiveResp
}

type testAWDReviewArchiveReader struct {
	archives []*assessmentqry.TeacherAWDReviewArchiveResp
	inputs   []assessmentqry.GetTeacherAWDReviewArchiveInput
}

func (r *testAWDReviewArchiveReader) GetContestArchive(ctx context.Context, requesterID, contestID int64, req assessmentqry.GetTeacherAWDReviewArchiveInput) (*assessmentqry.TeacherAWDReviewArchiveResp, error) {
	_ = ctx
	_ = requesterID
	_ = contestID
	r.inputs = append(r.inputs, req)
	if len(r.archives) == 0 {
		return nil, fmt.Errorf("unexpected GetContestArchive call")
	}
	archive := r.archives[0]
	r.archives = r.archives[1:]
	return archive, nil
}

func (b *testAWDReviewExportBuilder) BuildArchive(ctx context.Context, requesterID, contestID int64, roundNumber *int) (*assessmentqry.TeacherAWDReviewArchiveResp, error) {
	if b != nil && b.wait != nil {
		select {
		case <-b.wait:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if b != nil && b.archive != nil {
		return b.archive, nil
	}

	generatedAt := time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC)
	selectedRoundNumber := 0
	if roundNumber != nil {
		selectedRoundNumber = *roundNumber
	}
	if selectedRoundNumber <= 0 {
		selectedRoundNumber = 1
	}

	return &assessmentqry.TeacherAWDReviewArchiveResp{
		GeneratedAt: generatedAt,
		Scope: assessmentqry.TeacherAWDReviewScopeResp{
			SnapshotType: "final",
			RequestedBy:  requesterID,
			RequestedID:  contestID,
		},
		Contest: assessmentqry.TeacherAWDReviewContestMetaResp{
			ID:         contestID,
			Title:      "awd-review",
			Mode:       contestcontracts.ContestModeAWD,
			Status:     contestcontracts.ContestStatusEnded,
			RoundCount: 1,
			TeamCount:  1,
		},
		Overview: &assessmentqry.TeacherAWDReviewOverviewResp{
			RoundCount:   1,
			TeamCount:    1,
			ServiceCount: 1,
			AttackCount:  1,
			TrafficCount: 1,
		},
		Rounds: []assessmentqry.TeacherAWDReviewRoundResp{{
			ID:           1,
			ContestID:    contestID,
			RoundNumber:  selectedRoundNumber,
			Status:       contestcontracts.AWDRoundStatusFinished,
			ServiceCount: 1,
			AttackCount:  1,
			TrafficCount: 1,
		}},
		SelectedRound: &assessmentqry.TeacherAWDSelectedRoundResp{
			Round: assessmentqry.TeacherAWDReviewRoundResp{
				ID:           1,
				ContestID:    contestID,
				RoundNumber:  selectedRoundNumber,
				Status:       contestcontracts.AWDRoundStatusFinished,
				ServiceCount: 1,
				AttackCount:  1,
				TrafficCount: 1,
			},
			Teams: []assessmentqry.TeacherAWDReviewTeamResp{{
				TeamID:      1,
				TeamName:    "blue",
				CaptainID:   1,
				TotalScore:  100,
				MemberCount: 1,
			}},
			Services: []assessmentqry.TeacherAWDReviewServiceResp{{
				ID:                1,
				RoundID:           1,
				TeamID:            1,
				TeamName:          "blue",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				ServiceStatus:     contestcontracts.AWDServiceStatusUp,
			}},
			Attacks: []assessmentqry.TeacherAWDReviewAttackResp{{
				ID:                1,
				RoundID:           1,
				AttackerTeamID:    1,
				AttackerTeamName:  "blue",
				VictimTeamID:      2,
				VictimTeamName:    "red",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				AttackType:        contestcontracts.AWDAttackTypeFlagCapture,
				Source:            contestcontracts.AWDAttackSourceManual,
			}},
			Traffic: []assessmentqry.TeacherAWDReviewTrafficResp{{
				ID:                1,
				ContestID:         contestID,
				RoundID:           1,
				AttackerTeamID:    1,
				AttackerTeamName:  "blue",
				VictimTeamID:      2,
				VictimTeamName:    "red",
				AWDChallengeID:    1,
				AWDChallengeTitle: "web",
				Method:            "GET",
				Path:              "/health",
				StatusCode:        200,
				Source:            contestcontracts.AWDAttackSourceSubmission,
			}},
		},
	}, nil
}

func TestReportServiceCreateAWDReviewArchiveExportStartsProcessingTask(t *testing.T) {
	t.Parallel()

	db := newTestSQLiteDB(t)
	if err := db.AutoMigrate(&identitycontracts.User{}, &assessmententity.Report{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	teacher := &identitycontracts.User{
		ID:        11,
		Username:  "teacher-awd",
		Role:      identitycontracts.RoleTeacher,
		ClassName: "class-a",
		Status:    identitycontracts.UserStatusActive,
	}
	if err := db.Create(teacher).Error; err != nil {
		t.Fatalf("seed teacher: %v", err)
	}
	contest := &contestcontracts.Contest{
		ID:        21,
		Title:     "awd-ended",
		Mode:      contestcontracts.ContestModeAWD,
		Status:    contestcontracts.ContestStatusEnded,
		StartTime: time.Date(2026, 4, 12, 9, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 4, 12, 11, 0, 0, 0, time.UTC),
	}
	repo := &testReportRepository{
		db: db,
		contests: map[int64]*contestcontracts.Contest{
			contest.ID: contest,
		},
	}

	service := NewReportService(
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		repo,
		nil,
		config.ReportConfig{
			StorageDir:    t.TempDir(),
			DefaultFormat: assessmententity.ReportFormatPDF,
			MaxWorkers:    1,
		},
		nil,
	)
	service.StartBackgroundTasks(context.Background())
	releaseBuilder := make(chan struct{})
	service.SetAWDReviewExportBuilder(&testAWDReviewExportBuilder{wait: releaseBuilder})
	t.Cleanup(func() {
		close(releaseBuilder)
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = service.Close(closeCtx)
	})

	resp, err := service.CreateTeacherAWDReviewArchive(context.Background(), teacher.ID, contest.ID, CreateTeacherAWDReviewExportInput{
		RoundNumber: intPtr(2),
	})
	if err != nil {
		t.Fatalf("CreateTeacherAWDReviewArchive() error = %v", err)
	}
	if resp.Status != assessmententity.ReportStatusProcessing {
		t.Fatalf("expected processing status, got %+v", resp)
	}

	var report assessmententity.Report
	if err := db.First(&report, "id = ?", resp.ReportID).Error; err != nil {
		t.Fatalf("load report: %v", err)
	}
	if report.Type != assessmententity.ReportTypeAWDReviewArchive {
		t.Fatalf("expected report type %s, got %+v", assessmententity.ReportTypeAWDReviewArchive, report)
	}
	if report.Format != assessmententity.ReportFormatZIP {
		t.Fatalf("expected report format %s, got %+v", assessmententity.ReportFormatZIP, report)
	}
	if report.UserID == nil || *report.UserID != teacher.ID {
		t.Fatalf("expected report user_id %d, got %+v", teacher.ID, report)
	}
}

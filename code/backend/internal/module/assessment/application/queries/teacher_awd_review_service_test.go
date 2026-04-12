package queries_test

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestTeacherAWDReviewServiceListContestsReturnsOnlyAWDContests(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 101, now)
	if err := db.Create(&model.Contest{
		ID:        102,
		Title:     "jeopardy-contest",
		Mode:      model.ContestModeJeopardy,
		Status:    model.ContestStatusEnded,
		StartTime: now.Add(-4 * time.Hour),
		EndTime:   now.Add(-2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create jeopardy contest: %v", err)
	}

	resp, err := service.ListContests(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListContests() error = %v", err)
	}
	if len(resp.Contests) != 1 || resp.Contests[0].Mode != model.ContestModeAWD {
		t.Fatalf("expected awd-only contest list, got %+v", resp.Contests)
	}
}

func TestTeacherAWDReviewServiceGetContestArchiveBuildsOverviewAndRounds(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 201, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 20101, 201, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 20102, 201, 2, 70, 30, now.Add(-10*time.Minute), time.Time{})

	resp, err := service.GetContestArchive(context.Background(), 1, 201, &dto.GetTeacherAWDReviewArchiveReq{})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.Overview == nil {
		t.Fatalf("expected overview to be built, got %+v", resp)
	}
	if len(resp.Rounds) != 2 {
		t.Fatalf("expected 2 rounds, got %+v", resp.Rounds)
	}
	if resp.Scope.SnapshotType != "live" {
		t.Fatalf("expected live snapshot, got %s", resp.Scope.SnapshotType)
	}
}

func TestTeacherAWDReviewServiceGetContestArchiveSupportsSelectedRound(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 301, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 30101, 301, 1, 50, 50, now.Add(-50*time.Minute), now.Add(-30*time.Minute))
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 30102, 301, 2, 55, 45, now.Add(-20*time.Minute), now.Add(-5*time.Minute))

	resp, err := service.GetContestArchive(context.Background(), 1, 301, &dto.GetTeacherAWDReviewArchiveReq{
		RoundNumber: intPtr(2),
	})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.SelectedRound == nil || resp.SelectedRound.Round.RoundNumber != 2 {
		t.Fatalf("expected selected round 2, got %+v", resp.SelectedRound)
	}
}

func TestTeacherAWDReviewServiceMarksEndedContestAsFinalSnapshot(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	if err := db.Create(&model.Contest{
		ID:        401,
		Title:     "awd-ended",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusEnded,
		StartTime: now.Add(-2 * time.Hour),
		EndTime:   now.Add(-time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create ended contest: %v", err)
	}
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 40101, 401, 1, 60, 40, now.Add(-110*time.Minute), now.Add(-80*time.Minute))
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 40102, 401, 2, 70, 30, now.Add(-70*time.Minute), now.Add(-40*time.Minute))

	resp, err := service.GetContestArchive(context.Background(), 1, 401, &dto.GetTeacherAWDReviewArchiveReq{})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.Scope.SnapshotType != "final" {
		t.Fatalf("expected final snapshot, got %s", resp.Scope.SnapshotType)
	}
}

func setupTeacherAWDReviewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	return contesttestsupport.SetupAWDTestDB(t)
}

func intPtr(value int) *int {
	return &value
}

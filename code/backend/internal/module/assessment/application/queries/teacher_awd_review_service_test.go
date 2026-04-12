package queries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	"ctf-platform/pkg/errcode"
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

func TestTeacherAWDReviewServiceListContestsBuildsLatestEvidenceAt(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 111, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 11101, 111, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	seedTeacherAWDReviewTeamsAndChallenge(t, db, 111, now)

	attackAt := now.Add(-15 * time.Minute)
	trafficAt := now.Add(-5 * time.Minute)
	seedTeacherAWDReviewSignals(t, db, 111, 11101, attackAt, trafficAt)

	resp, err := service.ListContests(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListContests() error = %v", err)
	}
	if len(resp.Contests) != 1 {
		t.Fatalf("expected 1 contest, got %+v", resp.Contests)
	}
	if resp.Contests[0].LatestEvidenceAt == nil {
		t.Fatalf("expected latest evidence time, got %+v", resp.Contests[0])
	}
	if !resp.Contests[0].LatestEvidenceAt.Equal(trafficAt) {
		t.Fatalf("expected latest evidence %s, got %s", trafficAt, resp.Contests[0].LatestEvidenceAt)
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

func TestTeacherAWDReviewServiceGetContestArchiveBuildsLatestEvidenceAtFromSignals(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 211, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 21101, 211, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	seedTeacherAWDReviewTeamsAndChallenge(t, db, 211, now)

	attackAt := now.Add(-15 * time.Minute)
	trafficAt := now.Add(-5 * time.Minute)
	seedTeacherAWDReviewSignals(t, db, 211, 21101, attackAt, trafficAt)

	resp, err := service.GetContestArchive(context.Background(), 1, 211, &dto.GetTeacherAWDReviewArchiveReq{})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.Contest.LatestEvidenceAt == nil {
		t.Fatalf("expected contest latest evidence time, got %+v", resp.Contest)
	}
	if !resp.Contest.LatestEvidenceAt.Equal(trafficAt) {
		t.Fatalf("expected latest evidence %s, got %s", trafficAt, resp.Contest.LatestEvidenceAt)
	}
}

func TestTeacherAWDReviewServiceGetContestArchiveRejectsTeamIDWithoutRound(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 320, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 32001, 320, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	seedTeacherAWDReviewTeamsAndChallenge(t, db, 320, now)

	_, err := service.GetContestArchive(context.Background(), 1, 320, &dto.GetTeacherAWDReviewArchiveReq{
		TeamID: int64Ptr(3201),
	})
	assertInvalidParamsError(t, err)
}

func TestTeacherAWDReviewServiceGetContestArchiveRejectsUnknownTeamID(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 330, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 33001, 330, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	seedTeacherAWDReviewTeamsAndChallenge(t, db, 330, now)

	_, err := service.GetContestArchive(context.Background(), 1, 330, &dto.GetTeacherAWDReviewArchiveReq{
		RoundNumber: intPtr(1),
		TeamID:      int64Ptr(999999),
	})
	assertInvalidParamsError(t, err)
}

func TestTeacherAWDReviewServiceGetContestArchiveFiltersSelectedRoundByTeam(t *testing.T) {
	t.Parallel()

	db := setupTeacherAWDReviewTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(assessmentinfra.NewTeacherAWDReviewRepository(db))

	contesttestsupport.CreateAWDContestFixture(t, db, 340, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 34001, 340, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))
	seedTeacherAWDReviewTeamsAndChallenge(t, db, 340, now)
	seedTeacherAWDReviewFilterData(t, db, 340, 34001, now)

	resp, err := service.GetContestArchive(context.Background(), 1, 340, &dto.GetTeacherAWDReviewArchiveReq{
		RoundNumber: intPtr(1),
		TeamID:      int64Ptr(3401),
	})
	if err != nil {
		t.Fatalf("GetContestArchive() error = %v", err)
	}
	if resp.SelectedRound == nil {
		t.Fatalf("expected selected round, got %+v", resp)
	}
	if len(resp.SelectedRound.Teams) != 1 || resp.SelectedRound.Teams[0].TeamID != 3401 {
		t.Fatalf("expected selected team only, got %+v", resp.SelectedRound.Teams)
	}
	if len(resp.SelectedRound.Services) != 1 || resp.SelectedRound.Services[0].TeamID != 3401 {
		t.Fatalf("expected selected team services only, got %+v", resp.SelectedRound.Services)
	}
	if len(resp.SelectedRound.Attacks) != 1 {
		t.Fatalf("expected 1 related attack, got %+v", resp.SelectedRound.Attacks)
	}
	if resp.SelectedRound.Attacks[0].AttackerTeamID != 3401 && resp.SelectedRound.Attacks[0].VictimTeamID != 3401 {
		t.Fatalf("expected attack involving selected team, got %+v", resp.SelectedRound.Attacks[0])
	}
	if len(resp.SelectedRound.Traffic) != 1 {
		t.Fatalf("expected 1 related traffic event, got %+v", resp.SelectedRound.Traffic)
	}
	if resp.SelectedRound.Traffic[0].AttackerTeamID != 3401 && resp.SelectedRound.Traffic[0].VictimTeamID != 3401 {
		t.Fatalf("expected traffic involving selected team, got %+v", resp.SelectedRound.Traffic[0])
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

func int64Ptr(value int64) *int64 {
	return &value
}

func assertInvalidParamsError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected invalid params error")
	}

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params code, got %+v", appErr)
	}
}

func seedTeacherAWDReviewTeamsAndChallenge(t *testing.T, db *gorm.DB, contestID int64, now time.Time) {
	t.Helper()

	contesttestsupport.CreateAWDChallengeFixture(t, db, contestID*10+1, now)
	contesttestsupport.CreateAWDContestChallengeFixture(t, db, contestID, contestID*10+1, now)
	contesttestsupport.CreateAWDTeamFixture(t, db, contestID*10+1, contestID, "team-alpha", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, contestID*10+2, contestID, "team-beta", now)
	contesttestsupport.CreateAWDTeamFixture(t, db, contestID*10+3, contestID, "team-gamma", now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, contestID, contestID*10+1, contestID*100+1, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, contestID, contestID*10+2, contestID*100+2, now)
	contesttestsupport.CreateAWDTeamMemberFixture(t, db, contestID, contestID*10+3, contestID*100+3, now)
}

func seedTeacherAWDReviewSignals(t *testing.T, db *gorm.DB, contestID, roundID int64, attackAt, trafficAt time.Time) {
	t.Helper()

	if err := db.Create(&model.AWDTeamService{
		ID:            roundID*10 + 1,
		RoundID:       roundID,
		TeamID:        contestID*10 + 1,
		ChallengeID:   contestID*10 + 1,
		ServiceStatus: model.AWDServiceStatusUp,
		UpdatedAt:     attackAt.Add(-time.Minute),
		CreatedAt:     attackAt.Add(-2 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create team service: %v", err)
	}
	if err := db.Create(&model.AWDAttackLog{
		ID:             roundID*10 + 2,
		RoundID:        roundID,
		AttackerTeamID: contestID*10 + 1,
		VictimTeamID:   contestID*10 + 2,
		ChallengeID:    contestID*10 + 1,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceManual,
		IsSuccess:      true,
		ScoreGained:    10,
		CreatedAt:      attackAt,
	}).Error; err != nil {
		t.Fatalf("create attack log: %v", err)
	}
	if err := db.Create(&model.AWDTrafficEvent{
		ID:             roundID*10 + 3,
		ContestID:      contestID,
		RoundID:        roundID,
		AttackerTeamID: contestID*10 + 2,
		VictimTeamID:   contestID*10 + 1,
		ChallengeID:    contestID*10 + 1,
		Method:         "POST",
		Path:           "/flag",
		StatusCode:     200,
		Source:         model.AWDTrafficSourceRuntimeProxy,
		CreatedAt:      trafficAt,
	}).Error; err != nil {
		t.Fatalf("create traffic event: %v", err)
	}
}

func seedTeacherAWDReviewFilterData(t *testing.T, db *gorm.DB, contestID, roundID int64, now time.Time) {
	t.Helper()

	challengeID := contestID*10 + 1
	rows := []any{
		&model.AWDTeamService{
			ID:            roundID*10 + 1,
			RoundID:       roundID,
			TeamID:        contestID*10 + 1,
			ChallengeID:   challengeID,
			ServiceStatus: model.AWDServiceStatusUp,
			UpdatedAt:     now.Add(-10 * time.Minute),
			CreatedAt:     now.Add(-11 * time.Minute),
		},
		&model.AWDTeamService{
			ID:            roundID*10 + 2,
			RoundID:       roundID,
			TeamID:        contestID*10 + 2,
			ChallengeID:   challengeID,
			ServiceStatus: model.AWDServiceStatusDown,
			UpdatedAt:     now.Add(-9 * time.Minute),
			CreatedAt:     now.Add(-10 * time.Minute),
		},
		&model.AWDTeamService{
			ID:            roundID*10 + 3,
			RoundID:       roundID,
			TeamID:        contestID*10 + 3,
			ChallengeID:   challengeID,
			ServiceStatus: model.AWDServiceStatusCompromised,
			UpdatedAt:     now.Add(-8 * time.Minute),
			CreatedAt:     now.Add(-9 * time.Minute),
		},
		&model.AWDAttackLog{
			ID:             roundID*10 + 4,
			RoundID:        roundID,
			AttackerTeamID: contestID*10 + 1,
			VictimTeamID:   contestID*10 + 2,
			ChallengeID:    challengeID,
			AttackType:     model.AWDAttackTypeFlagCapture,
			Source:         model.AWDAttackSourceManual,
			IsSuccess:      true,
			ScoreGained:    10,
			CreatedAt:      now.Add(-7 * time.Minute),
		},
		&model.AWDAttackLog{
			ID:             roundID*10 + 5,
			RoundID:        roundID,
			AttackerTeamID: contestID*10 + 2,
			VictimTeamID:   contestID*10 + 3,
			ChallengeID:    challengeID,
			AttackType:     model.AWDAttackTypeFlagCapture,
			Source:         model.AWDAttackSourceManual,
			IsSuccess:      true,
			ScoreGained:    8,
			CreatedAt:      now.Add(-6 * time.Minute),
		},
		&model.AWDTrafficEvent{
			ID:             roundID*10 + 6,
			ContestID:      contestID,
			RoundID:        roundID,
			AttackerTeamID: contestID*10 + 1,
			VictimTeamID:   contestID*10 + 2,
			ChallengeID:    challengeID,
			Method:         "GET",
			Path:           "/health",
			StatusCode:     200,
			Source:         model.AWDTrafficSourceRuntimeProxy,
			CreatedAt:      now.Add(-5 * time.Minute),
		},
		&model.AWDTrafficEvent{
			ID:             roundID*10 + 7,
			ContestID:      contestID,
			RoundID:        roundID,
			AttackerTeamID: contestID*10 + 2,
			VictimTeamID:   contestID*10 + 3,
			ChallengeID:    challengeID,
			Method:         "POST",
			Path:           "/exploit",
			StatusCode:     500,
			Source:         model.AWDTrafficSourceRuntimeProxy,
			CreatedAt:      now.Add(-4 * time.Minute),
		},
	}

	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("seed teacher awd review filter data: %v", err)
		}
	}
}

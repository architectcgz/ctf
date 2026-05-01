package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	"ctf-platform/pkg/errcode"
)

type stubContestRepository struct{}

func (s *stubContestRepository) Create(context.Context, *model.Contest) error { return nil }
func (s *stubContestRepository) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}
func (s *stubContestRepository) Update(context.Context, *model.Contest) error { return nil }

func TestContestServiceCreateContestRejectsInvalidTimeRange(t *testing.T) {
	service := contestcmd.NewContestService(&stubContestRepository{}, nil, zap.NewNop())

	_, err := service.CreateContest(context.Background(), contestcmd.CreateContestInput{
		Title:     "contest",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now(),
	})
	if err != errcode.ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestContestServiceCreateContestNormalizesTimeFieldsToUTC(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	shanghai := time.FixedZone("Asia/Shanghai", 8*60*60)
	start := time.Date(2026, 4, 28, 12, 36, 54, 0, shanghai)
	end := start.Add(2 * time.Hour)

	resp, err := service.CreateContest(context.Background(), contestcmd.CreateContestInput{
		Title:       "utc contest",
		Description: "time contract",
		Mode:        model.ContestModeAWD,
		StartTime:   start,
		EndTime:     end,
	})
	if err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	if resp.StartTime.Location() != time.UTC || resp.EndTime.Location() != time.UTC {
		t.Fatalf("expected UTC response times, got start=%v end=%v", resp.StartTime.Location(), resp.EndTime.Location())
	}
	if !resp.StartTime.Equal(start) || !resp.EndTime.Equal(end) {
		t.Fatalf("response changed instant: start=%s end=%s", resp.StartTime, resp.EndTime)
	}

	var stored model.Contest
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load stored contest: %v", err)
	}
	if stored.StartTime.Location() != time.UTC || stored.EndTime.Location() != time.UTC {
		t.Fatalf("expected UTC stored times, got start=%v end=%v", stored.StartTime.Location(), stored.EndTime.Location())
	}
	if !stored.StartTime.Equal(start) || !stored.EndTime.Equal(end) {
		t.Fatalf("stored changed instant: start=%s end=%s", stored.StartTime, stored.EndTime)
	}
}

func TestContestServiceUpdateContestBlocksAWDStartWhenReadinessNotReady(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	now := time.Now()

	createContestForUpdateTest(t, db, &model.Contest{
		ID:        801,
		Title:     "awd-start-block",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRegistration,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})

	_, err := service.UpdateContest(context.Background(), 801, contestcmd.UpdateContestInput{
		Status: strPtr(model.ContestStatusRunning),
	})
	assertContestReadinessBlocked(t, err)
}

func TestContestServiceUpdateContestAllowsAWDStartOverride(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	now := time.Now()

	createContestForUpdateTest(t, db, &model.Contest{
		ID:        802,
		Title:     "awd-start-override",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRegistration,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})

	resp, err := service.UpdateContest(context.Background(), 802, contestcmd.UpdateContestInput{
		Status:         strPtr(model.ContestStatusRunning),
		ForceOverride:  boolPtr(true),
		OverrideReason: strPtr("teacher drill"),
	})
	if err != nil {
		t.Fatalf("UpdateContest() error = %v", err)
	}
	if resp == nil || resp.Status != model.ContestStatusRunning {
		t.Fatalf("unexpected contest response: %+v", resp)
	}
}

func TestContestServiceUpdateContestNormalizesTimeFieldsToUTC(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	now := time.Now().UTC()
	shanghai := time.FixedZone("Asia/Shanghai", 8*60*60)
	start := time.Date(2026, 4, 28, 12, 36, 54, 0, shanghai)
	end := start.Add(2 * time.Hour)

	createContestForUpdateTest(t, db, &model.Contest{
		ID:        805,
		Title:     "time-update",
		Mode:      model.ContestModeJeopardy,
		Status:    model.ContestStatusDraft,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})

	resp, err := service.UpdateContest(context.Background(), 805, contestcmd.UpdateContestInput{
		StartTime: &start,
		EndTime:   &end,
	})
	if err != nil {
		t.Fatalf("UpdateContest() error = %v", err)
	}
	if resp.StartTime.Location() != time.UTC || resp.EndTime.Location() != time.UTC {
		t.Fatalf("expected UTC response times, got start=%v end=%v", resp.StartTime.Location(), resp.EndTime.Location())
	}
	if !resp.StartTime.Equal(start) || !resp.EndTime.Equal(end) {
		t.Fatalf("response changed instant: start=%s end=%s", resp.StartTime, resp.EndTime)
	}

	var stored model.Contest
	if err := db.First(&stored, 805).Error; err != nil {
		t.Fatalf("load stored contest: %v", err)
	}
	if stored.StartTime.Location() != time.UTC || stored.EndTime.Location() != time.UTC {
		t.Fatalf("expected UTC stored times, got start=%v end=%v", stored.StartTime.Location(), stored.EndTime.Location())
	}
	if !stored.StartTime.Equal(start) || !stored.EndTime.Equal(end) {
		t.Fatalf("stored changed instant: start=%s end=%s", stored.StartTime, stored.EndTime)
	}
}

func TestContestServiceUpdateContestRejectsBlankOverrideReason(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	now := time.Now()

	createContestForUpdateTest(t, db, &model.Contest{
		ID:        803,
		Title:     "awd-start-blank-reason",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRegistration,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})

	_, err := service.UpdateContest(context.Background(), 803, contestcmd.UpdateContestInput{
		Status:         strPtr(model.ContestStatusRunning),
		ForceOverride:  boolPtr(true),
		OverrideReason: strPtr("  "),
	})
	if err != errcode.ErrInvalidParams {
		t.Fatalf("expected ErrInvalidParams, got %v", err)
	}
}

func TestContestServiceUpdateContestDoesNotGateNonAWDStatusUpdate(t *testing.T) {
	service, db := newContestCommandServiceForTest(t)
	now := time.Now()

	createContestForUpdateTest(t, db, &model.Contest{
		ID:        804,
		Title:     "jeopardy-start",
		Mode:      model.ContestModeJeopardy,
		Status:    model.ContestStatusRegistration,
		StartTime: now.Add(time.Hour),
		EndTime:   now.Add(2 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	})

	resp, err := service.UpdateContest(context.Background(), 804, contestcmd.UpdateContestInput{
		Status: strPtr(model.ContestStatusRunning),
	})
	if err != nil {
		t.Fatalf("UpdateContest() error = %v", err)
	}
	if resp == nil || resp.Status != model.ContestStatusRunning {
		t.Fatalf("unexpected contest response: %+v", resp)
	}
}

func newContestCommandServiceForTest(t *testing.T) (*contestcmd.ContestService, *gorm.DB) {
	t.Helper()

	db := contesttestsupport.SetupAWDTestDB(t)
	return contestcmd.NewContestService(contestinfra.NewRepository(db), contestinfra.NewAWDRepository(db), zap.NewNop()), db
}

func createContestForUpdateTest(t *testing.T, db *gorm.DB, contest *model.Contest) {
	t.Helper()

	if err := db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
}

func assertContestReadinessBlocked(t *testing.T, err error) {
	t.Helper()

	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrAWDReadinessBlocked.Code {
		t.Fatalf("expected ErrAWDReadinessBlocked, got %v", err)
	}
}

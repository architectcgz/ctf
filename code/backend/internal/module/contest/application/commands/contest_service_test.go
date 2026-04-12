package commands_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
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

	_, err := service.CreateContest(context.Background(), &dto.CreateContestReq{
		Title:     "contest",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now(),
	})
	if err != errcode.ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
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

	_, err := service.UpdateContest(context.Background(), 801, &dto.UpdateContestReq{
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

	resp, err := service.UpdateContest(context.Background(), 802, &dto.UpdateContestReq{
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

	_, err := service.UpdateContest(context.Background(), 803, &dto.UpdateContestReq{
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

	resp, err := service.UpdateContest(context.Background(), 804, &dto.UpdateContestReq{
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

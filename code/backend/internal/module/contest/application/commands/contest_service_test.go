package commands_test

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

type stubContestRepository struct{}

func (s *stubContestRepository) Create(context.Context, *model.Contest) error { return nil }
func (s *stubContestRepository) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}
func (s *stubContestRepository) Update(context.Context, *model.Contest) error { return nil }
func (s *stubContestRepository) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}
func (s *stubContestRepository) ListByStatusesAndTimeRange(context.Context, []string, time.Time, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}
func (s *stubContestRepository) UpdateStatus(context.Context, int64, string) error { return nil }
func (s *stubContestRepository) FindTeamsByIDs(context.Context, []int64) ([]*model.Team, error) {
	return nil, nil
}
func (s *stubContestRepository) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return nil, nil
}
func (s *stubContestRepository) FindScoreboardTeamStats(context.Context, int64, string, []int64) (map[int64]contestports.ScoreboardTeamStats, error) {
	return nil, nil
}

func TestContestServiceCreateContestRejectsInvalidTimeRange(t *testing.T) {
	service := contestcmd.NewContestService(&stubContestRepository{}, zap.NewNop())

	_, err := service.CreateContest(context.Background(), &dto.CreateContestReq{
		Title:     "contest",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now(),
	})
	if err != errcode.ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}

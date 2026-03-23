package application

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type stubRepository struct{}

func (s *stubRepository) Create(context.Context, *model.Contest) error {
	return nil
}

func (s *stubRepository) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, ErrContestNotFound
}

func (s *stubRepository) Update(context.Context, *model.Contest) error {
	return nil
}

func (s *stubRepository) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}

func (s *stubRepository) ListByStatusesAndTimeRange(context.Context, []string, time.Time, int, int) ([]*model.Contest, int64, error) {
	return nil, 0, nil
}

func (s *stubRepository) UpdateStatus(context.Context, int64, string) error {
	return nil
}

func (s *stubRepository) FindTeamsByIDs(context.Context, []int64) ([]*model.Team, error) {
	return nil, nil
}

func (s *stubRepository) FindTeamsByContest(context.Context, int64) ([]*model.Team, error) {
	return nil, nil
}

func (s *stubRepository) FindScoreboardTeamStats(context.Context, int64, string, []int64) (map[int64]ScoreboardTeamStats, error) {
	return nil, nil
}

func TestServiceCreateContestRejectsInvalidTimeRange(t *testing.T) {
	service := NewService(&stubRepository{}, zap.NewNop())

	_, err := service.CreateContest(context.Background(), &dto.CreateContestReq{
		Title:     "contest",
		StartTime: time.Now().Add(time.Hour),
		EndTime:   time.Now(),
	})
	if err != errcode.ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestScoreboardServiceCalculateDynamicScoreWithBaseUsesDefaultBaseWhenInputNonPositive(t *testing.T) {
	service := NewScoreboardService(&stubRepository{}, nil, &config.ContestConfig{
		BaseScore: 500,
		MinScore:  100,
		Decay:     0.8,
	}, zap.NewNop())

	got := service.CalculateDynamicScoreWithBase(0, 3)
	want := calculateDynamicScore(500, 100, 0.8, 3)
	if got != want {
		t.Fatalf("expected score %d, got %d", want, got)
	}
}

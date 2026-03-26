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
	"ctf-platform/pkg/errcode"
)

type stubContestRepository struct{}

func (s *stubContestRepository) Create(context.Context, *model.Contest) error { return nil }
func (s *stubContestRepository) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}
func (s *stubContestRepository) Update(context.Context, *model.Contest) error { return nil }

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

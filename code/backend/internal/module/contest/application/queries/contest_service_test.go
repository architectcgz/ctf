package queries

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

type contestListRepoStub struct{}

func (s *contestListRepoStub) FindByID(context.Context, int64) (*model.Contest, error) {
	return nil, contestdomain.ErrContestNotFound
}

func (s *contestListRepoStub) List(context.Context, *string, int, int) ([]*model.Contest, int64, error) {
	return []*model.Contest{}, 0, nil
}

func TestContestServiceGetContestReturnsContestNotFound(t *testing.T) {
	t.Parallel()

	service := NewContestService(&contestListRepoStub{}, zap.NewNop())

	_, err := service.GetContest(context.Background(), 42)
	if err != errcode.ErrContestNotFound {
		t.Fatalf("expected ErrContestNotFound, got %v", err)
	}
}

func TestContestServiceListContestsAcceptsNarrowRepository(t *testing.T) {
	t.Parallel()

	service := NewContestService(&contestListRepoStub{}, zap.NewNop())

	items, total, err := service.ListContests(context.Background(), &dto.ListContestsReq{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("ListContests() error = %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Fatalf("unexpected list result: items=%d total=%d", len(items), total)
	}
}
